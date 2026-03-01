---
id: com-2alb
status: open
deps: [com-ryd3]
links: []
created: 2026-03-01T13:18:07Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-5]
---
# comms-v1: Homebrew tap

## Summary
Configure GoReleaser to publish to the existing Homebrew tap at `wedow/homebrew-tools`. GoReleaser auto-updates the formula on release -- no manual publish step needed. The CI release job (Task 5.2) handles this automatically when a `v*` tag is pushed.

Add the following `brews` section to the existing `.goreleaser.yaml`:
```yaml
brews:
  - repository:
      owner: wedow
      name: homebrew-tools
    directory: Formula
    homepage: https://github.com/wedow/comms
    description: Filesystem-based messaging interface for AI agents
    license: MIT
    install: |
      bin.install "comms"
    service: |
      run [opt_bin/"comms", "daemon", "start"]
      keep_alive true
      working_dir Dir.home
      log_path var/"log/comms.log"
      error_log_path var/"log/comms.log"
```

The `service` block generates a launchd plist automatically -- Homebrew translates it into `~/Library/LaunchAgents/homebrew.mxcl.comms.plist`.

## Files
- `.goreleaser.yaml` (modify existing file from Task 5.1 -- add `brews` section)

## Implementation Notes
This task MODIFIES the `.goreleaser.yaml` created in Task 5.1. Append the `brews` section to the existing config. The repository is `wedow/homebrew-tools`, formula goes in the `Formula/` directory. The service block in the formula auto-generates a launchd plist for macOS users.

## Verification
```sh
cd /home/greg/p/comms && goreleaser check
```

## Acceptance Criteria
`goreleaser check` exits 0 after adding the brews section.

