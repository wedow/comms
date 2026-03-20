package telegram_test

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/wedow/comms/internal/protocol"
)

func TestSubprocessIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Build the binary.
	binPath := t.TempDir() + "/comms-telegram"
	build := exec.Command("go", "build", "-o", binPath, "./cmd/comms-telegram")
	build.Dir = projectRoot(t)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	// Spawn the subprocess.
	cmd := exec.Command(binPath, "subprocess")
	cmd.Env = append(os.Environ(), `COMMS_PROVIDER_CONFIG={"token":"fake-token"}`)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() { cmd.Process.Kill() })

	reader := bufio.NewReader(stdout)

	// Helper to read one JSONL event with timeout.
	readEvt := func() map[string]any {
		t.Helper()
		type result struct {
			m   map[string]any
			err error
		}
		ch := make(chan result, 1)
		go func() {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				ch <- result{err: err}
				return
			}
			var m map[string]any
			if err := json.Unmarshal(line, &m); err != nil {
				ch <- result{err: err}
				return
			}
			ch <- result{m: m}
		}()
		select {
		case r := <-ch:
			if r.err != nil {
				t.Fatalf("readEvt: %v", r.err)
			}
			return r.m
		case <-time.After(10 * time.Second):
			t.Fatal("readEvt: timeout waiting for event")
			return nil
		}
	}

	// Helper to write a command.
	writeCmd := func(msg any) {
		t.Helper()
		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		data = append(data, '\n')
		if _, err := stdin.Write(data); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	// 1. Verify ready event.
	evt := readEvt()
	if evt["type"] != "ready" {
		t.Fatalf("expected ready, got %v", evt["type"])
	}
	if evt["provider"] != "telegram" {
		t.Errorf("provider = %v, want telegram", evt["provider"])
	}

	// 2. Send start command -- polling will fail with fake token (expected).
	writeCmd(protocol.StartCommand{Type: "start", Offset: 0})

	// 3. Send command -- should get error response (fake token).
	writeCmd(protocol.SendCommand{
		Type:   "send",
		ID:     "int-send-1",
		ChatID: 123,
		Text:   "hello",
	})
	evt = readEvt()
	if evt["type"] != "response" {
		t.Fatalf("expected response, got %v", evt["type"])
	}
	if evt["id"] != "int-send-1" {
		t.Errorf("id = %v, want int-send-1", evt["id"])
	}
	if evt["ok"] != false {
		t.Errorf("ok = %v, want false (fake token should fail)", evt["ok"])
	}

	// 4. React command -- should get error response (fake token).
	writeCmd(protocol.ReactCommand{
		Type:      "react",
		ID:        "int-react-1",
		ChatID:    123,
		MessageID: 1,
		Emoji:     "thumbs_up",
	})
	evt = readEvt()
	if evt["type"] != "response" {
		t.Fatalf("expected response, got %v", evt["type"])
	}
	if evt["id"] != "int-react-1" {
		t.Errorf("id = %v, want int-react-1", evt["id"])
	}
	if evt["ok"] != false {
		t.Errorf("ok = %v, want false (fake token should fail)", evt["ok"])
	}

	// 5. Shutdown command -- should get shutdown_complete.
	writeCmd(protocol.ShutdownCommand{Type: "shutdown"})
	evt = readEvt()
	if evt["type"] != "shutdown_complete" {
		t.Fatalf("expected shutdown_complete, got %v", evt["type"])
	}

	// 6. Verify process exits cleanly.
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("process exited with error: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("process did not exit within timeout")
	}
}

// projectRoot returns the repository root (two levels up from providers/telegram/).
func projectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return filepath.Join(dir, "../..")
}
