---
id: com-r8wx
status: closed
deps: [com-9par, com-fvwr, com-hogs, com-yi8z, com-b40d]
links: []
created: 2026-03-01T13:18:43Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-4]
---
# comms-v1: Wire commands to root Cobra command

## Summary
Register `channels`, `list`, `unread`, and `send` as subcommands on the root Cobra command. Each command's `cobra.Command` is constructed by a function in its respective `internal/cli/*.go` file (e.g., `cli.NewChannelsCmd()`) and added to root.

## Files
- `internal/cli/root.go` -- add `rootCmd.AddCommand(...)` calls (modify existing file from Phase 1 Task 1.1)

## Implementation Notes
- Dependencies: Task 4.2 (channels cmd), Task 4.3 (list cmd), Task 4.4 (unread cmd), Task 4.5 (send cmd), Task 1.1 (root cmd)
- Each command file should export a constructor function: `NewChannelsCmd()`, `NewListCmd()`, `NewUnreadCmd()`, `NewSendCmd()`
- In `root.go`, add calls like: `rootCmd.AddCommand(NewChannelsCmd(), NewListCmd(), NewUnreadCmd(), NewSendCmd())`
- This is a simple wiring task -- no new logic, just connecting existing commands

## Verification
```sh
cd /home/greg/p/comms && go build ./cmd/comms && ./cmd/comms --help | grep -E 'channels|list|unread|send'
```

## Acceptance Criteria
- `--help` output lists all four commands: channels, list, unread, send

