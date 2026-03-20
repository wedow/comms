---
id: com-op5m
status: open
deps: [com-r9tt]
links: []
created: 2026-03-20T01:05:08Z
type: task
priority: 1
assignee: Greg Wedow
---
# Plugin Arch: Add ProviderConfig helper

## Summary
Add `ProviderConfig(provider string) ([]byte, error)` that returns JSON serialization of a specific provider's config section (for passing as `COMMS_PROVIDER_CONFIG` env var).

## Files
- Modify: `internal/config/config.go`

## Implementation Notes
- Returns JSON bytes for a given provider name from the Providers map
- Error on unknown provider (provider not found in map)
- Used by base binary to pass config to provider binaries via COMMS_PROVIDER_CONFIG env var

## Verification
```bash
go test ./internal/config/ -v
```

## Acceptance Criteria
- ProviderConfig returns JSON bytes for a given provider name
- Error on unknown provider

