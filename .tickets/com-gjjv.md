---
id: com-gjjv
status: open
deps: [com-p48x]
links: []
created: 2026-03-20T01:05:27Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Create provider CLI tests

## Summary
Move detailed telegram API mocking tests from `internal/cli/send_test.go` and `internal/cli/react_test.go` to the provider package. These tests exercise the telegram SDK calls directly via the provider's NewSendCmd and NewReactCmd.

## Files
- Create: `providers/telegram/cli_test.go`

## Implementation Notes
- Move telegram API mocking tests from internal/cli/send_test.go and internal/cli/react_test.go
- Tests should exercise NewSendCmd() and NewReactCmd() directly
- Use existing BotAPI mock patterns from the telegram package

## Verification
```bash
go test ./providers/telegram/ -v
```

## Acceptance Criteria
- Telegram API mocking tests live in provider package
- Tests pass

