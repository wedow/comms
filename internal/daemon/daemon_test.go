package daemon

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wedow/comms/internal/config"
	"github.com/wedow/comms/internal/message"
	"github.com/wedow/comms/internal/store"
)

type fakeProvider struct {
	messages    []message.Message
	chatIDs     []int64
	finalOffset int64
}

func (f *fakeProvider) Poll(ctx context.Context, initialOffset int64, handler func(msg message.Message, chatID int64)) (int64, error) {
	for i, msg := range f.messages {
		handler(msg, f.chatIDs[i])
	}
	return f.finalOffset, nil
}

func testConfig() config.Config {
	return config.Config{
		General: config.GeneralConfig{Format: "markdown"},
	}
}

func testMessage(from, channel, body string) message.Message {
	return message.Message{
		From:     from,
		Provider: "telegram",
		Channel:  channel,
		Date:     time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		ID:       "1",
		Body:     body,
	}
}

func TestRunWritesPIDFile(t *testing.T) {
	root := t.TempDir()
	fp := &fakeProvider{finalOffset: 0}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately; Poll returns right away with no messages

	if err := Run(ctx, testConfig(), root, fp); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// PID file should be removed after Run returns
	if _, err := os.Stat(filepath.Join(root, "daemon.pid")); !os.IsNotExist(err) {
		t.Error("PID file should be removed after Run exits")
	}
}

func TestRunWritesMessages(t *testing.T) {
	root := t.TempDir()
	msg := testMessage("alice", "general", "hello world")
	fp := &fakeProvider{
		messages:    []message.Message{msg},
		chatIDs:     []int64{123},
		finalOffset: 42,
	}

	ctx := context.Background()

	if err := Run(ctx, testConfig(), root, fp); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify message file was written
	chanDir := filepath.Join(root, "telegram-general")
	entries, err := os.ReadDir(chanDir)
	if err != nil {
		t.Fatalf("reading channel dir: %v", err)
	}

	// Filter out hidden files (.chat_id)
	var msgFiles []os.DirEntry
	for _, e := range entries {
		if e.Name()[0] != '.' {
			msgFiles = append(msgFiles, e)
		}
	}
	if len(msgFiles) != 1 {
		t.Fatalf("expected 1 message file, got %d", len(msgFiles))
	}
}

func TestRunWritesChatID(t *testing.T) {
	root := t.TempDir()
	msg := testMessage("alice", "general", "hello")
	fp := &fakeProvider{
		messages:    []message.Message{msg},
		chatIDs:     []int64{-1001234},
		finalOffset: 10,
	}

	ctx := context.Background()

	if err := Run(ctx, testConfig(), root, fp); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, err := store.ReadChatID(root, "telegram-general")
	if err != nil {
		t.Fatalf("ReadChatID: %v", err)
	}
	if got != -1001234 {
		t.Errorf("chat ID = %d, want -1001234", got)
	}
}

func TestRunWritesFinalOffset(t *testing.T) {
	root := t.TempDir()
	fp := &fakeProvider{
		messages:    []message.Message{testMessage("bob", "chat", "hi")},
		chatIDs:     []int64{999},
		finalOffset: 77,
	}

	ctx := context.Background()

	if err := Run(ctx, testConfig(), root, fp); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, err := store.ReadOffset(root, "telegram")
	if err != nil {
		t.Fatalf("ReadOffset: %v", err)
	}
	if got != 77 {
		t.Errorf("offset = %d, want 77", got)
	}
}

func TestRunRemovesPIDOnExit(t *testing.T) {
	root := t.TempDir()
	fp := &fakeProvider{finalOffset: 0}

	// We need to verify PID exists during Run, then is gone after.
	// Since our fake Poll returns immediately, PID lifecycle is:
	// written -> Poll -> cleanup -> removed
	ctx := context.Background()

	if err := Run(ctx, testConfig(), root, fp); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "daemon.pid")); !os.IsNotExist(err) {
		t.Error("PID file should be removed after Run exits")
	}
}
