---
id: com-36ho
status: open
deps: []
links: []
created: 2026-03-20T01:05:23Z
type: task
priority: 1
assignee: Greg Wedow
---
# Plugin Arch: Verify store offset functions are provider-parameterized

## Summary
Confirm `ReadOffset(root, provider)` and `WriteOffset(root, provider, offset)` already take a provider parameter. No code changes -- just verification. Note: `daemon.go` hardcodes `"telegram"` as the provider argument; this is fixed in Phase 03.

## Files
- None (verification only)

## Implementation Notes
- This is a verification-only task, no code changes needed
- Check function signatures in `internal/store/` for ReadOffset and WriteOffset
- Confirm they accept a provider parameter
- Note that daemon.go currently hardcodes "telegram" -- this will be fixed in Phase 03

## Verification
```bash
go test ./internal/store/ -v
```

## Acceptance Criteria
- Store offset functions confirmed to accept provider parameter
- No code changes needed

