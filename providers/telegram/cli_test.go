package telegram

import (
	"bytes"
	"context"
	"encoding/json"
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
