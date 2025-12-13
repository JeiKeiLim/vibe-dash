---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8]
inputDocuments:
  - 'docs/prd.md'
  - 'docs/analysis/brainstorming-session-2025-12-04T00:25:43.430Z.md'
  - 'docs/analysis/product-brief-bmad-test-2025-12-05.md'
  - 'docs/analysis/research/shards/00-index.md'
  - 'docs/analysis/research/shards/01-technology-stack.md'
  - 'docs/analysis/research/shards/02-architectural-patterns.md'
  - 'docs/analysis/research/shards/03-implementation-techniques.md'
  - 'docs/analysis/research/shards/04-vibe-coding-methods-CORRECTED.md'
  - 'docs/analysis/research/shards/05-technical-recommendations.md'
  - 'docs/analysis/research/shards/06-architecture-decisions.md'
  - 'docs/analysis/research/shards/07-implementation-roadmap.md'
  - 'docs/analysis/research/shards/08-risk-assessment.md'
  - 'docs/analysis/research/shards/09-executive-summary.md'
workflowType: 'architecture'
lastStep: 1
project_name: 'bmad-test'
user_name: 'Jongkuk Lim'
date: '2025-12-08'
---

# Architecture Decision Document

_This document builds collaboratively through step-by-step discovery. Sections are appended as we work through each architectural decision together._

## Project Context Analysis

### Requirements Overview

**Functional Requirements:**

66 functional requirements across 8 domains drive the architecture:

| Domain | Count | Architectural Impact |
|--------|-------|---------------------|
| Project Management | 8 | Core domain entity, path resolution, collision handling |
| Workflow Detection | 6 | Plugin interface, heuristic engine, uncertainty handling |
| Dashboard Visualization | 13 | TUI component hierarchy, real-time updates, keyboard navigation |
| Project State Management | 6 | State machine (active/hibernated/favorite), auto-promotion logic |
| Agent Monitoring | 5 | **Killer feature** - Inactivity detection, time-based heuristics, uncertainty model |
| Configuration Management | 9 | Cascading config system, YAML + SQLite persistence |
| Scripting & Automation | 14 | Non-interactive CLI layer, JSON serialization, exit codes |
| Error Handling | 5 | Recovery strategies, graceful degradation |

**Killer Feature Elevation - Agent Waiting Detection (FR34-38):**

This is the primary differentiator. Architecture must treat this as first-class:
- Dedicated detection engine with configurable thresholds
- Explicit uncertainty model for edge cases (vacation, inactive projects)
- Clear heuristic boundaries (10-minute MVP threshold, sub-1-minute post-MVP)
- Visual indicator system (⏸️ WAITING) integrated into TUI core

**Non-Functional Requirements:**

17 NFRs establish hard architectural constraints:

- **Performance:** <100ms render, <1s startup, <10s file detection (NFR-P1 through P6)
- **Reliability:** 95% detection accuracy - **launch blocker**, artifact-based recovery (NFR-R1 through R6)
- **Usability:** 1-minute onboarding, self-documenting TUI (NFR-U1 through U6)
- **Extensibility:** MethodDetector plugin interface, beta-phase flexibility (NFR-E1 through E6)

**Scale & Complexity:**

- Primary domain: CLI Developer Tool
- Complexity level: **Medium-High** (TUI + plugins + real-time file watching + SQLite + cross-platform)
- Estimated architectural components: 8-10 major modules

### Technical Constraints & Dependencies

**Hard Constraints:**
1. **Go language** - Required for Bubble Tea TUI, single binary distribution, cross-platform
2. **Centralized storage** (`~/.vibe-dash/`) - Avoids polluting project directories with tracking files
3. **SQLite per-project** - Lightweight embedded database, no external dependencies
4. **fsnotify** - File system watching (with platform-specific limitations)
5. **Artifact-based truth** - No user-maintained state; all detection from file scanning

**Platform Dependencies:**
- Go 1.21+ toolchain
- Bubble Tea (TUI framework)
- Cobra (CLI framework + shell completion)
- fsnotify (file watching)
- SQLite (embedded database)

**Platform Support Strategy:**
- **MVP:** Linux + macOS (primary development/dogfooding platforms)
- **Post-MVP:** Windows support via OS abstraction layer
- **Architecture:** OS abstraction layer from day 1 enables Windows addition without core rewrites

### Cross-Cutting Concerns Identified

1. **OS Abstraction Layer (Platform Independence)**
   - Abstract file system operations behind interface
   - Canonical path resolution (`filepath.EvalSymlinks()` + platform variants)
   - fsnotify event normalization across platforms
   - Home directory resolution (`~/.vibe-dash/` → platform-appropriate)
   - **MVP:** Linux/macOS implementations
   - **Post-MVP:** Windows implementation (junction points, different event semantics)

2. **Configuration Cascade**
   - Priority: CLI flags → project config → master config → defaults
   - Two file types: YAML (human-editable) + SQLite (state)
   - Auto-creation on first use

3. **Plugin Architecture**
   - MethodDetector interface for methodology support
   - Speckit as reference implementation
   - BMAD-Method as post-MVP plugin
   - Interface marked beta until community stabilization

4. **Error Recovery Strategy**
   - Artifacts are truth → re-scan recovers state
   - SQLite corruption → reinitialize from artifacts
   - Config syntax errors → report without corrupting state
   - fsnotify failures → fallback to manual refresh

5. **Database Connection Strategy**
   - Lazy-load SQLite connections (not all 20 projects at once)
   - Connection pooling or single-connection-per-operation pattern
   - Prevents file handle exhaustion at scale

### Testing Strategy Foundation

