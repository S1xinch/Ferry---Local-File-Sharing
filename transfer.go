package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// A transfer moves one file or folder between this computer and the server.
// Folders are expanded into their contents when their turn comes.
type transfer struct {
	id     int
	up     bool // true: this computer to the server
	name   string
	local  string
	remote string
	dir    bool
	size   int64
	done   atomic.Int64
	state  string // queued, running, done, failed, cancelled
	err    string
	mod    time.Time // the server's modified time, kept on downloads
}

// transferView is what the app window shows for each transfer.
type transferView struct {
	ID    int    `json:"id"`
	Up    bool   `json:"up"`
	Name  string `json:"name"`
	Dir   bool   `json:"dir"`
	Size  int64  `json:"size"`
	Done  int64  `json:"done"`
	State string `json:"state"`
	Error string `json:"error,omitempty"`
}

// transferQueue runs transfers one at a time over its own connection, so
// browsing the server stays quick while big files move.
type transferQueue struct {
	mu      sync.Mutex
	jobs    []*transfer
	nextID  int
	running bool
	cancel  context.CancelFunc
	cfg     *remoteConfig // the server transfers go to; nil when disconnected
	fs      remoteFS
	trust   *trustStore
	dial    func(remoteConfig, *trustStore) (remoteFS, error)
	onDone  func(t *transfer) // called after each finished item, for the Activity list
}

const maxFinishedShown = 200

func (q *transferQueue) views() []transferView {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]transferView, 0, len(q.jobs))
	for _, t := range q.jobs {
		out = append(out, transferView{ID: t.id, Up: t.up, Name: t.name, Dir: t.dir,
			Size: t.size, Done: t.done.Load(), State: t.state, Error: t.err})
	}
	return out
}

// setServer points the queue at a server, or at nothing. Anything waiting
// for the old server is cancelled.
func (q *transferQueue) setServer(c *remoteConfig) {
	q.mu.Lock()
	for _, t := range q.jobs {
		if t.state == "queued" {
			t.state = "cancelled"
		}
	}
	if q.cancel != nil {
		q.cancel()
	}
	q.cfg = c
	fs := q.fs
	q.fs = nil
	q.mu.Unlock()
	if fs != nil {
		go fs.Close()
	}
}

func (q *transferQueue) add(ts []*transfer) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, t := range ts {
		q.nextID++
		t.id, t.state = q.nextID, "queued"
		q.jobs = append(q.jobs, t)
	}
	q.trim()
	if !q.running {
		q.running = true
		go q.work()
	}
}

// trim drops the oldest finished transfers so the list stays readable.
func (q *transferQueue) trim() {
	finished := 0
	for _, t := range q.jobs {
		if t.state != "queued" && t.state != "running" {
			finished++
		}
	}
	if finished <= maxFinishedShown {
		return
	}
	keep := q.jobs[:0]
	for _, t := range q.jobs {
		if finished > maxFinishedShown && t.state != "queued" && t.state != "running" {
			finished--
			continue
		}
		keep = append(keep, t)
	}
	q.jobs = keep
}

// cancelJob stops one transfer, or all of them when id is 0.
func (q *transferQueue) cancelJob(id int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, t := range q.jobs {
		if id != 0 && t.id != id {
			continue
		}
		switch t.state {
		case "queued":
			t.state = "cancelled"
		case "running":
			if q.cancel != nil {
				q.cancel()
			}
		}
	}
}

// clearFinished removes everything that is no longer waiting or running.
func (q *transferQueue) clearFinished() {
	q.mu.Lock()
	defer q.mu.Unlock()
	keep := q.jobs[:0]
	for _, t := range q.jobs {
		if t.state == "queued" || t.state == "running" {
			keep = append(keep, t)
		}
	}
	q.jobs = keep
}

func (q *transferQueue) busy() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, t := range q.jobs {
		if t.state == "queued" || t.state == "running" {
			return true
		}
	}
	return false
}

