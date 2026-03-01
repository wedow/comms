---
id: com-rsmt
status: open
deps: [com-hyrn, com-n4d7]
links: []
created: 2026-03-01T20:37:19Z
type: feature
priority: 1
assignee: Greg Wedow
parent: com-d35c
tags: [telegram, v2, reactions]
---
# Send reactions via comms react

Add `comms react` CLI command to set a reaction on a message.

- Add `SetMessageReaction` to BotAPI interface: `SetMessageReaction(ctx context.Context, params *bot.SetMessageReactionParams) (bool, error)`
- New CLI command in `internal/cli/react.go`: `comms react --channel <channel> --message <msg-id> --emoji <emoji>`
- Use `ParseMessageID` (from com-zk1m) to extract numeric Telegram message ID
- Build `bot.SetMessageReactionParams{ChatID: chatID, MessageID: id, Reaction: []models.ReactionType{models.ReactionTypeEmoji{Emoji: emoji}}}`
- The `--emoji` flag takes a Unicode emoji character (e.g., `👍`), not a text name
- Wire new command into root.go

**SDK types:** `bot.SetMessageReactionParams{ChatID, MessageID int, Reaction []models.ReactionType, IsBig bool}`. `models.ReactionTypeEmoji{Type: "emoji", Emoji: string}`.

**Example:** `comms react --channel telegram-dm-123 --message telegram-456 --emoji 👍`

**Files to modify:** new `internal/cli/react.go` + test, `internal/provider/telegram/telegram.go` (BotAPI interface), `internal/cli/root.go` (wire command).

## TDD

Write failing tests first. Test that react command calls SetMessageReaction with correct params (mock BotAPI). Test emoji passthrough.

## Verification

```
go test ./internal/cli/... -v
go vet ./...
```

## Acceptance Criteria

`comms react` sets a reaction visible in Telegram. Tests mock the API and verify SetMessageReaction is called with correct ChatID, MessageID, and ReactionType.

