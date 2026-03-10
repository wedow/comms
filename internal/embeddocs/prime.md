# Agent Bootstrap Guide

You are an AI agent with access to `comms`, a filesystem-based messaging relay. This guide covers best practices for bootstrapping your environment and operating effectively.

**Important:** This guide is reference material. Copy any sections relevant to your use case into your project's `CLAUDE.md`, `AGENTS.md`, or equivalent agent instructions file — you will only see this content when you explicitly run `comms prime`.

## How comms Works

comms turns messaging platforms into files. A background daemon polls for messages and writes them as timestamped files in `.comms/`. You interact via CLI commands that output JSON lines.

```
.comms/
  config.toml                      # TOML config (token, callback, format)
  daemon.pid                       # running daemon PID
  offset.json                      # provider polling offset
  docs/                            # embedded setup guides
  <provider>-<channel>/
    .cursor                        # RFC3339Nano timestamp of last read message
    chat_id                        # numeric provider chat ID
    2026-03-01T12-30-00Z.md        # message file
    topic-<id>/                    # thread/topic subdirectory
```

Messages are markdown files with YAML frontmatter containing provider, channel, date, message ID, sender, and optional media metadata.

Run `comms --help` for the full list of commands and flags.

## Session Lifecycle

Every session has three phases: orient, act, persist.

### Orient

1. **Read** — `comms unread` returns new messages as JSON lines.
2. **Gather context** — Search `.comms/` message history and your state files for anything relevant to the conversation. If you have subagents available, fan them out for parallel search rather than searching sequentially yourself — conversation history can be large and keyword-matching benefits from multiple concurrent passes. Do not act in isolation.

### Act

3. **Respond appropriately** — Not every message needs a lengthy reply. Match your response to what the message calls for:
   - A quick reaction emoji (`comms react`) may be all that's needed — a thumbs-up, checkmark, or similar.
   - A short chat reply for simple questions or confirmations.
   - For requests that need research or work: send a brief acknowledgment first ("let me check"), do the work (browse, search, compute, whatever is needed), then send the result.
   - Sometimes the right action is just updating your state files silently — adding a note, updating a list — with a quick confirmation.
4. **Check for new messages** — `comms unread` between actions. Conversations move fast; don't reply to stale context.
5. **Acknowledge** — `comms ack <message-id>` to advance the read cursor past handled messages.

This is a chat, not a document exchange. Keep messages tight and conversational. Multiple short messages beat one wall of text.

### Persist

6. **Update state** — Write anything you learned to your state files.
7. **Commit and push** — `git add`, `git commit`, `git push`. Include everything: incoming messages, outgoing messages, cursor files, state file changes.

Never end a session with uncommitted state changes. Your repo is your memory — data loss means amnesia.

## Telegram Formatting

When sending to Telegram, use `--format markdown` to enable MarkdownV2.

**You must escape these characters** with a preceding `\` in message text (outside code blocks):

```
_ * [ ] ( ) ~ ` > # + - = | { } . !
```

Inside `` ` `` code spans and ``` ``` ``` blocks, only escape `` ` `` and `\`.

Unescaped special characters cause a 400 error and the message is not sent. There is no silent fallback.

## State Management

### Principles

- **Filesystem is your database.** All persistent state lives in files within your git repo.
- **State belongs in files, not in memory.** Each session starts fresh. If it matters, write it down.
- **Commit and push every session.** Your repo is the single source of truth.

### Recommended Structure

Use org-mode or markdown files to track persistent state. Start with a single file and split only when it gets unwieldy.

Organize by topic, not chronologically:

```org
* Preferences
** Kitchen
Household prefers brand X for dish soap.

* Shopping List
- [ ] Bread
- [ ] Milk

* Notes
** Dishwasher
Model number, purchase date, maintenance history.
```

### Cross-Linking

Use links liberally to make state searchable and traceable. Org-mode example:

```org
** Kitchen
Household prefers brand X for dish soap.
[[file:.comms/telegram-dm/2026-03-01T09:15:00.md][2026-03-01 re: dish soap]]
```

Link conversations to state entries. Link state entries to each other. Use keyword-rich descriptions — these are what future searches will match on.

### Citing Messages

When a past conversation is relevant to a reply, use `comms send --reply-to <message-id>` to reply directly to the original message. This gives the user a clickable reference in their chat client.

## Environment Bootstrapping

### Persistent Daemon

The comms daemon must run continuously to receive messages. Set it up as an OS-managed service:

- **Linux (systemd):** `comms daemon start` creates and starts a systemd user unit.
- **macOS (launchd/Homebrew):** Use `brew services start comms` or create a launchd plist.

### Callbacks

Configure a callback in `.comms/config.toml` to be notified of new messages:

```toml
[callback]
command = "/path/to/your/handler"
delay = "3s"
```

The callback receives context via environment variables (`COMMS_CHANNEL`, `COMMS_DIR`, etc.), not arguments. The delay debounces rapid message bursts.

### Persistent Tooling

Consider setting up a persistent process for state management:

- **Cron jobs or systemd timers** — for scheduled self-invocations and follow-ups.
- **Emacs daemon** — for org-mode operations, timers, and structured state manipulation.
- **Any long-lived process** — that can execute deferred tasks on your behalf.

The key idea: you need a way to schedule future actions (reminders, follow-ups, periodic checks) that survives session boundaries.

### Self-Invocation

If something needs checking later, schedule a follow-up rather than relying on the user to remind you. The mechanism depends on your environment (cron, systemd timers, at jobs, etc.), but the pattern is universal:

1. Define the task as a self-contained string (a fresh session will read it with no prior context).
2. Schedule it for the appropriate time.
3. When triggered, start a normal session: read unread, gather context, act.

## Operating Principles

- **Remember everything, surface only what's relevant.** Store aggressively. Reply concisely.
- **Err toward action.** If something clearly needs doing, do it and confirm — don't ask permission for routine actions.
- **Search before replying.** Your message history and state files contain context that this session doesn't have. Always search before composing a response.
- **Delegate expensive work.** Browser automation, large file processing, and deep research should go to subagents/subprocesses to keep your main session responsive.
- **Keep state flat.** A handful of well-named files beats a deep directory tree.
- **Persist your instructions.** Add the patterns you adopt from this guide to your `CLAUDE.md` or `AGENTS.md` so they're loaded automatically every session.
