package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/wedow/comms/internal/provider/telegram"
	"github.com/wedow/comms/internal/store"
)

type mockReactBot struct {
	sendFn func(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error)
	reactFn func(ctx context.Context, params *bot.SetMessageReactionParams) (bool, error)
}

func (m *mockReactBot) SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error) {
	if m.sendFn != nil {
		return m.sendFn(ctx, params)
	}
	return nil, errors.New("SendMessage not expected")
}

func (m *mockReactBot) SetMessageReaction(ctx context.Context, params *bot.SetMessageReactionParams) (bool, error) {
	return m.reactFn(ctx, params)
}

func mockReactBotFactory(b telegram.BotAPI) func(string) (telegram.BotAPI, error) {
	return func(_ string) (telegram.BotAPI, error) { return b, nil }
}

func TestReactCmd(t *testing.T) {
	t.Run("successful reaction", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		var gotParams *bot.SetMessageReactionParams
		mock := &mockReactBot{reactFn: func(_ context.Context, p *bot.SetMessageReactionParams) (bool, error) {
			gotParams = p
			return true, nil
		}}

		cmd := newReactCmd(mockReactBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetArgs([]string{"--channel", "general", "--message", "telegram-99", "--emoji", "\U0001F44D", "--dir", root})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v (stderr: %s)", err, errBuf.String())
		}

		if gotParams == nil {
			t.Fatal("SetMessageReaction was not called")
		}
		if gotParams.ChatID != int64(123) {
			t.Errorf("ChatID = %v, want 123", gotParams.ChatID)
		}
		if gotParams.MessageID != 99 {
			t.Errorf("MessageID = %d, want 99", gotParams.MessageID)
		}
		if len(gotParams.Reaction) != 1 {
			t.Fatalf("Reaction count = %d, want 1", len(gotParams.Reaction))
		}
		if gotParams.Reaction[0].ReactionTypeEmoji == nil {
			t.Fatal("ReactionTypeEmoji is nil")
		}
		if gotParams.Reaction[0].ReactionTypeEmoji.Emoji != "\U0001F44D" {
			t.Errorf("Emoji = %q, want thumbs up", gotParams.Reaction[0].ReactionTypeEmoji.Emoji)
		}

		got := out.String()
		if !strings.Contains(got, `"ok":true`) {
			t.Errorf("stdout = %q, want ok:true", got)
		}
		if !strings.Contains(got, `"channel":"general"`) {
			t.Errorf("stdout = %q, want channel:general", got)
		}
	})

	t.Run("api error", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		mock := &mockReactBot{reactFn: func(_ context.Context, _ *bot.SetMessageReactionParams) (bool, error) {
			return false, errors.New("reaction failed")
		}}

		cmd := newReactCmd(mockReactBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetArgs([]string{"--channel", "general", "--message", "telegram-99", "--emoji", "\U0001F44D", "--dir", root})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for API failure")
		}
		if !strings.Contains(errBuf.String(), `"error"`) {
			t.Errorf("stderr = %q, want JSON error", errBuf.String())
		}
	})

	t.Run("invalid message ID", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		mock := &mockReactBot{reactFn: func(_ context.Context, _ *bot.SetMessageReactionParams) (bool, error) {
			t.Fatal("SetMessageReaction should not be called with invalid message ID")
			return false, nil
		}}

		cmd := newReactCmd(mockReactBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetArgs([]string{"--channel", "general", "--message", "bad-id", "--emoji", "\U0001F44D", "--dir", root})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for invalid message ID")
		}
		if !strings.Contains(errBuf.String(), `"error"`) {
			t.Errorf("stderr = %q, want JSON error", errBuf.String())
		}
	})

	t.Run("missing channel", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "general", 123)

		mock := &mockReactBot{reactFn: func(_ context.Context, _ *bot.SetMessageReactionParams) (bool, error) {
			t.Fatal("SetMessageReaction should not be called with missing channel")
			return false, nil
		}}

		cmd := newReactCmd(mockReactBotFactory(mock))
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetArgs([]string{"--message", "telegram-99", "--emoji", "\U0001F44D", "--dir", root})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for missing channel flag")
		}
	})
}
