package cli

import (
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wedow/comms/internal/store"
)

func newChannelsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channels",
		Short: "List known channels as JSON lines",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")

			root, err := filepath.Abs(dir)
			if err != nil {
				return err
			}

			channels, err := store.ListChannels(root)
			if err != nil {
				return err
			}

			for _, name := range channels {
				provider := name
				if i := strings.Index(name, "-"); i > 0 {
					provider = name[:i]
				}

				out := struct {
					Name     string `json:"name"`
					Provider string `json:"provider"`
					Path     string `json:"path"`
					ChatID   *int64 `json:"chat_id,omitempty"`
				}{Name: name, Provider: provider, Path: filepath.Join(root, name)}

				if id, err := store.ReadChatID(root, name); err == nil {
					out.ChatID = &id
				}

				if err := PrintJSON(cmd.OutOrStdout(), out); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().String("dir", ".comms", "root directory")
	return cmd
}
