package message

import (
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

	if got != msg {
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

	if got != msg {
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

	if got != msg {
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

	if got != msg {
		t.Errorf("round-trip mismatch:\n got  %+v\n want %+v", got, msg)
	}
}
