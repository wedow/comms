# Phase 3: Daemon

**Dependencies:** Phase 1 (store, message, config), Phase 2 (telegram provider)

---

## Task 3.1: PID File Management

**Description:** Implement PID file helpers: write PID on daemon start, read PID for stop/status, check whether a PID is alive (signal 0), remove PID file on shutdown. PID file lives at `.comms/daemon.pid`. These are utility functions used by both the daemon itself and the CLI commands.

**Files:**
- `internal/daemon/pid.go` -- `WritePID(dir string) error`, `ReadPID(dir string) (int, error)`, `IsRunning(dir string) bool`, `RemovePID(dir string) error`
- `internal/daemon/pid_test.go` -- test write/read/remove cycle, test IsRunning with current process (alive) and stale pid (not alive)

**Dependencies:** none (stdlib only)

**Verification:**
```sh
cd /home/greg/p/comms && go test ./internal/daemon/... -run TestPID -v
```

**Pass/fail:** All tests pass. Write/read/remove cycle works. IsRunning returns true for current process, false for stale PID.

---

## Task 3.2: Daemon Core Loop

**Description:** Implement the main daemon loop in `internal/daemon`. The daemon loads config, initializes the Telegram provider, and polls for incoming messages. For each message received, it serializes via the message package and writes to disk via the store package. The daemon writes a PID file on startup and removes it on shutdown. Graceful shutdown on SIGTERM/SIGINT via `signal.NotifyContext`.

The poll handler receives both `message.Message` and `chatID int64` from `telegram.Poll`. For each message:
1. Write the message to disk via `store.WriteMessage`.
2. Write the chat ID via `store.WriteChatID` (Task 3.9) so `comms send` can resolve channel names later.
3. Fire the callback if configured (Task 3.6).

The daemon constructs the channel directory name as `msg.Provider + "-" + msg.Channel` (e.g., `telegram-general`) and passes that to store functions.

**Files:**
- `internal/daemon/daemon.go` -- `Run(ctx context.Context, cfg config.Config, root string) error`: write PID file, init provider, enter poll loop, remove PID file on exit
- `internal/daemon/daemon_test.go` -- test with a fake provider: verify messages are written to store, PID file created/removed

**Dependencies:** Task 3.1 (PID helpers), Phase 1 (config.Config, store.WriteMessage, message package), Phase 2 (telegram.Poll)

**Verification:**
```sh
cd /home/greg/p/comms && go build ./internal/daemon/...
cd /home/greg/p/comms && go test ./internal/daemon/... -run TestDaemon -v
```

**Pass/fail:** Package builds. Tests verify messages are written to store and PID file lifecycle is correct.

---

## Task 3.3: `comms daemon start` CLI Command

**Description:** Implement the `comms daemon start` subcommand. Runs the daemon in the foreground -- service managers (systemd, launchd) handle actual daemonization. Loads config, checks that no daemon is already running (via PID file + process check), then calls `daemon.Run`. Blocks until SIGTERM/SIGINT.

**Files:**
- `internal/cli/daemon.go` -- cobra command group `daemonCmd` with `startCmd` subcommand
- `internal/cli/daemon_test.go` -- test that start rejects when PID file indicates a running daemon

**Dependencies:** Task 3.1 (PID helpers), Task 3.2 (daemon.Run)

**Verification:**
```sh
cd /home/greg/p/comms && go build ./cmd/comms
cd /home/greg/p/comms && ./cmd/comms daemon start --help | grep -q 'start'
cd /home/greg/p/comms && go test ./internal/cli/... -run TestDaemonStart -v
```

**Pass/fail:** Binary builds. `daemon start --help` shows usage. Tests pass (rejects duplicate start).

---

## Task 3.4: `comms daemon stop` CLI Command

**Description:** Implement `comms daemon stop`. Reads the PID file, verifies the process is alive, sends SIGTERM, waits briefly for the process to exit, then confirms. If no PID file or process is not running, report that clearly as JSON on stderr.

