---
id: com-ryd3
status: open
deps: [com-b40d]
links: []
created: 2026-03-01T13:17:43Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-5]
---
# comms-v1: GoReleaser configuration

## Summary
Configure GoReleaser to produce multi-arch archives with checksums. Targets: macOS arm64, macOS amd64, Linux x86_64. Archives are tar.gz with the binary, LICENSE, and README.

Configuration details:
- `builds`: single build entry, `main: ./cmd/comms`, GOOS/GOARCH matrix: `darwin/arm64`, `darwin/amd64`, `linux/amd64`
- `archives`: tar.gz format, include `LICENSE*` and `README*`
- `checksum`: sha256 algorithm
- `release`: GitHub release with auto-generated changelog
- `snapshot`: name template with git short hash for non-tag builds
- No Docker images, no nfpm

## Files
- `.goreleaser.yaml` (create new)

## Implementation Notes
Minimal GoReleaser config. Single build targeting `./cmd/comms`. Three OS/arch combos in the matrix. Use tar.gz archives. sha256 checksums. Snapshot name template should include `{{ .ShortCommit }}` for non-tag builds.

## Verification
```sh
cd /home/greg/p/comms && goreleaser check
cd /home/greg/p/comms && goreleaser build --snapshot --clean
```

## Acceptance Criteria
`goreleaser check` exits 0. `goreleaser build --snapshot --clean` produces binaries for all three targets (darwin/arm64, darwin/amd64, linux/amd64).

