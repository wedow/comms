package store

import (
	"testing"
)

func TestOffsetRoundTrip(t *testing.T) {
	root := t.TempDir()

	if err := WriteOffset(root, "telegram", 42); err != nil {
		t.Fatalf("WriteOffset: %v", err)
	}

	got, err := ReadOffset(root, "telegram")
	if err != nil {
		t.Fatalf("ReadOffset: %v", err)
	}

	if got != 42 {
		t.Errorf("ReadOffset = %d, want 42", got)
	}
}

func TestReadOffsetMissingFile(t *testing.T) {
	root := t.TempDir()

	got, err := ReadOffset(root, "telegram")
	if err != nil {
		t.Fatalf("ReadOffset should not error on missing file: %v", err)
	}
	if got != 0 {
		t.Errorf("ReadOffset = %d, want 0 for missing file", got)
	}
}

func TestOffsetOverwrite(t *testing.T) {
	root := t.TempDir()

	if err := WriteOffset(root, "telegram", 10); err != nil {
		t.Fatalf("WriteOffset: %v", err)
	}
	if err := WriteOffset(root, "telegram", 99); err != nil {
		t.Fatalf("WriteOffset: %v", err)
	}

	got, err := ReadOffset(root, "telegram")
	if err != nil {
		t.Fatalf("ReadOffset: %v", err)
	}
	if got != 99 {
		t.Errorf("ReadOffset = %d, want 99", got)
	}
}
