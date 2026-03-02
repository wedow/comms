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

func TestUnreadNoCursor(t *testing.T) {
	tmpDir := t.TempDir()

	now := time.Date(2026, 2, 24, 14, 30, 5, 0, time.UTC)
	msgs := []message.Message{
		{From: "alice", Provider: "telegram", Channel: "general", Date: now, ID: "t-1", Body: "hello"},
		{From: "bob", Provider: "telegram", Channel: "general", Date: now.Add(time.Second), ID: "t-2", Body: "world"},
	}
	for _, m := range msgs {
		if _, err := store.WriteMessage(tmpDir, m, "markdown"); err != nil {
			t.Fatal(err)
		}
	}

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"unread", "--dir", tmpDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unread command: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2:\n%s", len(lines), buf.String())
	}

	var got []map[string]any
	for _, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("invalid JSON: %s", line)
		}
		got = append(got, m)
	}

	if got[0]["from"] != "alice" {
		t.Errorf("line 0 from = %v, want alice", got[0]["from"])
	}
	if got[1]["from"] != "bob" {
		t.Errorf("line 1 from = %v, want bob", got[1]["from"])
	}

	for i, m := range got {
		for _, key := range []string{"from", "provider", "channel", "date", "id", "body", "file"} {
			if _, ok := m[key]; !ok {
				t.Errorf("line %d missing key %q", i, key)
			}
		}
	}

	// Cursor should NOT be advanced
	cursor, err := store.ReadCursor(tmpDir, "telegram-general")
	if err != nil {
		t.Fatalf("reading cursor: %v", err)
	}
	if !cursor.IsZero() {
		t.Errorf("cursor should be zero, got %v", cursor)
	}
}

func TestUnreadIdempotent(t *testing.T) {
	tmpDir := t.TempDir()

	now := time.Date(2026, 2, 24, 14, 30, 5, 0, time.UTC)
	msg := message.Message{From: "alice", Provider: "telegram", Channel: "general", Date: now, ID: "t-1", Body: "hello"}
	if _, err := store.WriteMessage(tmpDir, msg, "markdown"); err != nil {
		t.Fatal(err)
	}

	// First run
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"unread", "--dir", tmpDir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("first unread: %v", err)
	}
	first := strings.TrimSpace(buf.String())

	// Second run should return same messages (cursor not advanced)
	cmd2 := newRootCmd()
	buf2 := new(bytes.Buffer)
	cmd2.SetOut(buf2)
	cmd2.SetArgs([]string{"unread", "--dir", tmpDir})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("second unread: %v", err)
	}
	second := strings.TrimSpace(buf2.String())

	if first != second {
		t.Errorf("unread should be idempotent:\nfirst:  %s\nsecond: %s", first, second)
	}
}

func TestUnreadChannelFilter(t *testing.T) {
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
	cmd.SetArgs([]string{"unread", "--dir", tmpDir, "--channel", "telegram-dev"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unread command: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1:\n%s", len(lines), buf.String())
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &m); err != nil {
		t.Fatalf("invalid JSON: %s", lines[0])
	}
	if m["from"] != "carol" {
		t.Errorf("from = %v, want carol", m["from"])
	}
}

func TestUnreadEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"unread", "--dir", tmpDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unread command: %v", err)
	}

	if buf.String() != "" {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestUnreadMediaFields(t *testing.T) {
	tmpDir := t.TempDir()

	now := time.Date(2026, 2, 24, 14, 30, 5, 0, time.UTC)
	msgs := []message.Message{
		{From: "alice", Provider: "telegram", Channel: "general", Date: now, ID: "t-1", Body: "text only"},
		{From: "bob", Provider: "telegram", Channel: "general", Date: now.Add(time.Second), ID: "t-2",
			MediaType: "photo", MediaURL: "/media/photo.jpg", Caption: "nice pic", Body: "nice pic"},
	}
	for _, m := range msgs {
		if _, err := store.WriteMessage(tmpDir, m, "markdown"); err != nil {
			t.Fatal(err)
		}
	}

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"unread", "--dir", tmpDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unread command: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2:\n%s", len(lines), buf.String())
	}

	var textMsg map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &textMsg); err != nil {
		t.Fatalf("invalid JSON: %s", lines[0])
	}
	if _, ok := textMsg["media_type"]; ok {
		t.Error("text message should not have media_type field")
	}

	var mediaMsg map[string]any
	if err := json.Unmarshal([]byte(lines[1]), &mediaMsg); err != nil {
		t.Fatalf("invalid JSON: %s", lines[1])
	}
	if mediaMsg["media_type"] != "photo" {
		t.Errorf("media_type = %v, want photo", mediaMsg["media_type"])
	}
}

func TestUnreadPartialCursor(t *testing.T) {
	tmpDir := t.TempDir()

	now := time.Date(2026, 2, 24, 14, 30, 5, 0, time.UTC)
	msgs := []message.Message{
		{From: "alice", Provider: "telegram", Channel: "general", Date: now, ID: "t-1", Body: "old"},
		{From: "bob", Provider: "telegram", Channel: "general", Date: now.Add(time.Second), ID: "t-2", Body: "new"},
		{From: "carol", Provider: "telegram", Channel: "general", Date: now.Add(2 * time.Second), ID: "t-3", Body: "newest"},
	}
	for _, m := range msgs {
		if _, err := store.WriteMessage(tmpDir, m, "markdown"); err != nil {
			t.Fatal(err)
		}
	}

	// Set cursor to first message's time
	if err := store.WriteCursor(tmpDir, "telegram-general", now); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"unread", "--dir", tmpDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unread command: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2:\n%s", len(lines), buf.String())
	}

	var got []map[string]any
	for _, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("invalid JSON: %s", line)
		}
		got = append(got, m)
	}

	if got[0]["body"] != "new" {
		t.Errorf("line 0 body = %v, want new", got[0]["body"])
	}
	if got[1]["body"] != "newest" {
		t.Errorf("line 1 body = %v, want newest", got[1]["body"])
	}

	// Cursor should NOT have been advanced
	cursor, err := store.ReadCursor(tmpDir, "telegram-general")
	if err != nil {
		t.Fatalf("reading cursor: %v", err)
	}
	if !cursor.Equal(now) {
		t.Errorf("cursor = %v, should still be %v", cursor, now)
	}
}
