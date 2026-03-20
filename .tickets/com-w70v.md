---
id: com-w70v
status: open
deps: [com-bl92]
links: []
created: 2026-03-20T01:06:21Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Daemon core loop with subprocess management

## Summary
Implement main `daemon.Run()` loop: (1) Write PID file. (2) Load allowed IDs. (3) Create `CallbackRunner` with `sendTypingCommand` closure. (4) For each provider: call `exec.LookPath("comms-" + provider)`, read offset, spawn subprocess via `RespawnManager`, start reader goroutine. (5) On each event: convert to message, call handler, persist offset. (6) Wait for context cancellation. Goroutine model: one reader per provider, main loop selects over channels. Signal handling and PID cleanup added in a later task.

## Files
- Modify: `internal/daemon/daemon.go`

## Implementation Notes
- sendTypingCommand closure writes TypingCommand to the correct provider's stdin
- One reader goroutine per provider, main loop selects over channels
- Signal handling and PID cleanup are NOT added here -- that comes in the signal handling task
- Uses RespawnManager from crash detection task for subprocess lifecycle
- Converts protocol events to messages using conversion functions from convert.go
- Calls handler functions from handler.go for each event type
- Persists offset after each event via store.WriteOffset()

## Verification
```bash
go test ./internal/daemon/ -run TestRun -v
```

## Acceptance Criteria
- Daemon spawns subprocesses, reads events, converts and handles them, persists offsets

