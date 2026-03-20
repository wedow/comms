package protocol

import "time"

// Type constants for all protocol messages.
const (
	TypeReady            = "ready"
	TypeMessage          = "message"
	TypeEdit             = "edit"
	TypeReaction         = "reaction"
	TypeResponse         = "response"
	TypeError            = "error"
	TypeShutdownComplete = "shutdown_complete"
	TypePing             = "ping"
	TypePong             = "pong"
	TypeStart            = "start"
	TypeSend             = "send"
	TypeSendMedia        = "send_media"
	TypeReact            = "react"
	TypeTyping           = "typing"
	TypeShutdown         = "shutdown"
)

// --- Provider-to-Daemon Events ---

type ReadyEvent struct {
	Type     string `json:"type"`
	Provider string `json:"provider"`
	Version  string `json:"version"`
}

type MessageEvent struct {
	Type         string    `json:"type"`
	Offset       int64     `json:"offset"`
	ID           int       `json:"id"`
	ChatID       int64     `json:"chat_id"`
	Channel      string    `json:"channel"`
	From         string    `json:"from"`
	Date         time.Time `json:"date"`
	Body         string    `json:"body"`
	ReplyTo      int       `json:"reply_to,omitempty"`
	ReplyToBody  string    `json:"reply_to_body,omitempty"`
	Quote        string    `json:"quote,omitempty"`
	ThreadID     int       `json:"thread_id,omitempty"`
	MediaType    string    `json:"media_type,omitempty"`
	MediaFileID  string    `json:"media_file_id,omitempty"`
	DownloadURL  string    `json:"download_url,omitempty"`
	MediaExt     string    `json:"media_ext,omitempty"`
	Caption      string    `json:"caption,omitempty"`
	ForwardFrom  string     `json:"forward_from,omitempty"`
	ForwardDate  *time.Time `json:"forward_date,omitempty"`
	EditDate     *time.Time `json:"edit_date,omitempty"`
	MediaGroupID string    `json:"media_group_id,omitempty"`
	Entities     []Entity  `json:"entities,omitempty"`
}

// EditEvent is a MessageEvent with type "edit"; discrimination is via the Type field.
type EditEvent = MessageEvent

type ReactionEvent struct {
	Type      string    `json:"type"`
	Offset    int64     `json:"offset"`
	Channel   string    `json:"channel"`
	MessageID int       `json:"message_id"`
	From      string    `json:"from"`
	Emoji     string    `json:"emoji"`
	Date      time.Time `json:"date"`
}

type ResponseEvent struct {
	Type    string      `json:"type"`
	ID      string      `json:"id"`
	OK      bool        `json:"ok"`
	Message *MsgSummary `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ErrorEvent struct {
	Type    string `json:"type"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ShutdownCompleteEvent struct {
	Type string `json:"type"`
}

type PingEvent struct {
	Type string    `json:"type"`
	TS   time.Time `json:"ts"`
}

type PongEvent struct {
	Type string    `json:"type"`
	TS   time.Time `json:"ts"`
}

// --- Daemon-to-Provider Commands ---

type StartCommand struct {
	Type   string `json:"type"`
	Offset int64  `json:"offset"`
}

type SendCommand struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
	ReplyToID int    `json:"reply_to_id,omitempty"`
	ThreadID  int    `json:"thread_id,omitempty"`
}

type SendMediaCommand struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	ChatID    int64  `json:"chat_id"`
	MediaType string `json:"media_type"`
	Path      string `json:"path"`
	Filename  string `json:"filename,omitempty"`
	Caption   string `json:"caption,omitempty"`
	ParseMode string `json:"parse_mode,omitempty"`
	ReplyToID int    `json:"reply_to_id,omitempty"`
	ThreadID  int    `json:"thread_id,omitempty"`
}

type ReactCommand struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	ChatID    int64  `json:"chat_id"`
	MessageID int    `json:"message_id"`
	Emoji     string `json:"emoji"`
}

type TypingCommand struct {
	Type   string `json:"type"`
	ChatID int64  `json:"chat_id"`
}

type ShutdownCommand struct {
	Type   string `json:"type"`
	Reason string `json:"reason,omitempty"`
}

// --- Supporting Types ---

type MsgSummary struct {
	ID      int       `json:"id"`
	ChatID  int64     `json:"chat_id"`
	Channel string    `json:"channel"`
	From    string    `json:"from"`
	Date    time.Time `json:"date"`
	Body    string    `json:"body"`
}

type Entity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	URL    string `json:"url,omitempty"`
}
