package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wedow/comms/internal/store"
)

func TestChannelsCommand(t *testing.T) {
	tmpDir := t.TempDir()

	// Create channel directories
	for _, name := range []string{"telegram-general", "telegram-dev", "docs"} {
		if err := os.MkdirAll(filepath.Join(tmpDir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// Create a plain file (should be excluded)
	os.WriteFile(filepath.Join(tmpDir, "config.toml"), []byte("x"), 0o644)

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"channels", "--dir", tmpDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("channels command: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2: %v", len(lines), lines)
	}

	// Channels are sorted, so telegram-dev comes first
	wantLines := []string{
		`{"name":"telegram-dev","provider":"telegram","path":"` + filepath.Join(tmpDir, "telegram-dev") + `"}`,
		`{"name":"telegram-general","provider":"telegram","path":"` + filepath.Join(tmpDir, "telegram-general") + `"}`,
	}
	for i, want := range wantLines {
		if lines[i] != want {
			t.Errorf("line %d = %q, want %q", i, lines[i], want)
		}
	}
}

func TestChannelsCommandEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"channels", "--dir", tmpDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("channels command: %v", err)
	}

	if buf.String() != "" {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestChannelsProviderExtraction(t *testing.T) {
	tmpDir := t.TempDir()

	// Channel with multiple hyphens - provider is only the first segment
	os.MkdirAll(filepath.Join(tmpDir, "discord-my-server"), 0o755)

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"channels", "--dir", tmpDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("channels command: %v", err)
	}

	want := `{"name":"discord-my-server","provider":"discord","path":"` + filepath.Join(tmpDir, "discord-my-server") + `"}` + "\n"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
}

func TestChannelsIncludesChatID(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "telegram-general"), 0o755)
	store.WriteChatID(tmpDir, "telegram-general", -1001234)

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"channels", "--dir", tmpDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("channels command: %v", err)
	}

	want := `{"name":"telegram-general","provider":"telegram","path":"` + filepath.Join(tmpDir, "telegram-general") + `","chat_id":-1001234}`
	if !strings.Contains(buf.String(), want) {
		t.Errorf("got %q, want line containing %q", buf.String(), want)
	}
}
