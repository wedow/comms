# Phase 03: Daemon Subprocess

## Overview

Replace the in-process `Provider` interface with a subprocess-per-provider model. This phase creates the subprocess manager (`daemon/provider.go`), the daemon core loop rewrite, callback refactoring, and removes all telegram coupling from `cli/daemon.go`. This is the largest phase and the point of no return -- once the `Provider` interface is removed, all existing daemon tests must be rewritten.

## Tasks

### Task 03-1: Config parsing for provider names

Add `ProviderNames()` method to config that returns sorted keys of the `Providers` map. The daemon uses this to determine which providers to spawn.

**File:** `internal/config/config.go`

### Task 03-2: Protocol event to message.Message conversion

Create `protocolToMessage(evt protocol.MessageEvent) message.Message` and `protocolToReaction(evt protocol.ReactionEvent)` conversion functions. Pure conversion, no side effects.

**File:** `internal/daemon/convert.go` (new)

### Task 03-3: Extract per-provider event handlers

Extract the inline handler logic from `daemon.Run()` into standalone functions:
- `handleMessageEvent(root, cfg, msg, chatID, cb)`
- `handleEditEvent(root, cfg, msg, chatID, cb)`
- `handleReactionEvent(root, cfg, evt, provider, cb)`

This is a pure refactor -- existing tests should still pass.

**File:** `internal/daemon/handler.go` (new)

### Task 03-4: Subprocess struct and spawn

Define `Subprocess` struct (cmd, pipes, state, event channel). Implement `Spawn(ctx, provider, binaryPath, root, providerConfig, offset) (*Subprocess, error)`:
1. Create `exec.Command(binaryPath, "subprocess")`
2. Set `COMMS_ROOT` and `COMMS_PROVIDER_CONFIG` env vars
3. Create stdin/stdout/stderr pipes
4. Start process
5. Read stdout with 10s deadline for `ready` event
6. Write `start` command with offset
7. Start stdout reader goroutine (posting to event channel)
8. Start stderr copy goroutine
9. Start `cmd.Wait()` goroutine for crash detection

Use swappable `startProcess` var for test injection.

**File:** `internal/daemon/provider.go` (new)

### Task 03-5: Subprocess shutdown

Implement `Subprocess.Shutdown(reason string) error`. Write `shutdown` command, wait 5s for `shutdown_complete`, then `cmd.Process.Kill()`.

**File:** `internal/daemon/provider.go`

### Task 03-6: Subprocess SendCommand

Implement `Subprocess.SendCommand(ctx, cmd) error`. Acquire write lock on stdin pipe, encode command, handle context cancellation. Only works when subprocess is in `Running` state.

**File:** `internal/daemon/provider.go`

### Task 03-7: Crash detection and respawn

Implement `RespawnManager` wrapping `Spawn` with crash recovery. Exponential backoff (1s, 2s, 4s, 8s, 30s cap), resets after 60s stable. 5 consecutive failures = fatal.

**File:** `internal/daemon/provider.go`

### Task 03-8: Subprocess tests

Tests using `os.Pipe()` pairs and helper goroutines simulating provider behavior. Test full spawn-handshake-running lifecycle, crash detection, shutdown timeout.

**File:** `internal/daemon/provider_test.go` (new)

### Task 03-9: Protocol I/O layer

Implement JSONL read/write on pipes using `internal/protocol` types. `DecodeEvent(r *json.Decoder) (ProtocolMessage, error)` and `EncodeCommand(w *json.Encoder, cmd Command) error`. This is the glue between the protocol package types and the pipe-based communication.

**File:** `internal/daemon/protocol.go` (new)

### Task 03-10: Protocol I/O tests

Round-trip encode/decode tests, blank line handling, oversized line rejection, unknown type error.

**File:** `internal/daemon/protocol_test.go` (new)

### Task 03-11: Remove Provider interface from daemon

Delete `Provider` and `TypingIndicator` interfaces from `daemon/daemon.go`. Replace `daemon.Run()` signature with `Run(ctx, cfg, root, providers []string) error` (stub for now).

**File:** `internal/daemon/daemon.go`

### Task 03-12: Daemon core loop with subprocess management

Implement the main `daemon.Run()` loop:
1. Write PID file
2. Load allowed IDs
3. Create CallbackRunner
4. For each provider: resolve binary, read offset, spawn subprocess, start reader goroutine
5. On each event: convert to message, call handler, persist offset
6. Wait for context cancellation (SIGTERM/SIGINT)
7. Send shutdown to all providers, wait for completion (5s timeout, then SIGKILL)
8. Remove PID file

