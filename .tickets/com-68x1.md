---
id: com-68x1
status: open
deps: [com-ksog, com-ab0w]
links: []
created: 2026-03-20T01:08:44Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Add subprocess command to provider binary

## Summary
Add hidden subprocess subcommand to cmd/comms-telegram/main.go. Command reads COMMS_PROVIDER_CONFIG env var (JSON), parses into telegram.ProviderConfig, calls telegram.RunSubprocess(ctx, os.Stdin, os.Stdout, cfg).

## Files
- Modify: cmd/comms-telegram/main.go

## Implementation Notes
- The subprocess command should be hidden (not shown in help)
- Reads config from COMMS_PROVIDER_CONFIG env var as JSON
- Parses into telegram.ProviderConfig struct
- Calls telegram.RunSubprocess(ctx, os.Stdin, os.Stdout, cfg)
- This is the entry point the daemon uses when spawning provider subprocesses

## Verification
```bash
go build -o /tmp/comms-telegram ./cmd/comms-telegram
```

## Acceptance Criteria
- comms-telegram subprocess command exists (hidden)
- Reads config from env var

