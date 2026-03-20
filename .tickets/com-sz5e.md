---
id: com-sz5e
status: open
deps: [com-xgku, com-kf56]
links: []
created: 2026-03-20T01:05:24Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Update root.go

## Summary
Remove `telegram.NewBot` parameter from `newSendCmd()` and `newReactCmd()` calls. Remove telegram import from root.go.

## Files
- Modify: `internal/cli/root.go`

## Implementation Notes
- Remove telegram.NewBot parameter from newSendCmd() and newReactCmd() calls
- Remove telegram import entirely from root.go
- These calls must match the new signatures from the send.go and react.go refactors

## Verification
```bash
go build ./cmd/comms
grep -r "go-telegram/bot" internal/cli/root.go  # should produce no output
```

## Acceptance Criteria
- root.go has zero telegram imports
- newSendCmd/newReactCmd calls updated

