package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/wedow/comms/internal/config"
	"github.com/wedow/comms/internal/message"
	"github.com/wedow/comms/internal/store"
)

func newSendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "send [message...]",
		Short:         "Send a message to a channel",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			channel, _ := cmd.Flags().GetString("channel")
			filePath, _ := cmd.Flags().GetString("file")

			root, err := filepath.Abs(dir)
			if err != nil {
				return err
			}

			cfg, err := config.Load(filepath.Join(root, "config.toml"))
			if err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("load config: %v", err)})
				return err
			}

			var body string
			if len(args) > 0 {
				body = strings.Join(args, " ")
			} else {
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("read stdin: %v", err)})
					return err
				}
				body = strings.TrimSpace(string(data))
			}

			// When no file is provided, stdin is the message body (must be non-empty).
			if filePath == "" && body == "" {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": "empty message body"})
				return fmt.Errorf("empty message body")
			}

			chatID, err := store.ReadChatID(root, channel)
			if err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("resolve channel %q: %v", channel, err)})
				return err
			}

			provider := extractProvider(channel)

			binary, err := resolveProviderBinary(provider)
			if err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("provider binary: %v", err)})
				return err
			}

			providerCfg, err := cfg.ProviderConfig(provider)
			if err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("provider config: %v", err)})
				return err
			}

			// Build args for provider binary
			providerArgs := []string{"send", "--chat-id", strconv.FormatInt(chatID, 10)}

			if format, _ := cmd.Flags().GetString("format"); format != "" {
				providerArgs = append(providerArgs, "--format", format)
			}
			if replyTo, _ := cmd.Flags().GetString("reply-to"); replyTo != "" {
				providerArgs = append(providerArgs, "--reply-to", replyTo)
			}
			if filePath != "" {
				providerArgs = append(providerArgs, "--file", filePath)
			}
			if mediaType, _ := cmd.Flags().GetString("media-type"); mediaType != "" {
				providerArgs = append(providerArgs, "--media-type", mediaType)
			}
			if threadID, _ := cmd.Flags().GetInt("thread"); threadID != 0 {
				providerArgs = append(providerArgs, "--thread", strconv.Itoa(threadID))
			}

			env := []string{"COMMS_PROVIDER_CONFIG=" + string(providerCfg)}

			out, err := delegateWithOutput(binary, providerArgs, env, strings.NewReader(body))
			if err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("provider send: %v", err)})
				return err
			}

			// Parse provider response to get message_id
			var resp struct {
				OK        bool `json:"ok"`
				MessageID int  `json:"message_id"`
			}
			if err := json.Unmarshal(out, &resp); err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("parse provider response: %v", err)})
				return err
			}

			// Construct message for local store
			channelSuffix := channel[len(provider)+1:]
			sent := message.Message{
				Provider: provider,
				Channel:  channelSuffix,
				Date:     time.Now().UTC(),
				ID:       fmt.Sprintf("%s-%d", provider, resp.MessageID),
				Body:     body,
			}

			// Write sent message to local store
			format := cfg.General.Format
			if format == "" {
				format = "markdown"
			}

			// Check for existing unreads before writing
			chanDir := sent.Provider + "-" + sent.Channel
			cursor, _ := store.ReadCursor(root, chanDir)
			unreads, _ := store.ListMessagesAfter(root, chanDir, cursor)

			store.WriteMessage(root, sent, format)

			// If no unreads existed, advance cursor past sent message
			// so the sender's own message doesn't show up in unread
			if len(unreads) == 0 {
				store.WriteCursor(root, chanDir, sent.Date)
			}

			return PrintJSON(cmd.OutOrStdout(), map[string]any{"ok": true, "channel": channel})
		},
	}
	cmd.Flags().String("dir", ".comms", "root directory")
	cmd.Flags().String("channel", "", "channel to send to")
	cmd.Flags().String("reply-to", "", "message ID to reply to")
	cmd.Flags().String("file", "", "path to file to send as media")
	cmd.Flags().String("media-type", "", "media type override (photo, document, audio, video, voice, animation)")
	cmd.Flags().String("format", "", "message format: markdown, html, or plain (default)")
	cmd.Flags().Int("thread", 0, "forum topic thread ID to send to")
	_ = cmd.MarkFlagRequired("channel")
	return cmd
}
