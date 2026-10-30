package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	repoName    = "S1xinch/Ferry---Local-File-Sharing"
	latestAPI   = "https://api.github.com/repos/" + repoName + "/releases/latest"
	downloadPfx = "https://github.com/" + repoName + "/releases/download/"
	maxDownload = 300 << 20 // no Ferry download is anywhere near this
)

// updateState is what the app window shows about updates.
type updateState struct {
	Version string `json:"version,omitempty"` // the newer version, e.g. "v0.1.2"
	Ready   bool   `json:"ready"`             // downloaded and waiting for a restart
}

// stagedUpdate is a checked new version, ready to be swapped in.
type stagedUpdate struct {
	version string
	path    string // the new program, or on Mac the new Ferry.app
}

// newer reports whether latest is a higher version than current.
// Versions look like "v0.1.2" or "0.1.2".
func newer(latest, current string) bool {
	a, okA := parseVersion(latest)
	b, okB := parseVersion(current)
	if !okA || !okB {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}

func parseVersion(s string) ([3]int, bool) {
	var v [3]int
	parts := strings.SplitN(strings.TrimPrefix(strings.TrimSpace(s), "v"), ".", 3)
	if len(parts) != 3 {
		return v, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return v, false
		}
		v[i] = n
	}
	return v, true
}

// assetFor names the release file for this computer, or "" if there is none.
func assetFor(goos, goarch string) string {
	switch {
	case goos == "darwin":
		return "ferry-mac.zip"
	case goos == "windows" && goarch == "amd64":
		return "ferry-windows-x64.exe"
	case goos == "windows" && goarch == "arm64":
		return "ferry-windows-arm64.exe"
	case goos == "linux" && goarch == "amd64":
		return "ferry-linux-x64"
	case goos == "linux" && goarch == "arm64":
		return "ferry-linux-arm64"
	}
	return ""
}

type release struct {
	Tag    string `json:"tag_name"`
	Assets []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func fetchLatest() (release, error) {
	req, err := http.NewRequest(http.MethodGet, latestAPI, nil)
	if err != nil {
		return release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Ferry/"+version)
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return release{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return release{}, fmt.Errorf("GitHub answered %s", resp.Status)
	}
	var r release
	err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&r)
	return r, err
}

// checkForUpdate runs once when the app opens. If a newer release exists it
// downloads it in the background; nothing is replaced until the person
// clicks "Restart to update". Offline or failing quietly is fine: it tries
// again next time the app opens.
func (a *app) checkForUpdate() {
	if version == "dev" {
		return
	}
	rel, err := fetchLatest()
	if err != nil || !newer(rel.Tag, version) {
		return
	}
	want := assetFor(runtime.GOOS, runtime.GOARCH)
	url := ""
	for _, as := range rel.Assets {
		if as.Name == want && strings.HasPrefix(as.URL, downloadPfx) {
			url = as.URL
		}
	}
	if url == "" {
		return
	}
	path, err := downloadUpdate(url, want)
	if err != nil {
		logf("Update to %s could not be downloaded: %v", rel.Tag, err)
		return
	}
	a.mu.Lock()
	a.staged = &stagedUpdate{version: rel.Tag, path: path}
	a.update = updateState{Version: rel.Tag, Ready: true}
	a.mu.Unlock()
	logf("Ferry %s is downloaded and ready to install", rel.Tag)
}

func downloadUpdate(url, asset string) (string, error) {
	dir := filepath.Join(configDir(), "update")
	os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	file := filepath.Join(dir, asset)
	if err := fetchTo(url, file); err != nil {
		os.RemoveAll(dir)
		return "", err
	}
	if err := checkDownload(file, asset); err != nil {
		os.RemoveAll(dir)
		return "", err
	}
	if !strings.HasSuffix(asset, ".zip") {
		return file, nil
	}
	if err := unzipApp(file, dir); err != nil {
		os.RemoveAll(dir)
		return "", err
	}
	os.Remove(file)
	bundle := filepath.Join(dir, "Ferry.app")
	if _, err := os.Stat(filepath.Join(bundle, "Contents", "MacOS", "ferry")); err != nil {
		os.RemoveAll(dir)
		return "", errors.New("the download does not contain Ferry")
	}
	return bundle, nil
}

func fetchTo(url, file string) error {
	resp, err := (&http.Client{Timeout: 15 * time.Minute}).Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	out, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	n, err := io.Copy(out, io.LimitReader(resp.Body, maxDownload+1))
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err == nil && n > maxDownload {
		err = errors.New("the update is too large")
	}
	return err
}

// checkDownload makes sure the file is what it claims to be before anything
// is replaced: a zip for Mac, or a real program for Windows and Linux.
func checkDownload(path, asset string) error {
	if strings.HasSuffix(asset, ".zip") {
		zr, err := zip.OpenReader(path)
		if err != nil {
			return errors.New("the download is not a valid zip file")
		}
		return zr.Close()
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	head := make([]byte, 4)
	io.ReadFull(f, head)
	bad := (strings.HasSuffix(asset, ".exe") && !bytes.HasPrefix(head, []byte("MZ"))) ||
		(strings.HasPrefix(asset, "ferry-linux") && !bytes.HasPrefix(head, []byte("\x7fELF")))
	if bad {
		return errors.New("the download is not a Ferry program")
	}
	return nil
}

// unzipApp extracts the Mac app. Any entry that would land outside dest, or a
// link, is refused.
func unzipApp(zipPath, dest string) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, f := range zr.File {
		name := filepath.FromSlash(f.Name)
		if !filepath.IsLocal(name) {
			return fmt.Errorf("unexpected file %q in the update", f.Name)
		}
		target := filepath.Join(dest, name)
		mode := f.Mode()
		switch {
		case mode.IsDir():
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		case !mode.IsRegular():
			return fmt.Errorf("unexpected link %q in the update", f.Name)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := writeZipFile(f, target); err != nil {
			return err
		}
	}
	return nil
}

func writeZipFile(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	perm := f.Mode().Perm()
	if strings.Contains(f.Name, "/MacOS/") {
		perm = 0o755 // the program itself must stay runnable
	}
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, rc)
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	return err
}

// installUpdate puts the staged version in place and returns what to start
// once this copy of Ferry has closed.
func (a *app) installUpdate() (string, error) {
	a.mu.Lock()
	st := a.staged
	a.mu.Unlock()
	if st == nil {
		return "", errors.New("there is no update ready")
	}
	self, err := os.Executable()
	if err != nil {
		return "", err
	}
	if self, err = filepath.EvalSymlinks(self); err != nil {
		return "", err
	}

	target := self
	if runtime.GOOS == "darwin" {
		bundle := filepath.Dir(filepath.Dir(filepath.Dir(self)))
		if filepath.Ext(bundle) != ".app" {
			return "", errors.New("this copy of Ferry is not the Mac app, so it cannot update itself")
		}
		err = swapBundle(bundle, st.path)
		target = bundle
	} else {
		err = swapBinary(self, st.path)
	}
	if err != nil {
		return "", fmt.Errorf("could not replace Ferry, move it to a folder you own and try again: %w", err)
	}
	a.mu.Lock()
	a.staged = nil
	a.update = updateState{}
	a.mu.Unlock()
	logf("Updated Ferry to %s, restarting", st.version)
	return target, nil
}

// swapBinary renames the running program aside (allowed on Windows too) and
// moves the new one into its place. If the second step fails, it puts the
// old one back.
func swapBinary(exe, newPath string) error {
	tmp := exe + ".new"
	if err := copyFile(newPath, tmp); err != nil {
		return err
	}
	old := exe + ".old"
	os.Remove(old)
	if err := os.Rename(exe, old); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, exe); err != nil {
		os.Rename(old, exe)
		return err
	}
	return nil
}

