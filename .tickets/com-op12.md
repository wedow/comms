---
id: com-op12
status: open
deps: [com-r9tt]
links: []
created: 2026-03-20T01:05:19Z
type: task
priority: 1
assignee: Greg Wedow
---
# Plugin Arch: Update init command for new config format

## Summary
Update `Default()` to produce config with `[providers.telegram]` section instead of `[telegram]`.

## Files
- Modify: `internal/config/config.go` (Default function)

## Implementation Notes
- Default() should populate the Providers map with a "telegram" entry
- The generated TOML should use `[providers.telegram]` section header
- Init command tests must still pass

## Verification
```bash
go test ./internal/cli/ -run TestInit -v
```

## Acceptance Criteria
- `Default()` produces `[providers.telegram]` format
- Init command tests pass

