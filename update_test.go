package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNewer(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"v0.1.2", "0.1.1", true},
		{"0.1.1", "0.1.1", false},
		{"v0.1.0", "0.1.1", false},
		{"v1.0.0", "0.9.9", true},
		{"v0.10.0", "0.9.0", true},
		{"nonsense", "0.1.1", false},
		{"v0.1.2", "dev", false},
	}
	for _, c := range cases {
		if got := newer(c.latest, c.current); got != c.want {
			t.Errorf("newer(%q, %q) = %v, want %v", c.latest, c.current, got, c.want)
		}
	}
}

func TestAssetFor(t *testing.T) {
	if got := assetFor("darwin", "arm64"); got != "ferry-mac.zip" {
		t.Errorf("mac: got %q", got)
	}
	if got := assetFor("windows", "amd64"); got != "ferry-windows-x64.exe" {
		t.Errorf("windows: got %q", got)
	}
	if got := assetFor("plan9", "amd64"); got != "" {
		t.Errorf("unknown platform should have no asset, got %q", got)
	}
}

func writeZip(t *testing.T, path string, files map[string]string, mode os.FileMode) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, body := range files {
		h := &zip.FileHeader{Name: name, Method: zip.Deflate}
		h.SetMode(mode)
		w, err := zw.CreateHeader(h)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(body))
	}
	zw.Close()
	f.Close()
}

func TestUnzipRefusesPathsOutsideFolder(t *testing.T) {
	dir := t.TempDir()
	zp := filepath.Join(dir, "bad.zip")
	writeZip(t, zp, map[string]string{"../evil.txt": "x"}, 0o644)
	if err := unzipApp(zp, filepath.Join(dir, "out")); err == nil {
		t.Fatal("expected an error for a path that escapes the folder")
	}
	if _, err := os.Stat(filepath.Join(dir, "evil.txt")); err == nil {
		t.Fatal("file was written outside the update folder")
	}
}

func TestUnzipKeepsProgramRunnable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file modes are not kept on Windows")
	}
	dir := t.TempDir()
	zp := filepath.Join(dir, "ok.zip")
	writeZip(t, zp, map[string]string{"Ferry.app/Contents/MacOS/ferry": "bin"}, 0o644)
	out := filepath.Join(dir, "out")
	if err := unzipApp(zp, out); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(out, "Ferry.app", "Contents", "MacOS", "ferry"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("program lost its execute bit: %v", info.Mode())
	}
}

func TestCheckDownloadRejectsNonPrograms(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "ferry-windows-x64.exe")
	os.WriteFile(p, []byte("<html>not a program</html>"), 0o600)
	if err := checkDownload(p, "ferry-windows-x64.exe"); err == nil {
		t.Fatal("an HTML page must not pass as a Windows program")
	}
}
