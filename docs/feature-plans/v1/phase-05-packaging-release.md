# Phase 5: Packaging & Release

**Dependencies:** Phases 1-4 for verification commands; Tasks 5.1-5.3, 5.5-5.6 are static config files that can be written without a buildable binary.

---

## Task 5.1: GoReleaser Configuration

**Description:** Configure GoReleaser to produce multi-arch archives with checksums. Targets: macOS arm64, macOS amd64, Linux x86_64. Archives are tar.gz with the binary, LICENSE, and README.

**Files:**
- `.goreleaser.yaml`

**Configuration details:**
- `builds`: single build entry, `main: ./cmd/comms`, GOOS/GOARCH matrix: `darwin/arm64`, `darwin/amd64`, `linux/amd64`
- `archives`: tar.gz format, include `LICENSE*` and `README*`
- `checksum`: sha256 algorithm
- `release`: GitHub release with auto-generated changelog
- `snapshot`: name template with git short hash for non-tag builds
- No Docker images, no nfpm

**Dependencies:** Phases 1-4 (binary must build)

**Verification:**
```sh
cd /home/greg/p/comms && goreleaser check
cd /home/greg/p/comms && goreleaser build --snapshot --clean
```

**Pass/fail:** `goreleaser check` exits 0. `goreleaser build --snapshot --clean` produces binaries for all three targets.

---

## Task 5.2: GitHub Actions CI Workflow

**Description:** Single workflow file with two jobs: (1) test on every push and PR, (2) release on version tags via GoReleaser.

**Files:**
- `.github/workflows/ci.yaml`

**Job: test:**
- Trigger: push to any branch, pull requests
- Steps: checkout, setup-go (match go.mod version), `go test ./...`, `go vet ./...`

**Job: release:**
- Trigger: tags matching `v*`
- Needs: test (runs tests first)
- Steps: checkout (with `fetch-depth: 0` for changelog), setup-go, goreleaser-action with `GITHUB_TOKEN`

**Dependencies:** Task 5.1 (GoReleaser config)

**Verification:**
```sh
# Validate YAML syntax:
cd /home/greg/p/comms && python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yaml'))" && echo "valid YAML"

# If actionlint is installed:
cd /home/greg/p/comms && actionlint .github/workflows/ci.yaml 2>/dev/null || true
```

**Pass/fail:** YAML parses without error. Workflow contains `test` and `release` jobs.

---

## Task 5.3: systemd User Unit

**Description:** A systemd user unit file so Linux users can run `systemctl --user enable --now comms`. The unit runs `comms daemon start` in the foreground.

**Files:**
- `dist/comms.service`

**Unit contents:**
```ini
[Unit]
Description=comms messaging daemon
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%h/go/bin/comms daemon start
WorkingDirectory=%h
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
```

Note: `ExecStart` path varies by installation method. AUR PKGBUILD (Task 5.5) installs its own copy with `/usr/bin/comms`.

**Dependencies:** none (static file)

**Verification:**
```sh
cd /home/greg/p/comms && systemd-analyze verify dist/comms.service 2>&1 || true
# Will warn about %h not resolving outside a user session -- expected
test -f dist/comms.service && echo "file exists"
```

**Pass/fail:** File exists with correct structure.

---

## Task 5.4: Homebrew Tap

**Description:** Configure GoReleaser to publish to the existing Homebrew tap at `wedow/homebrew-tools`. GoReleaser auto-updates the formula on release — no manual publish step needed. The CI release job (Task 5.2) handles this automatically when a `v*` tag is pushed.

**Files:**
- `.goreleaser.yaml` -- add `brews` section (modify existing file from Task 5.1)

**GoReleaser brews config:**
```yaml
brews:
  - repository:
      owner: wedow
      name: homebrew-tools
    directory: Formula
    homepage: https://github.com/wedow/comms
    description: Filesystem-based messaging interface for AI agents
    license: MIT
    install: |
      bin.install "comms"
    service: |
      run [opt_bin/"comms", "daemon", "start"]
      keep_alive true
      working_dir Dir.home
      log_path var/"log/comms.log"
      error_log_path var/"log/comms.log"
```

The `service` block generates a launchd plist automatically -- Homebrew translates it into `~/Library/LaunchAgents/homebrew.mxcl.comms.plist`.

**Dependencies:** Task 5.1 (GoReleaser config must exist)

**Verification:**
```sh
cd /home/greg/p/comms && goreleaser check
```

**Pass/fail:** `goreleaser check` exits 0 after adding the brews section.

---

## Task 5.5: AUR PKGBUILD

**Description:** A PKGBUILD for the Arch User Repository. Downloads the pre-built Linux amd64 binary from the GitHub release. Installs the binary and a systemd user unit.

**Files:**
- `dist/aur/PKGBUILD`
- `dist/aur/.SRCINFO` (generated from PKGBUILD)

**PKGBUILD details:**
- `pkgname=comms-bin` (binary package convention)
- `source`: GitHub release tarball URL with `$pkgver` substitution
- `package()`: install binary to `/usr/bin/comms`, install systemd unit to `/usr/lib/systemd/user/comms.service` (with `ExecStart=/usr/bin/comms daemon start`)
- `arch=('x86_64')`
- `sha256sums`: updated per release

**Dependencies:** Task 5.3 (ships the systemd unit)

**Verification:**
```sh
# If namcap is available (Arch Linux):
cd /home/greg/p/comms && namcap dist/aur/PKGBUILD 2>/dev/null || true
test -f dist/aur/PKGBUILD && echo "file exists"
```

**Pass/fail:** PKGBUILD exists with correct package name, source URL, and install steps.

---

## Task 5.6: launchd Plist (Manual Fallback)

**Description:** A launchd plist for macOS users who install the binary manually (not via Homebrew). Homebrew users get this automatically via the `service` block in the formula (Task 5.4).

**Files:**
- `dist/com.wedow.comms.plist`

**Plist contents:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.wedow.comms</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/comms</string>
        <string>daemon</string>
        <string>start</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/comms.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/comms.log</string>
</dict>
</plist>
```

Install: `cp dist/com.wedow.comms.plist ~/Library/LaunchAgents/ && launchctl load ~/Library/LaunchAgents/com.wedow.comms.plist`

**Dependencies:** none (static file)

**Verification:**
```sh
cd /home/greg/p/comms && plutil -lint dist/com.wedow.comms.plist 2>/dev/null || true
test -f dist/com.wedow.comms.plist && echo "file exists"
```

**Pass/fail:** File exists with valid XML structure.

---

## Task Dependency Graph

```
Task 5.1 (GoReleaser config)
  |
  +---> Task 5.2 (GitHub Actions CI)
  +---> Task 5.4 (Homebrew tap)

Task 5.3 (systemd unit) -- no deps
  |
  +---> Task 5.5 (AUR PKGBUILD)

Task 5.6 (launchd plist) -- no deps
```

Tasks 5.1, 5.3, 5.6 can proceed in parallel. Task 5.2 and 5.4 depend on 5.1. Task 5.5 depends on 5.3.
