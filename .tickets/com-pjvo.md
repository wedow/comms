---
id: com-pjvo
status: closed
deps: [com-bx3c]
links: []
created: 2026-03-01T13:17:42Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-2]
---
# comms-v1: SlugifyChat function

## Summary
Implement a pure function that converts a Telegram `models.Chat` into a filesystem-safe channel name string.

Rules:
1. Group/supergroup/channel -> slugify title: lowercase, replace `[^a-z0-9]` with `-`, collapse consecutive hyphens, trim leading/trailing hyphens.
2. Private chat -> username if non-empty, else `dm-<chat-id>`.
3. Empty title on a non-private chat -> fall back to `chat-<chat-id>`.

## Files
- `internal/provider/telegram/slug.go` -- `SlugifyChat` function
- `internal/provider/telegram/slug_test.go` -- table-driven tests

## Implementation Notes
Function signature:
```go
func SlugifyChat(chat models.Chat) string
```

The function takes a `models.Chat` (from `github.com/go-telegram/bot/models`) and returns a string.

Slugification algorithm for titles:
- Convert to lowercase
- Replace any character NOT matching `[a-z0-9]` with `-`
- Collapse consecutive hyphens into a single hyphen
- Trim leading and trailing hyphens

Chat type detection: use `chat.Type` field. Private chats have `chat.Type == "private"`.

Required test cases (table-driven):
- Group with title "My Cool Group" -> `my-cool-group`
- Title with special chars "Alerts!!! & Stuff" -> `alerts-stuff`
- Title with unicode -- strip non-ASCII -> hyphens collapsed
- Private chat with username "alice" -> `alice`
- Private chat without username, ID 12345 -> `dm-12345`
- Group with empty title, ID 67890 -> `chat-67890`
- Title with leading/trailing spaces and hyphens -> trimmed

## Verification
```sh
cd /home/greg/p/comms && go test ./internal/provider/telegram/ -run TestSlugifyChat -v
```

## Acceptance Criteria
- All table-driven tests pass

