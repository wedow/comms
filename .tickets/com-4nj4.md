---
id: com-4nj4
status: open
deps: [com-1wlb]
links: []
created: 2026-03-01T13:17:30Z
type: task
priority: 1
assignee: Greg Wedow
tags: [phase-1]
---
# comms-v1: Org-mode serialization

## Summary
Add org-mode round-trip serialization to the message package. Org-mode uses `#+KEYWORD: value` lines (all-caps keywords). A blank line separates headers from body.

Keywords: `FROM`, `PROVIDER`, `CHANNEL`, `DATE`, `ID`.

Functions:
- `MarshalOrg(msg Message) ([]byte, error)` -- render org keyword lines + body
- `UnmarshalOrg(data []byte) (Message, error)` -- parse keyword lines + body

Org-mode parsing is ~30 lines of custom code (no external dependency needed). Scan lines starting with `#+`, split on first `:`, trim. First line not matching `#+` (or blank line after headers) begins body.

## Files
- `internal/message/org.go` -- MarshalOrg / UnmarshalOrg
- `internal/message/org_test.go` -- round-trip tests, same edge cases as markdown

## Implementation Notes
- Depends on Task 1.2 (Message struct must exist first)
- No external dependencies needed for org-mode parsing
- Keywords are ALL-CAPS: `FROM`, `PROVIDER`, `CHANNEL`, `DATE`, `ID`
- Format: `#+KEYWORD: value` (one per line)
- A blank line separates headers from body
- Parsing: scan lines starting with `#+`, split on first `:`, trim. First line not matching `#+` (or blank line after headers) begins body
- Edge cases to test: empty body, multiline body, special characters in fields

## Verification
```sh
go test ./internal/message/... -run TestOrg -v
```

## Acceptance Criteria
- All tests pass
- Round-trip marshal/unmarshal produces identical Message values

