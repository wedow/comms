# Plugin Architecture

Refactor comms into a two-binary model: a provider-agnostic base binary (`comms`) that delegates provider-specific work to standalone provider binaries (`comms-telegram`, etc.) via exec (one-shot) and JSONL over stdio (daemon mode). After implementation, the base binary has zero provider SDK dependencies.

## Architecture

```
                    comms (base binary)
                   /      |        \
            CLI commands  daemon   store
           (send/react)   core    (fs)
              |              |       |
              v              v       v
         exec delegate   JSONL    shared
              |         protocol   pkgs
              v              |       |
    comms-telegram      comms-telegram
    (one-shot)         (subprocess mode)
              |              |
              v              v
         telegram API    telegram API
```

**Base binary**: parse flags, delegate send/react to provider, manage daemon lifecycle, run daemon core loop.

**Provider binary**: one-shot CLI mode (send, react) + daemon subprocess mode (JSONL stdin/stdout protocol).

**Shared packages**: `message`, `store`, `config`, `daemon` (core), `protocol` (JSONL types + codec).

## Phases

| Phase | Name | Tasks | Description |
|-------|------|-------|-------------|
| 00 | Provider Move | 1 | Move telegram package from `internal/provider/` to `providers/` |
| 01 | Shared Infrastructure | 10 | Protocol types/codec, config namespacing, store verification |
| 02 | One-Shot Delegation | 10 | CLI delegation layer, provider binary entry point, base binary cleanup |
| 03 | Daemon Subprocess | 19 | Subprocess manager, daemon core rewrite, callback refactor |
| 04 | Provider Subprocess | 5 | Telegram subprocess mode (JSONL loop), integration tests |
| 05 | Polish | 3 | Documentation, CI updates, final verification |

**Total tasks: 48**

## Dependency Graph

```
Phase 00 (provider move)
    |
    +---> Phase 01 (shared infra) --+
    |                               |
    +---> Phase 02 (one-shot)       |
    |       |                       |
    |       +---+ Phase 01 -------->+
    |           |
    +---> Phase 03 (daemon) <------+
    |       |
    |       +--- Phase 01 (protocol types)
    |       +--- Phase 00 (provider at new path)
    |
    +---> Phase 04 (provider subprocess)
            |
            +--- Phase 00 (provider move)
            +--- Phase 01 (protocol types/codec)
            +--- Phase 02 (provider binary exists)
            +--- Phase 03 (daemon core ready to consume)

Phase 05 (polish) -- depends on all
```

Execution order: 00 -> (01, 02) -> 03 -> 04 -> 05

Phase 01 and Phase 02 can run in parallel. Phase 02 technically depends on Phase 00 but Phase 01 is independent. Phase 03 depends on 01 (protocol types) and 00 (import paths). Phase 04 depends on 02 (provider binary skeleton) and 03 (daemon subprocess manager).

## File Changes Summary

| Phase | New Files | Modified Files |
|-------|-----------|----------------|
| 00 | 0 | 6 (import paths) |
| 01 | 4 | 3 |
| 02 | 4 | 8 |
| 03 | 5 | 5 |
| 04 | 3 | 2 |
| 05 | 0 | 3+ |

## Cross-Phase Issues Resolved

1. **Protocol types duplication**: Phase 1A and original Phase 5 both defined `internal/protocol/`. Consolidated into Phase 01; Phase 04 depends on it.
2. **Provider move duplication**: Phase 0 and original Phase 5 both moved the package. Consolidated into Phase 00.
3. **CLI binary duplication**: Phase 2A and original Phase 5 both created `cmd/comms-telegram/main.go` and `providers/telegram/cli.go`. Consolidated into Phase 02.
4. **Protocol package location**: Architecture doc listed `internal/daemon/protocol.go` for types. Resolved: types live in `internal/protocol/`, daemon's `protocol.go` is the I/O layer (pipe read/write using `internal/protocol` types).
5. **Phase numbering**: Source used 0/1A/1B/1C/2A/2B/3A-3D/5. Remapped to 6 sequential phases (00-05).
