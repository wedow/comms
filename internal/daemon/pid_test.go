package daemon

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteReadRemoveCycle(t *testing.T) {
	dir := t.TempDir()

	if err := WritePID(dir); err != nil {
		t.Fatalf("WritePID: %v", err)
	}

	pid, err := ReadPID(dir)
	if err != nil {
		t.Fatalf("ReadPID: %v", err)
	}
	if pid != os.Getpid() {
		t.Errorf("ReadPID = %d, want %d", pid, os.Getpid())
	}

	// Verify the file exists on disk
	if _, err := os.Stat(filepath.Join(dir, "daemon.pid")); err != nil {
		t.Fatalf("pid file should exist: %v", err)
	}

	if err := RemovePID(dir); err != nil {
		t.Fatalf("RemovePID: %v", err)
	}

	// Verify the file is gone
	if _, err := os.Stat(filepath.Join(dir, "daemon.pid")); !os.IsNotExist(err) {
		t.Fatalf("pid file should be removed, got err: %v", err)
	}
}

func TestIsRunningCurrentProcess(t *testing.T) {
	dir := t.TempDir()

	if err := WritePID(dir); err != nil {
		t.Fatalf("WritePID: %v", err)
	}

	if !IsRunning(dir) {
		t.Error("IsRunning should return true for current process")
	}
}

func TestIsRunningStalePID(t *testing.T) {
	dir := t.TempDir()

	// Write a fake high PID that almost certainly doesn't exist
	if err := os.WriteFile(filepath.Join(dir, "daemon.pid"), []byte("9999999"), 0o644); err != nil {
		t.Fatal(err)
	}

	if IsRunning(dir) {
		t.Error("IsRunning should return false for stale PID")
	}
}

func TestIsRunningNoPIDFile(t *testing.T) {
	dir := t.TempDir()

	if IsRunning(dir) {
		t.Error("IsRunning should return false when no PID file exists")
	}
}

func TestReadPIDMissingFile(t *testing.T) {
	dir := t.TempDir()

	_, err := ReadPID(dir)
	if err == nil {
		t.Fatal("ReadPID should return error for missing file")
	}
}
