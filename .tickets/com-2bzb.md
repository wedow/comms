---
id: com-2bzb
status: closed
deps: [com-hyrn]
links: []
created: 2026-03-01T20:37:03Z
type: feature
priority: 1
assignee: Greg Wedow
parent: com-d35c
tags: [telegram, v2, replies]
---
# Receive reply metadata from Telegram

Update `convertMessage()` in `internal/provider/telegram/convert.go` to extract reply fields:

- `m.ReplyToMessage.ID` → `Message.ReplyTo` (formatted as `telegram-{id}`)
- `m.ReplyToMessage.Text` → `Message.ReplyToBody`
- `m.Quote.Text` → `Message.Quote` (from `models.TextQuote`)
- For cross-chat replies: `m.ExternalReply` has `MessageID` and `Origin` but no full message — set ReplyTo from ExternalReply.MessageID if present

**SDK types:** `models.Message.ReplyToMessage` (*models.Message), `models.Message.Quote` (*models.TextQuote), `models.Message.ExternalReply` (*models.ExternalReplyInfo)

**Files to modify:** `internal/provider/telegram/convert.go`, `internal/provider/telegram/convert_test.go`

## TDD

Write failing tests first. Mock `models.Message` with ReplyToMessage set, verify converted Message has correct ReplyTo/ReplyToBody. Mock with Quote set, verify Quote field. Mock with ExternalReply, verify ReplyTo.

## Verification

```
go test ./internal/provider/telegram/... -v
go vet ./...
```

## Acceptance Criteria

When a Telegram user replies to a message, comms unread output includes reply_to field with the original message ID. When a partial quote is used, the quote field is populated. Tests cover same-chat reply, quoted reply, and external reply.