**Files:**
- `internal/cli/daemon.go` -- add `stopCmd` subcommand to `daemonCmd` group
- `internal/cli/daemon_test.go` -- test stop when no daemon running (error), test stop with stale PID file (cleanup + error)

**Dependencies:** Task 3.1 (PID helpers), Task 3.3 (daemon command group must exist)

**Verification:**
```sh
cd /home/greg/p/comms && go build ./cmd/comms
cd /home/greg/p/comms && ./cmd/comms daemon stop --help | grep -q 'stop'
cd /home/greg/p/comms && go test ./internal/cli/... -run TestDaemonStop -v
```

**Pass/fail:** Tests pass. Stop with no daemon reports JSON error. Stop with stale PID cleans up.

---

## Task 3.5: `comms daemon status` CLI Command

**Description:** Implement `comms daemon status`. Reads PID file, checks whether the process is alive. Outputs JSON to stdout: `{"running": true, "pid": 12345}` or `{"running": false}`. If PID file exists but process is dead, clean up the stale PID file and report not running.

**Files:**
- `internal/cli/daemon.go` -- add `statusCmd` subcommand to `daemonCmd` group
- `internal/cli/daemon_test.go` -- test status with no PID file, test status with stale PID file

**Dependencies:** Task 3.1 (PID helpers)

**Verification:**
```sh
cd /home/greg/p/comms && go build ./cmd/comms
cd /home/greg/p/comms && ./cmd/comms daemon status --help | grep -q 'status'
cd /home/greg/p/comms && go test ./internal/cli/... -run TestDaemonStatus -v
```

**Pass/fail:** Tests pass. Status output is valid JSON with `running` boolean field. Stale PID file is cleaned.

---

## Task 3.6: Callback Execution

**Description:** When the daemon writes a new message, optionally execute a shell callback. The callback command comes from `config.callback.command`. It runs as `$SHELL -c "<command>"` (or `sh -c` if `$SHELL` is unset) with environment variables: `$COMMS_FILE` (absolute path to written file), `$COMMS_CHANNEL` (channel name), `$COMMS_PROVIDER` (e.g. "telegram"), `$COMMS_SENDER` (sender). Callback runs asynchronously -- the daemon does not block on it. Stdout/stderr of the callback are discarded.

**Files:**
- `internal/daemon/callback.go` -- `ExecCallback(command string, env CallbackEnv) error`; `CallbackEnv` struct has `File`, `Channel`, `Provider`, `Sender` string fields
- `internal/daemon/callback_test.go` -- test with a command that writes env vars to a temp file, verify env vars are set correctly

**Dependencies:** Task 3.2 (integration point -- daemon calls callback after writing message)

**Verification:**
```sh
cd /home/greg/p/comms && go test ./internal/daemon/... -run TestCallback -v
```

**Pass/fail:** All tests pass. Callback receives correct env vars. Async execution does not block caller.

---

## Task 3.7: Callback Rate Limiting

**Description:** Enforce minimum interval between callback invocations using the `config.callback.delay` value (e.g. "5s"). If messages arrive faster than the delay, callbacks are skipped (not queued). Simple time-based gate: store the last callback invocation time, skip if less than `delay` has elapsed. Global scope (not per-channel).

**Files:**
- `internal/daemon/callback.go` -- add `CallbackRunner` struct with `lastRun time.Time` and `delay time.Duration`; method `Run(command string, env CallbackEnv)` that checks delay before executing
- `internal/daemon/callback_test.go` -- test that rapid invocations are throttled, test that invocations after delay pass through

**Dependencies:** Task 3.6

**Verification:**
```sh
cd /home/greg/p/comms && go test ./internal/daemon/... -run TestCallbackRate -v
```

**Pass/fail:** Rapid invocations are throttled. Invocations after delay proceed.

---

## Task 3.8: Telegram Polling Offset Persistence

**Description:** Persist the Telegram polling offset across daemon restarts so the daemon does not re-process old messages. Store the offset in `.comms/telegram.offset` as a plain integer. The daemon reads it on startup and passes it to `telegram.Poll`. On shutdown, the daemon writes the returned offset back.

