package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"github.com/wedow/comms/internal/config"
	"github.com/wedow/comms/internal/store"
)

const unitTemplate = `[Unit]
Description=comms daemon ({{.Name}})
After=network-online.target

[Service]
Type=simple
WorkingDirectory={{.WorkDir}}
Environment=PATH={{.Path}}
ExecStart={{.Binary}} daemon run --dir .comms
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
`

var unitTmpl = template.Must(template.New("unit").Parse(unitTemplate))

// Swappable for testing.
var (
	runSystemctl = defaultRunSystemctl
	checkActive  = defaultCheckActive
	getUnitDir   = defaultGetUnitDir
)

func defaultRunSystemctl(ctx context.Context, args ...string) error {
	out, err := exec.CommandContext(ctx, "systemctl", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return nil
}

func defaultCheckActive(svc string) bool {
	out, err := exec.Command("systemctl", "--user", "is-active", svc+".service").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "active"
}

func defaultGetUnitDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "systemd", "user"), nil
}

func serviceName(name, workDir string) string {
	if name != "" {
		return "comms-" + name
	}
	return "comms-" + filepath.Base(workDir)
}

func unitFilePath(service string) (string, error) {
	dir, err := getUnitDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, service+".service"), nil
}

func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install a systemd user service for the daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			name, _ := cmd.Flags().GetString("name")
			startNow, _ := cmd.Flags().GetBool("start")

			workDir, err := os.Getwd()
			if err != nil {
				return err
			}

			root, err := filepath.Abs(dir)
			if err != nil {
				return err
			}

			// Init .comms if needed
			if _, err := os.Stat(root); os.IsNotExist(err) {
				if err := store.InitDir(root); err != nil {
					return err
				}
				cfgPath := filepath.Join(root, "config.toml")
				f, err := os.Create(cfgPath)
				if err != nil {
					return err
				}
				defer f.Close()
				if err := toml.NewEncoder(f).Encode(config.Default()); err != nil {
					return err
				}
			}

			// Check for token
			cfg, err := config.Load(filepath.Join(root, "config.toml"))
			if err != nil {
				return err
			}

			if cfg.Telegram.Token == "" {
				fmt.Fprint(cmd.ErrOrStderr(), "Telegram bot token: ")
				scanner := bufio.NewScanner(cmd.InOrStdin())
				if !scanner.Scan() {
					return fmt.Errorf("no token provided")
				}
				token := strings.TrimSpace(scanner.Text())
				if token == "" {
					return fmt.Errorf("no token provided")
				}
				cfg.Telegram.Token = token
				cfgPath := filepath.Join(root, "config.toml")
				f, err := os.Create(cfgPath)
				if err != nil {
					return err
				}
				defer f.Close()
				if err := toml.NewEncoder(f).Encode(cfg); err != nil {
					return err
				}
			}

			binary, err := os.Executable()
			if err != nil {
				return err
			}

			svc := serviceName(name, workDir)
			wasActive := checkActive(svc)

			unitFile, err := unitFilePath(svc)
			if err != nil {
				return err
			}

			f, err := os.Create(unitFile)
			if err != nil {
				return err
			}
			err = unitTmpl.Execute(f, struct{ Name, WorkDir, Binary, Path string }{svc, workDir, binary, os.Getenv("PATH")})
			f.Close()
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			if err := runSystemctl(ctx, "--user", "daemon-reload"); err != nil {
				return err
			}
			if err := runSystemctl(ctx, "--user", "enable", svc+".service"); err != nil {
				return err
			}

			if wasActive {
				if err := runSystemctl(ctx, "--user", "restart", svc+".service"); err != nil {
					return err
				}
			} else if startNow {
				if err := runSystemctl(ctx, "--user", "start", svc+".service"); err != nil {
					return err
				}
			}

			return PrintJSON(cmd.OutOrStdout(), map[string]string{
				"status":  "installed",
				"service": svc,
				"unit":    unitFile,
			})
		},
	}
	cmd.Flags().String("dir", ".comms", "root directory")
	cmd.Flags().String("name", "", "service name suffix (default: directory basename)")
	cmd.Flags().Bool("start", false, "start the service after installing")
	return cmd
}

func newUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the systemd user service for the daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")

			workDir, err := os.Getwd()
			if err != nil {
				return err
			}

			svc := serviceName(name, workDir)
			ctx := cmd.Context()

			// stop and disable (ignore errors — may not be active/enabled)
			_ = runSystemctl(ctx, "--user", "stop", svc+".service")
			_ = runSystemctl(ctx, "--user", "disable", svc+".service")

			unitFile, err := unitFilePath(svc)
			if err != nil {
				return err
			}
			if err := os.Remove(unitFile); err != nil && !os.IsNotExist(err) {
				return err
			}

			if err := runSystemctl(ctx, "--user", "daemon-reload"); err != nil {
				return err
			}

			return PrintJSON(cmd.OutOrStdout(), map[string]string{
				"status":  "uninstalled",
				"service": svc,
			})
		},
	}
	cmd.Flags().String("name", "", "service name suffix (default: directory basename)")
	return cmd
}
