package daemon

import (
	"os"
	"os/exec"
	"sync"
	"time"
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

// CallbackRunner debounces callback execution. Each call to Run resets a timer;
// the callback fires once after delay elapses with no new calls.
// With zero delay, callbacks fire immediately.
type CallbackRunner struct {
	command string
	delay   time.Duration

	mu      sync.Mutex
	timer   *time.Timer
	lastEnv CallbackEnv
}

// NewCallbackRunner creates a CallbackRunner with the given command and debounce delay.
func NewCallbackRunner(command string, delay time.Duration) *CallbackRunner {
	return &CallbackRunner{command: command, delay: delay}
}

// Run schedules the callback. With debounce delay, each call resets the timer
// so the callback only fires after a quiet period.
func (r *CallbackRunner) Run(env CallbackEnv) {
	if r.delay == 0 {
		ExecCallback(r.command, env) //nolint:errcheck
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastEnv = env
	if r.timer != nil {
		r.timer.Stop()
	}
	r.timer = time.AfterFunc(r.delay, func() {
		r.mu.Lock()
		e := r.lastEnv
		r.mu.Unlock()
		ExecCallback(r.command, e) //nolint:errcheck
	})
}
