# Story 1.1: Project Scaffolding

**Status:** Done

## Story

**As a** developer,
**I want** the project structure initialized with all directories and configuration files,
**So that** I have a solid foundation to build upon.

## Acceptance Criteria

```gherkin
Given I am setting up the vibe-dash project
When I run the initialization commands
Then the following structure is created:
  - cmd/vibe/main.go (entry point)
  - internal/core/domain/ (entities)
  - internal/core/ports/ (interfaces)
  - internal/core/services/ (use cases)
  - internal/adapters/cli/ (Cobra commands)
  - internal/adapters/tui/ (Bubble Tea)
  - internal/adapters/persistence/ (SQLite + YAML)
  - internal/adapters/filesystem/ (OS abstraction)
  - internal/adapters/detectors/ (MethodDetector plugins)
  - internal/config/ (Viper integration)
  - test/fixtures/ (golden path test projects)
And go.mod is initialized with github.com/JeiKeiLim/vibe-dash
And Makefile contains: build, test, lint, fmt, run, clean targets
And .golangci.yml is configured for linting
And .gitignore excludes bin/, *.db, and IDE files
```

## Tasks / Subtasks

- [x] **Task 1: Create directory structure** (AC: structure creation)
  - [x] 1.1 Create cmd/vibe/ directory
  - [x] 1.2 Create internal/core/{domain,ports,services}/ directories
  - [x] 1.3 Create internal/adapters/{cli,tui,tui/components,persistence/sqlite,persistence/sqlite/migrations,persistence/yaml,filesystem,detectors,detectors/speckit}/ directories
  - [x] 1.4 Create internal/config/ directory
  - [x] 1.5 Create test/fixtures/ directory with ALL golden path subdirectories:
    - test/fixtures/speckit-stage-specify/
    - test/fixtures/speckit-stage-plan/
    - test/fixtures/speckit-stage-tasks/
    - test/fixtures/speckit-stage-implement/ (CRITICAL: includes implement.md stage)
    - test/fixtures/speckit-uncertain/
    - test/fixtures/no-method-detected/
    - test/fixtures/empty-project/
  - [x] 1.6 Create pkg/ directory (optional, for shareable utilities)
  - [x] 1.7 Create .github/workflows/ directory for CI pipeline

- [x] **Task 2: Initialize Go module** (AC: go.mod)
  - [x] 2.1 Run `go mod init github.com/JeiKeiLim/vibe-dash`
  - [x] 2.2 Ensure go.mod contains `go 1.21` directive (required for slog stdlib)
  - [x] 2.3 Add initial dependencies to go.mod:
    - github.com/charmbracelet/bubbletea (TUI framework)
    - github.com/charmbracelet/bubbles (TUI components)
    - github.com/charmbracelet/lipgloss (styling)
    - github.com/spf13/cobra (CLI framework)
    - github.com/spf13/viper (configuration)
    - github.com/jmoiron/sqlx (SQL extensions)
    - github.com/mattn/go-sqlite3 (SQLite driver - REQUIRES CGO_ENABLED=1)
    - github.com/fsnotify/fsnotify (file watching)
  - [x] 2.4 Run `go mod tidy` to download dependencies

- [x] **Task 3: Create main.go entry point** (AC: cmd/vibe/main.go)
  - [x] 3.1 Create minimal main.go with placeholder structure
  - [x] 3.2 Import root command from internal/adapters/cli
  - [x] 3.3 Implement graceful shutdown signal handling pattern
  - [x] 3.4 Ensure binary name will be `vibe` when compiled

- [x] **Task 4: Create Makefile** (AC: Makefile targets)
  - [x] 4.1 Create `build` target: `CGO_ENABLED=1 go build -o bin/vibe ./cmd/vibe`
  - [x] 4.2 Create `test` target: `go test ./...`
  - [x] 4.3 Create `test-all` target: `go test -tags=integration ./...`
  - [x] 4.4 Create `lint` target: `golangci-lint run`
  - [x] 4.5 Create `fmt` target: `goimports -w .`
  - [x] 4.6 Create `check-fmt` target for CI validation
  - [x] 4.7 Create `run` target: build and execute
  - [x] 4.8 Create `clean` target: remove bin/ directory
  - [x] 4.9 Create `install` target: `go install ./cmd/vibe`
  - [x] 4.10 Create `test-accuracy` target placeholder for detection accuracy testing

