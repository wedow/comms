---
id: com-g5fh
status: closed
deps: []
links: []
created: 2026-03-20T01:05:01Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Extract per-provider event handlers

## Summary
Extract inline handler logic from `daemon.Run()` into standalone functions: `handleMessageEvent(root, cfg, msg, chatID, cb)`, `handleEditEvent(root, cfg, msg, chatID, cb)`, `handleReactionEvent(root, cfg, evt, provider, cb)`. Pure refactor -- existing tests should still pass.

## Files
- Create: `internal/daemon/handler.go`

## Implementation Notes
- Pure refactor of existing code in daemon.go
- Extract message/edit/reaction handling into standalone functions
- No new behavior, just code reorganization
- All existing daemon tests must pass unchanged

## Verification
```bash
go test ./internal/daemon/ -v
```

## Acceptance Criteria
- Handler functions extracted. Existing daemon tests still pass unchanged.

