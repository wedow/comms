---
id: com-pqo8
status: open
deps: [com-aoxk, com-qcy9]
links: []
created: 2026-03-20T01:07:52Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Download media in handler tests

## Summary
Verify `downloadMedia` works when called from new handler path. Update existing `TestRunDownloadsMedia` test for new subprocess-based infrastructure.

## Files
- Modify: `internal/daemon/daemon_test.go`

## Implementation Notes
- Update TestRunDownloadsMedia to use mock subprocess infrastructure (not old fakeProvider)
- Test should write a MessageEvent with DownloadURL to the mock subprocess pipe
- Verify that media is downloaded and stored correctly
- Depends on both the media download handler wiring and the rewritten test infrastructure

## Verification
```bash
go test ./internal/daemon/ -run TestRunDownloadsMedia -v
```

## Acceptance Criteria
- Media download test passes with new subprocess infrastructure

