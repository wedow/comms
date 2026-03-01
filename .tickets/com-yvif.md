---
id: com-yvif
status: open
deps: [com-6qs4]
links: []
created: 2026-03-01T13:18:01Z
type: task
priority: 1
assignee: Greg Wedow
tags: [phase-1]
---
# comms-v1: Store cursor management and channel/message listing

## Summary
Add cursor files and directory enumeration to the store package.

Functions:
- `ReadCursor(root, channel string) (time.Time, error)` -- read `root/<channel>/.cursor`, parse as RFC3339Nano. Return zero time if file doesn't exist (meaning: all messages are unread).
- `WriteCursor(root, channel string, t time.Time) error` -- write RFC3339Nano timestamp to `root/<channel>/.cursor`
- `ListChannels(root string) ([]string, error)` -- return sorted directory names under `root/` that look like channel dirs (exclude `docs/`, files)
- `ListMessages(root, channel string) ([]string, error)` -- return sorted file paths in `root/<channel>/` (exclude `.cursor`), oldest first
- `ListMessagesAfter(root, channel string, after time.Time) ([]string, error)` -- like ListMessages but only files with timestamp > `after`

Sorting: by filename (which is a timestamp), lexicographic ascending = chronological ascending.

## Files
- `internal/store/cursor.go` -- ReadCursor, WriteCursor
- `internal/store/list.go` -- ListChannels, ListMessages, ListMessagesAfter
- `internal/store/cursor_test.go` -- cursor round-trip, missing file returns zero time
- `internal/store/list_test.go` -- listing with multiple channels and messages, filtering by time

## Implementation Notes
- Depends on Task 1.5 (store init/write/read must exist)
- Cursor file location: `root/<channel>/.cursor`
- Cursor format: RFC3339Nano timestamp
- Missing cursor file returns zero time (not an error)
- ListChannels must exclude `docs/` and plain files
- ListMessages must exclude `.cursor` files
- ListMessagesAfter filters by comparing filename timestamps
- All listings sorted lexicographically ascending (= chronological order since filenames are timestamps)

## Verification
```sh
go test ./internal/store/... -run 'TestCursor|TestList' -v
```

## Acceptance Criteria
- All tests pass
- Cursor round-trips correctly
- Missing cursor returns zero time
- ListChannels excludes `docs/`
- ListMessagesAfter filters correctly

