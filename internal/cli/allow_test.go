package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wedow/comms/internal/store"
)

func TestAllowAddAndList(t *testing.T) {
	root := t.TempDir()

	// Add an ID
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"allow", "add", "--dir", root, "123"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("allow add: %v", err)
	}
	if !strings.Contains(buf.String(), `"ok":true`) {
		t.Errorf("expected ok response, got %q", buf.String())
	}

	// List IDs
	cmd = newRootCmd()
	buf = new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"allow", "list", "--dir", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("allow list: %v", err)
	}
	if !strings.Contains(buf.String(), `"chat_id":123`) {
		t.Errorf("expected chat_id 123, got %q", buf.String())
	}
}

func TestAllowAddNegativeID(t *testing.T) {
	root := t.TempDir()

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"allow", "add", "--dir", root, "--", "-1001234"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("allow add negative: %v", err)
	}

	ids, _ := store.ReadAllowedIDs(root)
	if len(ids) != 1 || ids[0] != -1001234 {
		t.Errorf("got %v, want [-1001234]", ids)
	}
}

func TestAllowRemove(t *testing.T) {
	root := t.TempDir()
	store.AddAllowedID(root, 123)
	store.AddAllowedID(root, 456)

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"allow", "remove", "--dir", root, "123"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("allow remove: %v", err)
	}

	ids, _ := store.ReadAllowedIDs(root)
	if len(ids) != 1 || ids[0] != 456 {
		t.Errorf("got %v, want [456]", ids)
	}
}

func TestAllowListEmpty(t *testing.T) {
	root := t.TempDir()

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"allow", "list", "--dir", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("allow list: %v", err)
	}
	if buf.String() != "" {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestAllowAddInvalidID(t *testing.T) {
	root := t.TempDir()

	cmd := newRootCmd()
	errBuf := new(bytes.Buffer)
	cmd.SetErr(errBuf)
	cmd.SetArgs([]string{"allow", "add", "--dir", root, "notanumber"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestAllowDirDefaultsToRelative(t *testing.T) {
	root := t.TempDir()
	absRoot, _ := filepath.Abs(root)
	store.AddAllowedID(absRoot, 42)

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"allow", "list", "--dir", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("allow list: %v", err)
	}
	if !strings.Contains(buf.String(), `"chat_id":42`) {
		t.Errorf("expected chat_id 42, got %q", buf.String())
	}
}
