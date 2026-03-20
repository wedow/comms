---
id: com-0cej
status: open
deps: [com-w70v]
links: []
created: 2026-03-20T01:07:07Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Signal handling and graceful shutdown

## Summary
Depends on daemon core loop existing. Adds shutdown sequence: (1) Wrap context with `signal.NotifyContext(ctx, syscall.SIGTERM, os.Interrupt)` at entry. (2) After main select falls through (context cancelled): call `Shutdown()` on all live subprocesses, 5s timeout, then `cmd.Process.Kill()` on stragglers. (3) Wait for all per-provider goroutines to exit (use `sync.WaitGroup`). (4) Persist final offsets via `store.WriteOffset()`. (5) Remove PID file. Add test: cancel context (or send SIGTERM), verify all mock subprocesses receive shutdown command, verify offsets persisted, verify PID file removed.

## Files
- Modify: `internal/daemon/daemon.go`
- Modify: `internal/daemon/daemon_test.go`

## Implementation Notes
- Use signal.NotifyContext for SIGTERM and os.Interrupt
- 5-second timeout for graceful subprocess shutdown before force kill
- sync.WaitGroup to track per-provider goroutines
- Final offset persistence and PID file cleanup happen after all goroutines exit
- Tests should cancel context and verify shutdown sequence

## Verification
```bash
go test ./internal/daemon/ -run TestRunGracefulShutdown -v
go test ./internal/daemon/ -run TestRunForceKill -v
```

## Acceptance Criteria
- Signal handling triggers graceful shutdown
- Subprocesses receive shutdown command
- Offsets persisted
- PID file removed

