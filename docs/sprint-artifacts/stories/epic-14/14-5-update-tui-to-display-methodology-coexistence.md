# Story 14.5: Update TUI to Display Methodology Coexistence

Status: done

## Story

As a user,
I want to see both methodologies in the dashboard when coexistence is detected,
So that I can understand my project's mixed state.

## User-Visible Changes

- **New:** When two methodologies have similar activity timestamps (within 1 hour), the stage column shows both (e.g., "Speckit/BMAD") or a mixed indicator
- **New:** The detail panel shows a warning message explaining both methodologies are active when coexistence is detected
- **Changed:** Detection now uses timestamp-aware selection to determine which methodology to display as primary

## Acceptance Criteria

1. **AC1: Project row shows mixed methodology** - Given `DetectWithCoexistenceSelection` returns two methodologies with `CoexistenceWarning=true`, when dashboard renders, then stage column shows "Speckit/BMAD" or similar combined format

2. **AC2: Detail panel shows warning** - Given coexistence is detected, when detail panel renders for that project, then it shows warning section: "Warning: Both Speckit (Plan) and BMAD (Epic 8) detected with similar activity"

3. **AC3: Primary methodology for stage display** - When coexistence is detected, use the methodology with most recent artifact timestamp for the stage display (even in tie case, pick one consistently)

4. **AC4: Single methodology has no warning** - Given only one methodology detected, then no coexistence warning displayed (normal behavior)

5. **AC5: Clear winner has no warning** - Given methodologies with >1 hour timestamp difference, then display winner's stage/methodology only with no warning

6. **AC6: Integration with existing detection flow** - The TUI must call `DetectWithCoexistenceSelection` instead of `Detect` to receive coexistence information

## Tasks / Subtasks

- [x] Task 1: Add coexistence fields to Project domain model (AC: 4, 5, 6)
  - [x] Subtask 1.1: Add `CoexistenceWarning bool` field to `Project` struct in `internal/core/domain/project.go` (add after `DetectionReasoning` field at line 20)
  - [x] Subtask 1.2: Add `CoexistenceMessage string` field for display text
  - [x] Subtask 1.3: Add `SecondaryMethod string` field to store secondary methodology when coexistence detected
  - [x] Subtask 1.4: Add `SecondaryStage Stage` field to store secondary methodology's stage

- [x] Task 2: Update TUI model to use coexistence-aware detection (AC: 6)
  - [x] Subtask 2.1: In `internal/adapters/tui/model.go`, modify `refreshProjectsCmd()` (line 765) to call `DetectWithCoexistenceSelection` instead of `Detect`
  - [x] Subtask 2.2: Handle the multi-return: `(winner *DetectionResult, allResults []*DetectionResult, error)`
  - [x] Subtask 2.3: When winner is nil (tie case), pick first result as display primary, set coexistence fields. Note: `allResults[0]` is already the most recent by timestamp - results are pre-sorted.
  - [x] Subtask 2.4: Store all coexistence info in Project: `CoexistenceWarning`, `CoexistenceMessage`, `SecondaryMethod`, `SecondaryStage`
  - [x] Subtask 2.5: When clear winner, clear any existing coexistence warning on project (reset fields)
  - [x] Subtask 2.6: Audit ALL other `Detect()` calls in TUI layer to ensure consistent detection behavior

- [x] Task 3: Update detail panel to display coexistence warning (AC: 2, 4, 5)
  - [x] Subtask 3.1: In `internal/adapters/tui/components/detail_panel.go`, add coexistence warning section after Detection field
  - [x] Subtask 3.2: Use existing `styles.WarningStyle` for warning display (matches established pattern)
  - [x] Subtask 3.3: Format: "Warning: Both {Method1} ({Stage1}) and {Method2} ({Stage2}) detected with similar activity"
  - [x] Subtask 3.4: Use `emoji.Warning()` prefix for visual consistency
  - [x] Subtask 3.5: Only show section when `project.CoexistenceWarning == true`

- [x] Task 4: Update project row to show mixed methodology indicator (AC: 1, 3)
  - [x] Subtask 4.1: In `internal/adapters/tui/components/delegate.go`, modify stage column rendering
  - [x] Subtask 4.2: When `CoexistenceWarning=true`, prepend warning emoji to stage info
  - [x] Subtask 4.3: Update `stageformat.FormatStageInfo` or add wrapper to handle coexistence display
  - [x] Subtask 4.4: Alternative format option: "{PrimaryStage} (+{SecondaryMethod})" if full names too long

- [x] Task 5: Add unit tests (AC: 1, 2, 4, 5)
  - [x] Subtask 5.1: Test detail panel renders coexistence warning when flag is true
  - [x] Subtask 5.2: Test detail panel omits warning when flag is false
  - [x] Subtask 5.3: Test project row displays mixed indicator for coexistence
  - [x] Subtask 5.4: Test project row displays normal stage for single methodology
  - [x] Subtask 5.5: Test TUI model correctly populates coexistence fields from detection

