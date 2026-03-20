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
	Callback  CallbackConfig
	Providers map[string]map[string]any `toml:"providers"`
}

type GeneralConfig struct {
	Format string `toml:"format"`
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
	// Env var override
	if env := os.Getenv("COMMS_TELEGRAM_TOKEN"); env != "" {
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
