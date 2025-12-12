---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]
lastStep: 11
inputDocuments:
  - 'docs/analysis/product-brief-bmad-test-2025-12-05.md'
  - 'docs/analysis/brainstorming-session-2025-12-04T00:25:43.430Z.md'
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
workflowType: 'prd'
lastStep: 0
project_name: 'bmad-test'
user_name: 'Jongkuk Lim'
date: '2025-12-05'
---

# Product Requirements Document - bmad-test

**Author:** Jongkuk Lim
**Date:** 2025-12-05

---

## Executive Summary

**Vibe Dashboard** is a CLI-first dashboard that serves as the central nervous system for developers managing multiple AI-assisted vibe coding projects. It automatically tracks workflow state across projects using structured methodologies (BMAD-Method, Speckit), intelligently surfaces only active work while hibernating dormant projects, and proactively signals when human attention is needed.

**Target Users:** Solo developers using structured vibe coding methodologies (BMAD-Method, Speckit) with AI coding agents across 3-10 simultaneous projects.

**Core Problem Solved:** Developers lose mental context when switching between AI-assisted projects. The problem isn't code context (AI agents handle that) - it's **workflow methodology context**: WHERE am I in the methodology workflow? What stage was I at? What action comes next? Current reality means minutes lost to context reconstruction every time developers switch projects.

**Vision:** Eliminate the "what was I doing?" moments by providing instant context restoration. The dashboard becomes the environment developers live in - always visible, automatically updated, signaling when action is needed.

### What Makes This Special

**1. Artifacts Are Always Truth**
Unlike manual tracking tools, Vibe Dashboard NEVER asks users to maintain state. It reads directly from .bmad/ and specs/ artifacts. When uncertain, shows uncertainty transparently. No `set-stage` commands - refresh forces re-scan from artifacts. This principle ensures 95%+ detection accuracy without user maintenance burden.

**2. Hibernation Model - Natural Two-State System**
Borrowed from nature: automatic active vs hibernated states that mirror natural human mental model (working on now vs worked on before). No manual organization, tagging, or project management overhead. System learns from activity patterns and auto-promotes projects when you start working on them.

**3. Agent Waiting State Detection (Killer Feature)**
Track when AI coding agent is blocked waiting for human input. Dashboard shows distinct indicator (⏸️ WAITING) when agent idle, preventing lost workflow momentum and ensuring no project gets forgotten while agent waits.

**4. Signal Over Noise Philosophy**
Reverse thinking breakthrough: don't show everything (overwhelming), show only what needs attention (actionable). Smart filtering keeps active projects limited to working memory capacity (5-7 items). Dashboard becomes your command center, not a tool you occasionally check.

**5. Production-Grade Performance**
Go + Bubble Tea stack proven by k9s, kubectl, docker CLI. Single binary deployment, cross-platform, near-instant startup, real-time updates via fsnotify file monitoring.

## Project Classification

**Technical Type:** CLI Tool (Command-Line Interface)

**Domain:** Developer Tool

**Complexity:** Medium

This is a developer-focused CLI tool built for technical users comfortable with terminal environments. The medium complexity stems from:
- Real-time file system monitoring (fsnotify)
- Intelligent heuristic-based stage detection (95% accuracy target)
- Cross-platform support (Linux, macOS, Windows)
- Interactive TUI with keyboard-driven navigation
- SQLite-based state management (per-project `.vibe/` directory)
- Pluggable architecture for multiple vibe coding methodologies

**Key CLI Characteristics:**
- Primary interface: `vibe` command displays dashboard (no subcommands needed)
- Interactive TUI with keyboard shortcuts: [a] Add [h] Hibernated [r] Refresh [?] Help [q] Quit
- Zero configuration required - works immediately after `vibe add .`
- Scriptable output for automation potential
- Shell integration for seamless developer workflow

**Success Metrics:**
- GitHub stars as primary adoption indicator
- 95% stage detection accuracy (technical validation)
- Competition: human memory + tools like ccmanager (no methodology-specific alternatives)

---

## Success Criteria

### User Success

**Primary Success Indicators:**

1. **Jeff's Morning Test** - Dashboard becomes the first command users type after opening terminal
   - Success metric: Users report `vibe` is their default startup command
   - Validation: Dashboard stays open in terminal split throughout workday

2. **Context Reconstruction Speed**
   - Current state: 5+ minutes manually investigating .bmad folders
   - Target state: Under 10 seconds from `vibe` command to "I know what to do next"
   - Validation: Users stop manually navigating to project folders

3. **Scale Breakthrough** 
   - Users can manage **10+ vibe coding projects** simultaneously (up from cognitive limit of 5-7)
   - Success quote: "I can manage more than 10 vibe coding projects thanks to vibe dashboard"

4. **Zero Wasted Attention**
   - Users don't waste time watching agents work or wondering if agents are waiting
   - Dashboard proactively signals when human attention needed
   - Success quote: "I don't waste time watching over coding agent doing their job or agent waiting for me"

5. **Focus Shift to High-Value Work**
   - Users spend more time reviewing agent-created documents instead of status tracking
   - Cognitive load shifts from "where am I?" to "what's next?"
   - Success quote: "I can focus more on looking over the documents that agent created"

6. **Trust Level Achievement**
   - Users trust dashboard detection and stop manually checking project state
   - Zero "what was I doing?" moments
   - Dashboard becomes reliable source of truth

7. **Adoption Validation**
   - Users run `vibe` daily even for single project management (validates core value prop)
   - Success threshold: User continues using dashboard after 1 week
   - "I can't work without it anymore" testimonial level

### Business Success

**3-Month Success (Validation Phase):**
- **Community Proof**: 3-5 real user testimonials ("this changed how I work")
- **Organic Discovery**: At least 1 online post from real user introducing Vibe Dashboard
- **Active Users**: 50+ users running `vibe` daily
- **Fan Validation**: Evidence of someone championing the tool
- **GitHub Stars**: 100 stars (early adopter validation)

