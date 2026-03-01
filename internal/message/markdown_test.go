package message

import (
	"reflect"
	"testing"
	"time"
)

func TestMarkdownRoundTrip(t *testing.T) {
	msg := Message{
		From:     "alice",
		Provider: "telegram",
		Channel:  "general",
		Date:     time.Date(2026, 3, 1, 12, 30, 0, 0, time.UTC),
		ID:       "msg-001",
		Body:     "Hello, world!",
	}

	data, err := MarshalMarkdown(msg)
	if err != nil {
		t.Fatalf("MarshalMarkdown: %v", err)
	}

	got, err := UnmarshalMarkdown(data)
	if err != nil {
		t.Fatalf("UnmarshalMarkdown: %v", err)
	}

	if !reflect.DeepEqual(got, msg) {
		t.Errorf("round-trip mismatch:\n got  %+v\n want %+v", got, msg)
	}
}

func TestMarkdownEmptyBody(t *testing.T) {
	msg := Message{
		From:     "bob",
		Provider: "telegram",
		Channel:  "alerts",
		Date:     time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		ID:       "msg-002",
		Body:     "",
	}

	data, err := MarshalMarkdown(msg)
	if err != nil {
		t.Fatalf("MarshalMarkdown: %v", err)
	}

	got, err := UnmarshalMarkdown(data)
	if err != nil {
		t.Fatalf("UnmarshalMarkdown: %v", err)
	}

	if !reflect.DeepEqual(got, msg) {
		t.Errorf("round-trip mismatch:\n got  %+v\n want %+v", got, msg)
	}
}

func TestMarkdownMultilineBody(t *testing.T) {
	msg := Message{
		From:     "carol",
		Provider: "telegram",
		Channel:  "dev",
		Date:     time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC),
		ID:       "msg-003",
		Body:     "Line one.\n\nLine three.\n\n- item a\n- item b\n",
	}

	data, err := MarshalMarkdown(msg)
	if err != nil {
		t.Fatalf("MarshalMarkdown: %v", err)
	}

	got, err := UnmarshalMarkdown(data)
	if err != nil {
		t.Fatalf("UnmarshalMarkdown: %v", err)
	}

	if !reflect.DeepEqual(got, msg) {
		t.Errorf("round-trip mismatch:\n got  %+v\n want %+v", got, msg)
	}
}

func TestMarkdownSpecialCharacters(t *testing.T) {
	msg := Message{
		From:     "user: \"dave\"",
		Provider: "telegram",
		Channel:  "chat #1",
		Date:     time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
		ID:       "msg-004",
		Body:     "Special chars: --- ```yaml\nkey: value\n```\n",
	}

	data, err := MarshalMarkdown(msg)
	if err != nil {
		t.Fatalf("MarshalMarkdown: %v", err)
	}

	got, err := UnmarshalMarkdown(data)
	if err != nil {
		t.Fatalf("UnmarshalMarkdown: %v", err)
	}

	if !reflect.DeepEqual(got, msg) {
		t.Errorf("round-trip mismatch:\n got  %+v\n want %+v", got, msg)
	}
}

func TestMarkdownExtendedFieldsRoundTrip(t *testing.T) {
	editDate := time.Date(2026, 3, 1, 13, 0, 0, 0, time.UTC)
	fwdDate := time.Date(2026, 2, 28, 10, 0, 0, 0, time.UTC)
	msg := Message{
		From:         "alice",
		Provider:     "telegram",
		Channel:      "general",
		Date:         time.Date(2026, 3, 1, 12, 30, 0, 0, time.UTC),
		ID:           "msg-010",
		ReplyTo:      "msg-009",
		ReplyToBody:  "previous message text",
		Quote:        "quoted portion",
		MediaType:    "photo",
		MediaFileID:  "AgACAgIAAxk",
		MediaURL:     "/data/photos/abc.jpg",
		Caption:      "a nice photo",
		ForwardFrom:  "bob",
		ForwardDate:  &fwdDate,
		EditDate:     &editDate,
		ThreadID:     "topic-42",
		MediaGroupID: "album-7",
		Entities: []Entity{
			{Type: "bold", Offset: 0, Length: 5},
			{Type: "text_link", Offset: 6, Length: 4, URL: "https://example.com"},
		},
		Body: "Hello world!",
	}

	data, err := MarshalMarkdown(msg)
	if err != nil {
		t.Fatalf("MarshalMarkdown: %v", err)
	}

	got, err := UnmarshalMarkdown(data)
	if err != nil {
		t.Fatalf("UnmarshalMarkdown: %v", err)
	}

	if !reflect.DeepEqual(got, msg) {
		t.Errorf("round-trip mismatch:\n got  %+v\n want %+v", got, msg)
	}
}

func TestMarkdownBackwardCompatibility(t *testing.T) {
	// Old-format message without new fields should parse fine with zero-valued new fields.
	oldMd := "---\nfrom: alice\nprovider: telegram\nchannel: general\ndate: 2026-03-01T12:30:00Z\nid: msg-001\n---\nHello!"

	got, err := UnmarshalMarkdown([]byte(oldMd))
	if err != nil {
		t.Fatalf("UnmarshalMarkdown: %v", err)
	}

	want := Message{
		From:     "alice",
		Provider: "telegram",
		Channel:  "general",
		Date:     time.Date(2026, 3, 1, 12, 30, 0, 0, time.UTC),
		ID:       "msg-001",
		Body:     "Hello!",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("backward compat mismatch:\n got  %+v\n want %+v", got, want)
	}
}
