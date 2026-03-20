---
id: com-6bya
status: closed
deps: [com-ho3r, com-lnh0, com-2nnd, com-7pm4]
links: []
created: 2026-03-20T01:05:31Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Subprocess tests

## Summary
Tests using `os.Pipe()` pairs and helper goroutines simulating provider behavior. Test full spawn-handshake-running lifecycle, crash detection, shutdown timeout.

## Files
- Create: `internal/daemon/provider_test.go`

## Implementation Notes
- Use os.Pipe() pairs to simulate subprocess stdin/stdout
- Helper goroutines simulate provider binary behavior (write ready event, process commands)
- Test scenarios:
  - Full spawn-handshake-running lifecycle
  - Crash detection (process exit triggers event)
  - Shutdown timeout (force kill after 5s)
  - SendCommand with proper locking
  - RespawnManager backoff behavior
  - RespawnManager stable reset
  - RespawnManager fatal after 5 failures
- Use swappable `startProcess` var for test injection

## Verification
```bash
go test ./internal/daemon/ -run TestSubprocess -v
go test ./internal/daemon/ -run TestRespawn -v
```

## Acceptance Criteria
- Full subprocess lifecycle tested. Crash detection tested. Shutdown timeout tested.


## Notes

**2026-03-20T02:53:56Z**

All tests already written as part of TDD during com-ho3r, com-2nnd, com-lnh0, com-7pm4 implementation. Tests in subprocess_test.go and respawn_test.go cover: spawn handshake, ready timeout, bad ready, SendCommand, SendCommand canceled, shutdown graceful, shutdown timeout, respawn recovery, backoff, max failures, stability reset, context cancel.
