---
id: com-g34x
status: closed
deps: [com-azpc]
links: []
created: 2026-03-01T21:40:21Z
type: task
priority: 1
assignee: Greg Wedow
parent: com-d35c
tags: [telegram, v2, media]
---
# Download and store media files from Telegram

Add GetFile and FileDownloadLink to BotAPI interface. Add media download+save to store package. Wire download into daemon handler: if MediaFileID set after convert, call GetFile→download→save→set MediaURL before writing message. Storage layout: <channel>/<timestamp-dir>/001.<ext>. Preserve original extension from Telegram file_path. 20MB limit. Files: telegram.go (BotAPI), store/store.go, poll.go or daemon.go (wiring), test files. Depends on media metadata extraction subtask.

