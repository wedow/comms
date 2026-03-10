package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrimeCommand(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"prime"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("prime command: %v", err)
	}

	out := buf.String()
	if !strings.HasPrefix(out, "# Agent Bootstrap Guide") {
		t.Errorf("expected output to start with guide header, got %q", out[:min(80, len(out))])
	}
	if !strings.Contains(out, "comms unread") {
		t.Error("expected output to mention 'comms unread'")
	}
}
