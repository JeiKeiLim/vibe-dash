# bmad-test - Epic Breakdown

**Author:** Jongkuk Lim
**Date:** 2025-12-11
**Project:** Vibe Dashboard - CLI-first workflow state tracker for vibe coding projects

---

## Overview

This document provides the complete epic and story breakdown for Vibe Dashboard, decomposing the requirements from the [PRD](./prd.md) into implementable stories with full technical context from [Architecture](./architecture.md).

**Living Document Notice:** Stories include complete acceptance criteria with Architecture references for implementation-ready development.

---

## Functional Requirements Inventory

### Domain 1: Project Management (8 FRs)

| FR | Description |
|----|-------------|
| FR1 | Add project from current directory (`vibe add .`) |
| FR2 | Add project from specified path (`vibe add <path>`) |
| FR3 | View list of all tracked projects |
| FR4 | Remove project from tracking |
| FR5 | Set custom display name (nickname) for project |
| FR6 | Detect and resolve project name collisions |
| FR7 | Validate project paths at launch, detect missing directories |
| FR8 | Choose action when project path missing (Delete/Move/Keep) |

### Domain 2: Workflow Detection (6 FRs)

| FR | Description |
|----|-------------|
| FR9 | Detect Speckit methodology from project artifacts |
| FR10 | Identify current Speckit stage (Specify/Plan/Tasks/Implement) |
| FR11 | Show detection reasoning when stage identified |
| FR12 | Indicate uncertainty when stage detection unclear |
| FR13 | Support pluggable methodology detectors (MethodDetector interface) |
| FR14 | Detect multiple methodologies in same project |

### Domain 3: Dashboard Visualization (13 FRs)

| FR | Description |
|----|-------------|
| FR15 | View real-time dashboard of active projects in terminal UI |
| FR16 | See project name/nickname, stage, last modified timestamp |
| FR17 | Visual indicators (✨ recent, ⚡ active, 🤷 uncertain, ⏸️ waiting) |
| FR18 | Keyboard shortcuts [a/h/r/d/?/q] |
| FR19 | Vim-style keys [j/k/h/l] |
| FR20 | View detailed information for selected project |
| FR21 | Add/edit notes (memo) for a project |
| FR22 | View project notes in dashboard detail view |
| FR23 | Manual refresh forces artifact re-scan |
| FR24 | See count of active vs hibernated projects |
| FR25 | View hibernated projects list |
| FR26 | Display detection reasoning for current stage |
| FR27 | Automatically detect file system changes |

### Domain 4: Project State Management (6 FRs)

| FR | Description |
|----|-------------|
| FR28 | Auto-mark projects hibernated after configurable days |
| FR29 | Auto-mark projects active when file changes detected |
| FR30 | Manually mark project as favorite (always visible) |
| FR31 | Remove favorite status from project |
| FR32 | Configure hibernation threshold (global or per-project) |
| FR33 | Distinguish active/hibernated/favorite project states |

### Domain 5: Agent Monitoring - KILLER FEATURE (5 FRs)

| FR | Description |
|----|-------------|
| FR34 | Detect when AI agent waiting for user input (inactivity threshold) |
| FR35 | Display ⏸️ WAITING visual indicator |
| FR36 | Show elapsed time since agent started waiting |
| FR37 | Configure agent waiting threshold (minutes) |
| FR38 | Clear waiting state when activity resumes |

### Domain 6: Configuration Management (9 FRs)

| FR | Description |
|----|-------------|
| FR39 | Store project paths in master config (`~/.vibe-dash/config.yaml`) |
| FR40 | Store project-specific settings in project config files |
| FR41 | Override config values using CLI flags |
| FR42 | Modify global config by editing master config file |
| FR43 | Modify project config by editing project config files |
| FR44 | Auto-create default config on first project add |
| FR45 | Use canonical paths to handle symlinks correctly |
| FR46 | Configure global settings (hibernation, refresh, waiting threshold) |
| FR47 | Configure per-project settings overriding global defaults |

### Domain 7: Scripting & Automation (14 FRs)

| FR | Description |
|----|-------------|
| FR48 | List projects in plain text format |
| FR49 | JSON output format with API versioning |
| FR50 | Get specific project status non-interactively |
| FR51 | Add projects with automatic conflict resolution |
| FR52 | Force automatic conflict resolution with --force flag |
| FR53 | Remove projects via CLI |
| FR54 | Mark/unmark favorites via CLI |
| FR55 | Set/edit project notes via CLI |
| FR56 | Rename projects (set display name) via CLI |
| FR57 | Manually hibernate or activate projects via CLI |
| FR58 | Trigger manual refresh via CLI |
| FR59 | Check if project exists via CLI |
| FR60 | Return standard exit codes (0/1/2/3/4) |
| FR61 | Shell completion (Bash/Zsh/Fish) |

### Domain 8: Error Handling & User Feedback (5 FRs)

| FR | Description |
|----|-------------|
| FR62 | Gracefully handle file watching failures (fallback to manual refresh) |
| FR63 | Detect and report config syntax errors |
| FR64 | Recover from corrupted state databases by reinitializing |
| FR65 | View keyboard shortcut help |
| FR66 | Display progress indicators during long operations |

**Total: 66 Functional Requirements**

---

## Epic Structure Overview

| Epic | Title | User Value | Primary FRs |
|------|-------|------------|-------------|
| 1 | Foundation & First Launch | Install and see empty dashboard works | Infrastructure |
| 2 | Project Management & Detection | Add projects, see stages detected | FR1-14, FR39-45 |
| 3 | Dashboard Visualization | See projects with TUI navigation | FR15-27 |
| 4 | Agent Waiting Detection | Never forget waiting agents | FR34-38 |
| 5 | Project State & Hibernation | Auto-manage active/dormant projects | FR28-33 |
| 6 | Scripting & Automation | Script with JSON output | FR48-61 |
| 7 | Error Handling & Polish | Graceful errors, helpful feedback | FR62-66 |

### Dependency Graph

```
Epic 1 (Foundation)
    │
    ▼
Epic 2 (Project + Detection)
    │
    ├─────────────────┬───────────────────┐
    ▼                 ▼                   ▼
Epic 3 (TUI)    Epic 4 (WAITING)    Epic 5 (State)
    │                 │                   │
    └─────────────────┴───────────────────┘
                      │
                      ▼
              Epic 6 (Scripting)
                      │
                      ▼
              Epic 7 (Polish)
```

---

## Epic 1: Foundation & First Launch

**Goal:** Establish project infrastructure and validate the entire stack works end-to-end with a minimal vertical slice.

**User Value:** "I can install vibe-dash, run `vibe`, and see an empty dashboard - proving the tool is installed correctly."

**Technical Context:**
- Hexagonal architecture setup (Architecture Section: Project Structure)
- Go 1.21+ with Cobra CLI framework
- Bubble Tea TUI initialization
- Centralized storage at `~/.vibe-dash/`

**UX Context:**
- EmptyView welcome screen (UX: Component Strategy)
- Basic Lipgloss styling foundation

---

### Story 1.1: Project Scaffolding

**As a** developer,
**I want** the project structure initialized with all directories and configuration files,
**So that** I have a solid foundation to build upon.

**Acceptance Criteria:**

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

**Technical Notes:**
- Follow Architecture section "Complete Project Directory Structure"
- Use `go mod init github.com/JeiKeiLim/vibe-dash`
- golangci-lint configuration per Architecture "Build & Distribution"

**Prerequisites:** None (first story)

---

### Story 1.2: Domain Entities

**As a** developer,
**I want** core domain entities defined with zero external dependencies,
**So that** the domain layer is pure and testable.

**Acceptance Criteria:**

```gherkin
Given I need to model the core domain
When I create domain entities in internal/core/domain/
Then Project entity exists with fields:
  - ID (string)
  - Name (string)
  - Path (string, canonical)
  - DisplayName (string, optional nickname)
  - DetectedMethod (string)
  - CurrentStage (Stage)
  - IsFavorite (bool)
  - State (ProjectState: Active/Hibernated)
  - Notes (string)
  - LastActivityAt (time.Time)
  - CreatedAt (time.Time)
  - UpdatedAt (time.Time)

And Stage enum exists with values:
  - StageUnknown
  - StageSpecify
  - StagePlan
  - StageTasks
  - StageImplement

And DetectionResult value object exists with:
  - Method (string)
  - Stage (Stage)
  - Confidence (Confidence: Certain/Likely/Uncertain)
  - Reasoning (string)

And Confidence enum exists with:
  - ConfidenceCertain
  - ConfidenceLikely
  - ConfidenceUncertain

And domain errors exist in errors.go:
  - ErrProjectNotFound
  - ErrProjectAlreadyExists
  - ErrDetectionFailed
  - ErrConfigInvalid
  - ErrPathNotAccessible

And all entities have no external imports (only stdlib)
```

**Technical Notes:**
- Follow Architecture "Go Code Conventions"
- Stage.String() returns human-readable names
- Confidence determines 🤷 indicator display

**Prerequisites:** Story 1.1

---

### Story 1.3: Port Interfaces

**As a** developer,
**I want** port interfaces defined for all external dependencies,
**So that** adapters can be injected and the core remains testable.

**Acceptance Criteria:**

```gherkin
Given I need to define boundaries between core and adapters
When I create interfaces in internal/core/ports/
Then MethodDetector interface exists:
  - Name() string
  - CanDetect(ctx context.Context, path string) bool
  - Detect(ctx context.Context, path string) (*DetectionResult, error)

And ProjectRepository interface exists:
  - Save(ctx context.Context, project *Project) error
  - FindByID(ctx context.Context, id string) (*Project, error)
  - FindByPath(ctx context.Context, path string) (*Project, error)
  - FindAll(ctx context.Context) ([]*Project, error)
  - FindActive(ctx context.Context) ([]*Project, error)
  - FindHibernated(ctx context.Context) ([]*Project, error)
  - Delete(ctx context.Context, id string) error
  - UpdateState(ctx context.Context, id string, state ProjectState) error

And FileWatcher interface exists:
  - Watch(ctx context.Context, paths []string) (<-chan FileEvent, error)
  - Close() error

And ConfigLoader interface exists:
  - Load() (*Config, error)
  - Save(config *Config) error

And all interfaces accept context.Context as first parameter
```

**Technical Notes:**
- Follow Architecture "Context Propagation" pattern
- Interfaces enable mock injection for testing
- FileEvent contains Path, Operation, Timestamp

**Prerequisites:** Story 1.2

---

### Story 1.4: Cobra CLI Framework

**As a** user,
**I want** to run `vibe` command and have it recognized,
**So that** the CLI entry point works.

**Acceptance Criteria:**

```gherkin
Given vibe-dash is installed
When I run `vibe --help`
Then I see help text with:
  - Description: "CLI dashboard for vibe coding projects"
  - Available commands placeholder
  - Global flags: --verbose, --debug, --config

When I run `vibe --version`
Then I see version information

When I run `vibe` with no arguments
Then TUI dashboard launches (placeholder for now)

When I run with unknown flag
Then I see error message and exit code 1
```

**Technical Notes:**
- Use Cobra library for CLI framework
- Root command in internal/adapters/cli/root.go
- Version set at build time via ldflags
- Follow Architecture "CLI Framework" patterns

**Prerequisites:** Story 1.1

---

### Story 1.5: Bubble Tea TUI Shell

**As a** user,
**I want** `vibe` to launch a terminal UI,
**So that** I can see the dashboard interface.

**Acceptance Criteria:**

```gherkin
Given vibe-dash is running
When I run `vibe` command
Then Bubble Tea TUI launches in alternate screen buffer
And I see the EmptyView welcome screen:
  """
  ┌─ VIBE DASHBOARD ─────────────────────┐
  │                                      │
  │   Welcome to Vibe Dashboard! 🎯      │
  │                                      │
  │   Add your first project:            │
  │   $ vibe add /path/to/project        │
  │                                      │
  │   Or from a project directory:       │
  │   $ cd my-project && vibe add .      │
  │                                      │
  │   Press [?] for help, [q] to quit    │
  │                                      │
  └──────────────────────────────────────┘
  """

When I press 'q'
Then TUI exits cleanly
And terminal is restored to previous state
And exit code is 0

When I press '?'
Then help overlay displays available shortcuts

When I resize terminal
Then layout adapts without crash
```

