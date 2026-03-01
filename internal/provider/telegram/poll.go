package telegram

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/wedow/comms/internal/message"
)

func Poll(ctx context.Context, token string, initialOffset int64, handler func(msg message.Message, chatID int64, isEdit bool)) (int64, error) {
	lastOffset := initialOffset

	b, err := bot.New(token,
		bot.WithSkipGetMe(),
		bot.WithInitialOffset(initialOffset),
		bot.WithDefaultHandler(func(_ context.Context, _ *bot.Bot, update *models.Update) {
			if update.EditedMessage != nil {
				msg := convertMessage(update.EditedMessage)
				lastOffset = update.ID + 1
				handler(msg, update.EditedMessage.Chat.ID, true)
				return
			}
			if update.ChannelPost != nil {
				msg := convertMessage(update.ChannelPost)
				lastOffset = update.ID + 1
				handler(msg, update.ChannelPost.Chat.ID, false)
				return
			}
			if update.Message == nil {
				return
			}
			msg := convertMessage(update.Message)
			lastOffset = update.ID + 1
			handler(msg, update.Message.Chat.ID, false)
		}),
	)
	if err != nil {
		return lastOffset, err
	}

	b.Start(ctx)
	return lastOffset, nil
}
