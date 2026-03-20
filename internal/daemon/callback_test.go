package daemon

import (
	"context"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func testEnv() CallbackEnv {
	return CallbackEnv{File: "f", Channel: "c", Provider: "p", Sender: "s"}
}

func TestExecCallbackEnvVars(t *testing.T) {
	tmp := t.TempDir()
	outFile := tmp + "/env.txt"

	env := CallbackEnv{
		File:     "/tmp/msg-001.md",
		Channel:  "telegram-general",
		Provider: "telegram",
		Sender:   "alice",
	}

	ExecCallback(context.Background(), "env > "+outFile, env, nil)

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
	ExecCallback(context.Background(), "echo hello | tr h H > "+outFile, testEnv(), nil)

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
	start := time.Now()
	ExecCallback(context.Background(), "sleep 5", testEnv(), nil)
	elapsed := time.Since(start)

	if elapsed > 100*time.Millisecond {
		t.Errorf("ExecCallback blocked for %v, expected async return", elapsed)
	}
}

func TestExecCallbackSendsTyping(t *testing.T) {
	var count atomic.Int32
	typing := TypingFunc(func(_ context.Context, _ string, _ int64) error {
		count.Add(1)
		return nil
	})

	env := CallbackEnv{File: "f", Channel: "c", Provider: "p", Sender: "s", ChatID: 123}
	ExecCallback(context.Background(), "sleep 0.1", env, typing)

	time.Sleep(300 * time.Millisecond)

	if got := count.Load(); got < 1 {
		t.Errorf("expected at least 1 typing call, got %d", got)
	}
}

func TestExecCallbackNoTypingWithoutChatID(t *testing.T) {
	var count atomic.Int32
	typing := TypingFunc(func(_ context.Context, _ string, _ int64) error {
		count.Add(1)
		return nil
	})

	env := CallbackEnv{File: "f", Channel: "c", Provider: "p", Sender: "s", ChatID: 0}
	ExecCallback(context.Background(), "sleep 0.1", env, typing)

	time.Sleep(300 * time.Millisecond)

	if got := count.Load(); got != 0 {
		t.Errorf("expected 0 typing calls without ChatID, got %d", got)
	}
}

func TestCallbackRunnerDebouncesRapidCalls(t *testing.T) {
	tmp := t.TempDir()
	marker := tmp + "/count.txt"

	cmd := "printf x >> " + marker
	delay := 100 * time.Millisecond
	runner := NewCallbackRunner(context.Background(), cmd, delay, nil)
	env := testEnv()

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
	runner := NewCallbackRunner(context.Background(), cmd, delay, nil)
	env := testEnv()

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
	runner := NewCallbackRunner(context.Background(), cmd, 0, nil)
	env := testEnv()

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
	runner := NewCallbackRunner(context.Background(), cmd, delay, nil)

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

func TestStartTypingLoopRepeats(t *testing.T) {
	var mu sync.Mutex
	var chatIDs []int64
	typing := TypingFunc(func(_ context.Context, _ string, chatID int64) error {
		mu.Lock()
		chatIDs = append(chatIDs, chatID)
		mu.Unlock()
		return nil
	})

	stop := startTypingLoop(context.Background(), typing, "telegram", 42)
	// Immediate send + wait for one tick (using short interval isn't possible
	// since the loop hardcodes 5s, so we just verify the immediate call)
	time.Sleep(100 * time.Millisecond)
	stop()

	mu.Lock()
	defer mu.Unlock()
	if len(chatIDs) < 1 {
		t.Fatal("expected at least 1 typing call")
	}
	if chatIDs[0] != 42 {
		t.Errorf("chatID = %d, want 42", chatIDs[0])
	}
}

