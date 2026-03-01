package store

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// WriteChatID writes a chat ID to <root>/<channel>/.chat_id.
func WriteChatID(root, channel string, chatID int64) error {
	dir := filepath.Join(root, channel)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(
		filepath.Join(dir, ".chat_id"),
		[]byte(strconv.FormatInt(chatID, 10)+"\n"),
		0o644,
	)
}

// ReadChatID reads a chat ID from <root>/<channel>/.chat_id.
// Returns an error if the file does not exist.
func ReadChatID(root, channel string) (int64, error) {
	data, err := os.ReadFile(filepath.Join(root, channel, ".chat_id"))
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
}
