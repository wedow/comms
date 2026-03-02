package daemon

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// TypingIndicator is an optional interface that providers can implement to show
// a typing indicator while callbacks are running.
type TypingIndicator interface {
	SendTyping(ctx context.Context, chatID int64) error
}

// CallbackEnv holds the context passed to a callback command as environment variables.
type CallbackEnv struct {
	File     string
	Channel  string
	Provider string
	Sender   string
	ChatID   int64
}

// ExecCallback runs command asynchronously in a shell with COMMS_* env vars set.
// Stdout/stderr are logged. If typing is non-nil and ChatID is set, a typing
// indicator is sent for the duration of the command.
func ExecCallback(ctx context.Context, command string, env CallbackEnv, typing TypingIndicator) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}

	cmd := exec.Command(shell, "-c", command)

	// Build clean env, filtering out vars that interfere with child processes
	var cleanEnv []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "CLAUDECODE=") {
			continue
		}
		cleanEnv = append(cleanEnv, e)
	}
	cmd.Env = append(cleanEnv,
		"COMMS_FILE="+env.File,
		"COMMS_CHANNEL="+env.Channel,
		"COMMS_PROVIDER="+env.Provider,
		"COMMS_SENDER="+env.Sender,
	)

	go func() {
		var stopTyping func()
		if typing != nil && env.ChatID != 0 {
			stopTyping = startTypingLoop(ctx, typing, env.ChatID)
		}

		out, err := cmd.CombinedOutput()

		if stopTyping != nil {
			stopTyping()
		}
		if len(out) > 0 {
			log.Printf("callback: %s", out)
		}
		if err != nil {
			log.Printf("callback: %v", err)
		}
	}()
}

// startTypingLoop sends a typing indicator immediately and then every 5 seconds
// until the returned stop function is called.
func startTypingLoop(parent context.Context, typing TypingIndicator, chatID int64) func() {
	ctx, cancel := context.WithCancel(parent)
	go func() {
		if err := typing.SendTyping(ctx, chatID); err != nil {
			log.Printf("callback: typing: %v", err)
		}
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := typing.SendTyping(ctx, chatID); err != nil {
					log.Printf("callback: typing: %v", err)
				}
			}
		}
	}()
	return cancel
}

// CallbackRunner debounces callback execution. Each call to Run resets a timer;
// the callback fires once after delay elapses with no new calls.
// With zero delay, callbacks fire immediately.
type CallbackRunner struct {
	command string
	delay   time.Duration
	typing  TypingIndicator
	ctx     context.Context

	mu      sync.Mutex
	timer   *time.Timer
	lastEnv CallbackEnv
}

// NewCallbackRunner creates a CallbackRunner with the given command and debounce delay.
// typing may be nil if the provider does not support typing indicators.
func NewCallbackRunner(ctx context.Context, command string, delay time.Duration, typing TypingIndicator) *CallbackRunner {
	return &CallbackRunner{command: command, delay: delay, typing: typing, ctx: ctx}
}

// Run schedules the callback. With debounce delay, each call resets the timer
// so the callback only fires after a quiet period.
func (r *CallbackRunner) Run(env CallbackEnv) {
	if r.delay == 0 {
		ExecCallback(r.ctx, r.command, env, r.typing)
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
		ExecCallback(r.ctx, r.command, e, r.typing)
	})
}
