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

		got, err := Send(context.Background(), m, 123, "hello")
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

		_, err := Send(context.Background(), m, 456, "fail")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("nil response", func(t *testing.T) {
		m := &mockBot{sendFn: func(_ context.Context, _ *bot.SendMessageParams) (*models.Message, error) {
			return nil, nil
		}}

		_, err := Send(context.Background(), m, 789, "nil")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
