package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func TestRemoteConfigClean(t *testing.T) {
	cases := []struct {
		in   remoteConfig
		want remoteConfig
	}{
		{remoteConfig{Host: "example.com"}, remoteConfig{Protocol: "sftp", Host: "example.com", Port: 22}},
		{remoteConfig{Protocol: "ftp", Host: " ftp.example.com "}, remoteConfig{Protocol: "ftp", Host: "ftp.example.com", Port: 21}},
		{remoteConfig{Host: "sftp://pi@192.168.1.5:2222/home/pi"}, remoteConfig{Protocol: "sftp", Host: "192.168.1.5", Port: 2222, User: "pi"}},
		{remoteConfig{Protocol: "ftps", Host: "secure.example.com"}, remoteConfig{Protocol: "ftps", Host: "secure.example.com", Port: 990}},
		{remoteConfig{Protocol: "ftpes", Host: "[::1]:2121"}, remoteConfig{Protocol: "ftpes", Host: "::1", Port: 2121}},
	}
	for _, c := range cases {
		got, err := c.in.clean()
		if err != nil || got != c.want {
			t.Errorf("clean(%+v) = %+v, %v; want %+v", c.in, got, err, c.want)
		}
	}
	for _, bad := range []remoteConfig{{Host: ""}, {Host: "a b"}, {Host: "x", Port: 70000}, {Host: "x", Protocol: "gopher"}} {
		if _, err := bad.clean(); err == nil {
			t.Errorf("clean(%+v) should fail", bad)
		}
	}
	c := remoteConfig{Protocol: "sftp", Host: "h", Port: 22, User: "u"}
	if c.label() != "sftp://u@h" || c.key() != "sftp://u@h:22" {
		t.Errorf("label %q key %q", c.label(), c.key())
	}
}

func TestSafeName(t *testing.T) {
	for _, ok := range []string{"a.txt", "My Photos", ".hidden", "..dots"} {
		if !safeName(ok) {
			t.Errorf("%q should be allowed", ok)
		}
	}
	for _, bad := range []string{"", ".", "..", "../x", "a/b", `a\b`, "x\x00"} {
		if safeName(bad) {
			t.Errorf("%q should be refused", bad)
		}
	}
}

// startSFTP runs an SFTP server for user "me" with password "pw", serving
// root as its home folder. It returns the address and a way to swap the host key.
func startSFTP(t *testing.T, root string) (string, func()) {
	t.Helper()
	newKey := func() ssh.Signer {
		_, priv, _ := ed25519.GenerateKey(rand.Reader)
		s, err := ssh.NewSignerFromKey(priv)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	var mu sync.Mutex
	key := newKey()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			nc, err := ln.Accept()
			if err != nil {
				return
			}
			cfg := &ssh.ServerConfig{PasswordCallback: func(c ssh.ConnMetadata, pw []byte) (*ssh.Permissions, error) {
				if c.User() == "me" && string(pw) == "pw" {
					return nil, nil
				}
				return nil, errors.New("no")
			}}
			mu.Lock()
			cfg.AddHostKey(key)
			mu.Unlock()
			go serveSSH(nc, cfg, root)
		}
	}()
	return ln.Addr().String(), func() { mu.Lock(); key = newKey(); mu.Unlock() }
}

func serveSSH(nc net.Conn, cfg *ssh.ServerConfig, root string) {
	_, chans, reqs, err := ssh.NewServerConn(nc, cfg)
	if err != nil {
		nc.Close()
		return
	}
	go ssh.DiscardRequests(reqs)
	for nch := range chans {
		if nch.ChannelType() != "session" {
			nch.Reject(ssh.UnknownChannelType, "no")
			continue
		}
		ch, in, err := nch.Accept()
		if err != nil {
			continue
		}
		go func() {
			for req := range in {
				ok := req.Type == "subsystem" && len(req.Payload) > 4 && string(req.Payload[4:]) == "sftp"
				req.Reply(ok, nil)
				if ok {
					srv, err := sftp.NewServer(ch, sftp.WithServerWorkingDirectory(root))
					if err == nil {
						srv.Serve()
					}
					ch.Close()
				}
			}
		}()
	}
}

