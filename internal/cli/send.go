package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wedow/comms/internal/config"
	"github.com/wedow/comms/internal/provider/telegram"
	"github.com/wedow/comms/internal/store"
)

func newSendCmd(newBot func(string) (telegram.BotAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "send",
		Short:         "Send a message to a channel",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			channel, _ := cmd.Flags().GetString("channel")

			root, err := filepath.Abs(dir)
			if err != nil {
				return err
			}

			cfg, err := config.Load(filepath.Join(root, "config.toml"))
			if err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("load config: %v", err)})
				return err
			}

			data, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("read stdin: %v", err)})
				return err
			}
			body := strings.TrimSpace(string(data))
			if body == "" {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": "empty message body"})
				return fmt.Errorf("empty message body")
			}

			chatID, err := store.ReadChatID(root, channel)
			if err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("resolve channel %q: %v", channel, err)})
				return err
			}

			api, err := newBot(cfg.Telegram.Token)
			if err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("create bot: %v", err)})
				return err
			}

			replyToID := 0
			if replyTo, _ := cmd.Flags().GetString("reply-to"); replyTo != "" {
				replyToID, err = telegram.ParseMessageID(replyTo)
				if err != nil {
					_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": err.Error()})
					return err
				}
			}

			if _, err := telegram.Send(cmd.Context(), api, chatID, body, replyToID); err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": err.Error()})
				return err
			}

			return PrintJSON(cmd.OutOrStdout(), map[string]any{"ok": true, "channel": channel})
		},
	}
	cmd.Flags().String("dir", ".comms", "root directory")
	cmd.Flags().String("channel", "", "channel to send to")
	cmd.Flags().String("reply-to", "", "message ID to reply to")
	_ = cmd.MarkFlagRequired("channel")
	return cmd
}
