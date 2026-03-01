package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ListChannels returns sorted directory names under root, excluding "docs" and plain files.
func ListChannels(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var channels []string
	for _, e := range entries {
		if e.IsDir() && e.Name() != "docs" {
			channels = append(channels, e.Name())
		}
	}
	sort.Strings(channels)
	return channels, nil
}

// ListMessages returns sorted file paths in a channel directory, excluding .cursor.
// Also includes files from topic-* subdirectories.
// Files are sorted by basename (timestamp-based filenames) ascending (oldest first).
func ListMessages(root, channel string) ([]string, error) {
	chanDir := filepath.Join(root, channel)
	entries, err := os.ReadDir(chanDir)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "topic-") {
			subDir := filepath.Join(chanDir, e.Name())
			subEntries, err := os.ReadDir(subDir)
			if err != nil {
				return nil, err
			}
			for _, se := range subEntries {
				if !se.IsDir() && !strings.HasPrefix(se.Name(), ".") {
					paths = append(paths, filepath.Join(subDir, se.Name()))
				}
			}
		} else if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			paths = append(paths, filepath.Join(chanDir, e.Name()))
		}
	}
	sort.Slice(paths, func(i, j int) bool {
		return filepath.Base(paths[i]) < filepath.Base(paths[j])
	})
	return paths, nil
}

// ListMessagesAfter returns sorted file paths with timestamp strictly after the given time.
// If after is zero, all messages are returned.
func ListMessagesAfter(root, channel string, after time.Time) ([]string, error) {
	all, err := ListMessages(root, channel)
	if err != nil {
		return nil, err
	}

	if after.IsZero() {
		return all, nil
	}

	var filtered []string
	for _, path := range all {
		t, err := parseFilenameTime(filepath.Base(path))
		if err != nil {
			return nil, fmt.Errorf("parsing time from %s: %w", filepath.Base(path), err)
		}
		if t.After(after) {
			filtered = append(filtered, path)
		}
	}
	return filtered, nil
}

// parseFilenameTime reverses the filename encoding back to a time.Time.
// Filename format: "2026-03-01T12-30-00.123456789Z.md" -> time
func parseFilenameTime(name string) (time.Time, error) {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	idx := strings.Index(base, "T")
	if idx < 0 {
		return time.Time{}, fmt.Errorf("no T separator in %q", name)
	}
	datePart := base[:idx]
	timePart := base[idx+1:]

	// Replace hyphens with colons in the HH-MM-SS portion (before any fractional seconds dot)
	if dotIdx := strings.Index(timePart, "."); dotIdx >= 0 {
		hms := strings.ReplaceAll(timePart[:dotIdx], "-", ":")
		timePart = hms + timePart[dotIdx:]
	} else {
		timePart = strings.ReplaceAll(timePart, "-", ":")
	}

	return time.Parse(time.RFC3339Nano, datePart+"T"+timePart)
}
