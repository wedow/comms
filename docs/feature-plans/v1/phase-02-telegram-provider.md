# Phase 2: Telegram Provider

Implements `internal/provider/telegram` -- the package that wraps `github.com/go-telegram/bot` to provide polling and sending for the daemon and CLI.

**Depends on:** Phase 1 (message.Message struct must exist).

## Design Overview

The package exposes two stateless-style entry points:

- `Poll(ctx, token, offset, handler)` -- starts long-polling, converts Telegram updates to `message.Message`, calls handler for each. Returns the last update offset on context cancellation.
- `Send(ctx, api, chatID, text)` -- sends a text message to a chat, returns the sent `message.Message`.

For testability, a narrow `BotAPI` interface covers only the methods we call (`SendMessage`). Poll testing uses `ProcessUpdate` to feed synthetic updates into the bot's handler pipeline.

### Chat name resolution

`SlugifyChat(chat)` converts a Telegram chat to a directory-safe channel name:
- Group/supergroup/channel: slugify the chat title (lowercase, replace non-alphanum with hyphens, collapse runs, trim)
- Private (DM): use username if available, otherwise `"dm-<chat-id>"`

### Persistent state

Two pieces of state live on the filesystem (managed by callers, not this package):
- **Poll offset**: The daemon passes the last known offset to `Poll()` and persists the returned offset. This package does not touch the filesystem.
- **Chat ID mapping**: `Send()` takes a numeric `chatID int64` directly. The caller (CLI/daemon) resolves channel names to chat IDs via `.chat_id` files (see Phase 3, Task 3.10).

This keeps the provider package a pure API wrapper with no filesystem dependency.

---

## Task 2.1: Define BotAPI interface and package skeleton

**Description:** Create the package directory and define the narrow interface that wraps the `go-telegram/bot.Bot` methods we use. This interface enables mock-based testing without hitting Telegram's API.

**Interface definition:**
```go
// BotAPI is the subset of *bot.Bot methods this package uses.
type BotAPI interface {
    SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error)
}
```

Also define a factory type for dependency injection:
```go
// NewBotFunc creates a BotAPI from a token.
type NewBotFunc func(token string, opts ...bot.Option) (BotAPI, error)
```

**Files:**
- `internal/provider/telegram/telegram.go` -- package doc, BotAPI interface, NewBotFunc type
- `internal/provider/telegram/telegram_test.go` -- placeholder (confirms package compiles)

**Dependencies:** Task 2.6 (go-telegram/bot dependency must be in go.mod)

**Verification:**
```sh
cd /home/greg/p/comms && go build ./internal/provider/telegram/...
cd /home/greg/p/comms && go vet ./internal/provider/telegram/...
```

**Pass/fail:** Package builds and vets without error.

---

## Task 2.2: Implement SlugifyChat

**Description:** Pure function that converts a Telegram `models.Chat` into a filesystem-safe channel name string.

Rules:
1. Group/supergroup/channel -> slugify title: lowercase, replace `[^a-z0-9]` with `-`, collapse consecutive hyphens, trim leading/trailing hyphens.
2. Private chat -> username if non-empty, else `dm-<chat-id>`.
3. Empty title on a non-private chat -> fall back to `chat-<chat-id>`.

**Files:**
- `internal/provider/telegram/slug.go` -- `SlugifyChat` function
- `internal/provider/telegram/slug_test.go` -- table-driven tests

**Test cases (minimum):**
- Group with title "My Cool Group" -> `my-cool-group`
- Title with special chars "Alerts!!! & Stuff" -> `alerts-stuff`
- Title with unicode -- strip non-ASCII
- Private chat with username "alice" -> `alice`
- Private chat without username, ID 12345 -> `dm-12345`
- Group with empty title, ID 67890 -> `chat-67890`
- Title with leading/trailing spaces and hyphens -> trimmed

**Dependencies:** Task 2.1 (needs package to exist)

**Verification:**
```sh
cd /home/greg/p/comms && go test ./internal/provider/telegram/ -run TestSlugifyChat -v
```

**Pass/fail:** All table-driven tests pass.

---

## Task 2.3: Implement Send

**Description:** Function that sends a text message to a Telegram chat and returns a `message.Message` struct representing the sent message.

**Signature:**
```go
func Send(ctx context.Context, api BotAPI, chatID int64, text string) (message.Message, error)
```

Behavior:
1. Call `api.SendMessage` with the chatID and text.
2. Convert the returned `*models.Message` into a `message.Message` using `convertMessage` (Task 2.4).
3. Return the message or a wrapped error.

Field mapping (in `convertMessage`):
- `From`: `msg.From.Username` (or `msg.From.FirstName` if username empty, or `"unknown"` if From nil)
- `Provider`: `"telegram"`
- `Channel`: `SlugifyChat(msg.Chat)`
- `Date`: `time.Unix(msg.Date, 0).UTC()`
- `ID`: `fmt.Sprintf("telegram-%d", msg.MessageID)`
- `Body`: `msg.Text`

**Files:**
- `internal/provider/telegram/send.go` -- `Send` function
- `internal/provider/telegram/send_test.go` -- tests using mock BotAPI

**Test cases:**
- Successful send: verify returned message fields match expected values
- API error: verify error is propagated with context
- Nil response from API: verify graceful error

**Dependencies:** Task 2.1 (BotAPI interface), Task 2.2 (SlugifyChat), Task 2.4 (convertMessage), Phase 1 Task 1.2 (message.Message)

**Verification:**
```sh
cd /home/greg/p/comms && go test ./internal/provider/telegram/ -run TestSend -v
```

