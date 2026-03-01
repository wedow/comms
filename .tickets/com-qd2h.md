---
id: com-qd2h
status: closed
deps: [com-bx3c]
links: []
created: 2026-03-01T13:18:12Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-2]
---
# comms-v1: NewBot factory function

## Summary
Implement a convenience function that wraps `bot.New()` and returns a `BotAPI`. This is what production code passes as the bot constructor.

## Files
- `internal/provider/telegram/bot.go` -- `NewBot` function

## Implementation Notes
Signature:
```go
func NewBot(token string) (BotAPI, error)
```

Since `*bot.Bot` already satisfies `BotAPI` (it has `SendMessage`), this is a one-liner plus error handling. Pass `bot.WithSkipGetMe()` to avoid an unnecessary `getMe` network call on construction.

Implementation:
```go
func NewBot(token string) (BotAPI, error) {
    b, err := bot.New(token, bot.WithSkipGetMe())
    if err != nil {
        return nil, err
    }
    return b, nil
}
```

No tests needed -- this is a trivial wrapper. Verification is that the package builds and vets cleanly.

## Verification
```sh
cd /home/greg/p/comms && go build ./internal/provider/telegram/...
cd /home/greg/p/comms && go vet ./internal/provider/telegram/...
```

## Acceptance Criteria
- Package builds and vets without error.

