package cli

import (
	"fmt"
	"path/filepath"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/spf13/cobra"
	"github.com/wedow/comms/internal/config"
	"github.com/wedow/comms/providers/telegram"
	"github.com/wedow/comms/internal/store"
)

func newReactCmd(newBot func(string) (telegram.BotAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "react",
		Short:         "Set a reaction on a message",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			channel, _ := cmd.Flags().GetString("channel")
			message, _ := cmd.Flags().GetString("message")
			emoji, _ := cmd.Flags().GetString("emoji")

			root, err := filepath.Abs(dir)
			if err != nil {
				return err
			}

			cfg, err := config.Load(filepath.Join(root, "config.toml"))
			if err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("load config: %v", err)})
				return err
			}

			chatID, err := store.ReadChatID(root, channel)
			if err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("resolve channel %q: %v", channel, err)})
				return err
			}

			msgID, err := telegram.ParseMessageID(message)
			if err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": err.Error()})
				return err
			}

			api, err := newBot(cfg.Telegram.Token)
			if err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("create bot: %v", err)})
				return err
			}

			_, err = api.SetMessageReaction(cmd.Context(), &bot.SetMessageReactionParams{
				ChatID:    chatID,
				MessageID: msgID,
				Reaction: []models.ReactionType{
					{
						Type:              models.ReactionTypeTypeEmoji,
						ReactionTypeEmoji: &models.ReactionTypeEmoji{Type: models.ReactionTypeTypeEmoji, Emoji: emoji},
					},
				},
			})
			if err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("set reaction: %v", err)})
				return err
			}

			return PrintJSON(cmd.OutOrStdout(), map[string]any{"ok": true, "channel": channel})
		},
	}
	cmd.Flags().String("dir", ".comms", "root directory")
	cmd.Flags().String("channel", "", "channel to react in")
	cmd.Flags().String("message", "", "message ID to react to")
	cmd.Flags().String("emoji", "", "emoji reaction")
	_ = cmd.MarkFlagRequired("channel")
	_ = cmd.MarkFlagRequired("message")
	_ = cmd.MarkFlagRequired("emoji")
	return cmd
}
