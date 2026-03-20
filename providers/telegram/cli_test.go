package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/spf13/cobra"
)

func TestLoadProviderConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		t.Setenv("COMMS_PROVIDER_CONFIG", `{"token":"test-tok"}`)
		cfg, err := loadProviderConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Token != "test-tok" {
			t.Errorf("Token = %q, want %q", cfg.Token, "test-tok")
		}
	})

	t.Run("missing env var", func(t *testing.T) {
		t.Setenv("COMMS_PROVIDER_CONFIG", "")
		_, err := loadProviderConfig()
		if err == nil {
			t.Fatal("expected error for empty env var")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		t.Setenv("COMMS_PROVIDER_CONFIG", "not-json")
		_, err := loadProviderConfig()
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestParseFormatFlag(t *testing.T) {
	tests := []struct {
		flag    string
		want    models.ParseMode
		wantErr bool
	}{
		{"", "", false},
		{"plain", "", false},
		{"markdown", models.ParseModeMarkdown, false},
		{"html", models.ParseModeHTML, false},
		{"invalid", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("format", tt.flag, "")
			got, err := parseFormatFlag(cmd)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseFormatFlag(%q) error = %v, wantErr %v", tt.flag, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("parseFormatFlag(%q) = %q, want %q", tt.flag, got, tt.want)
			}
		})
	}
}

func TestNewSendCmd_TextMessage(t *testing.T) {
	var gotParams *bot.SendMessageParams
	mock := &subprocessMockBot{
		mockBot: mockBot{
			sendFn: func(_ context.Context, p *bot.SendMessageParams) (*models.Message, error) {
				gotParams = p
				return &models.Message{
					ID:   42,
					Chat: models.Chat{ID: 100, Type: models.ChatTypeGroup, Title: "Test"},
					From: &models.User{Username: "testbot"},
					Date: 1709312400,
					Text: "hello world",
				}, nil
			},
		},
	}

	old := newBotFunc
	newBotFunc = func(_ string) (BotAPI, error) { return mock, nil }
	t.Cleanup(func() { newBotFunc = old })
	t.Setenv("COMMS_PROVIDER_CONFIG", `{"token":"test-tok"}`)

	var stdout bytes.Buffer
	cmd := NewSendCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--chat-id", "100", "hello", "world"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if gotParams == nil {
		t.Fatal("SendMessage was not called")
	}
	if gotParams.ChatID != int64(100) {
		t.Errorf("ChatID = %v, want 100", gotParams.ChatID)
	}
	if gotParams.Text != "hello world" {
		t.Errorf("Text = %q, want %q", gotParams.Text, "hello world")
	}

	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result["ok"] != true {
		t.Errorf("ok = %v, want true", result["ok"])
	}
	if result["message_id"] != float64(42) {
		t.Errorf("message_id = %v, want 42", result["message_id"])
	}
}

func TestNewSendCmd_WithFormat(t *testing.T) {
	var gotParams *bot.SendMessageParams
	mock := &subprocessMockBot{
		mockBot: mockBot{
			sendFn: func(_ context.Context, p *bot.SendMessageParams) (*models.Message, error) {
				gotParams = p
				return &models.Message{
					ID:   43,
					Chat: models.Chat{ID: 100},
					Text: "*bold*",
				}, nil
			},
		},
	}

	old := newBotFunc
	newBotFunc = func(_ string) (BotAPI, error) { return mock, nil }
	t.Cleanup(func() { newBotFunc = old })
	t.Setenv("COMMS_PROVIDER_CONFIG", `{"token":"test-tok"}`)

	var stdout bytes.Buffer
	cmd := NewSendCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--chat-id", "100", "--format", "markdown", "*bold*"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if gotParams.ParseMode != models.ParseModeMarkdown {
		t.Errorf("ParseMode = %q, want %q", gotParams.ParseMode, models.ParseModeMarkdown)
	}
}

