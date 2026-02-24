# comms

## Project Vision

comms is a CLI tool and paired daemon that gives AI agents a unified, filesystem-based interface for sending and receiving messages across providers (Telegram first, more later). Messages are plain text files (markdown or org-mode), config is TOML, output is JSON lines, and callbacks are shell commands. Core philosophy: filesystem is the database, Unix composition over frameworks, agent ergonomics over human ergonomics, worse is better. See [docs/VISION.md](docs/VISION.md) for complete vision and decision framework.
