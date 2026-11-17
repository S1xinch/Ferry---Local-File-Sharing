package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// The "Transfer files" page: connect to an FTP, FTPS or SFTP server, browse
// it next to this computer's files, and copy between them.

const maxRecent = 10

// savedConn is a recent connection. Passwords are never saved.
type savedConn struct {
	Protocol string    `json:"protocol"`
	Host     string    `json:"host"`
	Port     int       `json:"port"`
	User     string    `json:"user"`
	Path     string    `json:"path"` // the folder that was open last time
	Used     time.Time `json:"used"`
}

func (s savedConn) config() remoteConfig {
	return remoteConfig{Protocol: s.Protocol, Host: s.Host, Port: s.Port, User: s.User}
}

type recentView struct {
	savedConn
	Key   string `json:"key"`
	Label string `json:"label"`
}

// remoteHub holds everything for the Transfer files page.
type remoteHub struct {
	trust      *trustStore
	queue      *transferQueue
	recentFile string
	dial       func(remoteConfig, *trustStore) (remoteFS, error)

	mu      sync.Mutex // guards the fields below and use of fs
	cfg     *remoteConfig
	fs      remoteFS
	cwd     string
	recents []savedConn
}

func (a *app) remote() *remoteHub {
	a.hubOnce.Do(func() {
		h := &remoteHub{
			trust:      &trustStore{path: filepath.Join(a.dir, "trusted-servers.json")},
			recentFile: filepath.Join(a.dir, "connections.json"),
			dial:       dialRemote,
		}
		if a.dialRemote != nil {
			h.dial = a.dialRemote
		}
		h.queue = &transferQueue{trust: h.trust, dial: h.dial, onDone: func(t *transfer) {
			if t.state == "failed" {
				verb := "Download"
				if t.up {
					verb = "Upload"
				}
				logf("%s of %s failed: %s", verb, t.name, t.err)
			}
		}}
		if b, err := os.ReadFile(h.recentFile); err == nil {
			json.Unmarshal(b, &h.recents)
		}
		a.hub = h
	})
	return a.hub
}

// remoteState is the part of the app window's state for this page.
type remoteState struct {
	Connected bool           `json:"connected"`
	Label     string         `json:"label,omitempty"`
	Protocol  string         `json:"protocol,omitempty"`
	Path      string         `json:"path,omitempty"`
	Transfers []transferView `json:"transfers"`
	Recent    []recentView   `json:"recent"`
	Trusted   []string       `json:"trusted"`
}

func (h *remoteHub) state() remoteState {
	h.mu.Lock()
	st := remoteState{Recent: []recentView{}}
	if h.cfg != nil {
		st.Connected, st.Label, st.Protocol, st.Path = true, h.cfg.label(), h.cfg.Protocol, h.cwd
	}
	for _, r := range h.recents {
		c := r.config()
		st.Recent = append(st.Recent, recentView{savedConn: r, Key: c.key(), Label: c.label()})
	}
	h.mu.Unlock()
	st.Transfers = h.queue.views()
	st.Trusted = h.trust.list()
	if st.Trusted == nil {
		st.Trusted = []string{}
	}
	return st
}

// rememberLocked moves a connection to the top of the recent list. h.mu is held.
func (h *remoteHub) rememberLocked(c remoteConfig, dir string) {
	key := c.key()
	entry := savedConn{Protocol: c.Protocol, Host: c.Host, Port: c.Port, User: c.User, Path: dir, Used: time.Now()}
	out := []savedConn{entry}
	for _, r := range h.recents {
		if r.config().key() != key {
			out = append(out, r)
		}
	}
	if len(out) > maxRecent {
		out = out[:maxRecent]
	}
	h.recents = out
	writeJSON(h.recentFile, h.recents)
}

func (h *remoteHub) savedPath(key string) string {
	for _, r := range h.recents {
		if r.config().key() == key {
			return r.Path
		}
	}
	return ""
}

