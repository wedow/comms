package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// MarshalOrg renders a Message as org-mode keyword lines + body.
func MarshalOrg(msg Message) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "#+FROM: %s\n", msg.From)
	fmt.Fprintf(&buf, "#+PROVIDER: %s\n", msg.Provider)
	fmt.Fprintf(&buf, "#+CHANNEL: %s\n", msg.Channel)
	fmt.Fprintf(&buf, "#+DATE: %s\n", msg.Date.Format(time.RFC3339))
	fmt.Fprintf(&buf, "#+ID: %s\n", msg.ID)

	// Optional string fields.
	for _, kv := range []struct{ key, val string }{
		{"REPLY_TO", msg.ReplyTo},
		{"REPLY_TO_BODY", msg.ReplyToBody},
		{"QUOTE", msg.Quote},
		{"MEDIA_TYPE", msg.MediaType},
		{"MEDIA_FILE_ID", msg.MediaFileID},
		{"MEDIA_URL", msg.MediaURL},
		{"CAPTION", msg.Caption},
		{"FORWARD_FROM", msg.ForwardFrom},
		{"THREAD_ID", msg.ThreadID},
		{"MEDIA_GROUP_ID", msg.MediaGroupID},
	} {
		if kv.val != "" {
			fmt.Fprintf(&buf, "#+%s: %s\n", kv.key, kv.val)
		}
	}

	// Optional time pointer fields.
	if msg.ForwardDate != nil {
		fmt.Fprintf(&buf, "#+FORWARD_DATE: %s\n", msg.ForwardDate.Format(time.RFC3339))
	}
	if msg.EditDate != nil {
		fmt.Fprintf(&buf, "#+EDIT_DATE: %s\n", msg.EditDate.Format(time.RFC3339))
	}

	// Entities as JSON array.
	if len(msg.Entities) > 0 {
		b, err := json.Marshal(msg.Entities)
		if err != nil {
			return nil, fmt.Errorf("marshal ENTITIES: %w", err)
		}
		fmt.Fprintf(&buf, "#+ENTITIES: %s\n", b)
	}

	if msg.Body != "" {
		buf.WriteByte('\n')
		buf.WriteString(msg.Body)
	}
	return buf.Bytes(), nil
}

// UnmarshalOrg parses org-mode keyword lines + body back into a Message.
func UnmarshalOrg(data []byte) (Message, error) {
	var msg Message
	rest := string(data)

	for {
		if rest == "" {
			break
		}
		if !strings.HasPrefix(rest, "#+") {
			break
		}
		line, after, _ := strings.Cut(rest, "\n")
		key, val, ok := strings.Cut(line[2:], ": ")
		if !ok {
			break
		}
		switch key {
		case "FROM":
			msg.From = val
		case "PROVIDER":
			msg.Provider = val
		case "CHANNEL":
			msg.Channel = val
		case "DATE":
			t, err := time.Parse(time.RFC3339, val)
			if err != nil {
				return Message{}, fmt.Errorf("parse DATE: %w", err)
			}
			msg.Date = t
		case "ID":
			msg.ID = val
		case "REPLY_TO":
			msg.ReplyTo = val
		case "REPLY_TO_BODY":
			msg.ReplyToBody = val
		case "QUOTE":
			msg.Quote = val
		case "MEDIA_TYPE":
			msg.MediaType = val
		case "MEDIA_FILE_ID":
			msg.MediaFileID = val
		case "MEDIA_URL":
			msg.MediaURL = val
		case "CAPTION":
			msg.Caption = val
		case "FORWARD_FROM":
			msg.ForwardFrom = val
		case "THREAD_ID":
			msg.ThreadID = val
		case "MEDIA_GROUP_ID":
			msg.MediaGroupID = val
		case "FORWARD_DATE":
			t, err := time.Parse(time.RFC3339, val)
			if err != nil {
				return Message{}, fmt.Errorf("parse FORWARD_DATE: %w", err)
			}
			msg.ForwardDate = &t
		case "EDIT_DATE":
			t, err := time.Parse(time.RFC3339, val)
			if err != nil {
				return Message{}, fmt.Errorf("parse EDIT_DATE: %w", err)
			}
			msg.EditDate = &t
		case "ENTITIES":
			if err := json.Unmarshal([]byte(val), &msg.Entities); err != nil {
				return Message{}, fmt.Errorf("parse ENTITIES: %w", err)
			}
		}
		rest = after
	}

	// Skip the blank separator line between headers and body.
	rest = strings.TrimPrefix(rest, "\n")
	msg.Body = rest

	return msg, nil
}
