package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListShowsPreviews(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pic.jpg"), []byte("not really a jpeg"), 0o600)
	os.WriteFile(filepath.Join(dir, "notes.pdf"), []byte("%PDF"), 0o600)
	os.Mkdir(filepath.Join(dir, "Holiday"), 0o755)

	fsh, err := newFileShare(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	fsh.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	body := rec.Body.String()

	for _, want := range []string{`<img src="pic.jpg"`, `class="tile">PDF<`, `class="tile dir">Folder<`} {
		if !strings.Contains(body, want) {
			t.Errorf("listing is missing %q", want)
		}
	}
}

func TestBadge(t *testing.T) {
	if got := badge(".pdf"); got != "PDF" {
		t.Errorf("badge(.pdf) = %q", got)
	}
	if got := badge(""); got != "FILE" {
		t.Errorf("badge(empty) = %q", got)
	}
	if got := badge(".toolongext"); got != "FILE" {
		t.Errorf("badge(long) = %q", got)
	}
}
