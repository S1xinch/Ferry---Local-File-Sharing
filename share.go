package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

// share is one running share: a folder or a website, behind the PIN.
type share struct {
	cfg  config
	auth *auth
	what string
	port int
	srv  *http.Server

	ctx      context.Context
	cancel   context.CancelFunc
	closer   func()
	stopFind func() // answers "who is sharing?" on the local network

	mu        sync.Mutex
	tunnelURL string
	tunnelMsg string
}

func startShare(cfg config) (*share, error) {
	if cfg.pin == "" {
		cfg.pin = randomPIN()
	}
	s := &share{cfg: cfg, auth: newAuth(cfg.pin), closer: func() {}}

	var app http.Handler
	switch cfg.mode {
	case "files":
		abs, err := filepath.Abs(cfg.folder)
		if err != nil {
			return nil, err
		}
		fs, err := newFileShare(abs, cfg.readOnly)
		if err != nil {
			return nil, err
		}
		app, s.what, s.closer = fs, abs, func() { fs.root.Close() }
	case "site":
		if cfg.target < 1 || cfg.target > 65535 {
			return nil, errors.New("the website port must be between 1 and 65535")
		}
		app, s.what = newSiteProxy(cfg.target), fmt.Sprintf("website on port %d", cfg.target)
	default:
		return nil, fmt.Errorf("unknown mode %q", cfg.mode)
	}

	ln, err := listen(cfg.port)
	if err != nil {
		s.closer()
		return nil, fmt.Errorf("could not open a port: %w", err)
	}
	s.port = ln.Addr().(*net.TCPAddr).Port
	s.stopFind = func() {}
	if cfg.mode == "files" {
		s.stopFind = answerDiscovery(s.what, s.port)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/__ferry/login", s.auth.loginHandler)
	mux.HandleFunc("/__ferry/logout", s.auth.logoutHandler)
	mux.Handle("/", s.auth.require(app))
	s.srv = &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		if err := s.srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logf("Server error: %v", err)
		}
	}()

	s.ctx, s.cancel = context.WithCancel(context.Background())
	logf("Started sharing %s", s.what)
	return s, nil
}

// lanURLs are the addresses other devices on the same network can use.
func (s *share) lanURLs() []string {
	var out []string
	for _, ip := range lanIPs() {
		out = append(out, fmt.Sprintf("http://%s:%d", ip, s.port))
	}
	if len(out) == 0 {
		out = append(out, fmt.Sprintf("http://localhost:%d", s.port))
	}
	return out
}

// withToken adds the auto-login token, for QR codes.
func (s *share) withToken(base string) string {
	return base + "/?t=" + s.auth.linkToken
}

func (s *share) tunnel() (url, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tunnelURL, s.tunnelMsg
}

func (s *share) setTunnel(url, msg string) {
	s.mu.Lock()
	s.tunnelURL, s.tunnelMsg = url, msg
	s.mu.Unlock()
}

// startInternet blocks until the internet link is ready or has failed.
// The link closes by itself when the share stops.
func (s *share) startInternet() error {
	s.setTunnel("", "Connecting...")
	t, err := startTunnel(s.ctx, s.port, func(msg string) { s.setTunnel("", msg) })
	if err != nil {
		s.setTunnel("", "Internet link failed: "+err.Error())
		logf("Internet link failed: %v", err)
		return err
	}
	s.setTunnel(t.URL, "")
	logf("Internet link ready: %s", t.URL)
	go func() {
		<-s.ctx.Done()
		t.Close()
	}()
	return nil
}

func (s *share) stop() {
	s.cancel()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s.srv.Shutdown(ctx)
	s.stopFind()
	s.closer()
	logf("Stopped sharing")
}

// listen uses the requested port, or the first free one from 8080 upward.
func listen(port int) (net.Listener, error) {
	if port != 0 {
		return net.Listen("tcp", fmt.Sprintf(":%d", port))
	}
	for p := 8080; p < 8100; p++ {
		if ln, err := net.Listen("tcp", fmt.Sprintf(":%d", p)); err == nil {
			return ln, nil
		}
	}
	return net.Listen("tcp", ":0")
}

// activity keeps recent events (logins, uploads) for the app window.
var activity struct {
	mu    sync.Mutex
	lines []string
}

func logf(format string, a ...any) {
	line := time.Now().Format("15:04:05") + "  " + fmt.Sprintf(format, a...)
	fmt.Println(line)
	activity.mu.Lock()
	activity.lines = append(activity.lines, line)
	if len(activity.lines) > 200 {
		activity.lines = activity.lines[len(activity.lines)-200:]
	}
	activity.mu.Unlock()
}

// recentActivity returns up to n events, newest first.
func recentActivity(n int) []string {
	activity.mu.Lock()
	defer activity.mu.Unlock()
	out := []string{}
	for i := len(activity.lines) - 1; i >= 0 && len(out) < n; i-- {
		out = append(out, activity.lines[i])
	}
	return out
}
