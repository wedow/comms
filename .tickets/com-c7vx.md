---
id: com-c7vx
status: open
deps: [com-mnvn, com-gvr2]
links: []
created: 2026-03-20T01:04:56Z
type: task
priority: 1
assignee: Greg Wedow
---
# Plugin Arch: Create protocol tests

## Summary
Test round-trip encode/decode for all types, blank line skipping, oversized line rejection.

## Files
- Create: `internal/protocol/codec_test.go`

## Implementation Notes
- Test round-trip for every protocol type defined in types.go
- Test blank line skipping behavior
- Test oversized line rejection (>1 MiB)
- Use table-driven tests where appropriate

## Verification
```bash
go test ./internal/protocol/ -v
```

## Acceptance Criteria
- All protocol types round-trip correctly
- Blank lines skipped
- Oversized lines rejected

