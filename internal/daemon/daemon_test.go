package daemon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wedow/comms/internal/config"
	"github.com/wedow/comms/internal/protocol"
	"github.com/wedow/comms/internal/store"
)

func testConfig() config.Config {
	return config.Config{
		General: config.GeneralConfig{Format: "markdown"},
	}
}

func TestRunWritesPIDAndCleansUp(t *testing.T) {
	root := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately so Run returns

	err := Run(ctx, testConfig(), root, []string{"telegram"})
	if err == nil || err != context.Canceled {
		t.Fatalf("Run returned %v, want context.Canceled", err)
	}

	// PID file should be removed after Run returns
	if _, err := os.Stat(filepath.Join(root, "daemon.pid")); !os.IsNotExist(err) {
		t.Error("PID file should be removed after Run exits")
	}
}

func TestRunProcessesMessageEvent(t *testing.T) {
	root := t.TempDir()

	origLookPath := lookPathFunc
	origSpawn := spawnFunc
	t.Cleanup(func() {
		lookPathFunc = origLookPath
		spawnFunc = origSpawn
	})

	lookPathFunc = func(name string) (string, error) {
		return "/fake/" + name, nil
	}

	eventsCh := make(chan any, 8)
	doneCh := make(chan error, 1)

	spawnFunc = func(ctx context.Context, provider, binaryPath, root string, providerConfig []byte, offset int64) (*Subprocess, error) {
		return &Subprocess{
			events:   eventsCh,
			done:     doneCh,
			provider: provider,
		}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		eventsCh <- protocol.MessageEvent{
			Type:    protocol.TypeMessage,
			Offset:  42,
			ID:      1,
			ChatID:  12345,
			Channel: "test-channel",
			From:    "alice",
			Date:    time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC),
			Body:    "hello world",
		}
		time.Sleep(50 * time.Millisecond)
		cancel()
		close(eventsCh)
		doneCh <- nil
	}()

	cfg := config.Config{
		General:   config.GeneralConfig{Format: "markdown"},
		Providers: map[string]map[string]any{"testprov": {"token": "fake"}},
	}

	err := Run(ctx, cfg, root, []string{"testprov"})
	if err != nil && err != context.Canceled {
		t.Fatalf("Run returned unexpected error: %v", err)
	}

	// Verify message was written.
	channelDir := filepath.Join(root, "testprov-test-channel")
	entries, err := os.ReadDir(channelDir)
	if err != nil {
		t.Fatalf("reading channel dir: %v", err)
	}

	foundMsg := false
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			foundMsg = true
		}
	}
	if !foundMsg {
		t.Error("expected message file in channel directory")
	}

	// Verify offset was persisted.
	offset, err := store.ReadOffset(root, "testprov")
	if err != nil {
		t.Fatalf("reading offset: %v", err)
	}
	if offset != 42 {
		t.Errorf("offset = %d, want 42", offset)
	}
}

