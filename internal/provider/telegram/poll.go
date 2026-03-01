package telegram

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/wedow/comms/internal/message"
)

func Poll(ctx context.Context, token string, initialOffset int64, handler func(msg message.Message, chatID int64)) (int64, error) {
	lastOffset := initialOffset

	b, err := bot.New(token,
		bot.WithSkipGetMe(),
		bot.WithInitialOffset(initialOffset),
		bot.WithDefaultHandler(func(_ context.Context, _ *bot.Bot, update *models.Update) {
			if update.Message == nil {
				return
			}
			msg := convertMessage(update.Message)
			lastOffset = update.ID + 1
			handler(msg, update.Message.Chat.ID)
		}),
	)
	if err != nil {
		return lastOffset, err
	}

	b.Start(ctx)
	return lastOffset, nil
}
