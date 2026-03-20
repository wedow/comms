---
id: com-bl92
status: open
deps: [com-ho3r, com-lnh0, com-2nnd, com-7pm4, com-6bya, com-ohfe, com-l3ws]
links: []
created: 2026-03-20T01:06:07Z
type: task
priority: 2
assignee: Greg Wedow
---
# Plugin Arch: Remove interfaces and replace TypingIndicator atomically

## Summary
These changes must land together -- deleting an interface before updating its sole user breaks the build. Atomic steps: (1) Define `TypingFunc func(ctx context.Context, provider string, chatID int64) error` in `callback.go`. (2) Replace `TypingIndicator` interface field and constructor parameter in `CallbackRunner` with `TypingFunc`. (3) Delete `TypingIndicator` interface from `callback.go`. (4) Delete `Provider` interface from `daemon.go`. (5) Replace `daemon.Run()` signature with `Run(ctx, cfg, root, providers []string) error` (stub for now). The daemon will implement `sendTypingCommand` (a closure matching `TypingFunc`) that writes `TypingCommand` to the correct provider's stdin, added in Task 03-12.

## Files
- Modify: `internal/daemon/callback.go`
- Modify: `internal/daemon/daemon.go`

## Implementation Notes
- TypingFunc replaces the TypingIndicator interface entirely
- Provider interface is deleted from daemon.go
- daemon.Run() gets a new signature: Run(ctx, cfg, root, providers []string) error
- The sendTypingCommand closure that writes TypingCommand to stdin is NOT added here -- that comes in the daemon core loop task
- All changes must be atomic to avoid breaking the build

## Verification
```bash
go test ./internal/daemon/ -v
```

## Acceptance Criteria
- Provider interface deleted
- TypingIndicator replaced with TypingFunc
- daemon.Run() has new signature

