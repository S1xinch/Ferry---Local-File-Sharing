package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const cookieName = "ferry_session"

// Limits on wrong PIN guesses. A 6 digit PIN has a million options, so
// these keep guessing hopeless even over the internet.
const (
	perIPTries    = 5
	perIPLockout  = time.Minute
	globalTries   = 30
	globalWindow  = 10 * time.Minute
	globalLockout = 10 * time.Minute
)

type auth struct {
	pin       string
	session   string // value stored in the browser cookie after login
	linkToken string // goes in QR codes so scanning skips the PIN

	mu            sync.Mutex
	failures      map[string]int
	lockedUntil   map[string]time.Time
	globalFails   []time.Time
	globalLockEnd time.Time
}

func newAuth(pin string) *auth {
	return &auth{
		pin:         pin,
		session:     randomHex(32),
		linkToken:   randomHex(16),
		failures:    map[string]int{},
		lockedUntil: map[string]time.Time{},
	}
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func randomPIN() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%06d", n.Int64())
}

func equal(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// clientIP finds who is connecting. Behind the internet tunnel every request
// arrives from this PC, so the tunnel's header is trusted only in that case.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		if cf := r.Header.Get("Cf-Connecting-Ip"); cf != "" {
			return cf
		}
	}
	return host
}

func (a *auth) loggedIn(r *http.Request) bool {
	c, err := r.Cookie(cookieName)
	return err == nil && equal(c.Value, a.session)
}

func (a *auth) setCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    a.session,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		MaxAge:   7 * 24 * 3600,
	})
}

// require lets logged in visitors through and sends everyone else to the PIN page.
func (a *auth) require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if t := r.URL.Query().Get("t"); t != "" && equal(t, a.linkToken) {
			a.setCookie(w, r)
			q := r.URL.Query()
			q.Del("t")
			u := *r.URL
			u.RawQuery = q.Encode()
			http.Redirect(w, r, u.RequestURI(), http.StatusSeeOther)
			return
		}
		if a.loggedIn(r) {
			next.ServeHTTP(w, r)
			return
		}
		http.Redirect(w, r, "/__ferry/login?next="+url.QueryEscape(r.URL.RequestURI()), http.StatusSeeOther)
	})
}

// locked reports how long until this visitor may try again (0 if they can try now).
func (a *auth) locked(ip string) time.Duration {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	if d := a.globalLockEnd.Sub(now); d > 0 {
		return d
	}
	if d := a.lockedUntil[ip].Sub(now); d > 0 {
		return d
	}
	return 0
}

func (a *auth) recordFailure(ip string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	a.failures[ip]++
	if a.failures[ip] >= perIPTries {
		a.failures[ip] = 0
		a.lockedUntil[ip] = now.Add(perIPLockout)
	}
	kept := a.globalFails[:0]
	for _, t := range a.globalFails {
		if now.Sub(t) < globalWindow {
			kept = append(kept, t)
		}
	}
	a.globalFails = append(kept, now)
	if len(a.globalFails) >= globalTries {
		a.globalFails = nil
		a.globalLockEnd = now.Add(globalLockout)
		logf("Too many wrong PINs, logins paused for 10 minutes")
	}
}

func (a *auth) loginHandler(w http.ResponseWriter, r *http.Request) {
	next := r.URL.Query().Get("next")
	if !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
		next = "/"
	}
	data := loginPage{Next: next}
	ip := clientIP(r)

	if r.Method == http.MethodPost {
		if wait := a.locked(ip); wait > 0 {
			data.Error = fmt.Sprintf("Too many wrong tries. Wait %d seconds.", int(wait.Seconds())+1)
		} else if equal(strings.TrimSpace(r.FormValue("pin")), a.pin) {
			a.mu.Lock()
			delete(a.failures, ip)
			a.mu.Unlock()
			a.setCookie(w, r)
			logf("%s logged in", ip)
			http.Redirect(w, r, next, http.StatusSeeOther)
			return
		} else {
			a.recordFailure(ip)
			data.Error = "That PIN is not right."
		}
		w.WriteHeader(http.StatusUnauthorized)
	}
	render(w, "login", data)
}

func (a *auth) logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/__ferry/login", http.StatusSeeOther)
}
