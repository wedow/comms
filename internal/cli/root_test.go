package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandName(t *testing.T) {
	cmd := newRootCmd()
	if cmd.Use != "comms" {
		t.Errorf("root command Use = %q, want %q", cmd.Use, "comms")
	}
}

func TestVersionFlag(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.HasPrefix(out, "comms") {
		t.Errorf("--version output = %q, want prefix %q", out, "comms")
	}
}

func TestHelpContainsComms(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "comms") {
		t.Errorf("--help output missing %q, got: %s", "comms", out)
	}
}
