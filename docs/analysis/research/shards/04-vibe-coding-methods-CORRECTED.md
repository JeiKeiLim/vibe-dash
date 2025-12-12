# Vibe Coding Methodologies (CORRECTED)

**Shard 4 of Technical Research**

---

## Overview

This section provides accurate, corrected information about BMAD-Method and Speckit vibe coding methodologies, including their folder structures, workflow stages, and detection strategies.

---

## Speckit (Specification-Driven Development) - CORRECTED

### Folder Structure (CORRECTED)

**`.specify/` or `.speckit/` Directory:**
- **Purpose:** Framework infrastructure (NOT specifications)
- **Contents:**
  - `memory/` - Agent memory and context
  - `scripts/` - Cross-platform automation scripts
  - `templates/` - Reusable templates
  - `config.*` - Configuration files

**`specs/` Directory:**
- **Purpose:** Actual specification documents
- **Structure:**
```
specs/
├── 001-user-authentication/
│   ├── spec.md (or speckit.specify.md)
│   ├── plan.md (or speckit.plan.md)
│   ├── tasks.md (or speckit.tasks.md)
│   ├── implement.md (or speckit.implement.md)
│   ├── clarify.md (optional)
│   └── analyze.md (optional)
├── 002-notification-system/
│   ├── spec.md
│   ├── plan.md
│   ├── tasks.md
│   └── implement.md
└── 003-payment-integration/
    └── ...
```

**Naming Convention:**
- `NNN-feature-name/` format
- `NNN` = zero-padded three-digit sequence number
- `feature-name` = kebab-case descriptive name

**Sources:**
- https://github.com/github/spec-kit
- https://deepwiki.com/github/spec-kit/8.4-project-structure-and-configuration
- https://deepwiki.com/github/spec-kit/4.10-feature-creation-and-branch-management

---

### Workflow Stages

**1. Specification ("Specify"):**
- **Command:** `/speckit.specify`
- **Output:** `specs/NNN-feature/spec.md`
- **Purpose:** Define WHAT the feature should do
- **Contains:**
  - User scenarios
  - Constraints and requirements
  - Success criteria
  - Acceptance criteria

**Example:**
```markdown
# Specification: User Authentication

## What We're Building
- User login/logout functionality
- JWT-based authentication
- Password reset flow
- Multi-factor authentication (optional)

## Success Criteria
- Users can log in within 2 seconds
- Password storage meets OWASP standards
- Session persistence across browser restarts
```

---

**2. Planning ("Plan"):**
- **Command:** `/speckit.plan`
- **Output:** `specs/NNN-feature/plan.md`
- **Purpose:** Define HOW to implement
- **Contains:**
  - Technology choices
  - Architecture design
  - Integration points
  - Compliance requirements
  - Technical approach

**Example:**
```markdown
# Plan: User Authentication

## Technology Stack
- Backend: Node.js + Express
- Auth: Passport.js + JWT
- Database: PostgreSQL
- Password Hashing: bcrypt

## Architecture
- REST API endpoints: /auth/login, /auth/logout
- Middleware: JWT verification
- Database schema: users table with hashed passwords
```

---

**3. Task Breakdown ("Tasks"):**
- **Command:** `/speckit.tasks`
- **Output:** `specs/NNN-feature/tasks.md`
- **Purpose:** Break down into actionable tasks
- **Contains:**
  - Granular, reviewable tasks
  - Explicit acceptance criteria per task
  - File path suggestions
  - Test-driven development notes

**Example:**
```markdown
# Tasks: User Authentication

## Task 1: Database Schema
- Create users table migration
- Add indexes on email field
- Test: migration runs successfully

## Task 2: Password Hashing
- Implement bcrypt wrapper functions
- Test: hashing and verification work
- File: src/utils/password.js
```

---

**4. Implementation ("Implement"):**
- **Command:** `/speckit.implement`
- **Output:** `specs/NNN-feature/implement.md`
- **Purpose:** Code generation and tracking
- **Contains:**
  - Generated code snippets
  - Implementation notes
  - Human review feedback
  - Traceability to spec

---

