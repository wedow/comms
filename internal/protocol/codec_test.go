package protocol

import (
	"bufio"
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestEncodeWritesJSONLine(t *testing.T) {
	var buf bytes.Buffer
	msg := PingEvent{Type: TypePing, TS: time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)}

	if err := Encode(&buf, msg); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	line := buf.String()
	if !strings.HasSuffix(line, "\n") {
		t.Errorf("output should end with newline, got %q", line)
	}
	if strings.Count(line, "\n") != 1 {
		t.Errorf("should have exactly one newline, got %d", strings.Count(line, "\n"))
	}
	if !strings.Contains(line, `"type":"ping"`) {
		t.Errorf("should contain type field, got %q", line)
	}
}

func TestDecodeReadsJSONLine(t *testing.T) {
	input := `{"type":"ping","ts":"2026-03-19T12:00:00Z"}` + "\n"
	r := bufio.NewReader(strings.NewReader(input))

	m, err := Decode(r)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if m["type"] != "ping" {
		t.Errorf("type = %v, want ping", m["type"])
	}
}

func TestDecodeSkipsBlankLines(t *testing.T) {
	input := "\n\n\n" + `{"type":"pong","ts":"2026-03-19T12:00:00Z"}` + "\n"
	r := bufio.NewReader(strings.NewReader(input))

	m, err := Decode(r)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if m["type"] != "pong" {
		t.Errorf("type = %v, want pong", m["type"])
	}
}

func TestDecodeInto(t *testing.T) {
	input := `{"type":"send","id":"req-1","chat_id":123,"text":"hello"}` + "\n"
	r := bufio.NewReader(strings.NewReader(input))

	var cmd SendCommand
	if err := DecodeInto(r, &cmd); err != nil {
		t.Fatalf("DecodeInto: %v", err)
	}
	if cmd.Type != TypeSend {
		t.Errorf("Type = %q, want %q", cmd.Type, TypeSend)
	}
	if cmd.ID != "req-1" {
		t.Errorf("ID = %q, want req-1", cmd.ID)
	}
	if cmd.Text != "hello" {
		t.Errorf("Text = %q, want hello", cmd.Text)
	}
}

func TestDecodeTypedReturnsConcreteTypes(t *testing.T) {
	tests := []struct {
		name string
		json string
		want any // expected concrete type (zero value)
	}{
		{"ready", `{"type":"ready","provider":"telegram","version":"1.0"}`, ReadyEvent{}},
		{"message", `{"type":"message","offset":1,"id":1,"chat_id":1,"channel":"ch","from":"u","date":"2026-03-19T12:00:00Z","body":"hi"}`, MessageEvent{}},
		{"edit", `{"type":"edit","offset":1,"id":1,"chat_id":1,"channel":"ch","from":"u","date":"2026-03-19T12:00:00Z","body":"edited"}`, MessageEvent{}},
		{"shutdown", `{"type":"shutdown"}`, ShutdownCommand{}},
		{"send", `{"type":"send","id":"r1","chat_id":1,"text":"hi"}`, SendCommand{}},
		{"error", `{"type":"error","code":500,"message":"fail"}`, ErrorEvent{}},
		{"ping", `{"type":"ping","ts":"2026-03-19T12:00:00Z"}`, PingEvent{}},
		{"pong", `{"type":"pong","ts":"2026-03-19T12:00:00Z"}`, PongEvent{}},
		{"reaction", `{"type":"reaction","offset":1,"channel":"ch","message_id":1,"from":"u","emoji":"👍","date":"2026-03-19T12:00:00Z"}`, ReactionEvent{}},
		{"response", `{"type":"response","id":"r1","ok":true}`, ResponseEvent{}},
		{"shutdown_complete", `{"type":"shutdown_complete"}`, ShutdownCompleteEvent{}},
		{"start", `{"type":"start","offset":0}`, StartCommand{}},
		{"send_media", `{"type":"send_media","id":"r1","chat_id":1,"media_type":"photo","path":"/tmp/a.jpg"}`, SendMediaCommand{}},
		{"react", `{"type":"react","id":"r1","chat_id":1,"message_id":1,"emoji":"👍"}`, ReactCommand{}},
		{"typing", `{"type":"typing","chat_id":1}`, TypingCommand{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tt.json + "\n"))
			v, err := DecodeTyped(r)
			if err != nil {
				t.Fatalf("DecodeTyped: %v", err)
			}
			// Check that the returned value has the same concrete type as want.
			gotType := concreteTypeName(v)
			wantType := concreteTypeName(tt.want)
			if gotType != wantType {
				t.Errorf("type = %s, want %s", gotType, wantType)
			}
		})
	}
}

func TestDecodeTypedUnknownType(t *testing.T) {
	input := `{"type":"bogus"}` + "\n"
	r := bufio.NewReader(strings.NewReader(input))

	_, err := DecodeTyped(r)
	if err == nil {
		t.Fatal("expected error for unknown type, got nil")
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error should mention unknown type, got %q", err.Error())
	}
}