**Golden Path Test Suite (Detection Oracle):**
- 20 real Speckit projects with known stages serve as **ground truth**
- Test fixtures ARE the truth (artifact-based validation)
- 95% accuracy threshold enforced via automated test pipeline
- Detection accuracy measurement integrated into CI

**Platform Test Matrix:**
- MVP: Linux + macOS CI runners
- Post-MVP: Add Windows CI runner when OS abstraction layer extended

**Heuristic Edge Case Coverage:**
- Agent Waiting Detection requires explicit edge case tests:
  - Project with no recent activity (not waiting, just dormant)
  - User on vacation (extended inactivity)
  - Rapid file changes followed by silence
  - Threshold boundary conditions (9:59 vs 10:01)

**Uncertainty Model:**
- Detection confidence levels: Certain, Likely, Uncertain (🤷)
- Architecture must propagate uncertainty through detection pipeline
- UI reflects confidence level to user (no false precision)

### FR-to-Component Mapping (Preliminary)

| Component | Primary FRs | Secondary FRs |
|-----------|-------------|---------------|
| Project Manager | FR1-8 | FR39-45 |
| Method Detector (Plugin) | FR9-14 | - |
| TUI Dashboard | FR15-27 | FR65-66 |
| State Manager | FR28-33 | FR57 |
| Agent Monitor | FR34-38 | - |
| Config Manager | FR39-47 | FR41, FR46-47 |
| CLI Layer | FR48-61 | FR62-64 |
| File Watcher | FR27 | FR23, FR58 |
| OS Abstraction | (cross-cutting) | FR6-7, FR45 |

## Starter Template Evaluation

### Primary Technology Domain

**CLI Tool / Developer Tooling** - Go-based terminal application with interactive TUI

Technology stack predetermined by PRD:
- Go 1.21+
- Bubble Tea (TUI framework)
- Cobra (CLI framework)
- fsnotify (file watching)
- SQLite (embedded database)

### Starter Options Considered

| Option | Source | Verdict |
|--------|--------|---------|
| bubbletea-app-template | Charmbracelet official | Too minimal - no Cobra, no architecture |
| go-cli-template | Community | Missing plugin structure, no hexagonal |
| Custom hexagonal structure | Custom | **Selected** - fits requirements |

### Selected Approach: Custom Hexagonal Structure

**Rationale for Selection:**

1. No existing starter combines Cobra + Bubble Tea + Hexagonal architecture
2. Plugin system (MethodDetector interface) requires ports & adapters pattern
3. PRD specifies hexagonal architecture explicitly
4. Reference apps (k9s, kubectl) use custom structures

**Project Structure:**

```
vibe-dash/
├── cmd/
│   └── vibe/
│       └── main.go              # Entry point (binary name remains 'vibe')
├── internal/
│   ├── core/                    # Domain layer (no external dependencies)
│   │   ├── domain/              # Entities: Project, Stage, DetectionResult
│   │   ├── ports/               # Interfaces: MethodDetector, ProjectRepository
│   │   └── services/            # Use cases: DetectionService, StateService
│   ├── adapters/                # Infrastructure layer
│   │   ├── cli/                 # Cobra commands
│   │   ├── tui/                 # Bubble Tea components
│   │   ├── persistence/         # SQLite + YAML adapters
│   │   ├── filesystem/          # OS abstraction, fsnotify
│   │   └── detectors/           # MethodDetector implementations
│   │       ├── speckit/         # Speckit detector plugin
│   │       └── bmad/            # BMAD detector plugin (post-MVP)
│   └── config/                  # Configuration loading
├── pkg/                         # Shareable utilities (if any)
├── test/
│   └── fixtures/                # Golden path test projects
├── go.mod
├── go.sum
└── Makefile
```

### Architectural Decisions Provided by Structure

**Language & Runtime:**
- Go 1.21+ with modules
- Single binary compilation
- Cross-platform build targets (Linux, macOS, Windows post-MVP)

**Code Organization:**
- Hexagonal architecture (ports & adapters)
- `internal/core/` contains pure domain logic with zero external dependencies
- `internal/adapters/` contains all infrastructure implementations
- Dependency injection via interfaces (ports)

**CLI Framework:**
- Cobra for command parsing and shell completion
- Commands in `internal/adapters/cli/`
- Root command launches TUI, subcommands for non-interactive mode

**TUI Framework:**
- Bubble Tea with Elm architecture (Model → Update → View)
- Components in `internal/adapters/tui/`
- Bubbles library for common components (spinner, table, text input)
- Lipgloss for styling

**Plugin Architecture:**
- `MethodDetector` interface defined in `internal/core/ports/`
- Implementations in `internal/adapters/detectors/`
- Registry pattern for detector discovery
- Speckit detector as reference implementation
- **Future consideration:** Scripting language plugins (Lua/Starlark) if community demands - deferred, not MVP

**Persistence:**
- SQLite adapter in `internal/adapters/persistence/`
- YAML config adapter in same location
- Repository interfaces in `internal/core/ports/`

**File System:**
- OS abstraction layer in `internal/adapters/filesystem/`
- fsnotify watcher with event debouncing
- Platform-specific implementations behind interface

**Testing:**
- Golden path fixtures in `test/fixtures/`
- Interface-based design enables mock injection
- Table-driven tests for detection accuracy

**Initialization Command:**

```bash
mkdir -p vibe-dash/cmd/vibe vibe-dash/internal/{core/{domain,ports,services},adapters/{cli,tui,persistence,filesystem,detectors/speckit},config} vibe-dash/test/fixtures
cd vibe-dash && go mod init github.com/JeiKeiLim/vibe-dash
```

