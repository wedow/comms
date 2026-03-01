---
id: com-e5ad
status: open
deps: []
links: []
created: 2026-03-01T13:18:22Z
type: task
priority: 3
assignee: Greg Wedow
tags: [phase-5]
---
# comms-v1: launchd plist (manual fallback)

## Summary
A launchd plist for macOS users who install the binary manually (not via Homebrew). Homebrew users get this automatically via the `service` block in the formula (Task 5.4).

Plist contents:
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

## Files
- `dist/com.wedow.comms.plist` (create new)

## Implementation Notes
Static XML file. Label is `com.wedow.comms`. Binary path is `/usr/local/bin/comms` (standard manual install location on macOS). RunAtLoad and KeepAlive both true. Logs go to `/tmp/comms.log`. This is a fallback for non-Homebrew users -- Homebrew users get an auto-generated plist from the formula's service block.

## Verification
```sh
cd /home/greg/p/comms && plutil -lint dist/com.wedow.comms.plist 2>/dev/null || true
test -f dist/com.wedow.comms.plist && echo "file exists"
```

## Acceptance Criteria
File exists with valid XML structure.

