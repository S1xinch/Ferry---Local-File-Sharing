package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

// The app window is a page served only to this computer, on a fixed port so
// opening the app a second time finds the one already running.
const (
	panelPort   = 47615
	panelCookie = "ferry_panel"
	idleQuit    = 3 * time.Minute // quit when nothing is shared and no window is open
)

type settings struct {
	Mode         string `json:"mode"`
	Folder       string `json:"folder"`
	Target       int    `json:"target"`
	Uploads      bool   `json:"uploads"`
	Internet     bool   `json:"internet"`
	KeepPIN      bool   `json:"keepPin"`      // reuse the same PIN when sharing starts again
	PIN          string `json:"pin"`          // saved only while KeepPIN is on
	Server       string `json:"server"`       // receive mode: the other computer's address
	MirrorFolder string `json:"mirrorFolder"` // receive mode: the folder kept up to date
	Prefs        prefs  `json:"prefs"`        // set on the Settings page
}

// prefs are the choices on the Settings page. They are kept when sharing
// starts, unlike the rest of settings which come from the sharing form.
type prefs struct {
	AutoUpdate bool `json:"autoUpdate"` // check for updates on open and every few hours
	Hidden     bool `json:"hidden"`     // do not answer "who is sharing?" on the Wi-Fi
	Port       int  `json:"port"`       // port to share on, 0 for automatic
}

// startRequest is what the app window sends with Start. The receive PIN is
// used once to log in and is never saved.
type startRequest struct {
	settings
	ReceivePIN string `json:"receivePin"`
}

type app struct {
	token string
	port  int
	dir   string
	quit  chan struct{}
	once  sync.Once

	mu       sync.Mutex
	cur      *share
	lastSeen time.Time
	settings settings
	update   updateState
	mir      *mirror // set while receiving from another computer
	staged   *stagedUpdate
	relaunch string // set when an update is installed; started after this copy closes
}

func configDir() string {
	d, err := os.UserConfigDir()
	if err != nil {
		d = os.TempDir()
	}
	return filepath.Join(d, "ferry")
}

// runApp opens the app window. If an update was installed, it returns the
// program to start once this copy has closed.
func runApp() (string, error) {
	dir := configDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	tokenFile := filepath.Join(dir, "panel-token")

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", panelPort))
	if err != nil {
		// Already running: just show its window again.
		if tok, e := os.ReadFile(tokenFile); e == nil && panelAlive(panelPort) {
			return "", openWindow(panelURL(panelPort, strings.TrimSpace(string(tok))))
		}
		if ln, err = net.Listen("tcp", "127.0.0.1:0"); err != nil {
			return "", err
		}
	}

	a := &app{
		token:    randomHex(24),
		port:     ln.Addr().(*net.TCPAddr).Port,
		dir:      dir,
		quit:     make(chan struct{}),
		lastSeen: time.Now(),
		settings: settings{Mode: "files", Target: 3000, Uploads: true, Prefs: prefs{AutoUpdate: true}},
	}
	a.loadSettings()
	cleanupUpdateLeftovers()
	if err := os.WriteFile(tokenFile, []byte(a.token), 0o600); err != nil {
		return "", err
	}
	defer os.Remove(tokenFile)

	srv := &http.Server{Handler: a.routes(), ReadHeaderTimeout: 10 * time.Second}
	go srv.Serve(ln)
	defer srv.Close()

	if err := openWindow(panelURL(a.port, a.token)); err != nil {
		return "", fmt.Errorf("could not open your browser: %w", err)
	}
	go a.idleWatch()
	go a.updateLoop()
	<-a.quit

	a.mu.Lock()
	if a.cur != nil {
		a.cur.stop()
	}
	relaunch := a.relaunch
	if a.mir != nil {
		a.mir.stop()
	}
	a.mu.Unlock()
	// The deferred cleanup (port, token file) runs before main starts the new copy.
	return relaunch, nil
}

func panelURL(port int, token string) string {
	return fmt.Sprintf("http://127.0.0.1:%d/?k=%s", port, token)
}

func panelAlive(port int) bool {
	c := http.Client{Timeout: time.Second}
	resp, err := c.Get(fmt.Sprintf("http://127.0.0.1:%d/api/ping", port))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 16))
	return string(b) == "ferry"
}

func (a *app) idleWatch() {
	for range time.Tick(10 * time.Second) {
		a.mu.Lock()
		idle := a.cur == nil && a.mir == nil && time.Since(a.lastSeen) > idleQuit
		a.mu.Unlock()
		if idle {
			a.stopApp()
			return
		}
	}
}