### State Detection Strategy (CORRECTED)

```go
type SpeckitState struct {
    CurrentPhase   string   // "specify", "plan", "tasks", "implement"
    SpecNumber     int      // Extracted from folder name
    FeatureName    string   // Extracted from folder name
    SpecComplete   bool
    PlanComplete   bool
    TotalTasks     int
    CompletedTasks int
    CurrentTask    string
    LastModified   time.Time
}

func DetectSpeckit(projectPath string) (bool, error) {
    // Check for .specify/ or .speckit/ directory
    if dirExists(path.Join(projectPath, ".specify")) ||
       dirExists(path.Join(projectPath, ".speckit")) {
        return true, nil
    }
    
    // Check for specs/ directory
    if dirExists(path.Join(projectPath, "specs")) {
        return true, nil
    }
    
    return false, nil
}

func ParseSpeckitState(projectPath string) (*SpeckitState, error) {
    specsPath := path.Join(projectPath, "specs")
    
    // Find all spec directories (NNN-feature-name pattern)
    specDirs, err := filepath.Glob(path.Join(specsPath, "[0-9][0-9][0-9]-*"))
    if err != nil {
        return nil, err
    }
    
    // Get most recent spec
    var latestSpec string
    var latestTime time.Time
    
    for _, dir := range specDirs {
        modTime := getLastModifiedInDir(dir)
        if modTime.After(latestTime) {
            latestTime = modTime
            latestSpec = dir
        }
    }
    
    if latestSpec == "" {
        return nil, ErrNoSpecsFound
    }
    
    state := &SpeckitState{
        LastModified: latestTime,
    }
    
    // Extract spec number and name from directory
    // Example: "001-user-authentication" -> 1, "user-authentication"
    baseName := filepath.Base(latestSpec)
    parts := strings.SplitN(baseName, "-", 2)
    if len(parts) == 2 {
        specNum, _ := strconv.Atoi(parts[0])
        state.SpecNumber = specNum
        state.FeatureName = parts[1]
    }
    
    // Check which artifacts exist
    specFile := path.Join(latestSpec, "spec.md")
    if !fileExists(specFile) {
        specFile = path.Join(latestSpec, "speckit.specify.md")
    }
    state.SpecComplete = fileExists(specFile)
    
    planFile := path.Join(latestSpec, "plan.md")
    if !fileExists(planFile) {
        planFile = path.Join(latestSpec, "speckit.plan.md")
    }
    state.PlanComplete = fileExists(planFile)
    
    // Determine current phase based on artifacts
    if !state.SpecComplete {
        state.CurrentPhase = "specify"
    } else if !state.PlanComplete {
        state.CurrentPhase = "plan"
    } else if fileExists(path.Join(latestSpec, "tasks.md")) ||
              fileExists(path.Join(latestSpec, "speckit.tasks.md")) {
        state.CurrentPhase = "tasks"
        // Count tasks if file exists
        state.TotalTasks = countTasksInFile(path.Join(latestSpec, "tasks.md"))
    }
    
    if fileExists(path.Join(latestSpec, "implement.md")) ||
       fileExists(path.Join(latestSpec, "speckit.implement.md")) {
        state.CurrentPhase = "implement"
    }
    
    return state, nil
}
```

---

## BMAD-Method (Breakthrough Method for Agile AI-Driven Development)

### Folder Structure

**`.bmad/` Directory:**
- **Purpose:** All BMAD artifacts and state
- **Structure:**
```
.bmad/
├── artifacts/
│   ├── briefs/
│   │   └── product-brief.md
│   ├── prd.md
│   ├── architecture.md
│   ├── ux-spec.md (optional)
│   ├── epics/
│   │   └── epic-*.md
│   └── stories/
│       └── story-*.md
├── qa/
│   ├── assessments/
│   └── gates/
└── config/
    └── workflow-state.yaml
```

**Sources:**
- https://buildmode.dev/blog/mastering-bmad-method-2025/
- https://deepwiki.com/bmadcode/BMAD-METHOD/4-development-workflow
- https://github.com/bmadcode/BMAD-METHOD-v5/blob/main/docs/user-guide.md

