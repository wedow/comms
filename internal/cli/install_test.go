package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/wedow/comms/internal/config"
)

// stubSystemd replaces systemd hooks with test stubs.
// Returns a slice that accumulates systemctl args from each call,
// and a cleanup function to restore originals.
func stubSystemd(t *testing.T, unitDir string, active bool) *[][]string {
	t.Helper()
	origRun, origCheck, origDir, origPlatform := runSystemctl, checkActive, getUnitDir, platform
	t.Cleanup(func() {
		runSystemctl, checkActive, getUnitDir, platform = origRun, origCheck, origDir, origPlatform
	})

	platform = "linux"
	var calls [][]string
	runSystemctl = func(_ context.Context, args ...string) error {
		calls = append(calls, args)
		return nil
	}
	checkActive = func(svc string) bool { return active }
	getUnitDir = func() (string, error) { return unitDir, nil }
	return &calls
}

// stubLaunchd replaces launchd hooks with test stubs.
// Returns a slice that accumulates launchctl args from each call.
func stubLaunchd(t *testing.T, agentDir string, loaded bool) *[][]string {
	t.Helper()
	origRun, origCheck, origDir, origLogDir, origPlatform := runLaunchctl, checkLoaded, getLaunchAgentDir, getLogDir, platform
	t.Cleanup(func() {
		runLaunchctl, checkLoaded, getLaunchAgentDir, getLogDir, platform = origRun, origCheck, origDir, origLogDir, origPlatform
	})

	logDir := filepath.Join(agentDir, "..", "Logs", "comms")

	platform = "darwin"
	var calls [][]string
	runLaunchctl = func(_ context.Context, args ...string) error {
		calls = append(calls, args)
		return nil
	}
	checkLoaded = func(label string) bool { return loaded }
	getLaunchAgentDir = func() (string, error) { return agentDir, nil }
	getLogDir = func() (string, error) { return logDir, nil }
	return &calls
}

func TestServiceName(t *testing.T) {
	tests := []struct {
		name    string
		workDir string
		want    string
	}{
		{"", "/home/user/p/stuart", "comms-stuart"},
		{"", "/home/user/myproject", "comms-myproject"},
		{"custom", "/home/user/p/stuart", "comms-custom"},
		{"foo", "/anything", "comms-foo"},
	}
	for _, tt := range tests {
		got := serviceName(tt.name, tt.workDir)
		if got != tt.want {
			t.Errorf("serviceName(%q, %q) = %q, want %q", tt.name, tt.workDir, got, tt.want)
		}
	}
}

func TestUnitTemplateRendering(t *testing.T) {
	var buf bytes.Buffer
	err := unitTmpl.Execute(&buf, struct{ Name, WorkDir, Binary, Dir, Path string }{
		"comms-stuart", "/home/user/p/stuart", "/usr/local/bin/comms", "/home/user/p/stuart/.comms", "/usr/local/bin:/usr/bin",
	})
	if err != nil {
		t.Fatalf("template execute: %v", err)
	}

	got := buf.String()
	expects := []string{
		"Description=comms daemon (comms-stuart)",
		"WorkingDirectory=/home/user/p/stuart",
		"ExecStart=/usr/local/bin/comms daemon run --dir /home/user/p/stuart/.comms",
		"Restart=on-failure",
		"WantedBy=default.target",
	}
	for _, want := range expects {
		if !strings.Contains(got, want) {
			t.Errorf("unit file missing %q\ngot:\n%s", want, got)
		}
	}
}

