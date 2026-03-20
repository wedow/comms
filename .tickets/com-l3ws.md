---
id: com-l3ws
status: open
deps: [com-6bya, com-ohfe]
links: []
created: 2026-03-20T01:05:41Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Protocol I/O tests

## Summary
Round-trip encode/decode tests, blank line handling, oversized line rejection, unknown type error. Tests exercise read/write calls in provider.go directly.

## Files
- Modify: `internal/daemon/provider_test.go`

## Implementation Notes
- Add tests to existing provider_test.go file
- Test scenarios:
  - Round-trip encode/decode of protocol messages through subprocess pipes
  - Blank line handling (skipped correctly)
  - Oversized line rejection (1 MiB limit)
  - Unknown type error handling
- Tests exercise the protocol I/O as integrated in provider.go
- Use os.Pipe() pairs to simulate subprocess communication

## Verification
```bash
go test ./internal/daemon/ -run TestProtocol -v
```

## Acceptance Criteria
- Protocol I/O round-trips correctly. Edge cases handled.

