package main

import (
	"fmt"
	"testing"
	"time"
)

func TestFindsSharingComputer(t *testing.T) {
	s, err := startShare(config{mode: "files", folder: t.TempDir(), pin: "123123", readOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	found, err := findComputers(time.Second)
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("127.0.0.1:%d", s.port)
	ok := false
	for _, c := range found {
		if c.Addr == want && c.Folder != "" {
			ok = true
		}
		if c.Addr == "" || c.Name == "" {
			t.Errorf("reply missing details: %+v", c)
		}
	}
	if !ok {
		t.Fatalf("share at %s not found in %+v", want, found)
	}

	// Once sharing stops, the computer must not answer any more.
	s.stop()
	found, _ = findComputers(time.Second)
	for _, c := range found {
		if c.Addr == want {
			t.Fatal("a stopped share should not answer")
		}
	}
}
