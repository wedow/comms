package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/wedow/comms/internal/message"
)

// ParseMessageID strips the "telegram-" prefix and returns the numeric message ID.
func ParseMessageID(id string) (int, error) {
	rest, ok := strings.CutPrefix(id, "telegram-")
	if !ok || rest == "" {
		return 0, fmt.Errorf("invalid telegram message ID: %q", id)
	}
	n, err := strconv.Atoi(rest)
	if err != nil {
		return 0, fmt.Errorf("invalid telegram message ID: %q", id)
	}
	return n, nil
}

// Send sends a text message to the given chat and returns the resulting message.
// If replyToID is non-zero, the message is sent as a reply to that message.
func Send(ctx context.Context, api BotAPI, chatID int64, text string, replyToID int) (message.Message, error) {
	params := &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	}
	if replyToID != 0 {
		params.ReplyParameters = &models.ReplyParameters{MessageID: replyToID}
	}
	resp, err := api.SendMessage(ctx, params)
	if err != nil {
		return message.Message{}, fmt.Errorf("telegram send: %w", err)
	}
	if resp == nil {
		return message.Message{}, fmt.Errorf("telegram send: nil response")
	}
	return convertMessage(resp), nil
}
