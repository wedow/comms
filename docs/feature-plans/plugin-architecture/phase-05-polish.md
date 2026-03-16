# Phase 05: Polish

## Overview

Final cleanup: update documentation, verify CI configuration, remove deprecated config fields, and do a comprehensive end-to-end verification. This phase ensures the codebase is clean and ready for use.

## Tasks

### Task 05-1: Remove deprecated config fields

Remove the `TelegramConfig` struct and `Telegram` field from `Config`. Remove the `TelegramToken()` backward-compat helper. Remove the migration logic in `Load()` that copies `[telegram]` to `[providers]`.

All callers should now use `Providers["telegram"]` or `ProviderConfig("telegram")`.

**Files:**
- `internal/config/config.go` -- remove `TelegramConfig`, `Telegram` field, `TelegramToken()`, migration logic
- `internal/cli/daemon.go` -- verify no remaining references (should already use new API from Phase 03)

Update `Default()` to only produce `[providers.telegram]` format.

### Task 05-2: Update CI and build configuration

Add a `Makefile` (or update existing build script) with targets for both binaries. Update `.github/workflows/` CI if it exists, or document the build commands in CLAUDE.md.

```makefile
.PHONY: build
build:
	go build -o comms ./cmd/comms
	go build -o comms-telegram ./cmd/comms-telegram

.PHONY: test
test:
	go test ./...
	go vet ./...
```

If `.goreleaser.yml` exists, add `comms-telegram` as a second build target with the same build settings as `comms`.

**Files:**
- `Makefile` (new, or update existing)
- `.github/workflows/*.yml` (if exists -- add `make build` step)
- `.goreleaser.yml` (if exists -- add second binary)

**Verification:**
```bash
make build
make test
ls -la comms comms-telegram
```

### Task 05-3: Final end-to-end verification

Run the complete verification suite and fix any remaining issues:

```bash
# Build both binaries
go build -o /tmp/comms ./cmd/comms
go build -o /tmp/comms-telegram ./cmd/comms-telegram

# Full test suite
go test ./... -v
go vet ./...

# Verify zero telegram imports in base binary
grep -r "go-telegram" cmd/comms/ internal/cli/ internal/daemon/ internal/store/ internal/config/ internal/message/
# Expect: no output

# Verify telegram imports only in provider binary
grep -r "go-telegram" providers/telegram/ cmd/comms-telegram/
# Expect: matches (provider package needs telegram SDK)

# Verify both binaries work
/tmp/comms --help
/tmp/comms-telegram --help
/tmp/comms-telegram send --help
/tmp/comms-telegram react --help

# Verify protocol package is standalone
go test ./internal/protocol/ -v

# Verify config works with new format only
go test ./internal/config/ -v
```

Fix any issues found. Document any known limitations or open decisions.

## Verification

```bash
go test ./...
go vet ./...
go build -o /tmp/comms ./cmd/comms
go build -o /tmp/comms-telegram ./cmd/comms-telegram
```

## Dependencies

- Phase 00, 01, 02, 03, 04 all complete

## Post-Phase State

- Two-binary plugin architecture fully implemented and tested
- Zero deprecated config fields remain
- CI builds and tests both binaries
- All verification commands pass
