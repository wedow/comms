---
id: com-0q0l
status: closed
deps: [com-77c4]
links: []
created: 2026-03-01T13:18:03Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-3]
---
# comms-v1: comms daemon status CLI command

## Summary
Implement `comms daemon status`. Reads PID file, checks whether the process is alive. Outputs JSON to stdout: `{"running": true, "pid": 12345}` or `{"running": false}`. If PID file exists but process is dead, clean up the stale PID file and report not running.

## Files
- `internal/cli/daemon.go` -- add `statusCmd` subcommand to `daemonCmd` group
- `internal/cli/daemon_test.go` -- test status with no PID file, test status with stale PID file

## Implementation Notes
- Dependencies: Task 3.1 (PID)
- The `statusCmd` is added as a subcommand of the existing `daemonCmd` group
- Output format (stdout, valid JSON):
  - Running: `{"running": true, "pid": 12345}`
  - Not running: `{"running": false}`
- If PID file exists but process is dead: remove stale PID file, report `{"running": false}`
- If no PID file: report `{"running": false}`

## Verification
```sh
cd /home/greg/p/comms && go build ./cmd/comms
cd /home/greg/p/comms && ./cmd/comms daemon status --help | grep -q 'status'
cd /home/greg/p/comms && go test ./internal/cli/... -run TestDaemonStatus -v
```

## Acceptance Criteria
- Tests pass
- Status output is valid JSON with `running` boolean field
- Stale PID file is cleaned

