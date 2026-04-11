# Contributing to Hefesto

Thanks for contributing! Here's how to get started.

## Prerequisites

- Go 1.26+
- golangci-lint (for linting)

## Build

```bash
cd cmd/hefesto
go build .
```

Or from the project root:

```bash
make build
```

## Run Tests

```bash
cd cmd/hefesto
go test ./... -count=1
```

Or from the project root:

```bash
make test
```

## Lint

```bash
cd cmd/hefesto
golangci-lint run ./...
```

Or from the project root:

```bash
make lint
```

## Commit Conventions

We use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` — New feature
- `fix:` — Bug fix
- `docs:` — Documentation only
- `refactor:` — Code refactor (no feature/fix)
- `test:` — Adding or updating tests
- `chore:` — Build, CI, tooling changes

**No AI attribution** — Do not add "Co-Authored-By" or similar lines.

## Pull Request Process

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-improvement`
3. Make your changes
4. Ensure tests pass: `go test ./... -count=1`
5. Ensure lint passes: `golangci-lint run ./...`
6. Commit with conventional commit format
7. Push and open a Pull Request

## Code Style

- Follow standard Go conventions ([Effective Go](https://go.dev/doc/effective_go))
- Run `go fmt` before committing
- Follow existing patterns in the codebase
- Keep functions focused and small

## Project Structure

```
cmd/hefesto/
├── main.go                    # CLI entry point (Cobra commands)
├── internal/
│   ├── install/               # Core logic (install, update, rollback, backup)
│   ├── tui/                   # Bubbletea TUI components
│   └── embed/config/          # Embedded config files (go:embed)
└── go.mod
```

## Adding New Commands

Commands are defined in `main.go` using [Cobra](https://github.com/spf13/cobra). Follow the existing pattern:

1. Add a new command function in `main.go`
2. Implement logic in `internal/install/` as a new file or extend existing
3. Add TUI support in `internal/tui/` if the command needs interactivity
4. Add tests in the same package (`*_test.go`)

## Adding New Skills

Skills live in `HefestoOpenCode/skills/`. Each skill is a directory with:

```
skills/my-skill/
└── SKILL.md    # Markdown with instructions, patterns, and rules
```

Skills are embedded at build time via `go:embed`.
