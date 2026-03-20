package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wedow/comms/internal/message"
	"github.com/wedow/comms/internal/store"
)

func writeTestConfig(t *testing.T, root string) {
	t.Helper()
	data := []byte("[telegram]\ntoken = \"test-token\"\n\n[providers.telegram]\ntoken = \"test-token\"\n")
	if err := os.WriteFile(filepath.Join(root, "config.toml"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// stubDelegate sets up lookPath and runDelegateOutput for testing send delegation.
// Returns a cleanup function and pointers to capture what was called.
type delegateCapture struct {
	binary string
	args   []string
	env    []string
	stdin  string
}

func stubSendDelegate(t *testing.T, response map[string]any, err error) *delegateCapture {
	t.Helper()
	cap := &delegateCapture{}

	origLookPath := lookPath
	origRunOutput := runDelegateOutput
	t.Cleanup(func() {
		lookPath = origLookPath
		runDelegateOutput = origRunOutput
	})

	lookPath = func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}

	runDelegateOutput = func(binary string, args []string, env []string, stdin io.Reader) ([]byte, error) {
		cap.binary = binary
		cap.args = args
		cap.env = env
		if stdin != nil {
			data, _ := io.ReadAll(stdin)
			cap.stdin = string(data)
		}
		if err != nil {
			return nil, err
		}
		out, _ := json.Marshal(response)
		return append(out, '\n'), nil
	}

	return cap
}

func TestSendCmd(t *testing.T) {
	t.Run("successful send via delegation", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		cap := stubSendDelegate(t, map[string]any{"ok": true, "message_id": float64(42)}, nil)

		cmd := newSendCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("hello world"))
		cmd.SetArgs([]string{"--channel", "telegram-general", "--dir", root})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v (stderr: %s)", err, errBuf.String())
		}

		// Verify delegation was called correctly
		if cap.binary != "/usr/bin/comms-telegram" {
			t.Errorf("binary = %q, want /usr/bin/comms-telegram", cap.binary)
		}
		// Verify chat-id flag was passed
		if !containsFlagValue(cap.args, "--chat-id", "123") {
			t.Errorf("args = %v, missing --chat-id 123", cap.args)
		}
		// Verify body was sent via stdin
		if cap.stdin != "hello world" {
			t.Errorf("stdin = %q, want %q", cap.stdin, "hello world")
		}

		got := out.String()
		if !strings.Contains(got, `"ok":true`) {
			t.Errorf("stdout = %q, want ok:true", got)
		}
		if !strings.Contains(got, `"channel":"telegram-general"`) {
			t.Errorf("stdout = %q, want channel:telegram-general", got)
		}
	})

	t.Run("positional args as message body", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		cap := stubSendDelegate(t, map[string]any{"ok": true, "message_id": float64(1)}, nil)

		cmd := newSendCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader(""))
		cmd.SetArgs([]string{"--channel", "telegram-general", "--dir", root, "hello", "world!"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Body from positional args should be sent via stdin to provider
		if cap.stdin != "hello world!" {
			t.Errorf("stdin = %q, want %q", cap.stdin, "hello world!")
		}
	})

	t.Run("empty stdin", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		_ = stubSendDelegate(t, nil, nil)

		cmd := newSendCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader(""))
		cmd.SetArgs([]string{"--channel", "telegram-general", "--dir", root})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for empty stdin")
		}
		if !strings.Contains(errBuf.String(), `"error"`) {
			t.Errorf("stderr = %q, want JSON error", errBuf.String())
		}
	})

	t.Run("missing chat_id", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)

		_ = stubSendDelegate(t, nil, nil)

		cmd := newSendCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("hello"))
		cmd.SetArgs([]string{"--channel", "telegram-nonexistent", "--dir", root})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for missing chat_id")
		}
		if !strings.Contains(errBuf.String(), `"error"`) {
			t.Errorf("stderr = %q, want JSON error", errBuf.String())
		}
	})

	t.Run("provider binary not found", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		origLookPath := lookPath
		t.Cleanup(func() { lookPath = origLookPath })
		lookPath = func(name string) (string, error) {
			return "", fmt.Errorf("not found: %s", name)
		}

		cmd := newSendCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("hello"))
		cmd.SetArgs([]string{"--channel", "telegram-general", "--dir", root})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for missing provider binary")
		}
		if !strings.Contains(errBuf.String(), `"error"`) {
			t.Errorf("stderr = %q, want JSON error", errBuf.String())
		}
	})

	t.Run("provider binary error", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		_ = stubSendDelegate(t, nil, fmt.Errorf("exit status 1"))

		cmd := newSendCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("hello"))
		cmd.SetArgs([]string{"--channel", "telegram-general", "--dir", root})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for provider binary failure")
		}
		if !strings.Contains(errBuf.String(), `"error"`) {
			t.Errorf("stderr = %q, want JSON error", errBuf.String())
		}
	})

	t.Run("reply-to flag forwarded", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		cap := stubSendDelegate(t, map[string]any{"ok": true, "message_id": float64(1)}, nil)

		cmd := newSendCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("reply text"))
		cmd.SetArgs([]string{"--channel", "telegram-general", "--dir", root, "--reply-to", "telegram-99"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !containsFlagValue(cap.args, "--reply-to", "telegram-99") {
			t.Errorf("args = %v, missing --reply-to telegram-99", cap.args)
		}
	})

	t.Run("file flag forwarded", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		tmpFile := filepath.Join(t.TempDir(), "photo.jpg")
		os.WriteFile(tmpFile, []byte("fake image data"), 0o644)

		cap := stubSendDelegate(t, map[string]any{"ok": true, "message_id": float64(1)}, nil)

		cmd := newSendCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("my caption"))
		cmd.SetArgs([]string{"--channel", "telegram-general", "--dir", root, "--file", tmpFile})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v (stderr: %s)", err, errBuf.String())
		}
		if !containsFlagValue(cap.args, "--file", tmpFile) {
			t.Errorf("args = %v, missing --file %s", cap.args, tmpFile)
		}
	})

	t.Run("media-type flag forwarded", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		tmpFile := filepath.Join(t.TempDir(), "data.bin")
		os.WriteFile(tmpFile, []byte("binary data"), 0o644)

		cap := stubSendDelegate(t, map[string]any{"ok": true, "message_id": float64(1)}, nil)

		cmd := newSendCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader(""))
		cmd.SetArgs([]string{"--channel", "telegram-general", "--dir", root, "--file", tmpFile, "--media-type", "video"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v (stderr: %s)", err, errBuf.String())
		}
		if !containsFlagValue(cap.args, "--media-type", "video") {
			t.Errorf("args = %v, missing --media-type video", cap.args)
		}
	})

	t.Run("thread flag forwarded", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		cap := stubSendDelegate(t, map[string]any{"ok": true, "message_id": float64(1)}, nil)

		cmd := newSendCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("topic message"))
		cmd.SetArgs([]string{"--channel", "telegram-general", "--dir", root, "--thread", "42"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !containsFlagValue(cap.args, "--thread", "42") {
			t.Errorf("args = %v, missing --thread 42", cap.args)
		}
	})

	t.Run("format flag forwarded", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		cap := stubSendDelegate(t, map[string]any{"ok": true, "message_id": float64(1)}, nil)

		cmd := newSendCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("*bold*"))
		cmd.SetArgs([]string{"--channel", "telegram-general", "--dir", root, "--format", "markdown"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !containsFlagValue(cap.args, "--format", "markdown") {
			t.Errorf("args = %v, missing --format markdown", cap.args)
		}
	})

	t.Run("provider config passed as env var", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		cap := stubSendDelegate(t, map[string]any{"ok": true, "message_id": float64(1)}, nil)

		cmd := newSendCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("hello"))
		cmd.SetArgs([]string{"--channel", "telegram-general", "--dir", root})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Check that COMMS_PROVIDER_CONFIG was set in env
		found := false
		for _, e := range cap.env {
			if strings.HasPrefix(e, "COMMS_PROVIDER_CONFIG=") {
				found = true
				// Should contain the token
				if !strings.Contains(e, "test-token") {
					t.Errorf("env var = %q, want token in config", e)
				}
			}
		}
		if !found {
			t.Errorf("env = %v, missing COMMS_PROVIDER_CONFIG", cap.env)
		}
	})

	t.Run("writes local message file", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		_ = stubSendDelegate(t, map[string]any{"ok": true, "message_id": float64(42)}, nil)

		cmd := newSendCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader("saved locally"))
		cmd.SetArgs([]string{"--channel", "telegram-general", "--dir", root})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v (stderr: %s)", err, errBuf.String())
		}

		// Check that a .md file was written in the channel directory
		channelDir := filepath.Join(root, "telegram-general")
		entries, err := os.ReadDir(channelDir)
		if err != nil {
			t.Fatalf("channel dir not created: %v", err)
		}
		found := false
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".md") {
				found = true
				msg, err := store.ReadMessage(filepath.Join(channelDir, e.Name()))
				if err != nil {
					t.Fatalf("read message: %v", err)
				}
				if msg.Body != "saved locally" {
					t.Errorf("body = %q, want 'saved locally'", msg.Body)
				}
				if msg.ID != "telegram-42" {
					t.Errorf("id = %q, want telegram-42", msg.ID)
				}
				if msg.Provider != "telegram" {
					t.Errorf("provider = %q, want telegram", msg.Provider)
				}
				if msg.Channel != "general" {
					t.Errorf("channel = %q, want general", msg.Channel)
				}
			}
		}
		if !found {
			t.Error("no .md file found in channel directory")
		}
	})

	t.Run("advances cursor when no unreads", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		_ = stubSendDelegate(t, map[string]any{"ok": true, "message_id": float64(50)}, nil)

		cmd := newSendCmd()
		out := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetIn(strings.NewReader("hello"))
		cmd.SetArgs([]string{"--channel", "telegram-general", "--dir", root})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Cursor should have advanced past the sent message
		cursor, err := store.ReadCursor(root, "telegram-general")
		if err != nil {
			t.Fatalf("read cursor: %v", err)
		}
		if cursor.IsZero() {
			t.Fatal("cursor should be set after send with no unreads")
		}

		// Unread should return nothing
		unreads, _ := store.ListMessagesAfter(root, "telegram-general", cursor)
		if len(unreads) != 0 {
			t.Errorf("expected 0 unreads, got %d", len(unreads))
		}
	})

	t.Run("preserves cursor when unreads exist", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		// Write an existing unread message
		existing := message.Message{
			From:     "someone",
			Provider: "telegram",
			Channel:  "general",
			Date:     time.Date(2024, 3, 1, 15, 0, 0, 0, time.UTC),
			ID:       "telegram-49",
			Body:     "hey there",
		}
		store.WriteMessage(root, existing, "markdown")

		_ = stubSendDelegate(t, map[string]any{"ok": true, "message_id": float64(50)}, nil)

		cmd := newSendCmd()
		out := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetIn(strings.NewReader("reply"))
		cmd.SetArgs([]string{"--channel", "telegram-general", "--dir", root})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Cursor should still be zero (not advanced)
		cursor, err := store.ReadCursor(root, "telegram-general")
		if err != nil {
			t.Fatalf("read cursor: %v", err)
		}
		if !cursor.IsZero() {
			t.Errorf("cursor should be zero when unreads exist, got %v", cursor)
		}

		// Both messages should be unread
		unreads, _ := store.ListMessagesAfter(root, "telegram-general", cursor)
		if len(unreads) != 2 {
			t.Errorf("expected 2 unreads, got %d", len(unreads))
		}
	})

	t.Run("file flag allows empty body", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		tmpFile := filepath.Join(t.TempDir(), "doc.pdf")
		os.WriteFile(tmpFile, []byte("fake pdf"), 0o644)

		_ = stubSendDelegate(t, map[string]any{"ok": true, "message_id": float64(1)}, nil)

		cmd := newSendCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetIn(strings.NewReader(""))
		cmd.SetArgs([]string{"--channel", "telegram-general", "--dir", root, "--file", tmpFile})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v (stderr: %s)", err, errBuf.String())
		}
		if !strings.Contains(out.String(), `"ok":true`) {
			t.Errorf("stdout = %q, want ok:true", out.String())
		}
	})
}

// containsFlagValue checks if args contains a flag with the given value.
func containsFlagValue(args []string, flag, value string) bool {
	for i, a := range args {
		if a == flag && i+1 < len(args) && args[i+1] == value {
			return true
		}
	}
	return false
}
