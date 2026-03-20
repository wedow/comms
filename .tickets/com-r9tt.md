---
id: com-r9tt
status: closed
deps: []
links: []
created: 2026-03-20T01:05:01Z
type: task
priority: 1
assignee: Greg Wedow
---
# Plugin Arch: Add Providers map to Config struct

## Summary
Change `Config` to support `[providers.<name>]` sections while keeping backward compatibility with `[telegram]`. Add `Providers map[string]map[string]any` field with `toml:"providers"` tag. Keep deprecated `Telegram TelegramConfig` field.

## Files
- Modify: `internal/config/config.go`

## Implementation Notes
- Add `Providers map[string]map[string]any` with `toml:"providers"` tag
- Keep deprecated `Telegram TelegramConfig` field for backward compat
- Existing tests must still pass

## Verification
```bash
go test ./internal/config/ -v
```

## Acceptance Criteria
- Config struct has Providers map field
- Existing tests still pass

