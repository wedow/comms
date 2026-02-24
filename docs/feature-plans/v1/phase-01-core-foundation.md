# Phase 1: Core Foundation

Leaf packages and project scaffolding. Everything here has zero project-internal dependencies. Later phases build on these packages.

## Task 1.1: Go project skeleton and Cobra CLI root

**Description:** Create the binary entrypoint and Cobra root command with `--version` flag. This establishes the build and gives every later task a place to wire into.

**Files:**
- `cmd/comms/main.go` -- calls `cli.Execute()`
- `internal/cli/root.go` -- Cobra root command (name "comms", short description, version flag)

**Dependencies:** none

**Verification:**
```sh
go build -o /dev/null ./cmd/comms
go build ./cmd/comms && ./cmd/comms --version | grep -qE '^comms'
go build ./cmd/comms && ./cmd/comms --help | grep -q 'comms'
```

**Pass/fail:** Binary builds without error. `--version` prints a version string. `--help` prints usage.

---

## Task 1.2: `internal/message` -- Message struct and markdown serialization

**Description:** Define the `Message` struct and implement markdown+YAML frontmatter round-trip serialization.

Struct fields (exported):
- `From` string
- `Provider` string
- `Channel` string
- `Date` time.Time
- `ID` string
- `Body` string

Frontmatter keys (lowercase, matching struct): `from`, `provider`, `channel`, `date`, `id`. The body is everything after the closing `---`.

Functions:
- `MarshalMarkdown(msg Message) ([]byte, error)` -- render YAML frontmatter + body
- `UnmarshalMarkdown(data []byte) (Message, error)` -- parse frontmatter + body back to Message

Use `gopkg.in/yaml.v3` for frontmatter writing and `github.com/adrg/frontmatter` for parsing.

**Files:**
- `internal/message/message.go` -- struct definition
- `internal/message/markdown.go` -- MarshalMarkdown / UnmarshalMarkdown
- `internal/message/markdown_test.go` -- round-trip tests, edge cases (empty body, multiline body, special characters in fields)

**Dependencies:** none

**Verification:**
```sh
go test ./internal/message/... -run TestMarkdown -v
```

**Pass/fail:** All tests pass. Round-trip marshal/unmarshal produces identical Message values.

---

## Task 1.3: `internal/message` -- Org-mode serialization

**Description:** Add org-mode round-trip serialization to the message package. Org-mode uses `#+KEYWORD: value` lines (all-caps keywords). A blank line separates headers from body.

Keywords: `FROM`, `PROVIDER`, `CHANNEL`, `DATE`, `ID`.

Functions:
- `MarshalOrg(msg Message) ([]byte, error)` -- render org keyword lines + body
- `UnmarshalOrg(data []byte) (Message, error)` -- parse keyword lines + body

Org-mode parsing is ~30 lines of custom code (no external dependency needed). Scan lines starting with `#+`, split on first `:`, trim. First line not matching `#+` (or blank line after headers) begins body.

**Files:**
- `internal/message/org.go` -- MarshalOrg / UnmarshalOrg
- `internal/message/org_test.go` -- round-trip tests, same edge cases as markdown

**Dependencies:** Task 1.2 (shares the Message struct)

**Verification:**
```sh
go test ./internal/message/... -run TestOrg -v
```

**Pass/fail:** All tests pass. Round-trip marshal/unmarshal produces identical Message values.

---

## Task 1.4: `internal/config` -- TOML config loading with env var resolution

**Description:** Define the `Config` struct and implement `Load(path string) (Config, error)`. After loading the TOML file, resolve `Telegram.Token`: if `$COMMS_TELEGRAM_TOKEN` is set, it overrides the file value. Also provide `Default() Config` returning sensible defaults.

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

**Files:**
- `internal/config/config.go` -- struct, Load, Default
- `internal/config/config_test.go` -- tests: load from file, env var override, defaults, missing file error

**Dependencies:** none

**Verification:**
```sh
go test ./internal/config/... -v
```

**Pass/fail:** All tests pass. Load reads TOML correctly. Env var overrides token. Default returns expected values. Missing file returns error.

---

## Task 1.5: `internal/store` -- Directory creation and message write/read

**Description:** Implement the filesystem operations for the `.comms/` tree. This task covers init, write, and read.

Functions:
- `InitDir(root string) error` -- create `root/` and `root/docs/` directories; caller (the init command) handles writing config and docs separately
- `WriteMessage(root string, msg message.Message, format string) (string, error)` -- serialize `msg` using the given format ("markdown" or "org"), write to `root/<provider>-<channel>/<timestamp>.<ext>`, return the written path. Create the channel directory if it doesn't exist. Filename: RFC3339Nano timestamp with colons replaced by hyphens (filesystem-safe). Extension: `.md` for markdown, `.org` for org.
- `ReadMessage(path string) (message.Message, error)` -- detect format from extension (.md -> markdown, .org -> org), read file, unmarshal

**Files:**
- `internal/store/store.go` -- InitDir, WriteMessage, ReadMessage, helper for filename generation
- `internal/store/store_test.go` -- tests using `t.TempDir()`: init creates dirs, write creates file with correct content, read round-trips, unknown extension errors

**Dependencies:** Task 1.2 (message.MarshalMarkdown/UnmarshalMarkdown), Task 1.3 (message.MarshalOrg/UnmarshalOrg)

