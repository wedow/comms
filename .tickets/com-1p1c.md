---
id: com-1p1c
status: closed
deps: [com-hyrn]
links: []
created: 2026-03-01T20:37:43Z
type: feature
priority: 2
assignee: Greg Wedow
parent: com-d35c
tags: [telegram, v2, channels]
---
# Receive channel posts from Telegram

Handle `update.ChannelPost` in poll.go (currently ignored).

- Process `update.ChannelPost` through the existing `convertMessage()` path
- Update `convertMessage()` to handle nil `m.From` with non-nil `m.SenderChat`: use `m.SenderChat.Title` or `m.SenderChat.Username` for the From field
- Channels are already supported in `SlugifyChat` (ChatTypeChannel)
- Do NOT handle `EditedChannelPost` here — that's covered by com-zpzt (edits)

**SDK fields:** `update.ChannelPost` is `*models.Message`. `m.From` is nil for channel posts. `m.SenderChat` (*models.Chat) identifies the channel.

**Files to modify:** `internal/provider/telegram/poll.go`, `internal/provider/telegram/convert.go` (SenderChat fallback), and test files.

## TDD

Write failing tests first. Test convertMessage with nil From and SenderChat set. Test that poll.go processes ChannelPost updates.

## Verification

```
go test ./internal/provider/telegram/... -v
go vet ./...
```

## Acceptance Criteria

Messages posted to a Telegram channel where the bot is admin appear as files in .comms/ and show in comms unread. From field shows channel name. Tests cover channel post conversion.

