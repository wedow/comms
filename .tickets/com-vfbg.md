---
id: com-vfbg
status: open
deps: [com-hyrn]
links: []
created: 2026-03-01T20:37:49Z
type: feature
priority: 2
assignee: Greg Wedow
parent: com-d35c
tags: [telegram, v2, threads]
---
# Support forum topics and message threads

**Design decision:** Topics are sub-directories under the channel. Messages in topic 123 of group "dev-chat" go in `telegram-dev-chat/topic-123/`. Messages not in any topic go directly in the channel directory (existing behavior).

**Receive:**
- Extract `m.MessageThreadID` from incoming messages
- In the daemon handler, if ThreadID is set, write the message file to `<channel>/topic-<id>/` instead of `<channel>/`
- ThreadID stored in message frontmatter for reference

**Send:**
- Add `--thread <id>` flag to `comms send`
- Pass `MessageThreadID` in `bot.SendMessageParams` when set
- Also works with `comms react` and media sends

**SDK fields:** `models.Message.MessageThreadID` (int), `bot.SendMessageParams.MessageThreadID` (int)

**Files to modify:** `internal/provider/telegram/convert.go`, `internal/store/` (topic-aware write path), `internal/cli/send.go`, and test files.

## TDD

Write failing tests first. Test ThreadID extraction in convertMessage. Test topic sub-directory creation in store. Test --thread flag sets MessageThreadID on send.

## Verification

```
go test ./internal/provider/telegram/... -v
go test ./internal/store/... -v
go test ./internal/cli/... -v
go vet ./...
```

## Acceptance Criteria

Messages from forum topics are stored in `<channel>/topic-<id>/` sub-directories. `comms send --thread <id>` posts to the correct topic. comms unread includes thread_id. Tests cover both receive and send.

