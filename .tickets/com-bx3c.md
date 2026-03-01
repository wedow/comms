---
id: com-bx3c
status: open
deps: [com-b40d]
links: []
created: 2026-03-01T13:17:31Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-2]
---
# comms-v1: BotAPI interface and package skeleton

## Summary
Create the `internal/provider/telegram` package directory and define the narrow interface that wraps the `go-telegram/bot.Bot` methods we use. This interface enables mock-based testing without hitting Telegram's API.

Interface definition:
```go
// BotAPI is the subset of *bot.Bot methods this package uses.
type BotAPI interface {
    SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error)
}
```

## Files
- `internal/provider/telegram/telegram.go` -- package doc, BotAPI interface
- `internal/provider/telegram/telegram_test.go` -- placeholder (confirms package compiles)
- `internal/provider/telegram/bot.go` -- `NewBot` factory function

## Implementation Notes
- Import `github.com/go-telegram/bot` for the `*bot.SendMessageParams` type
- Import `github.com/go-telegram/bot/models` for the `*models.Message` return type
- The interface is deliberately narrow -- only the methods this package actually uses
- Include a package-level doc comment explaining the package purpose
- The placeholder test just needs to confirm the package compiles (e.g., `func TestPackageCompiles(t *testing.T) {}`)

## Also: Add go-telegram/bot dependency
- Run `go get github.com/go-telegram/bot@latest` to add the dependency to go.mod/go.sum
- Run `go mod tidy` to clean up
- Dependency `github.com/go-telegram/bot` must appear in go.mod before the package code can compile

## Also: NewBot factory function
Implement a convenience function that wraps `bot.New()` and returns a `BotAPI`:
```go
func NewBot(token string) (BotAPI, error) {
    b, err := bot.New(token, bot.WithSkipGetMe())
    if err != nil {
        return nil, err
    }
    return b, nil
}
```
Since `*bot.Bot` already satisfies `BotAPI`, this is a trivial wrapper. No tests needed beyond package compilation.

## Verification
```sh
cd /home/greg/p/comms && go build ./internal/provider/telegram/...
cd /home/greg/p/comms && go vet ./internal/provider/telegram/...
```

## Acceptance Criteria
- Package builds and vets without error

