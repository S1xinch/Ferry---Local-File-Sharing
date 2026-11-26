package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLinkTokenLogsIn(t *testing.T) {
	a := newAuth("123456")
	h := a.require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) }))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/docs/?t="+a.linkToken+"&x=1", nil))
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/docs/?x=1" {
		t.Fatalf("got %d %q", rec.Code, rec.Header().Get("Location"))
	}
	cookie := rec.Result().Cookies()[0]

	req := httptest.NewRequest("GET", "/docs/", nil)
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Body.String() != "ok" {
		t.Fatalf("cookie did not log in: %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/?t=wrong", nil))
	if rec.Code != http.StatusSeeOther || !strings.HasPrefix(rec.Header().Get("Location"), "/__ferry/login?") {
		t.Fatalf("wrong token got through")
	}
}

func TestRecordFailureReportsLockout(t *testing.T) {
	a := newAuth("123456")
	for i := 1; i < perIPTries; i++ {
		if a.recordFailure("10.0.0.9") {
			t.Fatalf("locked out after %d tries", i)
		}
	}
	if !a.recordFailure("10.0.0.9") || a.locked("10.0.0.9") == 0 {
		t.Fatal("not locked out after the last try")
	}
}
