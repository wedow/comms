package telegram

import (
	"testing"

	"github.com/go-telegram/bot"
)

func TestBotSatisfiesBotAPI(t *testing.T) {
	// Compile-time check that *bot.Bot satisfies BotAPI
	var _ BotAPI = (*bot.Bot)(nil)
}