---

### Workflow Stages

**1. Strategic Planning Phase:**
- **Agents:** Analyst, PM, Architect, UX, PO
- **Artifacts:**
  - `docs/prd.md` - Product Requirements Document
  - `docs/architecture.md` - Technical Architecture
  - `docs/ux-spec.md` - UX Specifications (optional)
  - `docs/briefs/` - Initial briefs and context

**2. Document Sharding & Transition:**
- **Agents:** PO, QA
- **Artifacts:**
  - `docs/epics/` - Epics broken from PRD
  - `docs/stories/` - User stories from epics
  - `docs/qa/assessments/` - Initial QA assessments

**3. Core Development Cycle:**
- **Agents:** Dev, QA, Test Architect, Scrum Master
- **Artifacts:**
  - `src/` - Source code
  - `tests/` - Test suites
  - `docs/epics/` - Completed epics with notes
  - `docs/stories/` - Completed stories with results
  - `docs/qa/gates/` - Quality gate results

**4. Quality Assurance & Release:**
- **Agents:** QA, PO
- **Artifacts:**
  - Release notes
  - Audit trail in docs/

---

### State Detection Strategy

```go
type BMADState struct {
    CurrentPhase   string   // "planning", "sharding", "development", "qa"
    LastArtifact   string   // Most recently modified artifact
    CompletedEpics []string
    ActiveStories  []string
    AgentWaiting   bool
    LastModified   time.Time
}

func DetectBMAD(projectPath string) (bool, error) {
    // Check for .bmad directory
    bmadPath := path.Join(projectPath, ".bmad")
    return dirExists(bmadPath), nil
}

func ParseBMADState(projectPath string) (*BMADState, error) {
    bmadPath := path.Join(projectPath, ".bmad")
    
    state := &BMADState{}
    
    // Check for workflow-state.yaml if exists (explicit state)
    stateFile := path.Join(bmadPath, "config/workflow-state.yaml")
    if fileExists(stateFile) {
        return parseWorkflowStateYAML(stateFile)
    }
    
    // Infer state from artifacts
    artifactsPath := path.Join(bmadPath, "artifacts")
    
    // Check for PRD
    prdFile := path.Join(artifactsPath, "prd.md")
    if fileExists(prdFile) {
        state.CurrentPhase = "planning"
        state.LastModified = getFileModTime(prdFile)
    }
    
    // Check for epics
    epicsPath := path.Join(artifactsPath, "epics")
    if dirExists(epicsPath) {
        epics, _ := filepath.Glob(path.Join(epicsPath, "*.md"))
        if len(epics) > 0 {
            state.CurrentPhase = "sharding"
            state.CompletedEpics = epics
        }
    }
    
    // Check for stories
    storiesPath := path.Join(artifactsPath, "stories")
    if dirExists(storiesPath) {
        stories, _ := filepath.Glob(path.Join(storiesPath, "*.md"))
        if len(stories) > 0 {
            state.CurrentPhase = "development"
            state.ActiveStories = parseActiveStories(stories)
        }
    }
    
    // Check for QA gates
    qaPath := path.Join(bmadPath, "qa/gates")
    if dirExists(qaPath) {
        gates, _ := filepath.Glob(path.Join(qaPath, "*.md"))
        if len(gates) > 0 {
            state.CurrentPhase = "qa"
        }
    }
    
    // Detect waiting state via heuristics (file timestamp + stage)
    state.AgentWaiting = detectWaitingHeuristic(bmadPath, state.CurrentPhase)
    
    // Get last modified artifact
    state.LastArtifact = getLastModifiedInDir(bmadPath)
    state.LastModified = getFileModTime(state.LastArtifact)
    
    return state, nil
}

func detectWaitingHeuristic(bmadPath string, currentPhase string) bool {
    // Heuristic: if no activity in >1 hour AND in interactive stage
    lastModified := getFileModTime(getLastModifiedInDir(bmadPath))
    if time.Since(lastModified) > 1*time.Hour {
        // Interactive stages where agent typically waits
        interactiveStages := []string{"planning", "sharding"}
        for _, stage := range interactiveStages {
            if stage == currentPhase {
                return true
            }
        }
    }
    return false
}
```

