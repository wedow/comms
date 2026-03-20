---
id: com-ab0w
status: open
deps: [com-20s3, com-mnvn, com-gvr2]
links: []
created: 2026-03-20T01:08:31Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Implement subprocess JSONL loop

## Summary
Implement RunSubprocess(ctx, stdin, stdout, cfg) error in providers/telegram/subprocess.go. Add SendChatAction(ctx context.Context, params *bot.SendChatActionParams) (bool, error) to BotAPI interface in telegram.go, and add no-op stub for it to mockBot in existing tests.

RunSubprocess flow:
1. Handshake: create bot client from config, send {"type":"ready","provider":"telegram","version":"1"} to stdout
2. Read loop: read JSONL from stdin, dispatch by type:
   - start: start long-poll loop in goroutine, poll handler writes events to stdout
   - send: call telegram.Send(), write response
   - send_media: open file, call telegram.SendMedia(), write response
   - react: call api.SetMessageReaction(), write response
   - typing: call api.SendChatAction(), no response
   - shutdown: cancel poll, wait for stop, write shutdown_complete, return nil
   - ping: write pong
3. Error handling: init failure writes error event, stdin close = shutdown, command failure writes response with ok:false

Internal types: ProviderConfig struct (Token string), Subprocess struct (bot client, stdout writer mutex-protected, cancel func), messageToEvent(msg, offset) protocol.MessageEvent, messageToSummary(msg) *protocol.MsgSummary. Use swappable var subprocessPollFunc = Poll for test injection.

## Files
- Create: providers/telegram/subprocess.go
- Modify: providers/telegram/telegram.go (add SendChatAction to BotAPI interface)

## Implementation Notes
- Use internal/protocol types and codec for all JSONL encode/decode
- ProviderConfig reads from COMMS_PROVIDER_CONFIG env var (JSON)
- Use mutex on stdout writer to prevent interleaved writes
- Use swappable package-level var for Poll function injection in tests

## Verification
```bash
go test ./providers/telegram/ -run TestSubprocess -v
```

## Acceptance Criteria
- RunSubprocess implements full JSONL protocol
- Handshake, message events, send/react/typing commands all work
- SendChatAction added to BotAPI interface

