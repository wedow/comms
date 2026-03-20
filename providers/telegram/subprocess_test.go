package telegram

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/wedow/comms/internal/message"
	"github.com/wedow/comms/internal/protocol"
)

// subprocessMockBot extends mockBot with function fields for SendChatAction and SetMessageReaction.
type subprocessMockBot struct {
	mockBot
	sendChatActionFn    func(ctx context.Context, params *bot.SendChatActionParams) (bool, error)
	setMessageReactionFn func(ctx context.Context, params *bot.SetMessageReactionParams) (bool, error)
}

func (m *subprocessMockBot) SendChatAction(ctx context.Context, params *bot.SendChatActionParams) (bool, error) {
	if m.sendChatActionFn != nil {
		return m.sendChatActionFn(ctx, params)
	}
	return true, nil
}

func (m *subprocessMockBot) SetMessageReaction(ctx context.Context, params *bot.SetMessageReactionParams) (bool, error) {
	if m.setMessageReactionFn != nil {
		return m.setMessageReactionFn(ctx, params)
	}
	return true, nil
}

// writeCmd encodes a protocol message to the writer.
func writeCmd(w io.Writer, msg any) {
	protocol.Encode(w, msg)
}

// readEvent reads one JSON line from the reader and returns the decoded map.
func readEvent(t *testing.T, r *bufio.Reader) map[string]any {
	t.Helper()
	line, err := r.ReadBytes('\n')
	if err != nil {
		t.Fatalf("readEvent: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		t.Fatalf("readEvent unmarshal: %v (line: %s)", err, line)
	}
	return m
}

// runSubprocessHelper sets up pipes and runs RunSubprocess with the given mock bot.
// Returns the stdin writer, stdout reader, and a channel that receives the error from RunSubprocess.
func runSubprocessHelper(t *testing.T, mock BotAPI) (io.WriteCloser, *bufio.Reader, <-chan error) {
	t.Helper()
	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()

	oldNewBot := newBotFunc
	newBotFunc = func(_ string) (BotAPI, error) { return mock, nil }
	t.Cleanup(func() { newBotFunc = oldNewBot })

	// Prevent poll from being called unless explicitly set.
	oldPoll := subprocessPollFunc
	subprocessPollFunc = func(ctx context.Context, _ string, _ int64, _ func(message.Message, int64, bool), _ func(string, int, string, string, time.Time)) (int64, error) {
		<-ctx.Done()
		return 0, ctx.Err()
	}
	t.Cleanup(func() { subprocessPollFunc = oldPoll })

	errCh := make(chan error, 1)
	go func() {
		errCh <- RunSubprocess(context.Background(), stdinR, stdoutW, `{"token":"test-token"}`)
	}()

	return stdinW, bufio.NewReader(stdoutR), errCh
}

func TestRunSubprocess_Handshake(t *testing.T) {
	mock := &subprocessMockBot{}
	stdinW, stdout, _ := runSubprocessHelper(t, mock)

	evt := readEvent(t, stdout)
	if evt["type"] != "ready" {
		t.Errorf("type = %v, want ready", evt["type"])
	}
	if evt["provider"] != "telegram" {
		t.Errorf("provider = %v, want telegram", evt["provider"])
	}
	if evt["version"] != "1" {
		t.Errorf("version = %v, want 1", evt["version"])
	}

	// Clean shutdown.
	writeCmd(stdinW, protocol.ShutdownCommand{Type: "shutdown"})
	readEvent(t, stdout) // shutdown_complete
	stdinW.Close()
}

func TestRunSubprocess_PingPong(t *testing.T) {
	mock := &subprocessMockBot{}
	stdinW, stdout, _ := runSubprocessHelper(t, mock)

	// Consume ready event.
	readEvent(t, stdout)

	ts := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
	writeCmd(stdinW, protocol.PingEvent{Type: "ping", TS: ts})

	evt := readEvent(t, stdout)
	if evt["type"] != "pong" {
		t.Errorf("type = %v, want pong", evt["type"])
	}

	writeCmd(stdinW, protocol.ShutdownCommand{Type: "shutdown"})
	readEvent(t, stdout) // shutdown_complete
	stdinW.Close()
}

func TestRunSubprocess_Send(t *testing.T) {
	mock := &subprocessMockBot{
		mockBot: mockBot{
			sendFn: func(_ context.Context, p *bot.SendMessageParams) (*models.Message, error) {
				if p.ChatID != int64(123) {
					t.Errorf("ChatID = %v, want 123", p.ChatID)
				}
				if p.Text != "hello world" {
					t.Errorf("Text = %q, want %q", p.Text, "hello world")
				}
				return &models.Message{
					ID:   42,
					Chat: models.Chat{ID: 123, Type: models.ChatTypeGroup, Title: "Dev"},
					From: &models.User{Username: "testbot"},
					Date: 1709312400,
					Text: "hello world",
				}, nil
			},
		},
	}

	stdinW, stdout, _ := runSubprocessHelper(t, mock)
	readEvent(t, stdout) // ready

	writeCmd(stdinW, protocol.SendCommand{
		Type:   "send",
		ID:     "req-1",
		ChatID: 123,
		Text:   "hello world",
	})

	evt := readEvent(t, stdout)
	if evt["type"] != "response" {
		t.Fatalf("type = %v, want response", evt["type"])
	}
	if evt["id"] != "req-1" {
		t.Errorf("id = %v, want req-1", evt["id"])
	}
	if evt["ok"] != true {
		t.Errorf("ok = %v, want true", evt["ok"])
	}
	msg, ok := evt["message"].(map[string]any)
	if !ok {
		t.Fatalf("message not a map: %T", evt["message"])
	}
	if msg["id"] != float64(42) {
		t.Errorf("message.id = %v, want 42", msg["id"])
	}

	writeCmd(stdinW, protocol.ShutdownCommand{Type: "shutdown"})
	readEvent(t, stdout) // shutdown_complete
	stdinW.Close()
}

func TestRunSubprocess_Typing(t *testing.T) {
	actionCalled := false
	mock := &subprocessMockBot{
		sendChatActionFn: func(_ context.Context, p *bot.SendChatActionParams) (bool, error) {
			actionCalled = true
			if p.ChatID != int64(456) {
				t.Errorf("ChatID = %v, want 456", p.ChatID)
			}
			return true, nil
		},
	}

	stdinW, stdout, _ := runSubprocessHelper(t, mock)
	readEvent(t, stdout) // ready

	writeCmd(stdinW, protocol.TypingCommand{Type: "typing", ChatID: 456})

	// Typing doesn't produce a response. Send a ping to verify processing continued.
	writeCmd(stdinW, protocol.PingEvent{Type: "ping", TS: time.Now()})
	evt := readEvent(t, stdout)
	if evt["type"] != "pong" {
		t.Errorf("expected pong after typing, got %v", evt["type"])
	}
	if !actionCalled {
		t.Error("SendChatAction was not called")
	}

	writeCmd(stdinW, protocol.ShutdownCommand{Type: "shutdown"})
	readEvent(t, stdout) // shutdown_complete
	stdinW.Close()
}

func TestRunSubprocess_Shutdown(t *testing.T) {
	mock := &subprocessMockBot{}
	stdinW, stdout, errCh := runSubprocessHelper(t, mock)
	readEvent(t, stdout) // ready

	writeCmd(stdinW, protocol.ShutdownCommand{Type: "shutdown"})

	evt := readEvent(t, stdout)
	if evt["type"] != "shutdown_complete" {
		t.Errorf("type = %v, want shutdown_complete", evt["type"])
	}

	stdinW.Close()
	err := <-errCh
	if err != nil {
		t.Errorf("RunSubprocess returned error: %v", err)
	}
}

func TestRunSubprocess_React(t *testing.T) {
	reactionCalled := false
	mock := &subprocessMockBot{
		setMessageReactionFn: func(_ context.Context, p *bot.SetMessageReactionParams) (bool, error) {
			reactionCalled = true
			if p.ChatID != int64(789) {
				t.Errorf("ChatID = %v, want 789", p.ChatID)
			}
			if p.MessageID != 42 {
				t.Errorf("MessageID = %v, want 42", p.MessageID)
			}
			return true, nil
		},
	}

	stdinW, stdout, _ := runSubprocessHelper(t, mock)
	readEvent(t, stdout) // ready

	writeCmd(stdinW, protocol.ReactCommand{
		Type:      "react",
		ID:        "req-2",
		ChatID:    789,
		MessageID: 42,
		Emoji:     "👍",
	})

	evt := readEvent(t, stdout)
	if evt["type"] != "response" {
		t.Fatalf("type = %v, want response", evt["type"])
	}
	if evt["id"] != "req-2" {
		t.Errorf("id = %v, want req-2", evt["id"])
	}
	if evt["ok"] != true {
		t.Errorf("ok = %v, want true", evt["ok"])
	}
	if !reactionCalled {
		t.Error("SetMessageReaction was not called")
	}

	writeCmd(stdinW, protocol.ShutdownCommand{Type: "shutdown"})
	readEvent(t, stdout) // shutdown_complete
	stdinW.Close()
}

func TestRunSubprocess_StdinEOF(t *testing.T) {
	mock := &subprocessMockBot{}
	stdinW, stdout, errCh := runSubprocessHelper(t, mock)
	readEvent(t, stdout) // ready

	// Close stdin without sending shutdown - should treat as clean exit.
	stdinW.Close()

	err := <-errCh
	if err != nil {
		t.Errorf("RunSubprocess returned error on stdin EOF: %v", err)
	}
}

func TestMessageToEvent(t *testing.T) {
	msg := message.Message{
		ID:       "telegram-42",
		From:     "alice",
		Provider: "telegram",
		Channel:  "dev-team",
		Date:     time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC),
		Body:     "hello",
		ReplyTo:  "telegram-10",
		ThreadID: "5",
	}

	evt := messageToEvent(msg, 100, false)
	if evt.Type != "message" {
		t.Errorf("Type = %q, want message", evt.Type)
	}
	if evt.ID != 42 {
		t.Errorf("ID = %d, want 42", evt.ID)
	}
	if evt.ChatID != 100 {
		t.Errorf("ChatID = %d, want 100", evt.ChatID)
	}
	if evt.Offset != 0 {
		t.Errorf("Offset = %d, want 0", evt.Offset)
	}
	if evt.Channel != "dev-team" {
		t.Errorf("Channel = %q, want dev-team", evt.Channel)
	}
	if evt.From != "alice" {
		t.Errorf("From = %q, want alice", evt.From)
	}
	if evt.Body != "hello" {
		t.Errorf("Body = %q, want hello", evt.Body)
	}
	if evt.ReplyTo != 10 {
		t.Errorf("ReplyTo = %d, want 10", evt.ReplyTo)
	}
	if evt.ThreadID != 5 {
		t.Errorf("ThreadID = %d, want 5", evt.ThreadID)
	}
}

