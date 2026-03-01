---
id: com-m4pe
status: closed
deps: []
links: []
created: 2026-03-01T13:17:58Z
type: task
priority: 3
assignee: Greg Wedow
tags: [phase-5]
---
# comms-v1: systemd user unit and launchd plist

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
- `dist/com.wedow.comms.plist` (create new)

## Implementation Notes
Static file, no code generation needed. Use `%h` systemd specifier for user home directory. Type=simple since the daemon runs in the foreground. Restart=on-failure with 5 second delay. The AUR PKGBUILD will install a modified copy with `/usr/bin/comms` as the ExecStart path.

## Also: launchd plist (manual fallback for macOS)
A launchd plist for macOS users who install the binary manually (not via Homebrew). Homebrew users get this automatically via the `service` block in the formula.

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

## Verification
```sh
cd /home/greg/p/comms && systemd-analyze verify dist/comms.service 2>&1 || true
test -f dist/comms.service && echo "file exists"
cd /home/greg/p/comms && plutil -lint dist/com.wedow.comms.plist 2>/dev/null || true
test -f dist/com.wedow.comms.plist && echo "file exists"
```

## Acceptance Criteria
Both files exist with correct structure.

