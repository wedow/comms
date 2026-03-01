---
id: com-codv
status: closed
deps: [com-hyrn]
links: []
created: 2026-03-01T20:38:02Z
type: feature
priority: 2
assignee: Greg Wedow
parent: com-d35c
tags: [telegram, v2, entities]
---
# Support message entities (formatting, links, mentions)

**Receive:**
- Extract `m.Entities` ([]models.MessageEntity) from incoming messages in `convertMessage()`
- Convert to `[]message.Entity` (defined in com-hyrn): `Entity{Type, Offset, Length, URL}`
- Map `models.MessageEntity.Type` (string) directly — types include: bold, italic, underline, strikethrough, spoiler, code, pre, text_link, mention, hashtag, bot_command, url, email, phone_number
- For `text_link` entities, include `models.MessageEntity.URL` in `Entity.URL`
- Also extract `m.CaptionEntities` for media messages

**Send:**
- Add `--format` flag to `comms send`: `plain` (default), `markdown`, `html`
- When format is set, pass `ParseMode` in `bot.SendMessageParams`: `models.ParseModeMarkdown` (MarkdownV2) or `models.ParseModeHTML`
- This enables agents to send formatted messages (bold, links, code blocks)
- Also apply to media sends (SendPhotoParams etc. have ParseMode)

**SDK types:** `models.MessageEntity{Type string, Offset int, Length int, URL string, ...}`. `models.ParseModeMarkdown` = `"MarkdownV2"`, `models.ParseModeHTML` = `"HTML"`.

**Files to modify:** `internal/provider/telegram/convert.go`, `internal/provider/telegram/send.go`, `internal/cli/send.go`, and test files.

## TDD

Write failing tests first. Test entity extraction from mocked messages with various entity types. Test that --format markdown sets ParseMode on the API call.

## Verification

```
go test ./internal/provider/telegram/... -v
go test ./internal/cli/... -v
go vet ./...
```

## Acceptance Criteria

Incoming messages with formatting include entities array in output. `comms send --format markdown` sends with MarkdownV2 parse mode. Tests cover entity extraction and formatted sending.

