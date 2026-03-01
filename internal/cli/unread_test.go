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

	// Verify all expected fields present
	for i, m := range got {
		for _, key := range []string{"from", "provider", "channel", "date", "id", "body", "file"} {
			if _, ok := m[key]; !ok {
				t.Errorf("line %d missing key %q", i, key)
			}
		}
	}

	// Verify cursor was advanced
	cursor, err := store.ReadCursor(tmpDir, "telegram-general")
	if err != nil {
		t.Fatalf("reading cursor: %v", err)
	}
	if !cursor.Equal(now.Add(time.Second)) {
		t.Errorf("cursor = %v, want %v", cursor, now.Add(time.Second))
	}
}

func TestUnreadTwice(t *testing.T) {
	tmpDir := t.TempDir()

	now := time.Date(2026, 2, 24, 14, 30, 5, 0, time.UTC)
	msg := message.Message{From: "alice", Provider: "telegram", Channel: "general", Date: now, ID: "t-1", Body: "hello"}
	if _, err := store.WriteMessage(tmpDir, msg, "markdown"); err != nil {
		t.Fatal(err)
	}

	// First run: should return the message
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"unread", "--dir", tmpDir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("first unread: %v", err)
	}
	if strings.TrimSpace(buf.String()) == "" {
		t.Fatal("first run should have output")
	}

	// Second run: should produce no output
	cmd2 := newRootCmd()
	buf2 := new(bytes.Buffer)
	cmd2.SetOut(buf2)
	cmd2.SetArgs([]string{"unread", "--dir", tmpDir})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("second unread: %v", err)
	}
	if buf2.String() != "" {
		t.Errorf("second run should have no output, got %q", buf2.String())
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

	// Verify only filtered channel's cursor was advanced
	cursor, err := store.ReadCursor(tmpDir, "telegram-dev")
	if err != nil {
		t.Fatalf("reading cursor: %v", err)
	}
	if cursor.IsZero() {
		t.Error("telegram-dev cursor should be set")
	}

	// telegram-general cursor should not exist
	cursor2, err := store.ReadCursor(tmpDir, "telegram-general")
	if err != nil {
		t.Fatalf("reading cursor: %v", err)
	}
	if !cursor2.IsZero() {
		t.Error("telegram-general cursor should not be set")
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

	// Text-only message should NOT have media fields (omitempty)
	var textMsg map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &textMsg); err != nil {
		t.Fatalf("invalid JSON: %s", lines[0])
	}
	if _, ok := textMsg["media_type"]; ok {
		t.Error("text message should not have media_type field")
	}
	if _, ok := textMsg["media_url"]; ok {
		t.Error("text message should not have media_url field")
	}
	if _, ok := textMsg["caption"]; ok {
		t.Error("text message should not have caption field")
	}

	// Media message should have media fields
	var mediaMsg map[string]any
	if err := json.Unmarshal([]byte(lines[1]), &mediaMsg); err != nil {
		t.Fatalf("invalid JSON: %s", lines[1])
	}
	if mediaMsg["media_type"] != "photo" {
		t.Errorf("media_type = %v, want photo", mediaMsg["media_type"])
	}
	if mediaMsg["media_url"] != "/media/photo.jpg" {
		t.Errorf("media_url = %v, want /media/photo.jpg", mediaMsg["media_url"])
	}
	if mediaMsg["caption"] != "nice pic" {
		t.Errorf("caption = %v, want nice pic", mediaMsg["caption"])
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

	// Set cursor to first message's time (should skip it, return second and third)
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

	// Verify cursor advanced to newest
	cursor, err := store.ReadCursor(tmpDir, "telegram-general")
	if err != nil {
		t.Fatalf("reading cursor: %v", err)
	}
	if !cursor.Equal(now.Add(2 * time.Second)) {
		t.Errorf("cursor = %v, want %v", cursor, now.Add(2*time.Second))
	}
}
