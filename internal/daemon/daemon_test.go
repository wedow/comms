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
	"github.com/wedow/comms/internal/message"
	"github.com/wedow/comms/internal/store"
)

type fakeReaction struct {
	channel string
	msgID   int
	from    string
	emoji   string
	date    time.Time
}

type fakeProvider struct {
	messages    []message.Message
	chatIDs     []int64
	isEdits     []bool
	reactions   []fakeReaction
	finalOffset int64
}

func (f *fakeProvider) Poll(ctx context.Context, initialOffset int64, handler func(msg message.Message, chatID int64, isEdit bool), reactionHandler func(channel string, msgID int, from string, emoji string, date time.Time)) (int64, error) {
	for i, msg := range f.messages {
		isEdit := false
		if i < len(f.isEdits) {
			isEdit = f.isEdits[i]
		}
		handler(msg, f.chatIDs[i], isEdit)
	}
	for _, r := range f.reactions {
		reactionHandler(r.channel, r.msgID, r.from, r.emoji, r.date)
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

func TestRunHandlesEditedMessage(t *testing.T) {
	root := t.TempDir()

	// First, write an original message
	original := message.Message{
		From:     "alice",
		Provider: "telegram",
		Channel:  "general",
		Date:     time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		ID:       "telegram-42",
		Body:     "original text",
	}
	if _, err := store.WriteMessage(root, original, "markdown"); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}

	// Advance cursor past the message
	if err := store.WriteCursor(root, "telegram-general", time.Date(2026, 3, 1, 13, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("WriteCursor: %v", err)
	}

	// Now simulate an edit coming through
	editDate := time.Date(2026, 3, 1, 12, 5, 0, 0, time.UTC)
	edited := message.Message{
		From:     "alice",
		Provider: "telegram",
		Channel:  "general",
		Date:     time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		ID:       "telegram-42",
		EditDate: &editDate,
		Body:     "edited text",
	}

	fp := &fakeProvider{
		messages:    []message.Message{edited},
		chatIDs:     []int64{123},
		isEdits:     []bool{true},
		finalOffset: 50,
	}

	ctx := context.Background()
	if err := Run(ctx, testConfig(), root, fp); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify the edit was appended (find the message file)
	paths, err := store.ListMessages(root, "telegram-general")
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("expected 1 message file, got %d", len(paths))
	}

	data, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "---edit---") {
		t.Error("missing ---edit--- in file after edit")
	}
	if !strings.Contains(content, "edited text") {
		t.Error("missing edited text in file")
	}

	// Verify cursor was reset
	cursor, err := store.ReadCursor(root, "telegram-general")
	if err != nil {
		t.Fatalf("ReadCursor: %v", err)
	}
	if !cursor.Before(original.Date) {
		t.Errorf("cursor %v should be before message date %v", cursor, original.Date)
	}
}

func TestRunHandlesReaction(t *testing.T) {
	root := t.TempDir()

	// Write an original message
	original := message.Message{
		From:     "alice",
		Provider: "telegram",
		Channel:  "general",
		Date:     time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		ID:       "telegram-42",
		Body:     "original text",
	}
	if _, err := store.WriteMessage(root, original, "markdown"); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}

	// Advance cursor past the message
	if err := store.WriteCursor(root, "telegram-general", time.Date(2026, 3, 1, 13, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("WriteCursor: %v", err)
	}

	fp := &fakeProvider{
		reactions: []fakeReaction{
			{
				channel: "telegram-general",
				msgID:   42,
				from:    "bob",
				emoji:   "\U0001f44d",
				date:    time.Date(2026, 3, 1, 12, 10, 0, 0, time.UTC),
			},
		},
		finalOffset: 60,
	}

	ctx := context.Background()
	if err := Run(ctx, testConfig(), root, fp); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify reaction was appended
	paths, err := store.ListMessages(root, "telegram-general")
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("expected 1 message file, got %d", len(paths))
	}

	data, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "---reaction---") {
		t.Error("missing ---reaction--- in file after reaction")
	}
	if !strings.Contains(content, "from: bob") {
		t.Error("missing reaction from")
	}
	if !strings.Contains(content, "emoji: \U0001f44d") {
		t.Error("missing reaction emoji")
	}

	// Verify cursor was reset
	cursor, err := store.ReadCursor(root, "telegram-general")
	if err != nil {
		t.Fatalf("ReadCursor: %v", err)
	}
	if !cursor.Before(original.Date) {
		t.Errorf("cursor %v should be before message date %v", cursor, original.Date)
	}
}