func remoteTestApp(t *testing.T) (*app, http.Handler) {
	a := &app{token: "secret", port: 47615, dir: t.TempDir(), quit: make(chan struct{})}
	return a, a.routes()
}

func call(t *testing.T, h http.Handler, path string, body any) map[string]any {
	t.Helper()
	b, _ := json.Marshal(body)
	code, out := panelCall(t, h, "POST", path, string(b))
	if code != 200 {
		t.Fatalf("%s: %d %v", path, code, out)
	}
	return out
}

func waitTransfers(t *testing.T, a *app) []transferView {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if !a.remote().queue.busy() {
			return a.remote().queue.views()
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("transfers did not finish")
	return nil
}

func TestSFTPEndToEnd(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the test SFTP server does not serve Windows paths; the client itself is the same on every system")
	}
	server := t.TempDir()
	addr, swapKey := startSFTP(t, server)
	host, port, _ := net.SplitHostPort(addr)
	a, h := remoteTestApp(t)
	conn := map[string]any{"protocol": "sftp", "host": host, "port": atoi(port), "user": "me", "password": "pw"}

	// A wrong password is reported plainly.
	bad := map[string]any{"protocol": "sftp", "host": host, "port": atoi(port), "user": "me", "password": "nope"}
	b, _ := json.Marshal(bad)
	// First the server is unknown, so Ferry asks before even trying the password.
	out := call(t, h, "/api/remote/connect", bad)
	if out["needsTrust"] != true || !strings.HasPrefix(out["fingerprint"].(string), "SHA256:") {
		t.Fatalf("expected a trust question, got %v", out)
	}
	bad["trust"] = out["fingerprint"]
	b, _ = json.Marshal(bad)
	if code, out := panelCall(t, h, "POST", "/api/remote/connect", string(b)); code != 400 || !strings.Contains(out["error"].(string), "user name and password") {
		t.Fatalf("wrong password: %d %v", code, out)
	}

	// Now trusted, the right password connects and opens the home folder.
	out = call(t, h, "/api/remote/connect", conn)
	if out["path"] != filepath.ToSlash(server) && !strings.HasSuffix(out["path"].(string), filepath.Base(server)) {
		t.Fatalf("connect: %v", out)
	}
	remoteHome := out["path"].(string)

	// Upload a folder with a subfolder.
	local := t.TempDir()
	os.MkdirAll(filepath.Join(local, "Photos", "2026"), 0o755)
	os.WriteFile(filepath.Join(local, "Photos", "a.jpg"), []byte("aaaa"), 0o644)
	os.WriteFile(filepath.Join(local, "Photos", "2026", "b.jpg"), []byte(strings.Repeat("b", 300000)), 0o644)
	os.WriteFile(filepath.Join(local, "notes.txt"), []byte("hello"), 0o644)
	call(t, h, "/api/transfer", map[string]any{"up": true, "localDir": local, "remoteDir": remoteHome,
		"items": []map[string]any{{"name": "Photos", "dir": true}, {"name": "notes.txt", "size": 5}}})
	for _, v := range waitTransfers(t, a) {
		if v.State != "done" {
			t.Fatalf("upload %s: %s %s", v.Name, v.State, v.Error)
		}
	}
	if got, _ := os.ReadFile(filepath.Join(server, "Photos", "2026", "b.jpg")); len(got) != 300000 {
		t.Fatalf("nested upload size %d", len(got))
	}

	// Browse it.
	out = call(t, h, "/api/remote/list", map[string]any{"path": remoteHome})
	es := out["entries"].([]any)
	if len(es) != 2 || es[0].(map[string]any)["name"] != "Photos" || es[0].(map[string]any)["dir"] != true {
		t.Fatalf("listing: %v", es)
	}

	// Make a folder, rename a file, then download everything somewhere else.
	call(t, h, "/api/remote/mkdir", map[string]any{"dir": remoteHome, "name": "Empty"})
	call(t, h, "/api/remote/rename", map[string]any{"dir": remoteHome, "from": "notes.txt", "to": "notes-2.txt"})
	call(t, h, "/api/transfer", map[string]any{"up": false, "localDir": t.TempDir(), "remoteDir": remoteHome,
		"items": []map[string]any{{"name": "Photos", "dir": true}}})
	dest := t.TempDir()
	call(t, h, "/api/transfer/clear", map[string]any{})
	call(t, h, "/api/transfer", map[string]any{"up": false, "localDir": dest, "remoteDir": remoteHome,
		"items": []map[string]any{{"name": "Photos", "dir": true}, {"name": "notes-2.txt"}}})
	for _, v := range waitTransfers(t, a) {
		if v.State != "done" {
			t.Fatalf("download %s: %s %s", v.Name, v.State, v.Error)
		}
	}
	if got, _ := os.ReadFile(filepath.Join(dest, "Photos", "2026", "b.jpg")); len(got) != 300000 {
		t.Fatalf("nested download size %d", len(got))
	}
	if got, _ := os.ReadFile(filepath.Join(dest, "notes-2.txt")); string(got) != "hello" {
		t.Fatalf("renamed file: %q", got)
	}
	if leftovers, _ := filepath.Glob(filepath.Join(dest, ".*ferry-part")); len(leftovers) > 0 {
		t.Fatalf("partial files left behind: %v", leftovers)
	}

	// Delete a folder with things in it.
	call(t, h, "/api/remote/delete", map[string]any{"dir": remoteHome, "items": []map[string]any{{"name": "Photos", "dir": true}}})
	if _, err := os.Stat(filepath.Join(server, "Photos")); !os.IsNotExist(err) {
		t.Fatal("folder was not deleted")
	}

	// The connection is remembered, with no password anywhere on disk.
	st := a.remote().state()
	if !st.Connected || len(st.Recent) != 1 || st.Recent[0].User != "me" {
		t.Fatalf("state: %+v", st)
	}
	saved, _ := os.ReadFile(filepath.Join(a.dir, "connections.json"))
	if strings.Contains(string(saved), "pw") || strings.Contains(string(saved), "password") {
		t.Fatalf("password saved to disk: %s", saved)
	}
	call(t, h, "/api/remote/disconnect", map[string]any{})

	// A server whose key changed is refused, even with the right password.
	swapKey()
	b, _ = json.Marshal(conn)
	if code, out := panelCall(t, h, "POST", "/api/remote/connect", string(b)); code != 400 || !strings.Contains(out["error"].(string), "changed") {
		t.Fatalf("changed key: %d %v", code, out)
	}

	// Forgetting removes it from the list.
	call(t, h, "/api/remote/forget", map[string]any{"key": st.Recent[0].Key})
	if len(a.remote().state().Recent) != 0 {
		t.Fatal("forget did not remove the connection")
	}
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}

