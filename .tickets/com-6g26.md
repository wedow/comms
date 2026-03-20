---
id: com-6g26
status: open
deps: [com-7pm4, com-w70v]
links: []
created: 2026-03-20T01:06:43Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Crash recovery in daemon core

## Summary
Per-provider goroutine calls `RespawnManager.Run(ctx)` (from the crash detection and respawn task). All backoff and failure-counting logic lives in RespawnManager -- goroutine does not reimplement it. Goroutine responsible only for: logging when RespawnManager.Run returns (provider permanently failed or context cancelled), persisting final offset to disk before exiting.

## Files
- Modify: `internal/daemon/daemon.go`

## Implementation Notes
- Each per-provider goroutine wraps its subprocess lifecycle in RespawnManager.Run(ctx)
- RespawnManager handles all backoff and failure counting internally
- Goroutine only handles: logging RespawnManager exit reason, persisting final offset via store.WriteOffset()
- Do NOT reimplement backoff logic in the goroutine

## Verification
```bash
go test ./internal/daemon/ -run TestRun -v
```

## Acceptance Criteria
- Daemon uses RespawnManager for crash recovery
- Final offset persisted on exit

