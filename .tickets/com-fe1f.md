---
id: com-fe1f
status: open
deps: [com-bl92, com-w70v]
links: []
created: 2026-03-20T01:06:32Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Refactor CLI daemon command

## Summary
Remove `telegramProvider` struct, all `go-telegram/bot` imports, and direct `telegram.Poll()` call from `cli/daemon.go`. The `run` subcommand becomes: load config, check not already running, extract provider names, call `daemon.Run()`.

## Files
- Modify: `internal/cli/daemon.go`

## Implementation Notes
- Delete the telegramProvider struct entirely
- Remove all go-telegram/bot imports from cli/daemon.go
- Remove direct telegram.Poll() call
- The run subcommand flow: load config -> check PID -> extract provider names via config.ProviderNames() -> call daemon.Run(ctx, cfg, root, providers)
- No telegram-specific code should remain in cli/daemon.go

## Verification
```bash
go test ./internal/cli/ -run TestDaemon -v
grep -r "go-telegram" internal/cli/  # should produce no output
```

## Acceptance Criteria
- cli/daemon.go has zero telegram imports
- Uses new daemon.Run() API

