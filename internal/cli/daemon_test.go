package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
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
