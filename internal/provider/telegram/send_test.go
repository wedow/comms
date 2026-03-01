package telegram

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type mockBot struct {
	sendFn      func(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error)
	sendPhotoFn func(ctx context.Context, params *bot.SendPhotoParams) (*models.Message, error)
	sendDocFn   func(ctx context.Context, params *bot.SendDocumentParams) (*models.Message, error)
	sendAudioFn func(ctx context.Context, params *bot.SendAudioParams) (*models.Message, error)
	sendVideoFn func(ctx context.Context, params *bot.SendVideoParams) (*models.Message, error)
	sendVoiceFn func(ctx context.Context, params *bot.SendVoiceParams) (*models.Message, error)
	sendAnimFn  func(ctx context.Context, params *bot.SendAnimationParams) (*models.Message, error)
}

func (m *mockBot) SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error) {
	return m.sendFn(ctx, params)
}

func (m *mockBot) SendPhoto(ctx context.Context, params *bot.SendPhotoParams) (*models.Message, error) {
	if m.sendPhotoFn != nil {
		return m.sendPhotoFn(ctx, params)
	}
	return nil, nil
}

func (m *mockBot) SendDocument(ctx context.Context, params *bot.SendDocumentParams) (*models.Message, error) {
	if m.sendDocFn != nil {
		return m.sendDocFn(ctx, params)
	}
	return nil, nil
}

func (m *mockBot) SendAudio(ctx context.Context, params *bot.SendAudioParams) (*models.Message, error) {
	if m.sendAudioFn != nil {
		return m.sendAudioFn(ctx, params)
	}
	return nil, nil
}

func (m *mockBot) SendVideo(ctx context.Context, params *bot.SendVideoParams) (*models.Message, error) {
	if m.sendVideoFn != nil {
		return m.sendVideoFn(ctx, params)
	}
	return nil, nil
}

func (m *mockBot) SendVoice(ctx context.Context, params *bot.SendVoiceParams) (*models.Message, error) {
	if m.sendVoiceFn != nil {
		return m.sendVoiceFn(ctx, params)
	}
	return nil, nil
}

func (m *mockBot) SendAnimation(ctx context.Context, params *bot.SendAnimationParams) (*models.Message, error) {
	if m.sendAnimFn != nil {
		return m.sendAnimFn(ctx, params)
	}
	return nil, nil
}

func (m *mockBot) SetMessageReaction(_ context.Context, _ *bot.SetMessageReactionParams) (bool, error) {
	return false, nil
}

func (m *mockBot) GetFile(_ context.Context, _ *bot.GetFileParams) (*models.File, error) {
	return nil, nil
}

func (m *mockBot) FileDownloadLink(_ *models.File) string { return "" }