**Verification:**
```sh
go test ./internal/store/... -run 'TestInit|TestWrite|TestRead' -v
```

**Pass/fail:** All tests pass. InitDir creates expected directories. WriteMessage creates correctly-named file with correct content. ReadMessage round-trips through WriteMessage.

---

## Task 1.6: `internal/store` -- Cursor management and channel/message listing

**Description:** Add cursor files and directory enumeration to the store package.

Functions:
- `ReadCursor(root, channel string) (time.Time, error)` -- read `root/<channel>/.cursor`, parse as RFC3339Nano. Return zero time if file doesn't exist (meaning: all messages are unread).
- `WriteCursor(root, channel string, t time.Time) error` -- write RFC3339Nano timestamp to `root/<channel>/.cursor`
- `ListChannels(root string) ([]string, error)` -- return sorted directory names under `root/` that look like channel dirs (exclude `docs/`, files)
- `ListMessages(root, channel string) ([]string, error)` -- return sorted file paths in `root/<channel>/` (exclude `.cursor`), oldest first
- `ListMessagesAfter(root, channel string, after time.Time) ([]string, error)` -- like ListMessages but only files with timestamp > `after`

Sorting: by filename (which is a timestamp), lexicographic ascending = chronological ascending.

**Files:**
- `internal/store/cursor.go` -- ReadCursor, WriteCursor
- `internal/store/list.go` -- ListChannels, ListMessages, ListMessagesAfter
- `internal/store/cursor_test.go` -- cursor round-trip, missing file returns zero time
- `internal/store/list_test.go` -- listing with multiple channels and messages, filtering by time

**Dependencies:** Task 1.5 (uses InitDir and WriteMessage in tests to set up fixtures)

**Verification:**
```sh
go test ./internal/store/... -run 'TestCursor|TestList' -v
```

**Pass/fail:** All tests pass. Cursor round-trips correctly. Missing cursor returns zero time. ListChannels excludes `docs/`. ListMessagesAfter filters correctly.

---

## Task 1.7: Embedded docs

**Description:** Create the `embed/` directory with `telegram-setup.md` -- the setup guide that `comms init` deploys to `.comms/docs/`. Use Go's `//go:embed` directive to bundle the file into the binary.

The guide covers: creating a Telegram bot via BotFather, obtaining the token, adding the bot to groups, and configuring `config.toml`. Written for an AI agent reader (direct instructions, no screenshots).

**Files:**
- `embed/embed.go` -- `//go:embed telegram-setup.md` as `var TelegramSetupDoc []byte` (or `string`)
- `embed/telegram-setup.md` -- setup guide content

**Dependencies:** none

**Verification:**
```sh
go build ./embed
go vet ./embed
```

**Pass/fail:** Package builds and vets without error.

---

## Task 1.8: `comms init` command

**Description:** Implement the `init` subcommand. It scaffolds the `.comms/` directory tree, writes a default `config.toml`, and deploys embedded docs.

Behavior:
1. Resolve root as `.comms/` in the current working directory (or a `--dir` flag).
2. Call `store.InitDir(root)`.
3. Write `config.Default()` as TOML to `root/config.toml`. Skip if file already exists (don't overwrite user config).
4. Write `embed.TelegramSetupDoc` to `root/docs/telegram-setup.md`. Always overwrite (docs may update between versions).
5. Print JSON `{"status":"initialized","path":"<abs path>"}` to stdout.

Wire the subcommand into the Cobra root command.

**Files:**
- `internal/cli/init.go` -- init command definition and handler
- `internal/cli/init_test.go` -- test: init creates expected directory tree, config.toml exists with defaults, docs deployed, second init doesn't clobber config

**Dependencies:** Task 1.1 (root command), Task 1.4 (config.Default + TOML serialization), Task 1.5 (store.InitDir), Task 1.7 (embedded docs)

**Verification:**
```sh
go test ./internal/cli/... -run TestInit -v
go build ./cmd/comms && ./cmd/comms init --help | grep -q 'init'
```

**Pass/fail:** All tests pass. `init --help` shows usage. Init creates `.comms/` tree with config and docs.

---

## Task 1.9: External dependencies and go.sum

**Description:** Add all external dependencies needed by Phase 1. Run `go mod tidy` to populate `go.sum`. Verify the full project compiles and all tests pass.

**Files:**
- `go.mod` (updated with dependencies)
- `go.sum` (generated)

**Dependencies:** Tasks 1.1-1.8 (all code must exist before tidy resolves everything cleanly)

**Verification:**
```sh
go mod tidy
go build ./...
go test ./...
go vet ./...
```

**Pass/fail:** All four commands exit 0. No compilation errors, no test failures, no vet warnings.

---

## Task Dependency Graph

```
Task 1.1 (CLI skeleton)
Task 1.2 (message struct + markdown)
Task 1.4 (config)
Task 1.7 (embedded docs)
   |         |
   |   Task 1.3 (org) -- depends on 1.2
   |         |
   |   Task 1.5 (store write/read) -- depends on 1.2, 1.3
   |         |
   |   Task 1.6 (store cursor/list) -- depends on 1.5
   |         |
   +---------+
         |
   Task 1.8 (comms init) -- depends on 1.1, 1.4, 1.5, 1.7
         |
   Task 1.9 (go mod tidy) -- depends on 1.1-1.8
```

Tasks 1.1, 1.2, 1.4, 1.7 can proceed in parallel (no interdependencies).