- [x] **Task 5: Create .golangci.yml** (AC: linting configuration)
  - [x] 5.1 Configure recommended linters (govet, errcheck, staticcheck, unused, gosimple)
  - [x] 5.2 Set appropriate timeouts and severity levels
  - [x] 5.3 Exclude generated files and test fixtures

- [x] **Task 6: Create .gitignore** (AC: .gitignore)
  - [x] 6.1 Exclude bin/ directory
  - [x] 6.2 Exclude *.db files (SQLite databases)
  - [x] 6.3 Exclude IDE files (.idea/, .vscode/settings.json, *.swp)
  - [x] 6.4 Exclude OS files (.DS_Store, Thumbs.db)
  - [x] 6.5 Exclude coverage files (*.out, coverage.html)

- [x] **Task 7: Create placeholder and config files** (supports future stories)
  - [x] 7.1 Create .keep files in empty directories to ensure git tracking
  - [x] 7.2 Create README.md with project overview and build instructions
  - [x] 7.3 Create LICENSE file (MIT license recommended)
  - [x] 7.4 Create .github/workflows/ci.yml with lint, test, build jobs

- [x] **Task 8: Create CI Pipeline** (AC: automated validation)
  - [x] 8.1 Create .github/workflows/ci.yml with:
    - Checkout action
    - Setup Go 1.21+
    - Run `make check-fmt`
    - Run `make lint`
    - Run `make test-all`
    - Run `make build`
  - [x] 8.2 Configure CI to run on push and pull requests

- [x] **Task 9: Verify build** (validation)
  - [x] 9.1 Run `make build` successfully
  - [x] 9.2 Run `make lint` with no errors
  - [x] 9.3 Run `make test` (should pass with no tests yet)
  - [x] 9.4 Verify binary exists at bin/vibe
  - [x] 9.5 Run `./bin/vibe` and verify it starts without error

## Dev Notes

### Architecture Compliance

