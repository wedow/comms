package telegram

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/wedow/comms/internal/message"
	"github.com/wedow/comms/internal/protocol"
)

// ProviderConfig holds the JSON configuration for the telegram provider.
type ProviderConfig struct {
	Token string `json:"token"`
}

// Swappable functions for testing.
var (
	newBotFunc         = NewBot
	subprocessPollFunc = Poll
)

type subprocess struct {
	api    BotAPI
	stdout io.Writer
	mu     sync.Mutex
	cancel func() // cancels the poll goroutine
}

func (s *subprocess) write(msg any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	protocol.Encode(s.stdout, msg)
}

// RunSubprocess implements the provider-side JSONL protocol loop.
// It reads commands from stdin and writes events to stdout.
func RunSubprocess(ctx context.Context, stdin io.Reader, stdout io.Writer, configJSON string) error {
	var cfg ProviderConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	api, err := newBotFunc(cfg.Token)
	if err != nil {
		protocol.Encode(stdout, protocol.ErrorEvent{
			Type:    "error",
			Code:    1,
			Message: err.Error(),
		})
		return fmt.Errorf("create bot: %w", err)
	}

	s := &subprocess{
		api:    api,
		stdout: stdout,
	}

	s.write(protocol.ReadyEvent{
		Type:     "ready",
		Provider: "telegram",
		Version:  "1",
	})

	reader := bufio.NewReader(stdin)
	for {
		msg, err := protocol.DecodeTyped(reader)
		if err != nil {
			// EOF means the daemon closed stdin - treat as clean shutdown.
			if err == io.EOF {
				s.cancelPoll()
				return nil
			}
			return fmt.Errorf("decode: %w", err)
		}

		switch cmd := msg.(type) {
		case protocol.StartCommand:
			s.handleStart(ctx, cfg.Token, cmd.Offset)
		case protocol.SendCommand:
			s.handleSend(ctx, cmd)
		case protocol.SendMediaCommand:
			s.handleSendMedia(ctx, cmd)
		case protocol.ReactCommand:
			s.handleReact(ctx, cmd)
		case protocol.TypingCommand:
			s.api.SendChatAction(ctx, &bot.SendChatActionParams{
				ChatID: cmd.ChatID,
				Action: models.ChatActionTyping,
			})
		case protocol.ShutdownCommand:
			s.cancelPoll()
			s.write(protocol.ShutdownCompleteEvent{Type: "shutdown_complete"})
			return nil
		case protocol.PingEvent:
			s.write(protocol.PongEvent{Type: "pong", TS: cmd.TS})
		}
	}
}

func (s *subprocess) cancelPoll() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *subprocess) handleStart(ctx context.Context, token string, offset int64) {
	pollCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	go func() {
		subprocessPollFunc(pollCtx, token, offset,
			func(msg message.Message, chatID int64, isEdit bool) {
				evt := messageToEvent(msg, chatID, isEdit)
				s.write(evt)
			},
			func(channel string, msgID int, from string, emoji string, date time.Time) {
				s.write(protocol.ReactionEvent{
					Type:      "reaction",
					Channel:   channel,
					MessageID: msgID,
					From:      from,
					Emoji:     emoji,
					Date:      date,
				})
			},
		)
	}()
}

func (s *subprocess) handleSend(ctx context.Context, cmd protocol.SendCommand) {
	msg, err := Send(ctx, s.api, cmd.ChatID, cmd.Text, cmd.ReplyToID, cmd.ThreadID, models.ParseMode(cmd.ParseMode))
	if err != nil {
		s.write(protocol.ResponseEvent{
			Type:  "response",
			ID:    cmd.ID,
			OK:    false,
			Error: err.Error(),
		})
		return
	}
	s.write(protocol.ResponseEvent{
		Type:    "response",
		ID:      cmd.ID,
		OK:      true,
		Message: messageToSummary(msg),
	})
}

func (s *subprocess) handleSendMedia(ctx context.Context, cmd protocol.SendMediaCommand) {
	f, err := os.Open(cmd.Path)
	if err != nil {
		s.write(protocol.ResponseEvent{
			Type:  "response",
			ID:    cmd.ID,
			OK:    false,
			Error: err.Error(),
		})
		return
	}
	defer f.Close()

	filename := cmd.Filename
	if filename == "" {
		filename = cmd.Path
	}

	msg, err := SendMedia(ctx, s.api, cmd.ChatID, f, filename, cmd.MediaType, cmd.Caption, cmd.ReplyToID, cmd.ThreadID, models.ParseMode(cmd.ParseMode))
	if err != nil {
		s.write(protocol.ResponseEvent{
			Type:  "response",
			ID:    cmd.ID,
			OK:    false,
			Error: err.Error(),
		})
		return
	}
	s.write(protocol.ResponseEvent{
		Type:    "response",
		ID:      cmd.ID,
		OK:      true,
		Message: messageToSummary(msg),
	})
}

func (s *subprocess) handleReact(ctx context.Context, cmd protocol.ReactCommand) {
	_, err := s.api.SetMessageReaction(ctx, &bot.SetMessageReactionParams{
		ChatID:    cmd.ChatID,
		MessageID: cmd.MessageID,
		Reaction: []models.ReactionType{{
			Type:              models.ReactionTypeTypeEmoji,
			ReactionTypeEmoji: &models.ReactionTypeEmoji{Emoji: cmd.Emoji},
		}},
	})
	if err != nil {
		s.write(protocol.ResponseEvent{
			Type:  "response",
			ID:    cmd.ID,
			OK:    false,
			Error: err.Error(),
		})
		return
	}
	s.write(protocol.ResponseEvent{
		Type: "response",
		ID:   cmd.ID,
		OK:   true,
	})
}

func messageToEvent(msg message.Message, offset int64, isEdit bool) protocol.MessageEvent {
	typ := protocol.TypeMessage
	if isEdit {
		typ = protocol.TypeEdit
	}

	id, _ := ParseMessageID(msg.ID)
	replyTo, _ := ParseMessageID(msg.ReplyTo)
	threadID, _ := strconv.Atoi(msg.ThreadID)

	return protocol.MessageEvent{
		Type:         typ,
		Offset:       offset,
		ID:           id,
		ChatID:       0, // not available from message.Message
		Channel:      msg.Channel,
		From:         msg.From,
		Date:         msg.Date,
		Body:         msg.Body,
		ReplyTo:      replyTo,
		ReplyToBody:  msg.ReplyToBody,
		Quote:        msg.Quote,
		ThreadID:     threadID,
		MediaType:    msg.MediaType,
		MediaFileID:  msg.MediaFileID,
		Caption:      msg.Caption,
		ForwardFrom:  msg.ForwardFrom,
		ForwardDate:  msg.ForwardDate,
		EditDate:     msg.EditDate,
		MediaGroupID: msg.MediaGroupID,
	}
}

func messageToSummary(msg message.Message) *protocol.MsgSummary {
	id, _ := ParseMessageID(msg.ID)
	return &protocol.MsgSummary{
		ID:      id,
		Channel: msg.Channel,
		From:    msg.From,
		Date:    msg.Date,
		Body:    msg.Body,
	}
}