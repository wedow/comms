---
title: "Phase 00: Provider Move"
status: reviewing
---

# Phase 00: Provider Move

## Overview

Mechanical directory move of the telegram provider package from `internal/provider/telegram/` to `providers/telegram/`. No behavior change -- only import paths are updated. This establishes the directory layout that all subsequent phases build on.

## Tasks

### Task 00-1: Move package directory and update imports

Move all files from `internal/provider/telegram/` to `providers/telegram/`. Update all import paths from `github.com/wedow/comms/internal/provider/telegram` to `github.com/wedow/comms/providers/telegram`.

**Files to move:**
- All `.go` files in `internal/provider/telegram/` -> `providers/telegram/`

**Files to modify (import paths only):**
- `internal/cli/root.go`
- `internal/cli/send.go`
- `internal/cli/react.go`
- `internal/cli/daemon.go`
- `internal/cli/send_test.go`
- `internal/cli/react_test.go`

**No code changes.** Only directory location and import paths.

**Implementation notes:**
- Use `git mv` to preserve history
- Run `go mod tidy` after (should be a no-op -- same module)
- Remove empty `internal/provider/` directory

## Verification

```bash
go build ./...
go vet ./...
go test ./...
# Verify old directory is gone
ls internal/provider/telegram/ 2>&1  # should fail
# Verify new directory exists
ls providers/telegram/
```

## Dependencies

None. This is the first phase.

## Post-Phase State

The telegram provider package lives at `providers/telegram/`. All existing tests pass. No behavior has changed.
