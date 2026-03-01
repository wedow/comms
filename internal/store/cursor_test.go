package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCursorRoundTrip(t *testing.T) {
	root := t.TempDir()
	channel := "telegram-general"

	// Create channel directory
	if err := os.MkdirAll(filepath.Join(root, channel), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	ts := time.Date(2026, 3, 1, 12, 30, 0, 123456789, time.UTC)

	if err := WriteCursor(root, channel, ts); err != nil {
		t.Fatalf("WriteCursor: %v", err)
	}

	got, err := ReadCursor(root, channel)
	if err != nil {
		t.Fatalf("ReadCursor: %v", err)
	}

	if !got.Equal(ts) {
		t.Errorf("ReadCursor = %v, want %v", got, ts)
	}
}

func TestReadCursorMissingFile(t *testing.T) {
	root := t.TempDir()

	got, err := ReadCursor(root, "nonexistent-channel")
	if err != nil {
		t.Fatalf("ReadCursor on missing file should not error: %v", err)
	}

	if !got.IsZero() {
		t.Errorf("ReadCursor on missing file = %v, want zero time", got)
	}
}