**Note:** Project initialization and scaffolding should be the first implementation story.

## Core Architectural Decisions

### Decision Summary

| Category | Decision | Rationale |
|----------|----------|-----------|
| SQL Library | sqlx | Lightweight, honest SQL, struct scanning |
| Config Management | Viper | Cobra companion, handles config cascade automatically |
| Error Handling | Standard errors (Go 1.21+) | Idiomatic, domain error types in core |
| Logging | log/slog | Built-in Go 1.21+, structured, zero dependencies |
| Build Tool | Makefile | Simple for MVP, GoReleaser added at release time |

### Data Architecture

**SQL Library: sqlx**
- Lightweight SQL extensions over database/sql
- Direct SQL queries with struct scanning
- No ORM magic - full control over queries
- Well-suited for simple schema (projects, detection states, timestamps)

**Schema Location:** `internal/adapters/persistence/sqlite/`

**SQLite Concurrency: WAL Mode**
- Enable Write-Ahead Logging for concurrent read access
- Single writer, multiple readers pattern
- Prevents lock contention between file watcher goroutine and TUI refresh
- Connection opened on-demand, closed when operation completes (no persistent pool)

**Schema Versioning & Migration:**
- `schema_version` table tracks current version
- Migration files in `internal/adapters/persistence/sqlite/migrations/`
- On startup: check version, apply pending migrations sequentially
- Simple versioning: `v1`, `v2`, `v3` - no complex migration framework needed for CLI tool

### Configuration System

**Library: Viper**
- Handles configuration cascade
- Native integration with Cobra CLI framework
- YAML file parsing built-in
- Environment variable support if needed later

**Config Files:**
- Master: `~/.vibe-dash/config.yaml`
- Per-project: `~/.vibe-dash/<project>/config.yaml`

**Configuration Precedence (Explicit):**
```
1. CLI flags          (highest priority)  --hibernation-days=7
2. Project config     ~/.vibe-dash/<project>/config.yaml
3. Master config      ~/.vibe-dash/config.yaml
4. Built-in defaults  (lowest priority)   hibernation_days: 14
```

Viper merges these layers automatically. CLI flags always win.

### Error Handling Strategy

**Approach: Standard Go errors with domain types**

Domain error types defined in `internal/core/domain/errors.go`:
- `ErrProjectNotFound`
- `ErrDetectionFailed`
- `ErrConfigInvalid`
- `ErrPathNotAccessible`

**Error-to-Exit-Code Mapping:**

| Domain Error | Exit Code | PRD Reference |
|--------------|-----------|---------------|
| (none) | 0 | Success |
| (any unhandled) | 1 | General error |
| `ErrProjectNotFound` | 2 | Project not found |
| `ErrConfigInvalid` | 3 | Invalid configuration |
| `ErrDetectionFailed` | 4 | Detection failure |

Mapping implemented in CLI adapter layer (`internal/adapters/cli/`).

Error wrapping pattern:
```go
fmt.Errorf("failed to scan project %s: %w", path, err)
```

Adapters wrap with context, services return domain errors, CLI/TUI formats for display.

### Logging & Observability

**Library: log/slog (Go 1.21+ stdlib)**

Log levels controlled by CLI flags:
- Default: Errors only (quiet operation)
- `--verbose` / `-v`: Info level
- `--debug`: Debug level with file/line info

Logs to stderr to avoid interfering with JSON output in scripting mode.

**Testing Note:** Consider structured log capture in tests for debugging test failures.

### Build & Distribution

**MVP: Makefile**

```makefile
.PHONY: build test run clean lint

build:
	go build -o bin/vibe ./cmd/vibe

test:
	go test ./...

lint:
	golangci-lint run

run: build
	./bin/vibe

clean:
	rm -rf bin/
```

**Linting: golangci-lint**
- Run from day 1 to catch bugs early and enforce consistency
- Configure via `.golangci.yml` in project root
- Include in CI pipeline

**Post-MVP: GoReleaser** for automated multi-platform releases when publishing to GitHub.

### Future API Considerations

**Decision: Design services API-friendly, no API code in MVP**

- Services in `internal/core/services/` return domain types, not TUI-specific formats
- Hexagonal architecture already supports future API adapter
- When web view needed: add `internal/adapters/api/http/`
- Zero extra work now - architecture handles it

## Implementation Patterns & Consistency Rules

### Purpose

These patterns ensure AI agents implementing different parts of vibe-dash produce consistent, compatible code. Following these prevents merge conflicts and integration issues.

### Go Code Conventions

**Follow standard Go idioms - non-negotiable:**

| Element | Convention | Example |
|---------|------------|---------|
| Package names | lowercase, short, no underscores | `domain`, `sqlite`, `speckit` |
| File names | lowercase, underscores OK | `project_repository.go`, `detection_service.go` |
| Exported (public) | PascalCase | `Project`, `DetectionService`, `Detect()` |
| Unexported (private) | camelCase | `parseConfig`, `lastScanTime` |
| Interfaces | -er suffix when behavior | `Detector`, `Repository`, `Watcher` |
| Acronyms | ALL CAPS | `ID`, `URL`, `HTTP`, `SQL` |
| Constants | PascalCase or camelCase | `MaxProjects`, `defaultTimeout` |

**Context Propagation - required for cancellation:**

All service methods accept `context.Context` as first parameter:

```go
// Good - supports TUI cancellation when user hits 'q'
func (s *DetectionService) Detect(ctx context.Context, path string) (*Result, error)

// Bad - no cancellation support
func (s *DetectionService) Detect(path string) (*Result, error)
```

