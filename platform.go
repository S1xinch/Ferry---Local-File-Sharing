package main

import (
	"errors"
	"os"
	"os/exec"
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
