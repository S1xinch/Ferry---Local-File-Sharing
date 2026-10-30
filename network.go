package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

// lanIPs lists this PC's addresses on the local network (Wi-Fi or cable).
func lanIPs() []string {
	var out []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipnet.IP.To4()
			if ip != nil && ip.IsPrivate() {
				out = append(out, ip.String())
			}
		}
	}
	return out
}

func printQR(text string) {
	q, err := qrcode.New(text, qrcode.Low)
	if err != nil {
		return
	}
	fmt.Print(q.ToSmallString(false))
}

// tunnel is a free Cloudflare quick tunnel: no account, no port forwarding.
type tunnel struct {
	URL string
	cmd *exec.Cmd
}

var tunnelURL = regexp.MustCompile(`https://[a-z0-9-]+\.trycloudflare\.com`)

func startTunnel(ctx context.Context, port int, status func(string)) (*tunnel, error) {
	bin, err := findCloudflared(ctx, status)
	if err != nil {
		return nil, err
	}
	status("Connecting...")
	cmd := exec.CommandContext(ctx, bin, "tunnel", "--no-autoupdate", "--url", fmt.Sprintf("http://localhost:%d", port))
	hideWindow(cmd)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	found := make(chan string, 1)
	go func() {
		sc := bufio.NewScanner(stderr)
		for sc.Scan() {
			if m := tunnelURL.FindString(sc.Text()); m != "" {
				select {
				case found <- m:
				default:
				}
			}
		}
	}()

	select {
	case u := <-found:
		return &tunnel{URL: u, cmd: cmd}, nil
	case <-time.After(45 * time.Second):
		cmd.Process.Kill()
		return nil, errors.New("no link after 45 seconds, check your internet connection")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (t *tunnel) Close() {
	if t.cmd.Process != nil {
		t.cmd.Process.Kill()
	}
}

// findCloudflared uses an installed copy if there is one, otherwise
// downloads Cloudflare's official build once and keeps it.
func findCloudflared(ctx context.Context, status func(string)) (string, error) {
	name := "cloudflared"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if p, err := exec.LookPath("cloudflared"); err == nil {
		return p, nil
	}
	// Apps opened from Finder do not see Homebrew's folders on PATH.
	for _, p := range []string{"/opt/homebrew/bin/cloudflared", "/usr/local/bin/cloudflared"} {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dest := filepath.Join(dir, "ferry", name)
	if _, err := os.Stat(dest); err == nil {
		return dest, nil
	}

	status("Downloading the internet helper (first time only)...")
	logf("Downloading cloudflared to %s", dest)
	if err := downloadCloudflared(ctx, dest); err != nil {
		return "", fmt.Errorf("could not download the internet helper: %w", err)
	}
	return dest, nil
}

func cloudflaredAsset() (string, error) {
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "windows/amd64", "windows/arm64": // arm64 Windows runs the x64 build
		return "cloudflared-windows-amd64.exe", nil
	case "windows/386":
		return "cloudflared-windows-386.exe", nil
	case "darwin/arm64":
		return "cloudflared-darwin-arm64.tgz", nil
	case "darwin/amd64":
		return "cloudflared-darwin-amd64.tgz", nil
	case "linux/amd64", "linux/arm64", "linux/arm", "linux/386":
		return "cloudflared-linux-" + runtime.GOARCH, nil
	}
	return "", fmt.Errorf("no download for %s/%s, install cloudflared yourself", runtime.GOOS, runtime.GOARCH)
}

func downloadCloudflared(ctx context.Context, dest string) error {
	asset, err := cloudflaredAsset()
	if err != nil {
		return err
	}
	return downloadAsset(ctx, asset, dest)
}

func downloadAsset(ctx context.Context, asset, dest string) error {
	url := "https://github.com/cloudflare/cloudflared/releases/latest/download/" + asset
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := (&http.Client{Timeout: 5 * time.Minute}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned %s", resp.Status)
	}

	var src io.Reader = resp.Body
	if filepath.Ext(asset) == ".tgz" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return err
		}
		tr := tar.NewReader(gz)
		for {
			h, err := tr.Next()
			if err != nil {
				return errors.New("cloudflared not found in the download")
			}
			if filepath.Base(h.Name) == "cloudflared" && h.Typeflag == tar.TypeReg {
				src = tr
				break
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	tmp := dest + ".part"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, src)
	f.Close()
	if err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dest)
}
