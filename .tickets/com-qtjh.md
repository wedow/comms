---
id: com-qtjh
status: closed
deps: [com-77c4, com-1wlb, com-9oxg, com-6qs4, com-66kj]
links: []
created: 2026-03-01T13:17:42Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-3]
---
# comms-v1: Daemon core loop

## Summary
Implement the main daemon loop in `internal/daemon`. The daemon loads config, initializes the Telegram provider, and polls for incoming messages. For each message received, it serializes via the message package and writes to disk via the store package. The daemon writes a PID file on startup and removes it on shutdown. Graceful shutdown on SIGTERM/SIGINT via `signal.NotifyContext`.

The poll handler receives both `message.Message` and `chatID int64` from `telegram.Poll`. For each message:
1. Write the message to disk via `store.WriteMessage`.
2. Write the chat ID via `store.WriteChatID` (Task 3.9) so `comms send` can resolve channel names later.
3. Fire the callback if configured (Task 3.6).

The daemon constructs the channel directory name as `msg.Provider + "-" + msg.Channel` (e.g., `telegram-general`) and passes that to store functions.

## Files
- `internal/daemon/daemon.go` -- `Run(ctx context.Context, cfg config.Config, root string) error`: write PID file, init provider, enter poll loop, remove PID file on exit
- `internal/daemon/daemon_test.go` -- test with a fake provider: verify messages are written to store, PID file created/removed

## Implementation Notes
- Dependencies: Task 3.1 (PID), Task 1.2 (message), Task 1.4 (config), Task 1.5 (store), Task 2.4 (telegram poll)
- Function signature: `Run(ctx context.Context, cfg config.Config, root string) error`
- Flow: write PID -> init provider -> poll loop -> on shutdown: remove PID
- Channel directory naming convention: `msg.Provider + "-" + msg.Channel`
- Use `signal.NotifyContext` for graceful shutdown
- The provider should be abstracted via interface so tests can use a fake

## Verification
```sh
cd /home/greg/p/comms && go build ./internal/daemon/...
cd /home/greg/p/comms && go test ./internal/daemon/... -run TestDaemon -v
```

## Acceptance Criteria
- Package builds without error
- Tests verify messages are written to store and PID file lifecycle is correct

