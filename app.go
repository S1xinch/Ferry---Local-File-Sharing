package main

import (
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
	Mode     string `json:"mode"`
	Folder   string `json:"folder"`
	Target   int    `json:"target"`
	Uploads  bool   `json:"uploads"`
	Internet bool   `json:"internet"`
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
}

func configDir() string {
	d, err := os.UserConfigDir()
	if err != nil {
		d = os.TempDir()
	}
	return filepath.Join(d, "ferry")
}

func runApp() error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tokenFile := filepath.Join(dir, "panel-token")

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", panelPort))
	if err != nil {
		// Already running: just show its window again.
		if tok, e := os.ReadFile(tokenFile); e == nil && panelAlive(panelPort) {
			return openBrowser(panelURL(panelPort, strings.TrimSpace(string(tok))))
		}
		if ln, err = net.Listen("tcp", "127.0.0.1:0"); err != nil {
			return err
		}
	}

	a := &app{
		token:    randomHex(24),
		port:     ln.Addr().(*net.TCPAddr).Port,
		dir:      dir,
		quit:     make(chan struct{}),
		lastSeen: time.Now(),
		settings: settings{Mode: "files", Target: 3000, Uploads: true},
	}
	a.loadSettings()
	if err := os.WriteFile(tokenFile, []byte(a.token), 0o600); err != nil {
		return err
	}
	defer os.Remove(tokenFile)

	srv := &http.Server{Handler: a.routes(), ReadHeaderTimeout: 10 * time.Second}
	go srv.Serve(ln)
	defer srv.Close()

	if err := openBrowser(panelURL(a.port, a.token)); err != nil {
		return fmt.Errorf("could not open your browser: %w", err)
	}
	go a.idleWatch()
	<-a.quit

	a.mu.Lock()
	if a.cur != nil {
		a.cur.stop()
	}
	a.mu.Unlock()
	return nil
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
		idle := a.cur == nil && time.Since(a.lastSeen) > idleQuit
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
	mux.HandleFunc("POST /api/quit", a.api(a.quitHandler))
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
	Running        bool     `json:"running"`
	Settings       settings `json:"settings"`
	What           string   `json:"what"`
	PIN            string   `json:"pin"`
	Links          []link   `json:"links"`
	InternetStatus string   `json:"internetStatus"`
	Activity       []string `json:"activity"`
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
	s, st := a.cur, state{Settings: a.settings, Activity: recentActivity(30)}
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
	var in settings
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&in); err != nil {
		return nil, errors.New("bad request")
	}
	in.Folder = strings.Trim(strings.TrimSpace(in.Folder), `"'`)

	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cur != nil {
		return nil, errors.New("already sharing")
	}
	cfg := config{mode: in.Mode, internet: in.Internet, readOnly: !in.Uploads}
	switch in.Mode {
	case "files":
		if in.Folder == "" {
			return nil, errors.New("choose a folder to share first")
		}
		cfg.folder = in.Folder
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
	a.cur, a.settings = s, in
	a.saveSettings()
	if in.Internet {
		go s.startInternet()
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
