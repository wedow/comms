---
id: com-4r7p
status: open
deps: [com-20s3, com-mnvn, com-gvr2, com-c7vx, com-r9tt, com-tti8, com-op5m, com-0uyd, com-e1kk, com-op12, com-36ho, com-plln, com-un3y, com-ksog, com-p48x, com-xgku, com-kf56, com-sz5e, com-gjjv, com-98be, com-ogzs, com-1lmy, com-265l, com-g5fh, com-ho3r, com-lnh0, com-2nnd, com-7pm4, com-6bya, com-ohfe, com-l3ws, com-bl92, com-w70v, com-fe1f, com-6g26, com-aoxk, com-0cej, com-qcy9, com-pqo8, com-68x1, com-ab0w, com-oqd8, com-jrwi, com-6g4f, com-l7o6, com-3w5g, com-d7ni]
links: []
created: 2026-03-20T01:11:42Z
type: epic
priority: 2
assignee: Greg Wedow
---
# EPIC: Plugin Architecture - Complete Implementation

## Purpose

This epic tracks 100% completion of the plugin architecture implementation — decoupling the Telegram provider from the core binary via a subprocess-based JSONL protocol.

**Do not close this ticket until ALL dependent tickets are complete AND validation passes.**

## Plan Reference

Plan location: docs/feature-plans/plugin-architecture/

## Validation Instructions (REQUIRED before closing)

When all dependent tickets are complete, perform this validation:

### 1. Plan Coverage Check

Read the original plan and verify each planned task was implemented:

```bash
# List all plan phases
ls docs/feature-plans/plugin-architecture/phase-*.md

# Compare against completed tickets
tk ls --status closed | grep "Plugin Arch"
```

For each planned task, verify:
- [ ] A ticket exists that covers this task
- [ ] The ticket is marked closed
- [ ] Implementation matches plan intent (or deviation is justified)

### 2. Verification Commands

Run ALL verification commands from the plan:

```bash
# Build both binaries
go build -o /tmp/comms ./cmd/comms
go build -o /tmp/comms-telegram ./cmd/comms-telegram

# Run all tests
go test ./... -v
go vet ./...

# Verify telegram SDK isolation
grep -r "go-telegram" cmd/comms/ internal/cli/ internal/daemon/ internal/store/ internal/config/ internal/message/  # expect no output
grep -r "go-telegram" providers/telegram/ cmd/comms-telegram/  # expect matches

# Verify binaries work
/tmp/comms --help
/tmp/comms-telegram --help
/tmp/comms-telegram send --help
/tmp/comms-telegram react --help

# Protocol package standalone
go test ./internal/protocol/ -v

# Config works with new format
go test ./internal/config/ -v
```

### 3. Gap Analysis

If gaps are found:
1. Create new ticket(s) for missing work
2. Add them as dependencies to this epic: `tk dep [epic-id] [new-ticket-id]`
3. Do NOT close this epic until new tickets are also complete

### 4. Deviation Documentation

If implementation deviated from plan, document in this ticket:
- What changed and why
- Whether the plan should be updated for reference

## Acceptance Criteria

- [ ] All 47 dependent tickets are closed
- [ ] Plan coverage check: 100% of planned tasks implemented
- [ ] Both binaries build (comms + comms-telegram)
- [ ] All tests pass (go test ./...)
- [ ] go vet clean
- [ ] Zero telegram SDK imports in base binary
- [ ] Protocol package standalone
- [ ] Config works with new providers format
- [ ] No gaps remain (or new tickets created and completed)
- [ ] Deviations documented if any

