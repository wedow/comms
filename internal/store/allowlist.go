package store

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const allowlistFile = "allowed_ids"

func allowlistPath(root string) string {
	return filepath.Join(root, allowlistFile)
}

// ReadAllowedIDs reads the allowed chat IDs from <root>/allowed_ids.
// Returns nil (not error) if the file does not exist.
func ReadAllowedIDs(root string) ([]int64, error) {
	data, err := os.ReadFile(allowlistPath(root))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var ids []int64
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		id, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// AddAllowedID appends a chat ID to <root>/allowed_ids if not already present.
func AddAllowedID(root string, id int64) error {
	ids, err := ReadAllowedIDs(root)
	if err != nil {
		return err
	}
	for _, existing := range ids {
		if existing == id {
			return nil
		}
	}
	ids = append(ids, id)
	return writeAllowedIDs(root, ids)
}

// RemoveAllowedID removes a chat ID from <root>/allowed_ids.
func RemoveAllowedID(root string, id int64) error {
	ids, err := ReadAllowedIDs(root)
	if err != nil {
		return err
	}
	var filtered []int64
	for _, existing := range ids {
		if existing != id {
			filtered = append(filtered, existing)
		}
	}
	return writeAllowedIDs(root, filtered)
}

func writeAllowedIDs(root string, ids []int64) error {
	var lines []string
	for _, id := range ids {
		lines = append(lines, strconv.FormatInt(id, 10))
	}
	content := ""
	if len(lines) > 0 {
		content = strings.Join(lines, "\n") + "\n"
	}
	return os.WriteFile(allowlistPath(root), []byte(content), 0o644)
}
