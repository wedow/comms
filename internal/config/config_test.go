package config

import (
	"encoding/json"
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
	if c.Providers == nil {
		t.Fatal("Providers map is nil, want populated")
	}
	tg, ok := c.Providers["telegram"]
	if !ok {
		t.Fatal("Providers[\"telegram\"] not found")
	}
	if tg["token"] != "" {
		t.Errorf("Providers[\"telegram\"][\"token\"] = %v, want empty string", tg["token"])
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
	// Migration: old [telegram] should populate Providers["telegram"]
	if c.Providers == nil {
		t.Fatal("Providers map is nil after loading old [telegram] config")
	}
	tg, ok := c.Providers["telegram"]
	if !ok {
		t.Fatal("Providers[\"telegram\"] not populated from [telegram] migration")
	}
	if tg["token"] != "file-token" {
		t.Errorf("Providers[\"telegram\"][\"token\"] = %v, want %q", tg["token"], "file-token")
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
	// Env override should also populate Providers["telegram"]["token"]
	if c.Providers == nil {
		t.Fatal("Providers map is nil after env override")
	}
	tg, ok := c.Providers["telegram"]
	if !ok {
		t.Fatal("Providers[\"telegram\"] not found after env override")
	}
	if tg["token"] != "env-token" {
		t.Errorf("Providers[\"telegram\"][\"token\"] = %v, want %q", tg["token"], "env-token")
	}
}

func TestLoadProvidersOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	tomlData := `[general]
format = "markdown"

[providers.telegram]
token = "new-format-token"
`
	if err := os.WriteFile(path, []byte(tomlData), 0644); err != nil {
		t.Fatal(err)
	}

	c, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if c.Providers == nil {
		t.Fatal("Providers map is nil")
	}
	tg, ok := c.Providers["telegram"]
	if !ok {
		t.Fatal("Providers[\"telegram\"] not found")
	}
	if tg["token"] != "new-format-token" {
		t.Errorf("Providers[\"telegram\"][\"token\"] = %v, want %q", tg["token"], "new-format-token")
	}
	// Legacy Telegram.Token should be empty (no [telegram] section)
	if c.Telegram.Token != "" {
		t.Errorf("Telegram.Token = %q, want empty (no legacy section)", c.Telegram.Token)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.toml")
	if err == nil {
		t.Fatal("Load() expected error for missing file, got nil")
	}
}

func TestProviderConfig(t *testing.T) {
	c := Config{
		Providers: map[string]map[string]any{
			"telegram": {"token": "abc123", "timeout": 30.0},
		},
	}

	t.Run("valid provider", func(t *testing.T) {
		data, err := c.ProviderConfig("telegram")
		if err != nil {
			t.Fatalf("ProviderConfig() error: %v", err)
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("json.Unmarshal() error: %v", err)
		}
		if m["token"] != "abc123" {
			t.Errorf("token = %v, want %q", m["token"], "abc123")
		}
		if m["timeout"] != 30.0 {
			t.Errorf("timeout = %v, want %v", m["timeout"], 30.0)
		}
	})

	t.Run("unknown provider", func(t *testing.T) {
		_, err := c.ProviderConfig("slack")
		if err == nil {
			t.Fatal("ProviderConfig() expected error for unknown provider, got nil")
		}
	})
}

func TestProviderNames(t *testing.T) {
	t.Run("sorted names", func(t *testing.T) {
		c := Config{
			Providers: map[string]map[string]any{
				"telegram": {},
				"slack":    {},
				"discord":  {},
			},
		}
		names := c.ProviderNames()
		want := []string{"discord", "slack", "telegram"}
		if len(names) != len(want) {
			t.Fatalf("ProviderNames() len = %d, want %d", len(names), len(want))
		}
		for i, name := range names {
			if name != want[i] {
				t.Errorf("ProviderNames()[%d] = %q, want %q", i, name, want[i])
			}
		}
	})

	t.Run("empty map", func(t *testing.T) {
		c := Config{}
		names := c.ProviderNames()
		if len(names) != 0 {
			t.Errorf("ProviderNames() len = %d, want 0", len(names))
		}
	})
}

func TestTelegramToken(t *testing.T) {
	t.Run("returns token from Providers map", func(t *testing.T) {
		c := Config{
			Providers: map[string]map[string]any{
				"telegram": {"token": "my-token"},
			},
		}
		if got := c.TelegramToken(); got != "my-token" {
			t.Errorf("TelegramToken() = %q, want %q", got, "my-token")
		}
	})

	t.Run("nil Providers", func(t *testing.T) {
		c := Config{}
		if got := c.TelegramToken(); got != "" {
			t.Errorf("TelegramToken() = %q, want empty", got)
		}
	})

	t.Run("telegram not in Providers", func(t *testing.T) {
		c := Config{
			Providers: map[string]map[string]any{
				"slack": {"token": "slack-token"},
			},
		}
		if got := c.TelegramToken(); got != "" {
			t.Errorf("TelegramToken() = %q, want empty", got)
		}
	})

	t.Run("token key missing", func(t *testing.T) {
		c := Config{
			Providers: map[string]map[string]any{
				"telegram": {"timeout": 30},
			},
		}
		if got := c.TelegramToken(); got != "" {
			t.Errorf("TelegramToken() = %q, want empty", got)
		}
	})

	t.Run("token not a string", func(t *testing.T) {
		c := Config{
			Providers: map[string]map[string]any{
				"telegram": {"token": 12345},
			},
		}
		if got := c.TelegramToken(); got != "" {
			t.Errorf("TelegramToken() = %q, want empty", got)
		}
	})
}
