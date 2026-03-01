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
