package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPanelRequiresTokenHostAndHeader(t *testing.T) {
	a := &app{token: "secret", port: 47615, quit: make(chan struct{}), settings: settings{Mode: "files"}}
	h := a.routes()
	do := func(method, target, host string, cookie, header bool) int {
		r := httptest.NewRequest(method, target, strings.NewReader("{}"))
		r.Host = host
		if cookie {
			r.AddCookie(&http.Cookie{Name: panelCookie, Value: "secret"})
		}
		if header {
			r.Header.Set("X-Ferry", "1")
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Code
	}
	cases := []struct {
		name           string
		method, target string
		host           string
		cookie, header bool
		want           int
	}{
		{"state ok", "GET", "/api/state", "127.0.0.1:47615", true, true, 200},
		{"no cookie", "GET", "/api/state", "127.0.0.1:47615", false, true, 403},
		{"no header", "POST", "/api/stop", "127.0.0.1:47615", true, false, 403},
		{"other host", "GET", "/api/state", "evil.example:47615", true, true, 403},
		{"page locked", "GET", "/", "localhost:47615", false, false, 403},
		{"token sets cookie", "GET", "/?k=secret", "127.0.0.1:47615", false, false, 303},
		{"start needs folder", "POST", "/api/start", "127.0.0.1:47615", true, true, 400},
	}
	for _, c := range cases {
		if got := do(c.method, c.target, c.host, c.cookie, c.header); got != c.want {
			t.Errorf("%s: got %d, want %d", c.name, got, c.want)
		}
	}
}
