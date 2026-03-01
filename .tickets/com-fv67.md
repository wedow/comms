---
id: com-fv67
status: open
deps: [com-m4pe]
links: []
created: 2026-03-01T13:18:16Z
type: task
priority: 3
assignee: Greg Wedow
tags: [phase-5]
---
# comms-v1: AUR PKGBUILD

## Summary
A PKGBUILD for the Arch User Repository. Downloads the pre-built Linux amd64 binary from the GitHub release. Installs the binary and a systemd user unit.

PKGBUILD details:
- `pkgname=comms-bin` (binary package convention)
- `source`: GitHub release tarball URL with `$pkgver` substitution
- `package()`: install binary to `/usr/bin/comms`, install systemd unit to `/usr/lib/systemd/user/comms.service` (with `ExecStart=/usr/bin/comms daemon start`)
- `arch=('x86_64')`
- `sha256sums`: updated per release

The systemd unit installed by the PKGBUILD should have `ExecStart=/usr/bin/comms daemon start` (not the `%h/go/bin/comms` path used in the standalone unit from Task 5.3).

## Files
- `dist/aur/PKGBUILD` (create new)
- `dist/aur/.SRCINFO` (generated from PKGBUILD)

## Implementation Notes
Standard AUR binary PKGBUILD pattern. Use `pkgname=comms-bin` to follow AUR binary package convention. Source URL should reference the GitHub release tarball, e.g. `https://github.com/wedow/comms/releases/download/v${pkgver}/comms_${pkgver}_linux_amd64.tar.gz`. The `package()` function should:
1. Install binary: `install -Dm755 comms "$pkgdir/usr/bin/comms"`
2. Install systemd unit: `install -Dm644` a unit file to `"$pkgdir/usr/lib/systemd/user/comms.service"` with `ExecStart=/usr/bin/comms daemon start`

The `.SRCINFO` can be generated with `makepkg --printsrcinfo > .SRCINFO` or written manually matching the PKGBUILD fields.

## Verification
```sh
cd /home/greg/p/comms && namcap dist/aur/PKGBUILD 2>/dev/null || true
test -f dist/aur/PKGBUILD && echo "file exists"
```

## Acceptance Criteria
PKGBUILD exists with correct package name, source URL, and install steps.

