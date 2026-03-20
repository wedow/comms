---
id: com-lnh0
status: open
deps: [com-ho3r]
links: []
created: 2026-03-20T01:05:13Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Subprocess shutdown

## Summary
Implement `Subprocess.Shutdown(reason string) error`. Write `shutdown` command, wait 5s for `shutdown_complete`, then `cmd.Process.Kill()`.

## Files
- Modify: `internal/daemon/provider.go`

## Implementation Notes
- Method on Subprocess struct
- Write shutdown command to stdin pipe
- Wait up to 5s for shutdown_complete event
- Force kill process if timeout exceeded
- Handle already-dead process gracefully

## Verification
```bash
go test ./internal/daemon/ -run TestSubprocess -v
```

## Acceptance Criteria
- Shutdown sends command, waits for completion, force-kills on timeout.

