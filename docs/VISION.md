# Project Vision: comms

## Problem Statement

AI agents need to communicate across messaging platforms (Telegram, Discord, email) but each provider has its own API, authentication, and message format. Today, agents either get bespoke per-provider integrations or use heavy frameworks that couple messaging to agent-specific concepts. Neither approach composes well with Unix tools, and neither treats messages as what they are: text.

comms solves this by making the filesystem the interface. Messages are plain text files with YAML frontmatter (or org-mode headers). Config is TOML. The daemon is a process. Callbacks are shell commands. An agent that can read files and run CLI commands can use comms with zero SDK and zero framework dependency. A human who can `grep` can inspect every message ever received.

## Users & Use Cases

### Primary Users

**AI agents** operating via CLI. They run `comms unread`, process the output, and `comms send` replies. They may also configure comms from scratch by reading embedded setup docs -- no human hand-holding required.

**Humans** doing initial setup or auditing. They configure `config.toml`, start the daemon via brew services or systemd, and occasionally grep through `.comms/` to inspect message history.

### Core Use Cases

1. **Receive messages**: Daemon polls/listens on configured providers, writes each incoming message to `.comms/$provider-$channel/` as an ISO-8601-timestamped file.
2. **Read unread**: Agent runs `comms unread` to get new messages across all channels.
3. **Send messages**: Agent runs `comms send` targeting a provider/channel with message content.
4. **Trigger callbacks**: Daemon fires an optional shell command on new messages (with mutex/delay to prevent flooding). Callback gets context via env vars, not a template engine.
5. **Bootstrap**: `comms init` scaffolds `.comms/` directory structure, `config.toml`, and embedded provider setup docs in `.comms/docs/`. Agents read these docs and configure themselves.
6. **Manage state**: Agent marks messages read, adds tags, filters history via CLI commands. Advanced queries fall through to raw grep.

## Success Criteria

### Definition of Done

An agent can: `comms init` in a directory, configure Telegram (guided by embedded docs in `.comms/docs/`), start the daemon (brew service or systemd unit), receive messages as files, read unread messages, send replies, and trigger callbacks on incoming messages. The full loop works without human intervention after initial provider credentials are set.

### Quality Bar

Production MVP. Not a prototype.

- Multi-arch binaries via GoReleaser + GitHub Actions
- Homebrew tap for macOS
- AUR package for Arch Linux
- Brew service and systemd user unit for daemon management
- Real error messages, not panics

### Metrics

- **Setup time**: An agent with access to embedded docs can go from `comms init` to sending its first reply in under 5 minutes of wall time.
- **Message latency**: Incoming Telegram messages appear as files within 2 seconds of receipt.
- **Reliability**: Best-effort reliability -- daemon persists polling offset across restarts, but comms is not a message broker and makes no delivery guarantees beyond write-to-disk.

## Design Philosophy

### Guiding Principles

1. **Filesystem is the database**: Messages are files. Config is a file. State is file metadata or frontmatter. No SQLite, no hidden state, no binary formats. If you can `ls` and `grep`, you can inspect everything comms knows.

2. **Provider-agnostic interface**: The CLI never exposes provider-specific concepts. `comms send` works the same whether the backend is Telegram or (eventually) Discord. Provider differences are internal implementation details.

3. **Agent-ergonomic**: Optimized for programmatic use. CLI output is JSON lines (one JSON object per line) by default -- self-describing, composable with Unix tools via `jq`, and trivial for agents to parse. No interactive prompts, no color codes that break parsing, no "are you sure?" confirmations. A machine reads this output.

4. **Zero coupling to agent frameworks**: comms manages messages for whoever is looking at a directory. It has no concept of "tools," "chains," or any AI framework terminology. It is a mailbox, not an agent runtime.

5. **Simple composition**: Callback is `$SHELL -c "$COMMS_CALLBACK"` with env vars. Daemon is a long-running process you manage with your OS service manager. Messages are files you process with any tool. Each piece is independently useful.

6. **Worse is better**: Ship the simple thing. One provider done well beats five done poorly. Iterate based on real agent usage, not speculative requirements.

### Key Trade-offs

| We choose | Over | Because |
|---|---|---|
| Files | Databases | Grep-friendly, zero dependencies, agent-readable |
| CLI | TUI | Agents can't use TUIs |
| One provider done well | Many providers | Refine the interface on real usage first |
| Env var substitution | Template engines | Simpler, composable, no new syntax to learn |
| Embedded docs | Setup wizards | Agents can read docs; wizards require interactivity |
| One daemon per directory | Shared daemon | Isolation, simplicity, no IPC needed |

### Optimization Target

When in doubt, optimize for **agent ergonomics** over human ergonomics. The human touches comms during setup. The agent touches it hundreds of times a day.

## Anti-goals

- **Not a chat client.** No conversation view, no presence indicators, no typing status.
- **Not a message broker.** No queues, no pub/sub, no delivery guarantees beyond "write to disk."
- **Not an agent framework.** Zero agent-framework concepts. No tool definitions, no chain-of-thought hooks, no LLM integration.
- **Not a database-backed system.** If it needs a query planner, it's too complex.
- **Not a provider SDK wrapper.** The CLI abstraction is the interface; agents never touch provider-specific APIs.
- **No interactive setup wizards.** Docs are embedded; agents and humans read them.
- **No TUI in v1.** Future nice-to-have at best.
- **No multi-provider in v1.** Telegram only until the interface is proven.

## Technical Constraints

- **Language**: Go (latest stable). Single binary, fast startup, good concurrency for daemon.
- **Config format**: TOML at `.comms/config.toml`.
- **Directory scope**: `.comms/` is per-directory, like `.git/`. Each directory is an independent comms home with its own config, messages, and daemon. There is no global shared instance.
- **Message format**: Markdown with YAML frontmatter or org-mode. User-configurable, both equally supported. This is a deliberate trade-off: both formats share the same internal message struct with config-driven serialization -- the difference is two small serializers (~100 lines), not two architectures. Files stored at `.comms/$provider-$channel/YYYY-MM-DDTHH:MM:SS.{md,org}`.
- **CLI output format**: JSON lines (one JSON object per line) is the default output format for all CLI commands. Self-describing, composable with `jq` for human readability, trivial for agents to parse.
- **Error handling**: Structured JSON errors on stderr, meaningful exit codes. Agents parse errors the same way they parse output.
- **First provider**: Telegram via Go bot library.
- **Daemon management**: Homebrew services on macOS, systemd user units on Linux.
- **Release pipeline**: GoReleaser via GitHub Actions. Homebrew tap. AUR PKGBUILD.
- **Target platforms**: macOS (arm64, amd64) and Arch Linux (x86_64).
- **Embedded docs**: Deployed to `.comms/docs/` on `comms init`. Provider-specific setup guides live here.

## Decision Framework

When making decisions about comms, apply these rules in order:

- If choosing between simple and clever, **choose simple**. Clever breaks when agents parse it.
- If a feature requires a database, **don't build it**. Reshape the problem to work with files.
- If a feature requires interactivity, **embed docs instead**. Agents read; they don't click.
- If unsure whether to add a CLI flag or let agents use grep, **let them grep**. Add the flag only when multiple agents independently need it.
- If a provider abstraction leaks, **fix the abstraction**, don't add provider-specific CLI flags.
- If a feature is only useful for humans, **defer it**. Agent ergonomics come first.
- If in doubt about scope, **cut it**. Ship Telegram, learn, then expand.
