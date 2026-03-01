package telegram

import "github.com/go-telegram/bot"

// NewBot creates a BotAPI from a Telegram bot token.
func NewBot(token string) (BotAPI, error) {
	b, err := bot.New(token, bot.WithSkipGetMe())
	if err != nil {
		return nil, err
	}
	return b, nil
}
