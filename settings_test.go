package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func panelCall(t *testing.T, h http.Handler, method, path, body string) (int, map[string]any) {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Host = "127.0.0.1:47615"
	r.AddCookie(&http.Cookie{Name: panelCookie, Value: "secret"})
	r.Header.Set("X-Ferry", "1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var out map[string]any
	json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func TestSettingsAreSavedAndSurviveSharing(t *testing.T) {
	a := &app{token: "secret", port: 47615, dir: t.TempDir(), quit: make(chan struct{}),
		settings: settings{Mode: "files", Prefs: prefs{AutoUpdate: true}}}
	h := a.routes()

	for _, bad := range []string{`{"port":80}`, `{"port":70000}`, `{"port":47615}`} {
		if code, _ := panelCall(t, h, "POST", "/api/prefs", bad); code != 400 {
			t.Errorf("%s: got %d, want 400", bad, code)
		}
	}
	if code, out := panelCall(t, h, "POST", "/api/prefs", `{"autoUpdate":false,"hidden":true,"port":0}`); code != 200 {
		t.Fatalf("saving prefs: %d %v", code, out)
	}

	// Starting a share sends the sharing form only; the Settings choices must stay.
	body := `{"mode":"files","folder":` + jsonString(t, t.TempDir()) + `}`
	if code, out := panelCall(t, h, "POST", "/api/start", body); code != 200 {
		t.Fatalf("start: %d %v", code, out)
	}
	defer a.cur.stop()
	if a.settings.Prefs != (prefs{AutoUpdate: false, Hidden: true}) {
		t.Fatalf("prefs lost when sharing started: %+v", a.settings.Prefs)
	}
	if !a.cur.cfg.hidden {
		t.Fatal("sharing should not be findable when hidden is on")
	}

	// And they are read back from disk next time.
	b := &app{dir: a.dir, settings: settings{Prefs: prefs{AutoUpdate: true}}}
	b.loadSettings()
	if b.settings.Prefs.AutoUpdate || !b.settings.Prefs.Hidden {
		t.Fatalf("prefs not saved: %+v", b.settings.Prefs)
	}
}

func TestOldSettingsKeepAutoUpdateOn(t *testing.T) {
	dir := t.TempDir()
	// A settings file from before the Settings page existed.
	writeFile(t, dir+"/settings.json", `{"mode":"files","folder":"x"}`)
	a := &app{dir: dir, settings: settings{Prefs: prefs{AutoUpdate: true}}}
	a.loadSettings()
	if !a.settings.Prefs.AutoUpdate {
		t.Fatal("automatic update checks should be on for people upgrading")
	}
}

func TestCheckNowReportsStatus(t *testing.T) {
	a := &app{token: "secret", port: 47615, quit: make(chan struct{})}
	h := a.routes()
	if code, _ := panelCall(t, h, "POST", "/api/update/check", "{}"); code != 200 {
		t.Fatalf("check: %d", code)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		a.mu.Lock()
		u := a.update
		a.mu.Unlock()
		if !u.Checking && u.Status != "" {
			if !strings.Contains(u.Status, "development copy") {
				t.Fatalf("unexpected status for a dev build: %q", u.Status)
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("check never finished")
}

func jsonString(t *testing.T, s string) string {
	t.Helper()
	b, _ := json.Marshal(s)
	return string(b)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestShareStopsByItself(t *testing.T) {
	s, err := startShare(config{mode: "files", folder: t.TempDir(), hidden: true, stopAfter: 50 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	a := &app{cur: s}
	if s.ends.IsZero() {
		t.Fatal("share has no end time")
	}
	a.stopWhenDue(s) // returns once the time is up
	if a.cur != nil || !strings.Contains(a.stopped, "stopped by itself") {
		t.Fatalf("cur=%v stopped=%q", a.cur, a.stopped)
	}
	if s.ctx.Err() == nil {
		t.Fatal("the share is still running")
	}
}

func TestDropBoxAndTextsFromTheWindow(t *testing.T) {
	a := &app{token: "secret", port: 47615, dir: t.TempDir(), quit: make(chan struct{}),
		settings: settings{Mode: "files", Prefs: prefs{Hidden: true}}}
	h := a.routes()
	folder := jsonString(t, t.TempDir())

	for _, bad := range []string{`-5`, `20000`} {
		if code, _ := panelCall(t, h, "POST", "/api/start", `{"mode":"files","folder":`+folder+`,"stopAfter":`+bad+`}`); code != 400 {
			t.Errorf("stopAfter %s: got %d, want 400", bad, code)
		}
	}
	if code, _ := panelCall(t, h, "POST", "/api/text", `{"text":"hi"}`); code != 400 {
		t.Errorf("text before sharing: got %d, want 400", code)
	}

	code, out := panelCall(t, h, "POST", "/api/start", `{"mode":"files","folder":`+folder+`,"dropOnly":true,"uploads":false,"stopAfter":60}`)
	if code != 200 {
		t.Fatalf("start: %d %v", code, out)
	}
	defer a.cur.stop()
	if c := a.cur.cfg; !c.dropOnly || c.readOnly || c.stopAfter != time.Hour {
		t.Fatalf("share config %+v", c)
	}

	if code, out := panelCall(t, h, "POST", "/api/text", `{"text":"the wifi password"}`); code != 200 {
		t.Fatalf("send text: %d %v", code, out)
	}
	_, st := panelCall(t, h, "GET", "/api/state", "")
	ts, _ := st["texts"].([]any)
	if len(ts) != 1 || ts[0].(map[string]any)["host"] != true {
		t.Fatalf("texts in state: %v", st["texts"])
	}
	if n, _ := st["endsIn"].(float64); n < 3500 || n > 3601 {
		t.Errorf("endsIn = %v, want about an hour", st["endsIn"])
	}
	if _, ok := st["visitors"].([]any); !ok {
		t.Errorf("visitors missing from state: %v", st["visitors"])
	}
	if code, _ := panelCall(t, h, "POST", "/api/text/clear", "{}"); code != 200 || len(a.cur.texts.all()) != 0 {
		t.Error("clear did not remove the texts")
	}
}
