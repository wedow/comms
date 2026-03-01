package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReadCursor reads the cursor timestamp from root/<channel>/.cursor.
// Returns zero time if the file does not exist.
func ReadCursor(root, channel string) (time.Time, error) {
	data, err := os.ReadFile(filepath.Join(root, channel, ".cursor"))
	if errors.Is(err, os.ErrNotExist) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339Nano, strings.TrimSpace(string(data)))
}

// WriteCursor writes a RFC3339Nano timestamp to root/<channel>/.cursor.
func WriteCursor(root, channel string, t time.Time) error {
	return os.WriteFile(
		filepath.Join(root, channel, ".cursor"),
		[]byte(t.Format(time.RFC3339Nano)+"\n"),
		0o644,
	)
}
