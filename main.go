// Ferry: share a folder or a local website with other devices,
// on the same Wi-Fi or over the internet, protected by a PIN.
//
// Started with no arguments (double-clicked) it opens the app window.
// With arguments it runs in the terminal.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var version = "dev"

type config struct {
	mode     string // "files" or "site"
	folder   string
	target   int // local port to forward in site mode
	port     int
	pin      string
	internet bool
	readOnly bool
}

func main() {
	if len(os.Args) == 1 || isFinderLaunch(os.Args[1:]) {
		if err := runApp(); err != nil {
			showError("Ferry could not start: " + err.Error())
			os.Exit(1)
		}
		return
	}
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	if err := runCLI(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// Older macOS versions pass "-psn_..." when an app is opened from Finder.
func isFinderLaunch(args []string) bool {
	return len(args) == 1 && len(args[0]) > 5 && args[0][:5] == "-psn_"
}

func usage() {
	fmt.Printf(`Ferry %s: share a folder or a website from this computer.

Usage:
  ferry                        open the app window
  ferry files [folder]         share a folder (default: current folder)
  ferry site <port>            share a website already running on this PC

Options:
  --internet     also make it reachable from anywhere
  --pin 123456   choose the PIN instead of a random one
  --port 8080    port to listen on (default: first free from 8080)
  --read-only    files mode: disable uploads
`, version)
}

func parseArgs(args []string) (config, error) {
	cfg := config{mode: args[0]}
	fs := flag.NewFlagSet("ferry", flag.ContinueOnError)
	fs.Usage = usage
	fs.BoolVar(&cfg.internet, "internet", false, "")
	fs.StringVar(&cfg.pin, "pin", "", "")
	fs.IntVar(&cfg.port, "port", 0, "")
	fs.BoolVar(&cfg.readOnly, "read-only", false, "")

	// Allow flags before or after the positional argument.
	var positional []string
	rest := args[1:]
	for len(rest) > 0 {
		if err := fs.Parse(rest); err != nil {
			return cfg, err
		}
		rest = fs.Args()
		if len(rest) > 0 {
			positional = append(positional, rest[0])
			rest = rest[1:]
		}
	}

	switch cfg.mode {
	case "files":
		cfg.folder = "."
		if len(positional) > 0 {
			cfg.folder = positional[0]
		}
	case "site":
		if len(positional) == 0 {
			return cfg, errors.New("site mode needs the port your website runs on, e.g. ferry site 3000")
		}
		p, err := strconv.Atoi(positional[0])
		if err != nil || p < 1 || p > 65535 {
			return cfg, fmt.Errorf("%q is not a valid port", positional[0])
		}
		cfg.target = p
	case "help", "-h", "--help":
		usage()
		os.Exit(0)
	case "version", "--version":
		fmt.Println(version)
		os.Exit(0)
	default:
		usage()
		return cfg, fmt.Errorf("unknown mode %q", cfg.mode)
	}
	return cfg, nil
}

func runCLI(cfg config) error {
	s, err := startShare(cfg)
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("PIN:", s.cfg.pin)
	fmt.Println()
	for i, u := range s.lanURLs() {
		fmt.Println("Same Wi-Fi: ", u)
		if i == 0 {
			printQR(s.withToken(u))
			fmt.Println("Scan to open on your phone (the code includes the PIN, so only show it to people you trust).")
		}
	}
	if cfg.internet {
		fmt.Println()
		fmt.Println("Starting internet link...")
		if err := s.startInternet(); err == nil {
			u, _ := s.tunnel()
			fmt.Println("Anywhere:   ", u)
			printQR(s.withToken(u))
		}
	}
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	fmt.Println()
	s.stop()
	return nil
}
