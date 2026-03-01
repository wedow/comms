package store

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ReadOffset reads the polling offset for a provider.
// Returns 0 if the offset file does not exist.
func ReadOffset(root, provider string) (int64, error) {
	data, err := os.ReadFile(filepath.Join(root, provider+".offset"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
}

// WriteOffset writes the polling offset for a provider.
func WriteOffset(root, provider string, offset int64) error {
	return os.WriteFile(
		filepath.Join(root, provider+".offset"),
		[]byte(strconv.FormatInt(offset, 10)+"\n"),
		0o644,
	)
}
