package daemon

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/wedow/comms/internal/protocol"
)

// Subprocess manages a provider plugin subprocess.
type Subprocess struct {
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	stdout   *bufio.Reader
	stderr   io.ReadCloser
	events   chan any
	done     chan error
	provider string
}

// readyTimeout is the deadline for reading the initial ready event.
// Swappable for testing.
var readyTimeout = 10 * time.Second

// Swappable for testing (same pattern as runSystemctl in install.go).
var (
	startProcess   = func(cmd *exec.Cmd) error { return cmd.Start() }
	cmdStdinPipe   = func(cmd *exec.Cmd) (io.WriteCloser, error) { return cmd.StdinPipe() }
	cmdStdoutPipe  = func(cmd *exec.Cmd) (io.ReadCloser, error) { return cmd.StdoutPipe() }
	cmdStderrPipe  = func(cmd *exec.Cmd) (io.ReadCloser, error) { return cmd.StderrPipe() }
)

// Spawn starts a provider subprocess and performs the ready/start handshake.
func Spawn(ctx context.Context, provider, binaryPath, root string, providerConfig []byte, offset int64) (*Subprocess, error) {
	cmd := exec.CommandContext(ctx, binaryPath, "subprocess")
	cmd.Env = append(cmd.Environ(),
		"COMMS_ROOT="+root,
		"COMMS_PROVIDER_CONFIG="+string(providerConfig),
	)

	stdinW, err := cmdStdinPipe(cmd)
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}
	stdoutR, err := cmdStdoutPipe(cmd)
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	stderrR, err := cmdStderrPipe(cmd)
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	if err := startProcess(cmd); err != nil {
		return nil, fmt.Errorf("start process: %w", err)
	}

	stdout := bufio.NewReader(stdoutR)

	// Read ready event with deadline.
	type readyResult struct {
		evt any
		err error
	}
	ch := make(chan readyResult, 1)
	go func() {
		evt, err := protocol.DecodeTyped(stdout)
		ch <- readyResult{evt, err}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			return nil, fmt.Errorf("reading ready event: %w", res.err)
		}
		if _, ok := res.evt.(protocol.ReadyEvent); !ok {
			return nil, fmt.Errorf("expected ready event, got %T", res.evt)
		}
	case <-time.After(readyTimeout):
		return nil, fmt.Errorf("timed out waiting for ready event")
	}

	// Send start command with offset.
	if err := protocol.Encode(stdinW, protocol.StartCommand{
		Type:   protocol.TypeStart,
		Offset: offset,
	}); err != nil {
		return nil, fmt.Errorf("sending start command: %w", err)
	}

	sub := &Subprocess{
		cmd:      cmd,
		stdin:    stdinW,
		stdout:   stdout,
		stderr:   stderrR,
		events:   make(chan any, 8),
		done:     make(chan error, 1),
		provider: provider,
	}

	// Stdout reader goroutine: decode events and post to channel.
	go func() {
		defer close(sub.events)
		for {
			evt, err := protocol.DecodeTyped(stdout)
			if err != nil {
				return
			}
			sub.events <- evt
		}
	}()

	// Stderr discard goroutine.
	go func() {
		io.Copy(io.Discard, stderrR)
	}()

	// Wait goroutine: posts exit status to done channel.
	go func() {
		sub.done <- cmd.Wait()
	}()

	return sub, nil
}