// fakeFS is a server that lists whatever it is told to, for testing what
// Ferry does with odd or hostile answers.
type fakeFS struct {
	entries map[string][]remoteEntry
	files   map[string]string
	slow    chan struct{}
}

func (f *fakeFS) Home() (string, error)                { return "/", nil }
func (f *fakeFS) List(p string) ([]remoteEntry, error) { return f.entries[p], nil }
func (f *fakeFS) Mkdir(string) error                   { return nil }
func (f *fakeFS) Remove(string) error                  { return nil }
func (f *fakeFS) RemoveDir(string) error               { return nil }
func (f *fakeFS) Rename(string, string) error          { return nil }
func (f *fakeFS) Close() error                         { return nil }
func (f *fakeFS) Create(p string, r io.Reader) error   { _, err := io.Copy(io.Discard, r); return err }
func (f *fakeFS) Open(p string) (io.ReadCloser, error) {
	if f.slow != nil {
		return io.NopCloser(&blockingReader{f.slow}), nil
	}
	return io.NopCloser(strings.NewReader(f.files[p])), nil
}

type blockingReader struct{ stop chan struct{} }

func (b *blockingReader) Read(p []byte) (int, error) {
	select {
	case <-b.stop:
		return 0, io.EOF
	case <-time.After(5 * time.Millisecond):
		return copy(p, "x"), nil
	}
}

