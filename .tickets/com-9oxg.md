---
id: com-9oxg
status: closed
deps: []
links: []
created: 2026-03-01T13:17:39Z
type: task
priority: 1
assignee: Greg Wedow
tags: [phase-1]
---
# comms-v1: TOML config loading with env var resolution

## Summary
Define the Config struct and implement `Load(path string) (Config, error)`. After loading the TOML file, resolve `Telegram.Token`: if `$COMMS_TELEGRAM_TOKEN` is set, it overrides the file value. Also provide `Default() Config` returning sensible defaults.

Config struct:
```go
type Config struct {
    General  GeneralConfig
    Telegram TelegramConfig
    Callback CallbackConfig
}

type GeneralConfig struct {
    Format string // "markdown" or "org"
}

type TelegramConfig struct {
    Token string
}

type CallbackConfig struct {
    Command string
    Delay   string
}
```

Defaults: `General.Format = "markdown"`, `Callback.Delay = "5s"`, everything else empty string.

Env var resolution: after TOML decode, if `os.Getenv("COMMS_TELEGRAM_TOKEN") != ""`, overwrite `Telegram.Token`.

Use `github.com/BurntSushi/toml` for parsing.

## Files
- `internal/config/config.go` -- struct, Load, Default
- `internal/config/config_test.go` -- tests: load from file, env var override, defaults, missing file error

## Implementation Notes
- Run `go get github.com/BurntSushi/toml` before building
- No dependencies on other tasks
- Default values: `General.Format = "markdown"`, `Callback.Delay = "5s"`, everything else empty string
- Env var `COMMS_TELEGRAM_TOKEN` overrides `Telegram.Token` if set (non-empty)
- Load must return error for missing file

## Verification
```sh
go test ./internal/config/... -v
```

## Acceptance Criteria
- All tests pass
- Load reads TOML correctly
- Env var overrides token
- Default returns expected values
- Missing file returns error

