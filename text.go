package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// Sending text: a short list of links and notes that visitors and this
// computer can pass to each other while a folder is shared. It is kept in
// memory only and goes when sharing stops.

const (
	maxTexts   = 20
	maxTextLen = 10000 // characters
)

type textMsg struct {
	ID   int    `json:"id"`
	From string `json:"from"` // the visitor's device, empty when Host is set
	Host bool   `json:"host"` // sent from the Ferry window on this computer
	Text string `json:"text"`
	Time string `json:"time"` // "15:04"
}

// texts holds the newest messages first. The zero value is ready to use.
type texts struct {
	mu   sync.Mutex
	list []textMsg
	next int
}

func (t *texts) add(from string, host bool, text string) error {
	if strings.TrimSpace(text) == "" {
		return errors.New("type or paste some text first")
	}
	if utf8.RuneCountInString(text) > maxTextLen {
		return errors.New("that text is too long, the most is 10,000 characters")
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.next++
	msg := textMsg{ID: t.next, From: from, Host: host, Text: text, Time: time.Now().Format("15:04")}
	t.list = append([]textMsg{msg}, t.list...)
	if len(t.list) > maxTexts {
		t.list = t.list[:maxTexts]
	}
	return nil
}

func (t *texts) all() []textMsg {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]textMsg{}, t.list...)
}

func (t *texts) clear() {
	t.mu.Lock()
	t.list = nil
	t.mu.Unlock()
}

// textsHandler is /__ferry/texts on the visitor side. GET lists the texts;
// POST sends one, unless the share is download only.
func (s *share) textsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	fail := func(code int, msg string) {
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]string{"error": msg})
	}
	switch r.Method {
	case http.MethodGet:
	case http.MethodPost:
		if s.cfg.readOnly {
			fail(http.StatusForbidden, "sending text is turned off for this share")
			return
		}
		// Requiring JSON means another website cannot post here with a plain form.
		var in struct {
			Text string `json:"text"`
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") ||
			json.NewDecoder(io.LimitReader(r.Body, 64<<10)).Decode(&in) != nil {
			fail(http.StatusBadRequest, "bad request")
			return
		}
		if err := s.texts.add(deviceName(r.UserAgent()), false, in.Text); err != nil {
			fail(http.StatusBadRequest, err.Error())
			return
		}
		logf("%s sent a text", who(r))
		if s.cfg.printTexts {
			fmt.Printf("\nText from %s:\n%s\n\n", who(r), in.Text)
		}
	default:
		fail(http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"texts": s.texts.all(), "canSend": !s.cfg.readOnly})
}
