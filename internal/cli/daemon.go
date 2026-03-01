package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/wedow/comms/internal/config"
	"github.com/wedow/comms/internal/daemon"
	"github.com/wedow/comms/internal/message"
	"github.com/wedow/comms/internal/provider/telegram"
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

	startCmd := &cobra.Command{
		Use:           "start",
		Short:         "Start the comms daemon",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			root, err := filepath.Abs(dir)
			if err != nil {
				return err
			}

			cfg, err := config.Load(filepath.Join(root, "config.toml"))
			if err != nil {
				return err
			}

			if daemon.IsRunning(root) {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": "daemon already running"})
				return fmt.Errorf("daemon already running")
			}

			return daemon.Run(cmd.Context(), cfg, root, telegramProvider{token: cfg.Telegram.Token})
		},
	}
	startCmd.Flags().String("dir", ".comms", "root directory")

	stopCmd := &cobra.Command{
		Use:           "stop",
		Short:         "Stop the comms daemon",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			root, err := filepath.Abs(dir)
			if err != nil {
				return err
			}

			pid, pidErr := daemon.ReadPID(root)
			if pidErr != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": "daemon not running"})
				return fmt.Errorf("daemon not running")
			}

			if !daemon.IsRunning(root) {
				_ = daemon.RemovePID(root)
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": "daemon not running (stale pid)"})
				return fmt.Errorf("daemon not running (stale pid)")
			}

			if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
				_ = PrintJSON(cmd.ErrOrStderr(), map[string]string{"error": fmt.Sprintf("failed to stop daemon: %v", err)})
				return fmt.Errorf("failed to stop daemon: %w", err)
			}

			// Wait briefly for process to exit
			for i := 0; i < 10; i++ {
				time.Sleep(100 * time.Millisecond)
				if !daemon.IsRunning(root) {
					_ = daemon.RemovePID(root)
					return PrintJSON(cmd.OutOrStdout(), map[string]string{"status": "stopped"})
				}
			}

			return PrintJSON(cmd.OutOrStdout(), map[string]string{"status": "stopping", "pid": fmt.Sprintf("%d", pid)})
		},
	}
	stopCmd.Flags().String("dir", ".comms", "root directory")

	cmd.AddCommand(statusCmd, startCmd, stopCmd)
	return cmd
}

type telegramProvider struct {
	token string
}

func (t telegramProvider) Poll(ctx context.Context, initialOffset int64, handler func(msg message.Message, chatID int64, isEdit bool)) (int64, error) {
	return telegram.Poll(ctx, t.token, initialOffset, handler)
}
