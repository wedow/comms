---
id: com-ogzs
status: open
deps: [com-kf56, com-gjjv]
links: []
created: 2026-03-20T01:05:33Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Rewrite react_test.go for delegation

## Summary
Same pattern as send_test.go rewrite. Replace telegram-mock tests with delegation tests: verify correct args are passed to provider binary, provider resolution works, error propagation is correct.

## Files
- Modify: `internal/cli/react_test.go`

## Implementation Notes
- Same delegation test pattern as send_test.go rewrite
- Replace telegram-mock tests with delegation tests
- Verify correct args are passed to provider binary
- Test provider resolution via swappable vars
- Test error propagation from delegate()
- No telegram SDK references should remain

## Verification
```bash
go test ./internal/cli/ -run TestReact -v
```

## Acceptance Criteria
- react_test.go tests delegation behavior, not telegram API calls

