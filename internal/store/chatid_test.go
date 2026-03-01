package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChatIDRoundTrip(t *testing.T) {
	root := t.TempDir()
	channel := "telegram-general"

	if err := os.MkdirAll(filepath.Join(root, channel), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	var chatID int64 = 123456789

	if err := WriteChatID(root, channel, chatID); err != nil {
		t.Fatalf("WriteChatID: %v", err)
	}

	got, err := ReadChatID(root, channel)
	if err != nil {
		t.Fatalf("ReadChatID: %v", err)
	}

	if got != chatID {
		t.Errorf("ReadChatID = %d, want %d", got, chatID)
	}
}

func TestReadChatIDMissingFile(t *testing.T) {
	root := t.TempDir()

	_, err := ReadChatID(root, "nonexistent-channel")
	if err == nil {
		t.Fatal("expected error for missing .chat_id file, got nil")
	}
}

func TestChatIDNegative(t *testing.T) {
	root := t.TempDir()
	channel := "telegram-group"

	if err := os.MkdirAll(filepath.Join(root, channel), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	var chatID int64 = -1001234567890

	if err := WriteChatID(root, channel, chatID); err != nil {
		t.Fatalf("WriteChatID: %v", err)
	}

	got, err := ReadChatID(root, channel)
	if err != nil {
		t.Fatalf("ReadChatID: %v", err)
	}

	if got != chatID {
		t.Errorf("ReadChatID = %d, want %d", got, chatID)
	}
}
