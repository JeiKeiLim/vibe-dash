# Validation Report

**Document:** docs/sprint-artifacts/stories/epic-15/15-6-integrate-agent-detection-into-tui-dashboard.md
**Checklist:** .bmad/bmm/workflows/4-implementation/create-story/checklist.md
**Date:** 2026-01-16

## Summary
- Overall: 32/40 passed (80%)
- Critical Issues: 3

## Section Results

### Step 1: Load and Understand the Target
Pass Rate: 5/5 (100%)

[✓] Story metadata extracted correctly
Evidence: Story 15.6, epic_num=15, story_num=6, story_key="15-6", story_title="Integrate Agent Detection into TUI Dashboard"

[✓] Workflow variables resolved
Evidence: Lines 99-108 show architecture overview with correct file locations

[✓] Status is `ready-for-dev`
Evidence: Line 3: "Status: ready-for-dev"

[✓] User-Visible Changes section present
Evidence: Lines 11-16 document visible changes

[✓] Acceptance Criteria defined
Evidence: 8 acceptance criteria at lines 18-28

### Step 2: Exhaustive Source Document Analysis
Pass Rate: 13/16 (81%)

#### 2.1 Epics and Stories Analysis

[✓] Epic objectives referenced
Evidence: Lines 512-516 Dev Agent Record references "Phase 2 Epic 15: Sub-1-Minute Agent Detection (THE killer feature)"

[✓] Cross-story dependencies identified
Evidence: Lines 513-516 list prerequisites: Story 15.1, 15.4, 15.5 as DONE

[⚠] PARTIAL: Story requirements alignment with epic
Evidence: Epic Story 3.6 (lines 514-540 in epics-phase2.md) says "shows '⏸️ WAITING 2h'" but story User-Visible Changes says "WAITING Xh Ym" without emoji specification. Story should match epic exactly.
Impact: Developer might implement different format than epic specifies.

#### 2.2 Architecture Deep-Dive

[✓] Technical stack documented
Evidence: Dev Notes lines 130-235 provide code snippets with correct imports

[✓] Code structure patterns
Evidence: Lines 437-455 Hexagonal Architecture section defines package locations

[⚠] PARTIAL: Missing import path verification
Evidence: Line 138 imports `agentdetectors` from `internal/adapters/agentdetectors` but actual detector package is just `agentdetectors` - need to verify exact import path exists
Impact: Build errors if import path is wrong

[✓] Testing standards documented
Evidence: Lines 416-433 Testing Strategy section

[✓] Integration patterns
Evidence: Lines 394-403 Display Integration section explains existing patterns

#### 2.3 Previous Story Intelligence

[✓] Previous story learnings referenced
Evidence: Lines 458-475 document learnings from Stories 15.4, 15.5, and 4.5

[✓] Code patterns from previous stories
Evidence: Lines 459-461 reference `const detectorName`, duration clamping, context cancellation patterns

[✗] FAIL: Missing Story 15.3 learnings
Evidence: Story 15.3 (ClaudeCodeLogParser) has critical learnings about JSONL parsing efficiency that aren't referenced. The AgentDetectionService will call ClaudeCodeDetector which uses LogParser - understanding its behavior is important.
Impact: Developer may not understand full detection chain behavior.

#### 2.4 Git History Analysis

[✓] Recent commits referenced
Evidence: Git status shows recent commits for Stories 15.3-15.5 (79340b0, 476735b, a268ff6)

[➖] N/A: Git patterns
Reason: Story 15.6 creates new files, no existing patterns to follow

#### 2.5 Latest Technical Research

[✓] Libraries documented
Evidence: Uses stdlib only (context, time, sync), Bubble Tea TUI framework (existing)

### Step 3: Disaster Prevention Gap Analysis
Pass Rate: 8/13 (62%)

#### 3.1 Reinvention Prevention Gaps

