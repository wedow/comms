---
id: com-l7o6
status: open
deps: [com-6g4f, com-fe1f, com-6g26, com-0cej, com-qcy9, com-pqo8, com-sz5e, com-98be, com-ogzs]
links: []
created: 2026-03-20T01:09:48Z
type: task
priority: 3
assignee: Greg Wedow
---
# Plugin Arch: Remove deprecated config fields

## Summary
Remove TelegramConfig struct and Telegram field from Config. Remove TelegramToken() backward-compat helper. Remove migration logic in Load() that copies [telegram] to [providers]. Verify all callers access provider config through Providers map or ProviderConfig() helper. Update Default() to only produce [providers.telegram] format.

## Files
- Modify: internal/config/config.go (remove TelegramConfig, Telegram field, TelegramToken(), migration logic)
- Modify: internal/cli/daemon.go (verify no remaining references)

## Implementation Notes
- All callers should already be using the new Providers map API by this point
- TelegramToken() was a backward-compat helper added in Phase 01 -- remove it now
- Migration logic that auto-converts [telegram] to [providers.telegram] is no longer needed
- Default() should only produce [providers.telegram] format (no legacy format)

## Verification
```bash
go test ./internal/config/ -v
go test ./internal/cli/ -v
go test ./...
```

## Acceptance Criteria
- No deprecated config types/methods remain
- All callers use new API