**Technical Notes:**
- Use Bubble Tea with Elm architecture (Model → Update → View)
- Use alternate screen buffer for clean exit
- Follow UX "EmptyView" component specification
- Implement graceful shutdown pattern from Architecture

**Prerequisites:** Story 1.4

---

### Story 1.6: Lipgloss Styles Foundation

**As a** developer,
**I want** centralized Lipgloss styles defined,
**So that** all TUI components render consistently.

**Acceptance Criteria:**

```gherkin
Given I need consistent styling across the TUI
When I create styles.go in internal/adapters/tui/
Then the following styles are defined:

  SelectedStyle:
    - Cyan background (color 6)
    - Used for currently selected row

  WaitingStyle:
    - Bold + Red foreground (color 1)
    - Reserved ONLY for ⏸️ WAITING state

  RecentStyle:
    - Green foreground (color 2)
    - Used for ✨ today indicator

  ActiveStyle:
    - Yellow foreground (color 3)
    - Used for ⚡ this week indicator

  UncertainStyle:
    - Dim/Faint gray (color 8)
    - Used for 🤷 uncertain state

  FavoriteStyle:
    - Magenta foreground (color 5)
    - Used for ⭐ favorite indicator

  DimStyle:
    - Faint modifier
    - Used for hints and secondary info

  BorderStyle:
    - Normal border (square corners)
    - Used for panel boundaries

And NO_COLOR environment variable is respected
And styles work on both dark and light terminal themes
```

**Technical Notes:**
- Follow UX "Design System Foundation" color palette
- Use 16-color ANSI palette for compatibility
- Implement NO_COLOR check per UX Accessibility

**Prerequisites:** Story 1.5

---

### Story 1.7: Configuration Auto-Creation

**As a** user,
**I want** configuration to auto-create on first run,
**So that** I don't need manual setup.

**Acceptance Criteria:**

```gherkin
Given I run vibe for the first time
When ~/.vibe-dash/ directory doesn't exist
Then it is created automatically
And ~/.vibe-dash/config.yaml is created with defaults:
  """yaml
  settings:
    hibernation_days: 14
    refresh_interval_seconds: 10
    refresh_debounce_ms: 200
    agent_waiting_threshold_minutes: 10
  projects: {}
  """

When I run vibe again
Then existing config is preserved
And no duplicate creation occurs

When config.yaml has syntax errors
Then error is reported with line number
And application continues with defaults
And exit code is 0 (degraded operation, not failure)
```

**Technical Notes:**
- Use Viper for YAML parsing
- Follow Architecture "Configuration System" patterns
- Config path: `~/.vibe-dash/config.yaml`
- Handle cross-platform home directory resolution

**Prerequisites:** Story 1.4

---

**Epic 1 Complete**

**Stories Created:** 7
**FR Coverage:** Foundation (no specific FRs - infrastructure)
**Technical Context Used:** Architecture sections 1-4
**UX Patterns Incorporated:** EmptyView, Lipgloss styles, NO_COLOR support

---

## Epic 2: Project Management & Detection

**Goal:** Enable users to add projects and automatically detect their methodology stage.

**User Value:** "I can add my projects with `vibe add .` and immediately see what stage they're at in my methodology workflow."

**Technical Context:**
- Speckit detector implementation (Architecture: detectors/speckit/)
- SQLite repository (Architecture: persistence/sqlite/)
- Canonical path resolution (Architecture: filesystem/)
- Detection service orchestration

**UX Context:**
- Detection reasoning in detail panel
- Trust-building through accuracy
- Honest uncertainty with 🤷 indicator

**FRs Covered:** FR1-8 (Project Management), FR9-14 (Workflow Detection), FR39-45 (Configuration)

---

### Story 2.1: SQLite Repository Setup

**As a** developer,
**I want** SQLite persistence for project data,
**So that** project state survives between sessions.

**Acceptance Criteria:**

```gherkin
Given I need to persist project data
When SQLite repository is initialized
Then database is created at ~/.vibe-dash/projects.db
And schema includes projects table:
  """sql
  CREATE TABLE projects (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      path TEXT NOT NULL UNIQUE,
      display_name TEXT,
      detected_method TEXT,
      current_stage TEXT,
      confidence TEXT,
      detection_reasoning TEXT,
      is_favorite INTEGER DEFAULT 0,
      state TEXT DEFAULT 'active',
      notes TEXT,
      last_activity_at TEXT NOT NULL,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL
  );
  """
And schema_version table tracks migrations
And WAL mode is enabled for concurrent access

When repository.Save() is called with valid project
Then project is persisted to database
And timestamps are set correctly

When repository.FindByPath() is called
Then project is retrieved by canonical path
And all fields are hydrated correctly

When database file is corrupted
Then error is returned with recovery suggestion
```

**Technical Notes:**
- Use sqlx for struct scanning (Architecture: Data Architecture)
- Enable WAL mode: `PRAGMA journal_mode=WAL`
- Use ISO 8601 UTC for timestamps
- Follow Architecture "Database Naming Conventions"

**Prerequisites:** Story 1.3

---

### Story 2.2: Path Resolution Utilities

**As a** developer,
**I want** canonical path resolution,
**So that** symlinks and relative paths are handled correctly.

**Acceptance Criteria:**

```gherkin
Given I need to handle various path formats
When resolving "/home/user/project"
Then canonical absolute path is returned

When resolving "." from /home/user/project
Then "/home/user/project" is returned

When resolving "~/project"
Then home directory is expanded correctly

When resolving symlink "/home/user/link -> /home/user/actual"
Then canonical path "/home/user/actual" is returned

When resolving non-existent path
Then ErrPathNotAccessible is returned

When same physical location is added twice (via symlink)
Then collision is detected using canonical path comparison
```

**Technical Notes:**
- Use `filepath.EvalSymlinks()` for symlink resolution
- Use `filepath.Abs()` for absolute paths
- Handle cross-platform home directory (`~`)
- Implement in internal/adapters/filesystem/paths.go

**Prerequisites:** Story 1.3

---

### Story 2.3: Add Project Command

**As a** user,
**I want** to add projects using `vibe add`,
**So that** they appear in my dashboard.

**Acceptance Criteria:**

```gherkin
Given I am in a project directory
When I run `vibe add .`
Then project is added with:
  - Name derived from directory name
  - Canonical path stored
  - Methodology detection attempted
  - Stage detected if methodology found
And confirmation message shows:
  "✓ Added: project-name"
  "  Method: Speckit"
  "  Stage: Plan (plan.md found)"

When I run `vibe add /path/to/project`
Then project at specified path is added

When I run `vibe add .` on already-tracked project
Then error message shows:
  "✗ Project already tracked: project-name"
And exit code is 1

When path doesn't exist
Then error message shows:
  "✗ Path not found: /invalid/path"
And exit code is 2

When I run `vibe add . --name "Custom Name"`
Then project is added with custom display name
```

**Technical Notes:**
- Implement in internal/adapters/cli/add.go
- FR1: Add from current directory
- FR2: Add from specified path
- FR5: Custom display name
- FR44: Auto-create config on first add

**Prerequisites:** Story 2.1, Story 2.2

---

### Story 2.4: Speckit Detector Implementation

**As a** system,
**I want** to detect Speckit methodology and stage,
**So that** users see accurate workflow state.

**Acceptance Criteria:**

```gherkin
Given a project directory
When checking for Speckit methodology
Then detector looks for:
  - specs/ directory
  - .speckit/ directory
  - .specify/ directory

When Speckit is detected
Then CanDetect() returns true
And Detect() analyzes stage based on artifacts:

  Given specs/NNN-feature/ contains only spec.md
  Then Stage = StageSpecify
  And Confidence = ConfidenceCertain
  And Reasoning = "spec.md exists, no plan.md"

  Given specs/NNN-feature/ contains spec.md + plan.md
  Then Stage = StagePlan
  And Confidence = ConfidenceCertain
  And Reasoning = "plan.md exists, no tasks.md"

  Given specs/NNN-feature/ contains spec.md + plan.md + tasks.md
  Then Stage = StageTasks
  And Confidence = ConfidenceCertain
  And Reasoning = "tasks.md exists"

  Given specs/NNN-feature/ contains implement.md or implementation started
  Then Stage = StageImplement
  And Confidence = ConfidenceCertain

When multiple spec directories exist
Then most recently modified is used for stage

When artifacts are ambiguous
Then Confidence = ConfidenceUncertain
And Reasoning explains the ambiguity

When no Speckit markers found
Then CanDetect() returns false
```

**Technical Notes:**
- Implement MethodDetector interface
- FR9: Detect Speckit methodology
- FR10: Identify current stage
- FR11: Show detection reasoning
- FR12: Indicate uncertainty
- Target 95% accuracy (Architecture: launch blocker)

**Prerequisites:** Story 1.3

---

### Story 2.5: Detection Service

**As a** system,
**I want** a detection service that orchestrates detectors,
**So that** multiple methodologies can be supported.

**Acceptance Criteria:**

```gherkin
Given DetectionService with registered detectors
When Detect(path) is called
Then each detector's CanDetect() is called in order
And first matching detector's Detect() is invoked
And DetectionResult is returned

When multiple detectors match (future: BMAD + Speckit)
Then both results are returned
And FR14 is satisfied (multiple methodologies)

When no detector matches
Then DetectionResult with:
  - Method = "unknown"
  - Stage = StageUnknown
  - Confidence = ConfidenceUncertain
  - Reasoning = "No methodology markers found"

When detector throws error
Then error is logged
And next detector is tried
And partial results returned if any succeed
```

**Technical Notes:**
- Implement in internal/core/services/detection_service.go
- Use detector registry pattern
- FR13: Pluggable methodology detectors
- FR14: Detect multiple methodologies

**Prerequisites:** Story 2.4

---

### Story 2.6: Project Name Collision Handling

**As a** user,
**I want** project name collisions handled automatically,
**So that** I can track multiple projects with the same directory name.

**Acceptance Criteria:**

```gherkin
Given project "api-service" exists at /client-a/api-service
When I add /client-b/api-service
Then collision is detected
And user is prompted:
  "Project name 'api-service' already exists."
  "Suggestions:"
  "  1. client-b-api-service"
  "  2. Enter custom name"
  "Choose [1/2]: "

When user selects 1
Then project added as "client-b-api-service"

When user enters custom name
Then project added with that name

When --force flag is used
Then automatic disambiguation applied without prompt:
  - Add parent directory prefix
  - If still collision, add grandparent
  - Continue until unique

When displaying projects
Then display_name is shown if set, otherwise name
```

**Technical Notes:**
- FR6: Detect and resolve collisions
- Use parent directory disambiguation algorithm from PRD
- Store both name (derived) and display_name (user-set)

**Prerequisites:** Story 2.3

---

### Story 2.7: List Projects Command

**As a** user,
**I want** to list tracked projects from CLI,
**So that** I can see what's being tracked.

**Acceptance Criteria:**

```gherkin
Given projects are tracked
When I run `vibe list`
Then I see plain text output:
  """
  client-alpha     Plan      5m ago
  client-bravo     Tasks     2h ago
  client-charlie   Implement 3d ago
  """

When I run `vibe list --json`
Then I see JSON output:
  """json
  {
    "api_version": "v1",
    "projects": [
      {
        "name": "client-alpha",
        "display_name": null,
        "path": "/home/user/client-alpha",
        "method": "speckit",
        "stage": "plan",
        "confidence": "certain",
        "state": "active",
        "is_favorite": false,
        "last_activity_at": "2025-12-11T10:30:00Z"
      }
    ]
  }
  """

When no projects exist
Then plain text shows: "No projects tracked. Run 'vibe add .' to add one."
And JSON shows: {"api_version": "v1", "projects": []}
And exit code is 0
```

