---
id: com-gt6y
status: open
deps: [com-b40d, com-1wlb, com-4nj4, com-9oxg, com-6qs4, com-yvif, com-alwu, com-bx3c, com-25pr, com-66kj, com-77c4, com-qtjh, com-n1t1, com-mgu1, com-0q0l, com-zl1e, com-mwpq, com-0ljk, com-pwp2, com-4409, com-9par, com-fvwr, com-hogs, com-yi8z, com-ryd3, com-bz1j, com-m4pe, com-fv67]
links: []
created: 2026-03-01T13:19:41Z
type: epic
priority: 2
assignee: Greg Wedow
tags: [epic, comms-v1]
---
# EPIC: comms v1 - complete implementation

## Purpose

This epic tracks 100% completion of the comms v1 implementation.

**Do not close this ticket until ALL dependent tickets are complete AND validation passes.**

## Plan Reference

Plan location: docs/feature-plans/v1/

## Validation Instructions (REQUIRED before closing)

When all dependent tickets are complete, perform this validation:

### 1. Plan Coverage Check

Read the original plan and verify each planned task was implemented:

```bash
# List all completed tickets
tk closed | grep comms-v1
# Compare count against expected 37 tasks
```

For each planned task, verify:
- A ticket exists that covers this task
- The ticket is marked closed
- Implementation matches plan intent (or deviation is justified)

### 2. End-to-End Verification

Run the full verification sequence:

```bash
# Build
go build ./cmd/comms

# Full test suite
go test ./...

# Verify CLI commands work
./comms --version
./comms init --dir /tmp/comms-test
./comms channels --dir /tmp/comms-test
./comms list --dir /tmp/comms-test
./comms unread --dir /tmp/comms-test

# Verify daemon commands exist
./comms daemon start --help
./comms daemon stop --help
./comms daemon status --help

# Verify packaging configs
goreleaser check
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yaml'))"
test -f dist/comms.service
test -f dist/aur/PKGBUILD
test -f dist/com.wedow.comms.plist

# Cleanup
rm -rf /tmp/comms-test
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

- All dependent tickets are closed
- Plan coverage check: 100% of planned tasks implemented
- All verification commands pass
- No gaps remain (or new tickets created and completed)
- Deviations documented if any

