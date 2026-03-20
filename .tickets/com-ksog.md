---
id: com-ksog
status: closed
deps: [com-20s3, com-p48x]
links: []
created: 2026-03-20T01:04:53Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Create provider binary entry point

## Summary
Create `cmd/comms-telegram/main.go` -- thin Cobra binary with `send` and `react` subcommands. Separate `main` package that builds as `comms-telegram`. Imports `providers/telegram` for command factories.

## Files
- Create: `cmd/comms-telegram/main.go`

## Implementation Notes
- Thin Cobra binary with `send` and `react` subcommands
- Separate `main` package that builds as `comms-telegram`
- Imports `providers/telegram` for command factories (NewSendCmd, NewReactCmd)

## Verification
```bash
go build -o /tmp/comms-telegram ./cmd/comms-telegram
```

## Acceptance Criteria
- `comms-telegram` binary builds with send and react subcommands