Goroutine model: one reader per provider, main loop selects over channels.

**File:** `internal/daemon/daemon.go`

### Task 03-13: Typing indicator via protocol

Replace `TypingIndicator` interface with `TypingFunc func(ctx context.Context, provider string, chatID int64) error`. Update `CallbackRunner` to use function type instead of interface. The daemon implements `sendTypingCommand` that writes `TypingCommand` to the correct provider's stdin.

**File:** `internal/daemon/callback.go`

### Task 03-14: Refactor CLI daemon command

Remove `telegramProvider` struct, all `go-telegram/bot` imports, and direct `telegram.Poll()` call from `cli/daemon.go`. The `run` subcommand becomes: load config, check not already running, extract provider names, call `daemon.Run()`.

**File:** `internal/cli/daemon.go`

### Task 03-15: Crash recovery in daemon core

Implement crash recovery loop within daemon's per-provider goroutine: log exit, persist offset, wait with backoff, respawn. Handle binary-not-found (log once, skip). Handle 5 consecutive failures (skip permanently). Handle context cancellation during backoff.

**File:** `internal/daemon/daemon.go`

### Task 03-16: Media download from protocol events

Move `downloadMedia` into the handler path. `handleMessageEvent` calls `downloadMedia` when `msg.DownloadURL != ""`. No change to download logic.

**File:** `internal/daemon/daemon.go` or `internal/daemon/media.go`

### Task 03-17: Signal handling and graceful shutdown

Add signal handling in `daemon.Run()`: create a `signal.NotifyContext` for SIGTERM/SIGINT. On cancellation, call `Shutdown()` on all live subprocesses with 5s timeout, then `Process.Kill()` on stragglers. After all subprocesses exit, persist final offsets via `store.WriteOffset()` and remove PID file.

Add test: send SIGTERM to daemon process (or cancel context), verify all mock subprocesses receive shutdown command, verify offsets persisted, verify PID file removed.

**Files:**
- `internal/daemon/daemon.go` -- add signal context in `Run()`, add shutdown sequencing after context cancellation
- `internal/daemon/daemon_test.go` -- add `TestRunGracefulShutdown` and `TestRunForceKill`

### Task 03-18: Rewrite daemon tests

All existing daemon tests use `fakeProvider` implementing the old interface. Rewrite to use mock subprocesses:
- Create temp dir with a fake `comms-telegram` script (or inject `SpawnFunc` that creates `os.Pipe()` with JSONL streams)
- Write protocol events to pipe
- Verify store state after `Run()` completes

**File:** `internal/daemon/daemon_test.go`, `internal/daemon/daemon_integration_test.go`

### Task 03-19: Download media in handler tests

Verify `downloadMedia` works when called from the new handler path. Update existing `TestRunDownloadsMedia` test for the new infrastructure.

**File:** `internal/daemon/daemon_test.go`

## Verification

```bash
# Per-component tests
go test ./internal/daemon/ -run TestProtocol -v
go test ./internal/daemon/ -run TestSubprocess -v
go test ./internal/daemon/ -run TestRespawn -v
go test ./internal/daemon/ -run TestRun -v
go test ./internal/daemon/ -run TestIntegration -v
go test ./internal/cli/ -run TestDaemon -v

# Full suite
go test ./...
go vet ./...

# Verify no telegram imports remain in CLI
grep -r "go-telegram" internal/cli/
# Expect: no output

# Verify no telegram imports in daemon
grep -r "go-telegram" internal/daemon/
# Expect: no output
```

## Dependencies

- Phase 00 complete (provider at `providers/telegram/`)
- Phase 01 complete (protocol types in `internal/protocol/`, config `Providers` map)
- Phase 02 complete (provider binary entry point exists for reference, though daemon does not exec it for one-shot)

Internal task dependencies:
- Tasks 03-1 to 03-3 can run in parallel (pure refactor / additive)
- Tasks 03-4 to 03-10 are the subprocess infrastructure (sequential)
- Tasks 03-11 to 03-17 depend on 03-4 through 03-10
- Task 03-18 depends on all preceding tasks (rewrite tests after behavior changes)

## Post-Phase State

- `Provider` interface removed
- `TypingIndicator` interface replaced with `TypingFunc`
- Daemon spawns provider subprocesses, communicates via JSONL
- Zero telegram imports in `internal/cli/` and `internal/daemon/`
- All existing daemon tests rewritten for subprocess model
