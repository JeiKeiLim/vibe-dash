# Project Context for AI Agents

_Critical rules and patterns for implementing vibe-dash. Focus on unobvious details that agents might otherwise miss._

---

## Technology Stack & Versions

| Technology | Version | Notes |
|------------|---------|-------|
| Go | 1.21+ | Required for slog stdlib |
| Bubble Tea | Latest | TUI framework (Elm architecture) |
| Cobra | Latest | CLI framework |
| Viper | Latest | Config cascade |
| sqlx | Latest | SQL with struct scanning |
| fsnotify | Latest | File watching |
| SQLite | Embedded | WAL mode required |
| golangci-lint | Latest | Must pass before PR |
| goimports | stdlib | Must run before commit |

---

## Critical Implementation Rules

### Hexagonal Architecture Boundaries (CRITICAL)

```
internal/core/  →  ZERO external imports allowed
                   Only stdlib + own domain/ports/services

internal/adapters/  →  Can import core + external libs
                       Implements port interfaces

cmd/  →  Wires everything together
```

**NEVER let core import from adapters. This is the #1 architecture violation to avoid.**

### Killer Feature Priority

**Agent Waiting Detection (FR34-38) is THE differentiator** - prioritize its implementation quality. This feature detects when an AI agent is waiting for user input (10-minute inactivity threshold). It must work reliably.

### Go Patterns (MUST FOLLOW)

1. **Context first** - All service methods: `func (s *Svc) Do(ctx context.Context, ...)`
2. **New* constructors** - `func NewDetectionService(...) *DetectionService`
3. **Error wrapping** - `fmt.Errorf("failed to scan %s: %w", path, err)`
4. **Domain errors** - Return `ErrProjectNotFound`, not raw errors
5. **Log once** - Log at handling site only, never during propagation

### Registry Pattern

**Services call `registry.DetectAll()`, never individual detectors directly.**

The detector registry (`internal/adapters/detectors/registry.go`) is the only component that knows about all detector implementations.

### Configuration Cascade

**Priority order (highest to lowest):**
```
1. CLI flags           --hibernation-days=7
2. Project config      ~/.vibe-dash/<project>/config.yaml
3. Master config       ~/.vibe-dash/config.yaml
4. Built-in defaults   hibernation_days: 14
```

Viper merges these automatically. CLI flags always win.

### Naming Conventions

| Context | Convention | Example |
|---------|------------|---------|
| Go code | PascalCase/camelCase | `DetectionService`, `parseConfig` |
| DB tables | snake_case, plural | `projects`, `detection_states` |
| DB columns | snake_case | `project_id`, `last_scanned_at` |
| JSON/YAML keys | snake_case | `hibernation_days` |
| Test fixtures | `{method}-stage-{stage}` | `speckit-stage-plan` |

### Testing Rules

1. **Co-locate tests** - `detector_test.go` next to `detector.go`
2. **Table-driven** - Use `tests []struct{...}` pattern
3. **Build tags for integration** - `//go:build integration`
4. **Mocks in test file** - Extract to `*_mock.go` only if reused
5. **95% detection accuracy** - Launch blocker, CI must enforce

### Detection Accuracy Measurement

**Launch blocker formula:**
```
accuracy = correct_detections / total_fixtures * 100

19/20 = 95% ✅ Pass
18/20 = 90% ❌ Blocked - cannot ship
```

Test fixtures in `test/fixtures/` are the ground truth oracle.

### SQLite Rules

1. **WAL mode required** - Enable on connection open
2. **Lazy connections** - Open per-operation, close after
3. **Migrations** - `schema_version` table, sequential v1/v2/v3

### File Watching

1. **5-10 second debounce** - Don't react to every event
2. **OS abstraction** - Use `internal/adapters/filesystem/platform.go` interface
3. **MVP: Linux/macOS only** - Windows is post-MVP

---

## Anti-Patterns (NEVER DO)

| Don't | Do Instead |
|-------|------------|
| Import adapters from core | Keep core dependency-free |
| Call detectors directly | Use `registry.DetectAll()` |
| `userId` in JSON | `user_id` |
| `Users` table name | `users` |
| Skip `ctx context.Context` | Always first param |
| Log errors at every layer | Log once at handling site |
| Tests in `/tests` folder | Co-locate with source |
| Hardcode paths | Use OS abstraction layer |
| Keep SQLite connections open | Open-use-close pattern |

---

## Exit Codes (CLI)

| Code | Meaning | Domain Error |
|------|---------|--------------|
| 0 | Success | - |
| 1 | General error | Any unhandled |
| 2 | Project not found | `ErrProjectNotFound` |
| 3 | Invalid config | `ErrConfigInvalid` |
| 4 | Detection failure | `ErrDetectionFailed` |

---

## Quick Reference

**Run before commit:**
```bash
make fmt           # goimports
make lint          # golangci-lint
make test          # unit tests
make test-all      # unit + integration tests
make test-accuracy # detection accuracy (must be ≥95%)
```

**Key files:**
- Architecture: `docs/architecture.md`
- PRD: `docs/prd.md`
- Domain entities: `internal/core/domain/`
- Port interfaces: `internal/core/ports/`
- Services: `internal/core/services/`

**First user story:**
> "As a developer, I can run `vibe` and see an empty dashboard"

---

## Post-MVP References

### BMAD Progress Dashboard
**URL:** https://github.com/ibadmore/bmad-progress-dashboard

An existing implementation of a BMAD progress dashboard. Consider adopting this approach or integrating lessons learned during post-MVP phase for sprint/workflow visualization features.
