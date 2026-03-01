package daemon

import (
	"context"
	"log"

	"github.com/wedow/comms/internal/config"
	"github.com/wedow/comms/internal/message"
	"github.com/wedow/comms/internal/store"
)

// Provider abstracts a message source that delivers messages via a handler callback.
type Provider interface {
	Poll(ctx context.Context, initialOffset int64, handler func(msg message.Message, chatID int64)) (int64, error)
}

// Run is the daemon core loop. It writes a PID file, polls the provider for
// messages, persists them to the store, and cleans up on exit.
func Run(ctx context.Context, cfg config.Config, root string, p Provider) error {
	if err := WritePID(root); err != nil {
		return err
	}
	defer RemovePID(root)

	offset, err := store.ReadOffset(root, "telegram")
	if err != nil {
		return err
	}

	finalOffset, err := p.Poll(ctx, offset, func(msg message.Message, chatID int64) {
		channelDir := msg.Provider + "-" + msg.Channel
		if _, err := store.WriteMessage(root, msg, cfg.General.Format); err != nil {
			log.Printf("failed to write message: %v", err)
		}
		if err := store.WriteChatID(root, channelDir, chatID); err != nil {
			log.Printf("failed to write chat ID: %v", err)
		}
	})
	if err != nil {
		return err
	}

	return store.WriteOffset(root, "telegram", finalOffset)
}
