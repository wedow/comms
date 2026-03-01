---
id: com-mgu1
status: open
deps: [com-77c4, com-n1t1]
links: []
created: 2026-03-01T13:17:58Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-3]
---
# comms-v1: comms daemon stop CLI command

## Summary
Implement `comms daemon stop`. Reads the PID file, verifies the process is alive, sends SIGTERM, waits briefly for the process to exit, then confirms. If no PID file or process is not running, report that clearly as JSON on stderr.

## Files
- `internal/cli/daemon.go` -- add `stopCmd` subcommand to `daemonCmd` group
- `internal/cli/daemon_test.go` -- test stop when no daemon running (error), test stop with stale PID file (cleanup + error)

## Implementation Notes
- Dependencies: Task 3.1 (PID), Task 3.3 (daemon start -- provides the daemonCmd group)
- The `stopCmd` is added as a subcommand of the existing `daemonCmd` group
- Flow: read PID file -> check if alive -> send SIGTERM -> wait briefly -> confirm
- If no PID file exists: report JSON error on stderr
- If PID file exists but process is dead (stale): clean up PID file, report JSON error
- Error output format: JSON on stderr

## Verification
```sh
cd /home/greg/p/comms && go build ./cmd/comms
cd /home/greg/p/comms && ./cmd/comms daemon stop --help | grep -q 'stop'
cd /home/greg/p/comms && go test ./internal/cli/... -run TestDaemonStop -v
```

## Acceptance Criteria
- Tests pass
- Stop with no daemon reports JSON error
- Stop with stale PID cleans up

