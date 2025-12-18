# Validation Report

**Document:** `/docs/sprint-artifacts/stories/epic-3.5/3-5-6-update-cli-commands.md`
**Checklist:** `/bmm/workflows/4-implementation/create-story/checklist.md`
**Date:** 2025-12-18

## Summary
- Overall: 28/35 items passed (80%)
- Critical Issues: 5

## Section Results

### Section 1: Story Structure and Completeness
Pass Rate: 7/7 (100%)

[✓] **Story has clear user story format**
Evidence: Lines 5-9: "As a user, I want CLI commands to work with the new storage structure, So that I can add, list, and remove projects as before."

[✓] **Acceptance Criteria clearly defined**
Evidence: Lines 11-27: 7 ACs covering all CLI commands and test requirements.

[✓] **Tasks broken into subtasks**
Evidence: Lines 29-139: 9 tasks with detailed subtasks totaling 50+ subtasks.

[✓] **Dev Notes section comprehensive**
Evidence: Lines 141-450: Extensive architecture, design decisions, code patterns, and references.

[✓] **Status is ready-for-dev**
Evidence: Line 3: "Status: ready-for-dev"

[✓] **Dependencies documented**
Evidence: Lines 403-411: Clear "Depends on (COMPLETED)" and "Required by" sections.

[✓] **References section complete**
Evidence: Lines 442-450: 10 references to source documents and code files.

---

### Section 2: Technical Requirements Completeness
Pass Rate: 5/8 (63%)

[✓] **Current main.go analysis included**
Evidence: Lines 212-258: Shows current code and required changes with diff-style comparison.

[✓] **Files to modify list complete**
Evidence: Lines 262-277: Table with File/Action/Purpose columns.

[⚠] **PARTIAL: Missing GetDefaultBasePath helper implementation**
Evidence: Line 270 mentions `config/paths.go` should have `GetDefaultBasePath()` but the file doesn't exist. Story says "CREATE/MODIFY" but doesn't provide implementation details.
Impact: Dev agent may waste time figuring out where to put this helper.

[✗] **FAIL: Missing ports.DirectoryManager interface update for DeleteProjectDir**
Evidence: Task 4 Subtask 4.4 (line 89-91) mentions "Alternative approach - Add DeleteProjectDir method to DirectoryManager" but:
  - The `ports.DirectoryManager` interface (ports/directory.go) does NOT have this method
  - Story doesn't specify if this should be added to the interface
  - No code example for the port interface update provided
Impact: Dev agent may not add the interface method, causing compile errors when implementing.

[✓] **Error handling patterns documented**
Evidence: Lines 363-371: Error handling table with scenarios, types, and handling strategies.

[⚠] **PARTIAL: Incomplete configPathAdapter implementation**
Evidence: Lines 46-59 and 194-204 show adapter but:
  - Missing import statements for the adapter
  - Not clear where this adapter should live (in main.go? separate file?)
  - Missing test coverage requirements for adapter
Impact: Dev agent may implement adapter incorrectly or in wrong location.

[✗] **FAIL: Missing safety check for basePath emptiness**
Evidence: Line 241 creates DirectoryManager with `basePath` but doesn't validate:
  - What happens if `config.GetDefaultBasePath()` returns empty?
  - No error handling shown for nil DirectoryManager (NewDirectoryManager returns nil on failure)
Impact: Potential nil pointer panic if basePath not determined.

[⚠] **PARTIAL: Task 4 has conflicting approaches**
Evidence: Subtask 4.3-4.4 (lines 75-91) provide TWO different implementations for directory deletion without clear guidance on which to use. One inlines in runRemove(), other adds DirectoryManager method.
Impact: Dev agent may implement both or wrong one.

---

### Section 3: Previous Story Intelligence
Pass Rate: 6/7 (86%)

[✓] **Previous story (3.5.5) learnings referenced**
Evidence: Lines 413-419: Explicitly lists 4 patterns to follow from Story 3.5.5.