func TestNewSendCmd_MissingConfig(t *testing.T) {
	t.Setenv("COMMS_PROVIDER_CONFIG", "")

	var stderr bytes.Buffer
	cmd := NewSendCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--chat-id", "100", "hello"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestNewReactCmd(t *testing.T) {
	reactionCalled := false
	mock := &subprocessMockBot{
		setMessageReactionFn: func(_ context.Context, p *bot.SetMessageReactionParams) (bool, error) {
			reactionCalled = true
			if p.ChatID != int64(200) {
				t.Errorf("ChatID = %v, want 200", p.ChatID)
			}
			if p.MessageID != 55 {
				t.Errorf("MessageID = %v, want 55", p.MessageID)
			}
			return true, nil
		},
	}

	old := newBotFunc
	newBotFunc = func(_ string) (BotAPI, error) { return mock, nil }
	t.Cleanup(func() { newBotFunc = old })
	t.Setenv("COMMS_PROVIDER_CONFIG", `{"token":"test-tok"}`)

	var stdout bytes.Buffer
	cmd := NewReactCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--chat-id", "200", "--message", "telegram-55", "--emoji", "\u0001f44d"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !reactionCalled {
		t.Error("SetMessageReaction was not called")
	}

	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result["ok"] != true {
		t.Errorf("ok = %v, want true", result["ok"])
	}
}

func TestNewReactCmd_MissingConfig(t *testing.T) {
	t.Setenv("COMMS_PROVIDER_CONFIG", "")

	var stderr bytes.Buffer
	cmd := NewReactCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--chat-id", "200", "--message", "telegram-55", "--emoji", "\u0001f44d"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestNewSendCmd_WithFile(t *testing.T) {
	// Create a temporary file to send.
	dir := t.TempDir()
	filePath := filepath.Join(dir, "photo.png")
	if err := os.WriteFile(filePath, []byte("fake-png-data"), 0644); err != nil {
		t.Fatal(err)
	}

	var gotParams *bot.SendPhotoParams
	mock := &subprocessMockBot{
		mockBot: mockBot{
			sendPhotoFn: func(_ context.Context, p *bot.SendPhotoParams) (*models.Message, error) {
				gotParams = p
				return &models.Message{
					ID:   50,
					Chat: models.Chat{ID: 100},
					From: &models.User{Username: "testbot"},
				}, nil
			},
		},
	}

	old := newBotFunc
	newBotFunc = func(_ string) (BotAPI, error) { return mock, nil }
	t.Cleanup(func() { newBotFunc = old })
	t.Setenv("COMMS_PROVIDER_CONFIG", `{"token":"test-tok"}`)

	var stdout bytes.Buffer
	cmd := NewSendCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--chat-id", "100", "--file", filePath, "my caption"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if gotParams == nil {
		t.Fatal("SendPhoto was not called")
	}
	if gotParams.ChatID != int64(100) {
		t.Errorf("ChatID = %v, want 100", gotParams.ChatID)
	}
	if gotParams.Caption != "my caption" {
		t.Errorf("Caption = %q, want %q", gotParams.Caption, "my caption")
	}

	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result["ok"] != true {
		t.Errorf("ok = %v, want true", result["ok"])
	}
	if result["message_id"] != float64(50) {
		t.Errorf("message_id = %v, want 50", result["message_id"])
	}
}

func TestNewSendCmd_WithReplyTo(t *testing.T) {
	var gotParams *bot.SendMessageParams
	mock := &subprocessMockBot{
		mockBot: mockBot{
			sendFn: func(_ context.Context, p *bot.SendMessageParams) (*models.Message, error) {
				gotParams = p
				return &models.Message{
					ID:   44,
					Chat: models.Chat{ID: 100},
					Text: "reply text",
				}, nil
			},
		},
	}

	old := newBotFunc
	newBotFunc = func(_ string) (BotAPI, error) { return mock, nil }
	t.Cleanup(func() { newBotFunc = old })
	t.Setenv("COMMS_PROVIDER_CONFIG", `{"token":"test-tok"}`)

	var stdout bytes.Buffer
	cmd := NewSendCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--chat-id", "100", "--reply-to", "telegram-77", "reply text"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if gotParams == nil {
		t.Fatal("SendMessage was not called")
	}
	if gotParams.ReplyParameters == nil {
		t.Fatal("ReplyParameters is nil, want non-nil")
	}
	if gotParams.ReplyParameters.MessageID != 77 {
		t.Errorf("ReplyParameters.MessageID = %d, want 77", gotParams.ReplyParameters.MessageID)
	}
}

