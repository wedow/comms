package daemon

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/wedow/comms/internal/protocol"
)

// mockSubprocess creates a Subprocess with pre-wired channels for testing.
func mockSubprocess(events []any, crashAfter time.Duration) *Subprocess {
	evCh := make(chan any, len(events)+1)
	done := make(chan error, 1)

	go func() {
		for _, e := range events {
			evCh <- e
		}
		if crashAfter > 0 {
			time.Sleep(crashAfter)
		}
		close(evCh)
		done <- errors.New("process exited")
	}()

	return &Subprocess{
		events: evCh,
		done:   done,
	}
}

func TestRespawnRecovery(t *testing.T) {
	orig := spawnFunc
	t.Cleanup(func() { spawnFunc = orig })

	var mu sync.Mutex
	spawnCount := 0

	spawnFunc = func(ctx context.Context, provider, binaryPath, root string, providerConfig []byte, offset int64) (*Subprocess, error) {
		mu.Lock()
		spawnCount++
		n := spawnCount
		mu.Unlock()

		return mockSubprocess([]any{fmt.Sprintf("event-from-spawn-%d", n)}, 10*time.Millisecond), nil
	}

	origThreshold := respawnStableThreshold
	respawnStableThreshold = 1 * time.Second // won't trigger stability reset
	t.Cleanup(func() { respawnStableThreshold = origThreshold })

	origMax := respawnMaxFailures
	respawnMaxFailures = 3
	t.Cleanup(func() { respawnMaxFailures = origMax })

	origSleep := sleepFunc
	sleepFunc = func(ctx context.Context, d time.Duration) error {
		return ctx.Err()
	}
	t.Cleanup(func() { sleepFunc = origSleep })

	ctx, cancel := context.WithCancel(context.Background())

	rm := NewRespawnManager("test", "/bin/fake", "/tmp/root", nil, func() int64 { return 0 })

	// Collect events in background
	var collected []any
	var collMu sync.Mutex
	go func() {
		for evt := range rm.Events() {
			collMu.Lock()
			collected = append(collected, evt)
			collMu.Unlock()
		}
	}()

	// Cancel after collecting events from 2 spawns
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	_ = rm.Run(ctx)

	collMu.Lock()
	defer collMu.Unlock()

	if len(collected) < 2 {
		t.Errorf("expected events from at least 2 spawns, got %d events: %v", len(collected), collected)
	}

	mu.Lock()
	defer mu.Unlock()
	if spawnCount < 2 {
		t.Errorf("expected at least 2 spawns, got %d", spawnCount)
	}
}

func TestRespawnBackoff(t *testing.T) {
	orig := spawnFunc
	t.Cleanup(func() { spawnFunc = orig })

	spawnFunc = func(ctx context.Context, provider, binaryPath, root string, providerConfig []byte, offset int64) (*Subprocess, error) {
		return mockSubprocess(nil, 0), nil // crash immediately
	}

	origMax := respawnMaxFailures
	respawnMaxFailures = 4
	t.Cleanup(func() { respawnMaxFailures = origMax })

	origThreshold := respawnStableThreshold
	respawnStableThreshold = 1 * time.Hour // won't trigger
	t.Cleanup(func() { respawnStableThreshold = origThreshold })

	var mu sync.Mutex
	var sleepDurations []time.Duration

	origSleep := sleepFunc
	sleepFunc = func(ctx context.Context, d time.Duration) error {
		mu.Lock()
		sleepDurations = append(sleepDurations, d)
		mu.Unlock()
		return ctx.Err()
	}
	t.Cleanup(func() { sleepFunc = origSleep })

	rm := NewRespawnManager("test", "/bin/fake", "/tmp/root", nil, func() int64 { return 0 })
	// Drain events
	go func() {
		for range rm.Events() {
		}
	}()

	err := rm.Run(context.Background())
	if err == nil {
		t.Fatal("expected permanent error, got nil")
	}

	mu.Lock()
	defer mu.Unlock()

	// With maxFailures=4, we get 3 sleeps (between failure 1->2, 2->3, 3->4)
	expected := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}
	if len(sleepDurations) != len(expected) {
		t.Fatalf("expected %d sleep calls, got %d: %v", len(expected), len(sleepDurations), sleepDurations)
	}
	for i, want := range expected {
		if sleepDurations[i] != want {
			t.Errorf("sleep[%d] = %v, want %v", i, sleepDurations[i], want)
		}
	}
}

func TestRespawnMaxFailures(t *testing.T) {
	orig := spawnFunc
	t.Cleanup(func() { spawnFunc = orig })

	var mu sync.Mutex
	spawnCount := 0

	spawnFunc = func(ctx context.Context, provider, binaryPath, root string, providerConfig []byte, offset int64) (*Subprocess, error) {
		mu.Lock()
		spawnCount++
		mu.Unlock()
		return mockSubprocess(nil, 0), nil
	}

	origMax := respawnMaxFailures
	respawnMaxFailures = 3
	t.Cleanup(func() { respawnMaxFailures = origMax })

	origThreshold := respawnStableThreshold
	respawnStableThreshold = 1 * time.Hour
	t.Cleanup(func() { respawnStableThreshold = origThreshold })

	origSleep := sleepFunc
	sleepFunc = func(ctx context.Context, d time.Duration) error {
		return ctx.Err()
	}
	t.Cleanup(func() { sleepFunc = origSleep })

	rm := NewRespawnManager("test", "/bin/fake", "/tmp/root", nil, func() int64 { return 0 })
	go func() {
		for range rm.Events() {
		}
	}()

	err := rm.Run(context.Background())
	if err == nil {
		t.Fatal("expected permanent error after max failures")
	}

	mu.Lock()
	defer mu.Unlock()
	if spawnCount != 3 {
		t.Errorf("expected exactly %d spawn attempts, got %d", 3, spawnCount)
	}
}

