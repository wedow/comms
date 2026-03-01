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

	return message.Message{
		From:     from,
		Provider: "telegram",
		Channel:  SlugifyChat(m.Chat),
		Date:     time.Unix(int64(m.Date), 0).UTC(),
		ID:       fmt.Sprintf("telegram-%d", m.ID),
		Body:     m.Text,
	}
}
