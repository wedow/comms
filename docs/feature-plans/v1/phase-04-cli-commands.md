# Phase 4: CLI Commands

## Overview

Implement the four remaining CLI commands: `comms channels`, `comms list`, `comms unread`, and `comms send`. All commands emit JSON lines to stdout, structured JSON errors to stderr, and use exit code 0 for success, 1 for errors. No interactive prompts, no color codes.

**Dependencies:** Phase 1 (store, message, config packages), Phase 2 (telegram provider -- needed only for `send`), Phase 3 Task 3.9 (store.ReadChatID -- needed for `send`).

---

## Task 4.1: Shared JSON output helpers

**Description:** Establish the shared patterns for JSON line output and structured error output. This avoids duplicating serialization logic across four commands.

Two small helpers:
- `cli.PrintJSON(w io.Writer, v any) error` -- marshal `v` as JSON, write one line to `w`
- `cli.PrintError(msg string, args ...any)` -- write `{"error":"..."}` to stderr, used before `os.Exit(1)`

These are functions, not a framework. Each is 3-5 lines.

**Files:**
- `internal/cli/output.go` -- the two helper functions
- `internal/cli/output_test.go` -- verify JSON format, newline termination

**Dependencies:** none (stdlib only)

**Verification:**
```sh
cd /home/greg/p/comms && go test ./internal/cli/ -run TestOutput -v
```

**Pass/fail:** JSON output is valid, newline-terminated. Error format is `{"error":"..."}`.

---

## Task 4.2: `comms channels` command

**Description:** List all known channel directories under `.comms/` as JSON lines. Each line is a JSON object with the channel name and provider extracted from the directory name (`<provider>-<rest>` convention).

Output per line:
```json
{"name":"telegram-general","provider":"telegram","path":".comms/telegram-general"}
```

Channels are discovered by `store.ListChannels()` (Phase 1, Task 1.6).

**Files:**
- `internal/cli/channels.go` -- command definition, flag parsing, JSON output
- `internal/cli/channels_test.go` -- tests against a temp `.comms/` directory

**Dependencies:** Task 4.1 (output helpers), Phase 1 Task 1.6 (store.ListChannels)

**Verification:**
```sh
cd /home/greg/p/comms && go test ./internal/cli/ -run TestChannels -v
```

**Pass/fail:** Tests pass. Output is valid JSON lines. Each line has `name`, `provider`, `path` fields.

---

## Task 4.3: `comms list` command

**Description:** List all messages across all channels (or a specific channel with `--channel`). No cursor side effects -- read-only view of the full message history. Each message is printed as one JSON line.

Flags:
- `--channel <name>` -- optional, filter to a single channel directory

`--channel` takes the full channel directory name as shown by `comms channels` (e.g., `telegram-general`).

Output per line (all fields from `message.Message` struct plus file path):
```json
{"from":"alice","provider":"telegram","channel":"general","date":"2026-02-24T14:30:05Z","id":"telegram-98765","body":"Hey, the deploy finished.","file":".comms/telegram-general/2026-02-24T14-30-05.123.md"}
```

Flow:
1. Call `store.ListChannels()` (or filter to `--channel`)
2. For each channel, call `store.ListMessages(channel)` to get message file paths sorted chronologically
3. For each file, call `store.ReadMessage()` to parse the message (format auto-detected from file extension)
4. Serialize to JSON line and print to stdout

**Files:**
- `internal/cli/list.go` -- command definition
- `internal/cli/list_test.go` -- tests with pre-populated temp `.comms/` directory

**Test cases:**
- List all channels when no `--channel` flag
- List single channel with `--channel`
- Empty channel directory produces no output (exit 0)
- Invalid `--channel` name produces JSON error on stderr, exit 1

**Dependencies:** Task 4.1 (output helpers), Phase 1 Tasks 1.5-1.6 (store.ReadMessage, store.ListChannels, store.ListMessages). No config dependency -- `store.ReadMessage` auto-detects format from file extension.

**Verification:**
```sh
cd /home/greg/p/comms && go test ./internal/cli/ -run TestList -v
```

**Pass/fail:** All test cases pass. Output is valid JSON lines with correct fields.

---

## Task 4.4: `comms unread` command

**Description:** List unread messages across all channels (or filtered by `--channel`), then advance each channel's cursor to the newest message's timestamp. This is the primary command agents use to poll for new messages.

Flags:
- `--channel <name>` -- optional, filter to a single channel

`--channel` takes the full channel directory name as shown by `comms channels` (e.g., `telegram-general`).

Flow:
1. Call `store.ListChannels()` (or filter to `--channel`)
2. For each channel:
   a. Call `store.ReadCursor(channel)` to get the last-read timestamp
   b. Call `store.ListMessagesAfter(channel, cursor)` to get files newer than cursor
   c. For each file, `store.ReadMessage()` and print as JSON line
   d. If any messages were found, call `store.WriteCursor(channel, newestTimestamp)` to advance cursor
