# Story 1.4: Cobra CLI Framework

**Status:** Done

## Quick Reference

| Category | Items |
|----------|-------|
| **Commands** | `vibe` (root), `vibe --help`, `vibe --version` |
| **Global Flags** | `--verbose`, `--debug`, `--config` |
| **Files** | root.go, version.go, flags.go, exitcodes.go (+ tests) |
| **Dependencies** | github.com/spf13/cobra |
| **Exit Codes** | 0 (success), 1 (general error), 2 (not found), 3 (config), 4 (detection) |

### Flag Quick Reference

| Flag | Short | Type | Default | Effect |
|------|-------|------|---------|--------|
| `--verbose` | `-v` | bool | false | Info-level logging to stderr |
| `--debug` | | bool | false | Debug logging with file:line to stderr |
| `--config` | `-c` | string | "" | Custom config file path |
| `--version` | | | | Show version info and exit |
| `--help` | `-h` | | | Show help text and exit |

**Flag Precedence:** When both `--debug` and `--verbose` are set, `--debug` takes precedence (debug includes info-level output).

## Story

**As a** user,
**I want** to run `vibe` command and have it recognized,
**So that** the CLI entry point works.

## Acceptance Criteria

```gherkin
AC1: Given vibe-dash is installed
     When I run `vibe --help`
     Then I see help text with:
       - Description: "CLI dashboard for vibe coding projects"
       - Available commands placeholder
       - Global flags: --verbose, --debug, --config

AC2: Given vibe-dash is installed
     When I run `vibe --version`
     Then I see version information in format: "vibe version X.Y.Z"

AC3: Given vibe-dash is installed
     When I run `vibe` with no arguments
     Then TUI dashboard launches (placeholder message for now)
     And exit code is 0

AC4: Given vibe-dash is installed
     When I run with unknown flag (e.g., `vibe --invalid`)
     Then I see error message describing the unknown flag
     And exit code is 1

AC5: Given vibe-dash is installed
     When I run `vibe --verbose`
     Then info-level logging is enabled
     And normal operation continues

AC6: Given vibe-dash is installed
     When I run `vibe --debug`
     Then debug-level logging is enabled with file/line info
     And normal operation continues

AC7: Given vibe-dash is installed
     When I run `vibe --config /path/to/config.yaml`
     Then the specified config file is noted (actual loading in Story 1.7)
     And normal operation continues
```

## Tasks / Subtasks

- [x] **Task 1: Enhance root command with full description** (AC: 1)
  - [x] 1.1 Update `internal/adapters/cli/root.go` with full description
  - [x] 1.2 Set `Long` description: "CLI dashboard for vibe coding projects. Track AI-assisted coding project stages, detect when agents are waiting for input, and manage your workflow."
  - [x] 1.3 Add usage template showing available commands
  - [x] 1.4 Write test verifying help output contains expected text

- [x] **Task 2: Implement version command** (AC: 2)
  - [x] 2.1 Create `internal/adapters/cli/version.go`
  - [x] 2.2 Define version variables: Version, Commit, BuildDate (set via ldflags)
  - [x] 2.3 Add `--version` flag to root command
  - [x] 2.4 Output format: "vibe version X.Y.Z (commit: abc123, built: 2025-12-12)"
  - [x] 2.5 Update Makefile to inject version at build time via ldflags
  - [x] 2.6 Write test verifying version output format

- [x] **Task 3: Implement global flags** (AC: 5, 6, 7)
  - [x] 3.1 Create `internal/adapters/cli/flags.go` for global flag definitions
  - [x] 3.2 Add `--verbose` / `-v` flag (bool, default: false)
  - [x] 3.3 Add `--debug` flag (bool, default: false)
  - [x] 3.4 Add `--config` / `-c` flag (string, default: "")
  - [x] 3.5 Implement `initLogging()` function to configure slog based on flags
  - [x] 3.6 Call `initLogging()` in PersistentPreRun hook
  - [x] 3.7 Write tests for each flag behavior

- [x] **Task 4: Implement TUI placeholder** (AC: 3)
  - [x] 4.1 Add Run function to root command
  - [x] 4.2 Display placeholder message: "TUI dashboard coming soon. Press Ctrl+C to exit."
  - [x] 4.3 Wait for context cancellation (respects signal handling in main.go)
  - [x] 4.4 Exit with code 0 on clean shutdown
  - [x] 4.5 Write test verifying placeholder behavior

