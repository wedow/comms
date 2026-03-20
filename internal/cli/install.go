package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"github.com/wedow/comms/internal/config"
	"github.com/wedow/comms/internal/store"
)

var platform = runtime.GOOS

// --- systemd (Linux) ---

const unitTemplate = `[Unit]
Description=comms daemon ({{.Name}})
After=network-online.target

[Service]
Type=simple
WorkingDirectory={{.WorkDir}}
Environment=PATH={{.Path}}
ExecStart={{.Binary}} daemon run --dir {{.Dir}}
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

// --- launchd (macOS) ---

const plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Label}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.Binary}}</string>
        <string>daemon</string>
        <string>run</string>
        <string>--dir</string>
        <string>{{.Dir}}</string>
    </array>
    <key>WorkingDirectory</key>
    <string>{{.WorkDir}}</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <dict>
        <key>SuccessfulExit</key>
        <false/>
    </dict>
    <key>StandardOutPath</key>
    <string>{{.LogPath}}</string>
    <key>StandardErrorPath</key>
    <string>{{.LogPath}}</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>{{.Path}}</string>
    </dict>
</dict>
</plist>
`

var plistTmpl = template.Must(template.New("plist").Parse(plistTemplate))

// Swappable for testing.
var (
	runLaunchctl      = defaultRunLaunchctl
	checkLoaded       = defaultCheckLoaded
	getLaunchAgentDir = defaultGetLaunchAgentDir
	getLogDir         = defaultGetLogDir
)

func defaultRunLaunchctl(ctx context.Context, args ...string) error {
	out, err := exec.CommandContext(ctx, "launchctl", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("launchctl %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return nil
}

func defaultCheckLoaded(label string) bool {
	err := exec.Command("launchctl", "list", label).Run()
	return err == nil
}

func defaultGetLaunchAgentDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents"), nil
}

func defaultGetLogDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Logs", "comms"), nil
}

// --- shared helpers ---

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
	return filepath.Join(dir, service+".service"), nil
}

func plistLabel(service string) string {
	return "com.wedow." + service
}

func plistFilePath(service string) (string, error) {
	dir, err := getLaunchAgentDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, plistLabel(service)+".plist"), nil
}

func logFilePath(service string) (string, error) {
	dir, err := getLogDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, service+".log"), nil
}

func writeConfig(path string, cfg config.Config) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

// scaffoldConfig initialises the .comms directory and ensures a config with a
// Telegram token exists.
func scaffoldConfig(cmd *cobra.Command, root string) error {
	cfgPath := filepath.Join(root, "config.toml")

	// Init .comms if needed
	if _, err := os.Stat(root); os.IsNotExist(err) {
		if err := store.InitDir(root); err != nil {
			return err
		}
		if err := writeConfig(cfgPath, config.Default()); err != nil {
			return err
		}
	}

	// Check for token
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	if cfg.Providers == nil || cfg.Providers["telegram"] == nil || cfg.Providers["telegram"]["token"] == "" {
		fmt.Fprint(cmd.ErrOrStderr(), "Telegram bot token: ")
		scanner := bufio.NewScanner(cmd.InOrStdin())
		if !scanner.Scan() {
			return fmt.Errorf("no token provided")
		}
		token := strings.TrimSpace(scanner.Text())
		if token == "" {
			return fmt.Errorf("no token provided")
		}
		if cfg.Providers == nil {
			cfg.Providers = make(map[string]map[string]any)
		}
		if cfg.Providers["telegram"] == nil {
			cfg.Providers["telegram"] = make(map[string]any)
		}
		cfg.Providers["telegram"]["token"] = token
		if err := writeConfig(cfgPath, cfg); err != nil {
			return err
		}
	}

	return nil
}

// --- install ---

func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install a user service for the daemon",
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

			if err := scaffoldConfig(cmd, root); err != nil {
				return err
			}

			binary, err := os.Executable()
			if err != nil {
				return err
			}

			svc := serviceName(name, workDir)
			ctx := cmd.Context()

			switch platform {
			case "darwin":
				return installDarwin(ctx, cmd, svc, binary, workDir, root, startNow)
			default:
				return installLinux(ctx, cmd, svc, binary, workDir, root, startNow)
			}
		},
	}
	cmd.Flags().String("dir", ".comms", "root directory")
	cmd.Flags().String("name", "", "service name suffix (default: directory basename)")
	cmd.Flags().Bool("start", false, "start the service after installing")
	return cmd
}

func installLinux(ctx context.Context, cmd *cobra.Command, svc, binary, workDir, root string, startNow bool) error {
	wasActive := checkActive(svc)

	unitFile, err := unitFilePath(svc)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(unitFile), 0o755); err != nil {
		return err
	}

	f, err := os.Create(unitFile)
	if err != nil {
		return err
	}
	err = unitTmpl.Execute(f, struct{ Name, WorkDir, Binary, Dir, Path string }{svc, workDir, binary, root, os.Getenv("PATH")})
	f.Close()
	if err != nil {
		return err
	}

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
}

func installDarwin(ctx context.Context, cmd *cobra.Command, svc, binary, workDir, root string, startNow bool) error {
	label := plistLabel(svc)
	wasLoaded := checkLoaded(label)

	plistPath, err := plistFilePath(svc)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(plistPath), 0o755); err != nil {
		return err
	}

	logPath, err := logFilePath(svc)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return err
	}

	f, err := os.Create(plistPath)
	if err != nil {
		return err
	}
	err = plistTmpl.Execute(f, struct{ Label, Binary, Dir, WorkDir, LogPath, Path string }{
		label, binary, root, workDir, logPath, os.Getenv("PATH"),
	})
	f.Close()
	if err != nil {
		return err
	}

	if wasLoaded {
		_ = runLaunchctl(ctx, "unload", plistPath)
	}
	if err := runLaunchctl(ctx, "load", plistPath); err != nil {
		return err
	}

	return PrintJSON(cmd.OutOrStdout(), map[string]string{
		"status":  "installed",
		"service": svc,
		"plist":   plistPath,
	})
}

// --- uninstall ---

func newUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the user service for the daemon",
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
				return uninstallDarwin(ctx, cmd, svc)
			default:
				return uninstallLinux(ctx, cmd, svc)
			}
		},
	}
	cmd.Flags().String("name", "", "service name suffix (default: directory basename)")
	return cmd
}

func uninstallLinux(ctx context.Context, cmd *cobra.Command, svc string) error {
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
}

func uninstallDarwin(ctx context.Context, cmd *cobra.Command, svc string) error {
	plistPath, err := plistFilePath(svc)
	if err != nil {
		return err
	}

	_ = runLaunchctl(ctx, "unload", plistPath)

	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return PrintJSON(cmd.OutOrStdout(), map[string]string{
		"status":  "uninstalled",
		"service": svc,
	})
}
