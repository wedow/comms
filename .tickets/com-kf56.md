---
id: com-kf56
status: open
deps: [com-plln, com-p48x]
links: []
created: 2026-03-20T01:05:21Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Refactor react.go to delegate

## Summary
Same delegation pattern as send.go. Replace current react implementation with delegation logic: extract provider from channel, resolve `comms-<provider>` binary, forward flags as argv. Remove imports: `go-telegram/bot`, `go-telegram/bot/models`, `providers/telegram`, `store`.

## Files
- Modify: `internal/cli/react.go`

## Implementation Notes
- Same delegation pattern as send.go refactor
- Extract provider from channel, resolve comms-<provider> binary, forward flags as argv
- Remove imports: go-telegram/bot, go-telegram/bot/models, providers/telegram, store

## Verification
```bash
go test ./internal/cli/ -v
grep -r "go-telegram/bot" internal/cli/react.go  # should produce no output
```

## Acceptance Criteria
- react.go delegates to provider binary
- Zero telegram SDK imports in react.go

