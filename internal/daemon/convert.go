package daemon

import (
	"fmt"
	"strconv"
	"time"

	"github.com/wedow/comms/internal/message"
	"github.com/wedow/comms/internal/protocol"
)

// protocolToMessage converts a protocol.MessageEvent into a message.Message,
// prefixing ID fields with the provider name.
func protocolToMessage(provider string, evt protocol.MessageEvent) message.Message {
	msg := message.Message{
		Provider:     provider,
		ID:           fmt.Sprintf("%s-%d", provider, evt.ID),
		Channel:      evt.Channel,
		From:         evt.From,
		Date:         evt.Date,
		Body:         evt.Body,
		ReplyToBody:  evt.ReplyToBody,
		Quote:        evt.Quote,
		MediaType:    evt.MediaType,
		MediaFileID:  evt.MediaFileID,
		DownloadURL:  evt.DownloadURL,
		MediaExt:     evt.MediaExt,
		Caption:      evt.Caption,
		ForwardFrom:  evt.ForwardFrom,
		ForwardDate:  evt.ForwardDate,
		EditDate:     evt.EditDate,
		MediaGroupID: evt.MediaGroupID,
	}

	if evt.ReplyTo != 0 {
		msg.ReplyTo = fmt.Sprintf("%s-%d", provider, evt.ReplyTo)
	}
	if evt.ThreadID != 0 {
		msg.ThreadID = strconv.Itoa(evt.ThreadID)
	}

	if len(evt.Entities) > 0 {
		msg.Entities = make([]message.Entity, len(evt.Entities))
		for i, e := range evt.Entities {
			msg.Entities[i] = message.Entity{
				Type:   e.Type,
				Offset: e.Offset,
				Length: e.Length,
				URL:    e.URL,
			}
		}
	}

	return msg
}

// reactionInfo holds the fields needed to process a reaction event.
type reactionInfo struct {
	Channel string
	MsgID   string
	From    string
	Emoji   string
	Date    time.Time
}

// protocolToReaction converts a protocol.ReactionEvent into a reactionInfo,
// prefixing the message ID with the provider name.
func protocolToReaction(provider string, evt protocol.ReactionEvent) reactionInfo {
	return reactionInfo{
		Channel: evt.Channel,
		MsgID:   fmt.Sprintf("%s-%d", provider, evt.MessageID),
		From:    evt.From,
		Emoji:   evt.Emoji,
		Date:    evt.Date,
	}
}