- [x] Task 6: Run `make fmt && make lint && make test` to verify all passes

## Dev Notes

### Domain Model Changes

**File:** `internal/core/domain/project.go`

Add fields after `DetectionReasoning` field (line 20). Insert before the `IsFavorite` field:

```go
type Project struct {
	// ... existing fields through line 20 ...
	DetectionReasoning string       // Human-readable detection explanation (FR11, FR26)
	// NEW: Coexistence fields for Story 14.5 (runtime-only, not persisted)
	CoexistenceWarning bool   // True when multiple methodologies with similar timestamps
	CoexistenceMessage string // Warning text for display
	SecondaryMethod    string // Second methodology when coexistence detected
	SecondaryStage     Stage  // Second methodology's stage
	IsFavorite         bool         // Always visible regardless of activity (FR30)
	// ... rest of fields ...
}
```

**Important:** These fields are runtime-only for display. They do NOT need database schema changes or SQLite persistence - they are recomputed on each detection refresh.

### TUI Model Changes

**File:** `internal/adapters/tui/model.go` - modify `refreshProjectsCmd()` (around line 765)

Replace the existing detection call:
```go
result, err := m.detectionService.Detect(ctx, project.Path)
```

With coexistence-aware detection:
```go
winner, allResults, err := m.detectionService.DetectWithCoexistenceSelection(ctx, project.Path)
if err != nil {
	slog.Debug("refresh detection failed", "project", project.Name, "error", err)
	failedCount++
	continue
}

// Determine primary result and populate coexistence fields
var primary *domain.DetectionResult
if winner != nil {
	primary = winner
	// Clear any previous coexistence warning
	project.CoexistenceWarning = false
	project.CoexistenceMessage = ""
	project.SecondaryMethod = ""
	project.SecondaryStage = domain.StageUnknown
} else if len(allResults) > 0 {
	// Tie case - use first as primary (already sorted by most recent timestamp)
	primary = allResults[0]
	project.CoexistenceWarning = primary.HasCoexistenceWarning()
	project.CoexistenceMessage = primary.CoexistenceMessage
	if len(allResults) > 1 {
		project.SecondaryMethod = allResults[1].Method
		project.SecondaryStage = allResults[1].Stage
	}
} else {
	// No methodology detected - use unknown
	unknownResult := domain.NewDetectionResult("unknown", domain.StageUnknown, domain.ConfidenceUncertain, "No methodology detected")
	primary = &unknownResult
	project.CoexistenceWarning = false
	project.CoexistenceMessage = ""
	project.SecondaryMethod = ""
	project.SecondaryStage = domain.StageUnknown
}

// Update project with primary result (existing pattern)
project.DetectedMethod = primary.Method
project.CurrentStage = primary.Stage
project.Confidence = primary.Confidence
project.DetectionReasoning = primary.Reasoning
```