func TestRunDownloadsMedia(t *testing.T) {
	// Serve fake media content
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fake-image-bytes"))
	}))
	defer ts.Close()

	root := t.TempDir()
	msg := message.Message{
		From:        "alice",
		Provider:    "telegram",
		Channel:     "general",
		Date:        time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		ID:          "telegram-1",
		Body:        "photo caption",
		MediaType:   "photo",
		MediaFileID: "abc123",
		DownloadURL: ts.URL + "/file/bot/photos/file_0.jpg",
		MediaExt:    ".jpg",
	}

	fp := &fakeProvider{
		messages:    []message.Message{msg},
		chatIDs:     []int64{123},
		finalOffset: 10,
	}

	ctx := context.Background()
	if err := Run(ctx, testConfig(), root, fp); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify media file was downloaded
	chanDir := filepath.Join(root, "telegram-general")
	timestamp := strings.ReplaceAll(msg.Date.Format(time.RFC3339Nano), ":", "-")
	mediaPath := filepath.Join(chanDir, timestamp, "001.jpg")
	data, err := os.ReadFile(mediaPath)
	if err != nil {
		t.Fatalf("media file not found at %s: %v", mediaPath, err)
	}
	if string(data) != "fake-image-bytes" {
		t.Errorf("media content = %q, want %q", data, "fake-image-bytes")
	}

	// Verify message file has media_url set
	paths, err := store.ListMessages(root, "telegram-general")
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("expected 1 message file, got %d", len(paths))
	}
	written, err := store.ReadMessage(paths[0])
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	wantMediaURL := timestamp + "/001.jpg"
	if written.MediaURL != wantMediaURL {
		t.Errorf("MediaURL = %q, want %q", written.MediaURL, wantMediaURL)
	}
}

func TestRunSkipsMediaOnDownloadError(t *testing.T) {
	// Server that returns 500
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	root := t.TempDir()
	msg := message.Message{
		From:        "alice",
		Provider:    "telegram",
		Channel:     "general",
		Date:        time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		ID:          "telegram-1",
		Body:        "photo caption",
		MediaType:   "photo",
		MediaFileID: "abc123",
		DownloadURL: ts.URL + "/file/bot/photos/file_0.jpg",
		MediaExt:    ".jpg",
	}

	fp := &fakeProvider{
		messages:    []message.Message{msg},
		chatIDs:     []int64{123},
		finalOffset: 10,
	}

	ctx := context.Background()
	if err := Run(ctx, testConfig(), root, fp); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Message should still be written (without media_url)
	paths, err := store.ListMessages(root, "telegram-general")
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("expected 1 message file, got %d", len(paths))
	}
	written, err := store.ReadMessage(paths[0])
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if written.MediaURL != "" {
		t.Errorf("MediaURL = %q, want empty on download error", written.MediaURL)
	}
}

func TestAllowedIDsAcceptsAllWhenEmpty(t *testing.T) {
	root := t.TempDir()
	// No allowed_ids file — should accept all messages
	msg := testMessage("alice", "general", "hello")
	fp := &fakeProvider{
		messages:    []message.Message{msg},
		chatIDs:     []int64{123},
		finalOffset: 1,
	}

	if err := Run(context.Background(), testConfig(), root, fp); err != nil {
		t.Fatalf("Run: %v", err)
	}

	paths, _ := store.ListMessages(root, "telegram-general")
	if len(paths) != 1 {
		t.Errorf("expected 1 message (no allowlist), got %d", len(paths))
	}
}

func TestAllowedIDsAcceptsPermittedChat(t *testing.T) {
	root := t.TempDir()
	store.AddAllowedID(root, 123)

	msg := testMessage("alice", "general", "hello")
	fp := &fakeProvider{
		messages:    []message.Message{msg},
		chatIDs:     []int64{123},
		finalOffset: 1,
	}

	if err := Run(context.Background(), testConfig(), root, fp); err != nil {
		t.Fatalf("Run: %v", err)
	}

	paths, _ := store.ListMessages(root, "telegram-general")
	if len(paths) != 1 {
		t.Errorf("expected 1 message (allowed chat), got %d", len(paths))
	}
}

func TestAllowedIDsRejectsUnpermittedChat(t *testing.T) {
	root := t.TempDir()
	store.AddAllowedID(root, 999)

	msg := testMessage("alice", "general", "hello")
	fp := &fakeProvider{
		messages:    []message.Message{msg},
		chatIDs:     []int64{123}, // not in allowlist
		finalOffset: 1,
	}

	if err := Run(context.Background(), testConfig(), root, fp); err != nil {
		t.Fatalf("Run: %v", err)
	}

	paths, _ := store.ListMessages(root, "telegram-general")
	if len(paths) != 0 {
		t.Errorf("expected 0 messages (rejected chat), got %d", len(paths))
	}
}

func TestAllowedIDsRejectsReactionFromUnpermittedChat(t *testing.T) {
	root := t.TempDir()
	store.AddAllowedID(root, 999)

	// Write an original message and chat_id for the channel
	original := message.Message{
		From:     "alice",
		Provider: "telegram",
		Channel:  "general",
		Date:     time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		ID:       "telegram-42",
		Body:     "original text",
	}
	store.WriteMessage(root, original, "markdown")
	store.WriteChatID(root, "telegram-general", 123) // not in allowlist

	fp := &fakeProvider{
		reactions: []fakeReaction{
			{
				channel: "telegram-general",
				msgID:   42,
				from:    "bob",
				emoji:   "\U0001f44d",
				date:    time.Date(2026, 3, 1, 12, 10, 0, 0, time.UTC),
			},
		},
		finalOffset: 1,
	}

	if err := Run(context.Background(), testConfig(), root, fp); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Reaction should NOT be appended
	paths, _ := store.ListMessages(root, "telegram-general")
	if len(paths) != 1 {
		t.Fatalf("expected 1 message file, got %d", len(paths))
	}
	data, _ := os.ReadFile(paths[0])
	if strings.Contains(string(data), "---reaction---") {
		t.Error("reaction should have been rejected for unpermitted chat")
	}
}