func TestRunProcessesEditEvent(t *testing.T) {
	root := t.TempDir()

	origLookPath := lookPathFunc
	origSpawn := spawnFunc
	t.Cleanup(func() {
		lookPathFunc = origLookPath
		spawnFunc = origSpawn
	})

	lookPathFunc = func(name string) (string, error) {
		return "/fake/" + name, nil
	}

	eventsCh := make(chan any, 8)
	doneCh := make(chan error, 1)

	spawnFunc = func(ctx context.Context, provider, binaryPath, root string, providerConfig []byte, offset int64) (*Subprocess, error) {
		return &Subprocess{
			events:   eventsCh,
			done:     doneCh,
			provider: provider,
		}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	msgDate := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	editDate := time.Date(2026, 3, 20, 12, 5, 0, 0, time.UTC)

	go func() {
		// First send the original message.
		eventsCh <- protocol.MessageEvent{
			Type:    protocol.TypeMessage,
			Offset:  10,
			ID:      1,
			ChatID:  12345,
			Channel: "test-channel",
			From:    "alice",
			Date:    msgDate,
			Body:    "original text",
		}
		time.Sleep(50 * time.Millisecond)

		// Then send the edit.
		eventsCh <- protocol.MessageEvent{
			Type:     protocol.TypeEdit,
			Offset:   20,
			ID:       1,
			ChatID:   12345,
			Channel:  "test-channel",
			From:     "alice",
			Date:     msgDate,
			Body:     "edited text",
			EditDate: &editDate,
		}
		time.Sleep(50 * time.Millisecond)
		cancel()
		close(eventsCh)
		doneCh <- nil
	}()

	cfg := config.Config{
		General:   config.GeneralConfig{Format: "markdown"},
		Providers: map[string]map[string]any{"testprov": {"token": "fake"}},
	}

	err := Run(ctx, cfg, root, []string{"testprov"})
	if err != nil && err != context.Canceled {
		t.Fatalf("Run returned unexpected error: %v", err)
	}

	// Verify offset was updated to the edit's offset.
	offset, err := store.ReadOffset(root, "testprov")
	if err != nil {
		t.Fatalf("reading offset: %v", err)
	}
	if offset != 20 {
		t.Errorf("offset = %d, want 20", offset)
	}
}

func TestRunProcessesErrorEvent(t *testing.T) {
	root := t.TempDir()

	origLookPath := lookPathFunc
	origSpawn := spawnFunc
	t.Cleanup(func() {
		lookPathFunc = origLookPath
		spawnFunc = origSpawn
	})

	lookPathFunc = func(name string) (string, error) {
		return "/fake/" + name, nil
	}

	eventsCh := make(chan any, 8)
	doneCh := make(chan error, 1)

	spawnFunc = func(ctx context.Context, provider, binaryPath, root string, providerConfig []byte, offset int64) (*Subprocess, error) {
		return &Subprocess{
			events:   eventsCh,
			done:     doneCh,
			provider: provider,
		}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		eventsCh <- protocol.ErrorEvent{
			Type:    protocol.TypeError,
			Code:    500,
			Message: "test error",
		}
		time.Sleep(50 * time.Millisecond)
		cancel()
		close(eventsCh)
		doneCh <- nil
	}()

	cfg := config.Config{
		General:   config.GeneralConfig{Format: "markdown"},
		Providers: map[string]map[string]any{"testprov": {"token": "fake"}},
	}

	// Should not crash on error events -- just log them.
	err := Run(ctx, cfg, root, []string{"testprov"})
	if err != nil && err != context.Canceled {
		t.Fatalf("Run returned unexpected error: %v", err)
	}
}

func TestRunDownloadsMedia(t *testing.T) {
	root := t.TempDir()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fake image data"))
	}))
	defer ts.Close()

	origLookPath := lookPathFunc
	origSpawn := spawnFunc
	t.Cleanup(func() {
		lookPathFunc = origLookPath
		spawnFunc = origSpawn
	})

	lookPathFunc = func(name string) (string, error) {
		return "/fake/" + name, nil
	}

	eventsCh := make(chan any, 8)
	doneCh := make(chan error, 1)

	spawnFunc = func(ctx context.Context, provider, binaryPath, root string, providerConfig []byte, offset int64) (*Subprocess, error) {
		return &Subprocess{
			events:   eventsCh,
			done:     doneCh,
			provider: provider,
		}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		eventsCh <- protocol.MessageEvent{
			Type:        protocol.TypeMessage,
			Offset:      50,
			ID:          1,
			ChatID:      123,
			Channel:     "general",
			From:        "bob",
			Date:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Body:        "check this photo",
			DownloadURL: ts.URL + "/photo.jpg",
			MediaType:   "photo",
			MediaExt:    ".jpg",
		}
		time.Sleep(100 * time.Millisecond)
		cancel()
		close(eventsCh)
		doneCh <- nil
	}()

	cfg := config.Config{
		General:   config.GeneralConfig{Format: "markdown"},
		Providers: map[string]map[string]any{"testprov": {"token": "fake"}},
	}

	_ = Run(ctx, cfg, root, []string{"testprov"})

	// Verify media file was downloaded into a timestamp subdirectory.
	channelDir := filepath.Join(root, "testprov-general")
	entries, err := os.ReadDir(channelDir)
	if err != nil {
		t.Fatalf("reading channel dir: %v", err)
	}

	foundMedia := false
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		subEntries, err := os.ReadDir(filepath.Join(channelDir, e.Name()))
		if err != nil {
			continue
		}
		for _, se := range subEntries {
			if strings.HasSuffix(se.Name(), ".jpg") {
				foundMedia = true
				data, err := os.ReadFile(filepath.Join(channelDir, e.Name(), se.Name()))
				if err != nil {
					t.Fatalf("reading media file: %v", err)
				}
				if string(data) != "fake image data" {
					t.Errorf("media content = %q, want %q", string(data), "fake image data")
				}
			}
		}
	}
	if !foundMedia {
		t.Error("expected media file (.jpg) in channel directory")
	}

	// Verify offset was persisted.
	offset, err := store.ReadOffset(root, "testprov")
	if err != nil {
		t.Fatalf("reading offset: %v", err)
	}
	if offset != 50 {
		t.Errorf("offset = %d, want 50", offset)
	}
}
