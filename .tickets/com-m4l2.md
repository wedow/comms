---
id: com-m4l2
status: open
deps: []
links: []
created: 2026-03-01T13:18:09Z
type: task
priority: 1
assignee: Greg Wedow
tags: [phase-1]
---
# comms-v1: Embedded docs

## Summary
Create the `internal/embeddocs/` directory with `telegram-setup.md` -- the setup guide that `comms init` deploys to `.comms/docs/`. Use Go's `//go:embed` directive to bundle the file into the binary. The package is named `embeddocs` (not `embed`) to avoid shadowing Go's built-in `embed` package.

The guide covers: creating a Telegram bot via BotFather, obtaining the token, adding the bot to groups, and configuring `config.toml`. Written for an AI agent reader (direct instructions, no screenshots).

## Files
- `internal/embeddocs/embed.go` -- `//go:embed telegram-setup.md` as `var TelegramSetupDoc []byte` (or `string`)
- `internal/embeddocs/telegram-setup.md` -- setup guide content

## Implementation Notes
- No dependencies on other tasks
- Package name must be `embeddocs` (not `embed`) to avoid shadowing Go's built-in `embed` package
- Use `//go:embed telegram-setup.md` directive
- Export as `var TelegramSetupDoc []byte` (or `string`)
- The telegram-setup.md content should cover:
  1. Creating a Telegram bot via BotFather
  2. Obtaining the token
  3. Adding the bot to groups
  4. Configuring `config.toml`
- Written for an AI agent reader (direct instructions, no screenshots)

## Verification
```sh
go build ./internal/embeddocs
go vet ./internal/embeddocs
```

## Acceptance Criteria
- Package builds and vets without error

