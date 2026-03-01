package telegram

import (
	"fmt"
	"time"

	"github.com/go-telegram/bot/models"
	"github.com/wedow/comms/internal/message"
)

func convertMessage(m *models.Message) message.Message {
	from := "unknown"
	if m.From != nil {
		from = m.From.Username
		if from == "" {
			from = m.From.FirstName
		}
	}

	msg := message.Message{
		From:     from,
		Provider: "telegram",
		Channel:  SlugifyChat(m.Chat),
		Date:     time.Unix(int64(m.Date), 0).UTC(),
		ID:       fmt.Sprintf("telegram-%d", m.ID),
		Body:     m.Text,
	}

	if m.ReplyToMessage != nil {
		msg.ReplyTo = fmt.Sprintf("telegram-%d", m.ReplyToMessage.ID)
		msg.ReplyToBody = m.ReplyToMessage.Text
	} else if m.ExternalReply != nil && m.ExternalReply.MessageID != 0 {
		msg.ReplyTo = fmt.Sprintf("telegram-%d", m.ExternalReply.MessageID)
	}

	if m.Quote != nil {
		msg.Quote = m.Quote.Text
	}

	// Media metadata
	switch {
	case len(m.Photo) > 0:
		msg.MediaType = "photo"
		msg.MediaFileID = m.Photo[len(m.Photo)-1].FileID
	case m.Video != nil:
		msg.MediaType = "video"
		msg.MediaFileID = m.Video.FileID
	case m.Audio != nil:
		msg.MediaType = "audio"
		msg.MediaFileID = m.Audio.FileID
	case m.Document != nil:
		msg.MediaType = "document"
		msg.MediaFileID = m.Document.FileID
	case m.Voice != nil:
		msg.MediaType = "voice"
		msg.MediaFileID = m.Voice.FileID
	case m.Animation != nil:
		msg.MediaType = "animation"
		msg.MediaFileID = m.Animation.FileID
	case m.Sticker != nil:
		msg.MediaType = "sticker"
		msg.MediaFileID = m.Sticker.FileID
	case m.VideoNote != nil:
		msg.MediaType = "video_note"
		msg.MediaFileID = m.VideoNote.FileID
	}

	if m.Caption != "" {
		msg.Caption = m.Caption
	}
	if m.MediaGroupID != "" {
		msg.MediaGroupID = m.MediaGroupID
	}

	return msg
}
