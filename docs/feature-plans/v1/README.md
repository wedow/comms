---
title: comms v1
status: reviewing
review_passes: 1
last_review: 2026-03-01
blocking_issues_history:
  - pass: 1
    blocking_count: 18
    resolved: 18
---

# comms v1 Implementation Plan

Unified CLI tool and daemon giving AI agents a filesystem-based interface for Telegram messaging. Messages are plain text files, config is TOML, output is JSON lines, callbacks are shell commands.

## Total Tasks: 35

## Phases

| Phase | Tasks | Description |
|-------|-------|-------------|
| [Phase 1: Core Foundation](phase-01-core-foundation.md) | 9 | Go skeleton, message struct, serialization, config, store, embedded docs, `comms init` |
| [Phase 2: Telegram Provider](phase-02-telegram-provider.md) | 6 | BotAPI interface, chat slugification, send, poll, message converter, bot factory |
| [Phase 3: Daemon](phase-03-daemon.md) | 10 | Core loop, PID management, daemon CLI commands, callbacks, rate limiting, offset + chat ID persistence, integration test |
| [Phase 4: CLI Commands](phase-04-cli-commands.md) | 6 | JSON output helpers, `channels`, `list`, `unread`, `send`, command wiring |
| [Phase 5: Packaging & Release](phase-05-packaging-release.md) | 6 | GoReleaser, GitHub Actions CI, systemd unit, Homebrew tap, AUR PKGBUILD, launchd plist |

## Phase Dependencies

```
Phase 1 (Core Foundation)
  |
  +---> Phase 2 (Telegram Provider)  -- depends on message.Message from Phase 1
  |       |
  |       v
  +---> Phase 3 (Daemon)            -- depends on Phase 1 (store, config, message) + Phase 2 (telegram.Poll)
  |       |
  |       v
  +---> Phase 4 (CLI Commands)      -- depends on Phase 1 (store, config, message) + Phase 2 (telegram.Send)
  |                                     Phase 3 not required (CLI reads filesystem directly)
  v
Phase 5 (Packaging & Release)       -- depends on Phases 1-4 (functional binary)
```

Phases 2, 3, and 4 all depend on Phase 1. Phase 3 depends on Phase 2. Phase 4 depends on Phase 2 (for `send` only) but not Phase 3. Phase 5 depends on all prior phases.

## Key Architectural Decisions

1. **No IPC between CLI and daemon.** Daemon writes files; CLI reads files. `comms send` calls Telegram API directly.
2. **Per-channel cursor files.** Each channel directory has a `.cursor` file with the timestamp of the last-read message.
3. **All chats, CLI filtering.** Daemon writes messages from all chats the bot is in. Agent filters at read time via `--channel`.
4. **Minimal send stub.** stdin-only input, no `--body` flag. Revisit UX after real agent usage.
5. **Token from env or config.** `$COMMS_TELEGRAM_TOKEN` env var takes priority over `config.toml` value.

## Cross-Phase Interface Summary

| Interface | Defined In | Used By |
|-----------|-----------|---------|
| `message.Message` struct | Phase 1, Task 1.2 | Phases 2, 3, 4 |
| `message.MarshalMarkdown` / `UnmarshalMarkdown` | Phase 1, Task 1.2 | Phases 3, 4 (via store) |
| `message.MarshalOrg` / `UnmarshalOrg` | Phase 1, Task 1.3 | Phases 3, 4 (via store) |
| `config.Config` struct + `Load()` | Phase 1, Task 1.4 | Phases 3, 4 |
| `store.WriteMessage()` / `ReadMessage()` | Phase 1, Task 1.5 | Phase 3 (daemon), Phase 4 (list/unread) |
| `store.ReadCursor()` / `WriteCursor()` | Phase 1, Task 1.6 | Phase 4 (unread command) |
| `store.ListChannels()` / `ListMessages()` / `ListMessagesAfter()` | Phase 1, Task 1.6 | Phase 4 (channels, list, unread) |
| `telegram.BotAPI` interface | Phase 2, Task 2.1 | Phase 2 (send, poll), Phase 4 (send command) |
| `telegram.Send()` | Phase 2, Task 2.3 | Phase 4 (send command) |
| `telegram.Poll()` | Phase 2, Task 2.4 | Phase 3 (daemon core loop) |
| `store.ReadChatID()` / `WriteChatID()` | Phase 3, Task 3.9 | Phase 3 (daemon), Phase 4 (send command) |
| `store.ReadOffset()` / `WriteOffset()` | Phase 3, Task 3.8 | Phase 3 (daemon) |
| `daemon.WritePID()` / `ReadPID()` / `IsRunning()` | Phase 3, Task 3.1 | Phase 3 (daemon CLI commands) |

## How To Use

Run the `reviewing-plans` skill to validate and refine these plans, then run the `plan-to-tickets` skill to convert them into dependency-managed tasks for autonomous execution.
