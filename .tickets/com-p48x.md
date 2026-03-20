---
id: com-p48x
status: open
deps: [com-20s3, com-op5m]
links: []
created: 2026-03-20T01:05:02Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Create provider-specific CLI commands

## Summary
Create `providers/telegram/cli.go` with `NewSendCmd()` and `NewReactCmd()` Cobra factories. These call telegram SDK directly (not via delegation). Reads config using `COMMS_PROVIDER_CONFIG` env var (set by base binary via `config.ProviderConfig(provider)` from Phase 01). The provider binary is a pure network client: sends/receives via Telegram API, returns results as JSON on stdout, never touches the store. The `--format` flag is interpreted here (not in base binary -- `parseFormatFlag` is deleted from `send.go`). Config from `COMMS_PROVIDER_CONFIG`; no reference to `config.TelegramToken()`. Key difference from current `cli/send.go`: reads config from env var and calls `telegram.NewBot(token)` directly (no `newBot` parameter injection).

## Files
- Create: `providers/telegram/cli.go`

## Implementation Notes
- NewSendCmd() and NewReactCmd() are Cobra command factories
- Read config from COMMS_PROVIDER_CONFIG env var (JSON format)
- Call telegram.NewBot(token) directly -- no newBot parameter injection
- Provider binary is a pure network client: sends/receives via Telegram API, returns results as JSON on stdout, never touches the store
- The --format flag is interpreted here (moved from base binary's send.go)

## Verification
```bash
go build ./providers/telegram/
```

## Acceptance Criteria
- NewSendCmd() and NewReactCmd() factories exist
- Read config from COMMS_PROVIDER_CONFIG env var

