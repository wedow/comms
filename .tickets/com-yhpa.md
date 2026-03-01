---
id: com-yhpa
status: open
deps: [com-hyrn, com-vdst]
links: []
created: 2026-03-01T20:37:33Z
type: feature
priority: 1
assignee: Greg Wedow
parent: com-d35c
tags: [telegram, v2, media]
---
# Send media via comms send

Add `--file` and `--media-type` flags to `comms send` for sending media.

**CLI interface:**
- `--file <path>` — path to local file to send (multipart upload)
- `--media-type <type>` — optional override: photo/document/audio/video/voice/animation. If omitted, auto-detect from file extension
- Stdin becomes the caption when `--file` is provided
- Example: `echo 'check this out' | comms send --channel telegram-dm-123 --file ./photo.jpg`
- Example: `comms send --channel telegram-dm-123 --file ./data.csv --media-type document`

**Auto-detection mapping:**
- `.jpg`, `.jpeg`, `.png`, `.gif`, `.webp` → photo (`.gif` → animation)
- `.mp4`, `.mov`, `.avi` → video
- `.mp3`, `.ogg`, `.flac`, `.wav` → audio (`.ogg` → voice)
- Everything else → document

**BotAPI interface additions:**
- `SendPhoto(ctx, *bot.SendPhotoParams) (*models.Message, error)`
- `SendDocument(ctx, *bot.SendDocumentParams) (*models.Message, error)`
- `SendAudio(ctx, *bot.SendAudioParams) (*models.Message, error)`
- `SendVideo(ctx, *bot.SendVideoParams) (*models.Message, error)`
- `SendVoice(ctx, *bot.SendVoiceParams) (*models.Message, error)`
- `SendAnimation(ctx, *bot.SendAnimationParams) (*models.Message, error)`

**Files to modify:** `internal/provider/telegram/telegram.go` (BotAPI), `internal/provider/telegram/send.go` (new send functions), `internal/cli/send.go` (flags + dispatch), and test files.

## TDD

Write failing tests first. Test media type auto-detection. Test that --file dispatches to correct send method (mock BotAPI). Test --media-type override. Test caption from stdin.

## Verification

```
go test ./internal/provider/telegram/... -v
go test ./internal/cli/... -v
go vet ./...
```

## Acceptance Criteria

`comms send --file` sends media with optional caption. Auto-detection picks correct Telegram method. `--media-type` overrides detection. Tests mock each send method.