- [x] **Task 5: Implement exit codes** (AC: 4)
  - [x] 5.1 Create `internal/adapters/cli/exitcodes.go`
  - [x] 5.2 Define exit code constants matching domain errors:
    - ExitSuccess = 0
    - ExitGeneralError = 1
    - ExitProjectNotFound = 2
    - ExitConfigInvalid = 3
    - ExitDetectionFailed = 4
  - [x] 5.3 Create `MapErrorToExitCode(err error) int` function
  - [x] 5.4 Update main.go to use mapped exit codes
  - [x] 5.5 Write tests for error-to-exit-code mapping

- [x] **Task 6: Integration and validation** (AC: all)
  - [x] 6.1 Run `make build` and verify binary works
  - [x] 6.2 Test `./bin/vibe --help` shows expected output
  - [x] 6.3 Test `./bin/vibe --version` shows version info
  - [x] 6.4 Test `./bin/vibe --invalid` returns exit code 1
  - [x] 6.5 Run `make lint` to ensure code quality
  - [x] 6.6 Run `make test` to verify all tests pass

## Implementation Order (Recommended)

Execute tasks in this order to minimize rework and ensure dependencies are ready:

1. **Task 5: Exit codes** - Foundation for error handling, no dependencies
2. **Task 3: Global flags** - Foundation for logging configuration
3. **Task 2: Version command** - Simple, testable, depends on flags init
4. **Task 1: Root command enhancement** - Depends on flags being available
5. **Task 4: TUI placeholder** - Depends on context propagation
6. **Task 6: Integration** - Final validation of all components

**Test Priority:** Focus testing effort on:
1. Exit code mapping (highest risk for scripting users)
2. Flag behavior (--debug overrides --verbose)
3. Version output format (user-visible)

## Dev Notes

### Existing Codebase Context

The codebase already has these files that this story ENHANCES (not replaces):

| File | Existing Content | This Story Adds |
|------|------------------|-----------------|
| `cmd/vibe/main.go` | Signal handling, context cancellation, basic error exit | Use MapErrorToExitCode for domain errors |
| `internal/adapters/cli/root.go` | Basic RootCmd, Execute(ctx), Short/Long strings | Run function, global flags via init(), detailed Long description |
| `Makefile` | build, test, lint, fmt targets | ldflags for version injection |

**Important:** Review existing code before implementing. Enhance, don't replace.

### Architecture Compliance - CRITICAL

**CLI Adapter Location:**

```
internal/adapters/cli/
├── root.go         # Root command definition
├── version.go      # Version command and variables
├── flags.go        # Global flag definitions
├── exitcodes.go    # Domain error to exit code mapping
└── *_test.go       # Co-located tests
```

**Dependency Direction:**

```
cmd/vibe/main.go → internal/adapters/cli/ → github.com/spf13/cobra
                                          → internal/core/domain/ (for error types)
```

### Cobra Command Pattern

The existing `root.go` needs enhancement. Here's the target state:

```go
// root.go - enhance existing RootCmd
var RootCmd = &cobra.Command{
    Use:   "vibe",
    Short: "CLI dashboard for vibe coding projects",
    Long: `vibe-dash is a terminal dashboard for tracking AI-assisted coding projects.

Track project stages, detect when AI agents are waiting for input,
and manage your workflow across multiple projects.

Run 'vibe' with no arguments to launch the interactive dashboard.`,
    Run: func(cmd *cobra.Command, args []string) {
        // User-facing output goes to stdout (not slog)
        // slog output goes to stderr for diagnostics
        fmt.Println("TUI dashboard coming soon. Press Ctrl+C to exit.")
        <-cmd.Context().Done()
    },
}
```

**Note:** The `fmt.Println` here is intentional - it's user-facing output to stdout, not logging. Logging (slog) goes to stderr.

### Version Variables with ldflags

```go
// version.go - package cli

var (
    Version   = "dev"      // Set via ldflags: -X ...cli.Version=$(VERSION)
    Commit    = "unknown"  // Set via ldflags: -X ...cli.Commit=$(COMMIT)
    BuildDate = "unknown"  // Set via ldflags: -X ...cli.BuildDate=$(BUILD_DATE)
)

func init() {
    RootCmd.Version = Version
    RootCmd.SetVersionTemplate("vibe version {{.Version}} (commit: " + Commit + ", built: " + BuildDate + ")\n")
}
```

