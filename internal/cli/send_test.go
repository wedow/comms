package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/wedow/comms/internal/message"
	"github.com/wedow/comms/internal/provider/telegram"
	"github.com/wedow/comms/internal/store"
)

type mockSendBot struct {
	sendFn      func(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error)
	sendPhotoFn func(ctx context.Context, params *bot.SendPhotoParams) (*models.Message, error)
	sendDocFn   func(ctx context.Context, params *bot.SendDocumentParams) (*models.Message, error)
	sendAudioFn func(ctx context.Context, params *bot.SendAudioParams) (*models.Message, error)
	sendVideoFn func(ctx context.Context, params *bot.SendVideoParams) (*models.Message, error)
	sendVoiceFn func(ctx context.Context, params *bot.SendVoiceParams) (*models.Message, error)
	sendAnimFn  func(ctx context.Context, params *bot.SendAnimationParams) (*models.Message, error)
}

func (m *mockSendBot) SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error) {
	return m.sendFn(ctx, params)
}

func (m *mockSendBot) SendPhoto(ctx context.Context, params *bot.SendPhotoParams) (*models.Message, error) {
	if m.sendPhotoFn != nil {
		return m.sendPhotoFn(ctx, params)
	}
	return nil, nil
}

func (m *mockSendBot) SendDocument(ctx context.Context, params *bot.SendDocumentParams) (*models.Message, error) {
	if m.sendDocFn != nil {
		return m.sendDocFn(ctx, params)
	}
	return nil, nil
}

func (m *mockSendBot) SendAudio(ctx context.Context, params *bot.SendAudioParams) (*models.Message, error) {
	if m.sendAudioFn != nil {
		return m.sendAudioFn(ctx, params)
	}
	return nil, nil
}

func (m *mockSendBot) SendVideo(ctx context.Context, params *bot.SendVideoParams) (*models.Message, error) {
	if m.sendVideoFn != nil {
		return m.sendVideoFn(ctx, params)
	}
	return nil, nil
}

func (m *mockSendBot) SendVoice(ctx context.Context, params *bot.SendVoiceParams) (*models.Message, error) {
	if m.sendVoiceFn != nil {
		return m.sendVoiceFn(ctx, params)
	}
	return nil, nil
}

func (m *mockSendBot) SendAnimation(ctx context.Context, params *bot.SendAnimationParams) (*models.Message, error) {
	if m.sendAnimFn != nil {
		return m.sendAnimFn(ctx, params)
	}
	return nil, nil
}

func (m *mockSendBot) SetMessageReaction(_ context.Context, _ *bot.SetMessageReactionParams) (bool, error) {
	return false, nil
}

func (m *mockSendBot) GetFile(_ context.Context, _ *bot.GetFileParams) (*models.File, error) {
	return nil, nil
}

func (m *mockSendBot) FileDownloadLink(_ *models.File) string { return "" }

func mockBotFactory(b telegram.BotAPI) func(string) (telegram.BotAPI, error) {
	return func(_ string) (telegram.BotAPI, error) { return b, nil }
}

