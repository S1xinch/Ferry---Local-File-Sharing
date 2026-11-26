package main

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// Who is using a share, for the activity list and the Visitors list in the
// app window. Visitors are named after their device so people recognise
// them: "iPhone (192.168.1.23)" rather than just an address.

// deviceName turns a browser's User-Agent into a name people recognise.
// An iPad asking for the desktop version of sites says it is a Mac, which
// is close enough here.
func deviceName(ua string) string {
	switch {
	case strings.HasPrefix(ua, "Ferry/"), strings.HasPrefix(ua, "Go-http-client/"):
		return "Ferry" // another computer keeping a copy; older versions send Go's default
	case strings.Contains(ua, "iPhone"):
		return "iPhone"
	case strings.Contains(ua, "iPad"):
		return "iPad"
	case strings.Contains(ua, "Android"):
		if strings.Contains(ua, "Mobile") {
			return "Android phone"
		}
		return "Android tablet"
	case strings.Contains(ua, "CrOS"):
		return "Chromebook"
	case strings.Contains(ua, "Windows"):
		return "Windows PC"
	case strings.Contains(ua, "Macintosh"):
		return "Mac"
	case strings.Contains(ua, "Linux"):
		return "Linux computer"
	}
	return "Unknown device"
}

// who names the visitor making a request, like "iPhone (192.168.1.23)".
func who(r *http.Request) string {
	return deviceName(r.UserAgent()) + " (" + clientIP(r) + ")"
}

const maxVisitors = 50

// visitorView is one row of the Visitors list in the app window.
type visitorView struct {
	Device string `json:"device"`
	IP     string `json:"ip"`
	Seen   string `json:"seen"` // "just now", "5 min ago" or "at 14:02"
}

type visitor struct {
	device, ip string
	last       time.Time
}

// visitors remembers who has opened a share and when each was last seen.
// The zero value is ready to use.
type visitors struct {
	mu   sync.Mutex
	seen map[string]*visitor // by device and address
}

func (v *visitors) saw(r *http.Request) {
	device, ip, now := deviceName(r.UserAgent()), clientIP(r), time.Now()
	key := device + "|" + ip
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.seen == nil {
		v.seen = map[string]*visitor{}
	}
	if e := v.seen[key]; e != nil {
		e.last = now
		return
	}
	if len(v.seen) >= maxVisitors {
		var oldest string
		for k, e := range v.seen {
			if oldest == "" || e.last.Before(v.seen[oldest].last) {
				oldest = k
			}
		}
		delete(v.seen, oldest)
	}
	v.seen[key] = &visitor{device: device, ip: ip, last: now}
}

// list returns everyone seen, the most recent first.
func (v *visitors) list(now time.Time) []visitorView {
	v.mu.Lock()
	all := make([]visitor, 0, len(v.seen))
	for _, e := range v.seen {
		all = append(all, *e)
	}
	v.mu.Unlock()
	sort.Slice(all, func(i, j int) bool { return all[i].last.After(all[j].last) })
	out := []visitorView{}
	for _, e := range all {
		out = append(out, visitorView{Device: e.device, IP: e.ip, Seen: ago(e.last, now)})
	}
	return out
}

func ago(t, now time.Time) string {
	switch d := now.Sub(t); {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d min ago", int(d.Minutes()))
	}
	return "at " + t.Format("15:04")
}

// openQuiet is how long the same visitor opening the same thing again stays
// out of the activity list.
const openQuiet = 2 * time.Minute

// openLog notes files and zips that visitors open or download. A video
// streams in many pieces and pages get reloaded, so only the first request
// for the start of a file counts, once per visitor every couple of minutes.
// The zero value is ready to use.
type openLog struct {
	mu   sync.Mutex
	last map[string]time.Time
}

// note adds "<visitor> <what>" to the activity list, like
// "iPhone (192.168.1.23) downloaded Holiday/beach.jpg".
func (o *openLog) note(r *http.Request, what string) {
	if r.Method != http.MethodGet {
		return
	}
	if rg := r.Header.Get("Range"); rg != "" && !strings.HasPrefix(rg, "bytes=0-") {
		return
	}
	if deviceName(r.UserAgent()) == "Ferry" {
		return // another computer copying the folder; its login is already listed
	}
	label, now := who(r), time.Now()
	key := label + "|" + what
	o.mu.Lock()
	if o.last == nil {
		o.last = map[string]time.Time{}
	}
	if now.Sub(o.last[key]) < openQuiet {
		o.mu.Unlock()
		return
	}
	if len(o.last) > 500 {
		for k, t := range o.last {
			if now.Sub(t) >= openQuiet {
				delete(o.last, k)
			}
		}
	}
	o.last[key] = now
	o.mu.Unlock()
	logf("%s %s", label, what)
}
