---
id: com-plln
status: open
deps: [com-20s3]
links: []
created: 2026-03-20T01:04:46Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Create CLI delegation infrastructure

## Summary
Implement `resolveProviderBinary()`, `delegate()`, `extractProvider()` in `internal/cli/delegate.go`. Use swappable package-level vars for testing (same pattern as `runSystemctl` in `install.go`). `resolveProviderBinary` calls `exec.LookPath("comms-" + provider)`. `delegate` spawns provider binary with args, inherits stdin/stdout/stderr, passes through exit code. `extractProvider` extracts provider name from channel: `"telegram-general"` -> `"telegram"`.

## Files
- Create: `internal/cli/delegate.go`

## Implementation Notes
- Use swappable package-level vars for testing (same pattern as `runSystemctl` in `install.go`)
- `resolveProviderBinary` calls `exec.LookPath("comms-" + provider)`
- `delegate` spawns provider binary with args, inherits stdin/stdout/stderr, passes through exit code
- `extractProvider` extracts provider name from channel: `"telegram-general"` -> `"telegram"`

## Verification
```bash
go test ./internal/cli/ -v
```

## Acceptance Criteria
- Functions exist with swappable vars for testing
- extractProvider correctly parses channel names