// use runs fn on the browsing connection, reconnecting once if it dropped
// (servers close idle FTP connections after a few minutes).
func (h *remoteHub) use(fn func(fs remoteFS) error) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cfg == nil {
		return errors.New("not connected to a server")
	}
	if h.fs == nil {
		fs, err := h.dial(*h.cfg, h.trust)
		if err != nil {
			return err
		}
		h.fs = fs
	}
	err := fn(h.fs)
	if !isConnLost(err) {
		return err
	}
	h.fs.Close()
	h.fs = nil
	fs, derr := h.dial(*h.cfg, h.trust)
	if derr != nil {
		return err
	}
	h.fs = fs
	return fn(fs)
}

func decode(r *http.Request, v any) error {
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(v); err != nil {
		return errors.New("bad request")
	}
	return nil
}

type connectReply struct {
	Path        string `json:"path,omitempty"`
	NeedsTrust  bool   `json:"needsTrust,omitempty"`
	Kind        string `json:"kind,omitempty"`
	Host        string `json:"host,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

func (a *app) remoteConnect(r *http.Request) (any, error) {
	var in remoteConfig
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	c, err := in.clean()
	if err != nil {
		return nil, err
	}
	h := a.remote()
	fs, err := h.dial(c, h.trust)
	if err != nil {
		var te *trustError
		if errors.As(err, &te) && !te.Changed {
			return connectReply{NeedsTrust: true, Kind: te.Kind, Host: te.Host, Fingerprint: te.Fingerprint}, nil
		}
		logf("Could not connect to %s: %v", c.label(), err)
		return nil, err
	}
	home, err := fs.Home()
	if err != nil || home == "" {
		home = "/"
	}
	start := home
	h.mu.Lock()
	if p := h.savedPath(c.key()); p != "" {
		if _, err := fs.List(p); err == nil {
			start = p
		}
	}
	old := h.fs
	c.Trust = ""
	h.cfg, h.fs, h.cwd = &c, fs, start
	h.rememberLocked(c, start)
	h.mu.Unlock()
	if old != nil {
		go old.Close()
	}
	h.queue.setServer(&c)
	logf("Connected to %s", c.label())
	return connectReply{Path: start}, nil
}

func (a *app) remoteDisconnect(r *http.Request) (any, error) {
	h := a.remote()
	h.queue.setServer(nil)
	h.mu.Lock()
	fs, c := h.fs, h.cfg
	h.fs, h.cfg, h.cwd = nil, nil, ""
	h.mu.Unlock()
	if fs != nil {
		go fs.Close()
	}
	if c != nil {
		logf("Disconnected from %s", c.label())
	}
	return map[string]bool{"ok": true}, nil
}

type listing struct {
	Path    string        `json:"path"`
	Up      string        `json:"up"` // the folder above, "" at the top
	Entries []remoteEntry `json:"entries"`
}

func sortEntries(es []remoteEntry) {
	sort.Slice(es, func(i, j int) bool {
		if es[i].Dir != es[j].Dir {
			return es[i].Dir
		}
		return strings.ToLower(es[i].Name) < strings.ToLower(es[j].Name)
	})
}

func (a *app) remoteList(r *http.Request) (any, error) {
	var in struct {
		Path string `json:"path"`
	}
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	h := a.remote()
	h.mu.Lock()
	p := in.Path
	if p == "" {
		p = h.cwd
	}
	h.mu.Unlock()
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	p = path.Clean(p)
	var entries []remoteEntry
	err := h.use(func(fs remoteFS) error {
		var err error
		entries, err = fs.List(p)
		return err
	})
	if err != nil {
		return nil, friendlyRemoteError(err)
	}
	if entries == nil {
		entries = []remoteEntry{}
	}
	sortEntries(entries)
	h.mu.Lock()
	if h.cfg != nil && h.cwd != p {
		h.cwd = p
		h.rememberLocked(*h.cfg, p)
	}
	h.mu.Unlock()
	up := path.Dir(p)
	if up == p {
		up = ""
	}
	return listing{Path: p, Up: up, Entries: entries}, nil
}

func friendlyRemoteError(err error) error {
	msg := err.Error()
	switch {
	case errors.Is(err, os.ErrNotExist), strings.Contains(msg, "no such file"), strings.Contains(msg, "550"):
		return errors.New("that folder does not exist on the server, or you are not allowed to open it")
	case errors.Is(err, os.ErrPermission), strings.Contains(msg, "permission denied"):
		return errors.New("the server does not allow that")
	case isConnLost(err):
		return errors.New("lost the connection to the server. Connect again")
	}
	return err
}

// remotePath checks a path sent by the window.
func remotePath(p string) (string, error) {
	if p == "" || strings.ContainsRune(p, 0) {
		return "", errors.New("bad path")
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	p = path.Clean(p)
	if p == "/" {
		return "", errors.New("that is the top folder of the server")
	}
	return p, nil
}

func (a *app) remoteMkdir(r *http.Request) (any, error) {
	var in struct {
		Dir  string `json:"dir"`
		Name string `json:"name"`
	}
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(in.Name)
	if !safeName(name) {
		return nil, errors.New("folder names cannot contain / or \\")
	}
	p := path.Join("/", in.Dir, name)
	if err := a.remote().use(func(fs remoteFS) error { return fs.Mkdir(p) }); err != nil {
		return nil, friendlyRemoteError(err)
	}
	return map[string]bool{"ok": true}, nil
}

func (a *app) remoteRename(r *http.Request) (any, error) {
	var in struct {
		Dir  string `json:"dir"`
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	to := strings.TrimSpace(in.To)
	if !safeName(in.From) || !safeName(to) {
		return nil, errors.New("names cannot contain / or \\")
	}
	from, dest := path.Join("/", in.Dir, in.From), path.Join("/", in.Dir, to)
	if err := a.remote().use(func(fs remoteFS) error { return fs.Rename(from, dest) }); err != nil {
		return nil, friendlyRemoteError(err)
	}
	return map[string]bool{"ok": true}, nil
}

func (a *app) remoteDelete(r *http.Request) (any, error) {
	var in struct {
		Dir   string        `json:"dir"`
		Items []remoteEntry `json:"items"`
	}
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	h := a.remote()
	for _, it := range in.Items {
		if !safeName(it.Name) {
			return nil, errors.New("bad name")
		}
		p, err := remotePath(path.Join("/", in.Dir, it.Name))
		if err != nil {
			return nil, err
		}
		err = h.use(func(fs remoteFS) error {
			if it.Dir {
				return removeAllRemote(fs, p, 0)
			}
			return fs.Remove(p)
		})
		if err != nil {
			return nil, friendlyRemoteError(err)
		}
		logf("Deleted %s from the server", p)
	}
	return map[string]bool{"ok": true}, nil
}

// removeAllRemote deletes a folder and everything in it.
func removeAllRemote(fs remoteFS, p string, depth int) error {
	if depth > 64 {
		return errors.New("folders are nested too deeply to delete")
	}
	entries, err := fs.List(p)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !safeName(e.Name) {
			continue
		}
		child := path.Join(p, e.Name)
		if e.Dir {
			err = removeAllRemote(fs, child, depth+1)
		} else {
			err = fs.Remove(child)
		}
		if err != nil {
			return err
		}
	}
	return fs.RemoveDir(p)
}

func (a *app) remoteTransfer(r *http.Request) (any, error) {
	var in struct {
		Up        bool          `json:"up"`
		LocalDir  string        `json:"localDir"`
		RemoteDir string        `json:"remoteDir"`
		Items     []remoteEntry `json:"items"`
	}
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	h := a.remote()
	h.mu.Lock()
	connected := h.cfg != nil
	h.mu.Unlock()
	if !connected {
		return nil, errors.New("not connected to a server")
	}
	if len(in.Items) == 0 {
		return nil, errors.New("choose some files first")
	}
	fi, err := os.Stat(in.LocalDir)
	if err != nil || !fi.IsDir() || !filepath.IsAbs(in.LocalDir) {
		return nil, errors.New("that folder on this computer does not exist")
	}
	rdir := path.Clean(path.Join("/", in.RemoteDir))
	var ts []*transfer
	for _, it := range in.Items {
		if !safeName(it.Name) {
			return nil, errors.New("bad name: " + it.Name)
		}
		ts = append(ts, &transfer{up: in.Up, name: it.Name, dir: it.Dir, size: it.Size, mod: it.Mod,
			local: filepath.Join(in.LocalDir, it.Name), remote: path.Join(rdir, it.Name)})
	}
	h.queue.add(ts)
	return map[string]bool{"ok": true}, nil
}

func (a *app) transferCancel(r *http.Request) (any, error) {
	var in struct {
		ID int `json:"id"`
	}
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	a.remote().queue.cancelJob(in.ID)
	return map[string]bool{"ok": true}, nil
}

func (a *app) transferClear(r *http.Request) (any, error) {
	a.remote().queue.clearFinished()
	return map[string]bool{"ok": true}, nil
}

func (a *app) remoteForget(r *http.Request) (any, error) {
	var in struct {
		Key string `json:"key"`
	}
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	h := a.remote()
	h.mu.Lock()
	out := h.recents[:0]
	for _, s := range h.recents {
		if s.config().key() != in.Key {
			out = append(out, s)
		}
	}
	h.recents = out
	writeJSON(h.recentFile, h.recents)
	h.mu.Unlock()
	return map[string]bool{"ok": true}, nil
}

func (a *app) untrust(r *http.Request) (any, error) {
	var in struct {
		ID string `json:"id"`
	}
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	a.remote().trust.remove(in.ID)
	return map[string]bool{"ok": true}, nil
}

// --- this computer's files

const drivesPath = ":drives" // Windows: the list of drives, above C:\ and the rest

func (a *app) localList(r *http.Request) (any, error) {
	var in struct {
		Path string `json:"path"`
		Name string `json:"name"` // open this folder inside Path
	}
	if err := decode(r, &in); err != nil {
		return nil, err
	}
	p := in.Path
	if in.Name != "" {
		if in.Path != drivesPath && !safeName(in.Name) {
			return nil, errors.New("bad name")
		}
		p = localJoin(in.Path, in.Name)
	}
	if p == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = string(filepath.Separator)
		}
		p = home
	}
	if p == drivesPath && runtime.GOOS == "windows" {
		var out []remoteEntry
		for c := 'A'; c <= 'Z'; c++ {
			d := string(c) + `:\`
			if _, err := os.Stat(d); err == nil {
				out = append(out, remoteEntry{Name: d, Dir: true})
			}
		}
		return listing{Path: drivesPath, Entries: out}, nil
	}
	if !filepath.IsAbs(p) {
		return nil, errors.New("type a full path, like " + exampleLocalPath())
	}
	p = filepath.Clean(p)
	des, err := os.ReadDir(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.New("that folder does not exist on this computer")
		}
		if errors.Is(err, os.ErrPermission) {
			return nil, errors.New("this computer does not allow Ferry to open that folder")
		}
		return nil, err
	}
	out := make([]remoteEntry, 0, len(des))
	for _, d := range des {
		info, err := d.Info()
		if err != nil {
			continue
		}
		dir := d.IsDir()
		if d.Type()&os.ModeSymlink != 0 {
			if t, err := os.Stat(filepath.Join(p, d.Name())); err == nil {
				info, dir = t, t.IsDir()
			}
		}
		e := remoteEntry{Name: d.Name(), Dir: dir, Mod: info.ModTime()}
		if !dir {
			e.Size = info.Size()
		}
		out = append(out, e)
	}
	sortEntries(out)
	up := filepath.Dir(p)
	if up == p {
		up = ""
		if runtime.GOOS == "windows" {
			up = drivesPath
		}
	}
	return listing{Path: p, Up: up, Entries: out}, nil
}

func exampleLocalPath() string {
	if runtime.GOOS == "windows" {
		return `C:\Users\you\Documents`
	}
	return "/Users/you/Documents"
}

// localJoin builds the path of an item in a folder on this computer; on
// Windows the drives list holds whole drives rather than names.
func localJoin(dir, name string) string {
	if dir == drivesPath {
		return name
	}
	return filepath.Join(dir, name)
}
