package store

import (
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
