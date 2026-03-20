package protocol

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMessageEventRoundTrip(t *testing.T) {
	ts := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
	orig := MessageEvent{
		Type:    TypeMessage,
		Offset:  42,
		ID:      101,
		ChatID:  -100123,
		Channel: "general",
		From:    "alice",
		Date:    ts,
		Body:    "hello world",
		ReplyTo: 99,
		Entities: []Entity{
			{Type: "bold", Offset: 0, Length: 5},
		},
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got MessageEvent
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Type != orig.Type {
		t.Errorf("Type = %q, want %q", got.Type, orig.Type)
	}
	if got.Offset != orig.Offset {
		t.Errorf("Offset = %d, want %d", got.Offset, orig.Offset)
	}
	if got.ID != orig.ID {
		t.Errorf("ID = %d, want %d", got.ID, orig.ID)
	}
	if got.ChatID != orig.ChatID {
		t.Errorf("ChatID = %d, want %d", got.ChatID, orig.ChatID)
	}
	if got.Channel != orig.Channel {
		t.Errorf("Channel = %q, want %q", got.Channel, orig.Channel)
	}
	if got.From != orig.From {
		t.Errorf("From = %q, want %q", got.From, orig.From)
	}
	if !got.Date.Equal(orig.Date) {
		t.Errorf("Date = %v, want %v", got.Date, orig.Date)
	}
	if got.Body != orig.Body {
		t.Errorf("Body = %q, want %q", got.Body, orig.Body)
	}
	if got.ReplyTo != orig.ReplyTo {
		t.Errorf("ReplyTo = %d, want %d", got.ReplyTo, orig.ReplyTo)
	}
	if len(got.Entities) != 1 {
		t.Fatalf("Entities len = %d, want 1", len(got.Entities))
	}
	if got.Entities[0].Type != "bold" {
		t.Errorf("Entity.Type = %q, want %q", got.Entities[0].Type, "bold")
	}
}

func TestSendCommandRoundTrip(t *testing.T) {
	orig := SendCommand{
		Type:      TypeSend,
		ID:        "req-1",
		ChatID:    -100123,
		Text:      "hi there",
		ParseMode: "Markdown",
		ReplyToID: 5,
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got SendCommand
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got != orig {
		t.Errorf("got %+v, want %+v", got, orig)
	}
}

func TestResponseEventRoundTrip(t *testing.T) {
	ts := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
	orig := ResponseEvent{
		Type: TypeResponse,
		ID:   "req-1",
		OK:   true,
		Message: &MsgSummary{
			ID:      201,
			ChatID:  -100123,
			Channel: "general",
			From:    "bot",
			Date:    ts,
			Body:    "sent",
		},
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got ResponseEvent
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Type != orig.Type {
		t.Errorf("Type = %q, want %q", got.Type, orig.Type)
	}
	if got.ID != orig.ID {
		t.Errorf("ID = %q, want %q", got.ID, orig.ID)
	}
	if got.OK != orig.OK {
		t.Errorf("OK = %v, want %v", got.OK, orig.OK)
	}
	if got.Message == nil {
		t.Fatal("Message is nil, want non-nil")
	}
	if got.Message.ID != orig.Message.ID {
		t.Errorf("Message.ID = %d, want %d", got.Message.ID, orig.Message.ID)
	}
	if got.Message.Body != orig.Message.Body {
		t.Errorf("Message.Body = %q, want %q", got.Message.Body, orig.Message.Body)
	}
}

func TestOmitemptyFields(t *testing.T) {
	// MessageEvent with only required fields should omit optional ones
	msg := MessageEvent{
		Type:    TypeMessage,
		Offset:  1,
		ID:      1,
		ChatID:  1,
		Channel: "ch",
		From:    "user",
		Date:    time.Now(),
		Body:    "text",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	omittedFields := []string{"reply_to", "reply_to_body", "quote", "thread_id",
		"media_type", "media_file_id", "download_url", "media_ext",
		"caption", "forward_from", "forward_date", "edit_date",
		"media_group_id", "entities"}

	for _, field := range omittedFields {
		if _, ok := raw[field]; ok {
			t.Errorf("field %q should be omitted when zero, but present in JSON", field)
		}
	}
}

func TestEditEventIsMessageEvent(t *testing.T) {
	// EditEvent is a type alias for MessageEvent
	var e EditEvent
	e.Type = TypeEdit
	e.Body = "edited text"

	// Should marshal the same as MessageEvent
	data, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m MessageEvent
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if m.Type != TypeEdit {
		t.Errorf("Type = %q, want %q", m.Type, TypeEdit)
	}
	if m.Body != "edited text" {
		t.Errorf("Body = %q, want %q", m.Body, "edited text")
	}
}
