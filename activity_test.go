package main

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	uaIPhone  = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1"
	uaWindows = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36"
)

func TestDeviceName(t *testing.T) {
	cases := map[string]string{
		uaIPhone:  "iPhone",
		uaWindows: "Windows PC",
		"Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 Chrome/126.0 Mobile Safari/537.36":  "Android phone",
		"Mozilla/5.0 (Linux; Android 13; SM-X200) AppleWebKit/537.36 Chrome/126.0 Safari/537.36":         "Android tablet",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_5) AppleWebKit/605.1.15 Version/17.5 Safari/605.1.15": "Mac",
		"Mozilla/5.0 (X11; CrOS x86_64 14541.0.0) AppleWebKit/537.36 Chrome/126.0 Safari/537.36":         "Chromebook",
		"Ferry/0.3.0":        "Ferry",
		"Go-http-client/1.1": "Ferry",
		"curl/8.5.0":         "Unknown device",
		"":                   "Unknown device",
		"Mozilla/5.0 (iPad)": "iPad",
		"Mozilla/5.0 (X11; Linux x86_64) Firefox/127.0": "Linux computer",
	}
	for ua, want := range cases {
		if got := deviceName(ua); got != want {
			t.Errorf("deviceName(%q) = %q, want %q", ua, got, want)
		}
	}
}

func TestVisitorsNewestFirst(t *testing.T) {
	var v visitors
	phone := httptest.NewRequest("GET", "/", nil)
	phone.Header.Set("User-Agent", uaIPhone)
	phone.RemoteAddr = "192.168.1.23:5000"
	pc := httptest.NewRequest("GET", "/", nil)
	pc.Header.Set("User-Agent", uaWindows)
	pc.RemoteAddr = "192.168.1.40:5000"

	v.saw(phone)
	v.saw(pc)
	v.saw(phone) // the same visitor again is not a second row
	now := time.Now()
	v.seen["iPhone|192.168.1.23"].last = now.Add(-3 * time.Minute)
	v.seen["Windows PC|192.168.1.40"].last = now.Add(-10 * time.Second)

	got := v.list(now)
	if len(got) != 2 {
		t.Fatalf("want 2 visitors, got %+v", got)
	}
	if got[0].Device != "Windows PC" || got[0].Seen != "just now" {
		t.Errorf("first row = %+v", got[0])
	}
	if got[1].Device != "iPhone" || got[1].IP != "192.168.1.23" || got[1].Seen != "3 min ago" {
		t.Errorf("second row = %+v", got[1])
	}
}

func TestVisitorsListIsCapped(t *testing.T) {
	var v visitors
	for i := 0; i < maxVisitors+10; i++ {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = fmt.Sprintf("10.0.0.%d:1", i)
		v.saw(r)
	}
	if n := len(v.list(time.Now())); n != maxVisitors {
		t.Fatalf("kept %d visitors, want %d", n, maxVisitors)
	}
}

func TestOpenLogSkipsRepeatsAndPieces(t *testing.T) {
	var o openLog
	count := func(what string) int {
		n := 0
		for _, line := range recentActivity(200) {
			if strings.HasSuffix(line, what) {
				n++
			}
		}
		return n
	}
	req := func(method, ua, rng string) {
		r := httptest.NewRequest(method, "/clip.mp4", nil)
		r.RemoteAddr = "192.168.1.23:5000"
		r.Header.Set("User-Agent", ua)
		if rng != "" {
			r.Header.Set("Range", rng)
		}
		o.note(r, "opened open-log-test.mp4")
	}

	req("GET", uaIPhone, "")
	req("GET", uaIPhone, "")              // opened again straight away
	req("GET", uaIPhone, "bytes=500000-") // the rest of the video streaming in
	req("HEAD", uaWindows, "")            // only asking about the file
	req("GET", "Ferry/0.3.0", "")         // another computer copying the folder
	if n := count("iPhone (192.168.1.23) opened open-log-test.mp4"); n != 1 {
		t.Fatalf("iPhone open listed %d times, want 1", n)
	}
	if n := count("open-log-test.mp4"); n != 1 {
		t.Fatalf("listed %d lines in all, want 1", n)
	}

	req("GET", uaWindows, "bytes=0-") // a different visitor starting the video
	if n := count("Windows PC (192.168.1.23) opened open-log-test.mp4"); n != 1 {
		t.Fatalf("second visitor listed %d times, want 1", n)
	}
}
