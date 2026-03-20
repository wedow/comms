// Package telegram wraps the go-telegram/bot SDK for comms.
package telegram

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// BotAPI is the subset of *bot.Bot methods this package uses.
type BotAPI interface {
	SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error)
	SendPhoto(ctx context.Context, params *bot.SendPhotoParams) (*models.Message, error)
	SendDocument(ctx context.Context, params *bot.SendDocumentParams) (*models.Message, error)
	SendAudio(ctx context.Context, params *bot.SendAudioParams) (*models.Message, error)
	SendVideo(ctx context.Context, params *bot.SendVideoParams) (*models.Message, error)
	SendVoice(ctx context.Context, params *bot.SendVoiceParams) (*models.Message, error)
	SendAnimation(ctx context.Context, params *bot.SendAnimationParams) (*models.Message, error)
	SetMessageReaction(ctx context.Context, params *bot.SetMessageReactionParams) (bool, error)
	GetFile(ctx context.Context, params *bot.GetFileParams) (*models.File, error)
	FileDownloadLink(f *models.File) string
}
