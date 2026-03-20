---
id: com-d7ni
status: open
deps: [com-l7o6, com-3w5g]
links: []
created: 2026-03-20T01:10:24Z
type: task
priority: 3
assignee: Greg Wedow
---
# Plugin Arch: Final end-to-end verification

## Summary
Run complete verification suite and fix any remaining issues.

## Files
- None (verification only, fix any issues found)

## Implementation Notes
- This is the final verification task for the entire plugin architecture
- Both binaries must build and run correctly
- All tests must pass across all packages
- Zero telegram imports should exist in the base binary (cmd/comms, internal/cli, internal/daemon, internal/store, internal/config, internal/message)
- Telegram imports should only exist in providers/telegram and cmd/comms-telegram
- Protocol package should be standalone

## Verification
```bash
go build -o /tmp/comms ./cmd/comms
go build -o /tmp/comms-telegram ./cmd/comms-telegram
go test ./... -v
go vet ./...
grep -r "go-telegram" cmd/comms/ internal/cli/ internal/daemon/ internal/store/ internal/config/ internal/message/  # expect no output
grep -r "go-telegram" providers/telegram/ cmd/comms-telegram/  # expect matches
/tmp/comms --help
/tmp/comms-telegram --help
/tmp/comms-telegram send --help
/tmp/comms-telegram react --help
go test ./internal/protocol/ -v
go test ./internal/config/ -v
```

## Acceptance Criteria
- Both binaries build and run
- All tests pass
- Zero telegram imports in base binary
- Protocol package standalone
- Config works with new format only