[✓] **Context cancellation pattern documented**
Evidence: Lines 283-288: Shows required `select { case <-ctx.Done(): }` pattern.

[✓] **Error wrapping pattern documented**
Evidence: Lines 290-291: `fmt.Errorf("failed to X: %w", err)` pattern.

[✓] **Graceful degradation pattern documented**
Evidence: Lines 293-304: Shows safety check and slog.Warn for non-fatal errors.

[✓] **Thread safety notes included**
Evidence: Lines 358-360: Notes CLI runs sequentially, no extra locking needed.

[✓] **Code review feedback from 3.5.5 incorporated**
Evidence: Lines 413-419 mention config update in Delete (which was a code review fix in 3.5.5).

[⚠] **PARTIAL: Missing architectural compliance check**
Evidence: Story doesn't mention checking that CLI package doesn't import from core incorrectly. Project-context.md emphasizes "NEVER let core import from adapters" - reverse should also be validated.
Impact: Minor - CLI is in adapters, should be fine.

---

### Section 4: Disaster Prevention - Reinvention Analysis
Pass Rate: 4/6 (67%)

[✓] **Uses existing RepositoryCoordinator**
Evidence: Line 39: "Create RepositoryCoordinator" and throughout story uses coordinator.

[✓] **Uses existing DirectoryManager**
Evidence: Line 36: "Create DirectoryManager" wired from filesystem package.

[✓] **Uses existing ViperLoader**
Evidence: Line 33: "Create ViperLoader (already exists)".

[✗] **FAIL: Duplicate collision handling not addressed**
Evidence: `add.go` (lines 137-169) has its own collision handling logic (`generateUniqueName`, `promptCollisionResolution`). Story Subtask 3.2 (line 65) says "Remove duplicate collision handling from add.go if redundant" but:
  - Doesn't analyze whether it IS redundant
  - DirectoryManager handles directory NAME collision, not display NAME collision
  - These are DIFFERENT - DirectoryManager is for filesystem, add.go is for user-facing names
  - Story doesn't clarify this distinction
Impact: Dev agent may incorrectly remove display name collision handling.

[⚠] **PARTIAL: findProjectByName duplication not addressed**
Evidence: `remove.go:60-80` has `findProjectByName()` which is similar to `checkNameCollision()` in `add.go:209-223`. Comment at lines 63-67 notes similarity. Story doesn't mention refactoring this.
Impact: Minor tech debt continues.

[⚠] **PARTIAL: Missing basePath helper reuse analysis**
Evidence: Line 236 shows `basePath := config.GetDefaultBasePath()` but this function doesn't exist. Should check if `ViperLoader.configPath` derivation logic can be reused.
Impact: Dev agent may reimplement path logic differently.

---

### Section 5: LLM Dev Agent Optimization
Pass Rate: 6/7 (86%)

[✓] **Tasks numbered clearly**
Evidence: Tasks 1-9 numbered with subtasks using decimal notation (1.1, 1.2, etc.)

[✓] **Code examples provided**
Evidence: Multiple code blocks throughout (lines 34-42, 47-59, 75-88, etc.)

[✓] **Architecture diagram included**
Evidence: Lines 146-174: ASCII diagram showing component relationships.

[✓] **AC mapping in tasks**
Evidence: Each task header includes "(AC: X, Y)" notation.

[✓] **Manual testing checklist provided**
Evidence: Lines 385-399: Step-by-step manual testing checklist.

[✓] **File table with actions**
Evidence: Lines 262-277: Clear table format.

[⚠] **PARTIAL: Verbosity in Task 4**
Evidence: Task 4 (lines 72-95) provides multiple alternative approaches without clear decision. This wastes LLM tokens processing alternatives instead of single clear path.
Impact: Dev agent may request clarification or implement wrong approach.

---

## Failed Items

