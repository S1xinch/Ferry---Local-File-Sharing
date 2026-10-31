package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// A mirror keeps a local folder matching the folder shared by another
// computer running Ferry. It only ever changes files it has downloaded
// itself: anything else in the folder is left alone. Files removed on the
// other computer are moved into a trash folder here, never deleted.

const (
	mirrorStateFile = ".ferry-mirror.json" // list of files this mirror downloaded
	mirrorTrashDir  = ".ferry-removed"     // where removed files go
	mirrorEvery     = 2 * time.Minute
)

var errSession = errors.New("session expired")

// manifestEntry describes one file in the shared folder.
type manifestEntry struct {
	Path  string `json:"path"` // slash-separated, relative to the shared folder
	Size  int64  `json:"size"`
	Mtime int64  `json:"mtime"` // Unix seconds
}

type mirrorStatus struct {
	Running   bool   `json:"running"`
	Server    string `json:"server"`
	Folder    string `json:"folder"`
	LastSync  string `json:"lastSync"`
	LastError string `json:"lastError"`
	Files     int    `json:"files"`   // files on the other computer at the last check
	Changed   int    `json:"changed"` // downloaded at the last check
	Removed   int    `json:"removed"` // moved to the trash at the last check
}

type mirror struct {
	server string // host:port of the other computer
	pin    string // kept in memory only, never saved to disk
	dest   string

	client *http.Client
	cancel context.CancelFunc

	mu sync.Mutex
	st mirrorStatus
}

func newMirror(server, pin, dest string) *mirror {
	return &mirror{
		server: server,
		pin:    pin,
		dest:   dest,
		st:     mirrorStatus{Running: true, Server: server, Folder: dest},
	}
}

func (m *mirror) url(p string) string { return "http://" + m.server + p }

// login opens a session on the other computer with the PIN.
func (m *mirror) login(ctx context.Context) error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	m.client = &http.Client{
		Jar: jar,
		// Do not follow redirects: a redirect to the PIN page means "not logged in".
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		// Plain LAN or Tailscale traffic: ignore any proxy set on this computer.
		Transport: &http.Transport{Proxy: nil, ResponseHeaderTimeout: 30 * time.Second},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.url("/__ferry/login?next=/"),
		strings.NewReader(url.Values{"pin": {m.pin}}.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach %s, check that Ferry is running there and Tailscale is connected", m.server)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		return errors.New("the other computer did not accept that PIN (if you tried several times, wait a minute)")
	}
	return nil
}

// manifest fetches the list of every file in the other computer's folder.
func (m *mirror) manifest(ctx context.Context) ([]manifestEntry, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.url("/?manifest"), nil)
	if err != nil {
		return nil, err
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusSeeOther || resp.StatusCode == http.StatusUnauthorized {
		return nil, errSession
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the other computer answered %s", resp.Status)
	}
	var list []manifestEntry
	err = json.NewDecoder(io.LimitReader(resp.Body, 256<<20)).Decode(&list)
	return list, err
}

