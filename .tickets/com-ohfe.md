---
id: com-ohfe
status: closed
deps: [com-gvr2, com-ho3r]
links: []
created: 2026-03-20T01:05:37Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Protocol I/O layer

## Summary
The daemon imports `internal/protocol` directly and calls `protocol.DecodeTyped(*bufio.Reader)` to read events from subprocess stdout pipe. Commands encoded using `json.NewEncoder(stdin).Encode(cmd)` where `cmd` is a concrete `protocol` type. No wrapper file or intermediate types needed -- callers use Go type switches on concrete types from `internal/protocol`. No new file created; read/write calls live in `internal/daemon/provider.go` alongside Spawn.

## Files
- Modify: `internal/daemon/provider.go`

## Implementation Notes
- Import `internal/protocol` directly in provider.go
- Use `protocol.DecodeTyped(*bufio.Reader)` for reading events from stdout pipe
- Use `protocol.Encode(stdin, cmd)` for writing commands to stdin pipe
- No wrapper types or intermediate layer
- Callers use Go type switches on concrete protocol types
- This wires up the protocol codec into the subprocess infrastructure

## Verification
```bash
go test ./internal/daemon/ -run TestProtocol -v
```

## Acceptance Criteria
- Protocol I/O uses internal/protocol directly. No wrapper types.

