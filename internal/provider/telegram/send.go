package telegram

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/wedow/comms/internal/message"
)

// Send sends a text message to the given chat and returns the resulting message.
func Send(ctx context.Context, api BotAPI, chatID int64, text string) (message.Message, error) {
	resp, err := api.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		return message.Message{}, fmt.Errorf("telegram send: %w", err)
	}
	if resp == nil {
		return message.Message{}, fmt.Errorf("telegram send: nil response")
	}
	return convertMessage(resp), nil
}