**Note:** Uses `domain.ConfidenceUncertain` (not `ConfidenceUnknown` which doesn't exist). The error handling pattern matches existing code.

### Detail Panel Changes

**File:** `internal/adapters/tui/components/detail_panel.go`

Add after Detection field section (around line 160):

```go
// Coexistence warning section - Story 14.5
if p.CoexistenceWarning {
	warningText := fmt.Sprintf("Both %s (%s) and %s (%s) detected with similar activity",
		p.DetectedMethod,
		p.CurrentStage.String(),
		p.SecondaryMethod,
		p.SecondaryStage.String(),
	)
	warningLine := fmt.Sprintf("%s %s",
		emoji.Warning(),
		styles.WarningStyle.Render(warningText),
	)
	lines = append(lines, formatField("Coexistence", warningLine))
}
```

**Note:** Uses `Stage.String()` directly since `stageformat.FormatStage()` doesn't exist. Uses `formatField()` helper for consistent label alignment.

### Project Row Changes

**File:** `internal/adapters/tui/components/delegate.go`

Modify stage column rendering (around line 261-267). The current code is:

```go
stage := stageformat.FormatStageInfoWithWidth(item.Project, stageWidth)
```

Add coexistence handling after this line:

```go
// Stage column with coexistence indicator - Story 14.5
stage := stageformat.FormatStageInfoWithWidth(item.Project, stageWidth)

// Add coexistence indicator if warning set
if item.Project.CoexistenceWarning {
	// Prepend warning emoji to stage info
	stage = fmt.Sprintf("%s %s", emoji.Warning(), stage)
	// Truncate if needed to fit width (account for emoji width ~2 chars)
	if len(stage) > stageWidth {
		stage = stage[:stageWidth-3] + "..."
	}
}
```

**Note:** Uses existing `stageformat.FormatStageInfoWithWidth(item.Project, stageWidth)` signature (takes `*domain.Project`, not separate strings). Uses simple `len()` truncation since `lipgloss.Width` may not be needed for this use case.

### Architecture Compliance

| Rule | Compliance |
|------|------------|
| Hexagonal boundary | Domain fields in `internal/core/domain/`, TUI display in `internal/adapters/tui/` |
| No adapter imports in core | Domain only uses stdlib types |
| Established patterns | Uses existing `styles.WarningStyle`, `emoji.Warning()` patterns |
| Co-located tests | Tests in same packages as code |

### Existing Warning Pattern Reference

From `internal/adapters/tui/components/status_bar.go` and `detail_panel.go`:

```go
// Established pattern for warnings
styles.WarningStyle.Render(text)  // Yellow/orange styling
emoji.Warning()                    // "⚠️" character
```

### Test Pattern Reference

**Detail Panel Test** - Add to `internal/adapters/tui/components/detail_panel_test.go`:

```go
func TestDetailPanel_CoexistenceWarning(t *testing.T) {
	project := &domain.Project{
		ID:                 "test-id",
		Name:               "test-project",
		Path:               "/test/path",
		DetectedMethod:     "speckit",
		CurrentStage:       domain.StagePlan,
		CoexistenceWarning: true,
		CoexistenceMessage: "Multiple methodologies detected with similar activity",
		SecondaryMethod:    "bmad",
		SecondaryStage:     domain.StageEpic,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		LastActivityAt:     time.Now(),
	}

	panel := NewDetailPanelModel(80, 20)
	panel.SetVisible(true)
	panel.SetProject(project)
	output := panel.View()

	// Should contain warning indicator and both methodologies
	if !strings.Contains(output, "Coexistence") {
		t.Error("expected Coexistence label in output")
	}
	if !strings.Contains(output, "speckit") || !strings.Contains(output, "bmad") {
		t.Error("expected both methodologies in warning")
	}
}

func TestDetailPanel_NoWarningWhenFlagFalse(t *testing.T) {
	project := &domain.Project{
		ID:                 "test-id",
		Name:               "test-project",
		Path:               "/test/path",
		DetectedMethod:     "speckit",
		CurrentStage:       domain.StagePlan,
		CoexistenceWarning: false, // No warning
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		LastActivityAt:     time.Now(),
	}

	panel := NewDetailPanelModel(80, 20)
	panel.SetVisible(true)
	panel.SetProject(project)
	output := panel.View()

	// Should NOT contain coexistence warning
	if strings.Contains(output, "Coexistence") {
		t.Error("should not show Coexistence label when flag is false")
	}
}
```

**Delegate Test** - Add to `internal/adapters/tui/components/delegate_test.go`:

```go
func TestProjectItemDelegate_CoexistenceIndicator(t *testing.T) {
	tests := []struct {
		name           string
		coexistence    bool
		wantWarning    bool
	}{
		{"with coexistence", true, true},
		{"without coexistence", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := &domain.Project{
				ID:                 "test-id",
				Name:               "test-project",
				Path:               "/test/path",
				DetectedMethod:     "speckit",
				CurrentStage:       domain.StagePlan,
				CoexistenceWarning: tt.coexistence,
				LastActivityAt:     time.Now(),
			}
			item := ProjectItem{Project: project}
			delegate := NewProjectItemDelegate(100)

			var buf strings.Builder
			// Note: Render method writes to io.Writer
			// Create mock list.Model for testing
			// This may require additional setup based on existing test patterns

			// Assert: output should contain warning emoji only when coexistence is true
			// The actual test will depend on existing delegate test patterns in the file
		})
	}
}
```

### Files to Create/Modify

| File | Action | Purpose |
|------|--------|---------|
| `internal/core/domain/project.go` | MODIFY | Add coexistence fields |
| `internal/adapters/tui/model.go` | MODIFY | Use `DetectWithCoexistenceSelection` |
| `internal/adapters/tui/components/detail_panel.go` | MODIFY | Add coexistence warning section |
| `internal/adapters/tui/components/delegate.go` | MODIFY | Add warning indicator to stage column |
| `internal/adapters/tui/components/detail_panel_test.go` | MODIFY | Add coexistence warning tests |
| `internal/adapters/tui/components/delegate_test.go` | MODIFY or CREATE | Add coexistence indicator tests |

### Story Dependencies

| Dependency | Status | Relationship |
|------------|--------|--------------|
| Story 14.1 (DetectWithCoexistence) | COMPLETED | Provides multi-detection registry method |
| Story 14.2 (ArtifactTimestamp) | COMPLETED | Provides timestamp field for comparison |
| Story 14.3 (Most-Recent-Wins) | COMPLETED | Provides SelectByTimestamp logic |
| Story 14.4 (Coexistence Warning) | COMPLETED | Provides CoexistenceWarning flag and message |

### Critical Implementation Notes

1. **Preserve existing behavior:** When only one methodology detected or clear winner exists, behavior is identical to before. No user-visible change for normal cases.

2. **Secondary method sorting:** When tie detected, `allResults[0]` is the most recent by timestamp. Results are pre-sorted by `SelectByTimestamp`. AC3 is automatically satisfied.

3. **Stage format compatibility:** The `stageformat.FormatStageInfo` function already handles both Speckit and BMAD formats. Use `Stage.String()` for simple stage name display in warnings.

4. **Field persistence note:** The new domain fields (`CoexistenceWarning`, `SecondaryMethod`, etc.) are runtime-only for display. They do NOT need to be persisted to SQLite - they are recomputed on each detection.

5. **Width considerations:** The coexistence indicator (`emoji.Warning()`) adds ~2 characters. Ensure stage column truncation accounts for this.

6. **Confidence constant:** Use `domain.ConfidenceUncertain` (not `ConfidenceUnknown` - that doesn't exist). Valid values are: `ConfidenceCertain`, `ConfidenceLikely`, `ConfidenceUncertain`.

### Anti-Patterns to Avoid

| Don't | Do Instead | Why |
|-------|------------|-----|
| Call `DetectMultiple` and manually compare | Use `DetectWithCoexistenceSelection` | Service already handles timestamp comparison |
| Store coexistence in separate fields | Add to existing `Project` struct | Keeps detection state together |
| Create new warning style | Use existing `styles.WarningStyle` | Consistency with established patterns |
| Display full warning in project row | Use emoji indicator only | Row space is limited |
| Modify `SelectByTimestamp` | Handle display in TUI layer | Selection logic is correct, display is separate concern |

### References

| Document | Section |
|----------|---------|
| PRD Phase 2 | `docs/prd-phase2.md` - FR-P2-11: Display both methodologies on tie |
| Epic | `docs/epics-phase2.md` - Story 2.5 (mapped to 14.5) |
| Previous Story | Story 14.4 - CoexistenceWarning flag implementation |
| Architecture | `docs/architecture.md` - Hexagonal boundaries, TUI patterns |
| Project Context | `docs/project-context.md` - Story completion requirements |

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

None

### Completion Notes List

- Added 4 coexistence fields to Project domain model (CoexistenceWarning, CoexistenceMessage, SecondaryMethod, SecondaryStage) as runtime-only fields
- Modified TUI model's refreshProjectsCmd to use DetectWithCoexistenceSelection API
- Implemented coexistence logic: clear winner clears warning, tie case sets warning with secondary methodology info
- Added coexistence warning section to detail panel after Detection field
- Added warning emoji indicator to project row stage column when coexistence detected
- Audited TUI layer - only one Detect() call existed (in refreshProjectsCmd), now updated
- All unit tests added and passing (1338 total tests)
- Used StageTasks instead of StageEpic (which doesn't exist) in tests for BMAD secondary stage

#### Code Review Fixes Applied (2026-01-16)

- **H1 Fixed:** UTF-8 safe truncation in delegate.go - uses rune slicing instead of byte slicing to avoid cutting multi-byte emoji characters
- **M1 Fixed:** Added validation in detail_panel.go - only shows coexistence warning when SecondaryMethod is also populated (prevents empty string display)
- **M2 Fixed:** Added CoexistenceMessage assertion to model_refresh_test.go TestModel_RefreshWithCoexistence_TieCase
- **M4 Fixed:** Added TestProjectItemDelegate_CoexistenceIndicator_NarrowWidth test to verify stage column hidden at narrow widths
- Added TestDetailPanel_CoexistenceWarning_HiddenWhenSecondaryMethodEmpty edge case test
- Total tests: 1340 (2 new tests added)

### File List

- `internal/core/domain/project.go` - Added CoexistenceWarning, CoexistenceMessage, SecondaryMethod, SecondaryStage fields
- `internal/adapters/tui/model.go` - Updated refreshProjectsCmd to use DetectWithCoexistenceSelection
- `internal/adapters/tui/components/detail_panel.go` - Added coexistence warning section, added SecondaryMethod validation
- `internal/adapters/tui/components/delegate.go` - Added warning emoji to stage column for coexistence, fixed UTF-8 truncation
- `internal/adapters/tui/components/detail_panel_test.go` - Added TestDetailPanel_CoexistenceWarning_Shown/Hidden, HiddenWhenSecondaryMethodEmpty
- `internal/adapters/tui/components/delegate_test.go` - Added TestProjectItemDelegate_CoexistenceIndicator_Shown/Hidden/NarrowWidth
- `internal/adapters/tui/model_refresh_test.go` - Added TestModel_RefreshWithCoexistence_ClearWinner/TieCase with CoexistenceMessage assertion

