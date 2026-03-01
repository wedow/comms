---
id: com-n1t1
status: open
deps: [com-77c4, com-qtjh]
links: []
created: 2026-03-01T13:17:51Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-3]
---
# comms-v1: comms daemon start CLI command

## Summary
Implement the `comms daemon start` subcommand. Runs the daemon in the foreground -- service managers (systemd, launchd) handle actual daemonization. Loads config, checks that no daemon is already running (via PID file + process check), then calls `daemon.Run`. Blocks until SIGTERM/SIGINT.

## Files
- `internal/cli/daemon.go` -- cobra command group `daemonCmd` with `startCmd` subcommand
- `internal/cli/daemon_test.go` -- test that start rejects when PID file indicates a running daemon

## Implementation Notes
- Dependencies: Task 3.1 (PID), Task 3.2 (daemon core loop)
- Creates a cobra command group `daemonCmd` that is added to the root command
- The `startCmd` is a subcommand of `daemonCmd`
- Before starting, check `daemon.IsRunning()` and reject if already running
- Runs in foreground (no fork/daemonize) -- blocks until signal received
- On duplicate start attempt, report error clearly

## Verification
```sh
cd /home/greg/p/comms && go build ./cmd/comms
cd /home/greg/p/comms && ./cmd/comms daemon start --help | grep -q 'start'
cd /home/greg/p/comms && go test ./internal/cli/... -run TestDaemonStart -v
```

## Acceptance Criteria
- Binary builds
- `daemon start --help` shows usage
- Tests pass (rejects duplicate start)

