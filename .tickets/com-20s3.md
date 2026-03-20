---
id: com-20s3
status: closed
deps: []
links: []
created: 2026-03-20T01:04:35Z
type: task
priority: 1
assignee: Greg Wedow
---
# Plugin Arch: Move package directory and update imports

## Summary
Move all files from `internal/provider/telegram/` to `providers/telegram/`. Update all import paths from `github.com/wedow/comms/internal/provider/telegram` to `github.com/wedow/comms/providers/telegram`. Use `git mv` to preserve history. Run `go mod tidy` after. Remove empty `internal/provider/` directory. No code changes -- only directory location and import paths.

## Files
- Move: all `.go` files in `internal/provider/telegram/` -> `providers/telegram/`
- Modify (import paths only): `internal/cli/root.go`, `internal/cli/send.go`, `internal/cli/react.go`, `internal/cli/daemon.go`, `internal/cli/send_test.go`, `internal/cli/react_test.go`

## Implementation Notes
- Use `git mv` to preserve history
- Run `go mod tidy` after move
- Remove empty `internal/provider/` directory
- No code changes -- only directory location and import paths

## Verification
```bash
go build ./...
go vet ./...
go test ./...
ls internal/provider/telegram/ 2>&1  # should fail
ls providers/telegram/
```

## Acceptance Criteria
- Telegram provider package lives at `providers/telegram/`
- All existing tests pass
- No behavior change

