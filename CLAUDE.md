# Claude Code Project Instructions

## Project Overview

vdash (vibe-dash) is a terminal dashboard for tracking AI-assisted coding projects. Built with Go, Bubble Tea TUI, and SQLite.

**Note:** The binary was renamed from `vibe` to `vdash` in v0.1.0. Old design docs (`docs/prd.md`, `docs/project-context.md`) still reference `vibe` as historical artifacts. See `docs/IMPROVEMENTS.md` for details.

## Release Workflow

When asked to create a release tag (e.g., "create release tag v0.2.0"):

1. **Check recent changes** - Review commits since last tag:
   ```bash
   git log $(git describe --tags --abbrev=0)..HEAD --oneline
   ```

2. **Ensure CI passes** - Run before tagging:
   ```bash
   make fmt && make lint && make test
   ```

3. **Create annotated tag** with release notes following this template:
   ```bash
   git tag -a vX.X.X -m "## vdash vX.X.X - [Brief Title]

   [One-line summary of the release]

   ### New Features
   - **Feature Name** - Description

   ### Improvements
   - Description of improvement

   ### Bug Fixes
   - Fixed: description

   ### Breaking Changes (if any)
   - Description of breaking change
   "
   ```

4. **Push main and tag**:
   ```bash
   git push origin main
   git push origin vX.X.X
   ```

5. **GitHub Actions** automatically builds and creates the release with the tag message.

### Version Guidelines

- `vX.Y.Z` - Stable release
- `vX.Y.Z-beta` - Beta/prerelease (auto-marked as prerelease)
- `vX.Y.Z-rc.1` - Release candidate

## Development Commands

```bash
make build      # Build binary to bin/vdash
make test       # Run unit tests
make test-all   # Run all tests including integration
make lint       # Run golangci-lint
make fmt        # Format code with goimports
make install    # Install to ~/go/bin/vdash
```

## Architecture

- `cmd/vdash/` - Application entry point
- `internal/core/` - Domain layer (ports, services, domain models)
- `internal/adapters/` - Infrastructure (CLI, TUI, persistence, detectors)

## Code Style

- Follow hexagonal architecture (ports & adapters)
- Keep domain layer free of external dependencies
- Use table-driven tests
- Run `make fmt && make lint` before committing