**12-Month Success (Established Product):**
- **Scale**: Multiple testimonials and online posts from real users
- **GitHub Stars**: 1,000+ stars (community approval metric)
- **Active User Base**: 1,000+ daily active users
- **Retention**: 70%+ users still active 30 days after first use
- **Community Health**: 
  - Active issue/PR engagement
  - Community-created tutorials or blog posts
  - Tool mentioned in BMAD-Method/Speckit documentation

**Success Definition:**
- "Active user" = Running `vibe` daily (validates daily utility)
- Primary metric = GitHub stars (proves community approval)
- Secondary metric = User testimonials (validates problem solved)
- Maintained by solo developer for Year 1 (realistic scope)

### Technical Success

**Must-Haves (Launch Blockers):**

1. **95% Stage Detection Accuracy**
   - Golden path test suite: 20 real BMAD/Speckit projects with known stages
   - False positive rate: <5%
   - **Blocker status**: Users won't trust tool with wrong state detection
   - Validation: Automated test suite passes before launch

2. **Zero Configuration Friction**
   - `vibe add .` works immediately without setup
   - **Critical architectural decision**: Centralized `~/.vibe/` directory (not per-project)
   - Rationale: Avoid forcing developers to manually add `.vibe/` to `.gitignore`
   - Projects tracked centrally with per-project SQLite databases

3. **Production-Grade Performance**
   - Dashboard render: <100ms for 20 projects
   - File change detection: 5-10 seconds (fsnotify)
   - Startup time: Near-instant (<1 second)
   - Cross-platform: Works reliably on Linux, macOS, Windows

4. **Agent Waiting State Detection (Killer Feature)**
   - MVP threshold: 10 minutes of inactivity = agent waiting
   - Visual indicator: `⏸️ WAITING - Agent needs your input (Xh)`
   - Future refinement: Sub-1-minute detection based on user feedback
   - Rationale: This is the killer feature - ship it and iterate based on real usage

**MVP Features (Needs Improvement):**

1. **Hibernation System**
   - Auto-hibernation: Projects untouched 7-14 days (configurable)
   - Auto-promotion: Starting work on hibernated project brings it back
   - Always visible count: "X active, Y hibernated" prevents data-loss panic

2. **Method Detection**
   - MVP: Speckit detection with 95% accuracy
   - Future: BMAD-Method and additional methodologies via plugin architecture

**Technical Validation Requirements:**
- Cross-platform testing before launch (Linux, macOS, Windows)
- Golden path test suite passes at 95%+ accuracy
- Performance benchmarks met (<100ms render, <10s file detection)
- Zero critical bugs in core detection logic

### Measurable Outcomes

**Week 1 Validation:**
- 10+ early adopters testing MVP
- Zero critical bugs reported
- Detection accuracy validated on test suite

**Month 3 Validation:**
- 50+ active daily users
- 3-5 user testimonials collected
- 95%+ detection accuracy maintained
- At least 1 organic online post about the tool

**Month 6 Validation:**
- 200+ active daily users
- 500 GitHub stars
- Active community engagement (issues/PRs)
- Tool mentioned in vibe coding method communities

**Year 1 Success:**
- 1,000+ GitHub stars
- 1,000+ active daily users
- Multiple community-created tutorials/blog posts
- Tool becomes recommended in BMAD-Method/Speckit documentation
- "Jeff-style" viral testimonials appearing organically

## Product Scope

### MVP - Minimum Viable Product (4-6 Weeks)

**Core Dashboard (Essential):**
- `vibe` command displays dashboard of active projects
- Per-project display: name, stage, last modified, detection reasoning
- Interactive TUI: [a] Add [h] Hibernated [r] Refresh [d] Details [?] Help [q] Quit
- Visual indicators: ✨ recent (today), ⚡ active (this week), 🤷 uncertain

**Project Management (Essential):**
- `vibe add .` - Add project from current directory
- `vibe add <path>` - Add project from specified path
- Centralized state: `~/.vibe/` directory with per-project SQLite databases
- Auto-detect Speckit methodology (primary focus for MVP)

**Speckit Stage Detection (Essential):**
- Detect Speckit folder structure (`.specify/`, `.speckit/`, `specs/`)
- Parse `specs/NNN-feature-name/` directories
- Stage identification: spec.md → Specify, plan.md → Plan, tasks.md → Tasks, implement.md → Implement
- Transparent uncertainty handling when detection unclear

**Real-Time Refresh (Essential):**
- fsnotify file watcher monitors Speckit artifacts
- Debouncing for rapid changes (5-10 seconds)
- Manual `[r]` refresh forces immediate re-scan
- Visual "last updated Xs ago" timestamp

**Agent Waiting State Detection (Killer Feature - Essential):**
- Heuristic-based detection with 10-minute threshold
- Visual indicator: `⏸️ WAITING - Agent needs your input (Xh)`
- Manual refresh clears/updates state
- Future refinement: Sub-1-minute detection based on user feedback

**Architecture (Essential):**
- Hexagonal architecture with plugin-based method detection
- Interface-based design: MethodDetector interface for vibe coding methods
- Speckit detector implemented as plugin (not hardcoded)
- Ready for BMAD-Method expansion post-MVP

**Cross-Platform Support (Essential):**
- Single binary deployment (Go compiled)
- Works on Linux, macOS, Windows
- Centralized SQLite storage (`~/.vibe/`)

### Growth Features (Post-MVP - Month 2-3)

**Enhanced Intelligence:**
- BMAD-Method detection support (plug in via MethodDetector interface)
- Dual-method conflict resolution (both BMAD + Speckit in same project)
- Auto-hibernation system refinement (active vs dormant projects)
- `vibe hibernated` command for browsing dormant projects
- Improved agent waiting state detection (sub-1-minute threshold)

**Advanced Features:**
- Progress metrics and daily recap ("Good job! You completed 3 specs this week")
- Fuzzy search across projects
- Per-project notes (lightweight context)
- `vibe recent` command for morning startup
- Configurable hibernation thresholds

### Vision (Future - Month 6-12)

