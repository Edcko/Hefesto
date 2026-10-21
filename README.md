[🇪🇸 Español](README.es.md)

```
    🔥
   ╱│╲
  ╱ │ ╲
 ╱  │  ╲        HEFESTO
╱___▼___╲       AI Dev Environment Forge
 ║███████║
 ║███████║       Forge your perfect AI dev environment
 ╰═══════╯
```

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/Edcko/Hefesto?include_prereleases)](https://github.com/Edcko/Hefesto/releases)
[![Platforms](https://img.shields.io/badge/platforms-macOS%20%7C%20Linux%20%7C%20Android-lightgrey)](https://github.com/Edcko/Hefesto/releases)

---

## What is Hefesto?

Hefesto is an opinionated installer and configuration manager for [OpenCode](https://github.com/opencode-ai/opencode). It deploys a complete AI-powered development environment — agents, skills, SDD workflow, persistent memory, theme, and personality — in a single command.

No complex setup. No manual file copying. Just install and start building with AI assistance that actually understands your workflow.

```bash
brew install edcko/tap/hefesto
hefesto install
opencode
```

That's it. You're running with 14 agents, 32 skills, persistent memory, a full Spec-Driven Development workflow, and a router that computes the methodology profile for each request.

---

## Features

- **12 CLI Commands** — `install`, `status`, `update`, `uninstall`, `rollback`, `doctor`, `config`, `list`, `version`, `completion`, `route`, `verify-plan`
- **32 AI Skills** — Angular, React, Next.js, TypeScript, Tailwind, Zod, Django, .NET, Playwright, Pytest, and more
- **Methodology Router** — `hefesto route` computes the profile that fits a request; `hefesto verify-plan` checks the run against that profile's contract
- **10 SDD Phase Skills** — `init → explore → propose → spec → design → tasks → apply → verify → archive`
- **14 Agents** — Primary mentor, orchestrator, SDD phase agents, remote execution
- **Persistent Memory** — Engram integration for cross-session context that survives compactions
- **Background Agents Plugin** — Parallel task execution without blocking the main thread
- **Interactive TUI Installer** — Progress tracking with Bubbletea, anvil banner, spinners
- **Homebrew + curl|bash** — `brew install edcko/tap/hefesto` or `curl -fsSL ... | bash`
- **5 Platform Binaries** — macOS (ARM64/Intel), Linux (ARM64/AMD64), Android ARM64
- **Fuego/Forge Theme** — Cohesive amber/copper visual identity

---

## Installation

### Homebrew (recommended)

```bash
brew install edcko/tap/hefesto
```

### curl|bash

```bash
curl -fsSL https://raw.githubusercontent.com/Edcko/Hefesto/main/install.sh | bash
```

### Binary Download

Download the latest release for your platform from [GitHub Releases](https://github.com/Edcko/Hefesto/releases):

| Platform | Architecture | Binary |
|----------|-------------|--------|
| macOS | ARM64 (Apple Silicon) | `hefesto-darwin-arm64` |
| macOS | AMD64 (Intel) | `hefesto-darwin-amd64` |
| Linux | ARM64 | `hefesto-linux-arm64` |
| Linux | AMD64 | `hefesto-linux-amd64` |
| Android/Termux | ARM64 | `hefesto-android-arm64` |

### Post-Install

```bash
hefesto install          # Deploy configuration to ~/.config/opencode/
hefesto doctor           # Verify everything is healthy
opencode                 # Start coding
```

---

## CLI Reference

| Command | Description | Flags |
|--------|-------------|-------|
| `hefesto install` | Install configuration files | `--yes`, `--dry-run`, `--test` |
| `hefesto status` | Show installation status | `--verbose`, `--json` |
| `hefesto doctor` | Run comprehensive health diagnostics | `--json` |
| `hefesto update` | Update configuration (not the binary) | `--yes`, `--dry-run` |
| `hefesto uninstall` | Remove configuration files | `--yes`, `--purge` |
| `hefesto rollback` | Restore a previous backup | `--yes`, `--list` |
| `hefesto config show` | Display config paths and settings | — |
| `hefesto config path` | Print config directory path | — |
| `hefesto list skills` | List all embedded skills | `--json` |
| `hefesto list themes` | List available themes | `--json` |
| `hefesto list backups` | List timestamped backups | `--json` |
| `hefesto version` | Print version information | — |
| `hefesto route` | Decide the methodology profile for a request | `--type`, `--size`, `--risk`, `--change`, `--json` |
| `hefesto verify-plan` | Validate a plan against the contract of its profile | `--json` |

### Command Details

**`hefesto install`**
- Detects OpenCode configuration directory (`~/.config/opencode/`)
- Creates timestamped backups of existing configs
- Deploys embedded configuration files (skills, plugins, commands, theme, personality)
- Optionally installs OpenCode CLI and Engram binary if missing
- Flags: `--yes` (non-interactive), `--dry-run` (preview), `--test` (temp directory)

**`hefesto update`**
- Creates timestamped backup of current configuration
- Overlays latest embedded config files (preserves customizations where possible)
- **Note**: This updates configuration files only. To update the binary: `brew upgrade hefesto`

**`hefesto doctor`**
Runs comprehensive checks on:
- Configuration directory structure, AGENTS.md, opencode.json
- Skills, plugins (engram, background-agents), theme, personality, commands
- OpenCode binary, Engram binary

**`hefesto rollback`**
- Lists available backups with timestamps
- Creates safety backup before restoring
- Flags: `--list` (show backups), `--yes` (restore most recent)

**`hefesto route`**
- Decides the methodology profile for a request from three facts: the dominant
  uncertainty (`--type`), the size of the work (`--size`) and the high-risk item
  the change touches (`--risk`)
- Prints the decision, the contract of the chosen profile, the markdown scaffold
  and the plan the run starts from
- The profile is computed, never chosen by taste: the same inputs always
  produce the same decision
- Flags: `--change` (name of the change), `--json` (machine-readable payload)

**`hefesto verify-plan`**
- Validates a plan document against the contract of the profile it was routed to
- Reports every rule the plan breaks: missing artifacts, broken ordering,
  scenarios without verification, a missing independent verifier, artifacts
  outside the contract
- Exits non-zero while the plan still owes work
- Flags: `--json` (machine-readable report)

---

## Shell Completion

```bash
# Bash
hefesto completion bash > ~/.config/hefesto/completion.bash
source ~/.config/hefesto/completion.bash

# Zsh
hefesto completion zsh > "${fpath[1]}/_hefesto"

# Fish
hefesto completion fish > ~/.config/fish/completions/hefesto.fish
```

---

## What Gets Installed

Hefesto deploys the following into `~/.config/opencode/`:

```
~/.config/opencode/
├── AGENTS.md              # Persona + SDD orchestrator + Engram protocol
├── opencode.json          # 14 agent definitions with step limits
├── skills/                # 31 skill directories (30 + _shared)
│   ├── _shared/           # Phase-common patterns, persistence conventions
│   ├── ai-sdk-5/          # Vercel AI SDK 5 patterns
│   ├── angular/           # Angular 20+ Scope Rule architecture
│   ├── django-drf/        # Django REST Framework
│   ├── dotnet/            # .NET 9 / ASP.NET Core
│   ├── homebrew-release/  # Homebrew tap release workflow
│   ├── jira-epic/         # Jira epic creation
│   ├── jira-task/         # Jira task creation
│   ├── nextjs-15/         # Next.js 15 App Router
│   ├── playwright/        # E2E testing with Playwright
│   ├── pr-review/         # GitHub PR review workflow
│   ├── pytest/            # Python testing patterns
│   ├── react-19/          # React 19 with Compiler
│   ├── remote-exec/       # SSH/VPS remote execution
│   ├── sdd-*/             # 10 SDD phase skills
│   ├── skill-creator/     # AI agent skill creation
│   ├── skill-registry/    # Project skill registry management
│   ├── stream-deck/       # Presentation slide decks
│   ├── tailwind-4/        # Tailwind CSS 4 patterns
│   ├── technical-review/  # Code assessment workflow
│   ├── typescript/        # TypeScript strict patterns
│   ├── zod-4/             # Zod 4 schema validation
│   └── zustand-5/         # Zustand 5 state management
├── plugins/
│   ├── engram.ts          # Persistent memory integration
│   └── background-agents.ts  # Parallel agent execution
├── commands/              # 5 SDD slash commands
│   ├── sdd-init.md
│   ├── sdd-new.md
│   ├── sdd-ff.md
│   ├── sdd-apply.md
│   └── sdd-verify.md
├── themes/
│   └── hefesto.json       # Fuego/Forge theme (amber/copper)
└── personality/
    └── hefesto.md         # Hefesto persona definition
```

---

## Skills (31 Total)

### SDD Workflow (10 skills)
- **sdd-init** — Bootstrap SDD context and detect project stack
- **sdd-explore** — Investigate codebase before committing to a change
- **sdd-propose** — Create change proposal with intent and scope
- **sdd-design** — Create technical design with architecture decisions
- **sdd-plan** — Explore + propose (merged convenience phase)
- **sdd-spec** — Write detailed requirements and scenarios
- **sdd-tasks** — Break down specs into implementation checklist
- **sdd-apply** — Implement code changes from task definitions
- **sdd-verify** — Validate implementation against specs
- **sdd-archive** — Sync delta specs and archive completed change

### Frontend Frameworks
- **angular** — Angular 20+ with Scope Rule, Screaming Architecture, standalone components, signals
- **react-19** — React 19 patterns with React Compiler
- **nextjs-15** — Next.js 15 App Router (routing, Server Actions, data fetching)
- **tailwind-4** — Tailwind CSS 4 patterns (cn(), theme variables)
- **typescript** — TypeScript strict patterns (types, interfaces, generics)
- **zustand-5** — Zustand 5 state management

### Backend Frameworks
- **django-drf** — Django REST Framework (ViewSets, Serializers, Filters)
- **dotnet** — .NET 9 / ASP.NET Core Minimal APIs, Clean Architecture, EF Core

### Testing
- **playwright** — E2E testing with Page Objects, selectors, MCP workflow
- **pytest** — Python testing patterns (fixtures, mocking, markers)

### AI & Validation
- **ai-sdk-5** — Vercel AI SDK 5 patterns (breaking changes from v4)
- **zod-4** — Zod 4 schema validation (breaking changes from v3)

### Workflow & Review
- **pr-review** — GitHub PR/Issue review with structured analysis
- **technical-review** — Code assessment and candidate submission review
- **skill-creator** — Create new AI agent skills
- **skill-registry** — Project skill registry management
- **stream-deck** — Presentation slide decks for streams and courses

### DevOps & Project Management
- **remote-exec** — Execute commands on remote servers via SSH
- **homebrew-release** — Homebrew tap release workflow
- **jira-epic** — Jira epic creation
- **jira-task** — Jira task creation

### Shared
- **_shared** — Phase-common patterns, persistence conventions

---

## SDD Workflow

Spec-Driven Development is the structured planning layer for substantial code changes. It enforces a **design-before-code** discipline through discrete phases.

### Dependency Graph

```
init → explore → propose → spec ──→ tasks → apply → verify → archive
                                  ╲                    ↑
                                   → design ──────────╯
```

Each phase reads from previous artifacts and writes its own. Phases are designed to be run by specialized sub-agents.

### Slash Commands

| Command | Action |
|---------|--------|
| `/sdd-init` | Bootstrap SDD in your project |
| `/sdd-new <change>` | Create new change (runs explore → propose) |
| `/sdd-ff <change>` | Fast-forward: propose → spec → tasks |
| `/sdd-apply <change>` | Implement tasks in batches |
| `/sdd-verify <change>` | Validate implementation against specs |

### Persistence

All SDD artifacts are stored in **Engram** persistent memory. They survive session boundaries and context compactions — no lost work.

---

## Architecture

| Layer | Technology |
|-------|-----------|
| Language | Go 1.26 |
| TUI Framework | Bubbletea (Charmbracelet) |
| CLI Framework | Cobra |
| Embedding | `go:embed` — all configs baked into the binary |
| Distribution | Homebrew tap + GitHub Releases (5 platforms) |
| Binary Size | ~15 MB (includes all configs, skills, themes) |

### Design Decisions

- **Single binary**: Everything embedded via `go:embed`. No external dependencies at runtime.
- **Timestamped backups**: Every install/update creates a backup. `hefesto rollback` restores any previous state.
- **Engram-first**: Persistent memory is the primary artifact store. No file-based mode selection.
- **Non-interactive mode**: All commands support `--yes` for CI/CD and scripting.

---

## Development

### Prerequisites

- Go 1.26+
- Make (optional, for build automation)

### Build from Source

```bash
git clone https://github.com/Edcko/Hefesto.git
cd Hefesto

# Build for current platform
cd cmd/hefesto && go build .

# Or use the Makefile for all platforms
make build

# Run locally
./hefesto install --dry-run
```

### Testing

```bash
# Unit tests
cd cmd/hefesto && go test ./...

# Multi-platform Docker tests
./scripts/test.sh

# Lint
make lint
```

### Project Structure

```
Hefesto/
├── HefestoOpenCode/           # Source configuration (deployed to ~/.config/opencode/)
│   ├── AGENTS.md              # Agent orchestrator rules
│   ├── opencode.json          # Agent definitions
│   ├── skills/                # 32 skill directories
│   ├── plugins/               # Engram + background-agents
│   ├── commands/              # SDD slash commands and /route
│   ├── themes/                # Fuego/Forge theme
│   └── personality/           # Hefesto persona
├── cmd/hefesto/               # Go installer binary
│   ├── main.go                # CLI entry point (Cobra commands)
│   ├── route.go               # route + verify-plan commands
│   ├── router/                # Methodology router: decision table, contracts, plan validator
│   └── internal/
│       ├── install/           # Core install/backup/rollback/doctor logic
│       ├── tui/               # Bubbletea TUI screens
│       └── embed/             # go:embed config files (mirrors HefestoOpenCode/)
├── Formula/hefesto.rb         # Homebrew formula
├── install.sh                 # curl|bash installer
├── scripts/                   # Build, test, E2E scripts
├── Makefile                   # Build automation
└── .github/workflows/         # CI: test + release
```

---

## Techne Ecosystem

Hefesto is part of the Techne ecosystem — a set of tools designed to work together for AI-assisted development:

| Project | Role |
|---------|------|
| **Techne** | The foundation — philosophy and conventions |
| **Hefesto** | The forge — deploys and configures AI dev environments |
| **Engram** | The memory — persistent cross-session context for AI agents |
| **OpenCode** | The platform — AI coding assistant |

---

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-improvement`
3. Make your changes
4. Commit with conventional commits: `git commit -m "feat: clear description"`
5. Push: `git push origin feature/my-improvement`
6. Open a Pull Request

**Guidelines:**
- Use conventional commits (`feat:`, `fix:`, `docs:`, `refactor:`, `chore:`)
- No AI attribution in commits
- Test your changes with `hefesto install --dry-run`
- Keep embedded configs in sync with `HefestoOpenCode/` source

---

## License

[MIT](LICENSE) © 2026 Edcko
