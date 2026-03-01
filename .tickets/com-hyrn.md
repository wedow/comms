---
id: com-hyrn
status: open
deps: []
links: []
created: 2026-03-01T20:36:56Z
type: task
priority: 1
assignee: Greg Wedow
parent: com-d35c
tags: [telegram, v2]
---
# Extend Message struct with reply, media, reaction, edit, forward, thread, and entity fields

Add fields to `internal/message/Message` struct. All fields `yaml:",omitempty"` so existing messages stay compatible.

**String fields:**
- `ReplyTo string` — ID of message being replied to
- `ReplyToBody string` — preview/body of replied-to message
- `Quote string` — quoted text portion
- `MediaType string` — photo/video/audio/document/voice/animation/sticker/video_note
- `MediaFileID string` — Telegram file_id for download/reuse
- `MediaURL string` — local path to downloaded media file
- `Caption string` — media caption (separate from Body)
- `ForwardFrom string` — original sender for forwarded messages
- `ThreadID string` — forum topic ID
- `MediaGroupID string` — album grouping ID

**Structured fields:**
- `Entities []Entity` — text formatting. Define: `type Entity struct { Type string; Offset int; Length int; URL string }` (Type: bold/italic/code/pre/text_link/mention/url/etc. URL only for text_link.)

**Time fields:**
- `EditDate *time.Time` — when message was last edited (pointer for omitempty)
- `ForwardDate *time.Time` — original send date for forwards (pointer for omitempty)

Serialization: YAML frontmatter handles all fields automatically via struct tags. Org serialization in `org.go` needs new keyword lines for each field — follow existing pattern. Entities serialize as YAML list in frontmatter.

**Files to modify:** `internal/message/message.go`, `internal/message/markdown.go` (if needed), `internal/message/org.go`, and their test files.

## TDD

Write failing tests first. Test round-trip serialization for each new field in both markdown and org formats. Verify existing message files (without new fields) still parse correctly.

## Verification

```
go test ./internal/message/... -v
go vet ./...
```

## Acceptance Criteria

All new fields serialize/deserialize correctly in both markdown and org format. Existing message files still parse without error. Tests cover round-trip for each new field including the Entity slice.

