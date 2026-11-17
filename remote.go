package main

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

// The kinds of server Ferry can connect to, the same choices FileZilla offers.
const (
	protoSFTP  = "sftp"  // over SSH
	protoFTPES = "ftpes" // FTP, upgraded to TLS ("FTP over TLS", explicit)
	protoFTPS  = "ftps"  // FTP inside TLS from the start (implicit, port 990)
	protoFTP   = "ftp"   // plain FTP, nothing encrypted
)

const remoteTimeout = 20 * time.Second

// remoteConfig says how to reach a server. The password only lives in memory
// while connected and is never written to disk.
type remoteConfig struct {
	Protocol string `json:"protocol"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password,omitempty"`
	// Trust is a fingerprint the person has just agreed to trust.
	Trust string `json:"trust,omitempty"`
}

func defaultPort(protocol string) int {
	switch protocol {
	case protoSFTP:
		return 22
	case protoFTPS:
		return 990
	}
	return 21
}

// clean checks the config and fills in defaults. It accepts a host typed as
// "sftp://user@example.com:2222/", which is how people often paste them.
func (c remoteConfig) clean() (remoteConfig, error) {
	h := strings.TrimSpace(c.Host)
	if i := strings.Index(h, "://"); i >= 0 {
		switch p := strings.ToLower(h[:i]); p {
		case "sftp", "ftp", "ftps", "ftpes":
			c.Protocol = p
		}
		h = h[i+3:]
	}
	h = strings.TrimRight(h, "/")
	if i := strings.Index(h, "/"); i >= 0 {
		h = h[:i]
	}
	if i := strings.LastIndex(h, "@"); i >= 0 {
		if c.User == "" {
			c.User = h[:i]
		}
		h = h[i+1:]
	}
	if host, port, err := net.SplitHostPort(h); err == nil {
		h = host
		if n, err := strconv.Atoi(port); err == nil && c.Port == 0 {
			c.Port = n
		}
	}
	h = strings.Trim(h, "[]")
	switch c.Protocol {
	case protoSFTP, protoFTPES, protoFTPS, protoFTP:
	case "":
		c.Protocol = protoSFTP
	default:
		return c, errors.New("unknown kind of server")
	}
	if h == "" || strings.ContainsAny(h, " /\\") {
		return c, errors.New("type the server's address, for example ftp.example.com")
	}
	c.Host = h
	if c.Port == 0 {
		c.Port = defaultPort(c.Protocol)
	}
	if c.Port < 1 || c.Port > 65535 {
		return c, errors.New("the port must be between 1 and 65535")
	}
	c.User = strings.TrimSpace(c.User)
	return c, nil
}

func (c remoteConfig) addr() string { return net.JoinHostPort(c.Host, strconv.Itoa(c.Port)) }

// key identifies a saved connection: same kind, server and user.
func (c remoteConfig) key() string {
	return c.Protocol + "://" + c.User + "@" + c.addr()
}

// label is how a connection is shown, like "sftp://pi@192.168.1.5".
func (c remoteConfig) label() string {
	s := c.Protocol + "://"
	if c.User != "" {
		s += c.User + "@"
	}
	s += c.Host
	if c.Port != defaultPort(c.Protocol) {
		s += ":" + strconv.Itoa(c.Port)
	}
	return s
}

type remoteEntry struct {
	Name string    `json:"name"`
	Dir  bool      `json:"dir"`
	Size int64     `json:"size"`
	Mod  time.Time `json:"mod"`
}

// remoteFS is one connection to a server. Paths always use forward slashes.
// A connection is used by one goroutine at a time.
type remoteFS interface {
	Home() (string, error)
	List(p string) ([]remoteEntry, error)
	Open(p string) (io.ReadCloser, error)
	Create(p string, r io.Reader) error
	Mkdir(p string) error
	Remove(p string) error
	RemoveDir(p string) error
	Rename(from, to string) error
	Close() error
}

// trustError means the server's identity could not be checked, so Ferry
// asks the person before going on, like FileZilla does.
type trustError struct {
	Kind        string `json:"kind"` // "ssh" or "tls"
	Host        string `json:"host"`
	Fingerprint string `json:"fingerprint"`
	Changed     bool   `json:"changed"` // the server's key is different from last time
}

