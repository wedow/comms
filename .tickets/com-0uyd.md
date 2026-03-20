---
id: com-0uyd
status: open
deps: [com-r9tt, com-tti8, com-op5m]
links: []
created: 2026-03-20T01:05:12Z
type: task
priority: 1
assignee: Greg Wedow
---
# Plugin Arch: Update config tests

## Summary
Add tests for new config format, migration, env var override, ProviderConfig serialization.

## Files
- Modify: `internal/config/config_test.go`

## Implementation Notes
- Test new `[providers.telegram]` TOML format parses correctly
- Test migration from old `[telegram]` format to new Providers map
- Test COMMS_TELEGRAM_TOKEN env var override writes to Providers
- Test ProviderConfig() returns correct JSON for known provider
- Test ProviderConfig() errors for unknown provider

## Verification
```bash
go test ./internal/config/ -v
```

## Acceptance Criteria
- Tests cover new Providers map, migration from old format, env var override, ProviderConfig helper

