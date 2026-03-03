package cli

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/wedow/comms/internal/store"
)

func newAllowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "allow",
		Short: "Manage allowed chat IDs",
	}
	cmd.AddCommand(newAllowListCmd())
	cmd.AddCommand(newAllowAddCmd())
	cmd.AddCommand(newAllowRemoveCmd())
	cmd.PersistentFlags().String("dir", ".comms", "root directory")
	return cmd
}

func newAllowListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List allowed chat IDs",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			root, err := filepath.Abs(dir)
			if err != nil {
				return err
			}

			ids, err := store.ReadAllowedIDs(root)
			if err != nil {
				return err
			}
			for _, id := range ids {
				if err := PrintJSON(cmd.OutOrStdout(), struct {
					ChatID int64 `json:"chat_id"`
				}{id}); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func newAllowAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "add <chat-id>",
		Short:         "Add a chat ID to the allowlist",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			root, err := filepath.Abs(dir)
			if err != nil {
				return err
			}

			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("invalid chat ID: %v", err)})
				return err
			}

			if err := store.AddAllowedID(root, id); err != nil {
				return err
			}
			return PrintJSON(cmd.OutOrStdout(), map[string]any{"ok": true, "chat_id": id})
		},
	}
}

func newAllowRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "remove <chat-id>",
		Short:         "Remove a chat ID from the allowlist",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			root, err := filepath.Abs(dir)
			if err != nil {
				return err
			}

			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("invalid chat ID: %v", err)})
				return err
			}

			if err := store.RemoveAllowedID(root, id); err != nil {
				return err
			}
			return PrintJSON(cmd.OutOrStdout(), map[string]any{"ok": true, "chat_id": id})
		},
	}
}