func (e *trustError) Error() string {
	if e.Changed {
		return "the identity of " + e.Host + " has changed since you last connected. " +
			"This can mean someone is pretending to be the server, so Ferry will not connect. " +
			"If the server was reinstalled, remove it from the trusted servers in Settings and try again"
	}
	return "Ferry does not recognise " + e.Host + " yet"
}

// trustStore keeps the fingerprints of servers the person agreed to trust.
type trustStore struct {
	mu   sync.Mutex
	path string
}

func (t *trustStore) load() map[string]string {
	m := map[string]string{}
	if b, err := os.ReadFile(t.path); err == nil {
		json.Unmarshal(b, &m)
	}
	return m
}

func (t *trustStore) get(id string) (string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fp, ok := t.load()[id]
	return fp, ok
}

func (t *trustStore) set(id, fp string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	m := t.load()
	m[id] = fp
	writeJSON(t.path, m)
}

func (t *trustStore) list() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	var out []string
	for id := range t.load() {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func (t *trustStore) remove(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	m := t.load()
	delete(m, id)
	writeJSON(t.path, m)
}

// writeJSON saves v so that a crash halfway never leaves a broken file.
func writeJSON(file string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := file + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, file)
}

// check decides whether a server with this fingerprint may be used.
func (t *trustStore) check(c remoteConfig, kind, fp string) error {
	id := kind + " " + c.addr()
	if saved, ok := t.get(id); ok {
		if saved == fp {
			return nil
		}
		return &trustError{Kind: kind, Host: c.Host, Fingerprint: fp, Changed: true}
	}
	if c.Trust != "" && c.Trust == fp {
		t.set(id, fp)
		return nil
	}
	return &trustError{Kind: kind, Host: c.Host, Fingerprint: fp}
}

// dialRemote connects and logs in.
func dialRemote(c remoteConfig, trust *trustStore) (remoteFS, error) {
	if c.Protocol == protoSFTP {
		return dialSFTP(c, trust)
	}
	return dialFTP(c, trust)
}

// friendlyDialError turns network errors into something a person can act on.
func friendlyDialError(c remoteConfig, err error) error {
	var te *trustError
	if errors.As(err, &te) {
		return te
	}
	var ne net.Error
	msg := err.Error()
	switch {
	case errors.As(err, &ne) && ne.Timeout():
		return fmt.Errorf("%s did not answer. Check the address and port, and that the server is running", c.addr())
	case strings.Contains(msg, "connection refused"):
		return fmt.Errorf("%s refused the connection. Check the port and the kind of server (SFTP usually uses 22, FTP 21)", c.addr())
	case strings.Contains(msg, "no such host"):
		return fmt.Errorf("could not find a server called %s", c.Host)
	case strings.Contains(msg, "unable to authenticate"), strings.Contains(msg, "530"):
		return errors.New("the server did not accept that user name and password")
	case strings.Contains(msg, "first record does not look like a TLS handshake"):
		return errors.New("this server does not use implicit FTPS. Try \"FTP over TLS\" instead")
	case strings.Contains(msg, "AUTH") && (strings.Contains(msg, "500") || strings.Contains(msg, "502") || strings.Contains(msg, "534")):
		return errors.New("this server does not support encryption. Use plain FTP only if you trust the network")
	}
	return err
}

// --- SFTP

type sftpFS struct {
	ssh *ssh.Client
	c   *sftp.Client
}

