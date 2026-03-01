---
id: com-fvwr
status: closed
deps: [com-9par, com-yvif]
links: []
created: 2026-03-01T13:18:15Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-4]
---
# comms-v1: comms list command

## Summary
Implement `comms list` CLI command. Lists all messages across all channels (or a specific channel with `--channel`). No cursor side effects -- read-only view of the full message history. Each message is printed as one JSON line.

Flags:
- `--channel <name>` -- optional, filter to a single channel directory

`--channel` takes the full channel directory name as shown by `comms channels` (e.g., `telegram-general`).

Output per line (all fields from `message.Message` struct plus file path):
```json
{"from":"alice","provider":"telegram","channel":"general","date":"2026-02-24T14:30:05Z","id":"telegram-98765","body":"Hey, the deploy finished.","file":".comms/telegram-general/2026-02-24T14-30-05.123.md"}
```

Flow:
1. Call `store.ListChannels()` (or filter to `--channel`)
2. For each channel, call `store.ListMessages(channel)` to get message file paths sorted chronologically
3. For each file, call `store.ReadMessage()` to parse the message (format auto-detected from file extension)
4. Serialize to JSON line and print to stdout

## Files
- `internal/cli/list.go` -- command definition
- `internal/cli/list_test.go` -- tests with pre-populated temp `.comms/` directory

## Implementation Notes
- Dependencies: Task 4.1 (JSON output helpers), Task 1.5 (store.WriteMessage/ReadMessage), Task 1.6 (store.ListChannels/ListMessages)
- The command constructor should follow the pattern `cli.NewListCmd() *cobra.Command`
- Use `cli.PrintJSON()` from Task 4.1 to output each message
- Use `cli.PrintError()` for error cases
- Test cases to implement:
  - List all channels when no `--channel` flag
  - List single channel with `--channel`
  - Empty channel directory produces no output (exit 0)
  - Invalid `--channel` name produces JSON error on stderr, exit 1

## Verification
```sh
cd /home/greg/p/comms && go test ./internal/cli/ -run TestList -v
```

## Acceptance Criteria
- All test cases pass
- Output is valid JSON lines with correct fields

