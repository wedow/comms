package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wedow/comms/internal/message"
)

func TestInitDir(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".comms")

	if err := InitDir(root); err != nil {
		t.Fatalf("InitDir: %v", err)
	}

	for _, sub := range []string{"", "docs"} {
		dir := filepath.Join(root, sub)
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("expected directory %s to exist: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", dir)
		}
	}
}

func TestWriteMessage(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".comms")
	if err := InitDir(root); err != nil {
		t.Fatalf("InitDir: %v", err)
	}

	msg := message.Message{
		From:     "alice",
		Provider: "telegram",
		Channel:  "general",
		Date:     time.Date(2026, 3, 1, 12, 30, 0, 123456789, time.UTC),
		ID:       "42",
		Body:     "hello world",
	}

	path, err := WriteMessage(root, msg, "markdown")
	if err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}

	// Verify channel directory
	chanDir := filepath.Join(root, "telegram-general")
	if _, err := os.Stat(chanDir); err != nil {
		t.Fatalf("expected channel dir %s: %v", chanDir, err)
	}

	// Verify file is under channel dir with .md extension
	if filepath.Dir(path) != chanDir {
		t.Errorf("path dir = %s, want %s", filepath.Dir(path), chanDir)
	}
	if filepath.Ext(path) != ".md" {
		t.Errorf("extension = %s, want .md", filepath.Ext(path))
	}

	// Verify filename has no colons (RFC3339Nano colons replaced)
	base := filepath.Base(path)
	if strings.Contains(base, ":") {
		t.Errorf("filename contains colons: %s", base)
	}

	// Verify content is valid markdown-serialized message
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}
	if !strings.Contains(string(data), "hello world") {
		t.Errorf("file content missing body, got:\n%s", data)
	}
}

func TestWriteMessageWithThreadID(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".comms")
	if err := InitDir(root); err != nil {
		t.Fatalf("InitDir: %v", err)
	}

	msg := message.Message{
		From:     "alice",
		Provider: "telegram",
		Channel:  "general",
		Date:     time.Date(2026, 3, 1, 12, 30, 0, 0, time.UTC),
		ID:       "42",
		ThreadID: "99",
		Body:     "topic message",
	}

	path, err := WriteMessage(root, msg, "markdown")
	if err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}

	// Verify file is under topic-99 subdirectory
	wantDir := filepath.Join(root, "telegram-general", "topic-99")
	if filepath.Dir(path) != wantDir {
		t.Errorf("path dir = %s, want %s", filepath.Dir(path), wantDir)
	}
}

func TestWriteMessageOrg(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".comms")
	if err := InitDir(root); err != nil {
		t.Fatalf("InitDir: %v", err)
	}

	msg := message.Message{
		From:     "bob",
		Provider: "telegram",
		Channel:  "dev",
		Date:     time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		ID:       "99",
		Body:     "org content",
	}

	path, err := WriteMessage(root, msg, "org")
	if err != nil {
		t.Fatalf("WriteMessage org: %v", err)
	}

	if filepath.Ext(path) != ".org" {
		t.Errorf("extension = %s, want .org", filepath.Ext(path))
	}
}

func TestReadMessageRoundTrip(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".comms")
	if err := InitDir(root); err != nil {
		t.Fatalf("InitDir: %v", err)
	}

	original := message.Message{
		From:     "carol",
		Provider: "telegram",
		Channel:  "random",
		Date:     time.Date(2026, 3, 1, 15, 45, 30, 0, time.UTC),
		ID:       "7",
		Body:     "round trip test",
	}

	for _, format := range []string{"markdown", "org"} {
		t.Run(format, func(t *testing.T) {
			path, err := WriteMessage(root, original, format)
			if err != nil {
				t.Fatalf("WriteMessage(%s): %v", format, err)
			}

			got, err := ReadMessage(path)
			if err != nil {
				t.Fatalf("ReadMessage(%s): %v", format, err)
			}

			if got.From != original.From {
				t.Errorf("From = %q, want %q", got.From, original.From)
			}
			if got.Provider != original.Provider {
				t.Errorf("Provider = %q, want %q", got.Provider, original.Provider)
			}
			if got.Channel != original.Channel {
				t.Errorf("Channel = %q, want %q", got.Channel, original.Channel)
			}
			if !got.Date.Equal(original.Date) {
				t.Errorf("Date = %v, want %v", got.Date, original.Date)
			}
			if got.ID != original.ID {
				t.Errorf("ID = %q, want %q", got.ID, original.ID)
			}
			if got.Body != original.Body {
				t.Errorf("Body = %q, want %q", got.Body, original.Body)
			}
		})
	}
}

func TestReadMessageUnknownExtension(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "msg.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadMessage(path)
	if err == nil {
		t.Fatal("expected error for unknown extension, got nil")
	}
}

func TestFindMessageByID(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".comms")
	if err := InitDir(root); err != nil {
		t.Fatalf("InitDir: %v", err)
	}

	msgs := []message.Message{
		{From: "alice", Provider: "telegram", Channel: "general", Date: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC), ID: "telegram-10", Body: "first"},
		{From: "bob", Provider: "telegram", Channel: "general", Date: time.Date(2026, 3, 1, 12, 1, 0, 0, time.UTC), ID: "telegram-20", Body: "second"},
		{From: "carol", Provider: "telegram", Channel: "general", Date: time.Date(2026, 3, 1, 12, 2, 0, 0, time.UTC), ID: "telegram-30", Body: "third"},
	}

	for _, m := range msgs {
		if _, err := WriteMessage(root, m, "markdown"); err != nil {
			t.Fatalf("WriteMessage: %v", err)
		}
	}

	// Find existing message
	path, msg, err := FindMessageByID(root, "telegram-general", "telegram-20", "markdown")
	if err != nil {
		t.Fatalf("FindMessageByID: %v", err)
	}
	if msg.ID != "telegram-20" {
		t.Errorf("ID = %q, want telegram-20", msg.ID)
	}
	if path == "" {
		t.Error("path is empty")
	}

	// Not found
	_, _, err = FindMessageByID(root, "telegram-general", "telegram-99", "markdown")
	if err == nil {
		t.Fatal("expected error for missing ID, got nil")
	}
}

