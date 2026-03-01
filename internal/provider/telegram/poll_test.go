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

func TestPollProcessesChannelPost(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var got message.Message
	var gotChatID int64
	var gotIsEdit bool
	called := false
	handler := func(msg message.Message, chatID int64, isEdit bool) {
		got = msg
		gotChatID = chatID
		gotIsEdit = isEdit
		called = true
	}

	testBot, err := bot.New("test-token",
		bot.WithSkipGetMe(),
		bot.WithDefaultHandler(func(_ context.Context, _ *bot.Bot, update *models.Update) {
			if update.EditedMessage != nil {
				msg := convertMessage(update.EditedMessage)
				handler(msg, update.EditedMessage.Chat.ID, true)
				return
			}
			if update.ChannelPost != nil {
				msg := convertMessage(update.ChannelPost)
				handler(msg, update.ChannelPost.Chat.ID, false)
				return
			}
			if update.Message == nil {
				return
			}
			msg := convertMessage(update.Message)
			handler(msg, update.Message.Chat.ID, false)
		}),
	)
	if err != nil {
		t.Fatalf("bot.New: %v", err)
	}

	testBot.ProcessUpdate(ctx, &models.Update{
		ID: 300,
		ChannelPost: &models.Message{
			ID:         55,
			Date:       1709312400,
			Chat:       models.Chat{ID: 777, Type: models.ChatTypeChannel, Title: "News"},
			SenderChat: &models.Chat{Username: "newschannel", Title: "News"},
			Text:       "breaking news",
		},
	})

	time.Sleep(50 * time.Millisecond)

	if !called {
		t.Fatal("handler was not called for ChannelPost")
	}
	if got.From != "newschannel" {
		t.Errorf("From = %q, want %q", got.From, "newschannel")
	}
	if got.Provider != "telegram" {
		t.Errorf("Provider = %q, want %q", got.Provider, "telegram")
	}
	if got.Channel != "news" {
		t.Errorf("Channel = %q, want %q", got.Channel, "news")
	}
	if got.ID != "telegram-55" {
		t.Errorf("ID = %q, want %q", got.ID, "telegram-55")
	}
	if got.Body != "breaking news" {
		t.Errorf("Body = %q, want %q", got.Body, "breaking news")
	}
	if gotChatID != 777 {
		t.Errorf("chatID = %d, want %d", gotChatID, 777)
	}
	if gotIsEdit {
		t.Error("isEdit = true, want false")
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
