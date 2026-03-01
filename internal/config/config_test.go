package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	c := Default()

	if c.General.Format != "markdown" {
		t.Errorf("General.Format = %q, want %q", c.General.Format, "markdown")
	}
	if c.Callback.Delay != "5s" {
		t.Errorf("Callback.Delay = %q, want %q", c.Callback.Delay, "5s")
	}
	if c.Telegram.Token != "" {
		t.Errorf("Telegram.Token = %q, want empty", c.Telegram.Token)
	}
	if c.Callback.Command != "" {
		t.Errorf("Callback.Command = %q, want empty", c.Callback.Command)
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	toml := `[general]
format = "org"

[telegram]
token = "file-token"

[callback]
command = "notify-send"
delay = "10s"
`
	if err := os.WriteFile(path, []byte(toml), 0644); err != nil {
		t.Fatal(err)
	}

	c, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if c.General.Format != "org" {
		t.Errorf("General.Format = %q, want %q", c.General.Format, "org")
	}
	if c.Telegram.Token != "file-token" {
		t.Errorf("Telegram.Token = %q, want %q", c.Telegram.Token, "file-token")
	}
	if c.Callback.Command != "notify-send" {
		t.Errorf("Callback.Command = %q, want %q", c.Callback.Command, "notify-send")
	}
	if c.Callback.Delay != "10s" {
		t.Errorf("Callback.Delay = %q, want %q", c.Callback.Delay, "10s")
	}
}

func TestLoadEnvOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	toml := `[telegram]
token = "file-token"
`
	if err := os.WriteFile(path, []byte(toml), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("COMMS_TELEGRAM_TOKEN", "env-token")

	c, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if c.Telegram.Token != "env-token" {
		t.Errorf("Telegram.Token = %q, want %q", c.Telegram.Token, "env-token")
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.toml")
	if err == nil {
		t.Fatal("Load() expected error for missing file, got nil")
	}
}
