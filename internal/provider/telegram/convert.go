package telegram

import (
	"fmt"
	"strconv"
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
	} else if m.SenderChat != nil {
		from = m.SenderChat.Username
		if from == "" {
			from = m.SenderChat.Title
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

	if m.MessageThreadID != 0 {
		msg.ThreadID = strconv.Itoa(m.MessageThreadID)
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

	if len(m.Entities) > 0 {
		msg.Entities = convertEntities(m.Entities)
	} else if len(m.CaptionEntities) > 0 {
		msg.Entities = convertEntities(m.CaptionEntities)
	}

	if m.EditDate != 0 {
		t := time.Unix(int64(m.EditDate), 0).UTC()
		msg.EditDate = &t
	}

	if m.ForwardOrigin != nil {
		var fwdFrom string
		var fwdDate int
		switch m.ForwardOrigin.Type {
		case models.MessageOriginTypeUser:
			o := m.ForwardOrigin.MessageOriginUser
			fwdFrom = o.SenderUser.Username
			if fwdFrom == "" {
				fwdFrom = o.SenderUser.FirstName
			}
			fwdDate = o.Date
		case models.MessageOriginTypeHiddenUser:
			o := m.ForwardOrigin.MessageOriginHiddenUser
			fwdFrom = o.SenderUserName
			fwdDate = o.Date
		case models.MessageOriginTypeChat:
			o := m.ForwardOrigin.MessageOriginChat
			fwdFrom = o.SenderChat.Title
			fwdDate = o.Date
		case models.MessageOriginTypeChannel:
			o := m.ForwardOrigin.MessageOriginChannel
			fwdFrom = o.Chat.Title
			fwdDate = o.Date
		}
		msg.ForwardFrom = fwdFrom
		t := time.Unix(int64(fwdDate), 0).UTC()
		msg.ForwardDate = &t
	}

	return msg
}

func convertEntities(ents []models.MessageEntity) []message.Entity {
	out := make([]message.Entity, len(ents))
	for i, e := range ents {
		out[i] = message.Entity{
			Type:   string(e.Type),
			Offset: e.Offset,
			Length: e.Length,
			URL:    e.URL,
		}
	}
	return out
}
