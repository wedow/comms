---
id: com-bng0
status: closed
deps: [com-hyrn]
links: []
created: 2026-03-01T20:37:53Z
type: feature
priority: 2
assignee: Greg Wedow
parent: com-d35c
tags: [telegram, v2, forwards]
---
# Extract forward metadata from Telegram messages

Extract forwarded message metadata in `convertMessage()`:

- `m.ForwardOrigin` is `models.MessageOrigin` (interface). Check concrete type:
  - `*models.MessageOriginUser` → `ForwardFrom = User.FirstName` (or Username), `ForwardDate = Date`
  - `*models.MessageOriginHiddenUser` → `ForwardFrom = SenderUserName`, `ForwardDate = Date`
  - `*models.MessageOriginChat` → `ForwardFrom = SenderChat.Title`, `ForwardDate = Date`
  - `*models.MessageOriginChannel` → `ForwardFrom = Chat.Title`, `ForwardDate = Date`
- Set `Message.ForwardFrom` and `Message.ForwardDate` fields

**Files to modify:** `internal/provider/telegram/convert.go`, `internal/provider/telegram/convert_test.go`

## TDD

Write failing tests first. Test convertMessage with each of the four MessageOrigin types mocked. Verify ForwardFrom and ForwardDate are set correctly for each.

## Verification

```
go test ./internal/provider/telegram/... -v
go vet ./...
```

## Acceptance Criteria

Forwarded messages include forward_from and forward_date in comms unread output. Tests cover all four origin types (user, hidden user, chat, channel).

