---
id: com-mnvn
status: closed
deps: []
links: []
created: 2026-03-20T01:04:46Z
type: task
priority: 1
assignee: Greg Wedow
---
# Plugin Arch: Create protocol types

## Summary
Define all JSONL message types as Go structs in `internal/protocol/types.go`. Pure data types with JSON tags. No behavior.

Provider-to-daemon events: `ReadyEvent` (type, provider, version), `MessageEvent` (offset, id, chat_id, channel, from, date, body, reply_to, reply_to_body, quote, thread_id, media_type, media_file_id, download_url, media_ext, caption, forward_from, forward_date, edit_date, media_group_id, entities), `EditEvent` (type alias or wrapper for MessageEvent with type "edit"), `ReactionEvent` (offset, channel, message_id, from, emoji, date), `ResponseEvent` (id, ok, message (*MsgSummary), error), `ErrorEvent` (code, message), `ShutdownCompleteEvent` (just type field), `PingEvent`/`PongEvent` (type, ts RFC3339).

Daemon-to-provider commands: `StartCommand` (type, offset), `SendCommand` (type, id, chat_id, text, parse_mode, reply_to_id, thread_id), `SendMediaCommand` (type, id, chat_id, media_type, path, filename, caption, parse_mode, reply_to_id, thread_id), `ReactCommand` (type, id, chat_id, message_id, emoji), `TypingCommand` (type, chat_id), `ShutdownCommand` (type, reason).

Supporting types: `MsgSummary` (id, chat_id, channel, from, date, body), `Entity` (type, offset, length, url -- separate from message.Entity to avoid import coupling).

## Files
- Create: `internal/protocol/types.go`

## Implementation Notes
- Pure data types only, no behavior
- JSON tags on all fields
- Entity type is separate from message.Entity to avoid import coupling
- EditEvent can be a type alias or wrapper for MessageEvent with type "edit"

## Verification
```bash
go test ./internal/protocol/ -v
```

## Acceptance Criteria
- All protocol types defined as Go structs with JSON tags
- Package compiles

