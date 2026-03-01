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

				if err := PrintJSON(cmd.OutOrStdout(), struct {
					Name     string `json:"name"`
					Provider string `json:"provider"`
					Path     string `json:"path"`
				}{name, provider, filepath.Join(root, name)}); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().String("dir", ".comms", "root directory")
	return cmd
}
