package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/BurntSushi/toml"
)

type Config struct {
	General   GeneralConfig
	Telegram  TelegramConfig
	Callback  CallbackConfig
	Providers map[string]map[string]any `toml:"providers"`
}

type GeneralConfig struct {
	Format string `toml:"format"`
}

type TelegramConfig struct {
	Token string `toml:"token"`
}

type CallbackConfig struct {
	Command string `toml:"command"`
	Delay   string `toml:"delay"`
}

func Default() Config {
	return Config{
		General:  GeneralConfig{Format: "markdown"},
		Callback: CallbackConfig{Delay: "5s"},
		Providers: map[string]map[string]any{
			"telegram": {"token": ""},
		},
	}
}

// TelegramToken returns the Telegram token from the Providers map.
// Returns "" if not configured or not a string.
func (c *Config) TelegramToken() string {
	if c.Providers == nil {
		return ""
	}
	tg, ok := c.Providers["telegram"]
	if !ok {
		return ""
	}
	token, ok := tg["token"].(string)
	if !ok {
		return ""
	}
	return token
}

func (c *Config) ProviderConfig(provider string) ([]byte, error) {
	p, ok := c.Providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider %q not found", provider)
	}
	return json.Marshal(p)
}

func (c *Config) ProviderNames() []string {
	names := make([]string, 0, len(c.Providers))
	for k := range c.Providers {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func Load(path string) (Config, error) {
	var c Config
	if _, err := toml.DecodeFile(path, &c); err != nil {
		return Config{}, err
	}
	// Migration: copy [telegram] to [providers.telegram] if not already set
	if c.Telegram.Token != "" {
		if c.Providers == nil {
			c.Providers = make(map[string]map[string]any)
		}
		if c.Providers["telegram"] == nil {
			c.Providers["telegram"] = map[string]any{"token": c.Telegram.Token}
		}
	}
	// Env var override
	if env := os.Getenv("COMMS_TELEGRAM_TOKEN"); env != "" {
		c.Telegram.Token = env
		if c.Providers == nil {
			c.Providers = make(map[string]map[string]any)
		}
		if c.Providers["telegram"] == nil {
			c.Providers["telegram"] = make(map[string]any)
		}
		c.Providers["telegram"]["token"] = env
	}
	return c, nil
}
