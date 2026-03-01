package telegram

import (
	"testing"

	"github.com/go-telegram/bot/models"
)

func TestSlugifyChat(t *testing.T) {
	tests := []struct {
		name string
		chat models.Chat
		want string
	}{
		{
			name: "group with title",
			chat: models.Chat{Type: models.ChatTypeGroup, Title: "My Cool Group"},
			want: "my-cool-group",
		},
		{
			name: "title with special chars",
			chat: models.Chat{Type: models.ChatTypeGroup, Title: "Alerts!!! & Stuff"},
			want: "alerts-stuff",
		},
		{
			name: "title with unicode",
			chat: models.Chat{Type: models.ChatTypeSupergroup, Title: "Tes\u0301t Gro\u0308up"},
			want: "tes-t-gro-up",
		},
		{
			name: "private chat with username",
			chat: models.Chat{Type: models.ChatTypePrivate, Username: "alice"},
			want: "alice",
		},
		{
			name: "private chat without username",
			chat: models.Chat{Type: models.ChatTypePrivate, ID: 12345},
			want: "dm-12345",
		},
		{
			name: "group with empty title",
			chat: models.Chat{Type: models.ChatTypeGroup, ID: 67890},
			want: "chat-67890",
		},
		{
			name: "title with leading/trailing spaces and hyphens",
			chat: models.Chat{Type: models.ChatTypeChannel, Title: "  --Hello World--  "},
			want: "hello-world",
		},
		{
			name: "supergroup",
			chat: models.Chat{Type: models.ChatTypeSupergroup, Title: "Dev Chat"},
			want: "dev-chat",
		},
		{
			name: "channel",
			chat: models.Chat{Type: models.ChatTypeChannel, Title: "News"},
			want: "news",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SlugifyChat(tt.chat)
			if got != tt.want {
				t.Errorf("SlugifyChat() = %q, want %q", got, tt.want)
			}
		})
	}
}
