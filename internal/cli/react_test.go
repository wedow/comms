package cli

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/wedow/comms/internal/store"
)

func stubReactDelegate(t *testing.T, err error) *delegateCapture {
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
		if err != nil {
			return nil, err
		}
		return []byte(`{"ok":true}` + "\n"), nil
	}

	return cap
}

func TestReactCmd(t *testing.T) {
	t.Run("successful reaction via delegation", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		cap := stubReactDelegate(t, nil)

		cmd := newReactCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetArgs([]string{"--channel", "telegram-general", "--message", "telegram-99", "--emoji", "\U0001F44D", "--dir", root})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v (stderr: %s)", err, errBuf.String())
		}

		if cap.binary != "/usr/bin/comms-telegram" {
			t.Errorf("binary = %q, want /usr/bin/comms-telegram", cap.binary)
		}
		if !containsFlagValue(cap.args, "--chat-id", "123") {
			t.Errorf("args = %v, missing --chat-id 123", cap.args)
		}
		if !containsFlagValue(cap.args, "--message", "telegram-99") {
			t.Errorf("args = %v, missing --message telegram-99", cap.args)
		}
		if !containsFlagValue(cap.args, "--emoji", "\U0001F44D") {
			t.Errorf("args = %v, missing --emoji thumbs up", cap.args)
		}

		got := out.String()
		if !strings.Contains(got, `"ok":true`) {
			t.Errorf("stdout = %q, want ok:true", got)
		}
		if !strings.Contains(got, `"channel":"telegram-general"`) {
			t.Errorf("stdout = %q, want channel:telegram-general", got)
		}
	})

	t.Run("provider binary error", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		_ = stubReactDelegate(t, fmt.Errorf("exit status 1"))

		cmd := newReactCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetArgs([]string{"--channel", "telegram-general", "--message", "telegram-99", "--emoji", "\U0001F44D", "--dir", root})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for provider binary failure")
		}
		if !strings.Contains(errBuf.String(), `"error"`) {
			t.Errorf("stderr = %q, want JSON error", errBuf.String())
		}
	})

	t.Run("missing channel", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		_ = stubReactDelegate(t, nil)

		cmd := newReactCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetArgs([]string{"--message", "telegram-99", "--emoji", "\U0001F44D", "--dir", root})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for missing channel flag")
		}
	})

	t.Run("missing chat_id", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)

		_ = stubReactDelegate(t, nil)

		cmd := newReactCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetArgs([]string{"--channel", "telegram-nonexistent", "--message", "telegram-99", "--emoji", "\U0001F44D", "--dir", root})

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

		cmd := newReactCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetArgs([]string{"--channel", "telegram-general", "--message", "telegram-99", "--emoji", "\U0001F44D", "--dir", root})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for missing provider binary")
		}
		if !strings.Contains(errBuf.String(), `"error"`) {
			t.Errorf("stderr = %q, want JSON error", errBuf.String())
		}
	})

	t.Run("provider config passed as env var", func(t *testing.T) {
		root := t.TempDir()
		writeTestConfig(t, root)
		store.WriteChatID(root, "telegram-general", 123)

		cap := stubReactDelegate(t, nil)

		cmd := newReactCmd()
		out := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		cmd.SetArgs([]string{"--channel", "telegram-general", "--message", "telegram-99", "--emoji", "\U0001F44D", "--dir", root})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		found := false
		for _, e := range cap.env {
			if strings.HasPrefix(e, "COMMS_PROVIDER_CONFIG=") {
				found = true
				if !strings.Contains(e, "test-token") {
					t.Errorf("env var = %q, want token in config", e)
				}
			}
		}
		if !found {
			t.Errorf("env = %v, missing COMMS_PROVIDER_CONFIG", cap.env)
		}
	})
}
