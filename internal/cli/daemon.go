package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
				return PrintJSON(cmd.OutOrStdout(), map[string]any{"running": false})
			}

			if daemon.IsRunning(root) {
				return PrintJSON(cmd.OutOrStdout(), map[string]any{"running": true, "pid": pid})
			}

			_ = daemon.RemovePID(root)
			return PrintJSON(cmd.OutOrStdout(), map[string]any{"running": false})
		},
	}
	statusCmd.Flags().String("dir", ".comms", "root directory")

	runCmd := &cobra.Command{
		Use:           "run",
		Short:         "Run the daemon in the foreground",
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
	runCmd.Flags().String("dir", ".comms", "root directory")

	startCmd := &cobra.Command{
		Use:           "start",
		Short:         "Start the daemon via systemd",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			workDir, err := os.Getwd()
			if err != nil {
				return err
			}
			svc := serviceName(name, workDir)
			if err := runSystemctl(cmd.Context(), "--user", "start", svc+".service"); err != nil {
				return err
			}
			return PrintJSON(cmd.OutOrStdout(), map[string]string{"status": "started", "service": svc})
		},
	}
	startCmd.Flags().String("name", "", "service name suffix (default: directory basename)")

	stopCmd := &cobra.Command{
		Use:           "stop",
		Short:         "Stop the daemon via systemd",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			workDir, err := os.Getwd()
			if err != nil {
				return err
			}
			svc := serviceName(name, workDir)
			if err := runSystemctl(cmd.Context(), "--user", "stop", svc+".service"); err != nil {
				return err
			}
			return PrintJSON(cmd.OutOrStdout(), map[string]string{"status": "stopped", "service": svc})
		},
	}
	stopCmd.Flags().String("name", "", "service name suffix (default: directory basename)")

	logsCmd := &cobra.Command{
		Use:           "logs",
		Short:         "Show daemon logs",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			follow, _ := cmd.Flags().GetBool("follow")
			workDir, err := os.Getwd()
			if err != nil {
				return err
			}
			svc := serviceName(name, workDir)
			jArgs := []string{"--user-unit", svc + ".service", "--no-pager"}
			if follow {
				jArgs = append(jArgs, "--follow")
			} else {
				jArgs = append(jArgs, "-n", "100")
			}
			j := exec.CommandContext(cmd.Context(), "journalctl", jArgs...)
			j.Stdout = cmd.OutOrStdout()
			j.Stderr = cmd.ErrOrStderr()
			return j.Run()
		},
	}
	logsCmd.Flags().String("name", "", "service name suffix (default: directory basename)")
	logsCmd.Flags().BoolP("follow", "f", false, "follow log output")

	cmd.AddCommand(statusCmd, runCmd, startCmd, stopCmd, logsCmd, newInstallCmd(), newUninstallCmd())
	return cmd
}

type telegramProvider struct {
	token string
}

func (t telegramProvider) Poll(ctx context.Context, initialOffset int64, handler func(msg message.Message, chatID int64, isEdit bool), reactionHandler func(channel string, msgID int, from string, emoji string, date time.Time)) (int64, error) {
	return telegram.Poll(ctx, t.token, initialOffset, handler, reactionHandler)
}
