package main

import (
	"testing"
	"time"
)

func TestParseArgsDropAndStopAfter(t *testing.T) {
	cfg, err := parseArgs([]string{"files", "--drop", "/tmp/in", "--stop-after", "90m"})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.dropOnly || cfg.folder != "/tmp/in" || cfg.stopAfter != 90*time.Minute {
		t.Fatalf("got %+v", cfg)
	}
	for _, args := range [][]string{
		{"files", "--drop", "--read-only"},
		{"site", "3000", "--drop"},
		{"files", "--stop-after", "-5m"},
		{"files", "--stop-after", "soon"},
	} {
		if _, err := parseArgs(args); err == nil {
			t.Errorf("%v was accepted", args)
		}
	}
}
