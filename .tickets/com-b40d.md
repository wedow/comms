---
id: com-b40d
status: open
deps: []
links: []
created: 2026-03-01T13:17:15Z
type: task
priority: 1
assignee: Greg Wedow
tags: [phase-1]
---
# comms-v1: Go project skeleton and Cobra CLI root

## Summary
Create the binary entrypoint and Cobra root command with `--version` flag. This establishes the build and gives every later task a place to wire into. The root command should have name "comms", a short description, and a version flag.

## Files
- `cmd/comms/main.go` -- calls `cli.Execute()`
- `internal/cli/root.go` -- Cobra root command (name "comms", short description, version flag)

## Implementation Notes
- Run `go get github.com/spf13/cobra` before building
- No other dependencies
- The root command name must be "comms"
- Must have a short description
- Must have a `--version` flag that prints a version string starting with "comms"

## Verification
```sh
go build -o /dev/null ./cmd/comms
go build ./cmd/comms && ./cmd/comms --version | grep -qE '^comms'
go build ./cmd/comms && ./cmd/comms --help | grep -q 'comms'
```

## Acceptance Criteria
- Binary builds without error
- `--version` prints a version string
- `--help` prints usage

