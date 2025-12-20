# Sprint Change Proposal: BMAD Method State Detection

**Date:** 2025-12-20
**Proposed By:** Jongkuk Lim
**Status:** APPROVED
**Approval Date:** 2025-12-20

---

## Executive Summary

Pull BMAD Method State Detection from Post-MVP (Month 2-3) into the current sprint as Epic 4.5, to be executed immediately after Epic 4 and before Epic 5.

**Rationale:** vibe-dash is being built using the BMAD Method. Implementing BMAD detection now enables "eating your own dog food" - validating the dashboard while actively using it for development.

---

## Change Trigger

| Aspect | Details |
|--------|---------|
| **Type** | Strategic addition (pull-forward from post-MVP) |
| **Trigger Point** | Transition from Epic 4 to Epic 5 |
| **Discovery Context** | Recognized during sprint planning that dogfooding opportunity exists |
| **Original Timeline** | Post-MVP Month 2-3 (Growth Features) |
| **New Timeline** | Immediate (Epic 4.5) |

---

## Impact Analysis

### Epic Impact

| Epic | Impact |
|------|--------|
| Epic 1-4, 3.5 | None (completed) |
| Epic 4.5 | **NEW** - BMAD Method State Detection |
| Epic 5-7 | Deferred slightly, no scope changes |

### Artifact Impact

| Artifact | Change Required |
|----------|-----------------|
| PRD | Move BMAD detection from Growth Features to MVP |
| Epics | Add Epic 4.5 with 3 stories |
| Architecture | None - already designed for pluggable detectors |
| UX Design | None - methodology-agnostic display |
| Sprint Status | Add Epic 4.5 entries |
| Tests | Add BMAD-specific test fixtures |

### Risk Assessment

| Risk Factor | Level | Mitigation |
|-------------|-------|------------|
| Technical complexity | Low | MethodDetector pattern proven by Speckit |
| Timeline impact | Low | 3 focused stories, ~1-2 days each |
| Architecture changes | None | Interface already exists |
| Dependencies | None | Self-contained implementation |

---

## Proposed Epic 4.5: BMAD Method State Detection

### Epic Definition

**Goal:** Implement BMAD Method v6 detection as a second MethodDetector plugin, enabling vibe-dash to detect and display workflow state for projects using the BMAD Method.

**Scope:** BMAD v6 only (`.bmad/` folder structure). v4 support (`.bmad-core/`) deferred - can be added later if demand emerges.

**Success Criteria:**
- BMAD detector correctly identifies `.bmad/` folder structure
- Stage detection matches BMAD workflow phases (Ideation, Planning, Implementation, etc.)
- vibe-dash can display its own development state accurately
- All existing tests pass, new BMAD-specific tests added

### Stories

| ID | Title | Description | Complexity |
|----|-------|-------------|------------|
| 4.5-1 | BMAD v6 Detector Implementation | Implement MethodDetector interface for BMAD Method v6, detecting `.bmad/` folder presence and `bmm/config.yaml` | Medium |
| 4.5-2 | BMAD v6 Stage Detection Logic | Implement stage detection via `sprint-status.yaml` parsing (epic/story status fields) | Medium |
| 4.5-3 | BMAD Test Fixtures | Create test fixtures using vibe-dash's own `.bmad/` folder as real-world v6 test case | Low |

**Out of Scope (Future):**
- BMAD v4 support (`.bmad-core/` folder) - add if demand emerges

### Technical Approach

```
internal/adapters/detectors/
├── speckit/           # Existing - proven pattern
│   └── detector.go
└── bmad/              # New - follows same pattern
    ├── detector.go    # Implements MethodDetector interface
    ├── stages.go      # BMAD-specific stage detection
    └── detector_test.go
```

**Detection Strategy:**
1. Check for `.bmad/` folder at project root
2. Look for `config.yaml`, workflow files, sprint artifacts
3. Parse `sprint-status.yaml` for current epic/story status
4. Map to standardized stage indicators (Plan, Implement, Review, etc.)

---

## Updated Sprint Order

```
Before:
Epic 4 (Done) → Epic 5 → Epic 6 → Epic 7

After:
Epic 4 (Done) → Epic 4.5 (NEW) → Epic 5 → Epic 6 → Epic 7
```

---

## Handoff Plan

### Responsibilities

| Role | Action |
|------|--------|
| **SM (Bob)** | Create Epic 4.5 in epics.md, update sprint-status.yaml, draft stories |
| **Dev** | Implement stories via dev-story workflow |
| **Code Review** | Review each story before marking done |

### Execution Sequence

1. **SM:** Add Epic 4.5 definition to `docs/epics.md`
2. **SM:** Update `docs/sprint-artifacts/sprint-status.yaml` with Epic 4.5 entries
3. **SM:** Draft Story 4.5-1 → Mark ready-for-dev
4. **Dev:** Implement Story 4.5-1 → Code review → Done
5. **SM:** Draft Story 4.5-2 → Mark ready-for-dev
6. **Dev:** Implement Story 4.5-2 → Code review → Done
7. **SM:** Draft Story 4.5-3 → Mark ready-for-dev
8. **Dev:** Implement Story 4.5-3 → Code review → Done
9. **SM:** Run Epic 4.5 retrospective (optional for small epic)
10. **Continue:** Resume with Epic 5

---

## Approval

