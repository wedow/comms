---
id: com-qcy9
status: open
deps: [com-1lmy, com-265l, com-g5fh, com-ho3r, com-lnh0, com-2nnd, com-7pm4, com-6bya, com-ohfe, com-l3ws, com-bl92, com-w70v, com-fe1f, com-6g26, com-aoxk, com-0cej]
links: []
created: 2026-03-20T01:07:27Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Rewrite daemon tests

## Summary
All existing daemon tests use `fakeProvider` implementing old interface. Rewrite to use mock subprocesses: create temp dir with fake `comms-telegram` script (or inject `SpawnFunc` that creates `os.Pipe()` with JSONL streams), write protocol events to pipe, verify store state after `Run()` completes.

## Files
- Modify: `internal/daemon/daemon_test.go`
- Create: `internal/daemon/daemon_integration_test.go`

## Implementation Notes
- Replace all fakeProvider-based tests with mock subprocess tests
- Use os.Pipe() pairs to simulate subprocess stdin/stdout
- Write JSONL protocol events to pipe to simulate provider behavior
- Verify store state (messages written, offsets persisted, cursors updated) after Run() completes
- Integration tests can use a temp dir with a fake comms-telegram script
- No references to old Provider interface should remain in tests

## Verification
```bash
go test ./internal/daemon/ -v
go test ./internal/daemon/ -run TestIntegration -v
```

## Acceptance Criteria
- All daemon tests rewritten for subprocess model
- No references to old Provider interface