func dialSFTP(c remoteConfig, trust *trustStore) (remoteFS, error) {
	var userHosts ssh.HostKeyCallback
	if home, err := os.UserHomeDir(); err == nil {
		userHosts, _ = knownhosts.New(filepath.Join(home, ".ssh", "known_hosts"))
	}
	hostKey := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		fp := ssh.FingerprintSHA256(key)
		// Servers already in the person's own ~/.ssh/known_hosts are trusted.
		if userHosts != nil {
			err := userHosts(hostname, remote, key)
			if err == nil {
				return nil
			}
			var ke *knownhosts.KeyError
			if errors.As(err, &ke) && len(ke.Want) > 0 {
				return &trustError{Kind: "ssh", Host: c.Host, Fingerprint: fp, Changed: true}
			}
		}
		return trust.check(c, "ssh", fp)
	}

	var auth []ssh.AuthMethod
	if c.Password != "" {
		pw := c.Password
		auth = append(auth, ssh.Password(pw), ssh.KeyboardInteractive(
			func(_, _ string, questions []string, _ []bool) ([]string, error) {
				answers := make([]string, len(questions))
				for i := range answers {
					answers[i] = pw
				}
				return answers, nil
			}))
	} else {
		// No password: try the person's SSH agent and keys, like the ssh command does.
		if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
			if conn, err := net.Dial("unix", sock); err == nil {
				defer conn.Close()
				auth = append(auth, ssh.PublicKeysCallback(agent.NewClient(conn).Signers))
			}
		}
		if signers := defaultKeys(); len(signers) > 0 {
			auth = append(auth, ssh.PublicKeys(signers...))
		}
	}
	cfg := &ssh.ClientConfig{
		User:            c.User,
		Auth:            auth,
		HostKeyCallback: hostKey,
		Timeout:         remoteTimeout,
	}
	client, err := ssh.Dial("tcp", c.addr(), cfg)
	if err != nil {
		if c.Password == "" && strings.Contains(err.Error(), "unable to authenticate") {
			return nil, errors.New("the server needs a password, or an SSH key Ferry can use")
		}
		return nil, friendlyDialError(c, err)
	}
	sc, err := sftp.NewClient(client, sftp.UseConcurrentReads(true), sftp.UseConcurrentWrites(true))
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("the server allows SSH but not file transfers (SFTP): %w", err)
	}
	return &sftpFS{ssh: client, c: sc}, nil
}

// defaultKeys loads the person's SSH keys that have no passphrase.
func defaultKeys() []ssh.Signer {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	var out []ssh.Signer
	for _, name := range []string{"id_ed25519", "id_ecdsa", "id_rsa"} {
		b, err := os.ReadFile(filepath.Join(home, ".ssh", name))
		if err != nil {
			continue
		}
		if s, err := ssh.ParsePrivateKey(b); err == nil {
			out = append(out, s)
		}
	}
	return out
}

func (f *sftpFS) Home() (string, error) { return f.c.Getwd() }

func (f *sftpFS) List(p string) ([]remoteEntry, error) {
	infos, err := f.c.ReadDir(p)
	if err != nil {
		return nil, err
	}
	out := make([]remoteEntry, 0, len(infos))
	for _, fi := range infos {
		e := remoteEntry{Name: fi.Name(), Dir: fi.IsDir(), Size: fi.Size(), Mod: fi.ModTime()}
		if fi.Mode()&os.ModeSymlink != 0 {
			if t, err := f.c.Stat(path.Join(p, fi.Name())); err == nil {
				e.Dir, e.Size = t.IsDir(), t.Size()
			}
		}
		if e.Dir {
			e.Size = 0
		}
		out = append(out, e)
	}
	return out, nil
}

func (f *sftpFS) Open(p string) (io.ReadCloser, error) { return f.c.Open(p) }

func (f *sftpFS) Create(p string, r io.Reader) error {
	w, err := f.c.Create(p)
	if err != nil {
		return err
	}
	_, err = w.ReadFrom(r)
	if cerr := w.Close(); err == nil {
		err = cerr
	}
	return err
}

func (f *sftpFS) Mkdir(p string) error         { return f.c.Mkdir(p) }
func (f *sftpFS) Remove(p string) error        { return f.c.Remove(p) }
func (f *sftpFS) RemoveDir(p string) error     { return f.c.RemoveDirectory(p) }
func (f *sftpFS) Rename(from, to string) error { return f.c.Rename(from, to) }
func (f *sftpFS) Close() error                 { f.c.Close(); return f.ssh.Close() }

// --- FTP, FTPS and FTP over TLS

type ftpFS struct{ c *ftp.ServerConn }

