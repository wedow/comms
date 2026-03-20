---
id: com-265l
status: closed
deps: [com-mnvn]
links: []
created: 2026-03-20T01:04:56Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Protocol event to message.Message conversion

## Summary
Create `protocolToMessage(evt protocol.MessageEvent) message.Message` and `protocolToReaction(evt protocol.ReactionEvent)` conversion functions. Pure conversion, no side effects.

## Files
- Create: `internal/daemon/convert.go`

## Implementation Notes
- Pure data transformation functions
- Map all protocol event fields to corresponding message.Message fields
- No I/O or side effects
- protocolToMessage maps MessageEvent fields to message.Message
- protocolToReaction maps ReactionEvent fields

## Verification
```bash
go test ./internal/daemon/ -v
```

## Acceptance Criteria
- Conversion functions correctly map all protocol event fields to message types.

