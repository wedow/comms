package message

import (
	"reflect"
	"testing"
	"time"
)

func TestOrgRoundTrip(t *testing.T) {
	msg := Message{
		From:     "alice",
		Provider: "telegram",
		Channel:  "general",
		Date:     time.Date(2026, 3, 1, 12, 30, 0, 0, time.UTC),
		ID:       "msg-001",
		Body:     "Hello, world!",
	}

	data, err := MarshalOrg(msg)
	if err != nil {
		t.Fatalf("MarshalOrg: %v", err)
	}

	got, err := UnmarshalOrg(data)
	if err != nil {
		t.Fatalf("UnmarshalOrg: %v", err)
	}

	if !reflect.DeepEqual(got, msg) {
		t.Errorf("round-trip mismatch:\n got  %+v\n want %+v", got, msg)
	}
}

func TestOrgEmptyBody(t *testing.T) {
	msg := Message{
		From:     "bob",
		Provider: "telegram",
		Channel:  "alerts",
		Date:     time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		ID:       "msg-002",
		Body:     "",
	}

	data, err := MarshalOrg(msg)
	if err != nil {
		t.Fatalf("MarshalOrg: %v", err)
	}

	got, err := UnmarshalOrg(data)
	if err != nil {
		t.Fatalf("UnmarshalOrg: %v", err)
	}

	if !reflect.DeepEqual(got, msg) {
		t.Errorf("round-trip mismatch:\n got  %+v\n want %+v", got, msg)
	}
}

func TestOrgMultilineBody(t *testing.T) {
	msg := Message{
		From:     "carol",
		Provider: "telegram",
		Channel:  "dev",
		Date:     time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC),
		ID:       "msg-003",
		Body:     "Line one.\n\nLine three.\n\n- item a\n- item b\n",
	}

	data, err := MarshalOrg(msg)
	if err != nil {
		t.Fatalf("MarshalOrg: %v", err)
	}

	got, err := UnmarshalOrg(data)
	if err != nil {
		t.Fatalf("UnmarshalOrg: %v", err)
	}

	if !reflect.DeepEqual(got, msg) {
		t.Errorf("round-trip mismatch:\n got  %+v\n want %+v", got, msg)
	}
}

func TestOrgSpecialCharacters(t *testing.T) {
	msg := Message{
		From:     "user: \"dave\"",
		Provider: "telegram",
		Channel:  "chat #1",
		Date:     time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
		ID:       "msg-004",
		Body:     "Special chars: --- ```yaml\nkey: value\n```\n",
	}

	data, err := MarshalOrg(msg)
	if err != nil {
		t.Fatalf("MarshalOrg: %v", err)
	}

	got, err := UnmarshalOrg(data)
	if err != nil {
		t.Fatalf("UnmarshalOrg: %v", err)
	}

	if !reflect.DeepEqual(got, msg) {
		t.Errorf("round-trip mismatch:\n got  %+v\n want %+v", got, msg)
	}
}

func TestOrgExtendedFieldsRoundTrip(t *testing.T) {
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

	data, err := MarshalOrg(msg)
	if err != nil {
		t.Fatalf("MarshalOrg: %v", err)
	}

	got, err := UnmarshalOrg(data)
	if err != nil {
		t.Fatalf("UnmarshalOrg: %v", err)
	}

	if !reflect.DeepEqual(got, msg) {
		t.Errorf("round-trip mismatch:\n got  %+v\n want %+v", got, msg)
	}
}

func TestOrgBackwardCompatibility(t *testing.T) {
	// Old-format message without new fields should parse fine with zero-valued new fields.
	oldOrg := "#+FROM: alice\n#+PROVIDER: telegram\n#+CHANNEL: general\n#+DATE: 2026-03-01T12:30:00Z\n#+ID: msg-001\n\nHello!"

	got, err := UnmarshalOrg([]byte(oldOrg))
	if err != nil {
		t.Fatalf("UnmarshalOrg: %v", err)
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
