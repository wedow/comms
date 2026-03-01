---
id: com-hogs
status: closed
deps: [com-9par, com-yvif]
links: []
created: 2026-03-01T13:18:25Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-4]
---
# comms-v1: comms unread command

## Summary
Implement `comms unread` CLI command. Lists unread messages across all channels (or filtered by `--channel`), then advances each channel's cursor to the newest message's timestamp. This is the primary command agents use to poll for new messages.

Flags:
- `--channel <name>` -- optional, filter to a single channel

`--channel` takes the full channel directory name as shown by `comms channels` (e.g., `telegram-general`).

Flow:
1. Call `store.ListChannels()` (or filter to `--channel`)
2. For each channel:
   a. Call `store.ReadCursor(channel)` to get the last-read timestamp
   b. Call `store.ListMessagesAfter(channel, cursor)` to get files newer than cursor
   c. For each file, `store.ReadMessage()` and print as JSON line
   d. If any messages were found, call `store.WriteCursor(channel, newestTimestamp)` to advance cursor
3. Exit 0 even if no unread messages (empty output is valid)

Output format: same JSON line format as `comms list`.

Cursor advancement happens per-channel, after all messages for that channel have been printed.

## Files
- `internal/cli/unread.go` -- command definition
- `internal/cli/unread_test.go` -- tests

## Implementation Notes
- Dependencies: Task 4.1 (JSON output helpers), Task 1.5 (store read/write), Task 1.6 (store listing/cursor)
- The command constructor should follow the pattern `cli.NewUnreadCmd() *cobra.Command`
- Use `cli.PrintJSON()` from Task 4.1 to output each message
- Use `cli.PrintError()` for error cases
- Test cases to implement:
  - No cursor file exists: all messages returned, cursor created
  - Cursor exists: only newer messages returned, cursor advanced
  - No new messages: no output, cursor unchanged
  - `--channel` filters to one channel only
  - Running `unread` twice in succession: second run produces no output
  - Multiple channels: each cursor advances independently

## Verification
```sh
cd /home/greg/p/comms && go test ./internal/cli/ -run TestUnread -v
```

## Acceptance Criteria
- All test cases pass
- Cursors advance correctly
- Second run produces no output

