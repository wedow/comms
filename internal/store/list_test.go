package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestListChannels(t *testing.T) {
	root := t.TempDir()

	// Create channel dirs, a docs dir (should be excluded), and a plain file (should be excluded)
	for _, name := range []string{"telegram-general", "telegram-dev", "docs"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "config.toml"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	got, err := ListChannels(root)
	if err != nil {
		t.Fatalf("ListChannels: %v", err)
	}

	want := []string{"telegram-dev", "telegram-general"}
	if len(got) != len(want) {
		t.Fatalf("ListChannels = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ListChannels[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestListMessages(t *testing.T) {
	root := t.TempDir()
	channel := "telegram-general"
	chanDir := filepath.Join(root, channel)
	if err := os.MkdirAll(chanDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create message files and a .cursor file (should be excluded)
	files := []string{
		"2026-03-01T12-30-00Z.md",
		"2026-03-01T12-31-00Z.md",
		"2026-03-01T12-29-00Z.md",
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(chanDir, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(chanDir, ".cursor"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ListMessages(root, channel)
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}

	want := []string{
		filepath.Join(chanDir, "2026-03-01T12-29-00Z.md"),
		filepath.Join(chanDir, "2026-03-01T12-30-00Z.md"),
		filepath.Join(chanDir, "2026-03-01T12-31-00Z.md"),
	}
	if len(got) != len(want) {
		t.Fatalf("ListMessages = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ListMessages[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestListMessagesIncludesTopicSubdirs(t *testing.T) {
	root := t.TempDir()
	channel := "telegram-general"
	chanDir := filepath.Join(root, channel)
	topicDir := filepath.Join(chanDir, "topic-42")
	if err := os.MkdirAll(topicDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Root-level message and topic-level message (topic message is chronologically between)
	rootFiles := []string{"2026-03-01T12-29-00Z.md", "2026-03-01T12-31-00Z.md"}
	for _, f := range rootFiles {
		if err := os.WriteFile(filepath.Join(chanDir, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	topicFile := "2026-03-01T12-30-00Z.md"
	if err := os.WriteFile(filepath.Join(topicDir, topicFile), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ListMessages(root, channel)
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}

	want := []string{
		filepath.Join(chanDir, "2026-03-01T12-29-00Z.md"),
		filepath.Join(topicDir, "2026-03-01T12-30-00Z.md"),
		filepath.Join(chanDir, "2026-03-01T12-31-00Z.md"),
	}
	if len(got) != len(want) {
		t.Fatalf("ListMessages = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ListMessages[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestListMessagesAfter(t *testing.T) {
	root := t.TempDir()
	channel := "telegram-general"
	chanDir := filepath.Join(root, channel)
	if err := os.MkdirAll(chanDir, 0o755); err != nil {
		t.Fatal(err)
	}

	files := []string{
		"2026-03-01T12-29-00Z.md",
		"2026-03-01T12-30-00Z.md",
		"2026-03-01T12-31-00Z.md",
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(chanDir, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	after := time.Date(2026, 3, 1, 12, 29, 0, 0, time.UTC)
	got, err := ListMessagesAfter(root, channel, after)
	if err != nil {
		t.Fatalf("ListMessagesAfter: %v", err)
	}

	want := []string{
		filepath.Join(chanDir, "2026-03-01T12-30-00Z.md"),
		filepath.Join(chanDir, "2026-03-01T12-31-00Z.md"),
	}
	if len(got) != len(want) {
		t.Fatalf("ListMessagesAfter = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ListMessagesAfter[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestListMessagesAfterZeroTime(t *testing.T) {
	root := t.TempDir()
	channel := "telegram-general"
	chanDir := filepath.Join(root, channel)
	if err := os.MkdirAll(chanDir, 0o755); err != nil {
		t.Fatal(err)
	}

	files := []string{
		"2026-03-01T12-29-00Z.md",
		"2026-03-01T12-30-00Z.md",
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(chanDir, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Zero time should return all messages
	got, err := ListMessagesAfter(root, channel, time.Time{})
	if err != nil {
		t.Fatalf("ListMessagesAfter zero: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("ListMessagesAfter zero = %d files, want 2", len(got))
	}
}
