package store

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wedow/comms/internal/message"
)

// InitDir creates the root and root/docs directories.
func InitDir(root string) error {
	return os.MkdirAll(filepath.Join(root, "docs"), 0o755)
}

// WriteMessage serializes msg in the given format and writes it to the store.
// Returns the path of the written file.
func WriteMessage(root string, msg message.Message, format string) (string, error) {
	var data []byte
	var ext string
	var err error

	switch format {
	case "markdown":
		data, err = message.MarshalMarkdown(msg)
		ext = ".md"
	case "org":
		data, err = message.MarshalOrg(msg)
		ext = ".org"
	default:
		return "", fmt.Errorf("unknown format: %s", format)
	}
	if err != nil {
		return "", err
	}

	chanDir := filepath.Join(root, msg.Provider+"-"+msg.Channel)
	if err := os.MkdirAll(chanDir, 0o755); err != nil {
		return "", err
	}

	name := strings.ReplaceAll(msg.Date.Format(time.RFC3339Nano), ":", "-") + ext
	path := filepath.Join(chanDir, name)
	return path, os.WriteFile(path, data, 0o644)
}

// ReadMessage reads a message file and deserializes based on extension.
func ReadMessage(path string) (message.Message, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return message.Message{}, err
	}

	switch filepath.Ext(path) {
	case ".md":
		return message.UnmarshalMarkdown(data)
	case ".org":
		return message.UnmarshalOrg(data)
	default:
		return message.Message{}, fmt.Errorf("unknown extension: %s", filepath.Ext(path))
	}
}

// FindMessageByID scans message files in the channel directory for a matching ID.
// Returns the file path and parsed message, or an error if not found.
func FindMessageByID(root, channel, id, format string) (string, message.Message, error) {
	paths, err := ListMessages(root, channel)
	if err != nil {
		return "", message.Message{}, err
	}
	for _, p := range paths {
		msg, err := ReadMessage(p)
		if err != nil {
			continue
		}
		if msg.ID == id {
			return p, msg, nil
		}
	}
	return "", message.Message{}, fmt.Errorf("message %s not found in %s", id, channel)
}

// AppendEdit appends an edit section to an existing message file.
func AppendEdit(path string, editDate time.Time, newBody string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.Write(data)
	if len(data) > 0 && data[len(data)-1] != '\n' {
		buf.WriteByte('\n')
	}
	fmt.Fprintf(&buf, "---edit---\ndate: %s\n%s\n", editDate.Format(time.RFC3339), newBody)
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// AppendReaction appends a reaction section to an existing message file.
func AppendReaction(path string, reactionDate time.Time, from, emoji string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.Write(data)
	if len(data) > 0 && data[len(data)-1] != '\n' {
		buf.WriteByte('\n')
	}
	fmt.Fprintf(&buf, "---reaction---\ndate: %s\nfrom: %s\nemoji: %s\n", reactionDate.Format(time.RFC3339), from, emoji)
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// WriteMedia writes media data to <chanDir>/<timestamp>/<index>.<ext>.
// Returns the absolute path of the written file.
func WriteMedia(chanDir, timestamp string, index int, ext string, data []byte) (string, error) {
	dir := filepath.Join(chanDir, timestamp)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := fmt.Sprintf("%03d%s", index, ext)
	path := filepath.Join(dir, name)
	return path, os.WriteFile(path, data, 0o644)
}

// ResetCursorIfNeeded moves the cursor before msgDate if the cursor is currently after it.
func ResetCursorIfNeeded(root, channel string, msgDate time.Time) error {
	cursor, err := ReadCursor(root, channel)
	if err != nil {
		return err
	}
	if cursor.IsZero() || cursor.Before(msgDate) {
		return nil
	}
	return WriteCursor(root, channel, msgDate.Add(-time.Nanosecond))
}