func TestNewSendCmd_WithThread(t *testing.T) {
	var gotParams *bot.SendMessageParams
	mock := &subprocessMockBot{
		mockBot: mockBot{
			sendFn: func(_ context.Context, p *bot.SendMessageParams) (*models.Message, error) {
				gotParams = p
				return &models.Message{
					ID:   45,
					Chat: models.Chat{ID: 100},
					Text: "threaded msg",
				}, nil
			},
		},
	}

	old := newBotFunc
	newBotFunc = func(_ string) (BotAPI, error) { return mock, nil }
	t.Cleanup(func() { newBotFunc = old })
	t.Setenv("COMMS_PROVIDER_CONFIG", `{"token":"test-tok"}`)

	var stdout bytes.Buffer
	cmd := NewSendCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--chat-id", "100", "--thread", "12", "threaded msg"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if gotParams == nil {
		t.Fatal("SendMessage was not called")
	}
	if gotParams.MessageThreadID != 12 {
		t.Errorf("MessageThreadID = %d, want 12", gotParams.MessageThreadID)
	}
}

func TestNewSendCmd_APIError(t *testing.T) {
	mock := &subprocessMockBot{
		mockBot: mockBot{
			sendFn: func(_ context.Context, _ *bot.SendMessageParams) (*models.Message, error) {
				return nil, errors.New("telegram API unavailable")
			},
		},
	}

	old := newBotFunc
	newBotFunc = func(_ string) (BotAPI, error) { return mock, nil }
	t.Cleanup(func() { newBotFunc = old })
	t.Setenv("COMMS_PROVIDER_CONFIG", `{"token":"test-tok"}`)

	var stderr bytes.Buffer
	cmd := NewSendCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--chat-id", "100", "hello"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error from API failure")
	}

	var result map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal stderr: %v", err)
	}
	errMsg, _ := result["error"].(string)
	if errMsg == "" {
		t.Error("expected non-empty error in stderr JSON")
	}
}

func TestNewReactCmd_APIError(t *testing.T) {
	mock := &subprocessMockBot{
		setMessageReactionFn: func(_ context.Context, _ *bot.SetMessageReactionParams) (bool, error) {
			return false, errors.New("reaction rejected")
		},
	}

	old := newBotFunc
	newBotFunc = func(_ string) (BotAPI, error) { return mock, nil }
	t.Cleanup(func() { newBotFunc = old })
	t.Setenv("COMMS_PROVIDER_CONFIG", `{"token":"test-tok"}`)

	var stderr bytes.Buffer
	cmd := NewReactCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--chat-id", "200", "--message", "telegram-55", "--emoji", "\u0001f44d"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error from API failure")
	}

	var result map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal stderr: %v", err)
	}
	errMsg, _ := result["error"].(string)
	if !strings.Contains(errMsg, "reaction rejected") {
		t.Errorf("error = %q, want it to contain 'reaction rejected'", errMsg)
	}
}

func TestNewReactCmd_InvalidMessageID(t *testing.T) {
	mock := &subprocessMockBot{}

	old := newBotFunc
	newBotFunc = func(_ string) (BotAPI, error) { return mock, nil }
	t.Cleanup(func() { newBotFunc = old })
	t.Setenv("COMMS_PROVIDER_CONFIG", `{"token":"test-tok"}`)

	var stderr bytes.Buffer
	cmd := NewReactCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	// Pass a message ID without the "telegram-" prefix.
	cmd.SetArgs([]string{"--chat-id", "200", "--message", "bad-format", "--emoji", "\u0001f44d"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for invalid message ID format")
	}

	var result map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal stderr: %v", err)
	}
	errMsg, _ := result["error"].(string)
	if errMsg == "" {
		t.Error("expected non-empty error in stderr JSON")
	}
}

