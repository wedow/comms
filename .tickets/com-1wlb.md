---
id: com-1wlb
status: closed
deps: []
links: []
created: 2026-03-01T13:17:22Z
type: task
priority: 1
assignee: Greg Wedow
tags: [phase-1]
---
# comms-v1: Message struct and markdown serialization

## Summary
Define the `Message` struct and implement markdown+YAML frontmatter round-trip serialization.

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

## Files
- `internal/message/message.go` -- struct definition
- `internal/message/markdown.go` -- MarshalMarkdown / UnmarshalMarkdown
- `internal/message/markdown_test.go` -- round-trip tests, edge cases (empty body, multiline body, special characters in fields)

## Implementation Notes
- Run `go get gopkg.in/yaml.v3 github.com/adrg/frontmatter` before building
- Frontmatter keys must be lowercase and match struct field names exactly
- Body is everything after the closing `---` separator
- Edge cases to test: empty body, multiline body, special characters in fields

## Verification
```sh
go test ./internal/message/... -run TestMarkdown -v
```

## Acceptance Criteria
- All tests pass
- Round-trip marshal/unmarshal produces identical Message values

