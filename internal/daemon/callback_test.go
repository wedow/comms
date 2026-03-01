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

func TestCallbackRunnerThrottlesRapidCalls(t *testing.T) {
	tmp := t.TempDir()
	marker := tmp + "/count.txt"

	// Append "x" on each execution; count x's to know how many times it ran
	cmd := "printf x >> " + marker
	runner := NewCallbackRunner(cmd, 1*time.Second)
	env := CallbackEnv{File: "f", Channel: "c", Provider: "p", Sender: "s"}

	runner.Run(env)
	runner.Run(env) // should be throttled

	time.Sleep(200 * time.Millisecond) // let async exec finish

	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("reading marker file: %v", err)
	}
	if got := len(data); got != 1 {
		t.Errorf("expected 1 execution, got %d", got)
	}
}

func TestCallbackRunnerAllowsAfterDelay(t *testing.T) {
	tmp := t.TempDir()
	marker := tmp + "/count.txt"

	cmd := "printf x >> " + marker
	delay := 50 * time.Millisecond
	runner := NewCallbackRunner(cmd, delay)
	env := CallbackEnv{File: "f", Channel: "c", Provider: "p", Sender: "s"}

	runner.Run(env)
	time.Sleep(delay + 50*time.Millisecond) // wait past the delay
	runner.Run(env)

	time.Sleep(200 * time.Millisecond) // let async exec finish

	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("reading marker file: %v", err)
	}
	if got := len(data); got != 2 {
		t.Errorf("expected 2 executions, got %d", got)
	}
}
