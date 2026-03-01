---
id: com-zl1e
status: open
deps: [com-qtjh, com-9oxg]
links: []
created: 2026-03-01T13:18:13Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-3]
---
# comms-v1: Callback execution

## Summary
When the daemon writes a new message, optionally execute a shell callback. The callback command comes from `config.callback.command`. It runs as `$SHELL -c "<command>"` (or `sh -c` if `$SHELL` is unset) with environment variables: `$COMMS_FILE` (absolute path to written file), `$COMMS_CHANNEL` (channel name), `$COMMS_PROVIDER` (e.g. "telegram"), `$COMMS_SENDER` (sender). Callback runs asynchronously -- the daemon does not block on it. Stdout/stderr of the callback are discarded.

## Files
- `internal/daemon/callback.go` -- `ExecCallback(command string, env CallbackEnv) error`; `CallbackEnv` struct has `File`, `Channel`, `Provider`, `Sender` string fields
- `internal/daemon/callback_test.go` -- test with a command that writes env vars to a temp file, verify env vars are set correctly

## Implementation Notes
- Dependencies: Task 3.2 (daemon core loop)
- Function signature: `ExecCallback(command string, env CallbackEnv) error`
- `CallbackEnv` struct:
  ```go
  type CallbackEnv struct {
      File     string
      Channel  string
      Provider string
      Sender   string
  }
  ```
- Environment variable mapping:
  - `COMMS_FILE` = `env.File` (absolute path to written file)
  - `COMMS_CHANNEL` = `env.Channel`
  - `COMMS_PROVIDER` = `env.Provider`
  - `COMMS_SENDER` = `env.Sender`
- Shell selection: use `$SHELL` env var, fall back to `sh` if unset
- Execution: `$SHELL -c "<command>"` with env vars set
- Async: callback runs in a goroutine, does not block the caller
- Stdout/stderr of callback are discarded

## Verification
```sh
cd /home/greg/p/comms && go test ./internal/daemon/... -run TestCallback -v
```

## Acceptance Criteria
- All tests pass
- Callback receives correct env vars
- Async execution does not block caller