**Constructor Pattern - use `New*` functions:**

```go
// Standard Go constructor pattern
func NewDetectionService(repo ProjectRepository, detector MethodDetector) *DetectionService {
    return &DetectionService{
        repo:     repo,
        detector: detector,
    }
}
```

**Formatting - enforced by tooling:**

| Tool | Purpose | Usage |
|------|---------|-------|
| `goimports` | Format + manage imports | `make fmt` before commit |
| `golangci-lint` | Linting + style checks | `make lint` in CI |

```makefile
fmt:
	goimports -w .

check-fmt:
	@test -z "$$(goimports -l .)" || (echo "Run 'make fmt' to fix formatting" && exit 1)
```

### Database Naming Conventions

**SQLite tables and columns:**

| Element | Convention | Example |
|---------|------------|---------|
| Tables | snake_case, plural | `projects`, `detection_states` |
| Columns | snake_case | `project_id`, `last_scanned_at` |
| Primary keys | `id` | `id INTEGER PRIMARY KEY` |
| Foreign keys | `{singular_table}_id` | `project_id` |
| Timestamps | `*_at` suffix | `created_at`, `updated_at`, `hibernated_at` |
| Booleans | `is_*` or `has_*` prefix | `is_favorite`, `has_errors` |

**Example schema:**
```sql
CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,
    display_name TEXT,
    is_favorite INTEGER DEFAULT 0,
    detected_method TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```

### JSON/YAML Format Conventions

**For config files (`~/.vibe-dash/config.yaml`) and CLI JSON output:**

| Element | Convention | Example |
|---------|------------|---------|
| Keys | snake_case | `hibernation_days`, `project_path` |
| Booleans | true/false | `favorite: true` |
| Timestamps | ISO 8601 UTC | `"2025-12-11T10:30:00Z"` |
| Nulls | Omit field when possible | Don't include `field: null` |
| Arrays | Plural key names | `projects:`, `stages:` |

**Example config:**
```yaml
settings:
  hibernation_days: 14
  refresh_interval_seconds: 10
  refresh_debounce_ms: 200
  agent_waiting_threshold_minutes: 10

projects:
  my-project:
    path: "/home/user/my-project"
    favorite: false
```

### Test Organization Patterns

| Element | Convention | Example |
|---------|------------|---------|
| Location | Same package, `_test.go` suffix | `detector_test.go` beside `detector.go` |
| Function naming | `Test{Function}_{Scenario}` | `TestDetect_SpeckitProject` |
| Table-driven tests | `testCases` or `tests` slice | Standard Go pattern |
| Fixtures | `test/fixtures/` directory | Golden path test projects |
| Mocks | Interface-based, in test file | No mock framework needed |

**Test Fixture Naming Convention:**

```
test/fixtures/
├── speckit-stage-specify/     # Speckit project at Specify stage
├── speckit-stage-plan/        # Speckit project at Plan stage
├── speckit-stage-tasks/       # Speckit project at Tasks stage
├── speckit-uncertain/         # Edge case - unclear stage
├── bmad-stage-prd/            # BMAD project at PRD stage (post-MVP)
├── no-method-detected/        # No methodology markers found
└── empty-project/             # Empty directory
```

Pattern: `{method}-stage-{stage}` for normal cases, `{method}-{scenario}` for edge cases.

**Example test structure:**
```go
func TestDetect_SpeckitProject(t *testing.T) {
    tests := []struct {
        name     string
        path     string
        expected Stage
    }{
        {"spec exists", "fixtures/speckit-stage-specify", StageSpecify},
        {"plan exists", "fixtures/speckit-stage-plan", StagePlan},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### Logging Patterns

**Using log/slog consistently:**

| Element | Convention | Example |
|---------|------------|---------|
| Key names | snake_case | `"project_id"`, `"scan_duration_ms"` |
| Error logging | At handling site only | Log once where error is handled |
| Context | Include relevant identifiers | `slog.Info("scanning", "path", path)` |
| Levels | Error (user errors), Info (operations), Debug (internals) | Match CLI flags |

**Example usage:**
```go
// Good - structured, at handling site
slog.Error("detection failed", "path", path, "error", err)

// Bad - logging during propagation
return fmt.Errorf("scan failed: %w", err) // don't also log here
```

### File Watcher Patterns

**Debounce Configuration:**

File system events can fire rapidly during saves (editor writes temp file, renames, updates metadata). Without debouncing, this causes excessive CPU usage and flickering UI.

| Setting | Default | Range | Config Key |
|---------|---------|-------|------------|
| Debounce window | 200ms | 100-500ms | `refresh_debounce_ms` |

**Implementation Pattern:**

```go
// Debounce file events before triggering rescan
type DebouncedWatcher struct {
    watcher  *fsnotify.Watcher
    debounce time.Duration
    timer    *time.Timer
    mu       sync.Mutex
}

func (d *DebouncedWatcher) handleEvent(event fsnotify.Event) {
    d.mu.Lock()
    defer d.mu.Unlock()

    // Reset timer on each event - only fires after debounce period of silence
    if d.timer != nil {
        d.timer.Stop()
    }
    d.timer = time.AfterFunc(d.debounce, func() {
        // Trigger actual rescan here
    })
}
```

**Why 200ms default:**
- Fast enough to feel responsive (user sees update within 200ms of last save)
- Slow enough to batch rapid events (editor save sequences complete in ~100ms)
- Configurable for users with different needs

### Graceful Shutdown Pattern

**Shutdown Sequence:**

When user presses Ctrl+C or sends SIGTERM, the application must shut down cleanly to prevent data corruption:

```
1. Receive signal (SIGINT/SIGTERM)
        │
        ▼
