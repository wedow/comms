package message

import (
	"bytes"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v3"
)

// MarshalMarkdown renders a Message as markdown with YAML frontmatter.
func MarshalMarkdown(msg Message) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("---\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(msg); err != nil {
		return nil, err
	}
	enc.Close()
	buf.WriteString("---\n")
	buf.WriteString(msg.Body)

	return buf.Bytes(), nil
}

// UnmarshalMarkdown parses YAML frontmatter + body back into a Message.
func UnmarshalMarkdown(data []byte) (Message, error) {
	var msg Message
	body, err := frontmatter.Parse(bytes.NewReader(data), &msg)
	if err != nil {
		return Message{}, err
	}
	msg.Body = string(body)
	return msg, nil
}
