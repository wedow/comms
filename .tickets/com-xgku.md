---
id: com-xgku
status: open
deps: [com-plln, com-p48x]
links: []
created: 2026-03-20T01:05:12Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Refactor send.go to delegate

## Summary
Replace current send implementation with delegation logic: read `chat_id` from store, extract provider from channel, resolve `comms-<provider>` binary, forward explicit named flags as argv. Base binary retains store operations (ReadChatID, WriteMessage, WriteCursor) -- parses provider binary's JSON stdout and handles all store persistence. Delete `parseFormatFlag` from `send.go` entirely; `--format` flag is forwarded as raw string to provider binary. `delegate()` constructs explicit named argv -- NOT raw `os.Args` passthrough. Flags forwarded: `--channel`, `--format`, `--reply-to`, `--file`, `--media-type`, `--thread`, `--chat-id` (value from store via ReadChatID), and body args. Remove imports: `go-telegram/bot/models`, `providers/telegram`. Keep: `store`, `message`, `config`.

## Files
- Modify: `internal/cli/send.go`

## Implementation Notes
- Replace current send implementation with delegation logic
- Base binary retains store operations (ReadChatID, WriteMessage, WriteCursor)
- Parse provider binary's JSON stdout and handle all store persistence
- Delete `parseFormatFlag` from send.go entirely; --format flag is forwarded as raw string
- delegate() constructs explicit named argv -- NOT raw os.Args passthrough
- Flags forwarded: --channel, --format, --reply-to, --file, --media-type, --thread, --chat-id (value from store via ReadChatID), and body args
- Remove imports: go-telegram/bot/models, providers/telegram
- Keep imports: store, message, config

## Verification
```bash
go test ./internal/cli/ -v
grep -r "go-telegram/bot" internal/cli/send.go  # should produce no output
```

## Acceptance Criteria
- send.go delegates to provider binary
- Zero telegram SDK imports in send.go
- Store operations retained in base binary