### 1. Missing DeleteProjectDir interface method
**Severity:** HIGH
**Location:** Task 4 Subtask 4.4 (lines 89-91)
**Issue:** Story mentions adding `DeleteProjectDir` to DirectoryManager but doesn't update the `ports.DirectoryManager` interface definition.
**Recommendation:** Add explicit task to update `internal/core/ports/directory.go` with:
```go
// DeleteProjectDir removes the project directory and all its contents.
// Returns nil if directory doesn't exist (idempotent).
DeleteProjectDir(ctx context.Context, projectPath string) error
```

### 2. Missing basePath nil safety
**Severity:** HIGH
**Location:** Task 1 Subtask 1.4 (line 35)
**Issue:** `NewDirectoryManager` returns nil if basePath cannot be determined. main.go code doesn't check for nil.
**Recommendation:** Add nil check before using dirMgr:
```go
dirMgr := filesystem.NewDirectoryManager(basePath, configAdapter)
if dirMgr == nil {
    return fmt.Errorf("failed to initialize directory manager: cannot determine base path")
}
```

### 3. Missing GetDefaultBasePath helper
**Severity:** MEDIUM
**Location:** Lines 236, 270
**Issue:** Code references `config.GetDefaultBasePath()` which doesn't exist.
**Recommendation:** Create helper in `internal/config/paths.go`:
```go
// GetDefaultBasePath returns the default vibe-dash storage directory.
// Returns ~/.vibe-dash on success, empty string on home dir lookup failure.
func GetDefaultBasePath() string {
    home, err := os.UserHomeDir()
    if err != nil {
        return ""
    }
    return filepath.Join(home, ".vibe-dash")
}
```

### 4. Unclear collision handling distinction
**Severity:** MEDIUM
**Location:** Subtask 3.2 (line 65)
**Issue:** Story says "Remove duplicate collision handling" but doesn't clarify that DirectoryManager collision (filesystem) is DIFFERENT from add.go collision (display names).
**Recommendation:** Update story to clarify:
- DirectoryManager: handles filesystem directory name collisions (parent-prefixing)
- add.go: handles user-facing display name collisions (unrelated to filesystem)
- These are NOT duplicates and BOTH should remain

### 5. Conflicting Task 4 approaches
**Severity:** MEDIUM
**Location:** Subtasks 4.3-4.5 (lines 75-95)
**Issue:** Provides inline approach AND DirectoryManager method approach without recommendation.
**Recommendation:** Choose one approach. Recommend: DirectoryManager.DeleteProjectDir for consistency with EnsureProjectDir pattern.

---

## Partial Items

### 1. configPathAdapter location unclear
**Location:** Lines 46-59, 194-204
**Gap:** Adapter code shown but file location not specified.
**Fix:** Specify "Create in cmd/vibe/main.go as unexported type" or extract to separate file.

### 2. Task 4 verbosity
**Location:** Lines 72-95
**Gap:** Too many alternatives dilute focus.
**Fix:** Pick one approach, remove alternatives.

### 3. findProjectByName tech debt
**Location:** Comment at remove.go:63-67
**Gap:** Known duplication not addressed.
**Fix:** Add "Nice to have: refactor shared helper" note.

---

## Recommendations

### 1. Must Fix (Critical)

1. **Add DeleteProjectDir interface method** - Add to ports.DirectoryManager interface with proper godoc
2. **Add basePath nil safety** - Check NewDirectoryManager result in main.go
3. **Create GetDefaultBasePath helper** - Add to internal/config/paths.go
4. **Clarify collision handling** - Document that display name collision and directory collision are separate
5. **Choose single Task 4 approach** - Recommend DirectoryManager.DeleteProjectDir method

### 2. Should Improve

1. **Add import statements** - Show full import list for main.go changes
2. **Specify configPathAdapter location** - Explicitly state it goes in main.go
3. **Add FilesystemDirectoryManager.DeleteProjectDir** - Implementation code example

### 3. Consider (Nice to Have)

1. **Refactor findProjectByName** - Extract shared helper from add.go and remove.go
2. **Add test for nil DirectoryManager** - Edge case test
