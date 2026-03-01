---
id: com-alwu
status: open
deps: [com-b40d, com-9oxg, com-6qs4, com-m4l2]
links: []
created: 2026-03-01T13:18:20Z
type: task
priority: 1
assignee: Greg Wedow
tags: [phase-1]
---
# comms-v1: comms init command

## Summary
Implement the `init` subcommand. It scaffolds the `.comms/` directory tree, writes a default `config.toml`, and deploys embedded docs.

Behavior:
1. Resolve root as `.comms/` in the current working directory (or a `--dir` flag).
2. Call `store.InitDir(root)`.
3. Write `config.Default()` as TOML to `root/config.toml`. Skip if file already exists (don't overwrite user config).
4. Write `embeddocs.TelegramSetupDoc` to `root/docs/telegram-setup.md`. Always overwrite (docs may update between versions).
5. Print JSON `{"status":"initialized","path":"<abs path>"}` to stdout.

Wire the subcommand into the Cobra root command.

## Files
- `internal/cli/init.go` -- init command definition and handler
- `internal/cli/init_test.go` -- test: init creates expected directory tree, config.toml exists with defaults, docs deployed, second init doesn't clobber config

## Implementation Notes
- Depends on Tasks 1.1 (CLI root), 1.4 (config), 1.5 (store), 1.7 (embeddocs)
- `--dir` flag to override default `.comms/` directory
- Must NOT overwrite existing `config.toml` (skip if exists)
- Must ALWAYS overwrite docs (they may update between versions)
- Output must be JSON to stdout: `{"status":"initialized","path":"<abs path>"}`
- Use `config.Default()` to generate default config
- Use `github.com/BurntSushi/toml` encoder to write config as TOML
- Wire into Cobra root command from Task 1.1
- Test that second init doesn't clobber modified config

## Verification
```sh
go test ./internal/cli/... -run TestInit -v
go build ./cmd/comms && ./cmd/comms init --help | grep -q 'init'
```

## Acceptance Criteria
- All tests pass
- `init --help` shows usage
- Init creates `.comms/` tree with config and docs