func TestAppendEdit(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".comms")
	if err := InitDir(root); err != nil {
		t.Fatalf("InitDir: %v", err)
	}

	msg := message.Message{
		From:     "alice",
		Provider: "telegram",
		Channel:  "general",
		Date:     time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		ID:       "telegram-10",
		Body:     "original text",
	}
	path, err := WriteMessage(root, msg, "markdown")
	if err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}

	editDate := time.Date(2026, 3, 1, 12, 5, 0, 0, time.UTC)
	if err := AppendEdit(path, editDate, "edited text"); err != nil {
		t.Fatalf("AppendEdit: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "---edit---") {
		t.Error("missing ---edit--- separator")
	}
	if !strings.Contains(content, "date: 2026-03-01T12:05:00Z") {
		t.Error("missing edit date")
	}
	if !strings.Contains(content, "edited text") {
		t.Error("missing edited body")
	}
	// Original content should still be there
	if !strings.Contains(content, "original text") {
		t.Error("missing original body")
	}
}

func TestAppendReaction(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".comms")
	if err := InitDir(root); err != nil {
		t.Fatalf("InitDir: %v", err)
	}

	msg := message.Message{
		From:     "alice",
		Provider: "telegram",
		Channel:  "general",
		Date:     time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		ID:       "telegram-10",
		Body:     "original text",
	}
	path, err := WriteMessage(root, msg, "markdown")
	if err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}

	reactionDate := time.Date(2026, 3, 1, 12, 10, 0, 0, time.UTC)
	if err := AppendReaction(path, reactionDate, "bob", "\U0001f44d"); err != nil {
		t.Fatalf("AppendReaction: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "---reaction---") {
		t.Error("missing ---reaction--- separator")
	}
	if !strings.Contains(content, "date: 2026-03-01T12:10:00Z") {
		t.Error("missing reaction date")
	}
	if !strings.Contains(content, "from: bob") {
		t.Error("missing reaction from")
	}
	if !strings.Contains(content, "emoji: \U0001f44d") {
		t.Error("missing reaction emoji")
	}
	// Original content should still be there
	if !strings.Contains(content, "original text") {
		t.Error("missing original body")
	}
}

func TestWriteMedia(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".comms")
	chanDir := filepath.Join(root, "telegram-general")
	if err := os.MkdirAll(chanDir, 0o755); err != nil {
		t.Fatal(err)
	}

	ts := "2026-03-01T12-00-00Z"
	data := []byte("fake image data")

	path, err := WriteMedia(chanDir, ts, 1, ".jpg", data)
	if err != nil {
		t.Fatalf("WriteMedia: %v", err)
	}

	// Should be <chanDir>/<ts>/001.jpg
	wantDir := filepath.Join(chanDir, ts)
	if filepath.Dir(path) != wantDir {
		t.Errorf("dir = %s, want %s", filepath.Dir(path), wantDir)
	}
	if filepath.Base(path) != "001.jpg" {
		t.Errorf("base = %s, want 001.jpg", filepath.Base(path))
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("content = %q, want %q", got, data)
	}
}

func TestWriteMediaNoExtension(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".comms")
	chanDir := filepath.Join(root, "telegram-general")
	if err := os.MkdirAll(chanDir, 0o755); err != nil {
		t.Fatal(err)
	}

	path, err := WriteMedia(chanDir, "2026-03-01T12-00-00Z", 1, "", []byte("data"))
	if err != nil {
		t.Fatalf("WriteMedia: %v", err)
	}
	if filepath.Base(path) != "001" {
		t.Errorf("base = %s, want 001", filepath.Base(path))
	}
}

func TestResetCursorIfNeeded(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".comms")
	if err := InitDir(root); err != nil {
		t.Fatalf("InitDir: %v", err)
	}

	channel := "telegram-general"
	chanDir := filepath.Join(root, channel)
	if err := os.MkdirAll(chanDir, 0o755); err != nil {
		t.Fatal(err)
	}

	msgDate := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

	// Cursor past the message
	cursorTime := time.Date(2026, 3, 1, 13, 0, 0, 0, time.UTC)
	if err := WriteCursor(root, channel, cursorTime); err != nil {
		t.Fatalf("WriteCursor: %v", err)
	}

	if err := ResetCursorIfNeeded(root, channel, msgDate); err != nil {
		t.Fatalf("ResetCursorIfNeeded: %v", err)
	}

	got, err := ReadCursor(root, channel)
	if err != nil {
		t.Fatalf("ReadCursor: %v", err)
	}
	if !got.Before(msgDate) {
		t.Errorf("cursor %v should be before message date %v", got, msgDate)
	}

	// Cursor before the message should not change
	earlyTime := time.Date(2026, 3, 1, 11, 0, 0, 0, time.UTC)
	if err := WriteCursor(root, channel, earlyTime); err != nil {
		t.Fatalf("WriteCursor: %v", err)
	}
	if err := ResetCursorIfNeeded(root, channel, msgDate); err != nil {
		t.Fatalf("ResetCursorIfNeeded: %v", err)
	}
	got, err = ReadCursor(root, channel)
	if err != nil {
		t.Fatalf("ReadCursor: %v", err)
	}
	if !got.Equal(earlyTime) {
		t.Errorf("cursor = %v, want unchanged %v", got, earlyTime)
	}
}
