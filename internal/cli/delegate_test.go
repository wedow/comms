package cli

import (
	"errors"
	"testing"
)

func TestExtractProvider(t *testing.T) {
	tests := []struct {
		channel string
		want    string
	}{
		{"telegram-general", "telegram"},
		{"telegram-my-group", "telegram"},
		{"discord-server", "discord"},
	}
	for _, tt := range tests {
		t.Run(tt.channel, func(t *testing.T) {
			got := extractProvider(tt.channel)
			if got != tt.want {
				t.Errorf("extractProvider(%q) = %q, want %q", tt.channel, got, tt.want)
			}
		})
	}
}

func TestResolveProviderBinary(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		orig := lookPath
		t.Cleanup(func() { lookPath = orig })
		lookPath = func(name string) (string, error) {
			if name != "comms-telegram" {
				t.Errorf("lookPath called with %q, want %q", name, "comms-telegram")
			}
			return "/usr/bin/comms-telegram", nil
		}

		got, err := resolveProviderBinary("telegram")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "/usr/bin/comms-telegram" {
			t.Errorf("got %q, want %q", got, "/usr/bin/comms-telegram")
		}
	})

	t.Run("not found", func(t *testing.T) {
		orig := lookPath
		t.Cleanup(func() { lookPath = orig })
		lookPath = func(name string) (string, error) {
			return "", errors.New("not found")
		}

		_, err := resolveProviderBinary("slack")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestDelegate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		orig := runDelegate
		t.Cleanup(func() { runDelegate = orig })

		var gotBinary string
		var gotArgs []string
		runDelegate = func(binary string, args []string) error {
			gotBinary = binary
			gotArgs = args
			return nil
		}

		err := delegate("/usr/bin/comms-telegram", []string{"send", "--channel", "general"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotBinary != "/usr/bin/comms-telegram" {
			t.Errorf("binary = %q, want %q", gotBinary, "/usr/bin/comms-telegram")
		}
		if len(gotArgs) != 3 || gotArgs[0] != "send" || gotArgs[1] != "--channel" || gotArgs[2] != "general" {
			t.Errorf("args = %v, want [send --channel general]", gotArgs)
		}
	})

	t.Run("error", func(t *testing.T) {
		orig := runDelegate
		t.Cleanup(func() { runDelegate = orig })
		runDelegate = func(binary string, args []string) error {
			return errors.New("exit status 1")
		}

		err := delegate("/usr/bin/comms-telegram", []string{"send"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
