package daemon

import (
	"os"
	"os/exec"
)

// CallbackEnv holds the context passed to a callback command as environment variables.
type CallbackEnv struct {
	File     string
	Channel  string
	Provider string
	Sender   string
}

// ExecCallback runs command asynchronously in a shell with COMMS_* env vars set.
// It returns an error only if setup fails; execution errors are discarded (fire-and-forget).
func ExecCallback(command string, env CallbackEnv) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}

	cmd := exec.Command(shell, "-c", command)
	cmd.Env = append(os.Environ(),
		"COMMS_FILE="+env.File,
		"COMMS_CHANNEL="+env.Channel,
		"COMMS_PROVIDER="+env.Provider,
		"COMMS_SENDER="+env.Sender,
	)
	cmd.Stdout = nil
	cmd.Stderr = nil

	go cmd.Run() //nolint:errcheck

	return nil
}
