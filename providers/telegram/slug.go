package telegram

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-telegram/bot/models"
)

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

func SlugifyChat(chat models.Chat) string {
	if chat.Type == models.ChatTypePrivate {
		if chat.Username != "" {
			return chat.Username
		}
		return fmt.Sprintf("dm-%d", chat.ID)
	}

	title := strings.TrimSpace(chat.Title)
	if title == "" {
		return fmt.Sprintf("chat-%d", chat.ID)
	}

	slug := strings.ToLower(title)
	slug = nonAlphaNum.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}
