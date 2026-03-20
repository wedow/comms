package daemon

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"strings"
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

func TestSendCommand(t *testing.T) {
	// Set up a Subprocess with mocked pipes.
	stdinPR, stdinPW := io.Pipe()

	sub := &Subprocess{
		stdin:  stdinPW,
		events: make(chan any, 8),
		done:   make(chan error, 1),
	}

	// Read from stdin pipe in another goroutine.
	type result struct {
		cmd protocol.SendCommand
		err error
	}
	ch := make(chan result, 1)
	go func() {
		r := bufio.NewReader(stdinPR)
		evt, err := protocol.DecodeTyped(r)
		if err != nil {
			ch <- result{err: err}
			return
		}
		cmd, ok := evt.(protocol.SendCommand)
		if !ok {
			ch <- result{err: nil}
			return
		}
		ch <- result{cmd: cmd}
	}()

	err := sub.SendCommand(context.Background(), protocol.SendCommand{
		Type:   protocol.TypeSend,
		ID:     "abc",
		ChatID: 123,
		Text:   "hello",
	})
	if err != nil {
		t.Fatalf("SendCommand returned error: %v", err)
	}

	select {
	case res := <-ch:
		if res.err != nil {
			t.Fatalf("decode error: %v", res.err)
		}
		if res.cmd.ID != "abc" || res.cmd.Text != "hello" || res.cmd.ChatID != 123 {
			t.Errorf("unexpected command: %+v", res.cmd)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for command on stdin")
	}
}

func TestSendCommandCanceled(t *testing.T) {
	stdinPR, stdinPW := io.Pipe()
	defer stdinPR.Close()

	sub := &Subprocess{
		stdin:  stdinPW,
		events: make(chan any, 8),
		done:   make(chan error, 1),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := sub.SendCommand(ctx, protocol.SendCommand{
		Type: protocol.TypeSend,
		Text: "should not be written",
	})
	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestShutdownGraceful(t *testing.T) {
	stdinPR, stdinPW := io.Pipe()

	sub := &Subprocess{
		stdin:  stdinPW,
		events: make(chan any, 8),
		done:   make(chan error, 1),
	}

	// Read shutdown command from stdin and send shutdown_complete on events channel.
	go func() {
		r := bufio.NewReader(stdinPR)
		evt, err := protocol.DecodeTyped(r)
		if err != nil {
			return
		}
		cmd, ok := evt.(protocol.ShutdownCommand)
		if !ok || cmd.Type != protocol.TypeShutdown {
			return
		}
		sub.events <- protocol.ShutdownCompleteEvent{Type: protocol.TypeShutdownComplete}
	}()

	err := sub.Shutdown("test reason")
	if err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
}

func TestShutdownTimeout(t *testing.T) {
	stdinPR, stdinPW := io.Pipe()

	// We need a real process to kill. Use a simple cmd.
	cmd := exec.Command("sleep", "60")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start sleep process: %v", err)
	}

	sub := &Subprocess{
		cmd:    cmd,
		stdin:  stdinPW,
		events: make(chan any, 8),
		done:   make(chan error, 1),
	}

	// Post process exit to done channel when it dies.
	go func() {
		sub.done <- cmd.Wait()
	}()

	origTimeout := shutdownTimeout
	shutdownTimeout = 50 * time.Millisecond
	t.Cleanup(func() { shutdownTimeout = origTimeout })

	// Read stdin to unblock the write, but don't send shutdown_complete.
	go func() {
		r := bufio.NewReader(stdinPR)
		protocol.DecodeTyped(r)
		// intentionally don't send shutdown_complete
	}()

	err := sub.Shutdown("timeout test")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout error message, got: %v", err)
	}
}
