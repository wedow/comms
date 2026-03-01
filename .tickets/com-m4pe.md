---
id: com-m4pe
status: open
deps: []
links: []
created: 2026-03-01T13:17:58Z
type: task
priority: 3
assignee: Greg Wedow
tags: [phase-5]
---
# comms-v1: systemd user unit

## Summary
A systemd user unit file so Linux users can run `systemctl --user enable --now comms`. The unit runs `comms daemon start` in the foreground.

Unit contents:
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

## Files
- `dist/comms.service` (create new)

## Implementation Notes
Static file, no code generation needed. Use `%h` systemd specifier for user home directory. Type=simple since the daemon runs in the foreground. Restart=on-failure with 5 second delay. The AUR PKGBUILD will install a modified copy with `/usr/bin/comms` as the ExecStart path.

## Verification
```sh
cd /home/greg/p/comms && systemd-analyze verify dist/comms.service 2>&1 || true
test -f dist/comms.service && echo "file exists"
```

## Acceptance Criteria
File exists with correct structure.

