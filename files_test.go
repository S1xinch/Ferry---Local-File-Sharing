package main

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"net/url"
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
	// Close the folder before the test removes it; Windows refuses to delete
	// a folder that is still open.
	t.Cleanup(func() { fsh.root.Close() })
	rec := httptest.NewRecorder()
	fsh.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	body := rec.Body.String()

	for _, want := range []string{`<img src="pic.jpg?thumb"`, `class="tile">PDF<`, `class="tile dir">Folder<`} {
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

func TestDropBoxHidesTheFolder(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("private"), 0o600)
	os.Mkdir(filepath.Join(dir, "sub"), 0o755)
	fsh, err := newFileShare(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { fsh.root.Close() })
	fsh.dropOnly = true

	get := func(target string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		fsh.ServeHTTP(w, httptest.NewRequest("GET", target, nil))
		return w
	}
	page := get("/")
	if page.Code != 200 || !strings.Contains(page.Body.String(), `id="drop"`) {
		t.Fatalf("drop page: %d", page.Code)
	}
	for _, target := range []string{"/", "/?zip", "/?manifest"} {
		if body := get(target).Body.String(); strings.Contains(body, "secret") {
			t.Errorf("%s shows what is in the folder", target)
		}
	}
	for _, target := range []string{"/secret.txt", "/secret.txt?dl", "/sub/", "/sub/?zip"} {
		if code := get(target).Code; code != 404 {
			t.Errorf("%s: got %d, want 404", target, code)
		}
	}

	send := func(target, name string) *httptest.ResponseRecorder {
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		fw, _ := mw.CreateFormFile("files", name)
		fw.Write([]byte("from a phone"))
		mw.Close()
		r := httptest.NewRequest("POST", target, &body)
		r.Header.Set("Content-Type", mw.FormDataContentType())
		w := httptest.NewRecorder()
		fsh.ServeHTTP(w, r)
		return w
	}
	if w := send("/", "secret.txt"); w.Code != 303 || w.Header().Get("Location") != "/?done=1" {
		t.Fatalf("upload: %d %q", w.Code, w.Header().Get("Location"))
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "secret.txt")); string(b) != "private" {
		t.Fatal("an upload replaced a file already in the folder")
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "secret (1).txt")); string(b) != "from a phone" {
		t.Fatal("upload was not saved next to the existing file")
	}
	if w := send("/sub/", "x.txt"); w.Code != 404 {
		t.Errorf("upload into a subfolder: got %d, want 404", w.Code)
	}
	if !strings.Contains(get("/?done=1").Body.String(), "1 file uploaded.") {
		t.Error("no thank you after uploading")
	}
}

func TestDoneNotice(t *testing.T) {
	cases := map[string]string{"": "", "1": "1 file uploaded.", "3": "3 files uploaded.", "<b>": "", "-2": ""}
	for q, want := range cases {
		r := httptest.NewRequest("GET", "/?done="+url.QueryEscape(q), nil)
		if got := doneNotice(r); got != want {
			t.Errorf("done=%q: got %q, want %q", q, got, want)
		}
	}
}
