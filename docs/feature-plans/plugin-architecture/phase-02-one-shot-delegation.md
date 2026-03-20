---
title: "Phase 02: One-Shot Delegation"
status: reviewing
---

# Phase 02: One-Shot Delegation

## Overview

Establish the two-binary model where `comms send` and `comms react` delegate to `comms-telegram` via exec. Create the provider binary entry point with its own send/react commands. After this phase, `send.go`, `react.go`, and `root.go` in the base CLI have zero telegram SDK imports. The `daemon.go` still imports telegram (that is Phase 03).

## Tasks

### Task 02-1: Create CLI delegation infrastructure

Implement `resolveProviderBinary()`, `delegate()`, `extractProvider()` in `internal/cli/delegate.go`. Use swappable package-level vars for testing (same pattern as `runSystemctl` in `install.go`).

**File:** `internal/cli/delegate.go`

- `resolveProviderBinary` -- calls `exec.LookPath("comms-" + provider)`
- `delegate` -- spawns provider binary with args, inherits stdin/stdout/stderr, passes through exit code
- `extractProvider` -- extracts provider name from channel: `"telegram-general"` -> `"telegram"`

### Task 02-2: Create delegation tests

Test `extractProvider` with various channel names. Test `resolveProviderBinary` and `delegate` via swappable var overrides.

**File:** `internal/cli/delegate_test.go`

### Task 02-3: Create provider binary entry point

Create `cmd/comms-telegram/main.go` -- thin Cobra binary with `send` and `react` subcommands.

**File:** `cmd/comms-telegram/main.go`

This is a separate `main` package that builds as `comms-telegram`. Imports `providers/telegram` for command factories.

### Task 02-4: Create provider-specific CLI commands

Create `providers/telegram/cli.go` with `NewSendCmd()` and `NewReactCmd()` Cobra factories. These call telegram SDK directly (not via delegation). Reads config using the `COMMS_PROVIDER_CONFIG` env var (set by the base binary via `config.ProviderConfig(provider)` from Phase 01).

**File:** `providers/telegram/cli.go`

The provider binary is a pure network client: it sends/receives via the Telegram API and returns results as JSON on stdout. It never touches the store. The `--format` flag is interpreted here (not in the base binary — `parseFormatFlag` is deleted from `send.go`). Config is read from `COMMS_PROVIDER_CONFIG`; there is no reference to `config.TelegramToken()` here.

Key difference from current `cli/send.go`: the provider CLI reads config from the env var and calls `telegram.NewBot(token)` directly (no `newBot` parameter injection).

### Task 02-5: Refactor `internal/cli/send.go` to delegate

Replace the current send implementation with delegation logic: read `chat_id` from the store, extract provider from channel, resolve `comms-<provider>` binary, forward explicit named flags as argv.

**File:** `internal/cli/send.go`

The base binary retains store operations (`ReadChatID`, `WriteMessage`, `WriteCursor`) — it parses the provider binary's JSON stdout and handles all store persistence. Delete `parseFormatFlag` from `send.go` entirely; the `--format` flag is forwarded as a raw string to the provider binary, which interprets it.

`delegate()` constructs an explicit named argv — NOT raw `os.Args` passthrough. The flags forwarded are: `--channel`, `--format`, `--reply-to`, `--file`, `--media-type`, `--thread`, `--chat-id` (value read from store via `ReadChatID`), and body args.

Remove imports: `go-telegram/bot/models`, `providers/telegram`. Keep: `store`, `message`, `config`.

**Implementation order note:** Task 02-4 must be implemented before Task 02-3, because `cmd/comms-telegram/main.go` imports `providers/telegram` which is created in 02-4.

### Task 02-6: Refactor `internal/cli/react.go` to delegate

Same pattern as send.go.

**File:** `internal/cli/react.go`

Remove imports: `go-telegram/bot`, `go-telegram/bot/models`, `providers/telegram`, `store`.

### Task 02-7: Update `internal/cli/root.go`

Remove `telegram.NewBot` parameter from `newSendCmd()` and `newReactCmd()` calls. Remove telegram import.

**File:** `internal/cli/root.go`

### Task 02-8: Create provider CLI tests

Move detailed telegram API mocking tests from `internal/cli/send_test.go` and `internal/cli/react_test.go` to the provider package.

**File:** `providers/telegram/cli_test.go`

### Task 02-9: Rewrite send_test.go for delegation

Replace telegram-mock tests with delegation tests: verify correct args are passed, provider resolution, error propagation.

**File:** `internal/cli/send_test.go`

### Task 02-10: Rewrite react_test.go for delegation

Same pattern as send_test.go rewrite.

**File:** `internal/cli/react_test.go`

## Verification

```bash
go build -o /tmp/comms ./cmd/comms
go build -o /tmp/comms-telegram ./cmd/comms-telegram
go test ./internal/cli/ -v
go test ./providers/telegram/ -v
go vet ./...

# Verify no telegram imports in base CLI (except daemon.go)
grep -r "go-telegram/bot" internal/cli/send.go internal/cli/react.go internal/cli/root.go
# Expect: no output

# Verify daemon.go still has telegram (expected for Phase 03)
grep -c "go-telegram/bot" internal/cli/daemon.go
# Expect: > 0
```

## Dependencies

- Phase 00 must be complete (telegram at `providers/telegram/`)
- Phase 01 should be complete (config changes available), but Tasks 02-1 through 02-4 can start immediately after Phase 00 since the provider binary can use the old config API initially

## Post-Phase State

- Two binaries build: `comms` and `comms-telegram`
- `comms send` and `comms react` delegate to `comms-telegram` via exec
- Base CLI `send.go`, `react.go`, `root.go` have zero telegram SDK imports; `send.go` retains `store`, `message`, `config` imports for persistence
- Provider binary (`comms-telegram`) is a pure network client: no store access, reads config from `COMMS_PROVIDER_CONFIG`, returns JSON on stdout
- `daemon.go` still has telegram imports (Phase 03 scope)
