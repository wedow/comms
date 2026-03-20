---
id: com-aoxk
status: closed
deps: [com-g5fh]
links: []
created: 2026-03-20T01:06:54Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Media download from protocol events

## Summary
Move `downloadMedia` into handler path. `handleMessageEvent` calls `downloadMedia` when `msg.DownloadURL != ""`. No change to download logic itself.

## Files
- Modify: `internal/daemon/daemon.go` or `internal/daemon/media.go`

## Implementation Notes
- The downloadMedia function is called from handleMessageEvent when DownloadURL is present in the protocol event
- No changes to the actual download logic -- just wiring it into the new handler path
- The DownloadURL field comes from the protocol MessageEvent after conversion

## Verification
```bash
go test ./internal/daemon/ -v
```

## Acceptance Criteria
- Media download triggered from handler when DownloadURL present

