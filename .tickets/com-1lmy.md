---
id: com-1lmy
status: closed
deps: [com-r9tt]
links: []
created: 2026-03-20T01:04:51Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Config parsing for provider names

## Summary
Add `ProviderNames()` method to config that returns sorted keys of the `Providers` map. The daemon uses this to determine which providers to spawn.

## Files
- Modify: `internal/config/config.go`

## Implementation Notes
- Method on Config struct
- Return sorted slice of strings (keys from Providers map)
- Used by daemon to know which provider subprocesses to spawn

## Verification
```bash
go test ./internal/config/ -v
```

## Acceptance Criteria
- ProviderNames() returns sorted provider names from config.