// swapBundle copies the new app next to the old one (ditto keeps its
// signature and permissions), then swaps the two folders.
func swapBundle(bundle, newApp string) error {
	staged := filepath.Join(filepath.Dir(bundle), "Ferry.app.new")
	os.RemoveAll(staged)
	if out, err := exec.Command("ditto", newApp, staged).CombinedOutput(); err != nil {
		os.RemoveAll(staged)
		return fmt.Errorf("%v: %s", err, bytes.TrimSpace(out))
	}
	old := bundle + ".old"
	os.RemoveAll(old)
	if err := os.Rename(bundle, old); err != nil {
		os.RemoveAll(staged)
		return err
	}
	if err := os.Rename(staged, bundle); err != nil {
		os.Rename(old, bundle)
		return err
	}
	os.RemoveAll(old)
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	return err
}

// cleanupUpdateLeftovers removes files from earlier updates. Files still in
// use (on Windows) fail to delete and are tried again next time.
func cleanupUpdateLeftovers() {
	if version == "dev" {
		return
	}
	os.RemoveAll(filepath.Join(configDir(), "update"))
	self, err := os.Executable()
	if err != nil {
		return
	}
	if self, err = filepath.EvalSymlinks(self); err != nil {
		return
	}
	os.Remove(self + ".old")
	os.Remove(self + ".new")
	if runtime.GOOS == "darwin" {
		bundle := filepath.Dir(filepath.Dir(filepath.Dir(self)))
		if filepath.Ext(bundle) == ".app" {
			os.RemoveAll(bundle + ".old")
			os.RemoveAll(filepath.Join(filepath.Dir(bundle), "Ferry.app.new"))
		}
	}
}

// startProgram opens a new copy of Ferry after this one has closed.
func startProgram(target string) {
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("open", "-n", target)
	} else {
		cmd = exec.Command(target)
	}
	hideWindow(cmd)
	cmd.Start()
}

func (a *app) updateNow(r *http.Request) (any, error) {
	target, err := a.installUpdate()
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	a.relaunch = target
	a.mu.Unlock()
	go func() {
		time.Sleep(300 * time.Millisecond) // let this reply reach the window first
		a.stopApp()
	}()
	return map[string]bool{"ok": true}, nil
}
