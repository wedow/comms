package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/wedow/comms/internal/provider/telegram"
	"github.com/wedow/comms/internal/store"
)

type mockSendBot struct {
	sendFn func(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error)
}

func (m *mockSendBot) SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error) {
	return m.sendFn(ctx, params)
}

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
}
