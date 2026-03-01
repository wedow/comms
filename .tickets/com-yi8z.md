---
id: com-yi8z
status: open
deps: [com-9par, com-9oxg, com-25pr, com-pwp2]
links: []
created: 2026-03-01T13:18:36Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-4]
---
# comms-v1: comms send command

## Summary
Implement `comms send` CLI command. Minimal send command that reads message body from stdin and sends via the telegram provider.

Flags:
- `--channel <name>` -- required, the channel to send to

`--channel` takes the full channel directory name as shown by `comms channels` (e.g., `telegram-general`).

Flow:
1. Load config to get telegram token
2. Read all of stdin as the message body (trimmed)
3. If body is empty, emit JSON error to stderr, exit 1
4. Resolve channel name to a Telegram chat ID via `store.ReadChatID(root, channel)` (from Phase 3, Task 3.9). The daemon writes `.chat_id` files when it discovers channels during polling.
5. Call `telegram.Send(ctx, api, chatID, body)` (from Phase 2, Task 2.3)
6. Print result as JSON line: `{"ok":true,"channel":"<name>"}`

Error cases:
- No stdin / empty body: JSON error on stderr, exit 1
- Channel not found / no `.chat_id` file: JSON error on stderr, exit 1
- Telegram API error: JSON error on stderr, exit 1

## Files
- `internal/cli/send.go` -- command definition
- `internal/cli/send_test.go` -- tests (mock the telegram provider via BotAPI interface)

## Implementation Notes
- Dependencies: Task 4.1 (JSON output helpers), Task 1.4 (config), Task 2.1 (telegram.BotAPI interface), Task 2.3 (telegram.Send), Task 3.9 (store.ReadChatID)
- The command constructor should follow the pattern `cli.NewSendCmd() *cobra.Command`
- Use `cli.PrintJSON()` from Task 4.1 for success output
- Use `cli.PrintError()` from Task 4.1 for all error cases
- Mock the telegram provider via the BotAPI interface for testing
- Test cases to implement:
  - Successful send with mock: verify JSON success output
  - Empty stdin: verify JSON error
  - Missing channel: verify JSON error
  - Missing `.chat_id` file: verify JSON error

## Verification
```sh
cd /home/greg/p/comms && go test ./internal/cli/ -run TestSend -v
```

## Acceptance Criteria
- All test cases pass
- JSON output format is correct
- Errors emit to stderr

