package daemon

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"testing"
	"time"

	"github.com/wedow/comms/internal/protocol"
)

func TestSpawnHandshake(t *testing.T) {
	// Pipes: we control the "process" side, Spawn reads/writes the "daemon" side.
	// stdout pipe: process writes -> Spawn reads
	stdoutPR, stdoutPW := io.Pipe()
	// stdin pipe: Spawn writes -> process reads
	stdinPR, stdinPW := io.Pipe()
	// stderr pipe
	stderrPR, stderrPW := io.Pipe()

	var captured *exec.Cmd

	origStart := startProcess
	startProcess = func(cmd *exec.Cmd) error {
		captured = cmd
		return nil
	}
	t.Cleanup(func() { startProcess = origStart })

	origStdin := cmdStdinPipe
	cmdStdinPipe = func(cmd *exec.Cmd) (io.WriteCloser, error) {
		return stdinPW, nil
	}
	t.Cleanup(func() { cmdStdinPipe = origStdin })

	origStdout := cmdStdoutPipe
	cmdStdoutPipe = func(cmd *exec.Cmd) (io.ReadCloser, error) {
		return stdoutPR, nil
	}
	t.Cleanup(func() { cmdStdoutPipe = origStdout })

	origStderr := cmdStderrPipe
	cmdStderrPipe = func(cmd *exec.Cmd) (io.ReadCloser, error) {
		return stderrPR, nil
	}
	t.Cleanup(func() { cmdStderrPipe = origStderr })

	// Goroutine simulates the subprocess: write ready, read start, write message.
	go func() {
		defer stdoutPW.Close()
		defer stderrPW.Close()

		// Write ready event
		_ = protocol.Encode(stdoutPW, protocol.ReadyEvent{
			Type:     protocol.TypeReady,
			Provider: "test",
			Version:  "1",
		})

		// Read start command from stdin
		r := bufio.NewReader(stdinPR)
		_, _ = protocol.DecodeTyped(r)

		// Write a message event
		_ = protocol.Encode(stdoutPW, protocol.MessageEvent{
			Type:    protocol.TypeMessage,
			Offset:  42,
			ID:      1,
			ChatID:  100,
			Channel: "test-chan",
			From:    "alice",
			Date:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Body:    "hello",
		})
	}()

	ctx := context.Background()
	sub, err := Spawn(ctx, "test", "/usr/bin/test-provider", "/tmp/root", []byte(`{"token":"abc"}`), 99)
	if err != nil {
		t.Fatalf("Spawn returned error: %v", err)
	}

	// Verify command was set up correctly
	if captured == nil {
		t.Fatal("startProcess was not called")
	}
	if len(captured.Args) < 2 || captured.Args[1] != "subprocess" {
		t.Errorf("expected arg 'subprocess', got %v", captured.Args)
	}

	// Verify env vars
	envMap := make(map[string]string)
	for _, e := range captured.Env {
		for i := 0; i < len(e); i++ {
			if e[i] == '=' {
				envMap[e[:i]] = e[i+1:]
				break
			}
		}
	}
	if envMap["COMMS_ROOT"] != "/tmp/root" {
		t.Errorf("COMMS_ROOT = %q, want /tmp/root", envMap["COMMS_ROOT"])
	}
	if envMap["COMMS_PROVIDER_CONFIG"] != `{"token":"abc"}` {
		t.Errorf("COMMS_PROVIDER_CONFIG = %q, want {\"token\":\"abc\"}", envMap["COMMS_PROVIDER_CONFIG"])
	}

	// Verify start command was sent (the goroutine read it above)
	// Read the start command directly from the stdin pipe to verify offset
	// Actually the goroutine already consumed it. Let's check the event channel instead.

	// Verify event received on events channel
	select {
	case evt, ok := <-sub.events:
		if !ok {
			t.Fatal("events channel closed before receiving message")
		}
		msg, ok := evt.(protocol.MessageEvent)
		if !ok {
			t.Fatalf("expected MessageEvent, got %T", evt)
		}
		if msg.Body != "hello" {
			t.Errorf("message body = %q, want hello", msg.Body)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestSpawnStartCommandOffset(t *testing.T) {
	// Verify that Spawn sends the correct offset in the start command.
	stdoutPR, stdoutPW := io.Pipe()
	stdinPR, stdinPW := io.Pipe()
	stderrPR, stderrPW := io.Pipe()

	origStart := startProcess
	startProcess = func(cmd *exec.Cmd) error { return nil }
	t.Cleanup(func() { startProcess = origStart })

	origStdin := cmdStdinPipe
	cmdStdinPipe = func(cmd *exec.Cmd) (io.WriteCloser, error) { return stdinPW, nil }
	t.Cleanup(func() { cmdStdinPipe = origStdin })

	origStdout := cmdStdoutPipe
	cmdStdoutPipe = func(cmd *exec.Cmd) (io.ReadCloser, error) { return stdoutPR, nil }
	t.Cleanup(func() { cmdStdoutPipe = origStdout })

	origStderr := cmdStderrPipe
	cmdStderrPipe = func(cmd *exec.Cmd) (io.ReadCloser, error) { return stderrPR, nil }
	t.Cleanup(func() { cmdStderrPipe = origStderr })

	var gotStart protocol.StartCommand
	startRead := make(chan struct{})

	go func() {
		defer stdoutPW.Close()
		defer stderrPW.Close()

		// Write ready
		_ = protocol.Encode(stdoutPW, protocol.ReadyEvent{
			Type:     protocol.TypeReady,
			Provider: "test",
			Version:  "1",
		})

		// Read start command
		r := bufio.NewReader(stdinPR)
		v, _ := protocol.DecodeTyped(r)
		gotStart = v.(protocol.StartCommand)
		close(startRead)
	}()

	ctx := context.Background()
	_, err := Spawn(ctx, "test", "/bin/fake", "/tmp/r", nil, 42)
	if err != nil {
		t.Fatalf("Spawn returned error: %v", err)
	}

	select {
	case <-startRead:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for start command")
	}

	if gotStart.Offset != 42 {
		t.Errorf("start command offset = %d, want 42", gotStart.Offset)
	}
	if gotStart.Type != protocol.TypeStart {
		t.Errorf("start command type = %q, want %q", gotStart.Type, protocol.TypeStart)
	}
}

func TestSpawnReadyTimeout(t *testing.T) {
	stdoutPR, stdoutPW := io.Pipe()
	_, stdinPW := io.Pipe()
	stderrPR, stderrPW := io.Pipe()

	origStart := startProcess
	startProcess = func(cmd *exec.Cmd) error { return nil }
	t.Cleanup(func() { startProcess = origStart })

	origStdin := cmdStdinPipe
	cmdStdinPipe = func(cmd *exec.Cmd) (io.WriteCloser, error) { return stdinPW, nil }
	t.Cleanup(func() { cmdStdinPipe = origStdin })

	origStdout := cmdStdoutPipe
	cmdStdoutPipe = func(cmd *exec.Cmd) (io.ReadCloser, error) { return stdoutPR, nil }
	t.Cleanup(func() { cmdStdoutPipe = origStdout })

	origStderr := cmdStderrPipe
	cmdStderrPipe = func(cmd *exec.Cmd) (io.ReadCloser, error) { return stderrPR, nil }
	t.Cleanup(func() { cmdStderrPipe = origStderr })

	origTimeout := readyTimeout
	readyTimeout = 50 * time.Millisecond
	t.Cleanup(func() { readyTimeout = origTimeout })

	// Never write anything to stdout — should timeout
	go func() {
		defer stderrPW.Close()
		// Keep stdout open but write nothing, will close after Spawn returns
		<-time.After(1 * time.Second)
		stdoutPW.Close()
	}()

	ctx := context.Background()
	_, err := Spawn(ctx, "test", "/bin/fake", "/tmp/r", nil, 0)
	if err == nil {
		t.Fatal("expected error for ready timeout, got nil")
	}
}

func TestSpawnBadReady(t *testing.T) {
	stdoutPR, stdoutPW := io.Pipe()
	_, stdinPW := io.Pipe()
	stderrPR, stderrPW := io.Pipe()

	origStart := startProcess
	startProcess = func(cmd *exec.Cmd) error { return nil }
	t.Cleanup(func() { startProcess = origStart })

	origStdin := cmdStdinPipe
	cmdStdinPipe = func(cmd *exec.Cmd) (io.WriteCloser, error) { return stdinPW, nil }
	t.Cleanup(func() { cmdStdinPipe = origStdin })

	origStdout := cmdStdoutPipe
	cmdStdoutPipe = func(cmd *exec.Cmd) (io.ReadCloser, error) { return stdoutPR, nil }
	t.Cleanup(func() { cmdStdoutPipe = origStdout })

	origStderr := cmdStderrPipe
	cmdStderrPipe = func(cmd *exec.Cmd) (io.ReadCloser, error) { return stderrPR, nil }
	t.Cleanup(func() { cmdStderrPipe = origStderr })

	// Write a non-ready event as first message
	go func() {
		defer stdoutPW.Close()
		defer stderrPW.Close()
		_ = protocol.Encode(stdoutPW, protocol.PingEvent{
			Type: protocol.TypePing,
			TS:   time.Now(),
		})
	}()

	ctx := context.Background()
	_, err := Spawn(ctx, "test", "/bin/fake", "/tmp/r", nil, 0)
	if err == nil {
		t.Fatal("expected error for bad ready event, got nil")
	}
}
