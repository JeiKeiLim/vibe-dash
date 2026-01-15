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

## Story Completion & User Verification (MANDATORY)

**Lesson learned from Epics 3, 3.5, 4, 4.5:** Code was written but not properly wired/integrated. All stories now require user verification before marking `done`.

### Dev Agent Rules

1. **Never mark story `done` directly** - Always set to `review` first
2. **Stop after `review`** - Wait for user to verify integration
3. **Include User Testing Guide** - Every story with TUI/CLI changes must have one

### Story Status Flow

```
drafted → ready-for-dev → in-progress → review → done
                                          ↑        ↑
                                     Dev stops   User marks
```

### User Testing Guide Template

Every story that touches TUI, CLI, or integration must include:

```markdown
## User Testing Guide

**Time needed:** X minutes

### Step 1: Basic Check
- Command to run
- What to look for (table with Expected column)
- Red flags (what means FAIL)

### Step 2: Spot-Check (if applicable)
- One specific scenario to verify
- Copy-paste commands
- Expected output

### Decision Guide
| Situation | Action |
|-----------|--------|
| All checks pass | Mark `done` |
| Any check fails | Do NOT approve, document issue |
```

### What User Checks

| Story Type | User Verifies |
|------------|---------------|
| TUI feature | Run `./bin/vibe`, visually confirm feature works |
| CLI command | Run command, check output format |
| Detection logic | Run vibe on test project, verify stage/reasoning |
| Integration | End-to-end flow works, not just unit tests pass |

### User-Visible Changes Requirement

Every story MUST include a `## User-Visible Changes` section that describes:
- What users will see/experience differently (New/Changed/Removed)
- Or explicitly "None - [reason]" for internal changes

This section:
- Enables quick release note generation
- Helps reviewers focus on user impact
- Documents historical changes for future developers

**Reference example:** See the `## User-Visible Changes` section in Story 9.5-6 (`docs/sprint-artifacts/stories/epic-9.5/9-5-6-user-visible-changes-section.md`) for a correctly formatted section.

---

## Phase 2 Additions

### Agent Detection Interface

**New Port:**
- Location: `internal/core/ports/agent_activity_detector.go`
- Interface: `AgentActivityDetector` with `Detect(ctx, projectPath) (AgentState, error)`

**New Adapters:**
- `internal/adapters/agentdetectors/claude_code.go` - JSONL log parsing (high confidence)
- `internal/adapters/agentdetectors/generic.go` - File activity fallback (10-min threshold)

**Detection Logic:**
```
assistant + stop_reason: "end_turn"  → WAITING FOR USER
assistant + stop_reason: "tool_use"  → WORKING
```

**Log Location:** `~/.claude/projects/{project-hash}/*.jsonl`

### Methodology Coexistence

**Registry Enhancement:**
- New method: `DetectWithCoexistence(ctx, path) ([]*DetectionResult, error)`
- Runs ALL detectors (not first-match-wins)
- Compares artifact timestamps
- Returns single result if >1 hour difference, otherwise returns all with warning

**Timestamp Sources:**
- Speckit: mtime of most recent spec folder
- BMAD: mtime of `sprint-status.yaml`, `config.yaml`, or `implementation-artifacts/`

### Progress Metrics (Experimental - Tier 2)

**Isolation Principle:** Metrics feature is cleanly removable if unproven.

**Separate Database:** `~/.vibe-dash/metrics.db` (not in `state.db`)

**New Packages:**
- `internal/adapters/metrics/` - Event collection
- `internal/adapters/persistence/metrics/` - metrics.db access
- `internal/adapters/tui/statsview/` - Dedicated TUI view

**Schema Addition:**
```sql
CREATE TABLE stage_transitions (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    from_stage TEXT NOT NULL,
    to_stage TEXT NOT NULL,
    transitioned_at TEXT NOT NULL,
    FOREIGN KEY(project_id) REFERENCES projects(id)
);
```

**TUI Entry:** `'s'` key from Dashboard → Stats View → `Esc`/`'q'` back

### Phase 2 NFRs

| Requirement | Target |
|-------------|--------|
| Agent detection latency | < 1 second |
| Claude Code log parsing | Tail-optimized (last N entries) |
| Stats View render | < 500ms for 1 year of data |
| Metrics DB growth | < 20MB/year |

---

## Post-MVP References

### BMAD Progress Dashboard
**URL:** https://github.com/ibadmore/bmad-progress-dashboard

An existing implementation of a BMAD progress dashboard. Consider adopting this approach or integrating lessons learned during post-MVP phase for sprint/workflow visualization features.
