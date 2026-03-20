package daemon

import (
	"context"
	"fmt"
	"time"
)

// Swappable for testing.
var (
	spawnFunc              = Spawn
	sleepFunc              = contextSleep
	respawnBackoffCap      = 30 * time.Second
	respawnStableThreshold = 60 * time.Second
	respawnMaxFailures     = 5
)

// contextSleep sleeps for d, returning early if ctx is canceled.
func contextSleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// RespawnManager wraps Spawn with crash recovery and exponential backoff.
type RespawnManager struct {
	provider       string
	binaryPath     string
	root           string
	providerConfig []byte
	readOffset     func() int64
	events         chan any
}

// NewRespawnManager creates a RespawnManager.
func NewRespawnManager(provider, binaryPath, root string, providerConfig []byte, readOffset func() int64) *RespawnManager {
	return &RespawnManager{
		provider:       provider,
		binaryPath:     binaryPath,
		root:           root,
		providerConfig: providerConfig,
		readOffset:     readOffset,
		events:         make(chan any, 8),
	}
}

// Events returns the channel that receives forwarded subprocess events.
func (rm *RespawnManager) Events() <-chan any {
	return rm.events
}

// Run spawns the subprocess and respawns on crash with exponential backoff.
// Blocks until context cancellation or permanent failure (max consecutive crashes).
func (rm *RespawnManager) Run(ctx context.Context) error {
	defer close(rm.events)

	failures := 0
	backoff := 1 * time.Second

	for {
		if err := ctx.Err(); err != nil {
			return nil
		}

		offset := rm.readOffset()
		sub, err := spawnFunc(ctx, rm.provider, rm.binaryPath, rm.root, rm.providerConfig, offset)
		if err != nil {
			failures++
			if failures >= respawnMaxFailures {
				return fmt.Errorf("permanent failure after %d consecutive crashes: %w", failures, err)
			}
			if err := sleepFunc(ctx, backoff); err != nil {
				return nil // context canceled
			}
			backoff = nextBackoff(backoff)
			continue
		}

		spawnTime := time.Now()

		// Forward events and wait for crash.
	drain:
		for {
			select {
			case evt, ok := <-sub.events:
				if !ok {
					break drain
				}
				select {
				case rm.events <- evt:
				case <-ctx.Done():
					return nil
				}
			case <-sub.done:
				break drain
			case <-ctx.Done():
				return nil
			}
		}

		// Subprocess exited. Check stability.
		if time.Since(spawnTime) >= respawnStableThreshold {
			failures = 0
			backoff = 1 * time.Second
		} else {
			failures++
		}

		if failures >= respawnMaxFailures {
			return fmt.Errorf("permanent failure after %d consecutive crashes", failures)
		}

		if err := sleepFunc(ctx, backoff); err != nil {
			return nil // context canceled
		}
		backoff = nextBackoff(backoff)
	}
}

func nextBackoff(current time.Duration) time.Duration {
	next := current * 2
	if next > respawnBackoffCap {
		return respawnBackoffCap
	}
	return next
}
