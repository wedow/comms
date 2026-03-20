---
id: com-7pm4
status: open
deps: [com-ho3r]
links: []
created: 2026-03-20T01:05:25Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Crash detection and respawn

## Summary
Implement `RespawnManager` wrapping `Spawn` with crash recovery. Exponential backoff (1s, 2s, 4s, 8s, 30s cap), resets after 60s stable. 5 consecutive failures = fatal.

## Files
- Modify: `internal/daemon/provider.go`

## Implementation Notes
- RespawnManager struct wraps Spawn function
- Exponential backoff: 1s, 2s, 4s, 8s, capped at 30s
- Backoff resets to 1s after subprocess runs stable for 60s
- 5 consecutive failures triggers fatal error (returns permanent error)
- Each respawn reads current offset from store before re-spawning
- Context cancellation stops respawn loop

## Verification
```bash
go test ./internal/daemon/ -run TestRespawn -v
```

## Acceptance Criteria
- RespawnManager retries with backoff. Resets after stable period. Fatal after 5 consecutive failures.

