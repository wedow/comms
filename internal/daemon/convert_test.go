package daemon

import (
	"testing"
	"time"

	"github.com/wedow/comms/internal/protocol"
)

func TestProtocolToMessageAllFields(t *testing.T) {
	fwd := time.Date(2026, 3, 1, 11, 0, 0, 0, time.UTC)
	edit := time.Date(2026, 3, 1, 12, 5, 0, 0, time.UTC)

	evt := protocol.MessageEvent{
		Type:         protocol.TypeMessage,
		ID:           42,
		ChatID:       -1001234,
		Channel:      "general",
		From:         "alice",
		Date:         time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		Body:         "hello world",
		ReplyTo:      10,
		ReplyToBody:  "original message",
		Quote:        "quoted text",
		ThreadID:     5,
		MediaType:    "photo",
		MediaFileID:  "abc123",
		DownloadURL:  "https://example.com/file.jpg",
		MediaExt:     ".jpg",
		Caption:      "a photo",
		ForwardFrom:  "bob",
		ForwardDate:  &fwd,
		EditDate:     &edit,
		MediaGroupID: "grp-1",
		Entities: []protocol.Entity{
			{Type: "bold", Offset: 0, Length: 5},
			{Type: "text_link", Offset: 6, Length: 5, URL: "https://example.com"},
		},
	}

	msg := protocolToMessage("telegram", evt)

	if msg.Provider != "telegram" {
		t.Errorf("Provider = %q, want %q", msg.Provider, "telegram")
	}
	if msg.ID != "telegram-42" {
		t.Errorf("ID = %q, want %q", msg.ID, "telegram-42")
	}
	if msg.Channel != "general" {
		t.Errorf("Channel = %q, want %q", msg.Channel, "general")
	}
	if msg.From != "alice" {
		t.Errorf("From = %q, want %q", msg.From, "alice")
	}
	if !msg.Date.Equal(evt.Date) {
		t.Errorf("Date = %v, want %v", msg.Date, evt.Date)
	}
	if msg.Body != "hello world" {
		t.Errorf("Body = %q, want %q", msg.Body, "hello world")
	}
	if msg.ReplyTo != "telegram-10" {
		t.Errorf("ReplyTo = %q, want %q", msg.ReplyTo, "telegram-10")
	}
	if msg.ReplyToBody != "original message" {
		t.Errorf("ReplyToBody = %q, want %q", msg.ReplyToBody, "original message")
	}
	if msg.Quote != "quoted text" {
		t.Errorf("Quote = %q, want %q", msg.Quote, "quoted text")
	}
	if msg.ThreadID != "5" {
		t.Errorf("ThreadID = %q, want %q", msg.ThreadID, "5")
	}
	if msg.MediaType != "photo" {
		t.Errorf("MediaType = %q, want %q", msg.MediaType, "photo")
	}
	if msg.MediaFileID != "abc123" {
		t.Errorf("MediaFileID = %q, want %q", msg.MediaFileID, "abc123")
	}
	if msg.DownloadURL != "https://example.com/file.jpg" {
		t.Errorf("DownloadURL = %q, want %q", msg.DownloadURL, "https://example.com/file.jpg")
	}
	if msg.MediaExt != ".jpg" {
		t.Errorf("MediaExt = %q, want %q", msg.MediaExt, ".jpg")
	}
	if msg.Caption != "a photo" {
		t.Errorf("Caption = %q, want %q", msg.Caption, "a photo")
	}
	if msg.ForwardFrom != "bob" {
		t.Errorf("ForwardFrom = %q, want %q", msg.ForwardFrom, "bob")
	}
	if msg.ForwardDate == nil || !msg.ForwardDate.Equal(fwd) {
		t.Errorf("ForwardDate = %v, want %v", msg.ForwardDate, fwd)
	}
	if msg.EditDate == nil || !msg.EditDate.Equal(edit) {
		t.Errorf("EditDate = %v, want %v", msg.EditDate, edit)
	}
	if msg.MediaGroupID != "grp-1" {
		t.Errorf("MediaGroupID = %q, want %q", msg.MediaGroupID, "grp-1")
	}
	if len(msg.Entities) != 2 {
		t.Fatalf("Entities len = %d, want 2", len(msg.Entities))
	}
	if msg.Entities[0].Type != "bold" || msg.Entities[0].Offset != 0 || msg.Entities[0].Length != 5 {
		t.Errorf("Entities[0] = %+v, want bold/0/5", msg.Entities[0])
	}
	if msg.Entities[1].URL != "https://example.com" {
		t.Errorf("Entities[1].URL = %q, want %q", msg.Entities[1].URL, "https://example.com")
	}
}

func TestProtocolToMessageZeroOptionals(t *testing.T) {
	evt := protocol.MessageEvent{
		ID:      7,
		Channel: "chat",
		From:    "bob",
		Date:    time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		Body:    "hi",
	}

	msg := protocolToMessage("telegram", evt)

	if msg.ReplyTo != "" {
		t.Errorf("ReplyTo = %q, want empty for zero value", msg.ReplyTo)
	}
	if msg.ThreadID != "" {
		t.Errorf("ThreadID = %q, want empty for zero value", msg.ThreadID)
	}
	if msg.ForwardDate != nil {
		t.Errorf("ForwardDate = %v, want nil", msg.ForwardDate)
	}
	if msg.EditDate != nil {
		t.Errorf("EditDate = %v, want nil", msg.EditDate)
	}
	if len(msg.Entities) != 0 {
		t.Errorf("Entities len = %d, want 0", len(msg.Entities))
	}
}

func TestProtocolToReaction(t *testing.T) {
	evt := protocol.ReactionEvent{
		Type:      protocol.TypeReaction,
		Channel:   "telegram-general",
		MessageID: 42,
		From:      "bob",
		Emoji:     "\U0001f44d",
		Date:      time.Date(2026, 3, 1, 12, 10, 0, 0, time.UTC),
	}

	r := protocolToReaction("telegram", evt)

	if r.Channel != "telegram-general" {
		t.Errorf("Channel = %q, want %q", r.Channel, "telegram-general")
	}
	if r.MsgID != "telegram-42" {
		t.Errorf("MsgID = %q, want %q", r.MsgID, "telegram-42")
	}
	if r.From != "bob" {
		t.Errorf("From = %q, want %q", r.From, "bob")
	}
	if r.Emoji != "\U0001f44d" {
		t.Errorf("Emoji = %q, want %q", r.Emoji, "\U0001f44d")
	}
	if !r.Date.Equal(evt.Date) {
		t.Errorf("Date = %v, want %v", r.Date, evt.Date)
	}
}
