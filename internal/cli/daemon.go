package cli

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wedow/comms/internal/daemon"
)

func newDaemonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Manage the comms daemon",
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Check daemon status",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			root, err := filepath.Abs(dir)
			if err != nil {
				return err
			}

			pid, pidErr := daemon.ReadPID(root)
			if pidErr != nil {
				// No PID file
				return PrintJSON(cmd.OutOrStdout(), map[string]any{"running": false})
			}

			if daemon.IsRunning(root) {
				return PrintJSON(cmd.OutOrStdout(), map[string]any{"running": true, "pid": pid})
			}

			// Stale PID file: clean up
			_ = daemon.RemovePID(root)
			return PrintJSON(cmd.OutOrStdout(), map[string]any{"running": false})
		},
	}
	statusCmd.Flags().String("dir", ".comms", "root directory")

	cmd.AddCommand(statusCmd)
	return cmd
}