**Technical Notes:**
- FR3: View list of tracked projects
- FR48: Plain text format
- FR49: JSON output with API versioning
- Follow Architecture "JSON/YAML Format Conventions"

**Prerequisites:** Story 2.3

---

### Story 2.8: Remove Project Command

**As a** user,
**I want** to remove projects from tracking,
**So that** I can clean up my dashboard.

**Acceptance Criteria:**

```gherkin
Given project "client-alpha" is tracked
When I run `vibe remove client-alpha`
Then confirmation prompt shows:
  "Remove 'client-alpha' from tracking? [y/n]"

When I confirm with 'y'
Then project is removed from database
And message shows: "✓ Removed: client-alpha"
And exit code is 0

When I cancel with 'n'
Then project remains tracked
And message shows: "Cancelled"
And exit code is 0

When I run `vibe remove client-alpha --force`
Then no confirmation prompt
And project is removed immediately

When project doesn't exist
Then error shows: "✗ Project not found: client-alpha"
And exit code is 2
```

**Technical Notes:**
- FR4: Remove project from tracking
- FR53: Remove via CLI
- Implement in internal/adapters/cli/remove.go

**Prerequisites:** Story 2.3

---

### Story 2.9: Path Validation at Launch

**As a** user,
**I want** missing project paths detected at launch,
**So that** I can handle moved or deleted projects.

**Acceptance Criteria:**

```gherkin
Given project "client-alpha" tracked at /old/path
When /old/path no longer exists
And I launch `vibe`
Then TUI shows path validation dialog:
  """
  ⚠️ Project path not found: client-alpha
  /old/path

  [D] Delete - Remove from dashboard
  [M] Move - Update to current directory
  [K] Keep - Maybe network mount, keep tracking
  """

When I press 'D'
Then project is removed
And dashboard loads without it

When I press 'M' from /new/path
Then project path updated to /new/path
And detection re-runs

When I press 'K'
Then project kept with warning indicator
And dashboard loads

When all paths valid
Then no dialog shown
And dashboard loads normally
```

**Technical Notes:**
- FR7: Validate paths at launch
- FR8: Choose action for missing paths
- Only check at launch, not during runtime (per Architecture)

**Prerequisites:** Story 2.3, Story 1.5

---

### Story 2.10: Golden Path Test Fixtures

**As a** developer,
**I want** test fixtures for detection validation,
**So that** I can verify 95% accuracy.

**Acceptance Criteria:**

```gherkin
Given test/fixtures/ directory
When fixtures are created
Then the following exist:

  speckit-stage-specify/
    └── specs/001-feature/
        └── spec.md
  Expected: Stage=Specify, Confidence=Certain

  speckit-stage-plan/
    └── specs/001-feature/
        ├── spec.md
        └── plan.md
  Expected: Stage=Plan, Confidence=Certain

  speckit-stage-tasks/
    └── specs/001-feature/
        ├── spec.md
        ├── plan.md
        └── tasks.md
  Expected: Stage=Tasks, Confidence=Certain

  speckit-stage-implement/
    └── specs/001-feature/
        ├── spec.md
        ├── plan.md
        ├── tasks.md
        └── implement.md
  Expected: Stage=Implement, Confidence=Certain

  speckit-uncertain/
    └── specs/001-feature/
        └── partial.md
  Expected: Stage=Unknown, Confidence=Uncertain

  no-method-detected/
    └── README.md
  Expected: Method=unknown, Stage=Unknown

  empty-project/
    (empty directory)
  Expected: Method=unknown, Stage=Unknown

And detection tests run against all fixtures
And accuracy percentage calculated
And test fails if accuracy < 95%
```

**Technical Notes:**
- Follow Architecture "Test Fixture Naming Convention"
- Create `make test-accuracy` target
- 95% accuracy is launch blocker per PRD

**Prerequisites:** Story 2.4

---

**Epic 2 Complete**

**Stories Created:** 10
**FR Coverage:** FR1-8, FR9-14, FR39, FR44, FR45, FR48, FR49, FR53
**Technical Context Used:** Architecture sections on persistence, filesystem, detectors
**UX Patterns Incorporated:** Detection reasoning display, honest uncertainty

---

## Epic 3: Dashboard Visualization

**Goal:** Build the full interactive TUI dashboard with project list, navigation, and detail panel.

**User Value:** "I can see all my projects with their stages, navigate with keyboard, and view details - achieving the 'I know now!' moment in under 10 seconds."

**Technical Context:**
- Bubble Tea components (Architecture: adapters/tui/)
- Bubbles list with custom delegate
- Lipgloss styling
- Responsive layout

**UX Context:**
- Dashboard layout from UX Design Direction B
- ProjectItemDelegate rendering
- Detail panel with detection reasoning
- Status bar with counts and shortcuts
- Keyboard navigation patterns

**FRs Covered:** FR15-27 (Dashboard Visualization)

---

### Story 3.1: Project List Component

**As a** user,
**I want** to see my projects in a scrollable list,
**So that** I can browse all tracked projects.

**Acceptance Criteria:**

```gherkin
Given projects are tracked
When dashboard loads
Then project list displays with columns:
  - Selection indicator (> when selected)
  - Project name (or display_name if set)
  - Recency indicator (✨ today, ⚡ this week)
  - Stage name
  - Status (⏸️ WAITING if applicable)
  - Last activity time

And projects are sorted alphabetically by name
And first project is selected by default

When list exceeds visible rows
Then list becomes scrollable
And scroll position indicator shows

When no projects exist
Then EmptyView is shown instead

Example row format:
  "> client-bravo     ⚡ Plan      ⏸️ WAITING   2h ago"
```

**Technical Notes:**
- FR15: Real-time dashboard view
- FR16: Project name, stage, timestamp
- FR17: Visual indicators
- Use Bubbles list with custom ItemDelegate
- Follow UX "ProjectItemDelegate" specification

**Prerequisites:** Story 2.7, Story 1.6

---

### Story 3.2: Keyboard Navigation

**As a** user,
**I want** to navigate with keyboard shortcuts,
**So that** I can efficiently browse projects.

**Acceptance Criteria:**

```gherkin
Given project list is displayed
When I press 'j' or '↓'
Then selection moves down one item
And if at last item, wraps to first

When I press 'k' or '↑'
Then selection moves up one item
And if at first item, wraps to last

When I press 'g' (post-MVP)
Then selection moves to first item

When I press 'G' (post-MVP)
Then selection moves to last item

When I press 'q'
Then TUI exits cleanly

When I press 'Esc'
Then any active prompt is cancelled
And if no prompt, nothing happens (no exit)

And navigation feels instant (<50ms response)
And selection state persists during view switches
```

**Technical Notes:**
- FR18: Keyboard shortcuts [j/k/q]
- FR19: Vim-style keys
- Follow UX "Navigation Patterns"
- Implement in internal/adapters/tui/keys.go

**Prerequisites:** Story 3.1

---

### Story 3.3: Detail Panel Component

**As a** user,
**I want** to see detailed information about the selected project,
**So that** I understand detection reasoning and project context.

**Acceptance Criteria:**

```gherkin
Given a project is selected
When detail panel is visible
Then I see:
  """
  ┌─ DETAILS: client-bravo ─────────────────────┐
  │ Path:       /home/user/projects/client-bravo │
  │ Method:     Speckit                          │
  │ Stage:      Plan                             │
  │ Confidence: Certain                          │
  │ Detection:  plan.md exists, no tasks.md      │
  │ Notes:      Waiting on client API specs      │
  │ Added:      2025-12-01                        │
  │ Last Active: 2h ago                          │
  └──────────────────────────────────────────────┘
  """

When I press 'd'
Then detail panel toggles visibility
And layout adjusts smoothly

When detail panel is closed
Then project list expands to fill space

When terminal height < 30 rows
Then detail panel is closed by default
And hint shows: "Press [d] for details"

When project has no notes
Then Notes field shows: "(none)"

When detection is uncertain (🤷)
Then Confidence shows: "Uncertain"
And Detection shows reasoning for uncertainty
```

**Technical Notes:**
- FR20: View detailed information
- FR22: View project notes
- FR26: Display detection reasoning
- Follow UX "DetailPanel View" specification

**Prerequisites:** Story 3.1

---

### Story 3.4: Status Bar Component

**As a** user,
**I want** a persistent status bar,
**So that** I always see summary counts and available shortcuts.

**Acceptance Criteria:**

```gherkin
Given dashboard is displayed
Then status bar shows at bottom:
  """
  │ 5 active │ 2 hibernated │ ⏸️ 1 WAITING               │
  │ [j/k] nav [d] details [f] fav [?] help [q] quit     │
  """

And status bar is always visible (fixed position)
And counts update in real-time

When WAITING count > 0
Then "⏸️ N WAITING" displays in bold red (WaitingStyle)

When WAITING count = 0
Then WAITING section is hidden or shows "0 waiting" dimmed

When in hibernated view
Then shortcuts show "[h] back to active"

When terminal width < 80
Then shortcuts are abbreviated or wrapped
```

**Technical Notes:**
- FR24: Count of active vs hibernated
- Follow UX "StatusBar View" specification
- Two-line format: counts + shortcuts

**Prerequisites:** Story 3.1

---

### Story 3.5: Help Overlay

**As a** user,
**I want** to see all keyboard shortcuts,
**So that** I can learn available actions.

**Acceptance Criteria:**

```gherkin
Given dashboard is displayed
When I press '?'
Then help overlay appears:
  """
  ┌─ KEYBOARD SHORTCUTS ─────────────────────────┐
  │                                              │
  │  Navigation                                  │
  │  j/↓     Move down                          │
  │  k/↑     Move up                            │
  │                                              │
  │  Actions                                     │
  │  d        Toggle detail panel               │
  │  f        Toggle favorite                   │
  │  n        Edit notes                        │
  │  x        Remove project                    │
  │  a        Add project                       │
  │  r        Refresh/rescan                    │
  │                                              │
  │  Views                                       │
  │  h        View hibernated projects          │
  │                                              │
  │  General                                     │
  │  ?        Show this help                    │
  │  q        Quit                              │
  │  Esc      Cancel/close                      │
  │                                              │
  │         Press any key to close              │
  └──────────────────────────────────────────────┘
  """

When I press any key
Then help overlay closes
And previous view is restored
```

**Technical Notes:**
- FR65: View keyboard shortcut help
- Can use Bubbles help component or custom
- Follow UX "Help Overlay" pattern

**Prerequisites:** Story 3.2

---

### Story 3.6: Manual Refresh

**As a** user,
**I want** to manually refresh project detection,
**So that** I can force re-scan of artifacts.

**Acceptance Criteria:**

```gherkin
Given dashboard is displayed
When I press 'r'
Then all projects are re-scanned
And spinner shows in status bar: "⟳ Refreshing..."
And detection runs for each project
And stages update if artifacts changed
And status bar shows: "✓ Refreshed N projects"

When refresh completes
Then timestamp "last refreshed Xs ago" updates

When refresh encounters errors
Then partial success reported
And errors logged (not shown in UI unless all fail)

When I run `vibe refresh` from CLI
Then same refresh occurs non-interactively
```

**Technical Notes:**
- FR23: Manual refresh forces re-scan
- FR58: Trigger refresh via CLI
- Non-blocking - navigation still works during refresh

**Prerequisites:** Story 3.1, Story 2.5

---

### Story 3.7: Project Notes (View & Edit)

**As a** user,
**I want** to add notes to projects,
**So that** I can capture context that detection can't know.

**Acceptance Criteria:**

```gherkin
Given a project is selected
When I press 'n'
Then inline note editor opens:
  """
  ┌─ Edit note for "client-bravo" ────────────────┐
  │ > Waiting on client API specs█                │
  │ [Enter] save  [Esc] cancel                    │
  └───────────────────────────────────────────────┘
  """

When I type text and press Enter
Then note is saved to database
And detail panel updates to show note
And feedback shows: "✓ Note saved"

When I press Esc
Then edit is cancelled
And original note preserved

When note is empty and I save
Then note is cleared

When I run `vibe note client-bravo "New note"`
Then note is set via CLI
```