func TestDownloadIgnoresUnsafeNames(t *testing.T) {
	fs := &fakeFS{
		entries: map[string][]remoteEntry{"/box": {{Name: "../escape.txt"}, {Name: "ok.txt", Size: 2}, {Name: `..\win.txt`}}},
		files:   map[string]string{"/box/ok.txt": "ok"},
	}
	a, h := remoteTestApp(t)
	a.dialRemote = func(remoteConfig, *trustStore) (remoteFS, error) { return fs, nil }
	call(t, h, "/api/remote/connect", map[string]any{"protocol": "ftp", "host": "fake"})
	parent := t.TempDir()
	dest := filepath.Join(parent, "inside")
	os.Mkdir(dest, 0o755)
	call(t, h, "/api/transfer", map[string]any{"up": false, "localDir": dest, "remoteDir": "/",
		"items": []map[string]any{{"name": "box", "dir": true}}})
	waitTransfers(t, a)
	if _, err := os.Stat(filepath.Join(dest, "box", "ok.txt")); err != nil {
		t.Fatal("the safe file should be downloaded")
	}
	for _, p := range []string{filepath.Join(dest, "escape.txt"), filepath.Join(parent, "escape.txt"), filepath.Join(dest, "win.txt")} {
		if _, err := os.Stat(p); err == nil {
			t.Fatalf("%s should not exist", p)
		}
	}
	// And names sent straight from the window are checked too.
	b, _ := json.Marshal(map[string]any{"up": false, "localDir": dest, "remoteDir": "/", "items": []map[string]any{{"name": "../x"}}})
	if code, _ := panelCall(t, h, "POST", "/api/transfer", string(b)); code != 400 {
		t.Fatal("an unsafe name from the window should be refused")
	}
}

func TestCancelLeavesNoPartialFile(t *testing.T) {
	stop := make(chan struct{})
	defer close(stop)
	fs := &fakeFS{slow: stop}
	a, h := remoteTestApp(t)
	a.dialRemote = func(remoteConfig, *trustStore) (remoteFS, error) { return fs, nil }
	call(t, h, "/api/remote/connect", map[string]any{"protocol": "ftp", "host": "fake"})
	dest := t.TempDir()
	call(t, h, "/api/transfer", map[string]any{"up": false, "localDir": dest, "remoteDir": "/",
		"items": []map[string]any{{"name": "big.iso", "size": 1 << 30}, {"name": "next.txt"}}})
	time.Sleep(60 * time.Millisecond)
	call(t, h, "/api/transfer/cancel", map[string]any{"id": 0})
	vs := waitTransfers(t, a)
	if vs[0].State != "cancelled" || vs[1].State != "cancelled" {
		t.Fatalf("states: %+v", vs)
	}
	if vs[0].Done == 0 {
		t.Fatal("progress should have been counted before cancelling")
	}
	left, _ := os.ReadDir(dest)
	if len(left) != 0 {
		t.Fatalf("files left behind: %v", left)
	}
}

func TestLocalList(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "zeta"), 0o755)
	os.WriteFile(filepath.Join(dir, "Alpha.txt"), []byte("12345"), 0o644)
	_, h := remoteTestApp(t)
	out := call(t, h, "/api/local/list", map[string]any{"path": dir})
	es := out["entries"].([]any)
	if len(es) != 2 || es[0].(map[string]any)["name"] != "zeta" || es[1].(map[string]any)["size"] != float64(5) {
		t.Fatalf("folders first, then files with sizes: %v", es)
	}
	if out["up"] != filepath.Dir(dir) {
		t.Fatalf("up: %v", out["up"])
	}
	out = call(t, h, "/api/local/list", map[string]any{"path": dir, "name": "zeta"})
	if out["path"] != filepath.Join(dir, "zeta") {
		t.Fatalf("open child: %v", out["path"])
	}
	b, _ := json.Marshal(map[string]any{"path": dir, "name": "../.."})
	if code, _ := panelCall(t, h, "POST", "/api/local/list", string(b)); code != 400 {
		t.Fatal("../ should be refused")
	}
}
