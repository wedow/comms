---
id: com-tti8
status: open
deps: [com-r9tt]
links: []
created: 2026-03-20T01:05:05Z
type: task
priority: 1
assignee: Greg Wedow
---
# Plugin Arch: Add config migration logic

## Summary
In `Load()`, if `[telegram]` exists but `[providers]` does not, auto-migrate by copying `Telegram` into `Providers["telegram"]`. Also handle `COMMS_TELEGRAM_TOKEN` env var override writing to `Providers`.

## Files
- Modify: `internal/config/config.go`

## Implementation Notes
- Migration happens in Load() after TOML parsing
- If Telegram section is populated and Providers is nil/empty, create Providers["telegram"] from Telegram fields
- COMMS_TELEGRAM_TOKEN env var should also write into Providers map
- Backward compatibility: old config files must work without changes

## Verification
```bash
go test ./internal/config/ -v
```

## Acceptance Criteria
- Old `[telegram]` config auto-migrates to `Providers["telegram"]`
- Env var override works

