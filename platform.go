package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// openBrowser shows a page in the default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	hideWindow(cmd)
	return cmd.Start()
}

// openWindow shows the app window. It opens Chrome or Edge in app mode, which
// is a window with no tabs or address bar, so Ferry looks like its own app.
// If neither is installed it falls back to the default browser.
func openWindow(url string) error {
	if exe, ok := appBrowser(); ok {
		cmd := exec.Command(exe,
			"--app="+url,
			"--user-data-dir="+filepath.Join(configDir(), "window"), // its own profile, so it never mixes with your browsing
			"--no-first-run",
			"--window-size=920,840",
		)
		hideWindow(cmd)
		if err := cmd.Start(); err == nil {
			return nil
		}
	}
	return openBrowser(url)
}

// appBrowser finds a Chromium browser that can open an app window.
func appBrowser() (string, bool) {
	var candidates []string
	switch runtime.GOOS {
	case "windows":
		for _, c := range []struct{ env, rel string }{
			{"ProgramFiles(x86)", `Microsoft\Edge\Application\msedge.exe`},
			{"ProgramFiles", `Microsoft\Edge\Application\msedge.exe`},
			{"ProgramFiles", `Google\Chrome\Application\chrome.exe`},
			{"ProgramFiles(x86)", `Google\Chrome\Application\chrome.exe`},
			{"LOCALAPPDATA", `Google\Chrome\Application\chrome.exe`},
		} {
			if base := os.Getenv(c.env); base != "" {
				candidates = append(candidates, filepath.Join(base, c.rel))
			}
		}
	case "darwin":
		candidates = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		}
	default:
		for _, n := range []string{"google-chrome", "chromium", "chromium-browser", "microsoft-edge"} {
			if p, err := exec.LookPath(n); err == nil {
				candidates = append(candidates, p)
			}
		}
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, true
		}
	}
	return "", false
}

// pickFolder shows the system's folder chooser. An empty result means
// the person cancelled.
func pickFolder() (string, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		script := `[Console]::OutputEncoding = [Text.Encoding]::UTF8
Add-Type -AssemblyName System.Windows.Forms
$owner = New-Object System.Windows.Forms.Form -Property @{TopMost = $true; ShowInTaskbar = $false}
$d = New-Object System.Windows.Forms.FolderBrowserDialog
$d.Description = 'Choose a folder to share'
if ($d.ShowDialog($owner) -eq 'OK') { [Console]::Out.Write($d.SelectedPath) }`
		cmd = exec.Command("powershell", "-NoProfile", "-STA", "-Command", script)
	case "darwin":
		cmd = exec.Command("osascript",
			"-e", `set f to choose folder with prompt "Choose a folder to share"`,
			"-e", `POSIX path of f`)
	default:
		if _, err := exec.LookPath("zenity"); err != nil {
			return "", errors.New("no folder chooser available, type the path instead")
		}
		cmd = exec.Command("zenity", "--file-selection", "--directory", "--title=Choose a folder to share")
	}
	hideWindow(cmd)
	out, err := cmd.Output()
	path := strings.TrimSpace(string(out))
	if err != nil && path == "" {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return "", nil // cancelled
		}
		return "", err
	}
	if runtime.GOOS == "darwin" && len(path) > 1 {
		path = strings.TrimSuffix(path, "/")
	}
	return path, nil
}

// showError tells the person something went wrong when there is no terminal.
func showError(msg string) {
	os.Stderr.WriteString(msg + "\n")
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("powershell", "-NoProfile", "-Command",
			"Add-Type -AssemblyName PresentationFramework; [System.Windows.MessageBox]::Show($env:FERRY_MSG, 'Ferry')")
		cmd.Env = append(os.Environ(), "FERRY_MSG="+msg)
	case "darwin":
		cmd = exec.Command("osascript", "-e", `display alert "Ferry" message (system attribute "FERRY_MSG")`)
		cmd.Env = append(os.Environ(), "FERRY_MSG="+msg)
	default:
		return
	}
	hideWindow(cmd)
	cmd.Run()
}
