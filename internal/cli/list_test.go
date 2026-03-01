package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/wedow/comms/internal/message"
	"github.com/wedow/comms/internal/store"
)

func TestListCommand(t *testing.T) {
	tmpDir := t.TempDir()

	now := time.Date(2026, 2, 24, 14, 30, 5, 0, time.UTC)
	msgs := []message.Message{
		{From: "alice", Provider: "telegram", Channel: "general", Date: now, ID: "t-1", Body: "hello"},
		{From: "bob", Provider: "telegram", Channel: "general", Date: now.Add(time.Second), ID: "t-2", Body: "world"},
		{From: "carol", Provider: "telegram", Channel: "dev", Date: now.Add(2 * time.Second), ID: "t-3", Body: "deploy done"},
	}
	for _, m := range msgs {
		if _, err := store.WriteMessage(tmpDir, m, "markdown"); err != nil {
			t.Fatal(err)
		}
	}

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"list", "--dir", tmpDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("list command: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3:\n%s", len(lines), buf.String())
	}

	// Channels are sorted, so telegram-dev first, then telegram-general
	var got []map[string]any
	for _, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("invalid JSON: %s", line)
		}
		got = append(got, m)
	}

	if got[0]["from"] != "carol" {
		t.Errorf("line 0 from = %v, want carol", got[0]["from"])
	}
	if got[0]["channel"] != "dev" {
		t.Errorf("line 0 channel = %v, want dev", got[0]["channel"])
	}
	if got[1]["from"] != "alice" {
		t.Errorf("line 1 from = %v, want alice", got[1]["from"])
	}
	if got[2]["from"] != "bob" {
		t.Errorf("line 2 from = %v, want bob", got[2]["from"])
	}

	// Verify all lines have expected fields
	for i, m := range got {
		for _, key := range []string{"from", "provider", "channel", "date", "id", "body", "file"} {
			if _, ok := m[key]; !ok {
				t.Errorf("line %d missing key %q", i, key)
			}
		}
	}
}

func TestListCommandChannelFilter(t *testing.T) {
	tmpDir := t.TempDir()

	now := time.Date(2026, 2, 24, 14, 30, 5, 0, time.UTC)
	msgs := []message.Message{
		{From: "alice", Provider: "telegram", Channel: "general", Date: now, ID: "t-1", Body: "hello"},
		{From: "carol", Provider: "telegram", Channel: "dev", Date: now.Add(time.Second), ID: "t-3", Body: "deploy done"},
	}
	for _, m := range msgs {
		if _, err := store.WriteMessage(tmpDir, m, "markdown"); err != nil {
			t.Fatal(err)
		}
	}

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"list", "--dir", tmpDir, "--channel", "telegram-general"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("list command: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1:\n%s", len(lines), buf.String())
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &m); err != nil {
		t.Fatalf("invalid JSON: %s", lines[0])
	}
	if m["from"] != "alice" {
		t.Errorf("from = %v, want alice", m["from"])
	}
}

func TestListCommandEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"list", "--dir", tmpDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("list command: %v", err)
	}

	if buf.String() != "" {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestListCommandInvalidChannel(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(errBuf)
	cmd.SetArgs([]string{"list", "--dir", tmpDir, "--channel", "nonexistent"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for invalid channel, got nil")
	}
}