func (a *app) stopApp() { a.once.Do(func() { close(a.quit) }) }

func (a *app) loadSettings() {
	b, err := os.ReadFile(filepath.Join(a.dir, "settings.json"))
	if err == nil {
		json.Unmarshal(b, &a.settings)
	}
}

func (a *app) saveSettings() {
	b, _ := json.MarshalIndent(a.settings, "", "  ")
	os.WriteFile(filepath.Join(a.dir, "settings.json"), b, 0o600)
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/ping", func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "ferry") })
	mux.HandleFunc("GET /{$}", a.page)
	mux.HandleFunc("GET /api/state", a.api(a.state))
	mux.HandleFunc("POST /api/pick", a.api(a.pick))
	mux.HandleFunc("POST /api/start", a.api(a.start))
	mux.HandleFunc("POST /api/stop", a.api(a.stop))
	mux.HandleFunc("POST /api/mirror/stop", a.api(a.stopMirror))
	mux.HandleFunc("POST /api/discover", a.api(a.discover))
	mux.HandleFunc("POST /api/quit", a.api(a.quitHandler))
	mux.HandleFunc("POST /api/update", a.api(a.updateNow))
	mux.HandleFunc("POST /api/update/check", a.api(a.checkNow))
	mux.HandleFunc("POST /api/prefs", a.api(a.savePrefs))
	return a.guard(mux)
}

