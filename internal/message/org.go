package message

import (
	"bytes"
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
		}
		rest = after
	}

	// Skip the blank separator line between headers and body.
	rest = strings.TrimPrefix(rest, "\n")
	msg.Body = rest

	return msg, nil
}
