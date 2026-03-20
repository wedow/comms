package daemon

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
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

	allowedIDs, err := store.ReadAllowedIDs(root)
	if err != nil {
		return err
	}
	allowed := make(map[int64]bool, len(allowedIDs))
	for _, id := range allowedIDs {
		allowed[id] = true
	}
	if len(allowed) == 0 {
		log.Printf("warning: no allowed_ids configured, rejecting all messages")
	}

	var typing TypingIndicator
	if ti, ok := p.(TypingIndicator); ok {
		typing = ti
	}

	var cb *CallbackRunner
	if cfg.Callback.Command != "" {
		delay, _ := time.ParseDuration(cfg.Callback.Delay)
		cb = NewCallbackRunner(ctx, cfg.Callback.Command, delay, typing)
	}

	finalOffset, err := p.Poll(ctx, offset, func(msg message.Message, chatID int64, isEdit bool) {
		if !allowed[chatID] {
			log.Printf("rejected message from chat %d (not in allowed_ids)", chatID)
			return
		}
		if isEdit {
			handleEditEvent(root, cfg, msg, chatID, cb)
		} else {
			handleMessageEvent(root, cfg, msg, chatID, cb)
		}
	}, func(channel string, msgID int, from string, emoji string, date time.Time) {
		handleReactionEvent(root, cfg, channel, msgID, from, emoji, date, allowed, cb)
	})
	if err != nil {
		return err
	}

	return store.WriteOffset(root, "telegram", finalOffset)
}

// downloadMedia fetches media from msg.DownloadURL, saves it via store.WriteMedia,
// and sets msg.MediaURL to the relative path within the channel directory.
func downloadMedia(root, channelDir string, msg *message.Message) error {
	resp, err := http.Get(msg.DownloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: status %d", msg.DownloadURL, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	timestamp := strings.ReplaceAll(msg.Date.Format(time.RFC3339Nano), ":", "-")
	chanPath := filepath.Join(root, channelDir)
	if msg.ThreadID != "" {
		chanPath = filepath.Join(chanPath, "topic-"+msg.ThreadID)
	}
	path, err := store.WriteMedia(chanPath, timestamp, 1, msg.MediaExt, data)
	if err != nil {
		return err
	}

	// MediaURL is the relative path within the channel directory
	rel, err := filepath.Rel(chanPath, path)
	if err != nil {
		return err
	}
	msg.MediaURL = rel
	return nil
}