**Ecosystem & Community:**
- Plugin architecture for custom method detectors (community contributions)
- Community-contributed method detectors
- Additional vibe coding method support (community-requested)
- Integration hooks for other tools

**Optional Advanced Features:**
- Cloud sync via CRDTs for multi-device (optional, not core)
- Web dashboard companion (read-only view)
- Team features (shared project visibility - optional)
- Advanced analytics: methodology adherence tracking

**Long-Term North Star:**
- Vibe Dashboard becomes default tool in vibe coding method documentation
- Active community maintains and extends tool
- New vibe coding methodologies request official support
- "How do you manage multiple vibe coding projects?" → "vibe dashboard"

---

## User Journeys

### Journey 1: Jeff - The Super-Human Freelancer (Primary User)

**Who Jeff Is:**
Jeff is a senior developer turned freelance vibe coding specialist who manages 3-6 client projects simultaneously plus personal experiments. He uses Claude Code primarily and is eager to try new vibe coding tools. He works across multiple BMAD-Method and Speckit projects and considers himself a "super-human" because AI agents let him accomplish what seemed impossible before.

**The Problem - Before Vibe Dashboard:**
Jeff wakes up Monday morning after a weekend break. He opens his terminal and stares at his project directories. "Which client project was at PRD stage? Did I finish Epic breakdown on that other one? Was the personal project still in brainstorming?" He spends 5-10 minutes navigating through .bmad folders, reading timestamps, reconstructing context. By the time he figures out where everything stands, he's mentally exhausted before real work begins. He's become the bottleneck in his own projects - not because of coding, but because of mental context switching overhead.

**Discovery & First Use:**
Jeff hears about Vibe Dashboard from another freelancer managing multiple vibe coding projects who posts: "Never ask 'where was I?' again." Intrigued, Jeff installs the single binary and navigates to his first project directory.

```bash
$ cd ~/client-project-alpha
$ vibe add .
→ ✓ Added! Detected: Speckit
→ Tracking stage: Plan (plan.md found, no tasks yet)
→ Run `vibe` to see dashboard

$ vibe
┌─ ACTIVE PROJECTS (1) ─────────────────────────┐
│ client-project-alpha                           │
│   ✨ Plan Stage (5m ago)                       │
│   └─ Detected: plan.md exists, no tasks yet   │
└────────────────────────────────────────────────┘
```

Jeff's reaction: "Wait, it just... knew? No configuration?" He adds his other projects one by one, watching the dashboard populate.

**The "Aha!" Moment (Week 1):**
Tuesday morning. Jeff wakes up, opens terminal, types `vibe` out of habit now. The dashboard appears:

```
┌─ ACTIVE PROJECTS (5) ─────────────────────────────────┐
│ client-alpha: ✨ Plan Stage (16h ago)                 │
│ client-bravo: ⏸️ WAITING - Agent needs input (2h)     │
│ personal-experiment: ✨ Specify Stage (1d ago)        │
│ client-charlie: ⚡ Implementation (3d ago)            │
│ client-delta: ✨ Tasks Stage (today)                  │
└────────────────────────────────────────────────────────┘

💤 2 projects hibernated (inactive 14d+) - Press [h]
```

Jeff immediately knows: "Client Bravo needs my attention NOW - agent's been waiting 2 hours." He switches to that project, provides input, switches back. **Total context reconstruction time: 10 seconds.** No folder navigation, no mental reconstruction. Just instant clarity.

**Living in Vibe Dashboard (Month 1+):**
- Jeff keeps `vibe` open in a terminal split all day, every day
- When the ⏸️ indicator appears, he acts immediately - no forgotten blocked agents
- He's now managing 8 projects comfortably (up from his previous max of 5)
- His morning ritual: Coffee → `vibe` → Start working on flagged project
- Zero "what was I doing?" moments anymore

**Success Quote:**
"I can manage more than 10 vibe coding projects thanks to vibe dashboard. I don't waste time watching over coding agents doing their job or waiting for me. I can focus more on looking over the documents that agents created."

**Jeff's Transformation:**
- **Before**: 5-10 minutes context reconstruction per project switch
- **After**: <10 seconds to full context clarity
- **Before**: Managing 3-5 projects (cognitive limit)
- **After**: Managing 8-10 projects comfortably
- **Before**: Frequently forgot which agent was waiting
- **After**: Proactive agent attention through ⏸️ signals

---

### Journey 2: Sam - The Vibe Coding Learner (Secondary User)

**Who Sam Is:**
Sam is a developer who recently discovered vibe coding methodologies after realizing "just tossing prompts to coding agents goes nowhere." He's actively learning BMAD-Method and Speckit, frequently referring to official documentation to understand what comes next in the workflow. He's simultaneously learning the methodology while trying to execute projects with it.

**The Problem - Before Vibe Dashboard:**
Sam knows the stages exist (Brainstorming → Product Brief → PRD → Epics → Stories → Implementation) but keeps forgetting where he is and what comes next. He has the BMAD-Method documentation open in one tab, his project in another, constantly cross-referencing: "Okay, I finished the Product Brief... what's next? Oh right, PRD. Wait, did I already start the PRD or not?" He's learning workflow structure while trying to execute it - constant cognitive overload.

**Discovery & First Use:**
Sam finds Vibe Dashboard mentioned in the BMAD-Method documentation as a "helpful tool for tracking workflow state." He installs it and adds his learning project:

```bash
$ cd ~/learning-project
$ vibe add .
→ ✓ Added! Detected: BMAD-Method
→ Tracking stage: Product Brief (prd.md not found yet)

$ vibe
┌─ ACTIVE PROJECTS (1) ─────────────────────────┐
│ learning-project                               │
│   ✨ Product Brief Stage (today)               │
│   └─ Detected: product-brief.md exists        │
└────────────────────────────────────────────────┘
```

Sam's reaction: "Oh! So I'm at Product Brief stage. That means PRD comes next!" The dashboard makes the abstract workflow structure concrete and visible.

**The "Aha!" Moment (Week 2):**
Sam doesn't check the BMAD documentation for three days. He just works, occasionally glancing at `vibe` to see where he is. One morning he realizes: "Wait, I haven't opened the methodology docs in days. The dashboard has been teaching me the flow just by showing me where I am."

