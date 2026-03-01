---
id: com-77c4
status: open
deps: []
links: []
created: 2026-03-01T13:17:30Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-3]
---
# comms-v1: PID file management

## Summary
Implement PID file helpers: write PID on daemon start, read PID for stop/status, check whether a PID is alive (signal 0), remove PID file on shutdown. PID file lives at `.comms/daemon.pid`. These are utility functions used by both the daemon itself and the CLI commands.

## Files
- `internal/daemon/pid.go` -- `WritePID(dir string) error`, `ReadPID(dir string) (int, error)`, `IsRunning(dir string) bool`, `RemovePID(dir string) error`
- `internal/daemon/pid_test.go` -- test write/read/remove cycle, test IsRunning with current process (alive) and stale pid (not alive)

## Implementation Notes
- Dependencies: none (stdlib only)
- Function signatures:
  - `WritePID(dir string) error` -- writes current process PID to `dir/daemon.pid`
  - `ReadPID(dir string) (int, error)` -- reads PID integer from `dir/daemon.pid`
  - `IsRunning(dir string) bool` -- reads PID file and sends signal 0 to check if process is alive
  - `RemovePID(dir string) error` -- removes `dir/daemon.pid`
- PID file path: `<dir>/daemon.pid`

## Verification
```sh
cd /home/greg/p/comms && go test ./internal/daemon/... -run TestPID -v
```

## Acceptance Criteria
- All tests pass
- Write/read/remove cycle works
- IsRunning returns true for current process, false for stale PID