func TestPlistTemplateRendering(t *testing.T) {
	var buf bytes.Buffer
	err := plistTmpl.Execute(&buf, struct{ Label, Binary, Dir, WorkDir, LogPath, Path string }{
		"com.wedow.comms-test", "/usr/local/bin/comms", "/home/user/.comms",
		"/home/user/p/stuart", "/home/user/Library/Logs/comms/comms-test.log",
		"/usr/local/bin:/usr/bin",
	})
	if err != nil {
		t.Fatalf("template execute: %v", err)
	}

	got := buf.String()
	expects := []string{
		"<string>com.wedow.comms-test</string>",
		"<string>/usr/local/bin/comms</string>",
		"<string>daemon</string>",
		"<string>run</string>",
		"<string>--dir</string>",
		"<string>/home/user/.comms</string>",
		"<string>/home/user/p/stuart</string>",
		"<key>RunAtLoad</key>",
		"<key>SuccessfulExit</key>",
		"<string>/home/user/Library/Logs/comms/comms-test.log</string>",
	}
	for _, want := range expects {
		if !strings.Contains(got, want) {
			t.Errorf("plist missing %q\ngot:\n%s", want, got)
		}
	}
}

func TestPlistLabel(t *testing.T) {
	if got := plistLabel("comms-test"); got != "com.wedow.comms-test" {
		t.Errorf("plistLabel = %q, want com.wedow.comms-test", got)
	}
}

func TestInstallScaffoldsDir(t *testing.T) {
	tmpDir := t.TempDir()
	root := filepath.Join(tmpDir, ".comms")
	unitDir := filepath.Join(tmpDir, "units")
	os.MkdirAll(unitDir, 0o755)

	calls := stubSystemd(t, unitDir, false)

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetIn(strings.NewReader("test-token\n"))
	cmd.SetArgs([]string{"daemon", "install", "--dir", root, "--name", "test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install: %v", err)
	}

	// .comms dir created
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Fatal(".comms dir not created")
	}

	// config.toml has prompted token
	var cfg config.Config
	if _, err := toml.DecodeFile(filepath.Join(root, "config.toml"), &cfg); err != nil {
		t.Fatalf("config: %v", err)
	}
	if cfg.Telegram.Token != "test-token" {
		t.Errorf("token = %q, want %q", cfg.Telegram.Token, "test-token")
	}

	// unit file written
	unitFile := filepath.Join(unitDir, "comms-test.service")
	data, err := os.ReadFile(unitFile)
	if err != nil {
		t.Fatalf("unit file: %v", err)
	}
	if !strings.Contains(string(data), "daemon run --dir "+root) {
		t.Errorf("unit file missing ExecStart with absolute dir\n%s", data)
	}

	// systemctl calls: daemon-reload, enable
	if len(*calls) != 2 {
		t.Fatalf("systemctl calls = %d, want 2: %v", len(*calls), *calls)
	}
	if !containsArg((*calls)[0], "daemon-reload") {
		t.Errorf("call 0 = %v, want daemon-reload", (*calls)[0])
	}
	if !containsArg((*calls)[1], "enable") {
		t.Errorf("call 1 = %v, want enable", (*calls)[1])
	}

	// JSON output
	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output not JSON: %v\n%s", err, buf.String())
	}
	if result["status"] != "installed" {
		t.Errorf("status = %q, want installed", result["status"])
	}
	if result["service"] != "comms-test" {
		t.Errorf("service = %q, want comms-test", result["service"])
	}
}

func TestInstallSkipsPromptWhenTokenExists(t *testing.T) {
	tmpDir := t.TempDir()
	root := filepath.Join(tmpDir, ".comms")
	unitDir := filepath.Join(tmpDir, "units")
	os.MkdirAll(unitDir, 0o755)

	// Pre-create config with token
	os.MkdirAll(root, 0o755)
	f, _ := os.Create(filepath.Join(root, "config.toml"))
	toml.NewEncoder(f).Encode(config.Config{
		Telegram: config.TelegramConfig{Token: "existing"},
		General:  config.GeneralConfig{Format: "markdown"},
	})
	f.Close()

	stubSystemd(t, unitDir, false)

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	// No stdin — would block if prompt fires
	cmd.SetIn(strings.NewReader(""))
	cmd.SetArgs([]string{"daemon", "install", "--dir", root, "--name", "test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install: %v", err)
	}

	// Token unchanged
	var cfg config.Config
	toml.DecodeFile(filepath.Join(root, "config.toml"), &cfg)
	if cfg.Telegram.Token != "existing" {
		t.Errorf("token = %q, want existing", cfg.Telegram.Token)
	}
}