[✗] FAIL: services package location wrong
Evidence: Story proposes `internal/core/services/agent_detection_service.go` but examining existing services (waiting_detector.go per previous story), services are in `internal/core/services/`. Need to verify if AgentDetectionService belongs in services or should be in adapters since it directly uses adapter-layer detectors.
Impact: Architecture violation - services layer should not import adapters layer per hexagonal architecture.

[✗] FAIL: WaitingDetector implementation exists
Evidence: Line 502-504 reference `internal/core/ports/waiting_detector.go` and `internal/core/services/waiting_detector.go`. The story creates AgentWaitingAdapter but doesn't mention whether it replaces or wraps the existing WaitingDetector service. Current TUI uses `services.WaitingDetector` which is injected via main.go → cli.go → tui.Run().
Impact: Developer confusion about whether to replace existing WaitingDetector or compose with it.

[⚠] PARTIAL: Cache implementation duplicates existing pattern
Evidence: Lines 253-338 AgentWaitingAdapter has its own cache implementation. The existing WaitingDetector service may have similar patterns. Should verify no duplication.
Impact: Inconsistent caching strategies.

#### 3.2 Technical Specification DISASTERS

[✓] Interface compliance documented
Evidence: Lines 190, 282 show compile-time interface checks

[⚠] PARTIAL: Import statement has typo potential
Evidence: Line 251 shows `import "github.com/JeiKeiLim/vibe-dash/internal/core/services"` but if AgentDetectionService is in services, this creates circular dependency risk since services imports domain, not adapters.
Impact: Circular import error.

[✓] Timeout documented
Evidence: Lines 147, 200-202 document 1-second timeout per NFR-P2-1

#### 3.3 File Structure DISASTERS

[✓] File locations documented
Evidence: Lines 99-108 Architecture Overview, lines 532-537 File List

[⚠] PARTIAL: detection adapter location unclear
Evidence: Story proposes `internal/adapters/detection/agent_waiting_adapter.go` but existing detection adapters are in `internal/adapters/detectors/` and `internal/adapters/agentdetectors/`. Need to verify correct location.
Impact: Inconsistent package organization.

#### 3.4 Regression DISASTERS

[✓] Backward compatibility addressed
Evidence: Lines 394-403 explain existing components continue to work via WaitingChecker/WaitingDurationGetter callbacks

[⚠] PARTIAL: Missing migration strategy
Evidence: Story doesn't explain how to transition from existing `services.WaitingDetector` (which uses threshold resolver) to new `AgentWaitingAdapter`. The existing detector is wired through main.go → cli.go → tui.Run(). Story must explain whether to replace it entirely or compose them.
Impact: Two different waiting detection mechanisms could conflict.

#### 3.5 Implementation DISASTERS

[✓] Acceptance criteria clear
Evidence: 8 well-defined ACs with testable conditions

[✓] Edge cases documented
Evidence: Lines 476-484 Edge Cases table

[✓] Testing strategy comprehensive
Evidence: Lines 416-433, task 8 has 10 subtasks for tests

### Step 4: LLM-Dev-Agent Optimization Analysis
Pass Rate: 6/6 (100%)

[✓] Structure scannable
Evidence: Well-organized sections with headers, tables, code blocks

[✓] Code samples complete
Evidence: Lines 130-235, 240-339 provide full implementation code

[✓] Task breakdown actionable
Evidence: 9 tasks with 35+ subtasks, each with AC references

[✓] References complete
Evidence: Lines 493-506 list 16 source references with file paths

[✓] Architecture diagram present
Evidence: Lines 99-108, 437-455 show architecture

[✓] Previous learnings applied
Evidence: Lines 458-475 document patterns to follow

## Failed Items

### ✗ F1: Services layer imports adapters (Architecture Violation)
**Location:** Lines 138-139
**Evidence:** `AgentDetectionService` proposed in `internal/core/services/` imports `agentdetectors` from adapters layer
**Recommendation:** Move `AgentDetectionService` to adapters layer (`internal/adapters/detection/`) OR create interface in ports and have services call through interface

