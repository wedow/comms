---
id: com-3w5g
status: closed
deps: [com-6g4f, com-fe1f, com-6g26, com-0cej, com-qcy9, com-pqo8, com-sz5e, com-98be, com-ogzs]
links: []
created: 2026-03-20T01:10:07Z
type: task
priority: 3
assignee: Greg Wedow
---
# Plugin Arch: Update CI and build configuration

## Summary
Update .github/workflows/ CI to build both binaries. Add second go build step. Check if goreleaser config exists; if not, create .goreleaser.yaml with both build targets. Update brew formula install section for both binaries.

## Files
- Modify: .github/workflows/*.yml (add second go build step)
- Create (if absent): .goreleaser.yaml (both build targets)

## Implementation Notes
- CI must build both cmd/comms and cmd/comms-telegram
- GoReleaser needs two build targets: one for comms, one for comms-telegram
- Brew formula should install both binaries
- Check existing CI workflow files to understand current build setup

## Verification
```bash
go build -o /dev/null ./cmd/comms && go build -o /dev/null ./cmd/comms-telegram
```

## Acceptance Criteria
- CI builds both binaries
- Goreleaser configured for both targets

