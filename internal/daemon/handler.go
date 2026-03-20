package daemon

import (
	"fmt"
	"log"
	"time"

	"github.com/wedow/comms/internal/config"
	"github.com/wedow/comms/internal/message"
	"github.com/wedow/comms/internal/store"
)

// handleMessageEvent processes a new (non-edit) message: downloads media,
// writes the message file, writes chat ID, and fires the callback.
func handleMessageEvent(root string, cfg config.Config, msg message.Message, chatID int64, cb *CallbackRunner) {
	channelDir := msg.Provider + "-" + msg.Channel

	if msg.DownloadURL != "" {
		if err := downloadMedia(root, channelDir, &msg); err != nil {
			log.Printf("media: download failed: %v", err)
		}
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
			ChatID:   chatID,
		})
	}
}

// handleEditEvent processes an edited message: finds the original, appends the
// edit, resets the cursor, and fires the callback.
func handleEditEvent(root string, cfg config.Config, msg message.Message, chatID int64, cb *CallbackRunner) {
	channelDir := msg.Provider + "-" + msg.Channel

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
			ChatID:   chatID,
		})
	}
}

// handleReactionEvent processes a reaction: reads chat ID, checks allowed list,
// finds the message, appends the reaction, resets the cursor, and fires the callback.
func handleReactionEvent(root string, cfg config.Config, channel string, msgID int, from string, emoji string, date time.Time, allowed map[int64]bool, cb *CallbackRunner) {
	chatID, err := store.ReadChatID(root, channel)
	if err != nil || !allowed[chatID] {
		log.Printf("rejected reaction for chat %q (not in allowed_ids)", channel)
		return
	}

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
		reactionChatID, _ := store.ReadChatID(root, channel)
		cb.Run(CallbackEnv{
			File:     path,
			Channel:  channel,
			Provider: "telegram",
			Sender:   from,
			ChatID:   reactionChatID,
		})
	}
}
