# AGENTS.md — Wade Tool Master Context

> **Canonical source of truth for all AI agents working on this project.**
> Read this first before any change. Specs drive everything.

---

## 1. Project Identity

**Wade** — All-in-one Node.js version & npm registry manager.

- **Repo**: https://github.com/wadefengx/wade
- **Language**: Go 1.23+
- **Platforms**: macOS (Intel + ARM), Windows (x64), Linux (x64)
- **Author**: wadefengx (wadefengx@gmail.com)
- **License**: MIT

---

## 2. Problem Statement

### Problems solved

| Tool | Problem |
|------|---------|
| **nvm** | Each Node version has isolated global packages → switching versions loses CLI tools |
| **cgr/nrm** | Installed via `npm install -g` → tied to a specific Node version, breaks on switch |
| **Both** | Two separate tools, different ecosystems, installation conflicts |

### Wade's solution

1. **Single binary** — installed once via Homebrew/Scoop, never depends on Node
2. **Node version management** — install/switch/list/default via shim mechanism
3. **Registry management** — switch npm/yarn/pnpm registry with one command
4. **Cross-platform** — macOS, Windows, Linux from day one

---

## 3. SDD Workflow (Spec-Driven Development)

**This is the core development methodology.** Every feature follows this cycle:

```
┌─────────────────────────────────────────────────────┐
│                    SDD Cycle                         │
│                                                      │
│  SPEC.md (define behavior)                           │
│     │                                                 │
│     ▼                                                 │
│  Review (spec vs requirements)                        │
│     │                                                 │
│     ▼                                                 │
│  Implement (code to match spec) ▸ go test             │
│     │                                                 │
│     ▼                                                 │
│  Verify (test against spec)                           │
│     │                                                 │
│     ▼                                                 │
│  Commit + Release                                     │
│                                                      │
│  ♻ Back to Spec if requirements change               │
└─────────────────────────────────────────────────────┘
```

### Rules for AI agents

1. **Spec first** — Before writing any code for a feature, write or update SPEC.md
2. **Spec defines passing criteria** — A test passes if it matches the spec
3. **Refactoring** — Change implementation, keep spec; if spec needs to change, it's a feature change, not a refactor
4. **No code without spec** — Every function, every command, every behavior must be described in a spec
5. **Specs are living documents** — Update when behavior changes, not when implementation changes

---

## 4. Architecture Overview

```
wade (single Go binary)
├── cmd/                    # CLI entry points (cobra)
│   ├── root.go
│   ├── node.go             # wade node install/use/ls/default/uninstall
│   ├── registry.go         # wade registry ls/use/add/del/test
│   └── status.go           # wade status
├── internal/
│   ├── node/               # Node version management
│   │   ├── manager.go      # install/download/extract/switch
│   │   ├── shim.go         # PATH shim mechanism
│   │   └── versions.go     # version parsing, listing
│   ├── registry/           # Registry management
│   │   ├── manager.go      # switch/add/delete/list
│   │   ├── presets.go      # built-in registries
│   │   └── tester.go       # latency testing
│   ├── config/             # Config persistence (TOML)
│   │   └── config.go
│   └── platform/           # OS abstraction layer
│       ├── darwin.go
│       ├── windows.go
│       └── linux.go
├── spec/                   # Spec documents
│   ├── SPEC.md             # Master spec (entry point)
│   ├── node-spec.md        # Node version management spec
│   └── registry-spec.md    # Registry management spec
├── .github/
│   └── workflows/
│       └── release.yml     # Cross-compile + GitHub Releases
├── scripts/
│   └── install.sh
├── go.mod
├── go.sum
└── Makefile
```

---

## 5. Core Design Decisions

### 5.1 Shim-based version switching

Instead of modifying PATH on every `wade node use`, Wade uses a **shim directory** in `~/.wade/shims/`:

```
~/.wade/
├── shims/                  # ONE dir on PATH (setup once)
│   ├── node                # → symlink to current node binary
│   ├── npm                 # → symlink to current npm binary
│   ├── npx
│   ├── yarn
│   └── pnpm
├── versions/
│   ├── v18.20.0/
│   ├── v20.12.0/
│   └── v22.4.0/
├── config.toml
└── current                # Current version string
```

