package telegram

import (
	"context"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/wedow/comms/internal/message"
)

func Poll(ctx context.Context, token string, initialOffset int64, handler func(msg message.Message, chatID int64, isEdit bool), reactionHandler func(channel string, msgID int, from string, emoji string, date time.Time)) (int64, error) {
	lastOffset := initialOffset

	b, err := bot.New(token,
		bot.WithSkipGetMe(),
		bot.WithInitialOffset(initialOffset),
		bot.WithAllowedUpdates(bot.AllowedUpdates{"message", "edited_message", "channel_post", "message_reaction"}),
		bot.WithDefaultHandler(func(_ context.Context, _ *bot.Bot, update *models.Update) {
			if update.MessageReaction != nil {
				r := update.MessageReaction
				lastOffset = update.ID + 1
				from := "unknown"
				if r.User != nil {
					from = r.User.Username
					if from == "" {
						from = r.User.FirstName
					}
				}
				channel := "telegram-" + SlugifyChat(r.Chat)
				date := time.Unix(int64(r.Date), 0).UTC()
				for _, rt := range r.NewReaction {
					if rt.Type == models.ReactionTypeTypeEmoji && rt.ReactionTypeEmoji != nil {
						reactionHandler(channel, r.MessageID, from, rt.ReactionTypeEmoji.Emoji, date)
					}
				}
				return
			}
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