func TestMaxLineLengthEnforced(t *testing.T) {
	big := strings.Repeat("x", 1<<20+1) + "\n"
	r := bufio.NewReader(strings.NewReader(big))

	_, err := Decode(r)
	if err == nil {
		t.Fatal("expected error for oversized line, got nil")
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	orig := SendCommand{
		Type:   TypeSend,
		ID:     "req-42",
		ChatID: -100123,
		Text:   "round trip test",
	}

	var buf bytes.Buffer
	if err := Encode(&buf, orig); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	r := bufio.NewReader(&buf)
	var got SendCommand
	if err := DecodeInto(r, &got); err != nil {
		t.Fatalf("DecodeInto: %v", err)
	}

	if got != orig {
		t.Errorf("round trip mismatch:\n got  %+v\n want %+v", got, orig)
	}
}

func TestEncodeDecodeTypedRoundTrip(t *testing.T) {
	ts := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
	fwd := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	edit := time.Date(2026, 3, 19, 13, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		orig any
	}{
		{"ready", ReadyEvent{Type: TypeReady, Provider: "telegram", Version: "1.0"}},
		{"message", MessageEvent{
			Type: TypeMessage, Offset: 42, ID: 101, ChatID: -100123,
			Channel: "general", From: "alice", Date: ts, Body: "hello world",
			ReplyTo: 99, ReplyToBody: "original", Quote: "quoted text",
			ThreadID: 5, MediaType: "photo", MediaFileID: "abc123",
			DownloadURL: "https://example.com/photo.jpg", MediaExt: ".jpg",
			Caption: "nice photo", ForwardFrom: "bob", ForwardDate: &fwd,
			EditDate: &edit, MediaGroupID: "grp-1",
			Entities: []Entity{{Type: "bold", Offset: 0, Length: 5}},
		}},
		{"edit", MessageEvent{
			Type: TypeEdit, Offset: 43, ID: 102, ChatID: -100123,
			Channel: "general", From: "alice", Date: ts, Body: "edited text",
			EditDate: &edit,
		}},
		{"reaction", ReactionEvent{
			Type: TypeReaction, Offset: 44, Channel: "general",
			MessageID: 101, From: "bob", Emoji: "\U0001f44d", Date: ts,
		}},
		{"response", ResponseEvent{
			Type: TypeResponse, ID: "req-1", OK: true,
			Message: &MsgSummary{
				ID: 201, ChatID: -100123, Channel: "general",
				From: "bot", Date: ts, Body: "sent",
			},
		}},
		{"response_error", ResponseEvent{
			Type: TypeResponse, ID: "req-2", OK: false, Error: "not found",
		}},
		{"error", ErrorEvent{Type: TypeError, Code: 500, Message: "internal error"}},
		{"shutdown_complete", ShutdownCompleteEvent{Type: TypeShutdownComplete}},
		{"ping", PingEvent{Type: TypePing, TS: ts}},
		{"pong", PongEvent{Type: TypePong, TS: ts}},
		{"start", StartCommand{Type: TypeStart, Offset: 100}},
		{"send", SendCommand{
			Type: TypeSend, ID: "req-3", ChatID: -100123, Text: "hello",
			ParseMode: "Markdown", ReplyToID: 50, ThreadID: 7,
		}},
		{"send_media", SendMediaCommand{
			Type: TypeSendMedia, ID: "req-4", ChatID: -100123,
			MediaType: "photo", Path: "/tmp/photo.jpg", Filename: "photo.jpg",
			Caption: "a caption", ParseMode: "HTML", ReplyToID: 50, ThreadID: 7,
		}},
		{"react", ReactCommand{
			Type: TypeReact, ID: "req-5", ChatID: -100123, MessageID: 101, Emoji: "\U0001f44d",
		}},
		{"typing", TypingCommand{Type: TypeTyping, ChatID: -100123}},
		{"shutdown", ShutdownCommand{Type: TypeShutdown, Reason: "graceful"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := Encode(&buf, tt.orig); err != nil {
				t.Fatalf("Encode: %v", err)
			}

			r := bufio.NewReader(&buf)
			got, err := DecodeTyped(r)
			if err != nil {
				t.Fatalf("DecodeTyped: %v", err)
			}

			if !reflect.DeepEqual(got, tt.orig) {
				t.Errorf("round trip mismatch:\n got  %+v\n want %+v", got, tt.orig)
			}
		})
	}
}

// concreteTypeName returns a type name via type switch for test assertions.
func concreteTypeName(v any) string {
	switch v.(type) {
	case ReadyEvent:
		return "ReadyEvent"
	case MessageEvent:
		return "MessageEvent"
	case ReactionEvent:
		return "ReactionEvent"
	case ResponseEvent:
		return "ResponseEvent"
	case ErrorEvent:
		return "ErrorEvent"
	case ShutdownCompleteEvent:
		return "ShutdownCompleteEvent"
	case PingEvent:
		return "PingEvent"
	case PongEvent:
		return "PongEvent"
	case StartCommand:
		return "StartCommand"
	case SendCommand:
		return "SendCommand"
	case SendMediaCommand:
		return "SendMediaCommand"
	case ReactCommand:
		return "ReactCommand"
	case TypingCommand:
		return "TypingCommand"
	case ShutdownCommand:
		return "ShutdownCommand"
	default:
		return "unknown"
	}
}