func dialFTP(c remoteConfig, trust *trustStore) (remoteFS, error) {
	opts := []ftp.DialOption{
		ftp.DialWithDialer(net.Dialer{Timeout: remoteTimeout, KeepAlive: 30 * time.Second}),
		ftp.DialWithShutTimeout(10 * time.Second),
	}
	if c.Protocol != protoFTP {
		tc := &tls.Config{
			ServerName:         c.Host,
			ClientSessionCache: tls.NewLRUClientSessionCache(8), // many servers want data connections to resume the session
			// The certificate is checked below instead, so a self-signed one can
			// be trusted once by the person rather than refused outright.
			InsecureSkipVerify: true,
			VerifyConnection: func(cs tls.ConnectionState) error {
				return verifyCert(c, cs, trust)
			},
		}
		if c.Protocol == protoFTPS {
			opts = append(opts, ftp.DialWithTLS(tc))
		} else {
			opts = append(opts, ftp.DialWithExplicitTLS(tc))
		}
	}
	conn, err := ftp.Dial(c.addr(), opts...)
	if err != nil {
		return nil, friendlyDialError(c, err)
	}
	user, pw := c.User, c.Password
	if user == "" {
		user, pw = "anonymous", "anonymous@"
	}
	if err := conn.Login(user, pw); err != nil {
		conn.Quit()
		return nil, friendlyDialError(c, err)
	}
	return &ftpFS{c: conn}, nil
}

// verifyCert accepts certificates the system trusts, and otherwise ones the
// person has agreed to trust before.
func verifyCert(c remoteConfig, cs tls.ConnectionState, trust *trustStore) error {
	if len(cs.PeerCertificates) == 0 {
		return errors.New("the server sent no certificate")
	}
	leaf := cs.PeerCertificates[0]
	inter := x509.NewCertPool()
	for _, ic := range cs.PeerCertificates[1:] {
		inter.AddCert(ic)
	}
	if _, err := leaf.Verify(x509.VerifyOptions{DNSName: c.Host, Intermediates: inter}); err == nil {
		return nil
	}
	sum := sha256.Sum256(leaf.Raw)
	hexes := make([]string, len(sum))
	for i, b := range sum {
		hexes[i] = fmt.Sprintf("%02X", b)
	}
	return trust.check(c, "tls", "SHA256:"+strings.Join(hexes, ":"))
}

func (f *ftpFS) Home() (string, error) { return f.c.CurrentDir() }

func (f *ftpFS) List(p string) ([]remoteEntry, error) {
	entries, err := f.c.List(p)
	if err != nil {
		return nil, err
	}
	out := make([]remoteEntry, 0, len(entries))
	for _, e := range entries {
		if e.Name == "." || e.Name == ".." || e.Name == "" {
			continue
		}
		dir := e.Type == ftp.EntryTypeFolder
		size := int64(e.Size)
		if dir {
			size = 0
		}
		out = append(out, remoteEntry{Name: e.Name, Dir: dir, Size: size, Mod: e.Time})
	}
	return out, nil
}

func (f *ftpFS) Open(p string) (io.ReadCloser, error) { return f.c.Retr(p) }
func (f *ftpFS) Create(p string, r io.Reader) error   { return f.c.Stor(p, r) }
func (f *ftpFS) Mkdir(p string) error                 { return f.c.MakeDir(p) }
func (f *ftpFS) Remove(p string) error                { return f.c.Delete(p) }
func (f *ftpFS) RemoveDir(p string) error             { return f.c.RemoveDir(p) }
func (f *ftpFS) Rename(from, to string) error         { return f.c.Rename(from, to) }
func (f *ftpFS) Close() error                         { return f.c.Quit() }

// safeName reports whether a name from a server can be used as a file name
// here. A hostile server could otherwise send "../../something".
func safeName(name string) bool {
	return name != "" && name != "." && name != ".." &&
		!strings.ContainsAny(name, "/\\\x00") && filepath.IsLocal(name)
}

// isConnLost reports whether an error means the connection itself broke,
// so it is worth reconnecting and trying again.
func isConnLost(err error) bool {
	if err == nil {
		return false
	}
	var ne net.Error
	if errors.As(err, &ne) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, net.ErrClosed) || errors.Is(err, sftp.ErrSSHFxConnectionLost) {
		return true
	}
	msg := err.Error()
	for _, s := range []string{"421", "connection lost", "broken pipe", "connection reset", "use of closed"} {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}