// download copies one file into place. It writes to a temporary name first,
// so an interrupted download never leaves a half-finished file behind.
func (m *mirror) download(ctx context.Context, e manifestEntry, local string) error {
	segs := strings.Split(e.Path, "/")
	for i, s := range segs {
		segs[i] = url.PathEscape(s)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.url("/"+strings.Join(segs, "/")), nil)
	if err != nil {
		return err
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusSeeOther || resp.StatusCode == http.StatusUnauthorized {
		return errSession
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not download %s: %s", e.Path, resp.Status)
	}
	if err := os.MkdirAll(filepath.Dir(local), 0o755); err != nil {
		return err
	}
	tmp := local + ".ferry-part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	// Keep the original modified time, so an unchanged file is recognised next round.
	os.Chtimes(tmp, time.Now(), time.Unix(e.Mtime, 0))
	return os.Rename(tmp, local)
}

// loadMirrorState reads the list of files this mirror previously downloaded.
func loadMirrorState(dest string) map[string]bool {
	set := map[string]bool{}
	if b, err := os.ReadFile(filepath.Join(dest, mirrorStateFile)); err == nil {
		var paths []string
		if json.Unmarshal(b, &paths) == nil {
			for _, p := range paths {
				set[p] = true
			}
		}
	}
	return set
}

func saveMirrorState(dest string, set map[string]bool) error {
	paths := make([]string, 0, len(set))
	for p := range set {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	b, err := json.MarshalIndent(paths, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(dest, mirrorStateFile+".tmp")
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dest, mirrorStateFile))
}

// moveToTrash moves a removed file into a dated folder under the trash.
func (m *mirror) moveToTrash(p, stamp string) error {
	src := filepath.Join(m.dest, filepath.FromSlash(p))
	dst := filepath.Join(m.dest, mirrorTrashDir, stamp, filepath.FromSlash(p))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	err := os.Rename(src, dst)
	if os.IsNotExist(err) {
		return nil // already gone locally, nothing to do
	}
	return err
}

// apply brings the local folder in line with the list from the other computer.
func (m *mirror) apply(ctx context.Context, list []manifestEntry) error {
	state := loadMirrorState(m.dest)
	seen := make(map[string]bool, len(list))
	stamp := time.Now().Format("2006-01-02 15-04-05")
	changed, removed := 0, 0

	for _, e := range list {
		if e.Path == "" || !filepath.IsLocal(filepath.FromSlash(e.Path)) {
			continue // never write outside the folder
		}
		seen[e.Path] = true
		local := filepath.Join(m.dest, filepath.FromSlash(e.Path))
		if info, err := os.Stat(local); err == nil && info.Mode().IsRegular() &&
			info.Size() == e.Size && info.ModTime().Unix() == e.Mtime {
			state[e.Path] = true
			continue
		}
		if err := m.download(ctx, e, local); err != nil {
			m.setCounts(len(list), changed, removed)
			return err
		}
		state[e.Path] = true
		changed++
	}

	for p := range state {
		if seen[p] {
			continue
		}
		if err := m.moveToTrash(p, stamp); err != nil {
			return err
		}
		delete(state, p)
		removed++
	}
	m.setCounts(len(list), changed, removed)
	if changed+removed > 0 {
		logf("Mirror: %d updated, %d moved to %s", changed, removed, mirrorTrashDir)
	}
	return saveMirrorState(m.dest, state)
}

func (m *mirror) setCounts(files, changed, removed int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.st.Files, m.st.Changed, m.st.Removed = files, changed, removed
}

// round does one full check. If the session has expired it logs in again
// once and retries.
func (m *mirror) round(ctx context.Context) error {
	for attempt := 0; attempt < 2; attempt++ {
		if m.client == nil || attempt > 0 {
			if err := m.login(ctx); err != nil {
				return err
			}
		}
		list, err := m.manifest(ctx)
		if errors.Is(err, errSession) {
			continue
		}
		if err != nil {
			return err
		}
		err = m.apply(ctx, list)
		if errors.Is(err, errSession) {
			continue
		}
		return err
	}
	return errors.New("could not keep the session open with the other computer")
}

func (m *mirror) setResult(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err != nil {
		m.st.LastError = err.Error()
		return
	}
	m.st.LastError = ""
	m.st.LastSync = time.Now().Format("2 Jan 2006 15:04")
}

// run keeps checking until stopped.
func (m *mirror) run(ctx context.Context) {
	for {
		m.setResult(m.round(ctx))
		select {
		case <-ctx.Done():
			return
		case <-time.After(mirrorEvery):
		}
	}
}

func (m *mirror) start() {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	go m.run(ctx)
	logf("Receiving %s into %s", m.server, m.dest)
}

func (m *mirror) stop() {
	if m.cancel != nil {
		m.cancel()
	}
	m.mu.Lock()
	m.st.Running = false
	m.mu.Unlock()
	logf("Stopped receiving from %s", m.server)
}

func (m *mirror) snapshot() mirrorStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.st
}
