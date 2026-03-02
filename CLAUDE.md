# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

comms is a Go CLI tool and daemon that gives AI agents a filesystem-based interface for sending/receiving messages across providers (Telegram only currently). Messages are plain text files (markdown or org-mode with YAML frontmatter), config is TOML, CLI output is JSON lines, callbacks are shell commands. See [docs/VISION.md](docs/VISION.md) for the full vision.

## Commands

```bash
go test ./...          # run all tests
go test ./internal/store  # run tests for one package
go test -run TestWriteMessage ./internal/store  # run a single test
go vet ./...           # lint
go build -o comms ./cmd/comms  # build binary
```

CI runs `go test ./...` and `go vet ./...`.

## Architecture

**Entry point:** `cmd/comms/main.go` → `cli.Execute()` (cobra root command)

**Four layers, top to bottom:**

1. **cli** (`internal/cli/`) — Cobra commands. Each file is one command (send.go, list.go, unread.go, daemon.go, etc.). All output via `PrintJSON()` — one JSON object per line to stdout, errors to stderr.

2. **daemon** (`internal/daemon/`) — Long-running poll loop. `daemon.Run()` takes a `Provider` interface, polls for messages, writes them to the store, fires callbacks. The `Provider` interface decouples the daemon from Telegram specifics.

3. **store** (`internal/store/`) — Filesystem persistence. Messages written as timestamped files under `<root>/<provider>-<channel>/`. State files: `.cursor` (read position per channel), `offset.json` (Telegram update offset), `chat_id` (numeric chat ID per channel), `daemon.pid`. No database.

4. **provider/telegram** (`internal/provider/telegram/`) — Telegram bot SDK wrapper. `Poll()` for receiving, `Send()`/`SendMedia()` for sending. `BotAPI` interface abstracts the bot client for testing.

**Supporting packages:**
- `message` (`internal/message/`) — `Message` struct, markdown and org-mode serializers/deserializers
- `config` (`internal/config/`) — TOML config loading, env var override (`COMMS_TELEGRAM_TOKEN`)

**Filesystem layout (runtime):**
```
.comms/
  config.toml
  docs/                          # embedded setup guides
  daemon.pid
  telegram-<channel>/
    .cursor                      # RFC3339Nano timestamp of last read message
    chat_id                      # numeric Telegram chat ID
    2026-03-01T12-30-00Z.md      # message file (colons replaced with dashes)
    topic-<id>/                  # forum thread subdirectory
  offset.json                    # Telegram polling offset
```

## Key Patterns

- **Provider interface**: `daemon.Provider` has a single `Poll()` method. The daemon doesn't import telegram directly — `cli/daemon.go` bridges them via `telegramProvider`.
- **BotAPI interface**: `telegram.BotAPI` abstracts the Telegram bot SDK methods used, enabling test doubles throughout `telegram/` tests.
- **Callback debouncing**: `daemon.CallbackRunner` debounces shell callbacks with configurable delay. Zero delay = fire immediately.
- **Cursor-based unread**: `unread` command lists messages after the channel's `.cursor` timestamp. `ack` advances the cursor. Edits/reactions reset the cursor backward so the modified message reappears as unread.
- **Tests use `t.TempDir()`** for isolated filesystem state. CLI tests create a root command via `newRootCmd()`, set args, and capture stdout/stderr buffers.
- **Swappable functions for testing**: `install.go` uses package-level `var` functions (`runSystemctl`, `checkActive`, `getUnitDir`) replaced in tests.
