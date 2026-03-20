package cli

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/wedow/comms/internal/config"
	"github.com/wedow/comms/internal/store"
)

func newReactCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "react",
		Short:         "Set a reaction on a message",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			channel, _ := cmd.Flags().GetString("channel")
			msg, _ := cmd.Flags().GetString("message")
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

			providerArgs := []string{
				"react",
				"--chat-id", strconv.FormatInt(chatID, 10),
				"--message", msg,
				"--emoji", emoji,
			}

			env := []string{"COMMS_PROVIDER_CONFIG=" + string(providerCfg)}

			if _, err := delegateWithOutput(binary, providerArgs, env, nil); err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("provider react: %v", err)})
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
