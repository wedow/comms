package daemon

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wedow/comms/internal/config"
	"github.com/wedow/comms/internal/message"
	"github.com/wedow/comms/internal/protocol"
	"github.com/wedow/comms/internal/store"
)

// lookPathFunc is swappable for testing (same pattern as spawnFunc).
var lookPathFunc = exec.LookPath

// Run is the daemon core loop. It spawns provider subprocesses, processes
// events, and persists offsets. Blocks until context cancellation.
func Run(ctx context.Context, cfg config.Config, root string, providers []string) error {
	if err := WritePID(root); err != nil {
		return fmt.Errorf("write PID: %w", err)
	}
	defer RemovePID(root)

	allowed, err := loadAllowed(root)
	if err != nil {
		return fmt.Errorf("read allowed IDs: %w", err)
	}

	var cbDelay time.Duration
	if cfg.Callback.Delay != "" {
		cbDelay, _ = time.ParseDuration(cfg.Callback.Delay)
	}

	managers := make(map[string]*RespawnManager)

	sendTyping := TypingFunc(func(ctx context.Context, provider string, chatID int64) error {
		rm := managers[provider]
		if rm == nil {
			return fmt.Errorf("no manager for provider %s", provider)
		}
		return rm.SendCommand(ctx, protocol.TypingCommand{
			Type:   protocol.TypeTyping,
			ChatID: chatID,
		})
	})

	var cb *CallbackRunner
	if cfg.Callback.Command != "" {
		cb = NewCallbackRunner(ctx, cfg.Callback.Command, cbDelay, sendTyping)
	}

	type taggedEvent struct {
		provider string
		event    any
	}
	merged := make(chan taggedEvent, 16)

	var wg sync.WaitGroup

	for _, prov := range providers {
		binaryPath, err := lookPathFunc("comms-" + prov)
		if err != nil {
			log.Printf("provider %s: binary not found: %v", prov, err)
			continue
		}

		providerCfg, err := cfg.ProviderConfig(prov)
		if err != nil {
			log.Printf("provider %s: config error: %v", prov, err)
			continue
		}

		p := prov
		readOffset := func() int64 {
			offset, _ := store.ReadOffset(root, p)
			return offset
		}

		rm := NewRespawnManager(p, binaryPath, root, providerCfg, readOffset)
		managers[p] = rm

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := rm.Run(ctx); err != nil {
				log.Printf("provider %s: permanent failure: %v", p, err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for evt := range rm.Events() {
				select {
				case merged <- taggedEvent{provider: p, event: evt}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(merged)
	}()

	for te := range merged {
		switch evt := te.event.(type) {
		case protocol.MessageEvent:
			msg := protocolToMessage(te.provider, evt)
			if evt.Type == protocol.TypeEdit {
				handleEditEvent(root, cfg, msg, evt.ChatID, cb)
			} else {
				handleMessageEvent(root, cfg, msg, evt.ChatID, cb)
			}
			if evt.Offset > 0 {
				store.WriteOffset(root, te.provider, evt.Offset)
			}

		case protocol.ReactionEvent:
			channelDir := te.provider + "-" + evt.Channel
			handleReactionEvent(root, cfg, te.provider, channelDir, evt.MessageID, evt.From, evt.Emoji, evt.Date, allowed, cb)
			if evt.Offset > 0 {
				store.WriteOffset(root, te.provider, evt.Offset)
			}

		case protocol.ErrorEvent:
			log.Printf("provider %s error: [%d] %s", te.provider, evt.Code, evt.Message)
		}
	}

	return ctx.Err()
}

// loadAllowed reads allowed IDs from the store and returns them as a map.
func loadAllowed(root string) (map[int64]bool, error) {
	ids, err := store.ReadAllowedIDs(root)
	if err != nil {
		return nil, err
	}
	allowed := make(map[int64]bool, len(ids))
	for _, id := range ids {
		allowed[id] = true
	}
	return allowed, nil
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