**Pass/fail:** All tests pass. Returned message fields match expected values. Errors propagate correctly.

---

## Task 2.4: Implement Poll and convertMessage

**Description:** Implement the shared `convertMessage` helper and the `Poll` function that starts long-polling for Telegram updates.

**convertMessage signature:**
```go
func convertMessage(m *models.Message) message.Message
```

**Poll signature:**
```go
func Poll(ctx context.Context, token string, initialOffset int64, handler func(msg message.Message, chatID int64)) (lastOffset int64, err error)
```

Note: the handler receives `chatID int64` alongside the message so the daemon can maintain its chat registry without re-parsing.

Poll behavior:
1. Construct a `*bot.Bot` using `bot.New(token, opts...)` with a default handler that converts updates.
2. The default handler, for each update where `update.Message != nil`: build a `message.Message` via `convertMessage`, track `update.UpdateID + 1` as the next offset, call `handler(msg, update.Message.Chat.ID)`.
3. Call `b.Start(ctx)` -- blocks until ctx is cancelled.
4. Return the last tracked offset.

**Files:**
- `internal/provider/telegram/convert.go` -- `convertMessage` helper
- `internal/provider/telegram/convert_test.go` -- unit tests for the converter
- `internal/provider/telegram/poll.go` -- `Poll` function
- `internal/provider/telegram/poll_test.go` -- handler wiring tests using `ProcessUpdate`

**Test cases for convertMessage:**
- Message from a user in a group: all fields populated correctly
- Message from a user in a DM: channel derived from username
- Message with empty text: body is empty string
- Unix timestamp 0: date is zero time
- Message with `From` nil: `From` defaults to "unknown"

**Test cases for handler wiring:**
- ProcessUpdate with a text message: handler called with correct message.Message
- ProcessUpdate with a non-message update (e.g., callback query): handler not called
- ProcessUpdate with message but nil chat: no panic, handler not called

**Dependencies:** Task 2.1 (package skeleton), Task 2.2 (SlugifyChat), Phase 1 Task 1.2 (message.Message)

**Verification:**
```sh
cd /home/greg/p/comms && go test ./internal/provider/telegram/ -run TestConvert -v
cd /home/greg/p/comms && go test ./internal/provider/telegram/ -run TestPoll -v
```

**Pass/fail:** All tests pass. convertMessage produces correct field mappings. Handler fires for valid updates only.

---

## Task 2.5: Implement NewBot factory function

**Description:** Convenience function that wraps `bot.New()` and returns a `BotAPI`. This is what production code passes as the bot constructor.

**Signature:**
```go
func NewBot(token string) (BotAPI, error)
```

Since `*bot.Bot` already satisfies `BotAPI` (it has `SendMessage`), this is a one-liner plus error handling.

**Files:**
- `internal/provider/telegram/bot.go` -- `NewBot` function

**Dependencies:** Task 2.1 (BotAPI interface)

**Verification:**
```sh
cd /home/greg/p/comms && go build ./internal/provider/telegram/...
cd /home/greg/p/comms && go vet ./internal/provider/telegram/...
```

**Pass/fail:** Package builds and vets without error.

---

## Task 2.6: Add go-telegram/bot dependency

**Description:** Run `go get` to add the library to go.mod/go.sum.

**Files:**
- `go.mod` (updated)
- `go.sum` (updated)

**Dependencies:** none (can be done first or alongside Task 2.1)

**Verification:**
```sh
cd /home/greg/p/comms && go get github.com/go-telegram/bot@latest
cd /home/greg/p/comms && go mod tidy
cd /home/greg/p/comms && go build ./...
```

**Pass/fail:** `go build ./...` exits 0. Dependency appears in go.mod.

---

## Task Dependency Graph

```
Task 2.6 (go get dependency)
  |
  v
Task 2.1 (interface + skeleton)
  |
  +---> Task 2.2 (SlugifyChat)
  |       |
  |       +---> Task 2.3 (Send) -- also depends on 2.4 (convertMessage)
  |       |
  |       +---> Task 2.4 (Poll + convertMessage)
  |
  +---> Task 2.5 (NewBot factory)
```

Tasks 2.2 and 2.5 can proceed in parallel after 2.1. Tasks 2.3 and 2.4 depend on 2.2 but are independent of each other (though 2.3 uses convertMessage from 2.4).

## Files Created in This Phase

| File | Purpose |
|------|---------|
| `internal/provider/telegram/telegram.go` | Package doc, BotAPI interface, NewBotFunc type |
| `internal/provider/telegram/telegram_test.go` | Compilation smoke test |
| `internal/provider/telegram/slug.go` | SlugifyChat function |
| `internal/provider/telegram/slug_test.go` | Slug tests |
| `internal/provider/telegram/send.go` | Send function |
| `internal/provider/telegram/send_test.go` | Send tests with mock |
| `internal/provider/telegram/poll.go` | Poll function |
| `internal/provider/telegram/poll_test.go` | Poll handler wiring tests |
| `internal/provider/telegram/convert.go` | convertMessage helper |
| `internal/provider/telegram/convert_test.go` | Converter tests |
| `internal/provider/telegram/bot.go` | NewBot factory |

## Exit Criteria

```sh
cd /home/greg/p/comms && go build ./internal/provider/telegram/...
cd /home/greg/p/comms && go vet ./internal/provider/telegram/...
cd /home/greg/p/comms && go test ./internal/provider/telegram/ -v
```

All three commands exit 0. The package has zero filesystem dependencies and depends only on `message.Message` from the project and `github.com/go-telegram/bot` externally.