The implicit coaching worked:
- Dashboard shows "PRD Stage" → Sam knows he's past Product Brief, hasn't reached Epics
- Dashboard shows "Epic Stage" → Sam knows Stories come next
- Dashboard shows "Implementation" → Sam knows he's executing

**Living with Vibe Dashboard (Month 1+):**
- Sam runs `vibe` even on his single learning project (validates daily utility)
- The methodology workflow is now internalized - he doesn't need docs anymore
- When he starts a second project, context switching is effortless
- Dashboard confidence: He trusts where the tool says he is

**Success Quote:**
"I realized I'd internalized the vibe coding workflow without studying documentation - the dashboard trained me through daily use."

**Sam's Transformation:**
- **Before**: Constantly checking methodology documentation
- **After**: Workflow structure internalized through dashboard visibility
- **Before**: Uncertain if following methodology correctly
- **After**: Confident in current stage and next steps
- **Before**: Cognitive overload (learning + executing simultaneously)
- **After**: Dashboard handles "where am I?" so Sam focuses on execution

---

### Journey 3: Methodology Creator - Plugin Integration Path (Growth User)

**Who They Are:**
Creators of structured vibe coding methodologies (like BMAD-Method, Speckit) who want their methodology supported in Vibe Dashboard to provide better tooling for their users.

**The Problem:**
Their methodology users struggle with workflow state tracking. They recommend manual approaches or generic tools, but nothing understands their specific methodology structure. Users who adopt their method still lose context switching between projects.

**Discovery:**
They hear from their community: "I use Vibe Dashboard to track my [methodology] projects - works great!" They investigate and discover Vibe Dashboard already has plugin architecture for adding new methodology support.

**Integration Journey:**

**Phase 1: Research (Week 1)**
- Review Vibe Dashboard documentation on MethodDetector interface
- Examine existing Speckit and BMAD-Method detector implementations
- Understand artifact detection patterns and stage mapping

**Phase 2: Development (Week 2-3)**
- Implement MethodDetector interface for their methodology
- Define artifact patterns (folder structure, file signatures)
- Map methodology stages to detection heuristics
- Test against 20 real projects (golden path test suite)

**Phase 3: Contribution (Week 4)**
- Submit plugin via GitHub PR
- Community testing and feedback
- Merge and release with next Vibe Dashboard version

**Phase 4: Adoption**
- Announce in their methodology documentation: "Official Vibe Dashboard support!"
- Users install updated Vibe Dashboard, get automatic detection
- Methodology adoption increases due to better tooling

**Success Moment:**
"Our methodology is officially supported in Vibe Dashboard. Users love having automatic workflow state tracking."

**Methodology Creator's Value:**
- **Better tooling** for their methodology users
- **Increased adoption** through superior developer experience
- **Community growth** via improved workflow support
- **Ecosystem participation** in vibe coding tool landscape

**Design Implications:**
- Plugin architecture must be well-documented
- MethodDetector interface must be stable and extensible
- Golden path test suite must be easy to replicate
- Community contribution process must be frictionless

---

## Key Journey Insights

**Common Thread Across All Users:**
All users experience the transformation from "manual context reconstruction" to "instant clarity through automatic detection."

**MVP Focus:**
- Jeff and Sam journeys drive MVP feature requirements
- Methodology Creator journey informs plugin architecture design

**Success Validation:**
- Jeff's quote about managing 10+ projects = scale breakthrough metric
- Sam's internalization = implicit coaching success
- Methodology Creator adoption = ecosystem growth indicator

**Journey-to-Feature Mapping:**
- Jeff's "⏸️ WAITING" need → Agent Waiting State Detection (killer feature)
- Sam's learning curve → Transparent stage detection with reasoning
- Methodology Creator → Plugin architecture (MethodDetector interface)

---

## Design Philosophy & Differentiation

### Multi-Method Plugin Architecture

