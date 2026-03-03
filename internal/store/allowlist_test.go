package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadAllowedIDsMissing(t *testing.T) {
	root := t.TempDir()
	ids, err := ReadAllowedIDs(root)
	if err != nil {
		t.Fatalf("ReadAllowedIDs: %v", err)
	}
	if ids != nil {
		t.Errorf("expected nil, got %v", ids)
	}
}

func TestReadAllowedIDsEmpty(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "allowed_ids"), []byte(""), 0o644)

	ids, err := ReadAllowedIDs(root)
	if err != nil {
		t.Fatalf("ReadAllowedIDs: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected empty, got %v", ids)
	}
}

func TestReadAllowedIDs(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "allowed_ids"), []byte("123\n-456\n789\n"), 0o644)

	ids, err := ReadAllowedIDs(root)
	if err != nil {
		t.Fatalf("ReadAllowedIDs: %v", err)
	}
	want := []int64{123, -456, 789}
	if len(ids) != len(want) {
		t.Fatalf("got %d IDs, want %d", len(ids), len(want))
	}
	for i, id := range ids {
		if id != want[i] {
			t.Errorf("ids[%d] = %d, want %d", i, id, want[i])
		}
	}
}

func TestAddAllowedID(t *testing.T) {
	root := t.TempDir()

	if err := AddAllowedID(root, 123); err != nil {
		t.Fatalf("AddAllowedID: %v", err)
	}
	if err := AddAllowedID(root, -456); err != nil {
		t.Fatalf("AddAllowedID: %v", err)
	}

	ids, _ := ReadAllowedIDs(root)
	if len(ids) != 2 || ids[0] != 123 || ids[1] != -456 {
		t.Errorf("got %v, want [123 -456]", ids)
	}
}

func TestAddAllowedIDDuplicate(t *testing.T) {
	root := t.TempDir()

	AddAllowedID(root, 123)
	AddAllowedID(root, 123) // duplicate

	ids, _ := ReadAllowedIDs(root)
	if len(ids) != 1 {
		t.Errorf("expected 1 ID after duplicate add, got %d", len(ids))
	}
}

func TestRemoveAllowedID(t *testing.T) {
	root := t.TempDir()
	AddAllowedID(root, 123)
	AddAllowedID(root, 456)

	if err := RemoveAllowedID(root, 123); err != nil {
		t.Fatalf("RemoveAllowedID: %v", err)
	}

	ids, _ := ReadAllowedIDs(root)
	if len(ids) != 1 || ids[0] != 456 {
		t.Errorf("got %v, want [456]", ids)
	}
}

func TestRemoveAllowedIDNotPresent(t *testing.T) {
	root := t.TempDir()
	AddAllowedID(root, 123)

	if err := RemoveAllowedID(root, 999); err != nil {
		t.Fatalf("RemoveAllowedID: %v", err)
	}

	ids, _ := ReadAllowedIDs(root)
	if len(ids) != 1 || ids[0] != 123 {
		t.Errorf("got %v, want [123]", ids)
	}
}
