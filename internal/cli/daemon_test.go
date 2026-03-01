package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/wedow/comms/internal/config"
)

func TestDaemonStatusNoPID(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"daemon", "status", "--dir", tmpDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("daemon status failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if result["running"] != false {
		t.Errorf("running = %v, want false", result["running"])
	}
	if _, ok := result["pid"]; ok {
		t.Error("pid field should not be present when not running")
	}
}

func TestDaemonStatusStalePID(t *testing.T) {
	tmpDir := t.TempDir()

	// Write a stale PID (process almost certainly doesn't exist)
	pidPath := filepath.Join(tmpDir, "daemon.pid")
	if err := os.WriteFile(pidPath, []byte("9999999"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"daemon", "status", "--dir", tmpDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("daemon status failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if result["running"] != false {
		t.Errorf("running = %v, want false", result["running"])
	}

	// Stale PID file should have been cleaned up
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Error("stale PID file should have been removed")
	}
}

func TestDaemonStopNoDaemon(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := newRootCmd()
	stderr := new(bytes.Buffer)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"daemon", "stop", "--dir", tmpDir})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no daemon is running")
	}

	var errObj map[string]string
	if jsonErr := json.Unmarshal(stderr.Bytes(), &errObj); jsonErr != nil {
		t.Fatalf("stderr is not valid JSON: %v\nstderr: %s", jsonErr, stderr.String())
	}
	if errObj["error"] == "" {
		t.Error("expected non-empty error field in JSON output")
	}
}

func TestDaemonStopStalePID(t *testing.T) {
	tmpDir := t.TempDir()

	// Write a stale PID (process almost certainly doesn't exist)
	pidPath := filepath.Join(tmpDir, "daemon.pid")
	if err := os.WriteFile(pidPath, []byte("9999999"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	stderr := new(bytes.Buffer)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"daemon", "stop", "--dir", tmpDir})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when daemon PID is stale")
	}

	var errObj map[string]string
	if jsonErr := json.Unmarshal(stderr.Bytes(), &errObj); jsonErr != nil {
		t.Fatalf("stderr is not valid JSON: %v\nstderr: %s", jsonErr, stderr.String())
	}
	if errObj["error"] == "" {
		t.Error("expected non-empty error field in JSON output")
	}

	// Stale PID file should have been cleaned up
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Error("stale PID file should have been removed")
	}
}

func TestDaemonStartAlreadyRunning(t *testing.T) {
	tmpDir := t.TempDir()

	// Write a config.toml so config.Load succeeds
	cfgPath := filepath.Join(tmpDir, "config.toml")
	f, err := os.Create(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := toml.NewEncoder(f).Encode(config.Default()); err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Write current PID so IsRunning returns true
	pidPath := filepath.Join(tmpDir, "daemon.pid")
	if err := os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", os.Getpid())), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	stderr := new(bytes.Buffer)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"daemon", "start", "--dir", tmpDir})

	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error when daemon is already running")
	}

	// Stderr should contain JSON error about already running
	var errObj map[string]string
	if jsonErr := json.Unmarshal(stderr.Bytes(), &errObj); jsonErr != nil {
		t.Fatalf("stderr is not valid JSON: %v\nstderr: %s", jsonErr, stderr.String())
	}
	if errObj["error"] == "" {
		t.Error("expected non-empty error field in JSON output")
	}
}
