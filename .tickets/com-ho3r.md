---
id: com-ho3r
status: open
deps: [com-mnvn, com-gvr2]
links: []
created: 2026-03-20T01:05:10Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Subprocess struct and spawn

## Summary
Define `Subprocess` struct (cmd, pipes, state, event channel). Implement `Spawn(ctx, provider, binaryPath, root, providerConfig, offset) (*Subprocess, error)`: create `exec.Command(binaryPath, "subprocess")`, set `COMMS_ROOT` and `COMMS_PROVIDER_CONFIG` env vars, create stdin/stdout/stderr pipes, start process, read stdout with 10s deadline for `ready` event, write `start` command with offset, start stdout reader goroutine (posting to event channel), start stderr copy goroutine, start `cmd.Wait()` goroutine for crash detection. Use swappable `startProcess` var for test injection. Provider binary resolution: call `exec.LookPath("comms-" + provider)` inline.

## Files
- Create: `internal/daemon/provider.go`

## Implementation Notes
- Subprocess struct holds: cmd (*exec.Cmd), stdin/stdout/stderr pipes, state, event channel
- Spawn function handles full lifecycle: create command, set env vars, pipe setup, start, handshake, goroutine launch
- Use swappable `startProcess` var (same pattern as `runSystemctl` in install.go) for test injection
- 10s deadline for ready event during handshake
- Reader goroutine posts decoded events to channel
- Stderr copy goroutine for logging
- cmd.Wait() goroutine for crash detection

## Verification
```bash
go test ./internal/daemon/ -run TestSubprocess -v
```

## Acceptance Criteria
- Subprocess spawns, completes handshake, reads events into channel.