Unlike single-methodology tools (e.g., Specflow's Speckit-only dashboard), Vibe Dashboard uses a **MethodDetector interface** for pluggable methodology support. This architectural decision enables:

- **Current:** Speckit + BMAD-Method detection
- **Future:** Community-contributed method detectors
- **Flexibility:** Users working across multiple methodologies in different projects

This design choice differentiates Vibe Dashboard as a **methodology-agnostic workflow tracker** rather than a single-method tool.

### Hibernation Model Refinement

**Automatic Hibernation Logic:**
- Projects **untouched for 7-14 days** (configurable) automatically move to hibernated state
- Reduces cognitive noise by focusing attention on recent work
- Auto-promotion: Working on a hibernated project automatically brings it back to active state

**Manual Override - Favorites:**
- Users can **mark projects as favorites** to keep them visible regardless of activity
- Solves "important but inactive" problem with minimal complexity
- Clear mental model: Recent work + Manual favorites = Active view

**Visual Model:**
```
Active (Auto) = Touched recently (< 7-14 days)
Hibernated (Auto) = Untouched for extended period
Favorite (Manual) = Always visible regardless of activity
```

This three-tier system provides automatic noise reduction while preserving user control over exceptions.

### Execution Focus: Necessary Over Novel

Vibe Dashboard prioritizes **solving real developer pain** (cognitive overload from context switching) over breakthrough innovation. The value proposition is **reliable execution of workflow state detection** to reduce mental overhead and save developer time.

The product philosophy centers on being a **necessary tool** for vibe coding practitioners rather than pursuing innovation theater. Success means developers can't work without it anymore, not that it uses cutting-edge technology.

---


## CLI Tool Specific Requirements

### Command Structure & Interaction Modes

Vibe Dashboard supports both **interactive TUI** (primary use case) and **non-interactive scriptable mode** for automation.

#### Interactive Mode (Primary)
```bash
vibe                    # Launch full TUI dashboard
vibe [h]               # Show hibernated projects view
vibe [?]               # Show help/keyboard shortcuts
```

**Interactive Features:**
- Bubble Tea TUI with real-time updates
- Keyboard-driven navigation: [a] Add [h] Hibernated [r] Refresh [d] Details [?] Help [q] Quit
- Visual indicators: ✨ recent, ⚡ active, 🤷 uncertain, ⏸️ waiting
- Always-visible dashboard intended to stay open in terminal split

#### Non-Interactive Mode (Automation)
```bash
vibe list              # Plain text project list
vibe list --json       # JSON output for parsing
vibe status <project>  # Check specific project status
vibe add <path>        # Add project non-interactively
```

**Use Cases:**
- CI/CD pipeline integration
- Shell scripting and automation
- Logging and monitoring
- Remote project status checks

### Output Formats

**Human-Readable TUI:**
- Primary interactive display format
- Real-time updates via fsnotify
- Rich visual formatting with indicators

**JSON (Versioned):**
- Machine-readable structured output
- API versioning for schema stability: `vibe list --json --api-v1`
- Defaults to latest stable version when version flag omitted
- Prevents breaking changes in automation scripts
- Example: `vibe list --json` outputs project array with full state

**Plain Text:**
- Simple text output for logging
- Tab-separated or newline-separated
- Example: `vibe list` outputs project names and stages

### Configuration Schema

**Centralized Configuration Structure:**
```
~/.vibe/
  ├── config.yaml                 # Master index (single source of truth)
  ├── api-service/
  │   ├── config.yaml             # Project-specific settings only
  │   └── state.db                # SQLite state database
  └── client-b-api-service/
      ├── config.yaml
      └── state.db
```

**Master Config (`~/.vibe/config.yaml`):**
```yaml
# Single source of truth for project path mappings
projects:
  api-service:
    path: "/home/user/client-a/api-service"
    favorite: false
  client-b-api-service:
    path: "/home/user/client-b/api-service"
    favorite: true

settings:
  hibernation_days: 14
  refresh_interval_seconds: 10
  agent_waiting_threshold_minutes: 10
```

**Project Config (`~/.vibe/<project>/config.yaml`):**
```yaml
# Project-specific settings only (no path duplication)
detected_method: "speckit"
last_scanned: "2025-12-08T07:03:00Z"
custom_hibernation_days: null  # Override global setting if needed
```

**Configuration Priority:**
1. **CLI flags** (highest priority): `vibe --hibernation-days=7`
2. **Project config**: `~/.vibe/<project>/config.yaml`
3. **Master config**: `~/.vibe/config.yaml`
4. **Smart defaults** (lowest priority)

**Default Config Creation:**
- On first `vibe add .`, auto-create `~/.vibe/config.yaml` if not exists
- Create project-specific config directory with defaults
- Zero manual configuration required for basic operation

### Project Name Collision Handling

**Strategy:** Human-readable directory names with parent-directory disambiguation.

**Resolution Algorithm:**
```
1. First project: use project name
   ~/.vibe/api-service/

2. Collision: add parent directory
   ~/.vibe/client-b-api-service/

3. Still collision: add grandparent
   ~/.vibe/work-client-b-api-service/

4. Continue up directory tree until unique
```

**Performance Note:** Directory tree traversal occurs only during `vibe add` (add-time), not at runtime. Acceptable performance cost for intuitive, debuggable directory names.

**Path Resolution:**
- Use canonical path resolution via `filepath.EvalSymlinks()` to prevent symlink ambiguity
- Store and compare canonical paths in master config
- Eliminates issues with symlinks pointing to same physical location

**Path Change Detection:**

Path validation occurs:
- **At dashboard launch** (checks all tracked projects)
- **On manual refresh command** (user-initiated)
- **NEVER during runtime** file watching (prevents network mount flakiness)

When TUI detects project path no longer exists at launch:
```
⚠️ Project "api-service" path not found:
/home/user/old-path/api-service

[D] Deleted - Remove from dashboard
[M] Moved - Update path to current directory
[K] Keep - Maybe network mount, keep tracking
```

User chooses action once at launch, dashboard updates master config accordingly. Decision is remembered - no repeated prompts during runtime.

### Shell Integration & Completion

**Shell Completion Support:**
- Bash/Zsh/Fish tab-completion via Cobra library
- Command completion: `vibe <TAB>` shows available commands
- Project name completion: `vibe status <TAB>` lists tracked projects
- Flag completion: `vibe --<TAB>` shows available flags

**Installation:**
```bash
# Bash
vibe completion bash > /etc/bash_completion.d/vibe

# Zsh
vibe completion zsh > "${fpath[1]}/_vibe"

# Fish
vibe completion fish > ~/.config/fish/completions/vibe.fish
```

Built-in with Cobra library - essentially free to implement.

### Scripting Support

**Exit Codes:**
```
0 = Success
1 = General error
2 = Project not found
3 = Invalid configuration
4 = Detection failure
```

**Scriptable Commands:**
```bash
# Check if project exists
vibe exists <project-name>

# Get project stage
vibe status <project-name> --format plain

# Add project silently
vibe add /path/to/project --quiet

# List all projects
vibe list --format json --api-v1 | jq '.[] | select(.stage == "waiting")'
```

**Automation Examples:**
```bash
# Morning standup: show projects waiting for attention
vibe list --format plain | grep "⏸️" | mail -s "Projects waiting" user@example.com

# CI/CD: validate project state before deployment
if vibe status my-project --format json --api-v1 | jq -e '.stage == "implementation"'; then
  echo "Ready to deploy"
fi
```

---


## Project Scoping & Phased Development

### MVP Strategy & Philosophy

**MVP Approach:** Problem-Solving + Platform Foundation Hybrid

Vibe Dashboard follows a dual-purpose MVP strategy:

1. **Problem-Solving Core:** Eliminate the "what was I doing?" context reconstruction problem that wastes 5+ minutes per project switch. The MVP must deliver immediate value by reducing this to <10 seconds through artifact-based workflow state detection.

2. **Platform Foundation:** Build extensible plugin architecture (MethodDetector interface) from day 1 to support multiple vibe coding methodologies. This architectural investment enables community contributions and future method support without requiring core rewrites.

**Strategic Rationale:**

This hybrid approach balances immediate user value with long-term sustainability. The problem-solving focus ensures early adopters get real utility, while the platform foundation enables the ecosystem growth needed to reach 1,000+ GitHub stars within 12 months.

**Personal Validation Threshold:**

Primary creator (Jongkuk) serves as first validation: "If this tool is useful for me, it has potential." This internal usefulness test validates the core concept before broader market validation through testimonials and GitHub stars.

### Resource Requirements

**Team Size:** Solo developer (Jongkuk)

**Timeline:** 4-6 weeks to MVP launch

**Required Skills:**
- Go programming (Bubble Tea TUI framework)
- File system operations (fsnotify, path resolution)
- SQLite database operations
- Cross-platform development (Linux, macOS, Windows)
- CLI tool design patterns
- Speckit methodology knowledge

**Development Environment:**
- Go 1.21+ toolchain
- Bubble Tea library for TUI
- Cobra library for CLI framework and shell completion
- fsnotify for file watching
- SQLite for state management

### MVP Feature Set (Phase 1: Weeks 1-6)

**Core Dashboard (Essential):**
- `vibe` command displays real-time dashboard of active projects
- Per-project display: name, current stage, last modified timestamp, detection reasoning
- Interactive TUI with keyboard navigation: [a] Add [h] Hibernated [r] Refresh [d] Details [?] Help [q] Quit
- Visual indicators: ✨ recent (today), ⚡ active (this week), 🤷 uncertain, ⏸️ waiting
- Non-interactive mode: `vibe list --json --api-v1` for scripting/automation

**Project Management (Essential):**
- `vibe add .` - Add project from current directory with zero configuration
- `vibe add <path>` - Add project from specified path
- Centralized state: `~/.vibe/` directory with master config and per-project subdirectories
- Canonical path resolution via `filepath.EvalSymlinks()` to handle symlinks correctly
- Human-readable project directory names with parent-directory collision disambiguation
- Path change detection at launch with user-prompted resolution (Deleted/Moved/Keep)

**Speckit Stage Detection (Essential - 95% Accuracy Target):**
- Detect Speckit folder structure (`.specify/`, `.speckit/`, `specs/`)
- Parse `specs/NNN-feature-name/` directories for workflow artifacts
- Stage identification heuristics:
  - spec.md exists → Specify stage
  - plan.md exists → Plan stage
  - tasks.md exists → Tasks stage
  - implement.md exists → Implement stage
- Transparent uncertainty handling: Show 🤷 indicator when detection unclear
- Golden path test suite: 20 real Speckit projects with known stages for validation

**Real-Time Refresh (Essential):**
- fsnotify file watcher monitors Speckit artifact changes
- Debouncing for rapid changes (5-10 second delay)
- Launch-time path validation (not runtime)
- Manual `[r]` refresh forces immediate re-scan from artifacts
- Visual "last updated Xs ago" timestamp for status freshness

**Agent Waiting State Detection (Killer Feature - Essential):**
- Heuristic-based detection: 10 minutes of file inactivity = agent waiting
- Visual indicator: `⏸️ WAITING - Agent needs your input (Xh)` with hours elapsed
- Manual refresh clears/updates waiting state
- Prevents lost workflow momentum when agent blocked on user input
- Post-MVP refinement: Sub-1-minute detection threshold based on user feedback

**Architecture (Essential):**
- Hexagonal architecture with plugin-based method detection
- MethodDetector interface for vibe coding methodology support
- Speckit detector implemented as first plugin (demonstrates pattern)
- Ready for BMAD-Method expansion post-MVP via plugin addition

**Configuration System (Essential):**
- Master config (`~/.vibe/config.yaml`) as single source of truth for project paths
- Per-project config (`~/.vibe/<project>/config.yaml`) for project-specific settings
- Per-project SQLite database (`~/.vibe/<project>/state.db`) for detection state
- CLI flags override config values for flexibility
- Smart defaults: Auto-create config on first `vibe add`

**Shell Integration (Essential):**
- Bash/Zsh/Fish tab-completion via Cobra library (essentially free to implement)
- Command completion, project name completion, flag completion
- Scriptable exit codes (0=success, 1=error, 2=not found, 3=invalid config, 4=detection failure)

**Cross-Platform Support (Essential):**
- Single binary deployment (Go compiled, no dependencies)
- Works reliably on Linux, macOS, Windows
- Centralized SQLite storage at `~/.vibe/` (user home directory)

**Hibernation System (Essential):**
- Auto-hibernation: Projects untouched 7-14 days (configurable) move to hibernated state
- Manual override: Mark projects as favorites to keep visible regardless of inactivity
- Auto-promotion: Working on hibernated project automatically brings it back to active
- Visual model: Active (recent work) + Favorite (manual pin) + Hibernated (inactive)
- Always visible count: "X active, Y hibernated" prevents data-loss panic

### Post-MVP Features

**Phase 2: Growth Features (Month 2-3)**

**Enhanced Intelligence:**
- BMAD-Method detection support (second MethodDetector plugin)
- Dual-method conflict resolution (handle projects with both BMAD + Speckit)
- Improved agent waiting state detection (sub-1-minute threshold refinement)
- `vibe hibernated` command for browsing dormant projects
- Configurable hibernation thresholds per project

**Advanced Features:**
- Progress metrics: Daily/weekly recap ("Good job! You completed 3 specs this week")
- Fuzzy search across projects (quick navigation)
- Per-project notes field (lightweight context annotations)
- `vibe recent` command for morning startup routine
- Enhanced TUI navigation (filtering, sorting options)

**Phase 3: Expansion Features (Month 6-12)**

**Ecosystem & Community:**
- Public plugin API documentation for custom method detectors
- Community-contributed method detectors (GitHub contributions)
- Additional vibe coding method support (community-requested priorities)
- Integration hooks for external tools (CI/CD, IDEs)

**Optional Advanced Features (Not Core):**
- Cloud sync via CRDTs for multi-device synchronization (optional)
- Web dashboard companion (read-only browser view)
- Team features (shared project visibility - optional for larger teams)
- Advanced analytics: Methodology adherence tracking, velocity metrics

**Long-Term North Star (Year 2+):**
- Vibe Dashboard becomes default recommended tool in vibe coding method documentation
- Active community maintains and extends tool via plugin contributions
- New vibe coding methodologies request official Vibe Dashboard support
- Answer to "How do you manage multiple vibe coding projects?" becomes "vibe dashboard"

### Risk Mitigation Strategy

**Technical Risks:**

**Risk 1: 95% Stage Detection Accuracy Not Achieved**
- **Impact:** Users won't trust tool if detection frequently wrong (launch blocker)
- **Mitigation:** Golden path test suite with 20 real Speckit projects before launch
- **Validation:** Automated test suite must pass at 95%+ accuracy threshold
- **Contingency:** If <95%, reduce scope to single methodology initially, refine detection

**Risk 2: Cross-Platform Compatibility Issues**
- **Impact:** Tool fails on Windows or macOS reduces addressable market
- **Mitigation:** Test on all three platforms (Linux, macOS, Windows) before launch
- **Validation:** Canonical path resolution via `filepath.EvalSymlinks()` handles platform differences
- **Contingency:** Document known platform limitations, ship Linux/macOS first if needed

**Risk 3: Performance Degradation at Scale**
- **Impact:** Tool becomes slow with 20+ projects, users abandon
- **Mitigation:** Performance benchmarks before launch (<100ms render, <10s file detection)
- **Validation:** Test with 50 mock projects to validate performance at 2.5x expected scale
- **Contingency:** Lazy loading, pagination, or filtering if performance issues discovered

**Risk 4: fsnotify File Watching Reliability**
- **Impact:** Changes not detected, manual refresh required frequently
- **Mitigation:** Manual refresh always available as fallback
- **Validation:** Test file watching across platforms with various file system types
- **Contingency:** Document manual refresh requirement, consider polling fallback

**Market Risks:**

**Risk 1: Vibe Coding Adoption Too Niche**
- **Impact:** Insufficient users to reach 50+ daily active users by Month 3
- **Mitigation:** Personal validation first ("useful for me"), then early adopter outreach
- **Validation:** If useful for creator, likely useful for other vibe coding practitioners
- **Contingency:** Position as general project dashboard, not vibe-coding-specific

**Risk 2: Competing Tools Emerge**
- **Impact:** Better-funded or feature-rich competitors reduce adoption
- **Mitigation:** Plugin architecture enables fast methodology support additions
- **Validation:** ccmanager exists but has different approach (agent-launch-based vs artifact-based)
- **Contingency:** Focus on differentiation (methodology-agnostic, artifact-based truth)

**Risk 3: No Organic Discovery**
- **Impact:** GitHub stars don't grow, tool remains unknown
- **Mitigation:** Post in BMAD-Method and Speckit communities at launch
- **Validation:** 3-Month target: At least 1 organic online post from real user
- **Contingency:** Active community engagement, content creation (blog posts, demos)

**Resource Risks:**

**Risk 1: MVP Takes Longer Than 4-6 Weeks**
- **Impact:** Delays validation, creator motivation risk
- **Mitigation:** Strict MVP scope adherence, no feature creep
- **Validation:** Weekly progress check against timeline
- **Contingency:** Cut nice-to-have features (shell completion, non-interactive mode) if timeline slips

**Risk 2: Solo Developer Burnout**
- **Impact:** Project abandonment, users lose trust
- **Mitigation:** 4-6 week timeline keeps momentum, prevents extended grind
- **Validation:** Personal usefulness test ensures creator is primary beneficiary
- **Contingency:** Reduce scope to "useful for me" minimum if burnout imminent

**Risk 3: Post-Launch Maintenance Overhead**
- **Impact:** Bug reports and feature requests overwhelm solo developer
- **Mitigation:** Conservative launch (small initial user base), clear contribution guidelines
- **Validation:** Week 1 target: 10+ early adopters (manageable support load)
- **Contingency:** Pause new features, focus on stability, recruit community maintainers

### Success Validation Checkpoints

**Week 1 (MVP Launch):**
- ✅ 10+ early adopters testing MVP
- ✅ Zero critical bugs reported
- ✅ Detection accuracy validated on golden path test suite (95%+)
- ✅ Personal usefulness validated (creator uses it daily)

**Month 3 (Early Validation):**
- ✅ 50+ active daily users
- ✅ 3-5 user testimonials collected
- ✅ 95%+ detection accuracy maintained
- ✅ At least 1 organic online post about the tool
- ✅ 100 GitHub stars

**Month 6 (Growth Validation):**
- ✅ 200+ active daily users
- ✅ 500 GitHub stars
- ✅ Active community engagement (issues/PRs)
- ✅ Tool mentioned in vibe coding method communities

**Year 1 (Established Product):**
- ✅ 1,000+ GitHub stars
- ✅ 1,000+ active daily users
- ✅ Multiple community-created tutorials/blog posts
- ✅ Tool becomes recommended in BMAD-Method/Speckit documentation
- ✅ "Jeff-style" viral testimonials appearing organically

---


## Functional Requirements

### 1. Project Management

- FR1: Users can add a project from current directory using `vibe add .`
- FR2: Users can add a project from specified path using `vibe add <path>`
- FR3: Users can view list of all tracked projects
- FR4: Users can remove a project from tracking
- FR5: Users can set a custom display name (nickname) for a project
- FR6: System can detect and resolve project name collisions using parent directory names
- FR7: System can validate project paths at launch and detect missing directories
- FR8: Users can choose action when project path is missing (Delete/Move/Keep)

### 2. Workflow Detection

- FR9: System can detect Speckit methodology from project artifacts
- FR10: System can identify current Speckit stage (Specify/Plan/Tasks/Implement)
- FR11: System can show detection reasoning when stage is identified
- FR12: System can indicate uncertainty when stage detection is unclear
- FR13: System supports pluggable methodology detectors via MethodDetector interface
- FR14: System can detect multiple methodologies in same project

### 3. Dashboard Visualization

- FR15: Users can view real-time dashboard of active projects in terminal UI
- FR16: Users can see project name (or custom nickname), stage, and last modified timestamp for each project
- FR17: Users can see visual indicators for project status (✨ recent, ⚡ active, 🤷 uncertain, ⏸️ waiting)
- FR18: Users can navigate dashboard using keyboard shortcuts [a/h/r/d/?/q]
- FR19: Users can navigate dashboard using vim-style keys [j/k/h/l]
- FR20: Users can view detailed information for a selected project
- FR21: Users can add/edit notes (memo) for a project
- FR22: Users can view project notes in dashboard detail view
- FR23: Users can manually refresh dashboard to force artifact re-scan
- FR24: Users can see count of active vs hibernated projects
- FR25: Users can view hibernated projects list
- FR26: System can display detection reasoning for current stage
- FR27: System can automatically detect file system changes in tracked projects

### 4. Project State Management

- FR28: System can automatically mark projects as hibernated after configurable days of inactivity (7-14 days default)
- FR29: System can automatically mark projects as active when file changes are detected
- FR30: Users can manually mark a project as favorite to keep it always visible regardless of activity
- FR31: Users can manually remove favorite status from a project
- FR32: Users can configure hibernation threshold (days of inactivity per project or globally)
- FR33: System can distinguish between active, hibernated, and favorite project states

### 5. Agent Monitoring

- FR34: System can detect when AI coding agent is waiting for user input (configurable inactivity threshold)
- FR35: System can display visual indicator when agent is waiting (⏸️ WAITING)
- FR36: System can show elapsed time since agent started waiting
- FR37: Users can configure agent waiting threshold (minutes of inactivity)
- FR38: System can clear waiting state when activity resumes

### 6. Configuration Management

- FR39: System can store project path mappings in centralized master config (~/.vibe/config.yaml)
- FR40: System can store project-specific settings in project config files (~/.vibe/<project>/config.yaml)
- FR41: Users can override configuration values using CLI flags
- FR42: Users can modify global configuration by directly editing master config file
- FR43: Users can modify project-specific configuration by directly editing project config files
- FR44: System can auto-create default configuration on first project add
- FR45: System can use canonical paths to handle symlinks correctly
- FR46: Users can configure global settings (hibernation days, refresh interval, agent waiting threshold)
- FR47: Users can configure per-project settings that override global defaults

### 7. Scripting & Automation

- FR48: Users can list projects in plain text format for scripting
- FR49: Users can request JSON output format (with API versioning) for any query command
- FR50: Users can get specific project status non-interactively
- FR51: Users can add projects with automatic conflict resolution (interactive prompts only on errors)
- FR52: Users can force automatic conflict resolution with --force flag (bypasses interactive prompts)
- FR53: Users can remove projects via CLI
- FR54: Users can mark/unmark projects as favorites via CLI
- FR55: Users can set/edit project notes via CLI
- FR56: Users can rename projects (set custom display name) via CLI
- FR57: Users can manually hibernate or activate projects via CLI
- FR58: Users can trigger manual refresh via CLI
- FR59: Users can check if a project exists via CLI
- FR60: System can return standard exit codes (0=success, 1=error, 2=not found, 3=invalid config, 4=detection failure)
- FR61: System can support shell completion for commands, project names, and flags (Bash/Zsh/Fish)

### 8. Error Handling & User Feedback

- FR62: System can gracefully handle file watching failures with fallback to manual refresh
- FR63: System can detect and report configuration file syntax errors with helpful messages
- FR64: System can recover from corrupted state databases by reinitializing
- FR65: Users can view keyboard shortcut help
- FR66: System can display progress indicators during long-running operations (refresh, detection)

---


## Non-Functional Requirements

### Performance

**Response Time Requirements:**
- NFR-P1: Dashboard renders in <100ms for up to 20 tracked projects
- NFR-P2: Dashboard startup completes in <1 second
- NFR-P3: Non-interactive CLI commands (`vibe list`, `vibe status`) respond in <500ms
- NFR-P4: Project initialization (`vibe add`) completes directory scanning in <2 seconds
- NFR-P5: TUI auto-refreshes every 5-10 seconds via file system monitoring

**File Detection:**
- NFR-P6: File system changes detected and reflected in dashboard within 5-10 seconds

### Reliability

**Stability & Recovery:**
- NFR-R1: Dashboard exits gracefully with appropriate error codes on crashes (no auto-restart)
- NFR-R2: Dashboard recovers project state by rescanning artifacts on next launch
- NFR-R3: Stage detection achieves 95%+ accuracy on golden path test suite (20 real projects)
- NFR-R4: File watching failures fallback to manual refresh without data corruption

**Data Integrity:**
- NFR-R5: Project state corruption is recoverable through artifact re-scanning
- NFR-R6: Configuration file syntax errors are detected and reported without corrupting state

### Usability

**Learning Curve:**
- NFR-U1: New users can become productive within 1 minute (`vibe add .` → `vibe` → see dashboard)
- NFR-U2: Keyboard shortcuts are visible in TUI where shortcuts exist (contextual learning)
- NFR-U3: Help is accessible via [?] key without leaving TUI

**Error Handling:**
- NFR-U4: Error messages clearly state what failed without suggesting fixes (avoid dangerous suggestions)

**Documentation:**
- NFR-U5: README + help text (`vibe --help`) provides sufficient documentation for basic usage
- NFR-U6: Keyboard shortcuts are self-documenting through TUI display

### Extensibility

**Plugin Architecture:**
- NFR-E1: MethodDetector interface is clearly defined and documented with examples
- NFR-E2: MethodDetector interface is marked as beta until stabilized with community usage
- NFR-E3: Plugin documentation enables developer understanding of implementation patterns
- NFR-E4: New methodology detectors can be added without modifying core codebase
- NFR-E5: Interface breaking changes allowed during beta phase, frozen post-beta

**Community Contribution:**
- NFR-E6: MethodDetector implementation complexity depends on methodology structure (flexible abstraction)

---