2. Cancel root context
   └── All goroutines see ctx.Done()
        │
        ▼
3. Wait for in-flight operations (5s timeout)
   └── File watcher stops
   └── Active scans complete or abort
        │
        ▼
4. Flush pending writes
   └── SQLite WAL checkpoint
        │
        ▼
5. Close database connections
        │
        ▼
6. Exit with code 0 (or 1 if timeout exceeded)
```

**Implementation Pattern:**

```go
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
        os.Exit(1)
    }
}

func run(ctx context.Context) error {
    // ... setup ...

    defer func() {
        // Cleanup with timeout
        cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cleanupCancel()

        if err := db.Close(cleanupCtx); err != nil {
            slog.Error("cleanup failed", "error", err)
        }
    }()

    // ... run TUI/CLI ...
}
```

**Critical Rules:**
- Never call `os.Exit()` directly from goroutines - always use context cancellation
- Database writes must respect context cancellation
- 5-second timeout prevents hanging on stuck operations

### All AI Agents MUST

1. **Run `make fmt` before committing** - No unformatted code
2. **Run `make lint` before PR** - No linting errors
3. **Follow snake_case for database/JSON, Go conventions for code** - No mixing
4. **Write table-driven tests** - Standard Go testing pattern
5. **Log at error handling site only** - No duplicate logging
6. **Use domain error types** - Return `ErrProjectNotFound`, not raw errors
7. **Accept `context.Context` as first parameter** - All service methods
8. **Use `New*` constructor functions** - Standard Go pattern
9. **Follow fixture naming convention** - `{method}-stage-{stage}` pattern

### Anti-Patterns to Avoid

| Don't | Do Instead |
|-------|------------|
| `userId` in JSON | `user_id` |
| `Projects` table name | `projects` |
| `get_user_data()` function | `getUserData()` |
| Tests in separate `/tests` folder | `_test.go` co-located |
| Log same error at every layer | Log once at handling site |
| Custom error strings | Domain error types |
| Service methods without context | `func (s *Svc) Do(ctx context.Context, ...)` |
| Random fixture names | `speckit-stage-plan` pattern |

## Project Structure & Boundaries

### Complete Project Directory Structure

```
vibe-dash/
├── .github/
│   └── workflows/
│       └── ci.yml                    # GitHub Actions: lint, test, build
├── .golangci.yml                     # golangci-lint configuration
├── .gitignore
├── LICENSE
├── Makefile                          # build, test, lint, fmt, run, clean
├── README.md
├── go.mod
├── go.sum
│
├── cmd/
│   └── vibe/
│       └── main.go                   # Entry point - wires up adapters, launches CLI
│
├── internal/
│   ├── core/                         # Domain layer - ZERO external dependencies
│   │   ├── domain/
│   │   │   ├── project.go            # Project entity
│   │   │   ├── stage.go              # Stage enum and methods
│   │   │   ├── detection_result.go   # Detection result value object
│   │   │   ├── confidence.go         # Confidence levels: Certain, Likely, Uncertain
│   │   │   └── errors.go             # Domain errors: ErrProjectNotFound, etc.
│   │   ├── ports/
│   │   │   ├── detector.go           # MethodDetector interface
│   │   │   ├── repository.go         # ProjectRepository interface
│   │   │   ├── watcher.go            # FileWatcher interface
│   │   │   └── config.go             # ConfigLoader interface
│   │   └── services/
│   │       ├── detection_service.go  # Orchestrates detection across detectors
│   │       ├── project_service.go    # Project CRUD, state management
│   │       ├── state_service.go      # Active/hibernated/favorite transitions
│   │       └── agent_monitor.go      # Killer feature - Agent Waiting Detection
│   │
│   ├── adapters/
│   │   ├── cli/
│   │   │   ├── root.go               # Root command - launches TUI by default
│   │   │   ├── add.go                # vibe add <path>
│   │   │   ├── remove.go             # vibe remove <name>
│   │   │   ├── list.go               # vibe list [--json]
│   │   │   ├── status.go             # vibe status <name> [--json]
│   │   │   ├── scan.go               # vibe scan [--all]
│   │   │   ├── config.go             # vibe config [get|set]
│   │   │   └── exitcodes.go          # Maps domain errors to exit codes
│   │   │
│   │   ├── tui/
│   │   │   ├── app.go                # Main Bubble Tea application model
│   │   │   ├── dashboard.go          # Dashboard view - project list with stages
│   │   │   ├── project_detail.go     # Project detail view
│   │   │   ├── help.go               # Help overlay (? key)
│   │   │   ├── styles.go             # Lipgloss styles
│   │   │   ├── keys.go               # Key bindings
│   │   │   └── components/
│   │   │       ├── project_row.go    # Single project row component
│   │   │       ├── stage_badge.go    # Stage indicator with color
│   │   │       ├── waiting_indicator.go  # ⏸️ WAITING indicator (killer feature UI)
│   │   │       └── status_bar.go     # Bottom status bar
│   │   │
│   │   ├── persistence/
│   │   │   ├── sqlite/
│   │   │   │   ├── repository.go     # SQLite ProjectRepository implementation
│   │   │   │   ├── schema.go         # Schema definition and version checking
│   │   │   │   ├── migrations/
│   │   │   │   │   ├── 001_initial.sql
│   │   │   │   │   └── migrations.go # Migration runner
│   │   │   │   └── queries.go        # SQL query constants
│   │   │   └── yaml/
│   │   │       └── config_loader.go  # YAML config file adapter
│   │   │
│   │   ├── filesystem/
│   │   │   ├── platform.go           # OS abstraction interface
│   │   │   ├── platform_unix.go      # Linux/macOS implementation
│   │   │   ├── platform_windows.go   # Windows implementation (post-MVP stub)
│   │   │   ├── watcher.go            # fsnotify FileWatcher implementation
│   │   │   ├── paths.go              # Path resolution utilities
│   │   │   └── paths_test.go         # Path resolution tests
│   │   │
│   │   └── detectors/
│   │       ├── registry.go           # Detector registry - ONLY place that knows all detectors
│   │       └── speckit/
│   │           ├── detector.go       # Speckit MethodDetector implementation
│   │           ├── stages.go         # Speckit stage definitions
│   │           └── detector_test.go  # Speckit detection tests
│   │
│   └── config/
│       ├── config.go                 # Configuration struct and defaults
│       ├── loader.go                 # Viper integration, config cascade
│       └── defaults.go               # Built-in default values
│
└── test/
    └── fixtures/
        ├── speckit-stage-specify/    # Speckit at Specify stage
        │   └── spec.md
        ├── speckit-stage-plan/       # Speckit at Plan stage
        │   ├── spec.md
        │   └── plan.md
        ├── speckit-stage-tasks/      # Speckit at Tasks stage
        │   ├── spec.md
        │   ├── plan.md
        │   └── tasks.md
        ├── speckit-uncertain/        # Edge case - ambiguous markers
        │   └── partial-spec.md
        ├── no-method-detected/       # No methodology markers
        │   └── README.md
        └── empty-project/            # Empty directory
