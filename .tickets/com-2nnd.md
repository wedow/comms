---
id: com-2nnd
status: open
deps: [com-ho3r]
links: []
created: 2026-03-20T01:05:20Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Subprocess SendCommand

## Summary
Implement `Subprocess.SendCommand(ctx, cmd) error`. Acquire write lock on stdin pipe, encode command, handle context cancellation. Only works when subprocess is in `Running` state.

## Files
- Modify: `internal/daemon/provider.go`

## Implementation Notes
- Method on Subprocess struct
- Acquire write mutex/lock on stdin pipe before writing
- Encode command as JSONL using protocol codec
- Handle context cancellation (return error if ctx done)
- Return error if subprocess is not in Running state
- Thread-safe: multiple goroutines may call SendCommand concurrently

## Verification
```bash
go test ./internal/daemon/ -run TestSubprocess -v
```

## Acceptance Criteria
- SendCommand writes to stdin pipe with proper locking. Errors on non-running state.

