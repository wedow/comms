package daemon

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestExecCallbackEnvVars(t *testing.T) {
	tmp := t.TempDir()
	outFile := tmp + "/env.txt"

	env := CallbackEnv{
		File:     "/tmp/msg-001.md",
		Channel:  "telegram-general",
		Provider: "telegram",
		Sender:   "alice",
	}

	err := ExecCallback("env > "+outFile, env)
	if err != nil {
		t.Fatalf("ExecCallback: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}

	lines := string(data)
	for _, want := range []string{
		"COMMS_FILE=/tmp/msg-001.md",
		"COMMS_CHANNEL=telegram-general",
		"COMMS_PROVIDER=telegram",
		"COMMS_SENDER=alice",
	} {
		if !strings.Contains(lines, want) {
			t.Errorf("env output missing %q", want)
		}
	}
}

func TestExecCallbackUsesShell(t *testing.T) {
	tmp := t.TempDir()
	outFile := tmp + "/shell.txt"

	// Pipe is a shell feature; if this works, the command ran through a shell
	env := CallbackEnv{File: "f", Channel: "c", Provider: "p", Sender: "s"}
	err := ExecCallback("echo hello | tr h H > "+outFile, env)
	if err != nil {
		t.Fatalf("ExecCallback: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}

	got := strings.TrimSpace(string(data))
	if got != "Hello" {
		t.Errorf("got %q, want %q", got, "Hello")
	}
}

func TestExecCallbackAsync(t *testing.T) {
	env := CallbackEnv{File: "f", Channel: "c", Provider: "p", Sender: "s"}

	start := time.Now()
	err := ExecCallback("sleep 5", env)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("ExecCallback: %v", err)
	}

	if elapsed > 100*time.Millisecond {
		t.Errorf("ExecCallback blocked for %v, expected async return", elapsed)
	}
}

func TestCallbackRunnerDebouncesRapidCalls(t *testing.T) {
	tmp := t.TempDir()
	marker := tmp + "/count.txt"

	cmd := "printf x >> " + marker
	delay := 100 * time.Millisecond
	runner := NewCallbackRunner(cmd, delay)
	env := CallbackEnv{File: "f", Channel: "c", Provider: "p", Sender: "s"}

	// Rapid calls — each resets the timer
	runner.Run(env)
	runner.Run(env)
	runner.Run(env)

	// Callback hasn't fired yet (timer just reset)
	time.Sleep(50 * time.Millisecond)
	if _, err := os.Stat(marker); err == nil {
		t.Error("callback should not fire before debounce delay")
	}

	// Wait for debounce to fire
	time.Sleep(delay + 100*time.Millisecond)

	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("reading marker file: %v", err)
	}
	if got := len(data); got != 1 {
		t.Errorf("expected 1 execution after debounce, got %d", got)
	}
}

func TestCallbackRunnerFiresSeparatelyAfterQuietPeriod(t *testing.T) {
	tmp := t.TempDir()
	marker := tmp + "/count.txt"

	cmd := "printf x >> " + marker
	delay := 50 * time.Millisecond
	runner := NewCallbackRunner(cmd, delay)
	env := CallbackEnv{File: "f", Channel: "c", Provider: "p", Sender: "s"}

	// First call
	runner.Run(env)
	time.Sleep(delay + 100*time.Millisecond) // wait for it to fire

	// Second call after quiet period
	runner.Run(env)
	time.Sleep(delay + 100*time.Millisecond) // wait for it to fire

	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("reading marker file: %v", err)
	}
	if got := len(data); got != 2 {
		t.Errorf("expected 2 executions, got %d", got)
	}
}

func TestCallbackRunnerZeroDelayFiresImmediately(t *testing.T) {
	tmp := t.TempDir()
	marker := tmp + "/count.txt"

	cmd := "printf x >> " + marker
	runner := NewCallbackRunner(cmd, 0)
	env := CallbackEnv{File: "f", Channel: "c", Provider: "p", Sender: "s"}

	runner.Run(env)
	runner.Run(env)

	time.Sleep(200 * time.Millisecond)

	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("reading marker file: %v", err)
	}
	if got := len(data); got != 2 {
		t.Errorf("expected 2 executions with zero delay, got %d", got)
	}
}

func TestCallbackRunnerUsesLatestEnv(t *testing.T) {
	tmp := t.TempDir()
	outFile := tmp + "/sender.txt"

	cmd := "printf $COMMS_SENDER > " + outFile
	delay := 50 * time.Millisecond
	runner := NewCallbackRunner(cmd, delay)

	runner.Run(CallbackEnv{File: "f", Channel: "c", Provider: "p", Sender: "first"})
	runner.Run(CallbackEnv{File: "f", Channel: "c", Provider: "p", Sender: "second"})
	runner.Run(CallbackEnv{File: "f", Channel: "c", Provider: "p", Sender: "third"})

	time.Sleep(delay + 200*time.Millisecond)

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	if got := strings.TrimSpace(string(data)); got != "third" {
		t.Errorf("sender = %q, want 'third' (latest)", got)
	}
}