**Technical Notes:**
- FR21: Add/edit notes
- FR22: View notes in detail
- FR55: Set notes via CLI
- Use Bubbles textinput component

**Prerequisites:** Story 3.3

---

### Story 3.8: Toggle Favorite

**As a** user,
**I want** to mark projects as favorites,
**So that** they stay visible regardless of activity.

**Acceptance Criteria:**

```gherkin
Given a project is selected
When I press 'f'
Then favorite status toggles immediately
And ⭐ indicator appears/disappears in project row
And feedback shows: "⭐ Favorited" or "☆ Unfavorited"

When project is favorited
Then it displays with ⭐ prefix
And it never auto-hibernates

When I run `vibe favorite client-bravo`
Then favorite is toggled via CLI

When I run `vibe favorite client-bravo --off`
Then favorite is removed via CLI
```

**Technical Notes:**
- FR30: Mark as favorite
- FR31: Remove favorite status
- FR54: Mark/unmark via CLI
- No confirmation needed (easily reversible)

**Prerequisites:** Story 3.1

---

### Story 3.9: Remove Project from TUI

**As a** user,
**I want** to remove projects from the dashboard,
**So that** I can clean up without leaving TUI.

**Acceptance Criteria:**

```gherkin
Given a project is selected
When I press 'x'
Then inline confirmation shows:
  "Remove 'client-bravo' from tracking? [y/n]"

When I press 'y'
Then project is removed
And list updates immediately
And feedback shows: "✓ Removed: client-bravo"

When I press 'n' or Esc
Then removal is cancelled
And project remains

When confirmation times out (30s)
Then auto-cancel occurs

While confirmation is active
Then other keys are ignored (except y/n/Esc)
```

**Technical Notes:**
- Uses ConfirmPrompt component
- Follow UX "Confirmation Patterns"
- 30-second timeout per UX spec

**Prerequisites:** Story 3.1, Story 2.8

---

### Story 3.10: Responsive Layout

**As a** user,
**I want** the dashboard to adapt to my terminal size,
**So that** it works in various environments.

**Acceptance Criteria:**

```gherkin
Given terminal is resized during operation
When width < 60 columns
Then minimal view shows with warning:
  "Terminal too small. Minimum 60x20 required."

When width is 60-79 columns
Then truncated project names
And warning shown about limited view

When width is 80-99 columns
Then standard column widths
And full functionality

When width > 120 columns
Then content is centered
And maximum width is capped

When height < 20 rows
Then list-only view (no detail panel)
And status bar condensed

When height 20-34 rows
Then detail panel closed by default
And hint to open with 'd'

When height >= 35 rows
Then detail panel open by default

When resize occurs rapidly (drag)
Then layout recalculates with 50ms debounce
And no visual flicker
```

**Technical Notes:**
- Follow UX "Terminal Responsive Strategy"
- Implement debounced resize handling
- Use layout caching per UX spec

**Prerequisites:** Story 3.1, Story 3.3

---

**Epic 3 Complete**

**Stories Created:** 10
**FR Coverage:** FR15-27, FR30-31, FR54-55, FR58, FR65
**Technical Context Used:** TUI adapter architecture
**UX Patterns Incorporated:** Full dashboard layout, all components

---

## Epic 4: Agent Waiting Detection (Killer Feature)

**Goal:** Implement the killer feature - detecting when AI coding agents are waiting for user input.

**User Value:** "I never forget a waiting AI agent. The dashboard shows ⏸️ WAITING when agents need my input, saving hours of lost productivity."

**Technical Context:**
- File system monitoring via fsnotify
- Inactivity threshold heuristics
- Time-based detection service

**UX Context:**
- Bold red ⏸️ WAITING indicator
- Peripheral vision alerts in status bar
- Elapsed wait time display

**FRs Covered:** FR34-38 (Agent Monitoring)

---

### Story 4.1: File Watcher Service

**As a** system,
**I want** to monitor file changes in tracked projects,
**So that** I can detect activity and inactivity.

**Acceptance Criteria:**

```gherkin
Given projects are tracked
When FileWatcher starts
Then fsnotify watches are created for:
  - Each project's root directory
  - Key subdirectories (specs/, .bmad/, src/)
  - Recursive watch where supported

When file is created/modified/deleted
Then FileEvent is emitted with:
  - Path (canonical)
  - Operation (create/modify/delete)
  - Timestamp

When multiple events fire rapidly
Then events are debounced (200ms default)
And single aggregated event emitted

When watch fails on a path
Then error is logged
And fallback to polling suggested
And other watches continue

When project is added
Then watch is registered automatically

When project is removed
Then watch is unregistered

When application exits
Then all watches are cleaned up gracefully
```

**Technical Notes:**
- FR27: Automatically detect file system changes
- Use fsnotify library
- Implement DebouncedWatcher per Architecture
- 200ms debounce window configurable

**Prerequisites:** Story 1.3

---

### Story 4.2: Activity Timestamp Tracking

**As a** system,
**I want** to track last activity time per project,
**So that** I can calculate inactivity duration.

**Acceptance Criteria:**

```gherkin
Given file watcher is running
When file event occurs in project directory
Then project.LastActivityAt is updated
And database is updated
And "last active: Xm ago" display refreshes

When TUI is open
Then timestamps update every minute
And display shows relative time:
  - "just now" (< 1 min)
  - "5m ago" (< 1 hour)
  - "2h ago" (< 24 hours)
  - "3d ago" (< 7 days)
  - "2w ago" (>= 7 days)

When refresh is triggered
Then all timestamps recalculate
And display updates immediately
```

**Technical Notes:**
- Store LastActivityAt in SQLite
- Use time.Since() for relative calculation
- Update display on tick (every 60s)

**Prerequisites:** Story 4.1, Story 2.1

---

### Story 4.3: Agent Waiting Detection Logic

**As a** system,
**I want** to detect when an AI agent is waiting for input,
**So that** users are alerted to blocked agents.

**Acceptance Criteria:**

```gherkin
Given agent_waiting_threshold_minutes is 10 (default)
When project has no file activity for 10+ minutes
Then project is marked as "waiting"
And ⏸️ WAITING indicator appears

When file activity resumes
Then waiting state clears automatically
And indicator disappears

When calculating waiting duration
Then duration shows time since last activity:
  - "⏸️ WAITING 15m"
  - "⏸️ WAITING 2h"
  - "⏸️ WAITING 1d"

When project is hibernated
Then waiting detection is disabled
And no false WAITING indicators

When project was inactive before adding
Then initial state is NOT waiting
And waiting only triggers after observed activity then silence

Edge cases:
- Project with no activity ever: NOT waiting (never started)
- Project inactive 9m59s: NOT waiting (under threshold)
- Project inactive 10m01s: IS waiting
- Vacation scenario (7d inactive): IS waiting (user decision to acknowledge)
```

**Technical Notes:**
- FR34: Detect agent waiting via inactivity
- FR38: Clear waiting state on activity
- Threshold boundary testing critical (9:59 vs 10:01)

**Prerequisites:** Story 4.2

---

### Story 4.4: Waiting Threshold Configuration

**As a** user,
**I want** to configure the waiting threshold,
**So that** I can tune sensitivity to my workflow.

**Acceptance Criteria:**

```gherkin
Given default threshold is 10 minutes
When I set in config.yaml:
  """yaml
  settings:
    agent_waiting_threshold_minutes: 5
  """
Then waiting triggers after 5 minutes of inactivity

When I set per-project threshold:
  """yaml
  projects:
    client-bravo:
      agent_waiting_threshold_minutes: 30
  """
Then client-bravo uses 30 minute threshold
And other projects use global default

When I run `vibe --waiting-threshold=15`
Then CLI flag overrides config for this session

When threshold is set to 0
Then waiting detection is disabled for that scope
```

**Technical Notes:**
- FR37: Configure waiting threshold
- FR46: Global settings
- FR47: Per-project override
- Config cascade: CLI > project > global > default

**Prerequisites:** Story 4.3, Story 1.7

---

### Story 4.5: Waiting Indicator Display

**As a** user,
**I want** waiting projects to visually pop,
**So that** I notice them during quick scans.

**Acceptance Criteria:**

```gherkin
Given a project is in waiting state
When displayed in project list
Then row shows:
  - "⏸️ WAITING" in STATUS column
  - Text styled with WaitingStyle (bold red)
  - Elapsed time: "2h" format

When multiple projects are waiting
Then all show ⏸️ indicator
And status bar shows total: "⏸️ 3 WAITING" in red

When no projects waiting
Then status bar shows clean count without WAITING section

Visual treatment:
- ⏸️ emoji provides non-color indicator
- Bold makes text heavier
- Red color (ANSI 1) catches peripheral vision
- Entire "⏸️ WAITING Xh" styled together
```

**Technical Notes:**
- FR35: Display ⏸️ WAITING indicator
- FR36: Show elapsed time
- Follow UX "Color System" - red reserved for WAITING only

**Prerequisites:** Story 4.3, Story 3.4

---

### Story 4.6: Real-Time Dashboard Updates

**As a** user,
**I want** the dashboard to update automatically,
**So that** I see current state without manual refresh.

**Acceptance Criteria:**

```gherkin
Given dashboard is open
When file changes in tracked project
Then dashboard updates within 5-10 seconds:
  - LastActivity timestamp updates
  - WAITING state clears if was waiting
  - Recency indicator may change (⚡ → ✨)

When update occurs
Then no jarring visual changes
And selection position preserved
And smooth re-render

When file watcher fails
Then status bar shows warning: "⚠ File watching unavailable"
And manual refresh still works

When TUI is in background (unfocused)
Then updates still occur
And visible immediately when focused
```

**Technical Notes:**
- Use Bubble Tea Cmd for async updates
- Send custom Msg when file event received
- FR27: Auto-detect file system changes
- 5-10 second detection per PRD NFR-P6

**Prerequisites:** Story 4.1, Story 3.1

---

**Epic 4 Complete**

**Stories Created:** 6
**FR Coverage:** FR27, FR34-38, FR46-47
**Technical Context Used:** fsnotify, time-based heuristics
**UX Patterns Incorporated:** WaitingStyle, peripheral vision alerts

---

## Epic 4.5: BMAD Method v6 State Detection

**Goal:** Implement BMAD Method v6 detection as a second MethodDetector plugin, enabling vibe-dash to detect and display workflow state for projects using the BMAD Method.

**User Value:** "I can track my BMAD Method projects alongside Speckit projects - the dashboard shows me where I am in the BMAD workflow (Analysis, Planning, Solutioning, Implementation)."

**Scope:** BMAD v6 only (`.bmad/` folder structure). v4 support (`.bmad-core/`) deferred for future if demand emerges.

**Technical Context:**
- Implements existing `MethodDetector` interface (proven by Speckit detector)
- Detects `.bmad/` folder with `bmm/config.yaml`
- Parses `sprint-status.yaml` for epic/story status
- Maps BMAD phases to standardized stages

**Reference Implementation:** `github.com/ibadmore/bmad-progress-dashboard` (analyzed for detection patterns)

**FRs Covered:** FR13 (Pluggable methodology detectors), FR14 (Multiple methodologies)

---

### Story 4.5-1: BMAD v6 Detector Implementation

**As a** developer using BMAD Method,
**I want** vibe-dash to detect my BMAD v6 project structure,
**So that** my project appears with the correct methodology identified.

**Acceptance Criteria:**

```gherkin
Given a project directory with `.bmad/` folder
And `.bmad/bmm/config.yaml` exists
When MethodDetector.Detect() is called
Then returns MethodInfo with:
  - Name: "bmad"
  - Version: extracted from config.yaml (e.g., "6.0.0-alpha.13")
  - Confidence: High (folder structure matches)

Given a project directory without `.bmad/` folder
When MethodDetector.Detect() is called
Then returns nil (not a BMAD project)

Given a project with `.bmad-core/` (v4 structure)
When MethodDetector.Detect() is called
Then returns nil (v4 not supported in this story)
```

