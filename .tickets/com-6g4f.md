---
id: com-6g4f
status: open
deps: [com-68x1, com-ab0w, com-oqd8, com-jrwi]
links: []
created: 2026-03-20T01:09:28Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Full build and test verification

## Summary
Run go mod tidy. Build both binaries. Run all tests. Verify CI commands pass.

## Files
- Modify: go.mod, go.sum (auto-modified by go mod tidy)

## Implementation Notes
- This is a verification task that may require fixing any remaining issues
- Both comms and comms-telegram binaries must build cleanly
- All test suites must pass
- go vet must be clean

## Verification
```bash
go test ./providers/telegram/ -run TestSubprocess -v -count=1
go test ./providers/telegram/ -run TestIntegration -v -count=1
go test ./providers/telegram/ -v
go test ./internal/protocol/ -v
go test ./internal/daemon/ -v
go build -o /tmp/comms ./cmd/comms
go build -o /tmp/comms-telegram ./cmd/comms-telegram
/tmp/comms-telegram --version
/tmp/comms-telegram send --help
/tmp/comms-telegram react --help
go test ./...
go vet ./...
```

## Acceptance Criteria
- Both binaries build
- All tests pass
- go vet clean

