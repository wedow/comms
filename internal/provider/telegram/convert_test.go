package telegram

import (
	"testing"
	"time"

	"github.com/go-telegram/bot/models"
)

func TestConvertMessage(t *testing.T) {
	tests := []struct {
		name    string
		msg     *models.Message
		wantFrom        string
		wantProvider    string
		wantChannel     string
		wantDate        time.Time
		wantID          string
		wantBody        string
		wantReplyTo      string
		wantReplyToBody  string
		wantQuote        string
		wantMediaType    string
		wantMediaFileID  string
		wantCaption      string
		wantMediaGroupID string
		wantForwardFrom  string
		wantForwardDate  *time.Time
	}{
		{
			name: "user in group",
			msg: &models.Message{
				ID:   42,
				From: &models.User{Username: "alice"},
				Date: 1709312400,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Dev Team"},
				Text: "hello world",
			},
			wantFrom:     "alice",
			wantProvider: "telegram",
			wantChannel:  "dev-team",
			wantDate:     time.Unix(1709312400, 0).UTC(),
			wantID:       "telegram-42",
			wantBody:     "hello world",
		},
		{
			name: "user in DM",
			msg: &models.Message{
				ID:   7,
				From: &models.User{Username: "bob"},
				Date: 1000000,
				Chat: models.Chat{Type: models.ChatTypePrivate, Username: "bob"},
				Text: "hi",
			},
			wantFrom:     "bob",
			wantProvider: "telegram",
			wantChannel:  "bob",
			wantDate:     time.Unix(1000000, 0).UTC(),
			wantID:       "telegram-7",
			wantBody:     "hi",
		},
		{
			name: "empty text",
			msg: &models.Message{
				ID:   1,
				From: &models.User{Username: "carol"},
				Date: 100,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "General"},
			},
			wantFrom:     "carol",
			wantProvider: "telegram",
			wantChannel:  "general",
			wantDate:     time.Unix(100, 0).UTC(),
			wantID:       "telegram-1",
			wantBody:     "",
		},
		{
			name: "unix timestamp zero",
			msg: &models.Message{
				ID:   5,
				From: &models.User{Username: "dave"},
				Date: 0,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Test"},
			},
			wantFrom:     "dave",
			wantProvider: "telegram",
			wantChannel:  "test",
			wantDate:     time.Unix(0, 0).UTC(),
			wantID:       "telegram-5",
			wantBody:     "",
		},
		{
			name: "from nil",
			msg: &models.Message{
				ID:   99,
				Date: 500,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Ops"},
				Text: "system msg",
			},
			wantFrom:     "unknown",
			wantProvider: "telegram",
			wantChannel:  "ops",
			wantDate:     time.Unix(500, 0).UTC(),
			wantID:       "telegram-99",
			wantBody:     "system msg",
		},
		{
			name: "from with no username falls back to first name",
			msg: &models.Message{
				ID:   10,
				From: &models.User{FirstName: "Eve"},
				Date: 200,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Chat"},
				Text: "hey",
			},
			wantFrom:     "Eve",
			wantProvider: "telegram",
			wantChannel:  "chat",
			wantDate:     time.Unix(200, 0).UTC(),
			wantID:       "telegram-10",
			wantBody:     "hey",
		},
		{
			name: "reply to message",
			msg: &models.Message{
				ID:   50,
				From: &models.User{Username: "alice"},
				Date: 1000,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Dev"},
				Text: "my reply",
				ReplyToMessage: &models.Message{
					ID:   99,
					Text: "original",
				},
			},
			wantFrom:        "alice",
			wantProvider:    "telegram",
			wantChannel:     "dev",
			wantDate:        time.Unix(1000, 0).UTC(),
			wantID:          "telegram-50",
			wantBody:        "my reply",
			wantReplyTo:     "telegram-99",
			wantReplyToBody: "original",
		},
		{
			name: "reply with quote",
			msg: &models.Message{
				ID:   51,
				From: &models.User{Username: "bob"},
				Date: 2000,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Dev"},
				Text: "replying",
				ReplyToMessage: &models.Message{
					ID:   100,
					Text: "long message here",
				},
				Quote: &models.TextQuote{
					Text: "quoted part",
				},
			},
			wantFrom:        "bob",
			wantProvider:    "telegram",
			wantChannel:     "dev",
			wantDate:        time.Unix(2000, 0).UTC(),
			wantID:          "telegram-51",
			wantBody:        "replying",
			wantReplyTo:     "telegram-100",
			wantReplyToBody: "long message here",
			wantQuote:       "quoted part",
		},
		{
			name: "external reply",
			msg: &models.Message{
				ID:   52,
				From: &models.User{Username: "carol"},
				Date: 3000,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Dev"},
				Text: "ext reply",
				ExternalReply: &models.ExternalReplyInfo{
					MessageID: 77,
				},
			},
			wantFrom:     "carol",
			wantProvider: "telegram",
			wantChannel:  "dev",
			wantDate:     time.Unix(3000, 0).UTC(),
			wantID:       "telegram-52",
			wantBody:     "ext reply",
			wantReplyTo:  "telegram-77",
		},
		{
			name: "quote without reply",
			msg: &models.Message{
				ID:   53,
				From: &models.User{Username: "dave"},
				Date: 4000,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Dev"},
				Text: "just quoting",
				Quote: &models.TextQuote{
					Text: "some text",
				},
			},
			wantFrom:     "dave",
			wantProvider: "telegram",
			wantChannel:  "dev",
			wantDate:     time.Unix(4000, 0).UTC(),
			wantID:       "telegram-53",
			wantBody:     "just quoting",
			wantQuote:    "some text",
		},
		{
			name: "photo message",
			msg: &models.Message{
				ID:   60,
				From: &models.User{Username: "alice"},
				Date: 5000,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Photos"},
				Photo: []models.PhotoSize{
					{FileID: "small-id", Width: 100, Height: 100},
					{FileID: "large-id", Width: 800, Height: 600},
				},
				Caption: "nice photo",
			},
			wantFrom:        "alice",
			wantProvider:    "telegram",
			wantChannel:     "photos",
			wantDate:        time.Unix(5000, 0).UTC(),
			wantID:          "telegram-60",
			wantMediaType:   "photo",
			wantMediaFileID: "large-id",
			wantCaption:     "nice photo",
		},
		{
			name: "video message",
			msg: &models.Message{
				ID:   61,
				From: &models.User{Username: "bob"},
				Date: 6000,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Videos"},
				Video: &models.Video{FileID: "vid-id"},
			},
			wantFrom:        "bob",
			wantProvider:    "telegram",
			wantChannel:     "videos",
			wantDate:        time.Unix(6000, 0).UTC(),
			wantID:          "telegram-61",
			wantMediaType:   "video",
			wantMediaFileID: "vid-id",
		},
		{
			name: "document with caption",
			msg: &models.Message{
				ID:      62,
				From:    &models.User{Username: "carol"},
				Date:    7000,
				Chat:    models.Chat{Type: models.ChatTypeGroup, Title: "Files"},
				Document: &models.Document{FileID: "doc-id"},
				Caption: "read this",
			},
			wantFrom:        "carol",
			wantProvider:    "telegram",
			wantChannel:     "files",
			wantDate:        time.Unix(7000, 0).UTC(),
			wantID:          "telegram-62",
			wantMediaType:   "document",
			wantMediaFileID: "doc-id",
			wantCaption:     "read this",
		},
		{
			name: "voice message",
			msg: &models.Message{
				ID:   63,
				From: &models.User{Username: "dave"},
				Date: 8000,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Voice"},
				Voice: &models.Voice{FileID: "voice-id"},
			},
			wantFrom:        "dave",
			wantProvider:    "telegram",
			wantChannel:     "voice",
			wantDate:        time.Unix(8000, 0).UTC(),
			wantID:          "telegram-63",
			wantMediaType:   "voice",
			wantMediaFileID: "voice-id",
		},
		{
			name: "animation message",
			msg: &models.Message{
				ID:   64,
				From: &models.User{Username: "eve"},
				Date: 9000,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "GIFs"},
				Animation: &models.Animation{FileID: "anim-id"},
			},
			wantFrom:        "eve",
			wantProvider:    "telegram",
			wantChannel:     "gifs",
			wantDate:        time.Unix(9000, 0).UTC(),
			wantID:          "telegram-64",
			wantMediaType:   "animation",
			wantMediaFileID: "anim-id",
		},
		{
			name: "sticker message",
			msg: &models.Message{
				ID:   65,
				From: &models.User{Username: "frank"},
				Date: 10000,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Stickers"},
				Sticker: &models.Sticker{FileID: "sticker-id"},
			},
			wantFrom:        "frank",
			wantProvider:    "telegram",
			wantChannel:     "stickers",
			wantDate:        time.Unix(10000, 0).UTC(),
			wantID:          "telegram-65",
			wantMediaType:   "sticker",
			wantMediaFileID: "sticker-id",
		},
		{
			name: "album photo with media group",
			msg: &models.Message{
				ID:   66,
				From: &models.User{Username: "grace"},
				Date: 11000,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Album"},
				Photo: []models.PhotoSize{
					{FileID: "album-photo-id", Width: 800, Height: 600},
				},
				MediaGroupID: "group-123",
				Caption:      "album caption",
			},
			wantFrom:         "grace",
			wantProvider:     "telegram",
			wantChannel:      "album",
			wantDate:         time.Unix(11000, 0).UTC(),
			wantID:           "telegram-66",
			wantMediaType:    "photo",
			wantMediaFileID:  "album-photo-id",
			wantCaption:      "album caption",
			wantMediaGroupID: "group-123",
		},
		{
			name: "audio message",
			msg: &models.Message{
				ID:      67,
				From:    &models.User{Username: "heidi"},
				Date:    12000,
				Chat:    models.Chat{Type: models.ChatTypeGroup, Title: "Audio"},
				Audio:   &models.Audio{FileID: "audio-file-789"},
				Caption: "podcast episode",
			},
			wantFrom:        "heidi",
			wantProvider:    "telegram",
			wantChannel:     "audio",
			wantDate:        time.Unix(12000, 0).UTC(),
			wantID:          "telegram-67",
			wantMediaType:   "audio",
			wantMediaFileID: "audio-file-789",
			wantCaption:     "podcast episode",
		},
		{
			name: "video note message",
			msg: &models.Message{
				ID:        68,
				From:      &models.User{Username: "ivan"},
				Date:      13000,
				Chat:      models.Chat{Type: models.ChatTypeGroup, Title: "Notes"},
				VideoNote: &models.VideoNote{FileID: "vidnote-file-321"},
			},
			wantFrom:        "ivan",
			wantProvider:    "telegram",
			wantChannel:     "notes",
			wantDate:        time.Unix(13000, 0).UTC(),
			wantID:          "telegram-68",
			wantMediaType:   "video_note",
			wantMediaFileID: "vidnote-file-321",
		},
		{
			name: "channel post with sender_chat username",
			msg: &models.Message{
				ID:         101,
				Date:       19000,
				Chat:       models.Chat{Type: models.ChatTypeChannel, Title: "My Channel"},
				SenderChat: &models.Chat{Username: "mychannel", Title: "My Channel"},
				Text:       "channel post",
			},
			wantFrom:     "mychannel",
			wantProvider: "telegram",
			wantChannel:  "my-channel",
			wantDate:     time.Unix(19000, 0).UTC(),
			wantID:       "telegram-101",
			wantBody:     "channel post",
		},
		{
			name: "channel post with sender_chat title only",
			msg: &models.Message{
				ID:         102,
				Date:       20000,
				Chat:       models.Chat{Type: models.ChatTypeChannel, Title: "News Feed"},
				SenderChat: &models.Chat{Title: "News Feed"},
				Text:       "breaking news",
			},
			wantFrom:     "News Feed",
			wantProvider: "telegram",
			wantChannel:  "news-feed",
			wantDate:     time.Unix(20000, 0).UTC(),
			wantID:       "telegram-102",
			wantBody:     "breaking news",
		},
		{
			name: "forwarded from user",
			msg: &models.Message{
				ID:   70,
				From: &models.User{Username: "bob"},
				Date: 14000,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Dev"},
				Text: "fwd msg",
				ForwardOrigin: &models.MessageOrigin{
					Type: models.MessageOriginTypeUser,
					MessageOriginUser: &models.MessageOriginUser{
						Date:       1709300000,
						SenderUser: models.User{Username: "alice"},
					},
				},
			},
			wantFrom:        "bob",
			wantProvider:    "telegram",
			wantChannel:     "dev",
			wantDate:        time.Unix(14000, 0).UTC(),
			wantID:          "telegram-70",
			wantBody:        "fwd msg",
			wantForwardFrom: "alice",
			wantForwardDate: timePtr(time.Unix(1709300000, 0).UTC()),
		},
		{
			name: "forwarded from user without username",
			msg: &models.Message{
				ID:   71,
				From: &models.User{Username: "bob"},
				Date: 15000,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Dev"},
				Text: "fwd msg 2",
				ForwardOrigin: &models.MessageOrigin{
					Type: models.MessageOriginTypeUser,
					MessageOriginUser: &models.MessageOriginUser{
						Date:       1709300000,
						SenderUser: models.User{FirstName: "Bob"},
					},
				},
			},
			wantFrom:        "bob",
			wantProvider:    "telegram",
			wantChannel:     "dev",
			wantDate:        time.Unix(15000, 0).UTC(),
			wantID:          "telegram-71",
			wantBody:        "fwd msg 2",
			wantForwardFrom: "Bob",
			wantForwardDate: timePtr(time.Unix(1709300000, 0).UTC()),
		},
		{
			name: "forwarded from hidden user",
			msg: &models.Message{
				ID:   72,
				From: &models.User{Username: "carol"},
				Date: 16000,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Dev"},
				Text: "hidden fwd",
				ForwardOrigin: &models.MessageOrigin{
					Type: models.MessageOriginTypeHiddenUser,
					MessageOriginHiddenUser: &models.MessageOriginHiddenUser{
						Date:           1709300000,
						SenderUserName: "Anonymous",
					},
				},
			},
			wantFrom:        "carol",
			wantProvider:    "telegram",
			wantChannel:     "dev",
			wantDate:        time.Unix(16000, 0).UTC(),
			wantID:          "telegram-72",
			wantBody:        "hidden fwd",
			wantForwardFrom: "Anonymous",
			wantForwardDate: timePtr(time.Unix(1709300000, 0).UTC()),
		},
		{
			name: "forwarded from chat",
			msg: &models.Message{
				ID:   73,
				From: &models.User{Username: "dave"},
				Date: 17000,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Dev"},
				Text: "chat fwd",
				ForwardOrigin: &models.MessageOrigin{
					Type: models.MessageOriginTypeChat,
					MessageOriginChat: &models.MessageOriginChat{
						Date:       1709300000,
						SenderChat: models.Chat{Title: "My Group"},
					},
				},
			},
			wantFrom:        "dave",
			wantProvider:    "telegram",
			wantChannel:     "dev",
			wantDate:        time.Unix(17000, 0).UTC(),
			wantID:          "telegram-73",
			wantBody:        "chat fwd",
			wantForwardFrom: "My Group",
			wantForwardDate: timePtr(time.Unix(1709300000, 0).UTC()),
		},
		{
			name: "forwarded from channel",
			msg: &models.Message{
				ID:   74,
				From: &models.User{Username: "eve"},
				Date: 18000,
				Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Dev"},
				Text: "channel fwd",
				ForwardOrigin: &models.MessageOrigin{
					Type: models.MessageOriginTypeChannel,
					MessageOriginChannel: &models.MessageOriginChannel{
						Date: 1709300000,
						Chat: models.Chat{Title: "My Channel"},
					},
				},
			},
			wantFrom:        "eve",
			wantProvider:    "telegram",
			wantChannel:     "dev",
			wantDate:        time.Unix(18000, 0).UTC(),
			wantID:          "telegram-74",
			wantBody:        "channel fwd",
			wantForwardFrom: "My Channel",
			wantForwardDate: timePtr(time.Unix(1709300000, 0).UTC()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertMessage(tt.msg)
			if got.From != tt.wantFrom {
				t.Errorf("From = %q, want %q", got.From, tt.wantFrom)
			}
			if got.Provider != tt.wantProvider {
				t.Errorf("Provider = %q, want %q", got.Provider, tt.wantProvider)
			}
			if got.Channel != tt.wantChannel {
				t.Errorf("Channel = %q, want %q", got.Channel, tt.wantChannel)
			}
			if !got.Date.Equal(tt.wantDate) {
				t.Errorf("Date = %v, want %v", got.Date, tt.wantDate)
			}
			if got.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", got.ID, tt.wantID)
			}
			if got.Body != tt.wantBody {
				t.Errorf("Body = %q, want %q", got.Body, tt.wantBody)
			}
			if got.ReplyTo != tt.wantReplyTo {
				t.Errorf("ReplyTo = %q, want %q", got.ReplyTo, tt.wantReplyTo)
			}
			if got.ReplyToBody != tt.wantReplyToBody {
				t.Errorf("ReplyToBody = %q, want %q", got.ReplyToBody, tt.wantReplyToBody)
			}
			if got.Quote != tt.wantQuote {
				t.Errorf("Quote = %q, want %q", got.Quote, tt.wantQuote)
			}
			if got.MediaType != tt.wantMediaType {
				t.Errorf("MediaType = %q, want %q", got.MediaType, tt.wantMediaType)
			}
			if got.MediaFileID != tt.wantMediaFileID {
				t.Errorf("MediaFileID = %q, want %q", got.MediaFileID, tt.wantMediaFileID)
			}
			if got.Caption != tt.wantCaption {
				t.Errorf("Caption = %q, want %q", got.Caption, tt.wantCaption)
			}
			if got.MediaGroupID != tt.wantMediaGroupID {
				t.Errorf("MediaGroupID = %q, want %q", got.MediaGroupID, tt.wantMediaGroupID)
			}
			if got.ForwardFrom != tt.wantForwardFrom {
				t.Errorf("ForwardFrom = %q, want %q", got.ForwardFrom, tt.wantForwardFrom)
			}
			if tt.wantForwardDate == nil && got.ForwardDate != nil {
				t.Errorf("ForwardDate = %v, want nil", got.ForwardDate)
			} else if tt.wantForwardDate != nil && (got.ForwardDate == nil || !got.ForwardDate.Equal(*tt.wantForwardDate)) {
				t.Errorf("ForwardDate = %v, want %v", got.ForwardDate, tt.wantForwardDate)
			}
		})
	}
}

func TestConvertMessageEditDate(t *testing.T) {
	msg := &models.Message{
		ID:       80,
		From:     &models.User{Username: "alice"},
		Date:     1709312400,
		EditDate: 1709312500,
		Chat:     models.Chat{Type: models.ChatTypeGroup, Title: "Dev"},
		Text:     "edited text",
	}
	got := convertMessage(msg)
	if got.EditDate == nil {
		t.Fatal("EditDate is nil, want non-nil")
	}
	want := time.Unix(1709312500, 0).UTC()
	if !got.EditDate.Equal(want) {
		t.Errorf("EditDate = %v, want %v", *got.EditDate, want)
	}

	// Zero EditDate should remain nil
	msg2 := &models.Message{
		ID:   81,
		From: &models.User{Username: "bob"},
		Date: 1000,
		Chat: models.Chat{Type: models.ChatTypeGroup, Title: "Dev"},
		Text: "not edited",
	}
	got2 := convertMessage(msg2)
	if got2.EditDate != nil {
		t.Errorf("EditDate = %v, want nil", got2.EditDate)
	}
}

func timePtr(t time.Time) *time.Time { return &t }