func writeTestConfig(t *testing.T, root string) {
	t.Helper()
	data := []byte("[telegram]\ntoken = \"test-token\"\n")
	if err := os.WriteFile(filepath.Join(root, "config.toml"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSendCmd(t *testing.T) {
	t.Run("successful send", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		mock := &mockSendBot{sendFn: func(_ context.Context, p *bot.SendMessageParams) (*models.Message, error) {
			return &models.Message{
				ID:   1,
				Chat: models.Chat{ID: 123},
				Text: p.Text,
			}, nil
		}}

		cmd := newSendCmd(mockBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("hello world"))
		cmd.SetArgs([]string{"--channel", "general", "--dir", root})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := out.String()
		if !strings.Contains(got, `"ok":true`) {
			t.Errorf("stdout = %q, want ok:true", got)
		}
		if !strings.Contains(got, `"channel":"general"`) {
			t.Errorf("stdout = %q, want channel:general", got)
		}
	})

	t.Run("positional args as message body", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		var gotText string
		mock := &mockSendBot{sendFn: func(_ context.Context, p *bot.SendMessageParams) (*models.Message, error) {
			gotText = p.Text
			return &models.Message{
				ID:   1,
				Chat: models.Chat{ID: 123},
				Text: p.Text,
			}, nil
		}}

		cmd := newSendCmd(mockBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader(""))
		cmd.SetArgs([]string{"--channel", "general", "--dir", root, "hello", "world!"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotText != "hello world!" {
			t.Errorf("text = %q, want %q", gotText, "hello world!")
		}
	})

	t.Run("empty stdin", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		mock := &mockSendBot{sendFn: func(_ context.Context, _ *bot.SendMessageParams) (*models.Message, error) {
			t.Fatal("SendMessage should not be called for empty stdin")
			return nil, nil
		}}

		cmd := newSendCmd(mockBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader(""))
		cmd.SetArgs([]string{"--channel", "general", "--dir", root})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for empty stdin")
		}
		if !strings.Contains(errBuf.String(), `"error"`) {
			t.Errorf("stderr = %q, want JSON error", errBuf.String())
		}
	})

	t.Run("missing chat_id", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)

		mock := &mockSendBot{sendFn: func(_ context.Context, _ *bot.SendMessageParams) (*models.Message, error) {
			t.Fatal("SendMessage should not be called when chat_id is missing")
			return nil, nil
		}}

		cmd := newSendCmd(mockBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("hello"))
		cmd.SetArgs([]string{"--channel", "nonexistent", "--dir", root})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for missing chat_id")
		}
		if !strings.Contains(errBuf.String(), `"error"`) {
			t.Errorf("stderr = %q, want JSON error", errBuf.String())
		}
	})

	t.Run("api error", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		mock := &mockSendBot{sendFn: func(_ context.Context, _ *bot.SendMessageParams) (*models.Message, error) {
			return nil, errors.New("network timeout")
		}}

		cmd := newSendCmd(mockBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("hello"))
		cmd.SetArgs([]string{"--channel", "general", "--dir", root})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for API failure")
		}
		if !strings.Contains(errBuf.String(), `"error"`) {
			t.Errorf("stderr = %q, want JSON error", errBuf.String())
		}
	})

	t.Run("reply-to sets ReplyParameters", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		var gotParams *bot.SendMessageParams
		mock := &mockSendBot{sendFn: func(_ context.Context, p *bot.SendMessageParams) (*models.Message, error) {
			gotParams = p
			return &models.Message{ID: 1, Chat: models.Chat{ID: 123}, Text: p.Text}, nil
		}}

		cmd := newSendCmd(mockBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("reply text"))
		cmd.SetArgs([]string{"--channel", "general", "--dir", root, "--reply-to", "telegram-99"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotParams.ReplyParameters == nil {
			t.Fatal("ReplyParameters is nil, want non-nil")
		}
		if gotParams.ReplyParameters.MessageID != 99 {
			t.Errorf("ReplyParameters.MessageID = %d, want 99", gotParams.ReplyParameters.MessageID)
		}
	})

	t.Run("invalid reply-to format", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		mock := &mockSendBot{sendFn: func(_ context.Context, _ *bot.SendMessageParams) (*models.Message, error) {
			t.Fatal("SendMessage should not be called with invalid reply-to")
			return nil, nil
		}}

		cmd := newSendCmd(mockBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("hello"))
		cmd.SetArgs([]string{"--channel", "general", "--dir", root, "--reply-to", "bad-id"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for invalid reply-to")
		}
		if !strings.Contains(errBuf.String(), `"error"`) {
			t.Errorf("stderr = %q, want JSON error", errBuf.String())
		}
	})

	t.Run("file flag sends media", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		// Create a temp file to send
		tmpFile := filepath.Join(t.TempDir(), "photo.jpg")
		os.WriteFile(tmpFile, []byte("fake image data"), 0o644)

		var gotCaption string
		mock := &mockSendBot{sendPhotoFn: func(_ context.Context, p *bot.SendPhotoParams) (*models.Message, error) {
			gotCaption = p.Caption
			return &models.Message{ID: 1, Chat: models.Chat{ID: 123}}, nil
		}}

		cmd := newSendCmd(mockBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("my caption"))
		cmd.SetArgs([]string{"--channel", "general", "--dir", root, "--file", tmpFile})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v (stderr: %s)", err, errBuf.String())
		}
		if gotCaption != "my caption" {
			t.Errorf("caption = %q, want %q", gotCaption, "my caption")
		}
		if !strings.Contains(out.String(), `"ok":true`) {
			t.Errorf("stdout = %q, want ok:true", out.String())
		}
	})

	t.Run("file flag with empty caption", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		tmpFile := filepath.Join(t.TempDir(), "doc.pdf")
		os.WriteFile(tmpFile, []byte("fake pdf"), 0o644)

		called := false
		mock := &mockSendBot{sendDocFn: func(_ context.Context, p *bot.SendDocumentParams) (*models.Message, error) {
			called = true
			if p.Caption != "" {
				t.Errorf("caption = %q, want empty", p.Caption)
			}
			return &models.Message{ID: 1, Chat: models.Chat{ID: 123}}, nil
		}}

		cmd := newSendCmd(mockBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader(""))
		cmd.SetArgs([]string{"--channel", "general", "--dir", root, "--file", tmpFile})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v (stderr: %s)", err, errBuf.String())
		}
		if !called {
			t.Fatal("SendDocument was not called")
		}
	})

	t.Run("media-type override", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		tmpFile := filepath.Join(t.TempDir(), "data.bin")
		os.WriteFile(tmpFile, []byte("binary data"), 0o644)

		called := false
		mock := &mockSendBot{sendVideoFn: func(_ context.Context, _ *bot.SendVideoParams) (*models.Message, error) {
			called = true
			return &models.Message{ID: 1, Chat: models.Chat{ID: 123}}, nil
		}}

		cmd := newSendCmd(mockBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader(""))
		cmd.SetArgs([]string{"--channel", "general", "--dir", root, "--file", tmpFile, "--media-type", "video"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v (stderr: %s)", err, errBuf.String())
		}
		if !called {
			t.Fatal("SendVideo was not called for media-type override")
		}
	})

	t.Run("thread flag sets MessageThreadID", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		var gotParams *bot.SendMessageParams
		mock := &mockSendBot{sendFn: func(_ context.Context, p *bot.SendMessageParams) (*models.Message, error) {
			gotParams = p
			return &models.Message{ID: 1, Chat: models.Chat{ID: 123}, Text: p.Text}, nil
		}}

		cmd := newSendCmd(mockBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("topic message"))
		cmd.SetArgs([]string{"--channel", "general", "--dir", root, "--thread", "42"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotParams.MessageThreadID != 42 {
			t.Errorf("MessageThreadID = %d, want 42", gotParams.MessageThreadID)
		}
	})

	t.Run("format markdown sets parse mode", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		var gotParams *bot.SendMessageParams
		mock := &mockSendBot{sendFn: func(_ context.Context, p *bot.SendMessageParams) (*models.Message, error) {
			gotParams = p
			return &models.Message{ID: 1, Chat: models.Chat{ID: 123}, Text: p.Text}, nil
		}}

		cmd := newSendCmd(mockBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("*bold*"))
		cmd.SetArgs([]string{"--channel", "general", "--dir", root, "--format", "markdown"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotParams.ParseMode != models.ParseModeMarkdown {
			t.Errorf("ParseMode = %q, want %q", gotParams.ParseMode, models.ParseModeMarkdown)
		}
	})

	t.Run("format html sets parse mode", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		var gotParams *bot.SendMessageParams
		mock := &mockSendBot{sendFn: func(_ context.Context, p *bot.SendMessageParams) (*models.Message, error) {
			gotParams = p
			return &models.Message{ID: 1, Chat: models.Chat{ID: 123}, Text: p.Text}, nil
		}}

		cmd := newSendCmd(mockBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("<b>bold</b>"))
		cmd.SetArgs([]string{"--channel", "general", "--dir", root, "--format", "html"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotParams.ParseMode != models.ParseModeHTML {
			t.Errorf("ParseMode = %q, want %q", gotParams.ParseMode, models.ParseModeHTML)
		}
	})

	t.Run("format invalid returns error", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		mock := &mockSendBot{sendFn: func(_ context.Context, _ *bot.SendMessageParams) (*models.Message, error) {
			t.Fatal("SendMessage should not be called with invalid format")
			return nil, nil
		}}

		cmd := newSendCmd(mockBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("hello"))
		cmd.SetArgs([]string{"--channel", "general", "--dir", root, "--format", "rtf"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for invalid format")
		}
	})

	t.Run("writes local message file", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		mock := &mockSendBot{sendFn: func(_ context.Context, p *bot.SendMessageParams) (*models.Message, error) {
			return &models.Message{
				ID:   42,
				Date: 1709308800, // 2024-03-01T16:00:00Z
				Chat: models.Chat{ID: 123, Type: "private"},
				Text: p.Text,
				From: &models.User{Username: "mybot"},
			}, nil
		}}

		cmd := newSendCmd(mockBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("saved locally"))
		cmd.SetArgs([]string{"--channel", "general", "--dir", root})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v (stderr: %s)", err, errBuf.String())
		}

		// Check that a .md file was written in the channel directory
		channelDir := filepath.Join(root, "telegram-dm-123")
		entries, err := os.ReadDir(channelDir)
		if err != nil {
			t.Fatalf("channel dir not created: %v", err)
		}
		found := false
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".md") {
				found = true
				msg, err := store.ReadMessage(filepath.Join(channelDir, e.Name()))
				if err != nil {
					t.Fatalf("read message: %v", err)
				}
				if msg.From != "mybot" {
					t.Errorf("from = %q, want mybot", msg.From)
				}
				if msg.Body != "saved locally" {
					t.Errorf("body = %q, want 'saved locally'", msg.Body)
				}
			}
		}
		if !found {
			t.Error("no .md file found in channel directory")
		}
	})

	t.Run("advances cursor when no unreads", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		mock := &mockSendBot{sendFn: func(_ context.Context, p *bot.SendMessageParams) (*models.Message, error) {
			return &models.Message{
				ID:   50,
				Date: 1709308800,
				Chat: models.Chat{ID: 123, Type: "private"},
				Text: p.Text,
				From: &models.User{Username: "mybot"},
			}, nil
		}}

		cmd := newSendCmd(mockBotFactory(mock))
		out := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetIn(strings.NewReader("hello"))
		cmd.SetArgs([]string{"--channel", "general", "--dir", root})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Cursor should have advanced past the sent message
		cursor, err := store.ReadCursor(root, "telegram-dm-123")
		if err != nil {
			t.Fatalf("read cursor: %v", err)
		}
		if cursor.IsZero() {
			t.Fatal("cursor should be set after send with no unreads")
		}

		// Unread should return nothing
		unreads, _ := store.ListMessagesAfter(root, "telegram-dm-123", cursor)
		if len(unreads) != 0 {
			t.Errorf("expected 0 unreads, got %d", len(unreads))
		}
	})

	t.Run("preserves cursor when unreads exist", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		// Write an existing unread message
		// Provider + Channel must match what telegram.Send returns:
		// Provider="telegram", Channel="dm-123" → dir "telegram-dm-123"
		existing := message.Message{
			From:     "someone",
			Provider: "telegram",
			Channel:  "dm-123",
			Date:     time.Date(2024, 3, 1, 15, 0, 0, 0, time.UTC),
			ID:       "telegram-49",
			Body:     "hey there",
		}
		store.WriteMessage(root, existing, "markdown")

		mock := &mockSendBot{sendFn: func(_ context.Context, p *bot.SendMessageParams) (*models.Message, error) {
			return &models.Message{
				ID:   50,
				Date: 1709308800, // 2024-03-01T16:00:00Z (after the existing msg)
				Chat: models.Chat{ID: 123, Type: "private"},
				Text: p.Text,
				From: &models.User{Username: "mybot"},
			}, nil
		}}

		cmd := newSendCmd(mockBotFactory(mock))
		out := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetIn(strings.NewReader("reply"))
		cmd.SetArgs([]string{"--channel", "general", "--dir", root})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Cursor should still be zero (not advanced)
		cursor, err := store.ReadCursor(root, "telegram-dm-123")
		if err != nil {
			t.Fatalf("read cursor: %v", err)
		}
		if !cursor.IsZero() {
			t.Errorf("cursor should be zero when unreads exist, got %v", cursor)
		}

		// Both messages should be unread
		unreads, _ := store.ListMessagesAfter(root, "telegram-dm-123", cursor)
		if len(unreads) != 2 {
			t.Errorf("expected 2 unreads, got %d", len(unreads))
		}
	})

	t.Run("file not found", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		mock := &mockSendBot{}

		cmd := newSendCmd(mockBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader(""))
		cmd.SetArgs([]string{"--channel", "general", "--dir", root, "--file", "/nonexistent/file.jpg"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for missing file")
		}
		if !strings.Contains(errBuf.String(), `"error"`) {
			t.Errorf("stderr = %q, want JSON error", errBuf.String())
		}
	})
}
