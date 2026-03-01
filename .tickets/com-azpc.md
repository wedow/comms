---
id: com-azpc
status: closed
deps: []
links: []
created: 2026-03-01T21:40:13Z
type: task
priority: 1
assignee: Greg Wedow
parent: com-d35c
tags: [telegram, v2, media]
---
# Extract media metadata in convertMessage

Update convertMessage() to detect media type from non-nil fields (Photo, Video, Audio, Document, Voice, Animation, Sticker, VideoNote), extract FileID (for photos use last element of []PhotoSize for largest), set Caption from m.Caption, set MediaGroupID from m.MediaGroupID. For media-only messages, Body = Caption or empty. Pure conversion, no I/O, no interface changes. Files: convert.go, convert_test.go. TDD: test each media type with mock models.Message.

