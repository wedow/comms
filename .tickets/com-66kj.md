---
id: com-66kj
status: open
deps: [com-bx3c, com-1wlb]
links: []
created: 2026-03-01T13:18:06Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-2]
---
# comms-v1: Poll function and convertMessage helper

## Summary
Implement the shared `convertMessage` helper and the `Poll` function that starts long-polling for Telegram updates.

## Files
- `internal/provider/telegram/slug.go` -- `SlugifyChat` function
- `internal/provider/telegram/slug_test.go` -- table-driven tests for SlugifyChat
- `internal/provider/telegram/convert.go` -- `convertMessage` helper
- `internal/provider/telegram/convert_test.go` -- unit tests for the converter
- `internal/provider/telegram/poll.go` -- `Poll` function
- `internal/provider/telegram/poll_test.go` -- handler wiring tests using `ProcessUpdate`

## Implementation Notes

### SlugifyChat
Pure function that converts a Telegram `models.Chat` into a filesystem-safe channel name string.

Signature:
```go
func SlugifyChat(chat models.Chat) string
```

Rules:
1. Group/supergroup/channel -> slugify title: lowercase, replace `[^a-z0-9]` with `-`, collapse consecutive hyphens, trim leading/trailing hyphens.
2. Private chat -> username if non-empty, else `dm-<chat-id>`.
3. Empty title on a non-private chat -> fall back to `chat-<chat-id>`.

Test cases (table-driven):
- Group with title "My Cool Group" -> `my-cool-group`
- Title with special chars "Alerts!!! & Stuff" -> `alerts-stuff`
- Title with unicode -- strip non-ASCII -> hyphens collapsed
- Private chat with username "alice" -> `alice`
- Private chat without username, ID 12345 -> `dm-12345`
- Group with empty title, ID 67890 -> `chat-67890`
- Title with leading/trailing spaces and hyphens -> trimmed

### convertMessage
Signature:
```go
func convertMessage(m *models.Message) message.Message
```

Field mapping:
- `From`: `msg.From.Username` (or `msg.From.FirstName` if username empty, or `"unknown"` if From nil)
- `Provider`: `"telegram"`
- `Channel`: `SlugifyChat(msg.Chat)`
- `Date`: `time.Unix(int64(msg.Date), 0).UTC()`
- `ID`: `fmt.Sprintf("telegram-%d", msg.ID)`
- `Body`: `msg.Text`

Test cases for convertMessage:
- Message from a user in a group: all fields populated correctly
- Message from a user in a DM: channel derived from username
- Message with empty text: body is empty string
- Unix timestamp 0: date is zero time
- Message with `From` nil: `From` defaults to "unknown"

### Poll
Signature:
```go
func Poll(ctx context.Context, token string, initialOffset int64, handler func(msg message.Message, chatID int64)) (lastOffset int64, err error)
```

Note: the handler receives `chatID int64` alongside the message so the daemon can maintain its chat registry without re-parsing.

Poll behavior:
1. Construct a `*bot.Bot` using `bot.New(token, bot.WithSkipGetMe(), bot.WithInitialOffset(initialOffset), ...)` with a default handler that converts updates. Note: `bot.WithInitialOffset(initialOffset)` sets the starting offset for polling so the bot resumes from where it left off.
2. The default handler, for each update where `update.Message != nil`: build a `message.Message` via `convertMessage`, track `update.ID + 1` as the next offset, call `handler(msg, update.Message.Chat.ID)`.
3. Call `b.Start(ctx)` -- blocks until ctx is cancelled.
4. Return the last tracked offset.

Test cases for handler wiring:
- ProcessUpdate with a text message: handler called with correct message.Message
- ProcessUpdate with a non-message update (e.g., callback query): handler not called
- ProcessUpdate with message but nil chat: no panic, handler not called

## Verification
```sh
cd /home/greg/p/comms && go test ./internal/provider/telegram/ -run TestConvert -v
cd /home/greg/p/comms && go test ./internal/provider/telegram/ -run TestPoll -v
```

## Acceptance Criteria
- All tests pass. convertMessage produces correct field mappings. Handler fires for valid updates only.

