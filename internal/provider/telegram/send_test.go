package telegram

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type mockBot struct {
	sendFn func(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error)
}

func (m *mockBot) SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error) {
	return m.sendFn(ctx, params)
}

func TestSend(t *testing.T) {
	t.Run("successful send", func(t *testing.T) {
		m := &mockBot{sendFn: func(_ context.Context, _ *bot.SendMessageParams) (*models.Message, error) {
			return &models.Message{
				ID:   42,
				From: &models.User{Username: "botuser"},
				Date: 1709312400,
				Chat: models.Chat{ID: 123, Type: models.ChatTypeGroup, Title: "Dev Team"},
				Text: "hello",
			}, nil
		}}

		got, err := Send(context.Background(), m, 123, "hello", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.From != "botuser" {
			t.Errorf("From = %q, want %q", got.From, "botuser")
		}
		if got.Provider != "telegram" {
			t.Errorf("Provider = %q, want %q", got.Provider, "telegram")
		}
		if got.Channel != "dev-team" {
			t.Errorf("Channel = %q, want %q", got.Channel, "dev-team")
		}
		if !got.Date.Equal(time.Unix(1709312400, 0).UTC()) {
			t.Errorf("Date = %v, want %v", got.Date, time.Unix(1709312400, 0).UTC())
		}
		if got.ID != "telegram-42" {
			t.Errorf("ID = %q, want %q", got.ID, "telegram-42")
		}
		if got.Body != "hello" {
			t.Errorf("Body = %q, want %q", got.Body, "hello")
		}
	})

	t.Run("api error", func(t *testing.T) {
		m := &mockBot{sendFn: func(_ context.Context, _ *bot.SendMessageParams) (*models.Message, error) {
			return nil, errors.New("network timeout")
		}}

		_, err := Send(context.Background(), m, 456, "fail", 0)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("nil response", func(t *testing.T) {
		m := &mockBot{sendFn: func(_ context.Context, _ *bot.SendMessageParams) (*models.Message, error) {
			return nil, nil
		}}

		_, err := Send(context.Background(), m, 789, "nil", 0)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("with reply", func(t *testing.T) {
		var gotParams *bot.SendMessageParams
		m := &mockBot{sendFn: func(_ context.Context, p *bot.SendMessageParams) (*models.Message, error) {
			gotParams = p
			return &models.Message{ID: 99, Chat: models.Chat{ID: 123}, Text: "reply"}, nil
		}}

		_, err := Send(context.Background(), m, 123, "reply", 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotParams.ReplyParameters == nil {
			t.Fatal("ReplyParameters is nil, want non-nil")
		}
		if gotParams.ReplyParameters.MessageID != 42 {
			t.Errorf("ReplyParameters.MessageID = %d, want 42", gotParams.ReplyParameters.MessageID)
		}
	})

	t.Run("without reply", func(t *testing.T) {
		var gotParams *bot.SendMessageParams
		m := &mockBot{sendFn: func(_ context.Context, p *bot.SendMessageParams) (*models.Message, error) {
			gotParams = p
			return &models.Message{ID: 100, Chat: models.Chat{ID: 123}, Text: "no reply"}, nil
		}}

		_, err := Send(context.Background(), m, 123, "no reply", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotParams.ReplyParameters != nil {
			t.Errorf("ReplyParameters = %+v, want nil", gotParams.ReplyParameters)
		}
	})
}

func TestParseMessageID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"valid", "telegram-42", 42, false},
		{"no prefix", "abc", 0, true},
		{"empty suffix", "telegram-", 0, true},
		{"non-numeric suffix", "telegram-abc", 0, true},
		{"empty string", "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMessageID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseMessageID(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseMessageID(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
