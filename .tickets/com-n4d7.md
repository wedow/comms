---
id: com-n4d7
status: open
deps: [com-hyrn]
links: []
created: 2026-03-01T20:37:16Z
type: feature
priority: 1
assignee: Greg Wedow
parent: com-d35c
tags: [telegram, v2, reactions]
---
# Receive reactions from Telegram

**Design decision:** Reactions are appended to the existing message file (like edits), not stored as separate files. The append should include a clear timestamp and metadata section so agents can see ordering of events on a message.

Implementation:
- Subscribe to reaction updates: pass `bot.WithAllowedUpdates(bot.AllowedUpdates{...})` including `"message_reaction"` when creating the bot in `poll.go`
- Handle `update.MessageReaction` in poll.go's default handler (currently only `update.Message` is processed)
- On reaction update: find the existing message file by matching message ID, append a reaction section with timestamp, who reacted, and emoji
- Move the cursor back so the message shows as unread again
- If the original message file isn't found (reacted to old message), store reaction info in a new file

**SDK types:** `models.MessageReactionUpdated` has: `Chat`, `MessageID`, `User` (*models.User), `ActorChat` (*models.Chat), `Date`, `OldReaction` ([]models.ReactionType), `NewReaction` ([]models.ReactionType). `models.ReactionType` is interface — check for `models.ReactionTypeEmoji{Emoji string}`.

**Append format example:**
```
---reaction---
date: 2026-03-01T20:30:00Z
from: Greg
emoji: 👍
```

**Files to modify:** `internal/provider/telegram/poll.go`, `internal/store/` (append helper + cursor reset), and test files.

Note: bot must be admin in group chats to receive reaction updates. In DMs this works automatically.

## TDD

Write failing tests first. Test reaction update conversion. Test file append with demarcation. Test cursor reset after append.

## Verification

```
go test ./internal/provider/telegram/... -v
go test ./internal/store/... -v
go vet ./...
```

## Acceptance Criteria

When a user reacts to a message, the reaction is appended to the existing message file with timestamp/from/emoji. The message shows as unread again. Tests mock reaction updates and verify append + cursor behavior.

