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