func (q *transferQueue) work() {
	for {
		q.mu.Lock()
		var t *transfer
		for _, j := range q.jobs {
			if j.state == "queued" {
				t = j
				break
			}
		}
		if t == nil || q.cfg == nil {
			q.running = false
			fs := q.fs
			q.fs = nil
			q.mu.Unlock()
			if fs != nil {
				fs.Close()
			}
			return
		}
		t.state = "running"
		ctx, cancel := context.WithCancel(context.Background())
		q.cancel = cancel
		cfg := *q.cfg
		fs := q.fs
		q.mu.Unlock()

		var children []*transfer
		var err error
		if fs == nil {
			fs, err = q.dial(cfg, q.trust)
		}
		if err == nil {
			children, err = runTransfer(ctx, fs, t)
		}
		stopped := ctx.Err() != nil
		cancel()

		q.mu.Lock()
		q.cancel = nil
		switch {
		case stopped && (err != nil || t.dir):
			t.state, t.err = "cancelled", ""
		case err != nil:
			t.state, t.err = "failed", friendlyTransferError(err)
		default:
			t.state = "done"
		}
		// After any failure the connection may be stuck halfway through a
		// file, so the next transfer starts on a fresh one.
		if err != nil || stopped {
			if fs != nil {
				go fs.Close()
			}
			fs = nil
		}
		if q.cfg != nil && q.cfg.key() == cfg.key() {
			q.fs = fs
		} else if fs != nil {
			go fs.Close()
		}
		if len(children) > 0 && t.state == "done" {
			// Put the folder's contents right after it, in order.
			for _, c := range children {
				q.nextID++
				c.id, c.state = q.nextID, "queued"
			}
			pos := len(q.jobs) // the folder was cleared from the list meanwhile
			for i, j := range q.jobs {
				if j == t {
					pos = i + 1
				}
			}
			rest := append(children, q.jobs[pos:]...)
			q.jobs = append(q.jobs[:pos], rest...)
		}
		q.trim()
		cb := q.onDone
		q.mu.Unlock()
		if cb != nil && !t.dir {
			cb(t)
		}
	}
}

// runTransfer does one item. For a folder it creates the folder on the other
// side and returns its contents as new transfers.
func runTransfer(ctx context.Context, fs remoteFS, t *transfer) ([]*transfer, error) {
	if t.up {
		return upload(ctx, fs, t)
	}
	return download(ctx, fs, t)
}

func upload(ctx context.Context, fs remoteFS, t *transfer) ([]*transfer, error) {
	if t.dir {
		if err := fs.Mkdir(t.remote); err != nil {
			// Fine if it is already there.
			if _, lerr := fs.List(t.remote); lerr != nil {
				return nil, err
			}
		}
		entries, err := os.ReadDir(t.local)
		if err != nil {
			return nil, err
		}
		var out []*transfer
		for _, e := range entries {
			if e.Type()&os.ModeSymlink != 0 || (!e.IsDir() && !e.Type().IsRegular()) {
				continue // links could loop forever, and devices are not files
			}
			c := &transfer{up: true, name: path.Join(t.name, e.Name()), dir: e.IsDir(),
				local: filepath.Join(t.local, e.Name()), remote: path.Join(t.remote, e.Name())}
			if !c.dir {
				if fi, err := e.Info(); err == nil {
					c.size = fi.Size()
				}
			}
			out = append(out, c)
		}
		return out, nil
	}
	f, err := os.Open(t.local)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if fi, err := f.Stat(); err == nil {
		t.size = fi.Size()
	}
	return nil, fs.Create(t.remote, &progressReader{ctx: ctx, r: f, n: &t.done})
}

func download(ctx context.Context, fs remoteFS, t *transfer) ([]*transfer, error) {
	if t.dir {
		if err := os.MkdirAll(t.local, 0o755); err != nil {
			return nil, err
		}
		entries, err := fs.List(t.remote)
		if err != nil {
			return nil, err
		}
		var out []*transfer
		for _, e := range entries {
			if !safeName(e.Name) {
				continue
			}
			out = append(out, &transfer{name: path.Join(t.name, e.Name), dir: e.Dir, size: e.Size, mod: e.Mod,
				local: filepath.Join(t.local, e.Name), remote: path.Join(t.remote, e.Name)})
		}
		return out, nil
	}
	rc, err := fs.Open(t.remote)
	if err != nil {
		return nil, err
	}
	// Write to a hidden file first, so a cancelled download never leaves a
	// half-finished file that looks complete.
	part := filepath.Join(filepath.Dir(t.local), "."+filepath.Base(t.local)+".ferry-part")
	out, err := os.OpenFile(part, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		rc.Close()
		return nil, err
	}
	_, err = io.Copy(out, &progressReader{ctx: ctx, r: rc, n: &t.done})
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if rerr := rc.Close(); err == nil {
		err = rerr
	}
	if err != nil {
		os.Remove(part)
		return nil, err
	}
	if err := os.Rename(part, t.local); err != nil {
		os.Remove(part)
		return nil, err
	}
	if !t.mod.IsZero() {
		os.Chtimes(t.local, t.mod, t.mod)
	}
	return nil, nil
}

// progressReader counts bytes as they pass and stops when cancelled.
type progressReader struct {
	ctx context.Context
	r   io.Reader
	n   *atomic.Int64
}

func (p *progressReader) Read(b []byte) (int, error) {
	if err := p.ctx.Err(); err != nil {
		return 0, err
	}
	n, err := p.r.Read(b)
	p.n.Add(int64(n))
	return n, err
}

func friendlyTransferError(err error) string {
	msg := err.Error()
	switch {
	case errors.Is(err, os.ErrPermission), strings.Contains(msg, "550"), strings.Contains(msg, "permission denied"):
		return "not allowed: " + msg
	case errors.Is(err, os.ErrNotExist):
		return "not found: " + msg
	case isConnLost(err):
		return "the connection dropped: " + msg
	}
	return msg
}