```

### Architectural Boundaries

**Dependency Flow (Hexagonal):**

```
                    ┌─────────────────────────────────────────┐
                    │              cmd/vibe/                  │
                    │            (entry point)                │
                    └─────────────────┬───────────────────────┘
                                      │ wires up
                    ┌─────────────────▼───────────────────────┐
                    │           internal/adapters/            │
                    │  ┌─────┐  ┌─────┐  ┌──────────────────┐ │
                    │  │ CLI │  │ TUI │  │   persistence/   │ │
                    │  └──┬──┘  └──┬──┘  │ filesystem/      │ │
                    │     │       │      │ detectors/       │ │
                    │     │       │      └────────┬─────────┘ │
                    └─────┼───────┼──────────────┼───────────┘
                          │       │              │
                          │ calls │ calls        │ implements
                    ┌─────▼───────▼──────────────▼───────────┐
                    │           internal/core/               │
                    │  ┌─────────┐  ┌─────────┐  ┌─────────┐ │
                    │  │ domain/ │  │ ports/  │  │services/│ │
                    │  │(entities│  │(interfaces│ │(use     │ │
                    │  │ errors) │  │ only)   │  │ cases)  │ │
                    │  └─────────┘  └─────────┘  └─────────┘ │
                    │                                        │
                    │        ZERO EXTERNAL DEPENDENCIES      │
                    └────────────────────────────────────────┘
```

**Key Boundary Rules:**

| Boundary | Rule |
|----------|------|
| `internal/core/` → external | ❌ FORBIDDEN - core imports nothing from adapters |
| `internal/core/domain/` → services | ❌ FORBIDDEN - entities don't know about services |
| `internal/adapters/` → core | ✅ ALLOWED - adapters implement port interfaces |
| `internal/adapters/` → external libs | ✅ ALLOWED - Bubble Tea, sqlx, fsnotify, Viper |
| `cmd/` → everything | ✅ ALLOWED - wires up dependencies |

**Registry Coordination Role:**

The `adapters/detectors/registry.go` is the **only** component that knows about all detector implementations. Services interact with detectors through the registry, never directly:

```go
// Services call registry, not detectors directly
result := registry.DetectAll(ctx, path)  // Returns first match

// Registry internally iterates all registered detectors
for _, detector := range r.detectors {
    if result, err := detector.Detect(ctx, path); err == nil {
        return result
    }
}
```

### FR Category to Structure Mapping

| FR Category | Primary Location | Files |
|-------------|-----------------|-------|
| **Project Management (FR1-8)** | `core/services/project_service.go` | `domain/project.go`, `ports/repository.go`, `adapters/persistence/` |
| **Workflow Detection (FR9-14)** | `core/services/detection_service.go` | `domain/stage.go`, `ports/detector.go`, `adapters/detectors/` |
| **Dashboard Visualization (FR15-27)** | `adapters/tui/` | `dashboard.go`, `components/`, `styles.go` |
| **Project State Management (FR28-33)** | `core/services/state_service.go` | `domain/project.go` (state field) |
| **Agent Monitoring (FR34-38)** | `core/services/agent_monitor.go` | `adapters/tui/components/waiting_indicator.go` |
| **Configuration Management (FR39-47)** | `internal/config/` | `adapters/persistence/yaml/`, Viper integration |
| **Scripting & Automation (FR48-61)** | `adapters/cli/` | All CLI commands, `exitcodes.go` |
| **Error Handling (FR62-66)** | `core/domain/errors.go` | `adapters/cli/exitcodes.go` |

### Integration Points

**Internal Communication:**

| From | To | Pattern |
|------|-----|---------|
| CLI commands | Services | Direct function calls via injected interfaces |
| TUI | Services | Bubble Tea Cmd returns that call services |
| File watcher | TUI | Bubble Tea Msg sent on file change events |
| Services | Repository | Interface method calls (ports) |
| Services | Detectors | Registry lookup, interface method calls |

**Data Flow:**

```
User adds project
        │
        ▼
    CLI (add.go)
        │
        ▼
