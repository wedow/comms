---
id: com-un3y
status: closed
deps: [com-plln]
links: []
created: 2026-03-20T01:04:52Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Create delegation tests

## Summary
Test `extractProvider` with various channel names. Test `resolveProviderBinary` and `delegate` via swappable var overrides.

## Files
- Create: `internal/cli/delegate_test.go`

## Implementation Notes
- Test extractProvider with various channel name formats and edge cases
- Test resolveProviderBinary and delegate via swappable var overrides
- Follow existing test patterns in the cli package

## Verification
```bash
go test ./internal/cli/ -v
```

## Acceptance Criteria
- All delegation functions tested
- Channel name parsing covers edge cases

