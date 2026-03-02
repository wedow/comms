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

func TestAckAdvancesCursor(t *testing.T) {
	tmpDir := t.TempDir()

	now := time.Date(2026, 2, 24, 14, 30, 5, 0, time.UTC)
	msgs := []message.Message{
		{From: "alice", Provider: "telegram", Channel: "general", Date: now, ID: "t-1", Body: "first"},
		{From: "bob", Provider: "telegram", Channel: "general", Date: now.Add(time.Second), ID: "t-2", Body: "second"},
		{From: "carol", Provider: "telegram", Channel: "general", Date: now.Add(2 * time.Second), ID: "t-3", Body: "third"},
	}
	for _, m := range msgs {
		if _, err := store.WriteMessage(tmpDir, m, "markdown"); err != nil {
			t.Fatal(err)
		}
	}

	// Ack up to t-2
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"ack", "--dir", tmpDir, "t-2"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("ack: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output not JSON: %v\n%s", err, buf.String())
	}
	if result["status"] != "acked" {
		t.Errorf("status = %q, want acked", result["status"])
	}
	if result["id"] != "t-2" {
		t.Errorf("id = %q, want t-2", result["id"])
	}

	// Cursor should be at t-2's date
	cursor, err := store.ReadCursor(tmpDir, "telegram-general")
	if err != nil {
		t.Fatal(err)
	}
	if !cursor.Equal(now.Add(time.Second)) {
		t.Errorf("cursor = %v, want %v", cursor, now.Add(time.Second))
	}

	// Unread should now return only t-3
	cmd2 := newRootCmd()
	buf2 := new(bytes.Buffer)
	cmd2.SetOut(buf2)
	cmd2.SetArgs([]string{"unread", "--dir", tmpDir})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("unread: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf2.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1:\n%s", len(lines), buf2.String())
	}
	var msg map[string]any
	json.Unmarshal([]byte(lines[0]), &msg)
	if msg["id"] != "t-3" {
		t.Errorf("unread id = %v, want t-3", msg["id"])
	}
}

func TestAckNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	now := time.Date(2026, 2, 24, 14, 30, 5, 0, time.UTC)
	msg := message.Message{From: "alice", Provider: "telegram", Channel: "general", Date: now, ID: "t-1", Body: "hello"}
	if _, err := store.WriteMessage(tmpDir, msg, "markdown"); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	stderr := new(bytes.Buffer)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"ack", "--dir", tmpDir, "nonexistent"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for nonexistent message")
	}

	if !strings.Contains(stderr.String(), `"error"`) {
		t.Errorf("stderr = %q, want JSON error", stderr.String())
	}
}

func TestAckWithChannelFilter(t *testing.T) {
	tmpDir := t.TempDir()

	now := time.Date(2026, 2, 24, 14, 30, 5, 0, time.UTC)
	msgs := []message.Message{
		{From: "alice", Provider: "telegram", Channel: "general", Date: now, ID: "t-1", Body: "hello"},
		{From: "bob", Provider: "telegram", Channel: "dev", Date: now.Add(time.Second), ID: "t-2", Body: "world"},
	}
	for _, m := range msgs {
		if _, err := store.WriteMessage(tmpDir, m, "markdown"); err != nil {
			t.Fatal(err)
		}
	}

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"ack", "--dir", tmpDir, "--channel", "telegram-dev", "t-2"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("ack: %v", err)
	}

	// telegram-dev cursor advanced
	cursor, err := store.ReadCursor(tmpDir, "telegram-dev")
	if err != nil {
		t.Fatal(err)
	}
	if cursor.IsZero() {
		t.Error("telegram-dev cursor should be set")
	}

	// telegram-general cursor untouched
	cursor2, err := store.ReadCursor(tmpDir, "telegram-general")
	if err != nil {
		t.Fatal(err)
	}
	if !cursor2.IsZero() {
		t.Error("telegram-general cursor should not be set")
	}
}
