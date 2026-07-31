# Changelog

All notable changes to wade. This file is designed to be **AI-friendly** — structured so that Copilot, Claude, Codex, and other AI tools can quickly understand the project's evolution and current state.

---

## [v0.2.1] — 2026-07-31

### Added

- **M8 Distribution complete**: Homebrew tap `wadefengx/tap` + Scoop bucket `wadefengx/scoop-wade` live with production formula/manifest (real SHA256 digests)
- **GitHub Pages site** live at https://wadefengx.github.io/wade/ (multi-runtime features, install guide, command cheatsheet, milestone status)
- **release.yml version injection**: `-X main.version=<tag>` via `main.version` var
- **release.yml update-tap job**: auto-updates tap/bucket manifests on release via `scripts/release-shas.sh --update` (needs `TAP_TOKEN` secret)
- **install.sh now uploaded** as release asset (fixes `releases/latest/download/install.sh` 404)
- **scripts/release-shas.sh**: fetch latest release digests, print or `--update` templates

### Fixed

- `update-tap` job used dead sed placeholders (templates were hardcoded) → now uses release-shas.sh
- `docs/index.html` refreshed for multi-runtime + new commands (M0-M8 status table)

---

## [v0.2.0] — 2026-07-28

### Added

- **Interactive setup wizard** (`wade -i` / `wade init`)
  - 3-step flow: Node mirror → version → PATH
  - `wade init -y` for non-interactive auto-config
  - `wade init` (no flags) writes `.wade-version`
- **Node download mirror** (`wade node mirror`)
  - `wade node mirror official` → nodejs.org
  - `wade node mirror mirror` → npmmirror.com (default)
  - Separated from registry management — two independent concepts
- **Emoji-enhanced CLI output** for liveliness
- **GitHub Actions CI**: cross-compile darwin/arm64, darwin/amd64, linux/amd64, windows/amd64
- **Release templates**: Homebrew formula, Scoop manifest, install script
- **Self-update** (`wade update`)
- **Shell completions** via cobra built-in
- **GitHub Pages** landing page (Apple dark theme)
- **CHANGELOG.md** (this file)

### Changed

- `--version` flag now works correctly (ldflags injection)
- Registry switching tolerates pnpm failures on older Node versions
- `wade status` output simplified with helpful tips
- README rewritten with cross-platform install guide (pre-built + brew + scoop + curl)

### Fixed

- Stale binary at `~/.local/bin` shadowing updated binary at `/usr/local/bin`
- pnpm v11 crash on Node v20 (now shows warning, doesn't block)
- `wade registry ls` table rendering for wide URLs

---

## [v0.1.0] — 2026-07-28

### Added

- **Node version management** (`wade node`)
  - `install`, `use`, `ls`, `ls-remote`, `default`, `uninstall`, `current`
  - Shim-based switching via `~/.wade/shims/`
  - Mirror download from npmmirror.com
  - Partial version resolution (e.g., `20` → `v20.20.2`)
- **Registry management** (`wade registry`)
  - `ls`, `use`, `add`, `del`, `test`
  - 5 built-in registries: npm, taobao, tencent, huawei, cnpm
  - Switches npm + yarn + pnpm simultaneously
- **Core infrastructure**
  - Go module with cobra CLI framework
  - TOML config at `~/.wade/config.toml`
  - `wade setup` for directory creation + PATH hint
- **SDD foundation**
  - `AGENTS.md` — AI master context
  - `spec/SPEC.md` — complete specification
  - `.hermes/plans/` — implementation plan

---

## Architecture Notes (for AI tools)

### Key design decisions

1. **Shim-based switching** — `~/.wade/shims/` contains symlinks to current Node binaries. Set PATH once, switching is instant.
2. **No Node.js dependency** — Wade is a compiled Go binary. Full `http`, `os/exec`, `archive/tar` in stdlib.
3. **Registry vs Mirror** — `wade registry use` controls npm/yarn/pnpm package registries. `wade node mirror` controls where Node.js binaries are downloaded from. These are completely separate.
4. **Per-PM error tolerance** — If pnpm fails (e.g., wrong Node version), npm and yarn still succeed. Config is saved if at least one PM succeeds.

### File structure

```
wade/
├── main.go                 # Entry point, ldflags version injection
├── cmd/                    # CLI commands (cobra)
│   ├── root.go             # Root command, -i shortcut, version flag
│   ├── init.go             # Interactive setup wizard
│   ├── node.go             # Node version management + mirror
│   ├── registry.go         # Registry management + table render
│   ├── status.go           # Status dashboard
│   ├── update.go           # Self-update from GitHub Releases
│   └── setup.go            # Directory creation + shell detection
├── internal/
│   ├── config/config.go    # TOML config load/save
│   ├── registry/           # Registry switching logic
│   │   ├── presets.go      # 5 built-in registries
│   │   ├── manager.go      # Use/Add/Remove + per-PM exec
│   │   └── tester.go       # Concurrent latency testing
│   └── node/
│       ├── versions.go     # Version parsing (semver) + remote index
│       ├── manager.go      # Download + tar.gz extraction
│       └── shim.go         # Symlink management for PATH
├── AGENTS.md               # AI master context
├── spec/SPEC.md            # Complete specification
├── CHANGELOG.md            # This file
├── README.md + README_zh.md  # User docs
├── docs/index.html         # GitHub Pages
├── .github/workflows/      # CI/CD
└── scripts/                # Homebrew, Scoop, install.sh
```