This story establishes the **hexagonal architecture** foundation specified in [Source: docs/architecture.md#Selected-Approach-Custom-Hexagonal-Structure]:

```
vibe-dash/
├── cmd/vibe/              # Entry point - wires up adapters
├── internal/
│   ├── core/              # Domain layer - ZERO external dependencies
│   │   ├── domain/        # Entities
│   │   ├── ports/         # Interfaces only
│   │   └── services/      # Use cases
│   └── adapters/          # Infrastructure layer
│       ├── cli/           # Cobra commands
│       ├── tui/           # Bubble Tea components
│       ├── persistence/   # SQLite + YAML
│       ├── filesystem/    # OS abstraction
│       └── detectors/     # MethodDetector implementations
├── internal/config/       # Viper integration
├── test/fixtures/         # Golden path test projects
└── pkg/                   # Shareable utilities (if any)
```

### Critical Boundary Rules

**From [Source: docs/architecture.md#Architectural-Boundaries]:**

| Boundary | Rule |
|----------|------|
| `internal/core/` → external | **FORBIDDEN** - core imports nothing from adapters |
| `internal/core/domain/` → services | **FORBIDDEN** - entities don't know about services |
| `internal/adapters/` → core | ALLOWED - adapters implement port interfaces |
| `internal/adapters/` → external libs | ALLOWED - Bubble Tea, sqlx, fsnotify, Viper |
| `cmd/` → everything | ALLOWED - wires up dependencies |

### Centralized Storage Architecture (CRITICAL)

**From [Source: docs/prd.md#Technical-Success]:**

Application uses **centralized storage** at `~/.vibe-dash/`:
- Master config: `~/.vibe-dash/config.yaml`
- Per-project state: `~/.vibe-dash/<project>/state.db`

**DO NOT create per-project `.vibe/` directories inside tracked projects.**

Rationale: Avoids forcing developers to add `.vibe/` to every project's `.gitignore`.

### DO NOT (Anti-Patterns)

| DO NOT | DO INSTEAD |
|--------|------------|
| Create `.vibe/` inside tracked projects | Use centralized `~/.vibe-dash/` storage |
| Import adapters from core | Keep core dependency-free |
| Use `userId` in JSON/YAML | Use `user_id` (snake_case) |
| Create `Users` table | Use `users` (lowercase, plural) |
| Skip `ctx context.Context` | Always first param in service methods |
| Hardcode storage paths | Use OS abstraction for `~` expansion |

### Technology Stack (Exact Versions)

From [Source: docs/architecture.md#Platform-Dependencies]:

| Technology | Purpose | Import Path |
|------------|---------|-------------|
| Go 1.21+ | Required for slog stdlib | - |
| Bubble Tea | TUI framework | github.com/charmbracelet/bubbletea |
| Bubbles | TUI components | github.com/charmbracelet/bubbles |
| Lipgloss | TUI styling | github.com/charmbracelet/lipgloss |
| Cobra | CLI framework | github.com/spf13/cobra |
| Viper | Config cascade | github.com/spf13/viper |
| sqlx | SQL with struct scanning | github.com/jmoiron/sqlx |
| go-sqlite3 | SQLite driver | github.com/mattn/go-sqlite3 |
| fsnotify | File watching | github.com/fsnotify/fsnotify |

### Makefile Template

From [Source: docs/architecture.md#Build-Distribution]:

```makefile
.PHONY: build test test-all lint fmt check-fmt run clean install test-accuracy

# CGO_ENABLED=1 required for go-sqlite3
build:
	CGO_ENABLED=1 go build -o bin/vibe ./cmd/vibe

test:
	go test ./...

test-all:
	go test -tags=integration ./...

lint:
	golangci-lint run

fmt:
	goimports -w .

check-fmt:
	@test -z "$$(goimports -l .)" || (echo "Run 'make fmt' to fix formatting" && exit 1)

run: build
	./bin/vibe

clean:
	rm -rf bin/

install:
	CGO_ENABLED=1 go install ./cmd/vibe

# Placeholder for detection accuracy testing (95% threshold)
test-accuracy:
	@echo "Detection accuracy tests not yet implemented"
```

**CGO Requirement:** The `go-sqlite3` driver requires CGO. Ensure `CGO_ENABLED=1` for all build commands. This affects cross-compilation - native compilation on target platform recommended.

### main.go Pattern

From [Source: docs/architecture.md#Graceful-Shutdown-Pattern]:

```go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"log/slog"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		slog.Info("shutdown signal received")
		cancel()
	}()

	// Run application with cancellable context
	if err := run(ctx); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Placeholder - will wire up adapters and launch CLI/TUI
	slog.Info("vibe-dash starting")
	return nil
}
```

### .golangci.yml Configuration

```yaml
run:
  timeout: 5m
  go: "1.21"

linters:
  enable:
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - typecheck
    - goimports
    - misspell

linters-settings:
  goimports:
    local-prefixes: github.com/JeiKeiLim/vibe-dash

issues:
  exclude-dirs:
    - test/fixtures
```

### .gitignore Content

```gitignore
# Build artifacts
bin/
dist/

# Database files
*.db
*.db-shm
*.db-wal

# IDE and editor files
.idea/
.vscode/settings.json
*.swp
*.swo
*~

# OS files
.DS_Store
Thumbs.db

# Test coverage
*.out
coverage.html

# Go
vendor/
```

### CI Pipeline Template

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Install goimports
        run: go install golang.org/x/tools/cmd/goimports@latest

      - name: Check formatting
        run: make check-fmt

      - name: Install golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --timeout=5m

      - name: Run tests
        run: make test-all

      - name: Build
        run: make build
```

### Test Fixtures Structure

From [Source: docs/architecture.md#Test-Organization] and [Source: docs/epics.md#Story-2.10]:

```
test/fixtures/
├── speckit-stage-specify/     # Speckit at Specify stage
├── speckit-stage-plan/        # Speckit at Plan stage
├── speckit-stage-tasks/       # Speckit at Tasks stage
├── speckit-stage-implement/   # Speckit at Implement stage (CRITICAL - 4th stage)
├── speckit-uncertain/         # Edge case - ambiguous markers
├── no-method-detected/        # No methodology markers
└── empty-project/             # Empty directory
```

**Pattern:** `{method}-stage-{stage}` for normal cases, `{method}-{scenario}` for edge cases.
**Purpose:** Golden path test suite for 95% detection accuracy validation (launch blocker).

### Project Structure Notes

- **Alignment:** Follows hexagonal architecture exactly as specified in architecture.md
- **Naming:** All directories use lowercase, no underscores (Go convention)
- **Packages:** Each directory will become a Go package with matching name
- **Internal:** Using `internal/` prevents external imports of implementation details

### References

- [Source: docs/architecture.md#Complete-Project-Directory-Structure] - Full directory tree
- [Source: docs/architecture.md#Architectural-Boundaries] - Dependency flow rules
- [Source: docs/architecture.md#Go-Code-Conventions] - Naming conventions
- [Source: docs/architecture.md#Build-Distribution] - Makefile structure
- [Source: docs/architecture.md#Graceful-Shutdown-Pattern] - Signal handling
- [Source: docs/project-context.md] - Critical rules and patterns
- [Source: docs/epics.md#Story-1.1] - Original story definition

## Dev Agent Record

### Context Reference

Story context created by SM agent from:
- docs/epics.md (Story 1.1 requirements)
- docs/architecture.md (technical specifications)
- docs/project-context.md (critical rules)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - First story, no previous implementation

### Completion Notes List

- All directories created following hexagonal architecture structure
- Go module initialized with Go 1.24.3 (exceeds 1.21+ requirement)
- All required dependencies added: bubbletea, bubbles, lipgloss, cobra, viper, sqlx, go-sqlite3, fsnotify
- main.go implements graceful shutdown pattern with signal handling
- Created root.go in internal/adapters/cli with Cobra root command
- Makefile includes all required targets: build, test, test-all, lint, fmt, check-fmt, run, clean, install, test-accuracy
- .golangci.yml configured with recommended linters and test/fixtures excluded
- CI pipeline configured for GitHub Actions with Go 1.21, lint, test, and build jobs
- All verification steps passed: build successful, lint clean, tests pass, binary runs

### Change Log

| Date | Change | Author |
|------|--------|--------|
| 2025-12-12 | Initial project scaffolding complete | Dev Agent (Amelia) |
| 2025-12-12 | Code Review fixes: Context properly wired through CLI (H3/H4), CI Go version aligned to 1.24 (M2), LICENSE year updated to 2025 (L1) | Code Review (Amelia) |

### File List

**Files Created:**

**Entry Point & Config:**
- cmd/vibe/main.go (NEW)
- internal/adapters/cli/root.go (NEW)
- go.mod (NEW)
- go.sum (NEW)
- Makefile (NEW)
- .golangci.yml (NEW)
- .gitignore (NEW)
- README.md (NEW)
- LICENSE (NEW)

**CI Pipeline:**
- .github/workflows/ci.yml (NEW)

**Core Domain (hexagonal - ZERO external imports):**
- internal/core/domain/.keep (NEW)
- internal/core/ports/.keep (NEW)
- internal/core/services/.keep (NEW)

**Adapters (infrastructure implementations):**
- internal/adapters/tui/.keep (NEW)
- internal/adapters/tui/components/.keep (NEW)
- internal/adapters/persistence/sqlite/.keep (NEW)
- internal/adapters/persistence/sqlite/migrations/.keep (NEW)
- internal/adapters/persistence/yaml/.keep (NEW)
- internal/adapters/filesystem/.keep (NEW)
- internal/adapters/detectors/.keep (NEW)
- internal/adapters/detectors/speckit/.keep (NEW)

**Configuration:**
- internal/config/.keep (NEW)

**Test Fixtures (golden path - 95% accuracy requirement):**
- test/fixtures/speckit-stage-specify/.keep (NEW)
- test/fixtures/speckit-stage-plan/.keep (NEW)
- test/fixtures/speckit-stage-tasks/.keep (NEW)
- test/fixtures/speckit-stage-implement/.keep (NEW)
- test/fixtures/speckit-uncertain/.keep (NEW)
- test/fixtures/no-method-detected/.keep (NEW)
- test/fixtures/empty-project/.keep (NEW)

**Optional:**
- pkg/.keep (NEW)

**Build Artifacts (generated, in .gitignore):**
- bin/vibe (generated binary)

**Total: 30 files created**
