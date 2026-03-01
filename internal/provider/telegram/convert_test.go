package telegram

import (
	"testing"
	"time"

	"github.com/go-telegram/bot/models"
)

func TestConvertMessage(t *testing.T) {
	tests := []struct {
		name    string
		msg     *models.Message
		wantFrom    string
		wantProvider string
		wantChannel string
		wantDate    time.Time
		wantID      string
		wantBody    string
	}{
		{
			name: "user in group",
			msg: &models.Message{
				ID:   42,
				From: &models.User{Username: "alice"},
				Date: 1709312400,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Dev Team"},
				Text: "hello world",
			},
			wantFrom:     "alice",
			wantProvider: "telegram",
			wantChannel:  "dev-team",
			wantDate:     time.Unix(1709312400, 0).UTC(),
			wantID:       "telegram-42",
			wantBody:     "hello world",
		},
		{
			name: "user in DM",
			msg: &models.Message{
				ID:   7,
				From: &models.User{Username: "bob"},
				Date: 1000000,
				Chat: models.Chat{Type: models.ChatTypePrivate, Username: "bob"},
				Text: "hi",
			},
			wantFrom:     "bob",
			wantProvider: "telegram",
			wantChannel:  "bob",
			wantDate:     time.Unix(1000000, 0).UTC(),
			wantID:       "telegram-7",
			wantBody:     "hi",
		},
		{
			name: "empty text",
			msg: &models.Message{
				ID:   1,
				From: &models.User{Username: "carol"},
				Date: 100,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "General"},
			},
			wantFrom:     "carol",
			wantProvider: "telegram",
			wantChannel:  "general",
			wantDate:     time.Unix(100, 0).UTC(),
			wantID:       "telegram-1",
			wantBody:     "",
		},
		{
			name: "unix timestamp zero",
			msg: &models.Message{
				ID:   5,
				From: &models.User{Username: "dave"},
				Date: 0,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Test"},
			},
			wantFrom:     "dave",
			wantProvider: "telegram",
			wantChannel:  "test",
			wantDate:     time.Unix(0, 0).UTC(),
			wantID:       "telegram-5",
			wantBody:     "",
		},
		{
			name: "from nil",
			msg: &models.Message{
				ID:   99,
				Date: 500,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Ops"},
				Text: "system msg",
			},
			wantFrom:     "unknown",
			wantProvider: "telegram",
			wantChannel:  "ops",
			wantDate:     time.Unix(500, 0).UTC(),
			wantID:       "telegram-99",
			wantBody:     "system msg",
		},
		{
			name: "from with no username falls back to first name",
			msg: &models.Message{
				ID:   10,
				From: &models.User{FirstName: "Eve"},
				Date: 200,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Chat"},
				Text: "hey",
			},
			wantFrom:     "Eve",
			wantProvider: "telegram",
			wantChannel:  "chat",
			wantDate:     time.Unix(200, 0).UTC(),
			wantID:       "telegram-10",
			wantBody:     "hey",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertMessage(tt.msg)
			if got.From != tt.wantFrom {
				t.Errorf("From = %q, want %q", got.From, tt.wantFrom)
			}
			if got.Provider != tt.wantProvider {
				t.Errorf("Provider = %q, want %q", got.Provider, tt.wantProvider)
			}
			if got.Channel != tt.wantChannel {
				t.Errorf("Channel = %q, want %q", got.Channel, tt.wantChannel)
			}
			if !got.Date.Equal(tt.wantDate) {
				t.Errorf("Date = %v, want %v", got.Date, tt.wantDate)
			}
			if got.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", got.ID, tt.wantID)
			}
			if got.Body != tt.wantBody {
				t.Errorf("Body = %q, want %q", got.Body, tt.wantBody)
			}
		})
	}
}
