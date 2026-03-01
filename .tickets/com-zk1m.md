---
id: com-zk1m
status: closed
deps: [com-hyrn, com-2bzb]
links: []
created: 2026-03-01T20:37:08Z
type: feature
priority: 1
assignee: Greg Wedow
parent: com-d35c
tags: [telegram, v2, replies]
---
# Send replies via comms send

Add `--reply-to` flag to `comms send`. When set, include `ReplyParameters` in `bot.SendMessageParams`.

- Add `ParseMessageID(id string) (int, error)` helper in `internal/provider/telegram/` — strips `telegram-` prefix, parses int
- In `internal/cli/send.go`: read `--reply-to` flag, call ParseMessageID, pass `ReplyParameters{MessageID: id}` to telegram.Send
- Update `telegram.Send()` signature to accept optional reply-to ID (or pass SendMessageParams directly)
- BotAPI interface unchanged — `SendMessageParams` already has `ReplyParameters` field

**SDK type:** `bot.SendMessageParams.ReplyParameters` is `*models.ReplyParameters{MessageID: int}`

**Example:** `echo 'thanks' | comms send --channel telegram-dm-123 --reply-to telegram-456`

**Files to modify:** `internal/provider/telegram/send.go`, `internal/cli/send.go`, new `internal/provider/telegram/msgid.go` (or inline), and test files.

## TDD

Write failing tests first. Test ParseMessageID with valid/invalid inputs. Test that Send with reply-to sets ReplyParameters on the API call (mock BotAPI).

## Verification

```
go test ./internal/provider/telegram/... -v
go test ./internal/cli/... -v
go vet ./...
```

## Acceptance Criteria

`comms send --reply-to telegram-{id}` sends a Telegram reply. The reply is visible in Telegram UI as a threaded reply. Tests mock the API call and verify ReplyParameters is set.

