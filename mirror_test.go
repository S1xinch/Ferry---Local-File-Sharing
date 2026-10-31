package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMirrorFollowsSharedFolder(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	os.MkdirAll(filepath.Join(src, "photos"), 0o755)
	os.WriteFile(filepath.Join(src, "a.txt"), []byte("one"), 0o644)
	os.WriteFile(filepath.Join(src, "photos", "b.txt"), []byte("two"), 0o644)

	s, err := startShare(config{mode: "files", folder: src, pin: "424242", readOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.stop)

	ctx := context.Background()
	m := newMirror(fmt.Sprintf("127.0.0.1:%d", s.port), "424242", dst)
	if err := m.login(ctx); err != nil {
		t.Fatal(err)
	}
	if err := m.round(ctx); err != nil {
		t.Fatal(err)
	}
	mustRead(t, filepath.Join(dst, "a.txt"), "one")
	mustRead(t, filepath.Join(dst, "photos", "b.txt"), "two")

	// A change on the shared side is copied, with a new size so it is noticed at once.
	os.WriteFile(filepath.Join(src, "a.txt"), []byte("uno, changed"), 0o644)
	// A file removed on the shared side goes to the trash, not deleted.
	os.Remove(filepath.Join(src, "photos", "b.txt"))
	// A file the mirror did not download is never touched.
	os.WriteFile(filepath.Join(dst, "mine.txt"), []byte("keep me"), 0o644)
	if err := m.round(ctx); err != nil {
		t.Fatal(err)
	}
	mustRead(t, filepath.Join(dst, "a.txt"), "uno, changed")
	mustRead(t, filepath.Join(dst, "mine.txt"), "keep me")
	if _, err := os.Stat(filepath.Join(dst, "photos", "b.txt")); !os.IsNotExist(err) {
		t.Fatal("removed file should no longer be in the mirror folder")
	}
	trashed, _ := filepath.Glob(filepath.Join(dst, mirrorTrashDir, "*", "photos", "b.txt"))
	if len(trashed) != 1 {
		t.Fatalf("removed file should be in the trash, found %v", trashed)
	}
	if st := m.snapshot(); st.Changed != 1 || st.Removed != 1 || st.Files != 1 {
		t.Errorf("status after the changing round: %+v", st)
	}
	// A round with nothing new downloads nothing and moves nothing.
	if err := m.round(ctx); err != nil {
		t.Fatal(err)
	}
	if st := m.snapshot(); st.Changed != 0 || st.Removed != 0 {
		t.Errorf("a quiet round should change nothing: %+v", st)
	}
}

func TestMirrorRejectsWrongPIN(t *testing.T) {
	src := t.TempDir()
	s, err := startShare(config{mode: "files", folder: src, pin: "111111", readOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.stop)
	m := newMirror(fmt.Sprintf("127.0.0.1:%d", s.port), "999999", t.TempDir())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := m.login(ctx); err == nil {
		t.Fatal("a wrong PIN must not log in")
	}
}

func mustRead(t *testing.T, path, want string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != want {
		t.Errorf("%s = %q, want %q", path, b, want)
	}
}
