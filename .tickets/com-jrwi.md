---
id: com-jrwi
status: closed
deps: [com-68x1, com-ab0w, com-oqd8]
links: []
created: 2026-03-20T01:09:16Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Integration test for subprocess mode

## Summary
Build and run comms-telegram as real subprocess. No mock injection.

Test steps:
1. Build binary to temp path (go build -o $tmpdir/comms-telegram ./cmd/comms-telegram)
2. Spawn with COMMS_PROVIDER_CONFIG={"token":"fake-token"}
3. Ready event arrives before API calls (always observable regardless of token validity)
4. Send start, then send/react commands -- return errors from Telegram (expected with fake token), testable by reading response events with ok:false
5. Send shutdown, verify shutdown_complete

## Files
- Create: providers/telegram/subprocess_integration_test.go (or added to subprocess_test.go with build tags)

## Implementation Notes
- Builds real binary, no mocks
- Uses fake token so Telegram API calls will fail (expected)
- Verifies full protocol lifecycle: handshake -> start -> commands -> shutdown
- Ready event should arrive before any API calls are made
- Response events for send/react will have ok:false with error messages

## Verification
```bash
go test ./providers/telegram/ -run TestIntegration -v -count=1
```

## Acceptance Criteria
- Real binary subprocess completes full protocol lifecycle with fake token