**Files:**
- `internal/store/offset.go` -- `ReadOffset(root, provider string) (int64, error)`, `WriteOffset(root, provider string, offset int64) error`. File path: `root/<provider>.offset`.
- `internal/store/offset_test.go` -- test read/write cycle, test missing file returns 0
- `internal/daemon/daemon.go` -- integrate offset read on startup, write on shutdown (modify existing file from Task 3.2)

**Dependencies:** Task 3.2 (daemon core loop), Phase 1 store package

**Verification:**
```sh
cd /home/greg/p/comms && go test ./internal/store/... -run TestOffset -v
cd /home/greg/p/comms && go test ./internal/daemon/... -v
```

**Pass/fail:** Offset round-trips correctly. Missing file returns 0. Daemon integrates offset on startup/shutdown.

---

## Task 3.9: Chat ID Persistence

**Description:** When the daemon receives a message from a new chat, it needs to persist the chat ID so that `comms send` (Phase 4, Task 4.4) can resolve channel names to Telegram chat IDs. Store the chat ID in `<channel>/.chat_id` as a plain integer.

The daemon's poll handler already receives `chatID int64` from `telegram.Poll` (see Phase 2, Task 2.4). After writing the message file, the daemon calls `store.WriteChatID` to persist the mapping.

**Files:**
- `internal/store/chatid.go` -- `ReadChatID(root, channel string) (int64, error)`, `WriteChatID(root, channel string, chatID int64) error`. File path: `root/<channel>/.chat_id`.
- `internal/store/chatid_test.go` -- test read/write cycle, test missing file returns error

**Dependencies:** Phase 1 store package (directory structure)

**Verification:**
```sh
cd /home/greg/p/comms && go test ./internal/store/... -run TestChatID -v
```

**Pass/fail:** Chat ID round-trips correctly. Missing file returns a clear error.

---

## Task 3.10: Integration Test -- Full Daemon Lifecycle

**Description:** End-to-end test that exercises the full daemon lifecycle: start daemon with a mock Telegram provider, receive messages, verify files written to `.comms/`, verify callback fired, stop daemon via context cancellation, verify PID file cleaned up, verify polling offset persisted.

**Files:**
- `internal/daemon/daemon_integration_test.go` -- uses a fake provider, temp directory for `.comms/`, and a callback script that touches a marker file

**Dependencies:** All prior tasks in this phase (3.1-3.9)

**Verification:**
```sh
cd /home/greg/p/comms && go test ./internal/daemon/... -v -run TestIntegration
```

**Pass/fail:** Integration test passes. Messages written to disk. Callback fired. PID file cleaned. Offset persisted.

---

## Task Dependency Graph

```
Task 3.1 (PID helpers) -- no deps
  |
  +---> Task 3.2 (core loop) -- depends on 3.1, Phase 1, Phase 2
  |       |
  |       +---> Task 3.3 (daemon start CLI) -- depends on 3.1, 3.2
  |       |       |
  |       |       +---> Task 3.4 (daemon stop CLI) -- depends on 3.1, 3.3
  |       |
  |       +---> Task 3.6 (callback exec) -- depends on 3.2
  |       |       |
  |       |       +---> Task 3.7 (callback rate limit) -- depends on 3.6
  |       |
  |       +---> Task 3.8 (offset persistence) -- depends on 3.2
  |
  +---> Task 3.5 (daemon status CLI) -- depends on 3.1 only

Task 3.9 (chat ID persistence) -- depends on Phase 1 store only

Task 3.10 (integration test) -- depends on all above (3.1-3.9)
```

Parallelizable: (3.3 + 3.5 + 3.6 + 3.8 + 3.9) after their respective deps.

## Phase Exit Criteria

```sh
cd /home/greg/p/comms && go build ./cmd/comms
cd /home/greg/p/comms && go test ./internal/daemon/... -v
cd /home/greg/p/comms && go test ./internal/store/... -v
cd /home/greg/p/comms && go vet ./...
```

All commands exit 0.
