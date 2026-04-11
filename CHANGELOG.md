# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [1.0.0] - 2026-04-11

### Added
- CI pipeline with GitHub Actions for build, test, and release
- Structured logging throughout the CLI
- `--json` output flag for machine-readable command output
- `golangci-lint` configuration and lint-passing codebase
- Install `engram` binary automatically in Docker and TUI installer
- TUI redesign with anvil banner, step indicators, progress bars, and spinners
- Lifecycle commands: `hefesto uninstall`, `hefesto update`, `hefesto rollback`
- `hefesto doctor` command for environment health checks
- `hefesto status --verbose` for detailed installation diagnostics
- Comprehensive test suite — 128+ unit tests with E2E test script
- README rewrite with updated documentation

### Fixed
- Connect CLI installer to real logic (was placeholder code)
- Resolve technical debt — consolidate duplicate code, fix flaky tests
- Dynamic `engram` version resolution (no hardcoded versions)
- TUI error screen crash on failed steps
- Doctor warnings producing false positives

## [0.1.0] - 2026-04-02

### Added
- Initial forge — Hefesto AI dev environment configuration bundle
- `opencode.json` config with model and tool settings
- `AGENTS.md` with orchestrator rules, SDD workflow, and agent conventions
- 15 framework skills (Angular, React, Next.js, Django, .NET, Go, etc.)
- Slim `engram` plugin — persistent memory with 3 core features
- Background-agents plugin — async delegation for parallel sub-agent execution
- TUI installer — Go/Bubbletea single-binary with embedded config
- Homebrew formula (`hefesto`) + CI/CD release pipeline
- Docker testing environment
- Initial Go unit tests

### Fixed
- Correct `opencode.json` format for compatibility with opencode v1.3.13
- Complete `engram` protocol documentation in `AGENTS.md` — was dangerously incomplete
- Add steps limit to all SDD sub-agents to prevent infinite loop freezes
- Use absolute paths for `AGENTS.md` skill references

[Unreleased]: https://github.com/Edcko/Hefesto/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/Edcko/Hefesto/compare/v0.1.0...v1.0.0
[0.1.0]: https://github.com/Edcko/Hefesto/releases/tag/v0.1.0
