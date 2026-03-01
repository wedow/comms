package cli

import (
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/wedow/comms/internal/store"
)

func newUnreadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unread",
		Short: "List unread messages and advance cursors",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			channel, _ := cmd.Flags().GetString("channel")

			root, err := filepath.Abs(dir)
			if err != nil {
				return err
			}

			var channels []string
			if channel != "" {
				channels = []string{channel}
			} else {
				channels, err = store.ListChannels(root)
				if err != nil {
					return err
				}
			}

			for _, ch := range channels {
				cursor, err := store.ReadCursor(root, ch)
				if err != nil {
					return err
				}

				paths, err := store.ListMessagesAfter(root, ch, cursor)
				if err != nil {
					return err
				}

				var newest time.Time
				for _, p := range paths {
					msg, err := store.ReadMessage(p)
					if err != nil {
						return err
					}
					if err := PrintJSON(cmd.OutOrStdout(), struct {
						From     string    `json:"from"`
						Provider string    `json:"provider"`
						Channel  string    `json:"channel"`
						Date     time.Time `json:"date"`
						ID       string    `json:"id"`
						Body     string    `json:"body"`
						File     string    `json:"file"`
					}{msg.From, msg.Provider, msg.Channel, msg.Date, msg.ID, msg.Body, p}); err != nil {
						return err
					}
					if msg.Date.After(newest) {
						newest = msg.Date
					}
				}

				if !newest.IsZero() {
					if err := store.WriteCursor(root, ch, newest); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
	cmd.Flags().String("dir", ".comms", "root directory")
	cmd.Flags().String("channel", "", "filter to a single channel")
	return cmd
}