| Approver | Decision | Date |
|----------|----------|------|
| Jongkuk Lim | **APPROVED** | 2025-12-20 |

---

## Next Immediate Action

**Agent:** SM (Bob)
**Action:** Add Epic 4.5 definition to `docs/epics.md` and update sprint status

---

---

## Appendix: BMAD Detection Research

**Research Date:** 2025-12-20

### Existing Implementation Reference

**Repository:** https://github.com/ibadmore/bmad-progress-dashboard
- Supports both BMAD v4 and v6 (auto-detect)
- Analyzes story files with task checkboxes
- Calculates progress: 40% planning + 60% development

### BMAD v6 Structure (from vibe-dash .bmad/)

```
.bmad/
├── _cfg/                    # IDE/agent configurations
├── core/                    # Core BMAD infrastructure
├── docs/                    # BMAD documentation
└── bmm/                     # BMAD Method Module
    ├── config.yaml          # Version: 6.0.0-alpha.13
    ├── agents/              # Agent definitions
    └── workflows/           # Phase-organized workflows
        ├── 1-analysis/      # Research, product-brief
        ├── 2-plan-workflows/# PRD, UX design
        ├── 3-solutioning/   # Architecture, epics-and-stories
        └── 4-implementation/# dev-story, code-review, sprint-planning
```

### Primary Detection Artifacts

| Artifact | Location | Purpose |
|----------|----------|---------|
| sprint-status.yaml | `docs/sprint-artifacts/` | Epic/story status tracking |
| bmm/config.yaml | `.bmad/bmm/` | Version identification |
| Phase folders | `.bmad/bmm/workflows/` | Methodology phase detection |
| Output docs | `docs/` | PRD, architecture, epics presence |

### sprint-status.yaml Structure

```yaml
project_key: vibe-dash
development_status:
  epic-1: done
  epic-2: done
  epic-3: done
  epic-3.5: done
  epic-4: done      # Current: Just completed
  epic-5: backlog   # Next up
```

**Status Values:**
- Epic: `backlog`, `in-progress`, `done`
- Story: `backlog`, `drafted`, `ready-for-dev`, `in-progress`, `review`, `done`

### Detection Algorithm (Proposed)

```go
func (d *BmadDetector) Detect(ctx context.Context, path string) (*DetectionResult, error) {
    // 1. Check for .bmad folder
    bmadPath := filepath.Join(path, ".bmad")
    if !exists(bmadPath) {
        return nil, ErrNotBmadProject
    }

    // 2. Read version from bmm/config.yaml
    version := readBmmConfig(path)

    // 3. Find and parse sprint-status.yaml
    status := findSprintStatus(path) // docs/sprint-artifacts/ or output_folder

    // 4. Determine current phase from in-progress epic
    phase := determinePhase(status)

    // 5. Map to standardized stage
    return &DetectionResult{
        Method: "bmad",
        Stage: mapPhaseToStage(phase),
        Confidence: calculateConfidence(status),
        Reasoning: buildReasoning(status, phase),
    }
}
```

### Stage Mapping

| BMAD Phase | vibe-dash Stage | Indicator |
|------------|-----------------|-----------|
| 1-analysis | Research/Plan | No epics created |
| 2-planning | Plan | PRD exists, no architecture |
| 3-solutioning | Plan/Specify | Architecture exists |
| 4-implementation | Implement/Review | Epics in progress |
| All epics done | Validate | All epics marked done |

### Reference Implementation Analysis

**Source:** `github.com/ibadmore/bmad-progress-dashboard` (cloned and analyzed)

**Detection Logic (from update-progress.js):**

```javascript
// Config locations tried (in order)
const configPaths = [
  'tools/core-config.yaml',
  'tools/bmad-config.yaml',
  '.bmad-core/core-config.yaml'
];

// Planning artifacts
const planning = {
  brief: exists('docs/brief.md') || exists('docs/project-brief.md'),
  prd: exists('docs/prd.md') || hasFiles('docs/prd'),
  architecture: exists('docs/architecture.md') || hasFiles('docs/architecture')
};

// Story detection - counts checkboxes
function parseStoryFile(filepath) {
  const content = fs.readFileSync(filepath, 'utf8');
  let totalTasks = 0, completedTasks = 0;
  for (const line of lines) {
    if (line.includes('- [ ]')) totalTasks++;
    if (line.includes('- [x]')) { totalTasks++; completedTasks++; }
  }
  return { totalTasks, completedTasks, percentage };
}

// Progress calculation
overall = (planning * 0.4) + (development * 0.6);
```

**Key Differences from vibe-dash v6:**

| Aspect | bmad-progress-dashboard | vibe-dash .bmad v6 |
|--------|-------------------------|-------------------|
| Config location | `.bmad-core/` | `.bmad/bmm/config.yaml` |
| Stories location | `docs/stories/` | `docs/sprint-artifacts/stories/` |
| Progress tracking | Task checkbox counting | `sprint-status.yaml` status fields |
| Implementation | JavaScript/Node.js | Go |

**Recommended Approach for vibe-dash BMAD Detector:**

1. **Detection:** Check for `.bmad/` folder (not `.bmad-core/`)
2. **Version:** Read `bmm/config.yaml` for version info
3. **State:** Parse `sprint-status.yaml` for epic/story status (simpler than checkbox counting)
4. **Stage mapping:** Map epic status to standardized stages

---

*Generated by Correct Course Workflow - BMad Method*
