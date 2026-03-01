---
id: com-6qs4
status: open
deps: [com-1wlb, com-4nj4]
links: []
created: 2026-03-01T13:17:50Z
type: task
priority: 1
assignee: Greg Wedow
tags: [phase-1]
---
# comms-v1: Store directory creation and message write/read

## Summary
Implement the filesystem operations for the `.comms/` tree. This task covers init, write, and read.

Functions:
- `InitDir(root string) error` -- create `root/` and `root/docs/` directories; caller (the init command) handles writing config and docs separately
- `WriteMessage(root string, msg message.Message, format string) (string, error)` -- serialize `msg` using the given format ("markdown" or "org"), write to `root/<provider>-<channel>/<timestamp>.<ext>`, return the written path. Create the channel directory if it doesn't exist. Filename: RFC3339Nano timestamp with colons replaced by hyphens (filesystem-safe). Extension: `.md` for markdown, `.org` for org.
- `ReadMessage(path string) (message.Message, error)` -- detect format from extension (.md -> markdown, .org -> org), read file, unmarshal

## Files
- `internal/store/store.go` -- InitDir, WriteMessage, ReadMessage, helper for filename generation
- `internal/store/store_test.go` -- tests using `t.TempDir()`: init creates dirs, write creates file with correct content, read round-trips, unknown extension errors

## Implementation Notes
- Depends on Tasks 1.2 and 1.3 (message serialization must exist)
- Channel directory path pattern: `root/<provider>-<channel>/`
- Filename pattern: RFC3339Nano timestamp with colons replaced by hyphens (filesystem-safe)
- Extension: `.md` for markdown, `.org` for org
- WriteMessage must create channel directory if it doesn't exist
- ReadMessage detects format from file extension (.md -> markdown, .org -> org)
- ReadMessage must return error for unknown extensions
- Use `t.TempDir()` in tests for isolation

## Verification
```sh
go test ./internal/store/... -run 'TestInit|TestWrite|TestRead' -v
```

## Acceptance Criteria
- All tests pass
- InitDir creates expected directories
- WriteMessage creates correctly-named file with correct content
- ReadMessage round-trips through WriteMessage

