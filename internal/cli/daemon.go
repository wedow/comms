package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wedow/comms/internal/config"
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

			return daemon.Run(cmd.Context(), cfg, root, cfg.ProviderNames())
		},
	}
	runCmd.Flags().String("dir", ".comms", "root directory")

	startCmd := &cobra.Command{
		Use:           "start",
		Short:         "Start the daemon service",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			workDir, err := os.Getwd()
			if err != nil {
				return err
			}
			svc := serviceName(name, workDir)
			ctx := cmd.Context()
			switch platform {
			case "darwin":
				if err := runLaunchctl(ctx, "start", plistLabel(svc)); err != nil {
					return err
				}
			default:
				if err := runSystemctl(ctx, "--user", "start", svc+".service"); err != nil {
					return err
				}
			}
			return PrintJSON(cmd.OutOrStdout(), map[string]string{"status": "started", "service": svc})
		},
	}
	startCmd.Flags().String("name", "", "service name suffix (default: directory basename)")

	stopCmd := &cobra.Command{
		Use:           "stop",
		Short:         "Stop the daemon service",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			workDir, err := os.Getwd()
			if err != nil {
				return err
			}
			svc := serviceName(name, workDir)
			ctx := cmd.Context()
			switch platform {
			case "darwin":
				if err := runLaunchctl(ctx, "stop", plistLabel(svc)); err != nil {
					return err
				}
			default:
				if err := runSystemctl(ctx, "--user", "stop", svc+".service"); err != nil {
					return err
				}
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
			lines, _ := cmd.Flags().GetInt("lines")
			workDir, err := os.Getwd()
			if err != nil {
				return err
			}
			svc := serviceName(name, workDir)
			var c *exec.Cmd
			switch platform {
			case "darwin":
				logPath, err := logFilePath(svc)
				if err != nil {
					return err
				}
				if follow {
					c = exec.CommandContext(cmd.Context(), "tail", "-n", fmt.Sprintf("%d", lines), "-f", logPath)
				} else {
					c = exec.CommandContext(cmd.Context(), "tail", "-n", fmt.Sprintf("%d", lines), logPath)
				}
			default:
				jArgs := []string{"--user-unit", svc + ".service", "--no-pager"}
				if follow {
					jArgs = append(jArgs, "--follow")
				} else {
					jArgs = append(jArgs, "-n", fmt.Sprintf("%d", lines))
				}
				c = exec.CommandContext(cmd.Context(), "journalctl", jArgs...)
			}
			c.Stdout = cmd.OutOrStdout()
			c.Stderr = cmd.ErrOrStderr()
			return c.Run()
		},
	}
	logsCmd.Flags().String("name", "", "service name suffix (default: directory basename)")
	logsCmd.Flags().BoolP("follow", "f", false, "follow log output")
	logsCmd.Flags().IntP("lines", "n", 100, "number of lines to show")

	cmd.AddCommand(statusCmd, runCmd, startCmd, stopCmd, logsCmd, newInstallCmd(), newUninstallCmd())
	return cmd
}
