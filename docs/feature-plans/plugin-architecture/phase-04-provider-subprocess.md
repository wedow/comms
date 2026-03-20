---
title: "Phase 04: Provider Subprocess"
status: reviewing
---

# Phase 04: Provider Subprocess

## Overview

Implement the JSONL subprocess mode in the telegram provider. This phase creates `providers/telegram/subprocess.go` (the stdin/stdout event loop) and integration tests that exercise the full protocol. The provider binary now operates in both one-shot mode (send, react -- built in Phase 02) and daemon subprocess mode (JSONL protocol).

## Tasks

### Task 04-1: Add subprocess command to provider binary

Add a hidden `subprocess` subcommand to `cmd/comms-telegram/main.go`. This command:
- Reads `COMMS_PROVIDER_CONFIG` env var (JSON)
- Parses into `telegram.ProviderConfig`
- Calls `telegram.RunSubprocess(ctx, os.Stdin, os.Stdout, cfg)`

**File:** `cmd/comms-telegram/main.go` (modify -- add subprocess subcommand)

### Task 04-2: Implement subprocess JSONL loop

**Prerequisite**: Add `SendChatAction(ctx context.Context, params *bot.SendChatActionParams) (bool, error)` to the `BotAPI` interface in `telegram.go`, and add a no-op stub for it to `mockBot` in the existing tests. The `typing` command calls this method, so it must be in the interface before `subprocess.go` is written.

Create `RunSubprocess(ctx, stdin, stdout, cfg) error` in `providers/telegram/subprocess.go`:

1. **Handshake**: Create bot client from config, send `{"type":"ready","provider":"telegram","version":"1"}` to stdout
2. **Read loop**: Read JSONL from stdin, dispatch by `type`:
   - `start`: start long-poll loop in goroutine, poll handler writes events to stdout
   - `send`: call `telegram.Send()`, write `response` event
   - `send_media`: open file, call `telegram.SendMedia()`, write `response` event
   - `react`: call `api.SetMessageReaction()`, write `response` event
   - `typing`: call `api.SendChatAction()`, no response
   - `shutdown`: cancel poll, wait for stop, write `shutdown_complete`, return nil
   - `ping`: write `pong`
3. **Error handling**: init failure writes `error` event. stdin close = shutdown. command failure writes `response` with `ok:false`.

Internal types:
- `ProviderConfig` struct: `Token string`
- `Subprocess` struct: bot client, stdout writer (mutex-protected), cancel func
- `messageToEvent(msg, offset) protocol.MessageEvent` -- converts `message.Message` to protocol event; use `internal/provider/telegram/convert.go` as the authoritative field mapping source
- `messageToSummary(msg) *protocol.MsgSummary` -- converts for responses

Use swappable `var subprocessPollFunc = Poll` for test injection (same pattern as `install.go`). The var's type is inferred from `Poll`'s signature; tests replace it with a function literal. No wrapper type needed.

**File:** `providers/telegram/subprocess.go` (new)

### Task 04-3: Subprocess unit tests

Comprehensive tests using `io.Pipe` for stdin/stdout. Construct `Subprocess` struct directly with a `mockBot` (same pattern as existing `send_test.go` — no injection hook needed):

1. Handshake: verify `ready` event on stdout
2. Start + message event: inject mock update via swappable poll func, verify `message` event
3. Send command: mock BotAPI, verify `response` with `ok:true`
4. Send error: mock failure, verify `response` with `ok:false`
5. React command: verify `response` with `ok:true`
6. Shutdown: send `shutdown`, verify `shutdown_complete`
7. Stdin close: close stdin, verify graceful exit
8. Protocol error: malformed JSON, verify logged but continues
9. Ping/pong: verify pong response

Test helper: `startSubprocess(t) (*Subprocess, *io.PipeReader, *io.PipeWriter)`.

**File:** `providers/telegram/subprocess_test.go` (new)

### Task 04-4: Integration test for subprocess mode

Build and run `comms-telegram` as a real subprocess. No mock injection — the integration test exercises the real binary end-to-end:

1. Build binary to temp path (`go build -o $tmpdir/comms-telegram ./cmd/comms-telegram`)
2. Spawn with `COMMS_PROVIDER_CONFIG={"token":"fake-token"}`
3. The `ready` event arrives before any API calls, so it is always observable regardless of token validity
4. Send `start`, then send/react commands — these will return errors from Telegram (expected with a fake token), which is testable by reading `response` events with `ok:false`
5. Send `shutdown`, verify `shutdown_complete`

**File:** `providers/telegram/subprocess_integration_test.go` (new, or added to subprocess_test.go with build tags)

### Task 04-5: Full build and test verification

Run `go mod tidy`. Build both binaries. Run all tests. Verify CI commands pass.

**Files:** `go.mod`, `go.sum` (auto-modified)

## Verification

```bash
go test ./providers/telegram/ -run TestSubprocess -v -count=1
go test ./providers/telegram/ -run TestIntegration -v -count=1
go test ./providers/telegram/ -v
go test ./internal/protocol/ -v
go test ./internal/daemon/ -v

go build -o /tmp/comms ./cmd/comms
go build -o /tmp/comms-telegram ./cmd/comms-telegram

/tmp/comms-telegram --version
/tmp/comms-telegram send --help
/tmp/comms-telegram react --help

go test ./...
go vet ./...
```

## Dependencies

- Phase 00 complete (telegram at `providers/telegram/`)
- Phase 01 complete (protocol types and codec in `internal/protocol/`)
- Phase 02 complete (provider binary entry point exists, one-shot commands work)
- Phase 03 complete (daemon subprocess manager exists, daemon core can consume provider subprocess output)

## Post-Phase State

- `comms-telegram` binary supports both one-shot (send, react) and subprocess (JSONL daemon) modes
- Full protocol compliance verified via integration tests
- The entire two-binary architecture is functional
