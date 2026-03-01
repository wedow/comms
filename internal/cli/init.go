package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"github.com/wedow/comms/internal/config"
	"github.com/wedow/comms/internal/embeddocs"
	"github.com/wedow/comms/internal/store"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold the .comms/ directory with default config and docs",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")

			root, err := filepath.Abs(dir)
			if err != nil {
				return err
			}

			if err := store.InitDir(root); err != nil {
				return err
			}

			// Write config.toml only if it doesn't exist
			cfgPath := filepath.Join(root, "config.toml")
			if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
				f, err := os.Create(cfgPath)
				if err != nil {
					return err
				}
				defer f.Close()
				if err := toml.NewEncoder(f).Encode(config.Default()); err != nil {
					return err
				}
			}

			// Always overwrite docs
			docPath := filepath.Join(root, "docs", "telegram-setup.md")
			if err := os.WriteFile(docPath, embeddocs.TelegramSetupDoc, 0o644); err != nil {
				return err
			}

			out, _ := json.Marshal(map[string]string{
				"status": "initialized",
				"path":   root,
			})
			fmt.Fprintln(cmd.OutOrStdout(), string(out))
			return nil
		},
	}
	cmd.Flags().String("dir", ".comms", "root directory to initialize")
	return cmd
}