**Makefile ldflags:**

```makefile
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -ldflags "-X github.com/JeiKeiLim/vibe-dash/internal/adapters/cli.Version=$(VERSION) \
                    -X github.com/JeiKeiLim/vibe-dash/internal/adapters/cli.Commit=$(COMMIT) \
                    -X github.com/JeiKeiLim/vibe-dash/internal/adapters/cli.BuildDate=$(BUILD_DATE)"

build:
	CGO_ENABLED=1 go build $(LDFLAGS) -o bin/vibe ./cmd/vibe
```

### Global Flags Pattern

```go
// flags.go - package cli, imports: log/slog, os, github.com/spf13/cobra

var (
    verbose    bool
    debug      bool
    configFile string
)

func init() {
    RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
    RootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging with file/line info")
    RootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Config file path (default: ~/.vibe-dash/config.yaml)")

    RootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
        initLogging()
    }
}

func initLogging() {
    var level slog.Level
    var addSource bool

    switch {
    case debug:      // --debug takes precedence over --verbose
        level = slog.LevelDebug
        addSource = true
    case verbose:
        level = slog.LevelInfo
        addSource = false
    default:
        level = slog.LevelError
        addSource = false
    }

    opts := &slog.HandlerOptions{Level: level, AddSource: addSource}
    logger := slog.New(slog.NewTextHandler(os.Stderr, opts))  // stderr, not stdout
    slog.SetDefault(logger)
}
```

### Exit Code Mapping

```go
// exitcodes.go
package cli

import (
    "errors"

    "github.com/JeiKeiLim/vibe-dash/internal/core/domain"
)

const (
    ExitSuccess          = 0
    ExitGeneralError     = 1
    ExitProjectNotFound  = 2
    ExitConfigInvalid    = 3
    ExitDetectionFailed  = 4
)

// MapErrorToExitCode maps domain errors to CLI exit codes
// per Architecture "Error-to-Exit-Code Mapping"
//
// Exit code mapping (complete list):
//   - ErrProjectNotFound     → 2 (specific, recoverable)
//   - ErrConfigInvalid       → 3 (specific, user can fix config)
//   - ErrDetectionFailed     → 4 (specific, retry may help)
//   - ErrProjectAlreadyExists → 1 (general - user decision needed)
//   - ErrPathNotAccessible   → 1 (general - filesystem issue)
//   - ErrInvalidStage        → 1 (general - internal error)
//   - ErrInvalidConfidence   → 1 (general - internal error)
//   - Any other error        → 1 (general catch-all)
func MapErrorToExitCode(err error) int {
    if err == nil {
        return ExitSuccess
    }

    switch {
    case errors.Is(err, domain.ErrProjectNotFound):
        return ExitProjectNotFound
    case errors.Is(err, domain.ErrConfigInvalid):
        return ExitConfigInvalid
    case errors.Is(err, domain.ErrDetectionFailed):
        return ExitDetectionFailed
    // All other domain errors fall through to general error
    // This includes: ErrProjectAlreadyExists, ErrPathNotAccessible,
    // ErrInvalidStage, ErrInvalidConfidence
    default:
        return ExitGeneralError
    }
}
```

**Important:** The `errors.Is()` function traverses wrapped errors, so this works correctly with:
```go
fmt.Errorf("failed to find project %s: %w", name, domain.ErrProjectNotFound)
```

**Update main.go to use exit codes:**

```go
func main() {
    // ... signal handling ...

    if err := run(ctx); err != nil {
        slog.Error("application error", "error", err)
        os.Exit(cli.MapErrorToExitCode(err))
    }
}
```

### Testing Patterns

**Test help output:** (root_test.go)

```go
func TestRootCmd_Help(t *testing.T) {
    buf := new(bytes.Buffer)
    RootCmd.SetOut(buf)
    RootCmd.SetArgs([]string{"--help"})

    if err := RootCmd.Execute(); err != nil {
        t.Fatalf("Execute() error = %v", err)
    }

    output := buf.String()
    for _, phrase := range []string{"vibe coding projects", "--verbose", "--debug", "--config"} {
        if !strings.Contains(output, phrase) {
            t.Errorf("Help output missing %q", phrase)
        }
    }
}
```

