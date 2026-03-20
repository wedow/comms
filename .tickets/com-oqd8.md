---
id: com-oqd8
status: closed
deps: [com-ab0w]
links: []
created: 2026-03-20T01:08:56Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Subprocess unit tests

## Summary
Comprehensive tests using io.Pipe for stdin/stdout. Construct Subprocess struct directly with mockBot.

Test scenarios:
1. Handshake: verify ready event
2. Start + message event: inject mock update via swappable poll func, verify message event
3. Send command: mock BotAPI, verify response ok:true
4. Send error: mock failure, verify response ok:false
5. React command: verify response ok:true
6. Shutdown: send shutdown, verify shutdown_complete
7. Stdin close: verify graceful exit
8. Protocol error: malformed JSON, verify logged but continues
9. Ping/pong: verify pong response

Test helper: startSubprocess(t) (*Subprocess, *io.PipeReader, *io.PipeWriter)

## Files
- Create: providers/telegram/subprocess_test.go

## Implementation Notes
- Use io.Pipe() pairs for stdin/stdout simulation
- Construct Subprocess struct directly with mockBot (from existing test infrastructure)
- Use swappable subprocessPollFunc var to inject test updates
- Each test scenario should be independent

## Verification
```bash
go test ./providers/telegram/ -run TestSubprocess -v -count=1
```

## Acceptance Criteria
- All 9 test scenarios pass
- Full protocol coverage


## Notes

**2026-03-20T03:14:23Z**

Added 2 missing test scenarios (send error, malformed JSON). Other 10 scenarios already covered by TDD during com-ab0w.