---

## Method Comparison

| Aspect | BMAD-Method | Speckit |
|--------|-------------|---------|
| **Focus** | Agile AI-assisted development | Specification-driven development |
| **Structure** | Agent-based phases | Sequential workflow stages |
| **Main Folder** | `.bmad/` | `.specify/` (framework) + `specs/` (specs) |
| **Artifacts Location** | `.bmad/artifacts/` | `specs/NNN-feature-name/` |
| **Workflow Files** | `prd.md`, `architecture.md`, `epics/`, `stories/` | `spec.md`, `plan.md`, `tasks.md`, `implement.md` |
| **AI Role** | Agents for each phase | Spec-to-code generation |
| **Feature Organization** | Epic → Stories | Numbered spec folders |
| **State Tracking** | `config/workflow-state.yaml` (optional) | Inferred from artifact existence |
| **Best For** | AI-assisted agile projects | Formal spec-driven projects |
| **Learning Curve** | Medium | Medium-High |
| **Team Size** | Any | Medium-Large preferred |

---

## Detection Strategy Comparison

### Speckit Detection (CORRECTED)

```go
func DetectMethod(projectPath string) (string, error) {
    // Check for Speckit
    if dirExists(path.Join(projectPath, ".specify")) ||
       dirExists(path.Join(projectPath, ".speckit")) ||
       dirExists(path.Join(projectPath, "specs")) {
        return "speckit", nil
    }
    
    // Check for BMAD
    if dirExists(path.Join(projectPath, ".bmad")) {
        return "bmad", nil
    }
    
    return "", ErrNoMethodDetected
}
```

### Key Differences in Detection

**Speckit:**
- Look for `.specify/` or `.speckit/` directory (framework)
- Look for `specs/` directory (specifications)
- Parse `specs/NNN-feature-name/` directories
- Check for `spec.md`, `plan.md`, `tasks.md` files

**BMAD:**
- Look for `.bmad/` directory
- Parse `.bmad/artifacts/` subdirectories
- Check for `prd.md`, `epics/`, `stories/` files
- Optional: Read `config/workflow-state.yaml` for explicit state

---

## Implementation Recommendations

### Plugin Interface (Method-Agnostic)

```go
type MethodDetector interface {
    Name() string
    Detect(projectPath string) (bool, error)
    GetWorkflowStage(projectPath string) (*WorkflowStage, error)
    GetProjectState(projectPath string) (*ProjectState, error)
}

// Speckit-specific implementation
type SpeckitDetector struct {}

func (s *SpeckitDetector) Detect(projectPath string) (bool, error) {
    return dirExists(path.Join(projectPath, ".specify")) ||
           dirExists(path.Join(projectPath, ".speckit")) ||
           dirExists(path.Join(projectPath, "specs")), nil
}

// BMAD-specific implementation
type BMADDetector struct {}

func (b *BMADDetector) Detect(projectPath string) (bool, error) {
    return dirExists(path.Join(projectPath, ".bmad")), nil
}
```

---

## Sources and References

**Speckit (CORRECTED):**
- https://github.com/github/spec-kit
- https://speckit.org/
- https://deepwiki.com/github/spec-kit/8.4-project-structure-and-configuration
- https://deepwiki.com/github/spec-kit/4.10-feature-creation-and-branch-management
- https://developer.microsoft.com/blog/spec-driven-development-spec-kit

**BMAD-Method:**
- https://buildmode.dev/blog/mastering-bmad-method-2025/
- https://deepwiki.com/bmadcode/BMAD-METHOD/4-development-workflow
- https://solodev.app/a-look-into-the-bmad-method
- https://www.geeky-gadgets.com/bmad-agile-ai-coding-method/
- https://github.com/bmadcode/BMAD-METHOD-v5/blob/main/docs/user-guide.md

---

**Shard Status:** ✅ Complete and Corrected  
**Last Updated:** 2025-12-04T06:59:43.181Z  
**Confidence Level:** High (authoritative sources, corrected based on user feedback)
