# Phase 01: Shared Infrastructure

## Overview

Create the shared JSONL protocol package, refactor config for namespaced providers, and verify the store is provider-parameterized. After this phase, the protocol types exist, config supports `[providers.<name>]` sections with backward compatibility, and the daemon's store calls are confirmed provider-agnostic.

## Tasks

### Task 01-1: Create protocol types

Define all JSONL message types as Go structs in `internal/protocol/types.go`. Pure data types with JSON tags. No behavior.

**File:** `internal/protocol/types.go`

**Types (from protocol spec):**

Provider-to-daemon events:
- `ReadyEvent` -- `type`, `provider`, `version`
- `MessageEvent` -- offset, id, chat_id, channel, from, date, body, reply_to, reply_to_body, quote, thread_id, media_type, media_file_id, download_url, media_ext, caption, forward_from, forward_date, edit_date, media_group_id, entities
- `EditEvent` -- type alias or wrapper for `MessageEvent` (type is "edit")
- `ReactionEvent` -- offset, channel, message_id, from, emoji, date
- `ResponseEvent` -- id, ok, message (*MsgSummary), error
- `ErrorEvent` -- code, message
- `ShutdownCompleteEvent` -- just type field
- `PingEvent` / `PongEvent` -- type, ts (RFC3339)

Daemon-to-provider commands:
- `StartCommand` -- type, offset
- `SendCommand` -- type, id, chat_id, text, parse_mode, reply_to_id, thread_id
- `SendMediaCommand` -- type, id, chat_id, media_type, path, filename, caption, parse_mode, reply_to_id, thread_id
- `ReactCommand` -- type, id, chat_id, message_id, emoji
- `TypingCommand` -- type, chat_id
- `ShutdownCommand` -- type, reason

Supporting types:
- `MsgSummary` -- id, chat_id, channel, from, date, body
- `Entity` -- type, offset, length, url (separate from `message.Entity` to avoid import coupling)

### Task 01-2: Create protocol codec

Implement JSONL encode/decode in `internal/protocol/codec.go`.

- `Encode(w io.Writer, msg any) error` -- marshal to JSON, write with trailing newline
- `Decode(r *bufio.Reader) (map[string]any, error)` -- read one line, unmarshal to raw map
- `DecodeInto(r *bufio.Reader, target any) error` -- read one line, unmarshal into target struct
- 1 MiB max line length check
- Skip blank lines

### Task 01-3: Create protocol tests

Test round-trip encode/decode for all types, blank line skipping, oversized line rejection.

**File:** `internal/protocol/codec_test.go`

### Task 01-4: Add Providers map to Config struct

Change `Config` to support `[providers.<name>]` sections while keeping backward compatibility with `[telegram]`.

**File:** `internal/config/config.go`

```go
type Config struct {
    General   GeneralConfig
    Providers map[string]map[string]any `toml:"providers"`
    Callback  CallbackConfig
    // Deprecated: use Providers["telegram"]["token"] instead
    Telegram TelegramConfig `toml:"telegram"`
}
```

### Task 01-5: Add config migration logic

In `Load()`, if `[telegram]` exists but `[providers]` doesn't, auto-migrate by copying `Telegram` into `Providers["telegram"]`. Also handle `COMMS_TELEGRAM_TOKEN` env var override writing to `Providers`.

**File:** `internal/config/config.go`

### Task 01-6: Add ProviderConfig helper

Add `ProviderConfig(provider string) ([]byte, error)` that returns JSON serialization of a specific provider's config section (for passing as `COMMS_PROVIDER_CONFIG` env var).

**File:** `internal/config/config.go`

### Task 01-7: Update config tests

Add tests for new config format, migration, env var override, ProviderConfig serialization.

**File:** `internal/config/config_test.go`

### Task 01-8: Add TelegramToken backward-compat helper

Add `TelegramToken() string` method so existing callers (cli/daemon.go) continue working until Phase 03.

**File:** `internal/config/config.go`

### Task 01-9: Update init command for new config format

Update `Default()` to produce config with `[providers.telegram]` section instead of `[telegram]`.

**Files:** `internal/config/config.go` (Default function)

### Task 01-10: Verify store offset functions are provider-parameterized

Confirm `ReadOffset(root, provider)` and `WriteOffset(root, provider, offset)` already take a provider parameter. No code changes -- just verification. Note: `daemon.go` hardcodes `"telegram"` as the provider argument; this is fixed in Phase 03.

## Verification

```bash
go test ./internal/protocol/ -v
go test ./internal/config/ -v
go test ./internal/cli/ -run TestInit -v
go test ./...  # full suite
go vet ./...
```

## Dependencies

- Phase 00 must be complete (provider package at `providers/telegram/`)

## Post-Phase State

- `internal/protocol/` package exists with types and codec
- Config supports `[providers.<name>]` with `[telegram]` backward compat
- Store confirmed provider-parameterized (no changes needed)
