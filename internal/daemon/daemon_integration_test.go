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

// blockingFakeProvider delivers messages then blocks until ctx is cancelled,
// simulating a long-running poll loop.
type blockingFakeProvider struct {
	messages    []message.Message
	chatIDs     []int64
	finalOffset int64
}

func (f *blockingFakeProvider) Poll(ctx context.Context, initialOffset int64, handler func(msg message.Message, chatID int64, isEdit bool), _ func(string, int, string, string, time.Time)) (int64, error) {
	for i, msg := range f.messages {
		handler(msg, f.chatIDs[i], false)
	}
	<-ctx.Done()
	return f.finalOffset, nil
}

func TestIntegrationDaemonLifecycle(t *testing.T) {
	root := t.TempDir()
	markerFile := filepath.Join(t.TempDir(), "callback-fired")

	msg := message.Message{
		From:     "alice",
		Provider: "telegram",
		Channel:  "general",
		Date:     time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		ID:       "1",
		Body:     "integration test message",
	}

	fp := &blockingFakeProvider{
		messages:    []message.Message{msg},
		chatIDs:     []int64{-1001234},
		finalOffset: 42,
	}

	cfg := config.Config{
		General:  config.GeneralConfig{Format: "markdown"},
		Callback: config.CallbackConfig{Command: "touch " + markerFile, Delay: "0s"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, cfg, root, fp)
	}()

	// Wait for handler to process messages (the blocking provider delivers
	// them synchronously before blocking on ctx.Done).
	time.Sleep(200 * time.Millisecond)

	// --- Assertions while daemon is running ---

	// PID file exists
	if _, err := os.Stat(filepath.Join(root, "daemon.pid")); err != nil {
		t.Errorf("PID file should exist while daemon is running: %v", err)
	}

	// Message file written
	chanDir := filepath.Join(root, "telegram-general")
	entries, err := os.ReadDir(chanDir)
	if err != nil {
		t.Fatalf("reading channel dir: %v", err)
	}
	var msgFiles []os.DirEntry
	for _, e := range entries {
		if e.Name()[0] != '.' {
			msgFiles = append(msgFiles, e)
		}
	}
	if len(msgFiles) != 1 {
		t.Fatalf("expected 1 message file, got %d", len(msgFiles))
	}

	// Chat ID written
	chatID, err := store.ReadChatID(root, "telegram-general")
	if err != nil {
		t.Fatalf("ReadChatID: %v", err)
	}
	if chatID != -1001234 {
		t.Errorf("chat ID = %d, want -1001234", chatID)
	}

	// Callback marker file exists
	if _, err := os.Stat(markerFile); err != nil {
		t.Errorf("callback marker file should exist: %v", err)
	}

	// --- Shutdown ---
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return within 5 seconds after cancel")
	}

	// --- Post-shutdown assertions ---

	// PID file cleaned up
	if _, err := os.Stat(filepath.Join(root, "daemon.pid")); !os.IsNotExist(err) {
		t.Error("PID file should be removed after Run exits")
	}

	// Offset persisted
	offset, err := store.ReadOffset(root, "telegram")
	if err != nil {
		t.Fatalf("ReadOffset: %v", err)
	}
	if offset != 42 {
		t.Errorf("offset = %d, want 42", offset)
	}
}
