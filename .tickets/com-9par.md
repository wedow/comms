---
id: com-9par
status: open
deps: [com-yvif]
links: []
created: 2026-03-01T13:18:04Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-4]
---
# comms-v1: comms channels command

## Summary
Implement `comms channels` CLI command. Lists all known channel directories under `.comms/` as JSON lines. Each line is a JSON object with the channel name and provider extracted from the directory name (`<provider>-<rest>` convention).

Output per line:
```json
{"name":"telegram-general","provider":"telegram","path":".comms/telegram-general"}
```

Channels are discovered by `store.ListChannels()` (from Phase 1, Task 1.6).

## Files
- `internal/cli/output.go` -- shared JSON output helpers (PrintJSON, PrintError)
- `internal/cli/output_test.go` -- verify JSON format, newline termination
- `internal/cli/channels.go` -- command definition, flag parsing, JSON output
- `internal/cli/channels_test.go` -- tests against a temp `.comms/` directory

## Implementation Notes

### Shared JSON output helpers (build these first)
Two small helpers (functions, not a framework; each is 3-5 lines):
- `cli.PrintJSON(w io.Writer, v any) error` -- marshal `v` as JSON, write one line to `w`, append newline
- `cli.PrintError(msg string, args ...any)` -- format with `fmt.Sprintf`, marshal `{"error":"<formatted msg>"}` as JSON, write to `os.Stderr`
- Dependencies: none (stdlib only)
- Keep these minimal -- no error wrapping frameworks, no custom types

### Channels command
- The command should call `store.ListChannels()` to discover channel directories
- For each channel, extract the provider from the directory name using `<provider>-<rest>` convention (split on first `-`)
- Use `cli.PrintJSON()` from Task 4.1 to output each channel as a JSON line to stdout
- The command constructor should follow the pattern `cli.NewChannelsCmd() *cobra.Command`
- Test against a temp `.comms/` directory with some channel subdirectories

## Verification
```sh
cd /home/greg/p/comms && go test ./internal/cli/ -run TestChannels -v
```

## Acceptance Criteria
- Tests pass
- Output is valid JSON lines
- Each line has `name`, `provider`, `path` fields

