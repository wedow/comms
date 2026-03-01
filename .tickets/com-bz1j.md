---
id: com-bz1j
status: open
deps: [com-ryd3]
links: []
created: 2026-03-01T13:17:51Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-5]
---
# comms-v1: GitHub Actions CI workflow

## Summary
Single GitHub Actions workflow file with two jobs: (1) test on every push and PR, (2) release on version tags via GoReleaser.

Job: test:
- Trigger: push to any branch, pull requests
- Steps: checkout, setup-go (match go.mod version), `go test ./...`, `go vet ./...`

Job: release:
- Trigger: tags matching `v*`
- Needs: test (runs tests first)
- Steps: checkout (with `fetch-depth: 0` for changelog), setup-go, goreleaser-action with `GITHUB_TOKEN`

## Files
- `.github/workflows/ci.yaml` (create new)

## Implementation Notes
Use `actions/checkout@v4`, `actions/setup-go@v5`, `goreleaser/goreleaser-action@v6`. The test job triggers on push and pull_request. The release job triggers on tags `v*` and uses `needs: test`. The release job checkout must use `fetch-depth: 0` so GoReleaser can generate changelog from git history. Pass `GITHUB_TOKEN` as env to goreleaser-action.

## Verification
```sh
cd /home/greg/p/comms && python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yaml'))" && echo "valid YAML"
cd /home/greg/p/comms && actionlint .github/workflows/ci.yaml 2>/dev/null || true
```

## Acceptance Criteria
YAML parses without error. Workflow contains `test` and `release` jobs.

