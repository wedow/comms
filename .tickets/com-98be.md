---
id: com-98be
status: open
deps: [com-xgku, com-gjjv]
links: []
created: 2026-03-20T01:05:31Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Rewrite send_test.go for delegation

## Summary
Replace telegram-mock tests with delegation tests: verify correct args are passed to provider binary, provider resolution works, error propagation is correct. The detailed telegram API tests have been moved to the provider package.

## Files
- Modify: `internal/cli/send_test.go`

## Implementation Notes
- Replace telegram-mock tests with delegation tests
- Verify correct args are passed to provider binary
- Test provider resolution via swappable vars
- Test error propagation from delegate()
- No telegram SDK references should remain

## Verification
```bash
go test ./internal/cli/ -run TestSend -v
```

## Acceptance Criteria
- send_test.go tests delegation behavior, not telegram API calls

