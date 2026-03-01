package telegram

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
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
func Send(ctx context.Context, api BotAPI, chatID int64, text string, replyToID int, parseMode models.ParseMode) (message.Message, error) {
	params := &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: parseMode,
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

// DetectMediaType maps a filename's extension to a Telegram media type.
func DetectMediaType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return "photo"
	case ".gif":
		return "animation"
	case ".mp4", ".mov", ".avi":
		return "video"
	case ".mp3", ".flac", ".wav":
		return "audio"
	case ".ogg":
		return "voice"
	default:
		return "document"
	}
}

// SendMedia sends a media file to the given chat and returns the resulting message.
func SendMedia(ctx context.Context, api BotAPI, chatID int64, file io.Reader, filename string, mediaType string, caption string, replyToID int, parseMode models.ParseMode) (message.Message, error) {
	upload := &models.InputFileUpload{Filename: filename, Data: file}

	var replyParams *models.ReplyParameters
	if replyToID != 0 {
		replyParams = &models.ReplyParameters{MessageID: replyToID}
	}

	var resp *models.Message
	var err error

	switch mediaType {
	case "photo":
		resp, err = api.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:          chatID,
			Photo:           upload,
			Caption:         caption,
			ParseMode:       parseMode,
			ReplyParameters: replyParams,
		})
	case "video":
		resp, err = api.SendVideo(ctx, &bot.SendVideoParams{
			ChatID:          chatID,
			Video:           upload,
			Caption:         caption,
			ParseMode:       parseMode,
			ReplyParameters: replyParams,
		})
	case "audio":
		resp, err = api.SendAudio(ctx, &bot.SendAudioParams{
			ChatID:          chatID,
			Audio:           upload,
			Caption:         caption,
			ParseMode:       parseMode,
			ReplyParameters: replyParams,
		})
	case "voice":
		resp, err = api.SendVoice(ctx, &bot.SendVoiceParams{
			ChatID:          chatID,
			Voice:           upload,
			Caption:         caption,
			ParseMode:       parseMode,
			ReplyParameters: replyParams,
		})
	case "animation":
		resp, err = api.SendAnimation(ctx, &bot.SendAnimationParams{
			ChatID:          chatID,
			Animation:       upload,
			Caption:         caption,
			ParseMode:       parseMode,
			ReplyParameters: replyParams,
		})
	default:
		resp, err = api.SendDocument(ctx, &bot.SendDocumentParams{
			ChatID:          chatID,
			Document:        upload,
			Caption:         caption,
			ParseMode:       parseMode,
			ReplyParameters: replyParams,
		})
	}

	if err != nil {
		return message.Message{}, fmt.Errorf("telegram send media: %w", err)
	}
	if resp == nil {
		return message.Message{}, fmt.Errorf("telegram send media: nil response")
	}
	return convertMessage(resp), nil
}
