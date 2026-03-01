---
id: com-0ljk
status: closed
deps: [com-qtjh, com-6qs4]
links: []
created: 2026-03-01T13:18:28Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-3]
---
# comms-v1: Telegram polling offset persistence

## Summary
Persist the Telegram polling offset across daemon restarts so the daemon does not re-process old messages. Store the offset in `.comms/telegram.offset` as a plain integer. The daemon reads it on startup and passes it to `telegram.Poll`. On shutdown, the daemon writes the returned offset back.

## Files
- `internal/store/offset.go` -- `ReadOffset(root, provider string) (int64, error)`, `WriteOffset(root, provider string, offset int64) error`. File path: `root/<provider>.offset`.
- `internal/store/offset_test.go` -- test read/write cycle, test missing file returns 0
- `internal/daemon/daemon.go` -- integrate offset read on startup, write on shutdown (modify existing file from Task 3.2)

## Implementation Notes
- Dependencies: Task 3.2 (daemon core loop), Task 1.5 (store)
- Function signatures:
  - `ReadOffset(root, provider string) (int64, error)` -- reads offset from `root/<provider>.offset`; returns 0 if file does not exist
  - `WriteOffset(root, provider string, offset int64) error` -- writes offset as plain integer to `root/<provider>.offset`
- File path convention: `<root>/<provider>.offset` (e.g., `.comms/telegram.offset`)
- Offset is stored as a plain integer (text file, single number)
- Daemon integration:
  - On startup: call `ReadOffset` and pass to `telegram.Poll`
  - On shutdown: call `WriteOffset` with the final offset value

## Verification
```sh
cd /home/greg/p/comms && go test ./internal/store/... -run TestOffset -v
cd /home/greg/p/comms && go test ./internal/daemon/... -v
```

## Acceptance Criteria
- Offset round-trips correctly
- Missing file returns 0
- Daemon integrates offset on startup/shutdown