3. Exit 0 even if no unread messages (empty output is valid)

Output format: same JSON line format as `comms list`.

Cursor advancement happens per-channel, after all messages for that channel have been printed.

**Files:**
- `internal/cli/unread.go` -- command definition
- `internal/cli/unread_test.go` -- tests

**Test cases:**
- No cursor file exists: all messages returned, cursor created
- Cursor exists: only newer messages returned, cursor advanced
- No new messages: no output, cursor unchanged
- `--channel` filters to one channel only
- Running `unread` twice in succession: second run produces no output
- Multiple channels: each cursor advances independently

**Dependencies:** Task 4.1 (output helpers), Phase 1 Tasks 1.5-1.6 (store.ReadCursor, store.WriteCursor, store.ListMessagesAfter, store.ListChannels, store.ReadMessage). No config dependency -- `store.ReadMessage` auto-detects format from file extension.

**Verification:**
```sh
cd /home/greg/p/comms && go test ./internal/cli/ -run TestUnread -v
```

**Pass/fail:** All test cases pass. Cursors advance correctly. Second run produces no output.

---

## Task 4.5: `comms send` command

**Description:** Minimal send command. Reads message body from stdin, sends via the telegram provider.

Flags:
- `--channel <name>` -- required, the channel to send to

`--channel` takes the full channel directory name as shown by `comms channels` (e.g., `telegram-general`).

Flow:
1. Load config to get telegram token
2. Read all of stdin as the message body (trimmed)
3. If body is empty, emit JSON error to stderr, exit 1
4. Resolve channel name to a Telegram chat ID via `store.ReadChatID(root, channel)` (Phase 3, Task 3.9). The daemon writes `.chat_id` files when it discovers channels during polling.
5. Call `telegram.Send(ctx, api, chatID, body)` (Phase 2, Task 2.3)
6. Print result as JSON line: `{"ok":true,"channel":"<name>"}`

Error cases:
- No stdin / empty body: JSON error on stderr, exit 1
- Channel not found / no `.chat_id` file: JSON error on stderr, exit 1
- Telegram API error: JSON error on stderr, exit 1

**Files:**
- `internal/cli/send.go` -- command definition
- `internal/cli/send_test.go` -- tests (mock the telegram provider via BotAPI interface)

**Test cases:**
- Successful send with mock: verify JSON success output
- Empty stdin: verify JSON error
- Missing channel: verify JSON error
- Missing `.chat_id` file: verify JSON error

**Dependencies:** Task 4.1 (output helpers), Phase 1 Task 1.4 (config.Load), Phase 2 Tasks 2.1+2.3 (BotAPI interface, Send function), Phase 3 Task 3.9 (store.ReadChatID)

**Verification:**
```sh
cd /home/greg/p/comms && go test ./internal/cli/ -run TestSend -v
```

**Pass/fail:** All test cases pass. JSON output format is correct. Errors emit to stderr.

---

## Task 4.6: Wire commands to root Cobra command

**Description:** Register `channels`, `list`, `unread`, and `send` as subcommands on the root Cobra command. Each command's `cobra.Command` is constructed by a function in its respective `internal/cli/*.go` file (e.g., `cli.NewChannelsCmd()`) and added to root.

**Files:**
- `internal/cli/root.go` -- add `rootCmd.AddCommand(...)` calls (modify existing file from Phase 1 Task 1.1)

**Dependencies:** Tasks 4.2-4.5 (all four commands implemented), Phase 1 Task 1.1 (root command)

**Verification:**
```sh
cd /home/greg/p/comms && go build ./cmd/comms && ./cmd/comms --help | grep -E 'channels|list|unread|send'
```

**Pass/fail:** `--help` output lists all four commands.

---

## Task Dependency Graph

```
Task 4.1 (output helpers) -- no deps
  |
  +---> Task 4.2 (channels)
  +---> Task 4.3 (list)
  +---> Task 4.4 (unread)
  +---> Task 4.5 (send) -- also depends on Phase 2 (telegram.Send) + Phase 3 Task 3.9 (ReadChatID)
         |
         v
Task 4.6 (wire to root) -- depends on 4.2, 4.3, 4.4, 4.5
```

Tasks 4.2, 4.3, 4.4, 4.5 can proceed in parallel after 4.1. Natural order is 4.2 -> 4.3 -> 4.4 -> 4.5 (each builds on patterns from the previous).

## Exit Code Convention

| Code | Meaning |
|------|---------|
| 0 | Success (including "no unread messages" -- empty output is valid) |
| 1 | Error (config not found, channel not found, API failure, etc.) |

## JSON Error Format

All errors written to stderr as a single JSON line:
```json
{"error":"channel not found: telegram-nonexistent"}
```
