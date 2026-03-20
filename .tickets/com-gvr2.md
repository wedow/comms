---
id: com-gvr2
status: open
deps: [com-mnvn]
links: []
created: 2026-03-20T01:04:55Z
type: task
priority: 1
assignee: Greg Wedow
---
# Plugin Arch: Create protocol codec

## Summary
Implement JSONL encode/decode in `internal/protocol/codec.go`.

Functions:
- `Encode(w io.Writer, msg any) error` — marshal to JSON, write with trailing newline
- `Decode(r *bufio.Reader) (map[string]any, error)` — read one line, unmarshal to raw map
- `DecodeInto(r *bufio.Reader, target any) error` — read one line, unmarshal into target struct
- `DecodeTyped(r *bufio.Reader) (any, error)` — read one line, inspect "type" field, unmarshal into correct concrete struct (ReadyEvent, MessageEvent, EditEvent, ReactionEvent, ResponseEvent, ErrorEvent, ShutdownCompleteEvent, PingEvent, PongEvent, StartCommand, SendCommand, SendMediaCommand, ReactCommand, TypingCommand, ShutdownCommand); callers use Go type switches

1 MiB max line length check. Skip blank lines.

Note: `internal/daemon/protocol.go` from Phase 03 is NOT needed -- this codec handles all protocol encoding/decoding.

## Files
- Create: `internal/protocol/codec.go`

## Implementation Notes
- 1 MiB max line length
- Skip blank lines
- DecodeTyped inspects "type" field and returns correct concrete struct
- Callers use Go type switches on results

## Verification
```bash
go test ./internal/protocol/ -v
```

## Acceptance Criteria
- Encode/Decode/DecodeInto/DecodeTyped all work
- Max line length enforced
- Blank lines skipped

