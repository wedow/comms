package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const pidFile = "daemon.pid"

func pidPath(dir string) string {
	return filepath.Join(dir, pidFile)
}

// WritePID writes the current process PID to dir/daemon.pid.
func WritePID(dir string) error {
	return os.WriteFile(pidPath(dir), []byte(strconv.Itoa(os.Getpid())), 0o644)
}

// ReadPID reads the PID integer from dir/daemon.pid.
func ReadPID(dir string) (int, error) {
	data, err := os.ReadFile(pidPath(dir))
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid pid file: %w", err)
	}
	return pid, nil
}

// IsRunning reads the PID file and checks whether the process is alive.
func IsRunning(dir string) bool {
	pid, err := ReadPID(dir)
	if err != nil {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}

// RemovePID removes dir/daemon.pid.
func RemovePID(dir string) error {
	return os.Remove(pidPath(dir))
}