**Technical Notes:**
- Create `internal/adapters/detectors/bmad/detector.go`
- Implement `MethodDetector` interface from `internal/core/ports/detector.go`
- Register with detector registry (same pattern as Speckit)
- Config path: `.bmad/bmm/config.yaml`

**Prerequisites:** Epic 2 (MethodDetector interface exists)

---

### Story 4.5-2: BMAD v6 Stage Detection Logic

**As a** developer using BMAD Method,
**I want** vibe-dash to show my current workflow stage,
**So that** I know where I am in the BMAD process.

**Acceptance Criteria:**

```gherkin
Given a BMAD v6 project with `sprint-status.yaml`
When stage detection runs
Then parses development_status section
And determines current phase from epic statuses:
  - All epics backlog → "Planning"
  - Epic in-progress with stories → "Implementing"
  - All epics done → "Complete"

Given stage detection with epic/story analysis
Then maps to standardized stages:
  | BMAD State | Displayed Stage |
  |------------|-----------------|
  | No epics | Plan |
  | Epics backlog | Specify |
  | Epic in-progress | Implement |
  | Stories in review | Review |
  | All epics done | Validate |

Given detection with reasoning
Then provides explanation like:
  "Epic 4 in-progress, Story 4.3 being implemented"

Given no sprint-status.yaml found
Then falls back to artifact detection:
  - Has PRD → "Plan"
  - Has Architecture → "Specify"
  - Has Epics → "Implement"
```

**Technical Notes:**
- Sprint status location: `docs/sprint-artifacts/sprint-status.yaml` or per config
- Parse YAML for `development_status` section
- Epic status values: `backlog`, `in-progress`, `done`
- Story status values: `backlog`, `drafted`, `ready-for-dev`, `in-progress`, `review`, `done`

**Prerequisites:** Story 4.5-1

---

### Story 4.5-3: BMAD Test Fixtures

**As a** developer,
**I want** comprehensive test coverage for BMAD detection,
**So that** the detector is reliable and maintainable.

**Acceptance Criteria:**

```gherkin
Given test fixtures in `internal/adapters/detectors/bmad/testdata/`
Then includes:
  - valid-v6-project/ (complete .bmad structure)
  - minimal-v6-project/ (just .bmad/bmm/config.yaml)
  - mid-sprint-project/ (with sprint-status.yaml)
  - completed-project/ (all epics done)
  - no-bmad-project/ (control case)

Given vibe-dash's own `.bmad/` folder
Then use as real-world dogfooding test:
  - Copy relevant structure to testdata
  - Verify detection matches expected state

Given unit tests
Then covers:
  - Detector registration
  - Folder detection
  - Config parsing
  - Stage detection from sprint-status
  - Fallback to artifact detection
  - Edge cases (missing files, malformed YAML)

Given integration test
Then verifies end-to-end:
  - Add project with BMAD structure
  - Confirm methodology detected as "bmad"
  - Confirm stage displayed correctly
```

**Technical Notes:**
- Follow existing test patterns from Speckit detector
- Use table-driven tests for stage mapping
- Mock file system for unit tests where appropriate

**Prerequisites:** Story 4.5-1, Story 4.5-2

---

**Epic 4.5 Complete**

**Stories Created:** 3
**FR Coverage:** FR13 (pluggable detectors), FR14 (multiple methodologies)
**Technical Context Used:** MethodDetector interface, sprint-status.yaml parsing
**Dogfooding:** vibe-dash's own .bmad folder serves as test fixture

---

## Epic 5: Project State & Hibernation

**Goal:** Implement automatic state management with active/hibernated/favorite states.

**User Value:** "My dashboard stays clean - active projects are visible, dormant ones auto-hibernate, and favorites stay pinned regardless of activity."

**Technical Context:**
- State machine: Active ↔ Hibernated
- Favorite override (always visible)
- Auto-promotion on activity

**UX Context:**
- Hibernation flow from UX
- State indicators
- Separate hibernated view

**FRs Covered:** FR28-33 (Project State Management)

---

### Story 5.1: Project State Model

**As a** developer,
**I want** project state machine implemented,
**So that** state transitions are well-defined.

**Acceptance Criteria:**

```gherkin
Given Project domain entity
Then State field has values:
  - StateActive
  - StateHibernated

And state transitions are:
  Active → Hibernated (via auto-hibernate or manual)
  Hibernated → Active (via auto-promote or manual)

And IsFavorite is independent of State:
  - Favorite + Active: Always visible
  - Favorite + Hibernated: Still hibernated but marked

And state change triggers:
  - Database update
  - TUI refresh
  - Timestamp update
```

**Technical Notes:**
- FR33: Distinguish active/hibernated/favorite
- State stored in SQLite `state` column
- IsFavorite is separate boolean

**Prerequisites:** Story 1.2

---

### Story 5.2: Auto-Hibernation

**As a** user,
**I want** inactive projects to auto-hibernate,
**So that** my dashboard stays focused on active work.

**Acceptance Criteria:**

```gherkin
Given hibernation_days is 14 (default)
When project has no activity for 14+ days
Then project state changes to Hibernated
And project disappears from active list
And hibernated count increases in status bar

When project is favorited
Then it never auto-hibernates
And remains in active list always

When checking hibernation
Then check runs:
  - On application launch
  - On manual refresh
  - Every hour during TUI session

When project hibernates
Then no notification (silent transition)
And user sees updated count: "2 hibernated"
```

**Technical Notes:**
- FR28: Auto-mark hibernated after X days
- FR32: Configurable threshold
- Check on launch and periodically

**Prerequisites:** Story 5.1, Story 4.2

---

### Story 5.3: Auto-Activation on Activity

**As a** user,
**I want** hibernated projects to auto-activate when I work on them,
**So that** they reappear in my active dashboard.

**Acceptance Criteria:**

```gherkin
Given project "old-project" is hibernated
When file changes detected in old-project/
Then project state changes to Active
And project appears in active list
And hibernated count decreases

When file watcher detects activity
Then auto-activation is immediate
And user sees project appear in list

When manually opening hibernated project
Then it remains hibernated until file change
Or user manually activates it
```

**Technical Notes:**
- FR29: Auto-mark active on file changes
- Integrate with file watcher from Epic 4

**Prerequisites:** Story 5.2, Story 4.1

---

### Story 5.4: Hibernated Projects View

**As a** user,
**I want** to view and manage hibernated projects,
**So that** I can reactivate or remove them.

**Acceptance Criteria:**

```gherkin
Given some projects are hibernated
When I press 'h' in dashboard
Then view switches to hibernated list:
  """
  ┌─ HIBERNATED PROJECTS ────────────────────────┐
  │ PROJECT          LAST ACTIVE    ACTION       │
  │ old-client       3w ago         [Enter] wake │
  │ experiment-1     2mo ago        [Enter] wake │
  └──────────────────────────────────────────────┘
  │ 2 hibernated │ Press [h] to return           │
  """

When I press Enter on a project
Then project is reactivated
And moves to active list
And view switches to active dashboard

When I press 'h' again or Esc
Then view returns to active dashboard

When I press 'x' on hibernated project
Then removal confirmation shows
And project can be deleted

When no projects hibernated
Then message shows: "No hibernated projects."
```

**Technical Notes:**
- FR25: View hibernated projects list
- Same navigation keys work (j/k)
- Selection state separate from active list

**Prerequisites:** Story 5.2, Story 3.1

---

### Story 5.5: Manual State Control

**As a** user,
**I want** to manually hibernate or activate projects,
**So that** I can override automatic behavior.

**Acceptance Criteria:**

```gherkin
Given project "client-alpha" is active
When I run `vibe hibernate client-alpha`
Then project moves to hibernated state
And message shows: "✓ Hibernated: client-alpha"

Given project "old-project" is hibernated
When I run `vibe activate old-project`
Then project moves to active state
And message shows: "✓ Activated: old-project"

When in TUI on hibernated view
When I press Enter on project
Then project activates (same as CLI)

When trying to hibernate a favorite
Then warning shows: "Note: Favorites don't auto-hibernate but state was changed"
And state still changes (user explicitly requested)
```

**Technical Notes:**
- FR57: Manually hibernate/activate via CLI
- Implement in internal/adapters/cli/state.go

**Prerequisites:** Story 5.4

---

### Story 5.6: Hibernation Threshold Configuration

**As a** user,
**I want** to configure hibernation threshold,
**So that** I can tune to my project lifecycle.

**Acceptance Criteria:**

```gherkin
Given default threshold is 14 days
When I set in config.yaml:
  """yaml
  settings:
    hibernation_days: 7
  """
Then projects hibernate after 7 days

When I set per-project threshold:
  """yaml
  projects:
    long-term-project:
      hibernation_days: 30
  """
Then long-term-project uses 30 day threshold

When I set hibernation_days: 0
Then auto-hibernation is disabled
And projects never auto-hibernate

When I run `vibe --hibernation-days=21`
Then CLI flag overrides for this session
```

**Technical Notes:**
- FR32: Configure hibernation threshold
- FR46: Global settings
- FR47: Per-project override

**Prerequisites:** Story 5.2, Story 1.7

---

**Epic 5 Complete**

**Stories Created:** 6
**FR Coverage:** FR28-33, FR57
**Technical Context Used:** State machine, configuration cascade
**UX Patterns Incorporated:** Hibernation flow, separate view

---

## Epic 6: Scripting & Automation

**Goal:** Enable full CLI automation with JSON output, exit codes, and shell completion.

**User Value:** "I can script vibe-dash in my automation workflows - CI/CD, morning standup scripts, monitoring - with reliable JSON output and exit codes."

**Technical Context:**
- Cobra CLI commands
- JSON serialization with API versioning
- Standard exit codes
- Shell completion generation

**UX Context:**
- Non-interactive mode output formats
- Scriptable commands

**FRs Covered:** FR48-61 (Scripting & Automation)

---

### Story 6.1: JSON Output Format

**As a** scripter,
**I want** JSON output with API versioning,
**So that** my scripts are stable across updates.

**Acceptance Criteria:**

```gherkin
Given I need machine-readable output
When I run `vibe list --json`
Then output is valid JSON:
  """json
  {
    "api_version": "v1",
    "projects": [...]
  }
  """

When I run `vibe list --json --api-version=v1`
Then explicit v1 schema is used

When I run `vibe status client-alpha --json`
Then single project JSON returned:
  """json
  {
    "api_version": "v1",
    "project": {
      "name": "client-alpha",
      "path": "/home/user/client-alpha",
      "method": "speckit",
      "stage": "plan",
      "confidence": "certain",
      "state": "active",
      "is_favorite": false,
      "is_waiting": true,
      "waiting_duration_minutes": 45,
      "last_activity_at": "2025-12-11T10:30:00Z",
      "notes": "Waiting on API specs"
    }
  }
  """

When JSON schema changes in future
Then new api_version (v2) is introduced
And v1 remains supported for compatibility
```

**Technical Notes:**
- FR49: JSON output with API versioning
- Use snake_case for JSON keys (Architecture convention)
- ISO 8601 timestamps in UTC

**Prerequisites:** Story 2.7

---

### Story 6.2: Project Status Command

**As a** scripter,
**I want** to query specific project status,
**So that** I can check state in CI/CD pipelines.

**Acceptance Criteria:**

```gherkin
Given project "client-alpha" exists
When I run `vibe status client-alpha`
Then output shows:
  """
  client-alpha
    Method: Speckit
    Stage: Plan
    State: Active
    Last Active: 2h ago
  """

When I run `vibe status client-alpha --json`
Then JSON output as specified in Story 6.1

When project doesn't exist
Then error: "Project not found: client-alpha"
And exit code: 2

When I run `vibe status --all`
Then all projects status shown
```

**Technical Notes:**
- FR50: Get specific project status non-interactively

**Prerequisites:** Story 2.7

---

### Story 6.3: Exit Codes

**As a** scripter,
**I want** standard exit codes,
**So that** I can handle errors in scripts.

**Acceptance Criteria:**