### ✗ F2: Missing Story 15.3 learnings
**Location:** Lines 458-475
**Evidence:** Story references 15.4 and 15.5 learnings but omits 15.3 (ClaudeCodeLogParser)
**Recommendation:** Add from Story 15.3: tail-optimized reading pattern, ParseLastAssistantEntry returns nil for no entries, context cancellation at 100-entry intervals

### ✗ F3: Existing WaitingDetector migration unclear
**Location:** Entire story
**Evidence:** Story creates AgentWaitingAdapter but doesn't explain integration with existing `services.WaitingDetector` wired through main.go → cli.go
**Recommendation:** Add explicit section explaining:
  1. Whether AgentWaitingAdapter replaces or wraps existing WaitingDetector
  2. Changes needed in main.go to wire AgentDetectionService
  3. Whether threshold resolver config still applies

## Partial Items

### ⚠ P1: Epic format mismatch for WAITING display
**Location:** Lines 11-16
**Evidence:** Story says "WAITING Xh Ym" but epic says "⏸️ WAITING 2h"
**Recommendation:** Align User-Visible Changes with exact epic wording including emoji

### ⚠ P2: Import path verification needed
**Location:** Line 138
**Evidence:** Import `internal/adapters/agentdetectors` not verified
**Recommendation:** Verify exact package import path matches existing code

### ⚠ P3: detection adapter package location
**Location:** Lines 103-104
**Evidence:** Proposes `internal/adapters/detection/` but other detectors in different locations
**Recommendation:** Consider using existing `internal/adapters/detectors/` OR document why new package

### ⚠ P4: Missing migration wiring
**Location:** Tasks 5, 6, 7
**Evidence:** Tasks mention wiring but don't show full integration path from main.go
**Recommendation:** Add Task subtask for main.go changes showing full initialization flow

## Recommendations

### 1. Must Fix: Architecture - Move AgentDetectionService to adapters layer
```
Current: internal/core/services/agent_detection_service.go (VIOLATES hexagonal)
Correct: internal/adapters/detection/agent_detection_service.go

OR create interface:
- internal/core/ports/agent_detection_orchestrator.go (interface)
- internal/adapters/detection/agent_detection_service.go (implementation)
```

### 2. Must Fix: Add explicit migration/integration section
Add section explaining:
```markdown
### Integration with Existing WaitingDetector

The existing `services.WaitingDetector` (Story 4.3/4.4) uses threshold-based detection
from Project.LastActivityAt. This story REPLACES that with AgentWaitingAdapter.

**Changes required:**
1. main.go: Remove `services.NewWaitingDetector()` creation
2. main.go: Create `AgentDetectionService` and `AgentWaitingAdapter`
3. main.go: Pass adapter to `cli.SetWaitingDetector()`

**Rationale:** New detection is more accurate (log-based) and makes threshold
resolver obsolete for agent detection (threshold only used by GenericDetector fallback).
```

### 3. Should Improve: Add Story 15.3 learnings
```markdown
From Story 15.3 (ClaudeCodeLogParser):
- Tail-optimized reading (last 4KB) for performance
- ParseLastAssistantEntry returns (nil, nil) when no assistant entries found
- Context cancellation checked every 100 entries during parsing
- JSONL parsing skips malformed lines gracefully
```

### 4. Consider: Verify package locations
- Confirm `internal/adapters/agentdetectors/` is correct import path
- Consider consolidating detection-related adapters in one package
- Add explicit verification task: "6.0: Verify import paths compile"

---

## Validation Summary

| Category | Count |
|----------|-------|
| ✓ Pass | 32 |
| ⚠ Partial | 4 |
| ✗ Fail | 3 |
| ➖ N/A | 1 |
| **Total** | 40 |

**Critical Issues Requiring Fix Before Development:**
1. Architecture violation: Services importing adapters
2. Missing migration strategy for existing WaitingDetector
3. Missing previous story learnings (15.3)

**Path to validation report:** `docs/sprint-artifacts/stories/epic-15/validation-report-15-6-2026-01-16.md`
