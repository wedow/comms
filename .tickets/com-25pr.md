---
id: com-25pr
status: open
deps: [com-bx3c, com-pjvo, com-66kj, com-1wlb]
links: []
created: 2026-03-01T13:17:52Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-2]
---
# comms-v1: Telegram Send function

## Summary
Implement a function that sends a text message to a Telegram chat and returns a `message.Message` struct representing the sent message.

## Files
- `internal/provider/telegram/send.go` -- `Send` function
- `internal/provider/telegram/send_test.go` -- tests using mock BotAPI

## Implementation Notes
Function signature:
```go
func Send(ctx context.Context, api BotAPI, chatID int64, text string) (message.Message, error)
```

Behavior:
1. Call `api.SendMessage` with the chatID and text.
2. Convert the returned `*models.Message` into a `message.Message` using `convertMessage` (from the convert.go file, task 2.4).
3. Return the message or a wrapped error.

The `convertMessage` helper is defined in task 2.4 (convert.go). This task depends on that helper existing. The field mapping used by `convertMessage` is:
- `From`: `msg.From.Username` (or `msg.From.FirstName` if username empty, or `"unknown"` if From nil)
- `Provider`: `"telegram"`
- `Channel`: `SlugifyChat(msg.Chat)`
- `Date`: `time.Unix(int64(msg.Date), 0).UTC()`
- `ID`: `fmt.Sprintf("telegram-%d", msg.ID)`
- `Body`: `msg.Text`

For the mock in tests, define a mock struct implementing BotAPI:
```go
type mockBot struct {
    sendFn func(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error)
}
func (m *mockBot) SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error) {
    return m.sendFn(ctx, params)
}
```

Required test cases:
- Successful send: verify returned message fields match expected values
- API error: verify error is propagated with context
- Nil response from API: verify graceful error

## Verification
```sh
cd /home/greg/p/comms && go test ./internal/provider/telegram/ -run TestSend -v
```

## Acceptance Criteria
- All tests pass. Returned message fields match expected values. Errors propagate correctly.