func TestInstallRestartsWhenActive(t *testing.T) {
	tmpDir := t.TempDir()
	root := filepath.Join(tmpDir, ".comms")
	unitDir := filepath.Join(tmpDir, "units")
	os.MkdirAll(unitDir, 0o755)

	calls := stubSystemd(t, unitDir, true) // service is active

	// Pre-create config with token
	os.MkdirAll(root, 0o755)
	f, _ := os.Create(filepath.Join(root, "config.toml"))
	toml.NewEncoder(f).Encode(config.Config{
		Telegram: config.TelegramConfig{Token: "tok"},
		General:  config.GeneralConfig{Format: "markdown"},
	})
	f.Close()

	cmd := newRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetArgs([]string{"daemon", "install", "--dir", root, "--name", "test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install: %v", err)
	}

	// Should see: daemon-reload, enable, restart
	if len(*calls) != 3 {
		t.Fatalf("systemctl calls = %d, want 3: %v", len(*calls), *calls)
	}
	if !containsArg((*calls)[2], "restart") {
		t.Errorf("call 2 = %v, want restart", (*calls)[2])
	}
}

func TestInstallStartFlag(t *testing.T) {
	tmpDir := t.TempDir()
	root := filepath.Join(tmpDir, ".comms")
	unitDir := filepath.Join(tmpDir, "units")
	os.MkdirAll(unitDir, 0o755)

	calls := stubSystemd(t, unitDir, false) // not active

	os.MkdirAll(root, 0o755)
	f, _ := os.Create(filepath.Join(root, "config.toml"))
	toml.NewEncoder(f).Encode(config.Config{
		Telegram: config.TelegramConfig{Token: "tok"},
		General:  config.GeneralConfig{Format: "markdown"},
	})
	f.Close()

	cmd := newRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetArgs([]string{"daemon", "install", "--dir", root, "--name", "test", "--start"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install: %v", err)
	}

	// Should see: daemon-reload, enable, start
	if len(*calls) != 3 {
		t.Fatalf("systemctl calls = %d, want 3: %v", len(*calls), *calls)
	}
	if !containsArg((*calls)[2], "start") {
		t.Errorf("call 2 = %v, want start", (*calls)[2])
	}
}

func TestInstallNoStartByDefault(t *testing.T) {
	tmpDir := t.TempDir()
	root := filepath.Join(tmpDir, ".comms")
	unitDir := filepath.Join(tmpDir, "units")
	os.MkdirAll(unitDir, 0o755)

	calls := stubSystemd(t, unitDir, false)

	os.MkdirAll(root, 0o755)
	f, _ := os.Create(filepath.Join(root, "config.toml"))
	toml.NewEncoder(f).Encode(config.Config{
		Telegram: config.TelegramConfig{Token: "tok"},
		General:  config.GeneralConfig{Format: "markdown"},
	})
	f.Close()

	cmd := newRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetArgs([]string{"daemon", "install", "--dir", root, "--name", "test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install: %v", err)
	}

	// Should see: daemon-reload, enable (no start/restart)
	if len(*calls) != 2 {
		t.Fatalf("systemctl calls = %d, want 2: %v", len(*calls), *calls)
	}
}

func TestInstallTokenPromptRejectsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	root := filepath.Join(tmpDir, ".comms")
	unitDir := filepath.Join(tmpDir, "units")
	os.MkdirAll(unitDir, 0o755)

	stubSystemd(t, unitDir, false)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"daemon", "install", "--dir", root, "--name", "test"})
	cmd.SetIn(strings.NewReader("\n"))

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestInstallTokenPromptRejectsEOF(t *testing.T) {
	tmpDir := t.TempDir()
	root := filepath.Join(tmpDir, ".comms")
	unitDir := filepath.Join(tmpDir, "units")
	os.MkdirAll(unitDir, 0o755)

	stubSystemd(t, unitDir, false)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"daemon", "install", "--dir", root, "--name", "test"})
	cmd.SetIn(strings.NewReader(""))

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for EOF")
	}
}

