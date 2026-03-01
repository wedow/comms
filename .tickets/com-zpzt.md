---
id: com-zpzt
status: open
deps: [com-hyrn]
links: []
created: 2026-03-01T20:37:39Z
type: feature
priority: 2
assignee: Greg Wedow
parent: com-d35c
tags: [telegram, v2, edits]
---
# Receive edited messages from Telegram

**Design decision:** Edits append to the existing message file with clear demarcation (like reactions). The cursor resets so the message shows as unread again. Edits are not separate files.

Implementation:
- Handle `update.EditedMessage` in poll.go (currently ignored)
- Also handle `update.EditedChannelPost` (depends on com-1p1c for channel post support)
- On edit: find the existing message file by matching message ID (`telegram-{id}`)
- Append an edit section with timestamp and new body:
  ```
  ---edit---
  date: 2026-03-01T20:35:00Z
  New body text here after the edit.
  ```
- Set `EditDate` in the message frontmatter to the latest edit timestamp
- Move the cursor back so the message shows as unread again

**SDK fields:** `update.EditedMessage` is `*models.Message` with `EditDate` (int, Unix timestamp) set. Same struct as regular messages.

**Files to modify:** `internal/provider/telegram/poll.go`, `internal/store/` (append helper + cursor reset), and test files.

## TDD

Write failing tests first. Test edit append format. Test cursor reset after edit. Test that EditDate is set in frontmatter.

## Verification

```
go test ./internal/provider/telegram/... -v
go test ./internal/store/... -v
go vet ./...
```

## Acceptance Criteria

When a user edits a message, the new text is appended to the existing file with `---edit---` demarcation and timestamp. The message appears as unread again. Tests cover edit detection, append, and cursor reset.

