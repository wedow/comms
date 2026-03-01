---
id: com-vdst
status: open
deps: [com-hyrn]
links: []
created: 2026-03-01T20:37:28Z
type: feature
priority: 1
assignee: Greg Wedow
parent: com-d35c
tags: [telegram, v2, media]
---
# Receive media messages from Telegram

**Conversion** — update `convertMessage()` to extract media metadata:
- Detect media type from which field is non-nil: `m.Photo`, `m.Video`, `m.Audio`, `m.Document`, `m.Voice`, `m.Animation`, `m.Sticker`, `m.VideoNote`
- Extract `file_id` — for photos use last element of `[]models.PhotoSize` (largest resolution)
- Extract `m.Caption` (separate from `m.Text`) → `Message.Caption`. For media-only messages, `Message.Body` = caption or empty
- Extract `m.MediaGroupID` → `Message.MediaGroupID`
- Set `Message.MediaType`, `Message.MediaFileID`

**Download** — save raw media files alongside message files. No conversion.
- Add `GetFile` to BotAPI interface: `GetFile(ctx context.Context, params *bot.GetFileParams) (*models.File, error)`
- Download workflow: `getFile(file_id)` → get `file_path` → HTTP GET `https://api.telegram.org/file/bot<token>/<file_path>` → save to disk
- **Storage layout:** media directory matches the message filename:
  ```
  telegram-dm-123/
    2026-03-01T20-20-38Z.md          # message file with media metadata
    2026-03-01T20-20-38Z/            # media directory
      001.jpg                        # index-numbered, preserves order for albums
  ```
- Preserve original file extension from Telegram's `file_path`
- Set `Message.MediaURL` to the local relative path (e.g., `2026-03-01T20-20-38Z/001.jpg`)
- Download limit: 20 MB via standard Telegram API

**Files to modify:** `internal/provider/telegram/convert.go`, `internal/provider/telegram/telegram.go` (BotAPI), `internal/store/` (media dir creation + file save), daemon handler, and test files.

## TDD

Write failing tests first. Test convertMessage with each media type (mock models.Message with Photo/Video/etc set). Test media directory creation and file naming. Test download workflow with mock HTTP.

## Verification

```
go test ./internal/provider/telegram/... -v
go test ./internal/store/... -v
go vet ./...
```

## Acceptance Criteria

Photo, video, audio, document, voice messages are received with correct media type. Media files downloaded to `<channel>/<timestamp>/001.<ext>`. comms unread output includes media_type, media_url (local path), and caption. Tests cover each media type.