func TestRespawnStabilityReset(t *testing.T) {
	orig := spawnFunc
	t.Cleanup(func() { spawnFunc = orig })

	var mu sync.Mutex
	spawnCount := 0

	spawnFunc = func(ctx context.Context, provider, binaryPath, root string, providerConfig []byte, offset int64) (*Subprocess, error) {
		mu.Lock()
		spawnCount++
		n := spawnCount
		mu.Unlock()

		// First spawn runs longer than threshold, subsequent crash immediately
		if n == 1 {
			return mockSubprocess([]any{"stable-event"}, 100*time.Millisecond), nil
		}
		return mockSubprocess(nil, 0), nil
	}

	origMax := respawnMaxFailures
	respawnMaxFailures = 3
	t.Cleanup(func() { respawnMaxFailures = origMax })

	origThreshold := respawnStableThreshold
	respawnStableThreshold = 50 * time.Millisecond
	t.Cleanup(func() { respawnStableThreshold = origThreshold })

	origSleep := sleepFunc
	sleepFunc = func(ctx context.Context, d time.Duration) error {
		return ctx.Err()
	}
	t.Cleanup(func() { sleepFunc = origSleep })

	rm := NewRespawnManager("test", "/bin/fake", "/tmp/root", nil, func() int64 { return 0 })
	go func() {
		for range rm.Events() {
		}
	}()

	err := rm.Run(context.Background())
	if err == nil {
		t.Fatal("expected permanent error eventually")
	}

	mu.Lock()
	defer mu.Unlock()
	// Spawn 1: stable (runs > threshold) -> failure count resets to 0
	// Spawn 2: crash -> failures=1
	// Spawn 3: crash -> failures=2
	// Spawn 4: crash -> failures=3 -> permanent error
	// Total: 4 spawns (more than maxFailures=3 because of reset)
	if spawnCount <= respawnMaxFailures {
		t.Errorf("expected more than %d spawns due to stability reset, got %d", respawnMaxFailures, spawnCount)
	}
}

func TestRespawnContextCancel(t *testing.T) {
	orig := spawnFunc
	t.Cleanup(func() { spawnFunc = orig })

	// Subprocess that blocks forever until context cancel
	spawnFunc = func(ctx context.Context, provider, binaryPath, root string, providerConfig []byte, offset int64) (*Subprocess, error) {
		evCh := make(chan any, 1)
		done := make(chan error, 1)
		go func() {
			<-ctx.Done()
			close(evCh)
			done <- ctx.Err()
		}()
		return &Subprocess{events: evCh, done: done}, nil
	}

	rm := NewRespawnManager("test", "/bin/fake", "/tmp/root", nil, func() int64 { return 0 })
	go func() {
		for range rm.Events() {
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- rm.Run(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("expected nil error on context cancel, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after context cancel")
	}
}

func TestRespawnGracefulShutdown(t *testing.T) {
	orig := spawnFunc
	t.Cleanup(func() { spawnFunc = orig })

	origTimeout := shutdownTimeout
	shutdownTimeout = 2 * time.Second
	t.Cleanup(func() { shutdownTimeout = origTimeout })

	// Use an io.Pipe for stdin so we can read the shutdown command.
	stdinPR, stdinPW := io.Pipe()

	shutdownReceived := make(chan protocol.ShutdownCommand, 1)

	spawnFunc = func(ctx context.Context, provider, binaryPath, root string, providerConfig []byte, offset int64) (*Subprocess, error) {
		evCh := make(chan any, 8)
		done := make(chan error, 1)

		// Simulate subprocess: read stdin for shutdown command, then respond.
		go func() {
			r := bufio.NewReader(stdinPR)
			evt, err := protocol.DecodeTyped(r)
			if err != nil {
				return
			}
			cmd, ok := evt.(protocol.ShutdownCommand)
			if ok {
				shutdownReceived <- cmd
				// Respond with shutdown_complete so Shutdown() returns.
				evCh <- protocol.ShutdownCompleteEvent{Type: protocol.TypeShutdownComplete}
			}
		}()

		return &Subprocess{
			stdin:  stdinPW,
			events: evCh,
			done:   done,
		}, nil
	}

	rm := NewRespawnManager("test", "/bin/fake", "/tmp/root", nil, func() int64 { return 0 })
	go func() {
		for range rm.Events() {
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	runDone := make(chan error, 1)
	go func() {
		runDone <- rm.Run(ctx)
	}()

	// Let subprocess start.
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Verify shutdown command was sent.
	select {
	case cmd := <-shutdownReceived:
		if cmd.Type != protocol.TypeShutdown {
			t.Errorf("shutdown command type = %q, want %q", cmd.Type, protocol.TypeShutdown)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for shutdown command")
	}

	// Verify Run() returns nil.
	select {
	case err := <-runDone:
		if err != nil {
			t.Errorf("expected nil error on graceful shutdown, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after graceful shutdown")
	}
}
