package daemon

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/wedow/comms/internal/config"
	"github.com/wedow/comms/internal/message"
	"github.com/wedow/comms/internal/store"
)

// Provider abstracts a message source that delivers messages via a handler callback.
type Provider interface {
	Poll(ctx context.Context, initialOffset int64, handler func(msg message.Message, chatID int64, isEdit bool), reactionHandler func(channel string, msgID int, from string, emoji string, date time.Time)) (int64, error)
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

	var cb *CallbackRunner
	if cfg.Callback.Command != "" {
		delay, _ := time.ParseDuration(cfg.Callback.Delay)
		cb = NewCallbackRunner(cfg.Callback.Command, delay)
	}

	finalOffset, err := p.Poll(ctx, offset, func(msg message.Message, chatID int64, isEdit bool) {
		channelDir := msg.Provider + "-" + msg.Channel

		if isEdit {
			path, _, err := store.FindMessageByID(root, channelDir, msg.ID, cfg.General.Format)
			if err != nil {
				log.Printf("edit: message not found: %v", err)
				return
			}
			editDate := time.Now().UTC()
			if msg.EditDate != nil {
				editDate = *msg.EditDate
			}
			if err := store.AppendEdit(path, editDate, msg.Body); err != nil {
				log.Printf("edit: append failed: %v", err)
				return
			}
			// Find original message date from filename for cursor reset
			origMsg, err := store.ReadMessage(path)
			if err != nil {
				log.Printf("edit: read message for cursor reset: %v", err)
				return
			}
			if err := store.ResetCursorIfNeeded(root, channelDir, origMsg.Date); err != nil {
				log.Printf("edit: cursor reset failed: %v", err)
			}
			if cb != nil {
				cb.Run(CallbackEnv{
					File:     path,
					Channel:  channelDir,
					Provider: msg.Provider,
					Sender:   msg.From,
				})
			}
			return
		}

		filePath, err := store.WriteMessage(root, msg, cfg.General.Format)
		if err != nil {
			log.Printf("failed to write message: %v", err)
		}
		if err := store.WriteChatID(root, channelDir, chatID); err != nil {
			log.Printf("failed to write chat ID: %v", err)
		}
		if cb != nil {
			cb.Run(CallbackEnv{
				File:     filePath,
				Channel:  channelDir,
				Provider: msg.Provider,
				Sender:   msg.From,
			})
		}
	}, func(channel string, msgID int, from string, emoji string, date time.Time) {
		msgIDStr := fmt.Sprintf("telegram-%d", msgID)
		path, _, err := store.FindMessageByID(root, channel, msgIDStr, cfg.General.Format)
		if err != nil {
			log.Printf("reaction: message not found: %v", err)
			return
		}
		if err := store.AppendReaction(path, date, from, emoji); err != nil {
			log.Printf("reaction: append failed: %v", err)
			return
		}
		origMsg, err := store.ReadMessage(path)
		if err != nil {
			log.Printf("reaction: read message for cursor reset: %v", err)
			return
		}
		if err := store.ResetCursorIfNeeded(root, channel, origMsg.Date); err != nil {
			log.Printf("reaction: cursor reset failed: %v", err)
		}
		if cb != nil {
			cb.Run(CallbackEnv{
				File:     path,
				Channel:  channel,
				Provider: "telegram",
				Sender:   from,
			})
		}
	})
	if err != nil {
		return err
	}

	return store.WriteOffset(root, "telegram", finalOffset)
}
