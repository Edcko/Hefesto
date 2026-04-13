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

[![Go Version](https://img.shields.io/badge/Go-1.26.1-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/Edcko/Hefesto?include_prereleases)](https://github.com/Edcko/Hefesto/releases)

---

## What is Hefesto?

Hefesto is an opinionated installer and configuration manager for OpenCode. It deploys a complete AI development environment with agents, skills, SDD workflow, persistent memory, and more — in one command.

No complex setup. No manual file copying. Just install and start building with AI assistance that actually understands your workflow.

---

## Features

- **10 CLI Commands**: `install`, `status`, `update`, `uninstall`, `rollback`, `doctor`, `config`, `list`, `version`, `completion`
- **26 AI Skills**: Angular, React, Next.js, TypeScript, Tailwind, Zod, Django, .NET, Playwright, Pytest, and more
- **6-Phase SDD Workflow**: `init → plan → spec → tasks → apply → verify`
- **10 Agents**: Primary mentor, orchestrator, SDD phases, and remote execution
- **Persistent Memory**: Via Engram integration for cross-session context
- **Background Agents Plugin**: Parallel task execution without blocking
- **Interactive TUI Installer**: Progress tracking with Bubbletea
- **Homebrew Distribution**: `brew install edcko/tap/hefesto`
- **Multi-Platform**: darwin/linux × arm64/amd64 + android-arm64 binaries
- **Fuego/Forge Theme**: Cohesive visual identity with amber/copper palette

---

## Quick Start

```bash
# Install via Homebrew (macOS/Linux)
brew install edcko/tap/hefesto

# Or download binary from GitHub Releases
# https://github.com/Edcko/Hefesto/releases

# Android/Termux users: download the android-arm64 binary directly
# from GitHub Releases

# Install Hefesto configuration
hefesto install

# Check installation health
hefesto doctor

# Start using OpenCode
opencode
```

---

## CLI Commands

| Command | Description | Flags |
|--------|-------------|-------|
| `hefesto install` | Install Hefesto configuration files | `--yes`, `--dry-run` |
| `hefesto status` | Show installation status | `--verbose` |
| `hefesto doctor` | Run comprehensive health diagnostics | — |
| `hefesto update` | Update to latest configuration (not the binary) | `--yes`, `--dry-run` |
| `hefesto uninstall` | Remove Hefesto configuration | `--yes`, `--purge` |
| `hefesto rollback` | Restore a previous backup | `--yes`, `--list` |
| `hefesto config show` | Display current config paths and settings | — |
| `hefesto config path` | Print config directory path | — |
| `hefesto list skills` | List all embedded skills | `--json` |
| `hefesto list themes` | List available themes | `--json` |
| `hefesto list backups` | List timestamped backups | `--json` |
| `hefesto version` | Print version information | — |

### Command Details

**`hefesto install`**
- Detects OpenCode configuration directory (`~/.config/opencode/`)
- Creates timestamped backups of existing configs
- Deploys embedded configuration files
- Sets up skills, themes, plugins, and commands
- Flags: `--yes` (non-interactive), `--dry-run` (preview changes)

**`hefesto update`**
- Creates timestamped backup of current configuration
- Overlays latest embedded config files (preserves customizations where possible)
- **Important**: This only updates configuration files, NOT the Hefesto binary itself
- To update the binary: `brew upgrade hefesto` or download from [GitHub Releases](https://github.com/Edcko/Hefesto/releases)
- Flags: `--yes` (skip confirmation), `--dry-run` (preview changes)

**`hefesto doctor`**
Runs comprehensive checks on:
- Configuration directory structure
- AGENTS.md file validity
- opencode.json configuration
- Skills directory and structure
- Plugins directory (engram, background-agents)
- Theme configuration
- Personality settings
- Custom commands

**`hefesto rollback`**
- Lists available backups with timestamps
- Creates safety backup before restoring
- Flags: `--list` (show backups), `--yes` (restore most recent without prompt)

**`hefesto status`**
- Shows installation directory
- Displays installed version
- Lists available skills count
- Reports configuration health
- Flag: `--verbose` for detailed output

---

## Shell Completion

Hefesto supports shell completion for bash, zsh, fish, and PowerShell:

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
├── AGENTS.md              # Hefesto persona + SDD orchestrator + Engram protocol
├── opencode.json          # 10 agent definitions with step limits
├── skills/                # 26 skill directories
│   ├── _shared/           # Phase-common patterns, persistence conventions
│   ├── ai-sdk-5/          # Vercel AI SDK 5 patterns
│   ├── angular/           # Angular 20+ architecture
│   ├── django-drf/        # Django REST Framework
│   ├── dotnet/            # .NET 9 / ASP.NET Core
│   ├── go-testing/        # Go + Bubbletea TUI testing
│   ├── nextjs-15/         # Next.js 15 App Router
│   ├── playwright/        # E2E testing with Playwright
│   ├── pr-review/         # GitHub PR review workflow
│   ├── pytest/            # Python testing patterns
│   ├── react-19/          # React 19 with Compiler
│   ├── remote-exec/       # SSH/VPS remote execution
│   ├── sdd-*/             # SDD phase skills (init, plan, spec, tasks, apply, verify)
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

## Skills (26 Total)

### SDD Workflow
- **sdd-init** — Bootstrap SDD context and project configuration
- **sdd-plan** — Explore codebase and create change proposals (merged explore + propose)
- **sdd-spec** — Write detailed specifications from proposals
- **sdd-tasks** — Break down specs and designs into implementation tasks
- **sdd-apply** — Implement code changes from task definitions
- **sdd-verify** — Validate implementation against specs

### Frontend Frameworks
- **angular** — Angular 20+ architecture with Scope Rule, Screaming Architecture, standalone components, signals
- **react-19** — React 19 patterns with React Compiler (no useMemo/useCallback needed)
- **nextjs-15** — Next.js 15 App Router patterns (routing, Server Actions, data fetching)
- **tailwind-4** — Tailwind CSS 4 patterns and best practices (cn(), theme variables)
- **typescript** — TypeScript strict patterns and best practices (types, interfaces, generics)
- **zustand-5** — Zustand 5 state management patterns

### Backend Frameworks
- **django-drf** — Django REST Framework patterns (ViewSets, Serializers, Filters)
- **dotnet** — .NET 9 / ASP.NET Core with Minimal APIs, Clean Architecture, EF Core

### Testing
- **playwright** — Playwright E2E testing patterns (Page Objects, selectors, MCP workflow)
- **pytest** — Pytest testing patterns for Python (fixtures, mocking, markers)
- **go-testing** — Go tests and Bubbletea TUI testing

### AI & SDK
- **ai-sdk-5** — Vercel AI SDK 5 patterns (breaking changes from v4)
- **zod-4** — Zod 4 schema validation patterns (breaking changes from v3)

### Workflow & Review
- **pr-review** — Review GitHub PRs and Issues with structured analysis
- **technical-review** — Review technical exercises and candidate submissions
- **skill-creator** — Create new AI agent skills following Agent Skills spec
- **skill-registry** — Create or update project skill registry
- **stream-deck** — Create slide-deck presentation webs for streams and courses

### DevOps
- **remote-exec** — Execute commands on remote servers via SSH

### Shared
- **_shared** — Phase-common patterns, persistence conventions

---

## SDD Workflow

Spec-Driven Development is the structured planning layer for substantial changes.

### 6 Phases

```
init → plan → spec → tasks → apply → verify
```

- **init** — Bootstrap SDD context and detect project stack
- **plan** — Explore codebase + create change proposal (merged explore + propose)
- **spec** — Write detailed requirements and scenarios
- **tasks** — Break down specs into implementation checklist
- **apply** — Implement code changes from task definitions
- **verify** — Validate implementation matches specs

### Slash Commands

| Command | Action |
|---------|--------|
| `/sdd-init` | Bootstrap SDD in your project |
| `/sdd-new <change>` | Create new change (runs plan phase) |
| `/sdd-ff <change>` | Fast-forward: plan → spec → tasks |
| `/sdd-apply <change>` | Implement tasks |
| `/sdd-verify <change>` | Validate implementation |

### Persistence

**Engram only.** No mode selection. Artifacts are stored in persistent memory and survive sessions and compactions.

---

## Architecture

### Installer
- **Language**: Go 1.26.1
- **TUI Framework**: Bubbletea (Charmbracelet)
- **CLI Framework**: Cobra
- **Embedding**: Configuration files embedded in binary via `go:embed`

### Distribution
- **Homebrew Tap**: `edcko/tap/hefesto`
- **GitHub Releases**: 5 platform binaries (darwin/linux × arm64/amd64 + android-arm64)
- **Installation Size**: ~15MB (includes all configs, skills, themes)

### Configuration
- **Target Directory**: `~/.config/opencode/`
- **Backup Strategy**: Timestamped backups before each operation
- **Rollback Support**: Full backup restoration with safety backup

---

## Development

```bash
# Clone the repository
git clone https://github.com/Edcko/Hefesto.git
cd Hefesto

# Build the binary
cd cmd/hefesto
go build .

# Run locally
./hefesto install --dry-run

# Test in Docker (multi-platform)
cd ../..
./scripts/test.sh

# Run unit tests
cd cmd/hefesto
go test ./...
```

### Project Structure

```
Hefesto/
├── README.md
├── README.es.md
├── LICENSE
├── .gitignore
├── HefestoOpenCode/           # Configuration to be deployed
│   ├── AGENTS.md
│   ├── opencode.json
│   ├── skills/
│   ├── plugins/
│   ├── commands/
│   ├── themes/
│   └── personality/
├── cmd/hefesto/               # Installer binary
│   ├── main.go                # CLI entry point
│   ├── internal/
│   │   ├── install/           # Installation logic
│   │   ├── tui/               # Bubbletea TUI
│   │   └── embed/config/      # Embedded configs (via go:embed)
│   └── go.mod
└── scripts/
    └── test.sh                # Docker test runner
```

---

## Techne Ecosystem

Hefesto is part of the Techne ecosystem:

| Project | Role |
|---------|------|
| **Techne** | The foundation/matrix of the ecosystem |
| **Hefesto** | The forge — configures and deploys AI dev environments |
| **Engram** | Persistent memory for AI agents (integrated) |
| **OpenCode** | The AI coding assistant platform |

---

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-improvement`
3. Make your changes following conventional commits
4. Commit: `git commit -m "feat: clear description"`
5. Push: `git push origin feature/my-improvement`
6. Open a Pull Request

**Guidelines:**
- Use conventional commits (`feat:`, `fix:`, `docs:`, `refactor:`)
- NO AI attribution in commits (no "Co-Authored-By" lines)
- Clean code over quick code
- Test your changes with `hefesto install --dry-run`

---

## License

MIT License © 2026 Edcko

See [LICENSE](LICENSE) for full text.

---

> 🔥 *"A good blacksmith doesn't blame their hammer. They forge with what they have."*
