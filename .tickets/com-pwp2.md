---
id: com-pwp2
status: closed
deps: [com-6qs4]
links: []
created: 2026-03-01T13:18:35Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-3]
---
# comms-v1: Chat ID persistence

## Summary
When the daemon receives a message from a new chat, it needs to persist the chat ID so that `comms send` (Phase 4, Task 4.5) can resolve channel names to Telegram chat IDs. Store the chat ID in `<channel>/.chat_id` as a plain integer.

The daemon's poll handler already receives `chatID int64` from `telegram.Poll` (see Phase 2, Task 2.4). After writing the message file, the daemon calls `store.WriteChatID` to persist the mapping.

## Files
- `internal/store/chatid.go` -- `ReadChatID(root, channel string) (int64, error)`, `WriteChatID(root, channel string, chatID int64) error`. File path: `root/<channel>/.chat_id`.
- `internal/store/chatid_test.go` -- test read/write cycle, test missing file returns error

## Implementation Notes
- Dependencies: Task 1.5 (store)
- Function signatures:
  - `ReadChatID(root, channel string) (int64, error)` -- reads chat ID from `root/<channel>/.chat_id`; returns error if file does not exist
  - `WriteChatID(root, channel string, chatID int64) error` -- writes chat ID as plain integer to `root/<channel>/.chat_id`
- File path convention: `<root>/<channel>/.chat_id` (e.g., `.comms/telegram-general/.chat_id`)
- Chat ID is stored as a plain integer (text file, single number)
- Unlike offset (which returns 0 for missing file), missing chat ID file returns a clear error

## Verification
```sh
cd /home/greg/p/comms && go test ./internal/store/... -run TestChatID -v
```

## Acceptance Criteria
- Chat ID round-trips correctly
- Missing file returns a clear error

