---
id: com-gb3w
status: closed
deps: []
links: []
created: 2026-03-01T13:17:55Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-4]
---
# comms-v1: Shared JSON output helpers

## Summary
Establish shared patterns for JSON line output and structured error output. This avoids duplicating serialization logic across the four CLI commands (channels, list, unread, send).

Two small helpers (functions, not a framework; each is 3-5 lines):
- `cli.PrintJSON(w io.Writer, v any) error` -- marshal `v` as JSON, write one line to `w`
- `cli.PrintError(msg string, args ...any)` -- write `{"error":"..."}` to stderr, used before `os.Exit(1)`

## Files
- `internal/cli/output.go` -- the two helper functions
- `internal/cli/output_test.go` -- verify JSON format, newline termination

## Implementation Notes
- Dependencies: none (stdlib only)
- `PrintJSON` should marshal `v` with `json.Marshal`, write the result to `w`, and append a newline
- `PrintError` should format the message with `fmt.Sprintf`, marshal `{"error": "<formatted msg>"}` as JSON, and write to `os.Stderr`
- Function signatures:
  ```go
  func PrintJSON(w io.Writer, v any) error
  func PrintError(msg string, args ...any)
  ```
- Keep these minimal -- no error wrapping frameworks, no custom types

## Verification
```sh
cd /home/greg/p/comms && go test ./internal/cli/ -run TestOutput -v
```

## Acceptance Criteria
- JSON output is valid, newline-terminated
- Error format is `{"error":"..."}`
- Tests pass