ProjectService.Add(ctx, path)
        │
        ├──► PathResolver.Canonicalize(path)
        │
        ├──► DetectionService.Detect(ctx, path)
        │         │
        │         ▼
        │    Registry.DetectAll(ctx, path)
        │         │
        │         ▼
        │    SpeckitDetector.Detect(path)
        │         │
        │         ▼
        │    DetectionResult{Method, Stage, Confidence}
        │
        ▼
ProjectRepository.Save(project)
        │
        ▼
    SQLite write
```

### Test Organization

**Test File Location:**

| Test Type | Location | Naming | Run Command |
|-----------|----------|--------|-------------|
| Unit tests | Same directory as source | `*_test.go` | `go test ./...` |
| Integration tests | Same directory | `*_integration_test.go` | `go test -tags=integration ./...` |
| Fixture data | `test/fixtures/` | `{method}-stage-{stage}/` | N/A |

**Integration Test Build Tags:**

Use build tags to separate slow tests (SQLite, filesystem) from fast unit tests:

```go
//go:build integration

package sqlite_test

func TestRepository_Integration(t *testing.T) {
    // Tests that hit real SQLite database
}
```

Run commands:
- `go test ./...` - Unit tests only (fast, default)
- `go test -tags=integration ./...` - All tests including integration
- CI runs both

**Mock File Guidance:**

| Scenario | Location | Example |
|----------|----------|---------|
| Mock used in one test file | In the test file itself | `detector_test.go` contains `mockDetector` |
| Mock reused across multiple test files | Extract to `*_mock.go` | `detector_mock.go` next to `detector.go` |

Start with mocks in test files. Extract only when duplication appears.

### File Organization Patterns

**Configuration Files Location:**

| File | Location | Purpose |
|------|----------|---------|
| `.golangci.yml` | Project root | Linter configuration |
| `go.mod`, `go.sum` | Project root | Go modules |
| `Makefile` | Project root | Build commands |
| `.github/workflows/ci.yml` | `.github/workflows/` | CI pipeline |
| User config | `~/.vibe-dash/config.yaml` | User preferences |
| Project state | `~/.vibe-dash/<project>/state.db` | SQLite per project |

### Development Workflow

**Build Process:**

```bash
make build      # Compiles to bin/vibe
make test       # Runs unit tests only
make test-all   # Runs all tests including integration
make lint       # Runs golangci-lint
make fmt        # Formats with goimports
make check-fmt  # CI check for formatting
make run        # Build and run TUI
make clean      # Remove build artifacts
```

**CI Pipeline (GitHub Actions):**

```yaml
# .github/workflows/ci.yml
jobs:
  build:
    - checkout
    - setup-go 1.21
    - make check-fmt
    - make lint
    - make test-all
    - make build
```

### Future Extraction Note

When `cmd/vibe/main.go` grows too large with dependency wiring, extract to `internal/app/app.go`. Not needed for MVP - do when complexity warrants it.

## Architecture Validation Results

### Coherence Validation ✅

**Decision Compatibility:**
All technology choices are native Go libraries designed to work together. No version conflicts or incompatibilities detected. Hexagonal architecture cleanly separates concerns and supports all use cases including future API expansion.

**Pattern Consistency:**
Naming conventions are clearly separated (Go idioms for code, snake_case for persistence/JSON). Error handling flows from domain types through adapters to exit codes. Test organization follows standard Go practices with build tags for integration tests.

**Structure Alignment:**
Project structure directly implements hexagonal architecture with clear `core/` and `adapters/` boundaries. All integration points are defined with explicit communication patterns.

### Requirements Coverage Validation ✅

**Functional Requirements Coverage:**
All 66 functional requirements across 8 categories have explicit architectural support with mapped file locations. No FRs are missing architectural backing.

**Non-Functional Requirements Coverage:**
- Performance: Lazy loading, optimized rendering, efficient file watching
- Reliability: 95% accuracy via golden path test suite, artifact-based recovery
- Usability: Self-documenting TUI, help system, 1-minute onboarding target
- Extensibility: MethodDetector interface, registry pattern, future API support

### Implementation Readiness Validation ✅

**Decision Completeness:**
All critical decisions documented with technology versions (Go 1.21+, latest stable for all dependencies). Concrete examples provided for each pattern category.

**Structure Completeness:**
Complete project tree with 50+ files defined, each with purpose annotation. All boundaries explicitly documented with allowed/forbidden dependency rules.

**Pattern Completeness:**
Comprehensive coverage of naming, structure, communication, and process patterns. Anti-patterns documented. Enforcement guidelines specified.

### Gap Analysis Results

**Critical Gaps:** None

**Minor Gaps Identified:**
1. fsnotify debounce timing (5-10 seconds) - implement during file watcher story
2. Shell completion generation - implement as CLI polish story

**Post-MVP Considerations:**
- Performance benchmarking strategy
- Release versioning scheme (SemVer recommended)
- Windows platform implementation

### Architecture Completeness Checklist

**✅ Requirements Analysis**
- [x] Project context thoroughly analyzed (66 FRs, 17 NFRs)
- [x] Scale and complexity assessed (Medium-High)
- [x] Technical constraints identified (Go, centralized storage, artifact-based truth)
- [x] Cross-cutting concerns mapped (5 concerns with solutions)

**✅ Architectural Decisions**
- [x] Critical decisions documented with versions
- [x] Technology stack fully specified (Go, Bubble Tea, Cobra, sqlx, Viper, fsnotify, slog)
- [x] Integration patterns defined (hexagonal, ports & adapters)
- [x] Performance considerations addressed (lazy loading, WAL mode, debouncing)

**✅ Implementation Patterns**
- [x] Naming conventions established (Go idioms, snake_case for DB/JSON)
- [x] Structure patterns defined (co-located tests, fixture naming)
- [x] Communication patterns specified (context propagation, Bubble Tea Msgs)
- [x] Process patterns documented (error handling, logging, build tags)

**✅ Project Structure**
- [x] Complete directory structure defined (50+ files)
- [x] Component boundaries established (core ↔ adapters)
- [x] Integration points mapped (registry, services, repositories)
- [x] Requirements to structure mapping complete (FR → file mapping)

### Architecture Readiness Assessment

**Overall Status:** ✅ READY FOR IMPLEMENTATION

**Confidence Level:** HIGH

**Key Strengths:**
1. Clean hexagonal architecture enables testability and future API expansion
2. Plugin system designed from day 1 - not retrofitted
3. All standard Go idioms - minimal learning curve for Go developers
4. Explicit boundary rules prevent architecture erosion
5. Killer feature (Agent Waiting Detection) has first-class architectural support

**Areas for Future Enhancement:**
1. Windows platform support (OS abstraction layer ready)
2. Additional MethodDetector plugins (BMAD-Method, custom)
3. Web dashboard API adapter
4. Performance benchmarking suite
5. Scripting language plugins if community demands

### 95% Detection Accuracy Measurement

**Launch Blocker Definition:**

The PRD specifies 95% detection accuracy as a launch blocker. Here's how to measure it:

```
accuracy = (correct_detections / total_test_fixtures) * 100
```

**Test Oracle:**
- 20 golden path fixtures in `test/fixtures/`
- Each fixture has a known expected stage
- Detection must match expected stage to count as correct

**Pass/Fail Threshold:**
| Correct | Total | Accuracy | Status |
|---------|-------|----------|--------|
| 20 | 20 | 100% | ✅ Pass |
| 19 | 20 | 95% | ✅ Pass |
| 18 | 20 | 90% | ❌ Blocked |
| 17 | 20 | 85% | ❌ Blocked |

**Implementation:** Create `make test-accuracy` target that runs detection against all fixtures and calculates percentage. CI fails if below 95%.

### Implementation Handoff

**AI Agent Guidelines:**
1. Follow all architectural decisions exactly as documented
2. Use implementation patterns consistently across all components
3. Respect project structure and boundaries (especially core → adapters rule)
4. Refer to this document for all architectural questions
5. Run `make fmt` and `make lint` before every commit

**Dependency Injection Wiring Order (cmd/vibe/main.go):**

```go
// 1. Create adapters (concrete implementations)
repo := sqlite.NewRepository(dbPath)
watcher := filesystem.NewWatcher()
detector := speckit.NewDetector()