```gherkin
Given vibe command executes
When operation succeeds
Then exit code is 0

When general error occurs
Then exit code is 1

When project not found
Then exit code is 2
And error message shows which project

When configuration invalid
Then exit code is 3
And error message shows config issue

When detection fails
Then exit code is 4
And error message shows what failed

Example script usage:
  """bash
  if vibe status my-project > /dev/null 2>&1; then
    echo "Project exists"
  else
    case $? in
      2) echo "Project not found" ;;
      *) echo "Error occurred" ;;
    esac
  fi
  """
```

**Technical Notes:**
- FR60: Standard exit codes (0/1/2/3/4)
- Map domain errors to exit codes in CLI adapter
- Follow Architecture "Error-to-Exit-Code Mapping"

**Prerequisites:** Story 1.4

---

### Story 6.4: Project Exists Check

**As a** scripter,
**I want** to check if a project is tracked,
**So that** I can conditionally add or skip.

**Acceptance Criteria:**

```gherkin
Given I need to check project existence
When I run `vibe exists client-alpha`
And project exists
Then exit code is 0
And no output (silent success)

When project doesn't exist
Then exit code is 2
And no output (silent failure)

Usage in scripts:
  """bash
  if vibe exists my-project; then
    echo "Already tracked"
  else
    vibe add .
  fi
  """
```

**Technical Notes:**
- FR59: Check if project exists via CLI
- Silent operation for scripting

**Prerequisites:** Story 2.3

---

### Story 6.5: Rename Project Command

**As a** user,
**I want** to rename projects via CLI,
**So that** I can set display names without TUI.

**Acceptance Criteria:**

```gherkin
Given project "api-service" exists
When I run `vibe rename api-service "Client A API"`
Then display_name is set to "Client A API"
And message shows: "✓ Renamed: api-service → Client A API"

When I run `vibe rename api-service --clear`
Then display_name is cleared
And project shows original name derived from path

When project doesn't exist
Then error and exit code 2
```

**Technical Notes:**
- FR5: Set custom display name
- FR56: Rename via CLI

**Prerequisites:** Story 2.3

---

### Story 6.6: Shell Completion

**As a** user,
**I want** shell tab completion,
**So that** I can efficiently type commands.

**Acceptance Criteria:**

```gherkin
Given vibe is installed
When I run `vibe completion bash`
Then Bash completion script is output

When I run `vibe completion zsh`
Then Zsh completion script is output

When I run `vibe completion fish`
Then Fish completion script is output

When completion is installed
Then `vibe <TAB>` shows available commands
And `vibe status <TAB>` shows project names
And `vibe --<TAB>` shows available flags
```

**Technical Notes:**
- FR61: Shell completion (Bash/Zsh/Fish)
- Use Cobra's built-in completion generation
- Essentially free with Cobra

**Prerequisites:** Story 1.4

---

### Story 6.7: Quiet and Force Flags

**As a** scripter,
**I want** quiet mode and force flags,
**So that** I can automate without prompts.

**Acceptance Criteria:**

```gherkin
Given I need silent operation
When I run `vibe add . --quiet`
Then no output on success
And exit code indicates result

When I run `vibe remove client-alpha --force`
Then no confirmation prompt
And project removed immediately

When I run `vibe add . --force`
Then collision resolved automatically
And no prompt for disambiguation

Combining flags:
  `vibe remove client-alpha --force --quiet`
  → Silent removal, no prompts, exit code only
```

**Technical Notes:**
- FR51: Add with automatic conflict resolution
- FR52: Force flag bypasses prompts
- Quiet mode suppresses stdout, errors still go to stderr

**Prerequisites:** Story 2.3, Story 2.8

---

**Epic 6 Complete**

**Stories Created:** 7
**FR Coverage:** FR48-61
**Technical Context Used:** Cobra CLI, JSON serialization
**UX Patterns Incorporated:** Non-interactive output formats

---

## Epic 7: Error Handling & Polish

**Goal:** Implement graceful error handling, helpful feedback, and final polish.

**User Value:** "The tool handles errors gracefully, gives me helpful feedback, and feels polished and reliable."

**Technical Context:**
- Domain error types
- slog structured logging
- Recovery strategies

**UX Context:**
- Error recovery patterns from UX
- Feedback patterns
- Loading indicators

**FRs Covered:** FR62-66 (Error Handling & User Feedback)

---

### Story 7.1: File Watcher Error Recovery

**As a** user,
**I want** graceful handling when file watching fails,
**So that** I can still use the dashboard.

**Acceptance Criteria:**

```gherkin
Given file watcher is running
When fsnotify fails on a path
Then warning shows in status bar:
  "⚠ File watching unavailable for: client-alpha"
And other watches continue working
And manual refresh still works

When all watches fail
Then status bar shows:
  "⚠ File watching unavailable. Use [r] to refresh."
And dashboard remains functional
And real-time updates disabled

When watch recovers (e.g., network mount reconnects)
Then watching resumes automatically
And warning clears
```

**Technical Notes:**
- FR62: Gracefully handle file watching failures
- Fallback to manual refresh mode
- Log errors with slog for debugging

**Prerequisites:** Story 4.1

---

### Story 7.2: Configuration Error Handling

**As a** user,
**I want** helpful messages when config has errors,
**So that** I can fix issues easily.

**Acceptance Criteria:**

```gherkin
Given config.yaml has syntax error
When vibe launches
Then error shows:
  """
  ⚠ Config syntax error in ~/.vibe-dash/config.yaml
  Line 5: invalid YAML - unexpected character
  Using default settings.
  """
And application continues with defaults
And exit code is 0 (degraded, not failed)

When config has invalid values
Then error shows:
  """
  ⚠ Invalid config value: hibernation_days must be >= 0
  Using default: 14
  """
And valid values are used

When I run `vibe config --validate`
Then config is validated
And all issues reported
And exit code reflects validity (0 or 3)
```

**Technical Notes:**
- FR63: Detect and report config syntax errors
- Viper provides line numbers for YAML errors
- Continue with defaults rather than failing

**Prerequisites:** Story 1.7

---

### Story 7.3: Database Recovery

**As a** user,
**I want** corrupted state to be recoverable,
**So that** I don't lose my project tracking.

**Acceptance Criteria:**

```gherkin
Given database is corrupted
When vibe launches
Then error shows:
  """
  ⚠ Database corrupted: ~/.vibe-dash/projects.db
  Attempting recovery from config...
  """

When recovery succeeds
Then projects are restored from config.yaml paths
And detection re-runs for all projects
And message shows: "✓ Recovered N projects"

When recovery fails
Then error shows:
  """
  ✗ Database recovery failed.
  Run 'vibe reset --confirm' to start fresh.
  Your projects.yaml in config will be preserved.
  """

When I run `vibe reset --confirm`
Then database is deleted
And fresh database created
And projects re-added from config paths
```

**Technical Notes:**
- FR64: Recover from corrupted state
- Config.yaml stores paths as backup
- Re-detection rebuilds state

**Prerequisites:** Story 2.1

---

### Story 7.4: Progress Indicators

**As a** user,
**I want** progress indication during long operations,
**So that** I know the tool is working.

**Acceptance Criteria:**

```gherkin
Given a long operation is running
When refresh scans many projects
Then spinner shows in status bar:
  "⟳ Scanning... (5/20)"
And count updates as projects complete

When operation completes
Then spinner stops
And result shows: "✓ Scanned 20 projects"

When operation is cancelled (Esc)
Then partial results are kept
And message shows: "Cancelled. 12/20 projects scanned."

Progress scenarios:
- Initial load: "Loading projects..."
- Refresh: "⟳ Scanning... (N/total)"
- Add project: "Detecting methodology..."
```

**Technical Notes:**
- FR66: Display progress indicators
- Use Bubbles spinner component
- Non-blocking - navigation still works

**Prerequisites:** Story 3.1

---

### Story 7.5: Verbose and Debug Logging

**As a** user,
**I want** logging options for troubleshooting,
**So that** I can diagnose issues.

**Acceptance Criteria:**

```gherkin
Given default logging
When vibe runs normally
Then only errors go to stderr
And TUI is clean

When I run `vibe --verbose`
Then info-level logs included:
  "INFO: Scanning project: client-alpha"
  "INFO: Detected: Speckit at Plan stage"

When I run `vibe --debug`
Then debug-level logs included:
  "DEBUG: [detection_service.go:45] Starting detection"
  "DEBUG: [speckit.go:32] Checking for specs/ directory"
And file/line info included

Logs go to stderr to not interfere with JSON output:
  `vibe list --json 2>/dev/null`  # Clean JSON
  `vibe list --json --debug 2>debug.log`  # JSON + logs
```

**Technical Notes:**
- Use log/slog (Go 1.21+ stdlib)
- Follow Architecture "Logging & Observability"
- Log at handling site only

**Prerequisites:** Story 1.4

---

### Story 7.6: Feedback Messages Polish

**As a** user,
**I want** consistent, helpful feedback messages,
**So that** I understand what happened.

**Acceptance Criteria:**

```gherkin
Given action is performed
Then feedback follows UX patterns:

Success messages (green, 3s):
  "✓ Added: client-alpha"
  "✓ Removed: client-bravo"
  "✓ Note saved"
  "✓ Refreshed 5 projects"

Error messages (red, persistent):
  "✗ Path not found: /invalid/path"
  "✗ Project not found: unknown"
  "✗ Database error: connection failed"

Warning messages (yellow, 4s):
  "⚠ Could not detect methodology"
  "⚠ File watching unavailable"
  "⚠ Config using defaults"

Info messages (default, 3s):
  "⭐ Favorited"
  "☆ Unfavorited"
  "Cancelled"

All messages:
- Clear and concise
- Actionable when possible
- Consistent iconography (✓ ✗ ⚠)
```

**Technical Notes:**
- Follow UX "Feedback Patterns"
- Implement feedback message component
- 3-4 second display per accessibility

**Prerequisites:** Story 3.1

---

### Story 7.7: Graceful Shutdown

**As a** user,
**I want** clean shutdown on Ctrl+C,
**So that** no data is corrupted.

**Acceptance Criteria:**

```gherkin
Given vibe TUI is running
When I press Ctrl+C or send SIGTERM
Then shutdown sequence executes:
  1. Cancel root context
  2. Wait for in-flight operations (5s max)
  3. Flush pending database writes
  4. Close database connections
  5. Restore terminal state
  6. Exit with code 0

When shutdown takes too long
Then 5 second timeout forces exit
And warning logged: "Shutdown timeout exceeded"

When database write was in progress
Then write completes or rolls back cleanly
And no partial state corruption

When terminal was in alternate screen
Then alternate screen is exited
And user's previous content restored
```

**Technical Notes:**
- Follow Architecture "Graceful Shutdown Pattern"
- Use context cancellation
- Never call os.Exit() from goroutines

**Prerequisites:** Story 1.5

---

**Epic 7 Complete**

**Stories Created:** 7
**FR Coverage:** FR62-66
**Technical Context Used:** Error handling patterns, logging
**UX Patterns Incorporated:** Feedback patterns, error recovery

---

## FR Coverage Matrix