func TestMessageToEvent_Edit(t *testing.T) {
	editDate := time.Date(2026, 3, 19, 13, 0, 0, 0, time.UTC)
	msg := message.Message{
		ID:       "telegram-42",
		From:     "alice",
		Provider: "telegram",
		Channel:  "dev-team",
		Date:     time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC),
		Body:     "edited",
		EditDate: &editDate,
	}

	evt := messageToEvent(msg, 101, true)
	if evt.Type != "edit" {
		t.Errorf("Type = %q, want edit", evt.Type)
	}
}

func TestMessageToSummary(t *testing.T) {
	msg := message.Message{
		ID:       "telegram-42",
		From:     "bob",
		Provider: "telegram",
		Channel:  "general",
		Date:     time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC),
		Body:     "test message",
	}

	summary := messageToSummary(msg)
	if summary.ID != 42 {
		t.Errorf("ID = %d, want 42", summary.ID)
	}
	if summary.Channel != "general" {
		t.Errorf("Channel = %q, want general", summary.Channel)
	}
	if summary.From != "bob" {
		t.Errorf("From = %q, want bob", summary.From)
	}
	if summary.Body != "test message" {
		t.Errorf("Body = %q, want test message", summary.Body)
	}
}

func TestRunSubprocess_Start(t *testing.T) {
	// Test that start command triggers polling and message events are emitted.
	mock := &subprocessMockBot{}

	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()

	oldNewBot := newBotFunc
	newBotFunc = func(_ string) (BotAPI, error) { return mock, nil }
	t.Cleanup(func() { newBotFunc = oldNewBot })

	// Set up poll to emit one message then block until cancelled.
	oldPoll := subprocessPollFunc
	subprocessPollFunc = func(ctx context.Context, _ string, offset int64, handler func(message.Message, int64, bool), _ func(string, int, string, string, time.Time)) (int64, error) {
		handler(message.Message{
			ID:       "telegram-99",
			From:     "alice",
			Provider: "telegram",
			Channel:  "dev-team",
			Date:     time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC),
			Body:     "hello from poll",
		}, 123, false)
		<-ctx.Done()
		return offset + 1, nil
	}
	t.Cleanup(func() { subprocessPollFunc = oldPoll })

	errCh := make(chan error, 1)
	go func() {
		errCh <- RunSubprocess(context.Background(), stdinR, stdoutW, `{"token":"test-token"}`)
	}()

	stdout := bufio.NewReader(stdoutR)
	readEvent(t, stdout) // ready

	writeCmd(stdinW, protocol.StartCommand{Type: "start", Offset: 0})

	// Read the message event emitted by the poll handler.
	evt := readEvent(t, stdout)
	if evt["type"] != "message" {
		t.Fatalf("type = %v, want message", evt["type"])
	}
	if evt["from"] != "alice" {
		t.Errorf("from = %v, want alice", evt["from"])
	}
	if evt["body"] != "hello from poll" {
		t.Errorf("body = %v, want 'hello from poll'", evt["body"])
	}
	if evt["chat_id"] != float64(123) {
		t.Errorf("chat_id = %v, want 123", evt["chat_id"])
	}

	writeCmd(stdinW, protocol.ShutdownCommand{Type: "shutdown"})
	readEvent(t, stdout) // shutdown_complete
	stdinW.Close()
}

func TestRunSubprocess_InvalidConfig(t *testing.T) {
	var buf bytes.Buffer
	err := RunSubprocess(context.Background(), strings.NewReader(""), &buf, "not-json")
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}