// 2. Register detectors with registry
registry := detectors.NewRegistry()
registry.Register(detector)

// 3. Create services with injected ports
detectionSvc := services.NewDetectionService(registry)
projectSvc := services.NewProjectService(repo, detectionSvc)
stateSvc := services.NewStateService(repo)
agentMonitor := services.NewAgentMonitor(watcher)

// 4. Create CLI/TUI with injected services
rootCmd := cli.NewRootCmd(projectSvc, stateSvc, agentMonitor)

// 5. Run
rootCmd.Execute()
```

**First Implementation Priority:**

```bash
# 1. Initialize project structure
mkdir -p vibe-dash/cmd/vibe vibe-dash/internal/{core/{domain,ports,services},adapters/{cli,tui,persistence/sqlite,filesystem,detectors/speckit},config} vibe-dash/test/fixtures
cd vibe-dash && go mod init github.com/JeiKeiLim/vibe-dash

# 2. Create Makefile with all targets

# 3. Implement domain entities (zero dependencies)

# 4. Define port interfaces

# 5. Implement adapters
```

**First User Story (Vertical Slice):**

> "As a developer, I can run `vibe` and see an empty dashboard, so I know the tool is installed correctly."

This validates the entire stack: CLI entry → TUI initialization → empty state rendering. First demonstrable user value.

## Architecture Completion Summary

### Workflow Completion

**Architecture Decision Workflow:** COMPLETED ✅
**Total Steps Completed:** 8
**Date Completed:** 2025-12-11
**Document Location:** docs/architecture.md

### Final Architecture Deliverables

**📋 Complete Architecture Document**
- All architectural decisions documented with specific versions
- Implementation patterns ensuring AI agent consistency
- Complete project structure with all files and directories
- Requirements to architecture mapping
- Validation confirming coherence and completeness

**🏗️ Implementation Ready Foundation**
- 5 core architectural decisions made (sqlx, Viper, slog, standard errors, Makefile)
- 9 implementation pattern categories defined
- 8 architectural components specified
- 66 functional requirements fully supported

**📚 AI Agent Implementation Guide**
- Technology stack with verified versions
- Consistency rules that prevent implementation conflicts
- Project structure with clear boundaries
- Integration patterns and communication standards

### Quality Assurance Checklist

**✅ Architecture Coherence**
- [x] All decisions work together without conflicts
- [x] Technology choices are compatible
- [x] Patterns support the architectural decisions
- [x] Structure aligns with all choices

**✅ Requirements Coverage**
- [x] All functional requirements are supported
- [x] All non-functional requirements are addressed
- [x] Cross-cutting concerns are handled
- [x] Integration points are defined

**✅ Implementation Readiness**
- [x] Decisions are specific and actionable
- [x] Patterns prevent agent conflicts
- [x] Structure is complete and unambiguous
- [x] Examples are provided for clarity

---

**Architecture Status:** READY FOR IMPLEMENTATION ✅

**Next Phase:** Create Epics & Stories, then begin implementation.