### 5.2 Registry switching

`wade registry use taobao` writes to ALL package manager configs simultaneously:
- `npm config set registry <url>`
- `yarn config set registry <url>` (if yarn installed)
- `pnpm config set registry <url>` (if pnpm installed)

### 5.3 No Node.js dependency

Wade itself is a compiled Go binary. It does NOT require Node.js to be installed.
`wade node install` downloads pre-compiled Node binaries from mirrors.

### 5.4 Mirror-first Node downloads

Default download mirror: `https://npmmirror.com/mirrors/node/` (China-friendly).
Configurable via `wade node mirror <url>` or `~/.wade/config.toml`.

---

## 6. CLI Reference

```bash
# Node version management
wade node install 18          # Install Node 18 (latest 18.x)
wade node install 20.12.0    # Install specific version
wade node use 18             # Switch to Node 18
wade node ls                 # List installed versions
wade node ls-remote          # List available versions from mirror
wade node default 20         # Set default version
wade node uninstall 18       # Remove a version
wade node current            # Print current version

# Registry management
wade registry ls             # List all registries, mark current
wade registry use taobao    # Switch ALL package managers to taobao
wade registry use npm       # Switch back to official
wade registry add myreg https://my.registry.com/
wade registry del myreg
wade registry test          # Test latency of all registries

# Status
wade status                 # Show: Node version + registry + shim health
wade setup                  # Add ~/.wade/shims to PATH (shell integration)
wade version                # Print wade version
```

---

## 7. Built-in Registries

| Name | URL |
|------|-----|
| npm | https://registry.npmjs.org/ |
| taobao | https://registry.npmmirror.com/ |
| tencent | https://mirrors.tencent.com/npm/ |
| huawei | https://repo.huaweicloud.com/repository/npm/ |
| cnpm | http://r.cnpmjs.org/ |

---

## 8. Tech Stack

| Layer | Library | Reason |
|-------|---------|--------|
| CLI framework | `github.com/spf13/cobra` | Industry standard, subcommands |
| Config | `github.com/pelletier/go-toml/v2` | TOML for config files |
| HTTP | `net/http` (stdlib) | No deps needed for downloads |
| Progress bar | `github.com/schollz/progressbar/v3` | Download UX |
| Table output | `github.com/olekukonko/tablewriter` | Nice CLI tables |
| Cross-compile | GOOS/GOARCH env vars | No CGO, pure Go |

---

## 9. Conventions

### Git
- Local: `user.name = wadefengx`, `user.email = wadefengx@gmail.com`
- Branch: `master` (stable), feature branches (dev)
- Commit: `type(scope): description` — `feat(node): add install command`

### Code
- `gofmt` formatting
- `internal/pkg` layout
- Errors: wrap with context at boundaries
- No panics in library code
- Table-driven tests

### File naming
- `snake_case.go`
- Build constraints: `//go:build darwin`

---

## 10. Pitfalls to Avoid (from nvm/cgr experience)

1. **NEVER** install wade via `npm install -g` — only via brew/scoop/binary
2. **NEVER** put wade binary inside any Node-managed directory
3. **NEVER** depend on Node being installed for wade to function
4. **ALWAYS** put `~/.wade/shims` at front of PATH
5. **ALWAYS** skip package managers that aren't installed (no error if yarn missing)
6. **NEVER** write code before writing spec

---

## 11. Milestone Roadmap

| Milestone | Content | Depends on |
|-----------|---------|------------|
| **M0: Skeleton** | Go module, CLI framework, AGENTS.md, SPEC.md | — |
| **M1: Registry** | `wade registry ls/use/add/del/test` | M0 |
| **M2: Node** | `wade node install/use/ls/default/uninstall` | M0 |
| **M3: Release** | GitHub Actions, Homebrew, Scoop, install script | M1+M2 |
| **M4: Polish** | `wade status`, shell completions, self-update | M3 |