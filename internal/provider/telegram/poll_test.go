package telegram

import (
	"context"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/wedow/comms/internal/message"
)

func TestPollProcessesTextMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var got message.Message
	var gotChatID int64
	called := false
	handler := func(msg message.Message, chatID int64) {
		got = msg
		gotChatID = chatID
		called = true
	}

	testBot, err := bot.New("test-token",
		bot.WithSkipGetMe(),
		bot.WithDefaultHandler(func(_ context.Context, _ *bot.Bot, update *models.Update) {
			if update.Message == nil {
				return
			}
			msg := convertMessage(update.Message)
			handler(msg, update.Message.Chat.ID)
		}),
	)
	if err != nil {
		t.Fatalf("bot.New: %v", err)
	}

	testBot.ProcessUpdate(ctx, &models.Update{
		ID: 100,
		Message: &models.Message{
			ID:   42,
			From: &models.User{Username: "alice"},
			Date: 1709312400,
			Chat: models.Chat{ID: 999, Type: models.ChatTypeGroup, Title: "Dev Team"},
			Text: "hello",
		},
	})

	// ProcessUpdate runs handler async by default; give it a moment
	time.Sleep(50 * time.Millisecond)

	if !called {
		t.Fatal("handler was not called")
	}
	if got.From != "alice" {
		t.Errorf("From = %q, want %q", got.From, "alice")
	}
	if got.Provider != "telegram" {
		t.Errorf("Provider = %q, want %q", got.Provider, "telegram")
	}
	if got.Channel != "dev-team" {
		t.Errorf("Channel = %q, want %q", got.Channel, "dev-team")
	}
	if got.ID != "telegram-42" {
		t.Errorf("ID = %q, want %q", got.ID, "telegram-42")
	}
	if got.Body != "hello" {
		t.Errorf("Body = %q, want %q", got.Body, "hello")
	}
	if gotChatID != 999 {
		t.Errorf("chatID = %d, want %d", gotChatID, 999)
	}
}

func TestPollSkipsNilMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	called := false
	handler := func(msg message.Message, chatID int64) {
		called = true
	}

	testBot, err := bot.New("test-token",
		bot.WithSkipGetMe(),
		bot.WithDefaultHandler(func(_ context.Context, _ *bot.Bot, update *models.Update) {
			if update.Message == nil {
				return
			}
			msg := convertMessage(update.Message)
			handler(msg, update.Message.Chat.ID)
		}),
	)
	if err != nil {
		t.Fatalf("bot.New: %v", err)
	}

	testBot.ProcessUpdate(ctx, &models.Update{
		ID: 200,
		// Message is nil - e.g. an edited_message or callback_query update
	})

	time.Sleep(50 * time.Millisecond)

	if called {
		t.Error("handler should not be called for nil Message")
	}
}