func TestUninstallRemovesUnitFile(t *testing.T) {
	tmpDir := t.TempDir()
	unitDir := filepath.Join(tmpDir, "units")
	os.MkdirAll(unitDir, 0o755)

	// Create a unit file to remove
	unitFile := filepath.Join(unitDir, "comms-test.service")
	os.WriteFile(unitFile, []byte("[Unit]\n"), 0o644)

	calls := stubSystemd(t, unitDir, false)

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"daemon", "uninstall", "--name", "test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("uninstall: %v", err)
	}

	// Unit file removed
	if _, err := os.Stat(unitFile); !os.IsNotExist(err) {
		t.Error("unit file should be removed")
	}

	// systemctl calls: stop, disable, daemon-reload
	if len(*calls) != 3 {
		t.Fatalf("systemctl calls = %d, want 3: %v", len(*calls), *calls)
	}
	if !containsArg((*calls)[0], "stop") {
		t.Errorf("call 0 = %v, want stop", (*calls)[0])
	}
	if !containsArg((*calls)[1], "disable") {
		t.Errorf("call 1 = %v, want disable", (*calls)[1])
	}
	if !containsArg((*calls)[2], "daemon-reload") {
		t.Errorf("call 2 = %v, want daemon-reload", (*calls)[2])
	}

	// JSON output
	var result map[string]string
	json.Unmarshal(buf.Bytes(), &result)
	if result["status"] != "uninstalled" {
		t.Errorf("status = %q, want uninstalled", result["status"])
	}
}

func TestUninstallIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	unitDir := filepath.Join(tmpDir, "units")
	os.MkdirAll(unitDir, 0o755)

	// No unit file exists
	stubSystemd(t, unitDir, false)

	cmd := newRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetArgs([]string{"daemon", "uninstall", "--name", "test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("uninstall should be idempotent: %v", err)
	}
}

// --- Darwin (launchd) tests ---

func prepareConfigDir(t *testing.T, tmpDir string) string {
	t.Helper()
	root := filepath.Join(tmpDir, ".comms")
	os.MkdirAll(root, 0o755)
	f, _ := os.Create(filepath.Join(root, "config.toml"))
	toml.NewEncoder(f).Encode(config.Config{
		Telegram: config.TelegramConfig{Token: "tok"},
		General:  config.GeneralConfig{Format: "markdown"},
	})
	f.Close()
	return root
}

func TestInstallDarwinWritesPlist(t *testing.T) {
	tmpDir := t.TempDir()
	root := prepareConfigDir(t, tmpDir)
	agentDir := filepath.Join(tmpDir, "LaunchAgents")
	os.MkdirAll(agentDir, 0o755)

	calls := stubLaunchd(t, agentDir, false)

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"daemon", "install", "--dir", root, "--name", "test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install darwin: %v", err)
	}

	// Plist file written
	plistPath := filepath.Join(agentDir, "com.wedow.comms-test.plist")
	data, err := os.ReadFile(plistPath)
	if err != nil {
		t.Fatalf("plist file: %v", err)
	}
	if !strings.Contains(string(data), "<string>com.wedow.comms-test</string>") {
		t.Errorf("plist missing label\n%s", data)
	}
	if !strings.Contains(string(data), "<string>daemon</string>") {
		t.Errorf("plist missing daemon arg\n%s", data)
	}

	// Should see: load (always loads to register the agent)
	if len(*calls) != 1 {
		t.Fatalf("launchctl calls = %d, want 1: %v", len(*calls), *calls)
	}
	if !containsArg((*calls)[0], "load") {
		t.Errorf("call 0 = %v, want load", (*calls)[0])
	}

	// JSON output
	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output not JSON: %v\n%s", err, buf.String())
	}
	if result["status"] != "installed" {
		t.Errorf("status = %q, want installed", result["status"])
	}
	if result["plist"] != plistPath {
		t.Errorf("plist = %q, want %q", result["plist"], plistPath)
	}
}