**Test exit codes:**

```go
// exitcodes_test.go
func TestMapErrorToExitCode(t *testing.T) {
    tests := []struct {
        name     string
        err      error
        expected int
    }{
        {"nil error", nil, ExitSuccess},
        {"project not found", domain.ErrProjectNotFound, ExitProjectNotFound},
        // CRITICAL: Test wrapped errors - errors.Is() must traverse the chain
        {"wrapped project not found", fmt.Errorf("failed: %w", domain.ErrProjectNotFound), ExitProjectNotFound},
        {"deeply wrapped", fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", domain.ErrProjectNotFound)), ExitProjectNotFound},
        {"config invalid", domain.ErrConfigInvalid, ExitConfigInvalid},
        {"detection failed", domain.ErrDetectionFailed, ExitDetectionFailed},
        // Other domain errors → general error
        {"project already exists", domain.ErrProjectAlreadyExists, ExitGeneralError},
        {"path not accessible", domain.ErrPathNotAccessible, ExitGeneralError},
        {"unknown error", errors.New("unknown"), ExitGeneralError},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := MapErrorToExitCode(tt.err)
            if got != tt.expected {
                t.Errorf("MapErrorToExitCode(%v) = %d, want %d", tt.err, got, tt.expected)
            }
        })
    }
}
```

**Why wrapped error tests matter:** Services wrap domain errors with context (e.g., `fmt.Errorf("project %s: %w", name, domain.ErrProjectNotFound)`). The `errors.Is()` function must correctly identify the underlying error through the chain. Without these tests, exit code mapping could silently break.

### Previous Story Context (Story 1.3)

**Port interfaces available:**
- `ports.MethodDetector` - for future detection commands
- `ports.ProjectRepository` - for future project commands
- `ports.FileWatcher` - for future file watching
- `ports.ConfigLoader` - for future config loading

**Domain errors available for exit code mapping:**
- `domain.ErrProjectNotFound` → exit 2
- `domain.ErrProjectAlreadyExists` → exit 1
- `domain.ErrDetectionFailed` → exit 4
- `domain.ErrConfigInvalid` → exit 3
- `domain.ErrPathNotAccessible` → exit 1

### Context Propagation

The existing main.go already implements proper context propagation:

```go
ctx, cancel := context.WithCancel(context.Background())
// ... signal handling calls cancel() ...
cli.Execute(ctx)
```

Root command uses `ExecuteContext(ctx)` which propagates context to all subcommands via `cmd.Context()`.

### Files to Create/Modify

| File | Action | Purpose |
|------|--------|---------|
| `internal/adapters/cli/root.go` | Modify | Enhance with full description, Run function |
| `internal/adapters/cli/root_test.go` | Create | Tests for root command |
| `internal/adapters/cli/version.go` | Create | Version variables and template |
| `internal/adapters/cli/version_test.go` | Create | Tests for version output |
| `internal/adapters/cli/flags.go` | Create | Global flag definitions, logging init |
| `internal/adapters/cli/flags_test.go` | Create | Tests for flag behavior |
| `internal/adapters/cli/exitcodes.go` | Create | Exit code constants and mapping |
| `internal/adapters/cli/exitcodes_test.go` | Create | Tests for error mapping |
| `cmd/vibe/main.go` | Modify | Use mapped exit codes |
| `Makefile` | Modify | Add ldflags for version injection |

### DO NOT (Anti-Patterns)

| DO NOT | DO INSTEAD |
|--------|------------|
| Hard-code version string | Inject via ldflags at build time |
| Use fmt.Printf for logging | Use slog package (logs to stderr) |
| Exit with os.Exit in command handlers | Return errors, let main.go handle exit |
| Put business logic in CLI layer | CLI just parses args, calls services |
| Skip context propagation | Always use cmd.Context() |
| Log at multiple layers | Log once at handling site (main.go) |
| Use panic for errors | Return error and let caller handle |
| Mix stdout and stderr | User output → stdout, diagnostics → stderr |

**Critical:** Never use `panic()` in CLI command handlers. Panics bypass the graceful shutdown sequence and prevent proper exit code reporting. Always return errors and let main.go handle them.

### Project Structure Notes

