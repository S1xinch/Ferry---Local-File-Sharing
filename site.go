package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// newSiteProxy forwards visitors to a website already running on this PC,
// for example a dev server on port 3000. WebSockets pass through too.
func newSiteProxy(port int) http.Handler {
	target, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", port))
	return &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(target) // also sets Host to localhost, which dev servers expect
			pr.SetXForwarded()
			stripCookie(pr.Out)
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			w.WriteHeader(http.StatusBadGateway)
			render(w, "offline", struct{ Port int }{port})
		},
	}
}

// stripCookie hides Ferry's own cookies from the website being shared.
func stripCookie(r *http.Request) {
	var kept []string
	for _, c := range r.Cookies() {
		if c.Name != cookieName && c.Name != panelCookie {
			kept = append(kept, c.Name+"="+c.Value)
		}
	}
	r.Header.Del("Cookie")
	if len(kept) > 0 {
		r.Header.Set("Cookie", strings.Join(kept, "; "))
	}
}
