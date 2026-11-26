package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTextsKeepTheNewestAndCheckLength(t *testing.T) {
	var ts texts
	for i := 0; i < maxTexts+5; i++ {
		if err := ts.add("iPhone", false, "note"); err != nil {
			t.Fatal(err)
		}
	}
	all := ts.all()
	if len(all) != maxTexts || all[0].ID != maxTexts+5 {
		t.Fatalf("kept %d texts, newest id %d", len(all), all[0].ID)
	}
	if err := ts.add("iPhone", false, "  \n "); err == nil {
		t.Error("blank text was accepted")
	}
	if err := ts.add("iPhone", false, strings.Repeat("é", maxTextLen)); err != nil {
		t.Errorf("text at the limit was refused: %v", err)
	}
	if err := ts.add("iPhone", false, strings.Repeat("a", maxTextLen+1)); err == nil {
		t.Error("text over the limit was accepted")
	}
	ts.clear()
	if len(ts.all()) != 0 {
		t.Error("clear left texts behind")
	}
}

func TestVisitorTexts(t *testing.T) {
	do := func(s *share, method, ctype, body string) (int, map[string]any) {
		r := httptest.NewRequest(method, "/__ferry/texts", strings.NewReader(body))
		r.Header.Set("User-Agent", uaIPhone)
		if ctype != "" {
			r.Header.Set("Content-Type", ctype)
		}
		w := httptest.NewRecorder()
		s.textsHandler(w, r)
		var out map[string]any
		json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}

	s := &share{cfg: config{mode: "files"}}
	s.texts.add("", true, "wifi password: hunter2")
	if code, out := do(s, "POST", "application/json", `{"text":"https://example.com/a"}`); code != 200 || len(out["texts"].([]any)) != 2 {
		t.Fatalf("send: %d %v", code, out)
	}
	if got := s.texts.all()[0]; got.From != "iPhone" || got.Host || got.Text != "https://example.com/a" {
		t.Fatalf("stored %+v", got)
	}
	// A plain form, which any website could submit, is refused.
	if code, _ := do(s, "POST", "application/x-www-form-urlencoded", `text=hi`); code != 400 {
		t.Errorf("form post: got %d, want 400", code)
	}
	if code, out := do(s, "POST", "application/json", `{"text":""}`); code != 400 || out["error"] == nil {
		t.Errorf("empty text: %d %v", code, out)
	}

	// Download only: visitors can read texts but not send them.
	ro := &share{cfg: config{mode: "files", readOnly: true}}
	ro.texts.add("", true, "hello")
	if code, _ := do(ro, "POST", "application/json", `{"text":"hi"}`); code != 403 {
		t.Errorf("download only send: got %d, want 403", code)
	}
	if code, out := do(ro, "GET", "", ""); code != 200 || out["canSend"] != false || len(out["texts"].([]any)) != 1 {
		t.Errorf("download only read: %d %v", code, out)
	}
}
