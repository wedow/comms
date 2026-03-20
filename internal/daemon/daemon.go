package daemon

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/wedow/comms/internal/config"
	"github.com/wedow/comms/internal/message"
	"github.com/wedow/comms/internal/store"
)

// Run is the daemon core loop. It writes a PID file and waits for cancellation.
// Provider-specific polling will be handled by plugin subprocesses in a follow-up.
func Run(ctx context.Context, cfg config.Config, root string, providers []string) error {
	if err := WritePID(root); err != nil {
		return fmt.Errorf("write PID: %w", err)
	}
	defer RemovePID(root)

	<-ctx.Done()
	return ctx.Err()
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