| FR | Description | Epic | Story |
|----|-------------|------|-------|
| FR1 | Add project from current directory | 2 | 2.3 |
| FR2 | Add project from specified path | 2 | 2.3 |
| FR3 | View list of all tracked projects | 2 | 2.7 |
| FR4 | Remove project from tracking | 2 | 2.8 |
| FR5 | Set custom display name | 2, 6 | 2.3, 6.5 |
| FR6 | Detect and resolve name collisions | 2 | 2.6 |
| FR7 | Validate project paths at launch | 2 | 2.9 |
| FR8 | Choose action for missing paths | 2 | 2.9 |
| FR9 | Detect Speckit methodology | 2 | 2.4 |
| FR10 | Identify current Speckit stage | 2 | 2.4 |
| FR11 | Show detection reasoning | 2 | 2.4 |
| FR12 | Indicate uncertainty | 2 | 2.4 |
| FR13 | Support pluggable detectors | 2 | 2.5 |
| FR14 | Detect multiple methodologies | 2 | 2.5 |
| FR15 | View real-time dashboard | 3 | 3.1 |
| FR16 | See project name, stage, timestamp | 3 | 3.1 |
| FR17 | Visual indicators | 3 | 3.1 |
| FR18 | Keyboard shortcuts | 3 | 3.2 |
| FR19 | Vim-style keys | 3 | 3.2 |
| FR20 | View detailed project information | 3 | 3.3 |
| FR21 | Add/edit notes | 3 | 3.7 |
| FR22 | View notes in detail | 3 | 3.3, 3.7 |
| FR23 | Manual refresh | 3 | 3.6 |
| FR24 | Count of active vs hibernated | 3 | 3.4 |
| FR25 | View hibernated list | 5 | 5.4 |
| FR26 | Display detection reasoning | 3 | 3.3 |
| FR27 | Auto-detect file changes | 4 | 4.1, 4.6 |
| FR28 | Auto-hibernate after X days | 5 | 5.2 |
| FR29 | Auto-activate on file changes | 5 | 5.3 |
| FR30 | Mark project as favorite | 3 | 3.8 |
| FR31 | Remove favorite status | 3 | 3.8 |
| FR32 | Configure hibernation threshold | 5 | 5.6 |
| FR33 | Distinguish states | 5 | 5.1 |
| FR34 | Detect agent waiting | 4 | 4.3 |
| FR35 | Display ⏸️ WAITING indicator | 4 | 4.5 |
| FR36 | Show elapsed waiting time | 4 | 4.5 |
| FR37 | Configure waiting threshold | 4 | 4.4 |
| FR38 | Clear waiting on activity | 4 | 4.3 |
| FR39 | Store paths in master config | 2 | 2.1 |
| FR40 | Store project-specific settings | 2 | 2.1 |
| FR41 | Override config via CLI flags | 4, 5 | 4.4, 5.6 |
| FR42 | Modify global config | 1 | 1.7 |
| FR43 | Modify project config | 2 | 2.1 |
| FR44 | Auto-create default config | 1 | 1.7 |
| FR45 | Use canonical paths | 2 | 2.2 |
| FR46 | Configure global settings | 4, 5 | 4.4, 5.6 |
| FR47 | Configure per-project settings | 4, 5 | 4.4, 5.6 |
| FR48 | List in plain text format | 2 | 2.7 |
| FR49 | JSON output with versioning | 6 | 6.1 |
| FR50 | Get project status non-interactively | 6 | 6.2 |
| FR51 | Add with automatic conflict resolution | 6 | 6.7 |
| FR52 | Force flag for auto-resolution | 6 | 6.7 |
| FR53 | Remove via CLI | 2 | 2.8 |
| FR54 | Mark/unmark favorites via CLI | 3 | 3.8 |
| FR55 | Set notes via CLI | 3 | 3.7 |
| FR56 | Rename via CLI | 6 | 6.5 |
| FR57 | Hibernate/activate via CLI | 5 | 5.5 |
| FR58 | Trigger refresh via CLI | 3 | 3.6 |
| FR59 | Check project exists via CLI | 6 | 6.4 |
| FR60 | Standard exit codes | 6 | 6.3 |
| FR61 | Shell completion | 6 | 6.6 |
| FR62 | Handle file watching failures | 7 | 7.1 |
| FR63 | Detect config syntax errors | 7 | 7.2 |
| FR64 | Recover from corrupted state | 7 | 7.3 |
| FR65 | View shortcut help | 3 | 3.5 |
| FR66 | Display progress indicators | 7 | 7.4 |

---

## Summary

**Total Epics:** 9 (including 4.5, excluding deferred Epic 5)
**Total Stories:** 64

| Epic | Stories | FRs Covered |
|------|---------|-------------|
| 1 - Foundation | 7 | Infrastructure |
| 2 - Project Management | 10 | FR1-14, FR39-45, FR48, FR53 |
| 3 - Dashboard | 10 | FR15-24, FR26, FR30-31, FR54-55, FR58, FR65 |
| 4 - Agent Waiting | 6 | FR27, FR34-38, FR46-47 |
| 4.5 - BMAD Detection | 3 | FR13-14 |
| 5 - State & Hibernation | 6 | FR25, FR28-33, FR57 (DEFERRED) |
| 6 - Scripting | 7 | FR49-52, FR56, FR59-61 |
| 7 - Error Handling | 7 | FR62-66 |
| 8 - UX Polish | 14 | UX improvements |
| 9 - Scalable Watching | 8 | FR27, FR34-38 (architecture) |

**FR Coverage:** 66/66 (100%)

**Ready for Phase 4:** Sprint Planning and Development Implementation

---

## Epic 9: Scalable File Watching Architecture

**Goal:** Re-architect file watching to scale to 20+ projects with deep directory structures without exhausting OS file descriptors.

**User Value:** "I can track 20+ projects simultaneously, each with thousands of files, and the dashboard remains responsive - agent waiting detection and project status updates work reliably at scale."

**Background & Problem Statement:**

The current fsnotify-based approach (Stories 4.1, 8.1) has fundamental scalability limitations:

| Current State | Impact |
|---------------|--------|
| fsnotify uses kqueue on macOS | 1 file descriptor per watched directory |
| Story 8.1 watches up to 500 dirs/project | 10 projects × 500 dirs = 5000 FDs |
| macOS default ulimit is 256 | System-wide file descriptor exhaustion |
| Re-enumeration on every Watch() call | High CPU during periodic refresh |

**Symptoms observed:**
- "too many open files" errors after extended runtime
- Dashboard shows "file watching unavailable"
- System-wide impact (tmux, other processes affected)
- Welcome screen displayed despite projects existing

**What File Watching Enables:**
1. **Agent Waiting Detection** (FR34-38): Detects when AI agent is idle by monitoring file activity
2. **Project Status Detection**: Triggers re-detection when methodology artifacts change
3. **Real-time Dashboard Updates** (FR27): Live activity timestamps

**Technical Context:**
- Platform-specific optimal approaches needed
- macOS: FSEvents API can watch entire tree with single stream
- Linux: inotify with higher limits, or fanotify for recursive
- Fallback: Polling-based detection for universal compatibility

**FRs Impacted:** FR27, FR34-38 (agent waiting), project status detection

---

### Story 9.1: File Watching Architecture Research

**As a** developer,
**I want** a comprehensive analysis of file watching strategies,
**So that** we choose the optimal approach for multi-platform, multi-project scale.

**Acceptance Criteria:**

```gherkin
Given the current fsnotify limitations
When research is conducted
Then document includes:

Platform Analysis:
| Platform | Current | Optimal | FD Usage |
|----------|---------|---------|----------|
| macOS | kqueue (fsnotify) | FSEvents | 1 per tree |
| Linux | inotify (fsnotify) | inotify/fanotify | 1 per dir, higher limits |
| Windows | ReadDirectoryChangesW | Same | Recursive by default |

Go Library Evaluation:
- fsnotify: Current, kqueue on macOS, inotify on Linux
- rjeczalik/notify: Uses FSEvents on macOS, more efficient
- fsevents (macOS only): Direct FSEvents bindings
- Custom polling: Universal fallback

Resource Analysis:
- FD usage per 10/20/50 projects
- CPU usage: event-based vs polling at 1s/5s/30s intervals
- Memory usage per approach
- Latency from file change to detection

Decision Criteria:
- Must detect file changes in <10 seconds
- Must support 20+ projects with 1000+ files each
- Must not exhaust system resources
- Must work on macOS and Linux (Windows nice-to-have)
```

**Technical Notes:**
- Output: `docs/architecture/file-watching-research.md`
- Include benchmark methodology
- Reference: `github.com/rjeczalik/notify` documentation

**Prerequisites:** None (research story)

---

### Story 9.2: Scalable Watching Strategy Design

**As a** architect,
**I want** a detailed design for scalable file watching,
**So that** implementation has clear direction.

**Acceptance Criteria:**

```gherkin
Given research from Story 9.1
When architecture design is complete
Then design document includes:

Strategy Decision:
- Primary approach for each platform
- Fallback approach when primary fails
- Hybrid options (e.g., FSEvents + polling)

Interface Design:
- Abstract FileWatcher interface (already exists)
- Platform-specific implementations
- Factory pattern for platform detection

Configuration:
- max_watch_directories: global limit
- watching_strategy: "auto" | "fsevents" | "polling" | "hybrid"
- polling_interval_seconds: for polling/hybrid modes
- watch_depth: how deep to recurse (current: 10)

Graceful Degradation:
- When FD limit approached → switch to polling
- When polling fails → disable real-time, manual refresh only
- Warning UI patterns for degraded modes

Migration Path:
- Backwards compatibility with current config
- Feature flag for gradual rollout
- Rollback mechanism

Architecture Diagram:
┌─────────────────────────────────────────────┐
│                FileWatcher Port             │
├─────────────────────────────────────────────┤
│ Watch(ctx, paths) → <-chan FileEvent        │
│ Close() error                               │
│ GetFailedPaths() []string                   │
└─────────────────────────────────────────────┘
         │
         ├── FSEventsWatcher (macOS)
         │     └── Single stream per project tree
         │
         ├── InotifyWatcher (Linux)
         │     └── Optimized directory enumeration
         │
         ├── PollingWatcher (fallback)
         │     └── Periodic file stat comparison
         │
         └── HybridWatcher
               └── Events + periodic validation
```

**Technical Notes:**
- Output: `docs/architecture/file-watching-design.md`
- Include decision matrix for strategy selection
- Define metrics for monitoring

**Prerequisites:** Story 9.1

---

### Story 9.3: Platform-Optimized Watcher Implementation

**As a** developer,
**I want** platform-specific file watcher implementations,
**So that** each platform uses its most efficient mechanism.

**Acceptance Criteria:**

```gherkin
Given design from Story 9.2
When implementations are complete
Then:

macOS (FSEventsWatcher):
- Uses FSEvents API via appropriate Go bindings
- Single event stream per project root
- Recursive watching without per-directory FDs
- Coalesces rapid events (built-in)
- Handles symlinks correctly

Linux (OptimizedInotifyWatcher):
- Uses inotify but with smarter enumeration
- Watches project root + key directories only
- Relies on inotify recursive events where available
- Falls back to polling for deep structures

Fallback (PollingWatcher):
- Configurable interval (default: 5 seconds)
- Uses file modification time comparison
- Caches previous state efficiently
- Minimal CPU when idle
- Works universally

Factory:
func NewPlatformWatcher(config WatcherConfig) ports.FileWatcher {
    switch runtime.GOOS {
    case "darwin":
        return NewFSEventsWatcher(config)
    case "linux":
        return NewOptimizedInotifyWatcher(config)
    default:
        return NewPollingWatcher(config)
    }
}

All implementations:
- Implement existing ports.FileWatcher interface
- Emit same FileEvent format
- Thread-safe
- Graceful shutdown
- Resource cleanup on Close()
```

**Technical Notes:**
- May require build tags for platform-specific code
- FSEvents: consider `github.com/fsnotify/fsevents` or `github.com/rjeczalik/notify`
- Polling: efficient file tree walking with caching

**Prerequisites:** Story 9.2

---

### Story 9.4: Intelligent Watch Scope Reduction

**As a** system,
**I want** to watch only essential directories,
**So that** resource usage is minimized while detection still works.

**Acceptance Criteria:**

```gherkin
Given a project with deep directory structure
When determining what to watch
Then prioritize:

Essential (always watched):
- Project root (for top-level changes)
- .bmad/ (methodology artifacts - recursive)
- .speckit/ or specs/ (methodology artifacts - recursive)
- src/ (if exists, common code location)
- Configuration files (go.mod, package.json, etc.)

Optional (watched if under limit):
- Other top-level directories
- Depth-limited subdirectories

Never watched:
- node_modules/, vendor/, target/, __pycache__
- .git/, .svn/, .hg/
- Hidden directories (except .bmad, .speckit)

Smart depth:
- Full depth for small projects (<100 dirs)
- Limited depth for large projects
- Dynamic adjustment based on available FDs

Configuration:
```yaml
settings:
  file_watching:
    max_directories_per_project: 100  # Reduced from 500
    priority_directories:
      - ".bmad"
      - ".speckit"
      - "specs"
      - "src"
    watch_depth: 5  # Reduced from 10