- **Location:** `internal/adapters/cli/`
- **Package name:** `cli`
- **Naming:** Files named after functionality (root.go, version.go, flags.go)
- **Tests:** Co-located `*_test.go` files

### References

| Document | Section | Lines | Key Content |
|----------|---------|-------|-------------|
| architecture.md | CLI Framework | 169-174 | Cobra for commands, root launches TUI |
| architecture.md | Error Handling Strategy | 335-359 | Error-to-Exit-Code mapping table |
| architecture.md | Logging & Observability | 362-377 | slog levels, stderr output |
| architecture.md | Graceful Shutdown Pattern | 629-695 | Context cancellation sequence |
| architecture.md | Go Code Conventions | 419-468 | Context first, New* constructors |
| project-context.md | Exit Codes (CLI) | 129-138 | Exit code 0-4 definitions |
| epics.md | Story 1.4 | 326-362 | Original acceptance criteria |
| 1-3-port-interfaces.md | Domain errors | 393-410 | Available error types for mapping |

## Dev Agent Record

### Context Reference

Story context created from comprehensive analysis of:
- docs/epics.md (Story 1.4 requirements)
- docs/architecture.md (CLI framework, error handling, logging patterns)
- docs/prd.md (functional requirements FR60)
- docs/project-context.md (critical rules, exit codes)
- docs/sprint-artifacts/1-3-port-interfaces.md (previous story learnings)
- internal/adapters/cli/root.go (existing CLI implementation)
- cmd/vibe/main.go (existing entry point with signal handling)

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

None - story created in YOLO mode.

### Completion Notes List

- **Task 5 (Exit codes):** Implemented exit code constants and MapErrorToExitCode function using errors.Is() for proper wrapped error handling. Updated main.go to use mapped exit codes. Full test coverage including wrapped and deeply wrapped error scenarios.

- **Task 3 (Global flags):** Created flags.go with --verbose/-v, --debug, and --config/-c persistent flags. Implemented initLogging() with slog configuration based on flag values. Debug takes precedence over verbose. Logging outputs to stderr.

- **Task 2 (Version command):** Created version.go with Version, Commit, BuildDate variables. Updated Makefile with LDFLAGS for build-time injection via git describe/rev-parse. Version template shows format: "vibe version X.Y.Z (commit: abc123, built: timestamp)".

- **Task 1 (Root command):** Enhanced root.go with comprehensive Long description. Short description now reads "CLI dashboard for vibe coding projects". Help output includes all global flags.

- **Task 4 (TUI placeholder):** Implemented Run function that displays "TUI dashboard coming soon. Press Ctrl+C to exit." and waits for context cancellation via <-cmd.Context().Done().

- **Task 6 (Integration):** All validation passed: make build works, --help shows expected output, --version shows injected version, --invalid returns exit code 1, make lint passes, make test passes.

### File List

**New files:**
- internal/adapters/cli/exitcodes.go
- internal/adapters/cli/exitcodes_test.go
- internal/adapters/cli/flags.go
- internal/adapters/cli/flags_test.go
- internal/adapters/cli/version.go
- internal/adapters/cli/version_test.go
- internal/adapters/cli/root_test.go
- internal/adapters/cli/test_helpers_test.go (added during code review)

**Modified files:**
- internal/adapters/cli/root.go (enhanced description, Run function, startup log moved here)
- cmd/vibe/main.go (use MapErrorToExitCode, removed premature startup log)
- Makefile (added LDFLAGS for version injection)

### Code Review Fixes Applied

**Review Date:** 2025-12-12
**Reviewer:** Amelia (Dev Agent)

**Issues Fixed:**
1. **[HIGH] INFO log pollution** - Moved `slog.Info("vibe-dash starting")` from main.go to root.go's Run function so --help/--version output is clean
2. **[HIGH] Test coverage** - Improved from 50% to 93.5% by adding comprehensive tests
3. **[HIGH] GetConfigFile() untested** - Added test in flags_test.go
4. **[MEDIUM] Debug precedence test** - Added TestDebugPrecedenceOverVerbose
5. **[MEDIUM] slog state cleanup** - Added saveSlogDefault() and resetLoggingState() helpers
6. **[LOW] Flaky TUI test** - Replaced time.Sleep with polling loop
7. **[LOW] Test helper organization** - Created test_helpers_test.go for shared test utilities
