---
id: com-e1kk
status: closed
deps: [com-r9tt, com-tti8]
links: []
created: 2026-03-20T01:05:15Z
type: task
priority: 1
assignee: Greg Wedow
---
# Plugin Arch: Add TelegramToken backward-compat helper

## Summary
Add `TelegramToken() string` method so existing callers (cli/daemon.go) continue working until Phase 03.

## Files
- Modify: `internal/config/config.go`

## Implementation Notes
- Method on Config struct
- Returns token from Providers["telegram"]["token"]
- Existing callers (cli/daemon.go) use this during transition period
- Will be removed in Phase 05 (task 05-1)

## Verification
```bash
go test ./internal/config/ -v
```

## Acceptance Criteria
- `TelegramToken()` returns token from Providers map
- Existing callers unaffected