```
```

**Technical Notes:**
- Detect essential directories at project add time
- Store watched paths for faster re-watch
- Log when directories are skipped due to limits

**Prerequisites:** Story 9.3

---

### Story 9.5: Watcher Resource Monitoring

**As a** system,
**I want** to monitor file watcher resource usage,
**So that** problems are detected before system impact.

**Acceptance Criteria:**

```gherkin
Given file watching is active
When monitoring resource usage
Then track:

Metrics:
- Total directories being watched
- File descriptors in use (where available)
- Events received per second
- Time since last event per project

Thresholds:
- Warning at 80% of estimated safe limit
- Critical at 95% of limit
- Auto-degrade to polling at critical

UI Integration:
- Status bar warning: "⚠ High watch usage (450/500 dirs)"
- Detail panel shows watching status per project
- Debug mode shows full metrics

Logging:
- slog.Warn when approaching limits
- slog.Info for watching strategy changes
- Metrics logged every 5 minutes in debug mode

Recovery:
- Automatic switch to polling when limits hit
- Attempt to recover to events after 5 minutes
- User can force strategy via config
```

**Technical Notes:**
- Use `lsof` or `/proc/self/fd` for FD counting
- Store metrics in memory (not persisted)
- Expose via debug endpoint if CLI has server mode

**Prerequisites:** Story 9.3

---

### Story 9.6: Watcher Migration & Configuration

**As a** user,
**I want** to configure file watching strategy,
**So that** I can tune for my environment.

**Acceptance Criteria:**

```gherkin
Given the new watcher implementations
When configuring file watching
Then config.yaml supports:

```yaml
settings:
  file_watching:
    # Strategy: auto, fsevents, inotify, polling, hybrid
    strategy: auto

    # Polling interval (for polling/hybrid strategies)
    polling_interval_seconds: 5

    # Maximum directories to watch per project
    max_directories_per_project: 100

    # Maximum depth for recursive watching
    watch_depth: 5

    # Disable file watching entirely (manual refresh only)
    disabled: false
```

When strategy is "auto"
Then platform-optimal strategy is selected:
- macOS: FSEvents
- Linux: Optimized inotify
- Other: Polling

When strategy is explicitly set
Then that strategy is used regardless of platform

When --watch-strategy=polling CLI flag used
Then polling is used for this session only

Migration:
- Existing configs continue to work
- New defaults are applied for missing settings
- Deprecation warning for old settings (if any)
```

**Technical Notes:**
- Viper bindings for new config keys
- CLI flag override support
- Validate strategy is valid for platform

**Prerequisites:** Story 9.3

---

### Story 9.7: Scale Testing & Validation

**As a** developer,
**I want** comprehensive scale testing,
**So that** the new architecture is validated.

**Acceptance Criteria:**

```gherkin
Given new watcher implementations
When scale testing
Then verify:

Test Scenarios:
| Scenario | Projects | Dirs/Project | Total Dirs | Expected |
|----------|----------|--------------|------------|----------|
| Small | 5 | 50 | 250 | All event-based |
| Medium | 20 | 100 | 2000 | Event-based, no warnings |
| Large | 50 | 200 | 10000 | Graceful degradation |
| Stress | 100 | 500 | 50000 | Polling fallback |

Metrics to Capture:
- File descriptor usage over 1 hour
- CPU usage (idle, during activity burst)
- Memory usage
- Event latency (file change → detection)
- Detection accuracy (no missed changes)

Test Implementation:
- Automated test script for scale scenarios
- Docker-based for reproducibility
- CI integration for regression

Acceptance:
- Medium scenario: No resource warnings
- Large scenario: Graceful degradation, no crashes
- All scenarios: No missed file changes
- All scenarios: <10 second detection latency
```

**Technical Notes:**
- Use synthetic file generation for testing
- Monitor with `time`, `pprof`, system metrics
- Document results in test report

**Prerequisites:** Stories 9.3, 9.4, 9.5

---

### Story 9.8: Documentation & Troubleshooting Guide

**As a** user,
**I want** clear documentation on file watching,
**So that** I can troubleshoot issues and tune performance.

**Acceptance Criteria:**

```gherkin
Given new file watching architecture
When documentation is complete
Then includes:

User Guide:
- What file watching does
- How to configure for your environment
- Common issues and solutions

Troubleshooting:
| Symptom | Cause | Solution |
|---------|-------|----------|
| "too many open files" | FD exhaustion | Reduce max_dirs, use polling |
| High CPU usage | Polling too aggressive | Increase polling_interval |
| Missed changes | Polling too slow | Decrease interval or use events |
| "file watching unavailable" | Platform issue | Check logs, try polling |

Performance Tuning:
- Small setups (<10 projects): default settings
- Medium setups (10-30 projects): reduce dirs, increase interval
- Large setups (30+ projects): polling strategy, longer intervals

Architecture Documentation:
- Platform-specific implementation details
- Resource usage characteristics
- Extension points for new platforms
```

**Technical Notes:**
- Add to existing docs/project-context.md or separate file
- Include in --help output for relevant commands
- Link from error messages

**Prerequisites:** Story 9.7

---

**Epic 9 Complete**

**Stories Created:** 8
**FR Impact:** FR27 (file change detection), FR34-38 (agent waiting)
**Technical Context:** Platform-specific file watching, resource management
**Risk Mitigation:** Graceful degradation prevents system-wide impact

---

## Epic 9 Dependency Graph

```
Epic 9 (Scalable Watching)
    │
    ├── Story 9.1 (Research)
    │       │
    │       ▼
    ├── Story 9.2 (Design)
    │       │
    │       ▼
    ├── Story 9.3 (Implementation)
    │       │
    │       ├────────────────────────┐
    │       │                        │
    │       ▼                        ▼
    ├── Story 9.4 (Scope)     Story 9.5 (Monitoring)
    │       │                        │
    │       └────────────────────────┘
    │                   │
    │                   ▼
    ├── Story 9.6 (Migration & Config)
    │                   │
    │                   ▼
    ├── Story 9.7 (Scale Testing)
    │                   │
    │                   ▼
    └── Story 9.8 (Documentation)
```

**Estimated Effort:** Medium-Large (significant research + platform-specific implementation)
**Priority:** High (blocks reliable 20+ project support)
**Can Run In Parallel With:** None (foundational infrastructure change)

---

## Epic 12: Agentic Tool Log Viewing

**User Value:** View Claude Code (and future agentic tool) session logs directly from vibe-dash TUI without manual path navigation.

**Business Context:** When monitoring multiple AI-assisted projects, manually checking Claude Code logs is tedious. The path mapping (`/path/to/project` → `~/.claude/projects/-path-to-project/`) requires mental effort. Integrating log viewing into vibe-dash provides instant access and complements the existing "Agent Waiting" detection feature.

**Technical Context:** Follows the `MethodDetector` adapter pattern for extensibility. The `LogReader` port interface enables future support for other agentic tools (Cursor, Aider, Windsurf).

**Scope:**
- Claude Code log viewer (MVP)
- Raw JSON display (user's separate CLI handles formatting)
- Live tailing with auto-scroll
- Session switching

**Out of Scope (Deferred):**
- Pretty formatting (handled by user's log viewer CLI)
- Other agentic tool adapters
- Entry type filtering

---

### Story 12.1: Claude Code Log Viewer

**As a** developer monitoring multiple AI-assisted projects,
**I want** to view Claude Code session logs directly from the vibe-dash TUI,
**So that** I can quickly check what Claude is doing without manually navigating to log files.

**Acceptance Criteria:**

```gherkin
Scenario: Enter on project opens log view
  Given a project with Claude Code logs at ~/.claude/projects/{path-with-dashes}/
  When I press Enter on the project in normal view
  Then the TUI switches to full-screen log view
  And shows the latest session's log entries

Scenario: Live tailing updates
  Given I am viewing a Claude Code session
  When Claude writes new entries to the log file
  Then the log view updates within 2 seconds
  And auto-scrolls to show the latest entry

Scenario: Manual scroll pauses auto-scroll
  Given I am viewing a Claude Code session with auto-scroll
  When I press k or up arrow to scroll up
  Then auto-scroll pauses
  And I can read earlier entries without interruption

Scenario: Resume auto-scroll with G
  Given auto-scroll is paused
  When I press G
  Then the view jumps to the latest entry
  And auto-scroll resumes

Scenario: Session switching
  Given I am viewing a Claude Code session
  When I press S
  Then a session picker overlay appears
  And shows all available sessions sorted by recency
  And each session shows: truncated ID, start time, entry count, summary

Scenario: No logs graceful handling
  Given a project without Claude Code logs
  When I press Enter on the project
  Then a flash message "No Claude Code logs for this project" appears
  And the message clears after 2 seconds

Scenario: LogReader follows MethodDetector pattern
  Given the architecture uses hexagonal patterns
  When LogReader port is implemented
  Then it has CanRead(), Tool(), ListSessions(), TailSession() methods
  And follows the same adapter/registry pattern as MethodDetector
```

**Technical Notes:**
- Path mapping: `/Users/foo/bar` → `-Users-foo-bar`
- Log location: `~/.claude/projects/{path-with-dashes}/*.jsonl`
- Entry types: `summary`, `user`, `assistant`, `system`, `file-history-snapshot`
- Use polling-based live tail (1-2s interval) - boring technology that works

**Architecture References:**
- [Source: internal/core/ports/detector.go] - MethodDetector pattern to follow
- [Source: internal/adapters/detectors/registry.go] - Registry pattern
- [Source: internal/adapters/tui/validation.go:15-21] - viewMode enum pattern

**Prerequisites:** None (can start immediately)

---

**Future Enhancement: Log-Based Agent Waiting Detection**

The log viewing infrastructure could enhance the existing agent waiting detection (FR34-38). Instead of relying solely on file inactivity heuristics, we could detect waiting state with higher accuracy by analyzing log entries:

- `AskUserQuestion` tool call → **Certain** waiting
- Assistant message ends with `?` → **Likely** waiting
- `tool_use` with no `tool_result` → **Likely** pending approval

This creates a synergy between Epic 4 (Agent Waiting) and Epic 12 (Log Viewing). Deferred to future story after Story 12.1 proves the infrastructure.

---

### Story 12.2: Log Viewer Polish

**As a** developer using the log viewer,
**I want** improved keyboard navigation and visual feedback,
**So that** I can efficiently navigate, search, and read logs with familiar vim-style controls.

**Acceptance Criteria:**

```gherkin
Scenario: 'L' key opens session selector (bug fix)
  Given I am viewing the project list
  When I press 'L'
  Then the session picker overlay appears
  And help overlay shows 'L' as shortcut

Scenario: Half-page navigation with Ctrl+U/D
  Given I am in log viewer mode
  When I press Ctrl+D or Ctrl+U
  Then the view scrolls half a page down/up

Scenario: Double 'g' jumps to top (vim-standard)
  Given I am in log viewer mode
  When I press 'g' twice within 500ms
  Then the view jumps to the first line

Scenario: Search mode with '/'
  Given I am in log viewer mode
  When I press '/'
  Then search input appears
  And 'n'/'N' navigate matches
  And match counter shows (e.g., "3/15")

Scenario: ANSI colors preserved
  Given cclv outputs ANSI color codes
  When I view the log
  Then colors are rendered correctly
  And long lines truncate without breaking colors
```

**Technical Notes:**
- 'L' key missing from keys.go - add and wire to session picker
- Implement gg detection with timer state (500ms threshold)
- Search: ~100-150 lines (state, UI, matching)
- ANSI: Add go-runewidth dependency for visual width calculation
- Files: keys.go, model.go, views.go

**Prerequisites:** Story 12.1 (Claude Code Log Viewer)

---

**Epic 12 In Progress**

**Stories Created:** 2
**FR Impact:** New functionality - complements FR34-38 (Agent Monitoring)
**Technical Context:** LogReader adapter pattern, JSONL parsing, TUI view mode
**Extensibility:** Pattern enables future Cursor/Aider/Windsurf adapters + enhanced waiting detection