// guard only answers requests addressed to this computer by name, so other
// websites cannot reach the app window through tricks with DNS.
func (a *app) guard(next http.Handler) http.Handler {
	ok := map[string]bool{
		fmt.Sprintf("127.0.0.1:%d", a.port): true,
		fmt.Sprintf("localhost:%d", a.port): true,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !ok[r.Host] {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func (a *app) hasCookie(r *http.Request) bool {
	c, err := r.Cookie(panelCookie)
	return err == nil && equal(c.Value, a.token)
}

func (a *app) page(w http.ResponseWriter, r *http.Request) {
	if k := r.URL.Query().Get("k"); k != "" {
		if equal(k, a.token) {
			http.SetCookie(w, &http.Cookie{Name: panelCookie, Value: a.token, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode})
		}
		// The app was just opened: play the intro once.
		http.Redirect(w, r, "/?intro=1", http.StatusSeeOther)
		return
	}
	if !a.hasCookie(r) {
		w.WriteHeader(http.StatusForbidden)
		render(w, "panel-locked", nil)
		return
	}
	render(w, "panel", struct{ Version string }{version})
}

type apiFunc func(r *http.Request) (any, error)

// api checks the request came from the app window and returns JSON.
// The custom header cannot be sent by other websites without permission.
func (a *app) api(fn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !a.hasCookie(r) || r.Header.Get("X-Ferry") != "1" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		a.mu.Lock()
		a.lastSeen = time.Now()
		a.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		out, err := fn(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(out)
	}
}

type link struct {
	Label string `json:"label"`
	URL   string `json:"url"`
	QR    string `json:"qr"`
}

type state struct {
	Running        bool          `json:"running"`
	Settings       settings      `json:"settings"`
	What           string        `json:"what"`
	PIN            string        `json:"pin"`
	Links          []link        `json:"links"`
	InternetStatus string        `json:"internetStatus"`
	Activity       []string      `json:"activity"`
	Update         updateState   `json:"update"`
	Mirror         *mirrorStatus `json:"mirror,omitempty"`
}

func qrDataURL(text string) string {
	png, err := qrcode.Encode(text, qrcode.Medium, 320)
	if err != nil {
		return ""
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
}

func (a *app) state(r *http.Request) (any, error) {
	a.mu.Lock()
	s, st := a.cur, state{Settings: a.settings, Activity: recentActivity(30), Update: a.update}
	if a.mir != nil {
		snap := a.mir.snapshot()
		st.Mirror = &snap
	}
	a.mu.Unlock()
	if s == nil {
		return st, nil
	}
	st.Running, st.What, st.PIN = true, s.what, s.cfg.pin
	for _, u := range s.lanURLs() {
		st.Links = append(st.Links, link{Label: "Same Wi-Fi", URL: u, QR: qrDataURL(s.withToken(u))})
	}
	if s.cfg.internet {
		u, msg := s.tunnel()
		if u != "" {
			st.Links = append(st.Links, link{Label: "Anywhere", URL: u, QR: qrDataURL(s.withToken(u))})
		}
		st.InternetStatus = msg
	}
	return st, nil
}

func (a *app) pick(r *http.Request) (any, error) {
	p, err := pickFolder()
	if err != nil {
		return nil, err
	}
	return map[string]string{"folder": p}, nil
}

func (a *app) start(r *http.Request) (any, error) {
	var req startRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&req); err != nil {
		return nil, errors.New("bad request")
	}
	in := req.settings
	in.Folder = strings.Trim(strings.TrimSpace(in.Folder), `"'`)
	in.MirrorFolder = strings.Trim(strings.TrimSpace(in.MirrorFolder), `"'`)
	in.Server = strings.TrimSpace(in.Server)
	if in.Mode == "receive" {
		return a.startReceive(in, strings.TrimSpace(req.ReceivePIN))
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cur != nil {
		return nil, errors.New("already sharing")
	}
	// Keep the receive settings and the Settings page choices when sharing.
	in.Server, in.MirrorFolder, in.PIN, in.Prefs = a.settings.Server, a.settings.MirrorFolder, "", a.settings.Prefs
	cfg := config{mode: in.Mode, internet: in.Internet, readOnly: !in.Uploads,
		port: in.Prefs.Port, hidden: in.Prefs.Hidden}
	switch in.Mode {
	case "files":
		if in.Folder == "" {
			return nil, errors.New("choose a folder to share first")
		}
		cfg.folder = in.Folder
		if in.KeepPIN && a.settings.PIN != "" {
			cfg.pin = a.settings.PIN
		}
	case "site":
		if in.Target == a.port {
			return nil, errors.New("that port is Ferry itself, pick your website's port")
		}
		cfg.target = in.Target
	default:
		return nil, errors.New("choose what to share")
	}
	s, err := startShare(cfg)
	if err != nil {
		return nil, err
	}
	if in.KeepPIN {
		in.PIN = s.cfg.pin
	}
	a.cur, a.settings = s, in
	a.saveSettings()
	if in.Internet {
		go s.startInternet()
	}
	return map[string]bool{"ok": true}, nil
}

// startReceive logs in to another computer and starts keeping the folder up to date.
func (a *app) startReceive(in settings, pin string) (any, error) {
	if in.Server == "" {
		return nil, errors.New("type the address of the other computer first")
	}
	if pin == "" {
		return nil, errors.New("type the PIN shown on the other computer")
	}
	if in.MirrorFolder == "" {
		return nil, errors.New("choose the folder to keep up to date")
	}
	if err := os.MkdirAll(in.MirrorFolder, 0o755); err != nil {
		return nil, fmt.Errorf("cannot use that folder: %w", err)
	}
	m := newMirror(in.Server, pin, in.MirrorFolder)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := m.login(ctx); err != nil {
		return nil, err
	}

	a.mu.Lock()
	if a.mir != nil {
		a.mir.stop()
	}
	a.mir = m
	a.settings.Mode, a.settings.Server, a.settings.MirrorFolder = "receive", in.Server, in.MirrorFolder
	a.saveSettings()
	a.mu.Unlock()
	m.start()
	return map[string]bool{"ok": true}, nil
}

func (a *app) stopMirror(r *http.Request) (any, error) {
	a.mu.Lock()
	m := a.mir
	a.mir = nil
	a.mu.Unlock()
	if m != nil {
		m.stop()
	}
	return map[string]bool{"ok": true}, nil
}

func (a *app) stop(r *http.Request) (any, error) {
	a.mu.Lock()
	s := a.cur
	a.cur = nil
	a.mu.Unlock()
	if s != nil {
		s.stop()
	}
	return map[string]bool{"ok": true}, nil
}

func (a *app) quitHandler(r *http.Request) (any, error) {
	go func() {
		time.Sleep(200 * time.Millisecond)
		a.stopApp()
	}()
	return map[string]bool{"ok": true}, nil
}

// savePrefs stores the Settings page. A new port is used the next time
// sharing starts.
func (a *app) savePrefs(r *http.Request) (any, error) {
	var p prefs
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<12)).Decode(&p); err != nil {
		return nil, errors.New("bad request")
	}
	if p.Port != 0 && (p.Port < 1024 || p.Port > 65535) {
		return nil, errors.New("the port must be between 1024 and 65535, or empty for automatic")
	}
	if p.Port != 0 && p.Port == a.port {
		return nil, errors.New("that port is used by Ferry itself, pick another one")
	}
	a.mu.Lock()
	a.settings.Prefs = p
	a.saveSettings()
	a.mu.Unlock()
	return map[string]bool{"ok": true}, nil
}