func TestInstallDarwinStartFlag(t *testing.T) {
	tmpDir := t.TempDir()
	root := prepareConfigDir(t, tmpDir)
	agentDir := filepath.Join(tmpDir, "LaunchAgents")
	os.MkdirAll(agentDir, 0o755)

	calls := stubLaunchd(t, agentDir, false)

	cmd := newRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetArgs([]string{"daemon", "install", "--dir", root, "--name", "test", "--start"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install darwin --start: %v", err)
	}

	// Should see: load
	if len(*calls) != 1 {
		t.Fatalf("launchctl calls = %d, want 1: %v", len(*calls), *calls)
	}
	if !containsArg((*calls)[0], "load") {
		t.Errorf("call 0 = %v, want load", (*calls)[0])
	}
}

func TestInstallDarwinReloadsWhenLoaded(t *testing.T) {
	tmpDir := t.TempDir()
	root := prepareConfigDir(t, tmpDir)
	agentDir := filepath.Join(tmpDir, "LaunchAgents")
	os.MkdirAll(agentDir, 0o755)

	calls := stubLaunchd(t, agentDir, true) // already loaded

	cmd := newRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetArgs([]string{"daemon", "install", "--dir", root, "--name", "test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install darwin reload: %v", err)
	}

	// Should see: unload, load
	if len(*calls) != 2 {
		t.Fatalf("launchctl calls = %d, want 2: %v", len(*calls), *calls)
	}
	if !containsArg((*calls)[0], "unload") {
		t.Errorf("call 0 = %v, want unload", (*calls)[0])
	}
	if !containsArg((*calls)[1], "load") {
		t.Errorf("call 1 = %v, want load", (*calls)[1])
	}
}

func TestUninstallDarwin(t *testing.T) {
	tmpDir := t.TempDir()
	agentDir := filepath.Join(tmpDir, "LaunchAgents")
	os.MkdirAll(agentDir, 0o755)

	// Create a plist file to remove
	plistPath := filepath.Join(agentDir, "com.wedow.comms-test.plist")
	os.WriteFile(plistPath, []byte("<plist/>"), 0o644)

	calls := stubLaunchd(t, agentDir, false)

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"daemon", "uninstall", "--name", "test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("uninstall darwin: %v", err)
	}

	// Plist file removed
	if _, err := os.Stat(plistPath); !os.IsNotExist(err) {
		t.Error("plist file should be removed")
	}

	// launchctl calls: unload
	if len(*calls) != 1 {
		t.Fatalf("launchctl calls = %d, want 1: %v", len(*calls), *calls)
	}
	if !containsArg((*calls)[0], "unload") {
		t.Errorf("call 0 = %v, want unload", (*calls)[0])
	}

	// JSON output
	var result map[string]string
	json.Unmarshal(buf.Bytes(), &result)
	if result["status"] != "uninstalled" {
		t.Errorf("status = %q, want uninstalled", result["status"])
	}
}

func TestUninstallDarwinIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	agentDir := filepath.Join(tmpDir, "LaunchAgents")
	os.MkdirAll(agentDir, 0o755)

	// No plist file exists
	stubLaunchd(t, agentDir, false)

	cmd := newRootCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetArgs([]string{"daemon", "uninstall", "--name", "test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("uninstall darwin should be idempotent: %v", err)
	}
}

func containsArg(args []string, want string) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}