func TestSend(t *testing.T) {
	t.Run("successful send", func(t *testing.T) {
		m := &mockBot{sendFn: func(_ context.Context, _ *bot.SendMessageParams) (*models.Message, error) {
			return &models.Message{
				ID:   42,
				From: &models.User{Username: "botuser"},
				Date: 1709312400,
				Chat: models.Chat{ID: 123, Type: models.ChatTypeGroup, Title: "Dev Team"},
				Text: "hello",
			}, nil
		}}

		got, err := Send(context.Background(), m, 123, "hello", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.From != "botuser" {
			t.Errorf("From = %q, want %q", got.From, "botuser")
		}
		if got.Provider != "telegram" {
			t.Errorf("Provider = %q, want %q", got.Provider, "telegram")
		}
		if got.Channel != "dev-team" {
			t.Errorf("Channel = %q, want %q", got.Channel, "dev-team")
		}
		if !got.Date.Equal(time.Unix(1709312400, 0).UTC()) {
			t.Errorf("Date = %v, want %v", got.Date, time.Unix(1709312400, 0).UTC())
		}
		if got.ID != "telegram-42" {
			t.Errorf("ID = %q, want %q", got.ID, "telegram-42")
		}
		if got.Body != "hello" {
			t.Errorf("Body = %q, want %q", got.Body, "hello")
		}
	})

	t.Run("api error", func(t *testing.T) {
		m := &mockBot{sendFn: func(_ context.Context, _ *bot.SendMessageParams) (*models.Message, error) {
			return nil, errors.New("network timeout")
		}}

		_, err := Send(context.Background(), m, 456, "fail", 0)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("nil response", func(t *testing.T) {
		m := &mockBot{sendFn: func(_ context.Context, _ *bot.SendMessageParams) (*models.Message, error) {
			return nil, nil
		}}

		_, err := Send(context.Background(), m, 789, "nil", 0)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("with reply", func(t *testing.T) {
		var gotParams *bot.SendMessageParams
		m := &mockBot{sendFn: func(_ context.Context, p *bot.SendMessageParams) (*models.Message, error) {
			gotParams = p
			return &models.Message{ID: 99, Chat: models.Chat{ID: 123}, Text: "reply"}, nil
		}}

		_, err := Send(context.Background(), m, 123, "reply", 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotParams.ReplyParameters == nil {
			t.Fatal("ReplyParameters is nil, want non-nil")
		}
		if gotParams.ReplyParameters.MessageID != 42 {
			t.Errorf("ReplyParameters.MessageID = %d, want 42", gotParams.ReplyParameters.MessageID)
		}
	})

	t.Run("without reply", func(t *testing.T) {
		var gotParams *bot.SendMessageParams
		m := &mockBot{sendFn: func(_ context.Context, p *bot.SendMessageParams) (*models.Message, error) {
			gotParams = p
			return &models.Message{ID: 100, Chat: models.Chat{ID: 123}, Text: "no reply"}, nil
		}}

		_, err := Send(context.Background(), m, 123, "no reply", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotParams.ReplyParameters != nil {
			t.Errorf("ReplyParameters = %+v, want nil", gotParams.ReplyParameters)
		}
	})
}

func TestDetectMediaType(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"photo.jpg", "photo"},
		{"photo.jpeg", "photo"},
		{"photo.png", "photo"},
		{"photo.webp", "photo"},
		{"anim.gif", "animation"},
		{"video.mp4", "video"},
		{"video.mov", "video"},
		{"video.avi", "video"},
		{"song.mp3", "audio"},
		{"song.flac", "audio"},
		{"song.wav", "audio"},
		{"voice.ogg", "voice"},
		{"file.pdf", "document"},
		{"file.txt", "document"},
		{"file.zip", "document"},
		{"PHOTO.JPG", "photo"},
		{"noext", "document"},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := DetectMediaType(tt.filename)
			if got != tt.want {
				t.Errorf("DetectMediaType(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestSendMedia(t *testing.T) {
	stubResp := &models.Message{
		ID:   10,
		Chat: models.Chat{ID: 123, Type: models.ChatTypePrivate},
		From: &models.User{Username: "testbot"},
	}

	t.Run("photo dispatch", func(t *testing.T) {
		called := false
		m := &mockBot{sendPhotoFn: func(_ context.Context, p *bot.SendPhotoParams) (*models.Message, error) {
			called = true
			if p.ChatID != int64(123) {
				t.Errorf("ChatID = %v, want 123", p.ChatID)
			}
			if p.Caption != "my caption" {
				t.Errorf("Caption = %q, want %q", p.Caption, "my caption")
			}
			return stubResp, nil
		}}
		got, err := SendMedia(context.Background(), m, 123, strings.NewReader("data"), "pic.jpg", "photo", "my caption", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("SendPhoto was not called")
		}
		if got.ID != "telegram-10" {
			t.Errorf("ID = %q, want %q", got.ID, "telegram-10")
		}
	})

	t.Run("document dispatch", func(t *testing.T) {
		called := false
		m := &mockBot{sendDocFn: func(_ context.Context, p *bot.SendDocumentParams) (*models.Message, error) {
			called = true
			if p.ChatID != int64(456) {
				t.Errorf("ChatID = %v, want 456", p.ChatID)
			}
			return stubResp, nil
		}}
		_, err := SendMedia(context.Background(), m, 456, strings.NewReader("data"), "file.pdf", "document", "", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("SendDocument was not called")
		}
	})

	t.Run("audio dispatch", func(t *testing.T) {
		called := false
		m := &mockBot{sendAudioFn: func(_ context.Context, _ *bot.SendAudioParams) (*models.Message, error) {
			called = true
			return stubResp, nil
		}}
		_, err := SendMedia(context.Background(), m, 123, strings.NewReader("data"), "song.mp3", "audio", "", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("SendAudio was not called")
		}
	})

	t.Run("video dispatch", func(t *testing.T) {
		called := false
		m := &mockBot{sendVideoFn: func(_ context.Context, _ *bot.SendVideoParams) (*models.Message, error) {
			called = true
			return stubResp, nil
		}}
		_, err := SendMedia(context.Background(), m, 123, strings.NewReader("data"), "clip.mp4", "video", "", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("SendVideo was not called")
		}
	})

	t.Run("voice dispatch", func(t *testing.T) {
		called := false
		m := &mockBot{sendVoiceFn: func(_ context.Context, _ *bot.SendVoiceParams) (*models.Message, error) {
			called = true
			return stubResp, nil
		}}
		_, err := SendMedia(context.Background(), m, 123, strings.NewReader("data"), "msg.ogg", "voice", "", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("SendVoice was not called")
		}
	})

	t.Run("animation dispatch", func(t *testing.T) {
		called := false
		m := &mockBot{sendAnimFn: func(_ context.Context, _ *bot.SendAnimationParams) (*models.Message, error) {
			called = true
			return stubResp, nil
		}}
		_, err := SendMedia(context.Background(), m, 123, strings.NewReader("data"), "funny.gif", "animation", "", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("SendAnimation was not called")
		}
	})

	t.Run("with reply", func(t *testing.T) {
		m := &mockBot{sendPhotoFn: func(_ context.Context, p *bot.SendPhotoParams) (*models.Message, error) {
			if p.ReplyParameters == nil {
				t.Fatal("ReplyParameters is nil, want non-nil")
			}
			if p.ReplyParameters.MessageID != 42 {
				t.Errorf("ReplyParameters.MessageID = %d, want 42", p.ReplyParameters.MessageID)
			}
			return stubResp, nil
		}}
		_, err := SendMedia(context.Background(), m, 123, strings.NewReader("data"), "pic.jpg", "photo", "", 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("api error", func(t *testing.T) {
		m := &mockBot{sendPhotoFn: func(_ context.Context, _ *bot.SendPhotoParams) (*models.Message, error) {
			return nil, errors.New("upload failed")
		}}
		_, err := SendMedia(context.Background(), m, 123, strings.NewReader("data"), "pic.jpg", "photo", "", 0)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("unknown media type falls back to document", func(t *testing.T) {
		called := false
		m := &mockBot{sendDocFn: func(_ context.Context, _ *bot.SendDocumentParams) (*models.Message, error) {
			called = true
			return stubResp, nil
		}}
		_, err := SendMedia(context.Background(), m, 123, strings.NewReader("data"), "file.xyz", "unknown", "", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("SendDocument was not called for unknown type")
		}
	})
}

func TestParseMessageID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"valid", "telegram-42", 42, false},
		{"no prefix", "abc", 0, true},
		{"empty suffix", "telegram-", 0, true},
		{"non-numeric suffix", "telegram-abc", 0, true},
		{"empty string", "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMessageID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseMessageID(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseMessageID(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
