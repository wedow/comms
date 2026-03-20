package telegram

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/spf13/cobra"
)

// loadProviderConfig reads COMMS_PROVIDER_CONFIG env var and unmarshals it.
func loadProviderConfig() (ProviderConfig, error) {
	raw := os.Getenv("COMMS_PROVIDER_CONFIG")
	if raw == "" {
		return ProviderConfig{}, fmt.Errorf("COMMS_PROVIDER_CONFIG not set")
	}
	var cfg ProviderConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return ProviderConfig{}, fmt.Errorf("parse provider config: %w", err)
	}
	return cfg, nil
}

// parseFormatFlag converts the --format flag value to a Telegram ParseMode.
func parseFormatFlag(cmd *cobra.Command) (models.ParseMode, error) {
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "", "plain":
		return "", nil
	case "markdown":
		return models.ParseModeMarkdown, nil
	case "html":
		return models.ParseModeHTML, nil
	default:
		return "", fmt.Errorf("unsupported format %q (use markdown, html, or plain)", format)
	}
}

// printJSON writes a JSON object followed by a newline to w.
func printJSON(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

// NewSendCmd returns a Cobra command that sends a message via the Telegram API.
func NewSendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "send [message...]",
		Short:         "Send a message to a chat",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadProviderConfig()
			if err != nil {
				_ = printJSON(cmd.ErrOrStderr(), map[string]string{"error": err.Error()})
				return err
			}

			api, err := newBotFunc(cfg.Token)
			if err != nil {
				_ = printJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("create bot: %v", err)})
				return err
			}

			chatID, _ := cmd.Flags().GetInt64("chat-id")

			var body string
			if len(args) > 0 {
				body = strings.Join(args, " ")
			} else {
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					_ = printJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("read stdin: %v", err)})
					return err
				}
				body = strings.TrimSpace(string(data))
			}

			filePath, _ := cmd.Flags().GetString("file")
			if filePath == "" && body == "" {
				err := fmt.Errorf("empty message body")
				_ = printJSON(cmd.ErrOrStderr(), map[string]string{"error": err.Error()})
				return err
			}

			replyToID := 0
			if replyTo, _ := cmd.Flags().GetString("reply-to"); replyTo != "" {
				replyToID, err = ParseMessageID(replyTo)
				if err != nil {
					_ = printJSON(cmd.ErrOrStderr(), map[string]string{"error": err.Error()})
					return err
				}
			}

			threadID, _ := cmd.Flags().GetInt("thread")

			parseMode, err := parseFormatFlag(cmd)
			if err != nil {
				_ = printJSON(cmd.ErrOrStderr(), map[string]string{"error": err.Error()})
				return err
			}

			if filePath != "" {
				f, err := os.Open(filePath)
				if err != nil {
					_ = printJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("open file: %v", err)})
					return err
				}
				defer f.Close()

				mediaType, _ := cmd.Flags().GetString("media-type")
				if mediaType == "" {
					mediaType = DetectMediaType(filepath.Base(filePath))
				}

				sent, err := SendMedia(cmd.Context(), api, chatID, f, filepath.Base(filePath), mediaType, body, replyToID, threadID, parseMode)
				if err != nil {
					_ = printJSON(cmd.ErrOrStderr(), map[string]string{"error": err.Error()})
					return err
				}

				msgID, _ := ParseMessageID(sent.ID)
				return printJSON(cmd.OutOrStdout(), map[string]any{"ok": true, "message_id": msgID})
			}

			sent, err := Send(cmd.Context(), api, chatID, body, replyToID, threadID, parseMode)
			if err != nil {
				_ = printJSON(cmd.ErrOrStderr(), map[string]string{"error": err.Error()})
				return err
			}

			msgID, _ := ParseMessageID(sent.ID)
			return printJSON(cmd.OutOrStdout(), map[string]any{"ok": true, "message_id": msgID})
		},
	}
	cmd.Flags().Int64("chat-id", 0, "chat ID to send to")
	cmd.Flags().String("reply-to", "", "message ID to reply to")
	cmd.Flags().Int("thread", 0, "forum topic thread ID")
	cmd.Flags().String("format", "", "message format: markdown, html, or plain (default)")
	cmd.Flags().String("file", "", "path to file to send as media")
	cmd.Flags().String("media-type", "", "media type override (photo, document, audio, video, voice, animation)")
	_ = cmd.MarkFlagRequired("chat-id")
	return cmd
}

// NewReactCmd returns a Cobra command that sets a reaction on a message.
func NewReactCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "react",
		Short:         "Set a reaction on a message",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadProviderConfig()
			if err != nil {
				_ = printJSON(cmd.ErrOrStderr(), map[string]string{"error": err.Error()})
				return err
			}

			api, err := newBotFunc(cfg.Token)
			if err != nil {
				_ = printJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("create bot: %v", err)})
				return err
			}

			chatID, _ := cmd.Flags().GetInt64("chat-id")
			message, _ := cmd.Flags().GetString("message")
			emoji, _ := cmd.Flags().GetString("emoji")

			msgID, err := ParseMessageID(message)
			if err != nil {
				_ = printJSON(cmd.ErrOrStderr(), map[string]string{"error": err.Error()})
				return err
			}

			_, err = api.SetMessageReaction(cmd.Context(), &bot.SetMessageReactionParams{
				ChatID:    chatID,
				MessageID: msgID,
				Reaction: []models.ReactionType{
					{
						Type:              models.ReactionTypeTypeEmoji,
						ReactionTypeEmoji: &models.ReactionTypeEmoji{Emoji: emoji},
					},
				},
			})
			if err != nil {
				_ = printJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("set reaction: %v", err)})
				return err
			}

			return printJSON(cmd.OutOrStdout(), map[string]any{"ok": true})
		},
	}
	cmd.Flags().Int64("chat-id", 0, "chat ID")
	cmd.Flags().String("message", "", "message ID to react to")
	cmd.Flags().String("emoji", "", "emoji reaction")
	_ = cmd.MarkFlagRequired("chat-id")
	_ = cmd.MarkFlagRequired("message")
	_ = cmd.MarkFlagRequired("emoji")
	return cmd
}