func TestNewSendCmd_WithMediaTypeOverride(t *testing.T) {
	// Create a .png file but override media type to "document".
	dir := t.TempDir()
	filePath := filepath.Join(dir, "image.png")
	if err := os.WriteFile(filePath, []byte("fake-png"), 0644); err != nil {
		t.Fatal(err)
	}

	var docCalled bool
	mock := &subprocessMockBot{
		mockBot: mockBot{
			sendDocFn: func(_ context.Context, p *bot.SendDocumentParams) (*models.Message, error) {
				docCalled = true
				if p.ChatID != int64(100) {
					t.Errorf("ChatID = %v, want 100", p.ChatID)
				}
				return &models.Message{
					ID:   51,
					Chat: models.Chat{ID: 100},
				}, nil
			},
		},
	}

	old := newBotFunc
	newBotFunc = func(_ string) (BotAPI, error) { return mock, nil }
	t.Cleanup(func() { newBotFunc = old })
	t.Setenv("COMMS_PROVIDER_CONFIG", `{"token":"test-tok"}`)

	var stdout bytes.Buffer
	cmd := NewSendCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--chat-id", "100", "--file", filePath, "--media-type", "document", "caption"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !docCalled {
		t.Error("SendDocument was not called despite --media-type document override")
	}
}

func TestNewSendCmd_FileNotFound(t *testing.T) {
	mock := &subprocessMockBot{}

	old := newBotFunc
	newBotFunc = func(_ string) (BotAPI, error) { return mock, nil }
	t.Cleanup(func() { newBotFunc = old })
	t.Setenv("COMMS_PROVIDER_CONFIG", `{"token":"test-tok"}`)

	var stderr bytes.Buffer
	cmd := NewSendCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--chat-id", "100", "--file", "/nonexistent/path/photo.png"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for nonexistent file")
	}

	var result map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal stderr: %v", err)
	}
	errMsg, _ := result["error"].(string)
	if !strings.Contains(errMsg, "open file") {
		t.Errorf("error = %q, want it to contain 'open file'", errMsg)
	}
}

func TestNewSendCmd_FileWithEmptyBody(t *testing.T) {
	// Send a file with no body args (caption-less media).
	dir := t.TempDir()
	filePath := filepath.Join(dir, "doc.pdf")
	if err := os.WriteFile(filePath, []byte("fake-pdf"), 0644); err != nil {
		t.Fatal(err)
	}

	var gotParams *bot.SendDocumentParams
	mock := &subprocessMockBot{
		mockBot: mockBot{
			sendDocFn: func(_ context.Context, p *bot.SendDocumentParams) (*models.Message, error) {
				gotParams = p
				return &models.Message{
					ID:   52,
					Chat: models.Chat{ID: 100},
				}, nil
			},
		},
	}

	old := newBotFunc
	newBotFunc = func(_ string) (BotAPI, error) { return mock, nil }
	t.Cleanup(func() { newBotFunc = old })
	t.Setenv("COMMS_PROVIDER_CONFIG", `{"token":"test-tok"}`)

	var stdout bytes.Buffer
	cmd := NewSendCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	// No body args, just --file. Use stdin with empty content.
	cmd.SetIn(strings.NewReader(""))
	cmd.SetArgs([]string{"--chat-id", "100", "--file", filePath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if gotParams == nil {
		t.Fatal("SendDocument was not called")
	}
	if gotParams.Caption != "" {
		t.Errorf("Caption = %q, want empty", gotParams.Caption)
	}

	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if result["ok"] != true {
		t.Errorf("ok = %v, want true", result["ok"])
	}
}
