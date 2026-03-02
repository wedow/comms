package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wedow/comms/internal/store"
)

func newAckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "ack <message-id>",
		Short:         "Advance the read cursor to a message",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			channel, _ := cmd.Flags().GetString("channel")
			msgID := args[0]

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
				_, msg, err := store.FindMessageByID(root, ch, msgID, "")
				if err != nil {
					continue // not in this channel
				}
				if err := store.WriteCursor(root, ch, msg.Date); err != nil {
					return err
				}
				return PrintJSON(cmd.OutOrStdout(), map[string]string{
					"status":  "acked",
					"channel": ch,
					"id":      msgID,
				})
			}

			_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("message %s not found", msgID)})
			return fmt.Errorf("message %s not found", msgID)
		},
	}
	cmd.Flags().String("dir", ".comms", "root directory")
	cmd.Flags().String("channel", "", "channel to search (default: all)")
	return cmd
}
