---
stepsCompleted: [1, 2, 3, 4, 5]
inputDocuments:
  - 'docs/analysis/research/shards/09-executive-summary.md'
  - 'docs/analysis/research/shards/00-index.md'
  - 'docs/analysis/research/shards/04-vibe-coding-methods-CORRECTED.md'
  - 'docs/analysis/research/shards/06-architecture-decisions.md'
  - 'docs/analysis/research/shards/01-technology-stack.md'
  - 'docs/analysis/research/shards/07-implementation-roadmap.md'
  - 'docs/analysis/research/shards/05-technical-recommendations.md'
  - 'docs/analysis/research/shards/03-implementation-techniques.md'
  - 'docs/analysis/research/shards/02-architectural-patterns.md'
  - 'docs/analysis/research/shards/08-risk-assessment.md'
  - 'docs/analysis/brainstorming-session-2025-12-04T00:25:43.430Z.md'
workflowType: 'product-brief'
lastStep: 0
project_name: 'bmad-test'
user_name: 'Jongkuk Lim'
date: '2025-12-05'
---

# Product Brief: bmad-test

**Date:** 2025-12-05
**Author:** Jongkuk Lim

---

<!-- Content will be appended sequentially through collaborative workflow steps -->

## Executive Summary

**Vibe Dashboard** is a CLI-first dashboard that serves as the central nervous system for developers managing multiple AI-assisted vibe coding projects. It eliminates cognitive load by automatically tracking workflow state across projects using structured methodologies (BMAD-Method, Speckit), intelligently surfacing only active work while hibernating dormant projects, and proactively signaling when human attention is needed.

**Target Users:** Solo developers using structured vibe coding methodologies with AI coding agents across 3-10 simultaneous projects.

**Core Problem Solved:** Reduces context reconstruction time from minutes to seconds when switching between AI-assisted projects, eliminating "what was I doing?" moments through automatic workflow state tracking from project artifacts.

**Business Model:** Open source (MIT license) with community-driven growth via GitHub stars and optional buy-me-a-coffee support. Success measured by adoption, not revenue.

**Technology Stack:** Go + Cobra + Bubble Tea for production-grade CLI dashboard with real-time performance via fsnotify file watching and cross-platform single binary deployment.

**Killer Feature:** Agent waiting state detection - dashboard shows when coding agent is blocked waiting for human command/decision, ensuring no lost workflow momentum.

**Core Design Principle:** Artifacts are ALWAYS the source of truth. The dashboard reads reality from project files, never asks users to manually define state. Transparency over guessing when uncertain.

---

## Core Vision

### Problem Statement

Developers using AI coding agents across multiple projects lose mental context when switching between them. The problem isn't about code context (AI agents handle that) - it's about **workflow methodology context**: WHERE am I in the BMAD-Method or Speckit workflow? What stage was I at? What action comes next?

Current reality: Developers manage multiple Claude sessions through dashboards, manually track which project is at "PRD stage" vs "Implementation," and waste minutes reconstructing "what was I doing here?" every time they switch projects.

The cognitive reload time adds up: switching between 5-10 projects multiple times daily means **hours lost to context reconstruction** instead of actual development work.

**Key Insight:** Users need this tool BECAUSE they don't know where they are. Any solution requiring manual state maintenance defeats the purpose.

### Problem Impact

**Time Cost:**
- Minutes lost per context switch × multiple switches daily = hours of wasted productivity
- Mental energy spent remembering "where was I?" instead of "what's next?"
- Context reconstruction compounds across growing number of simultaneous projects

**Cognitive Load:**
- Human working memory limited to 5-7 items, but developers manage 10+ projects
- Stress from "did I forget something?" across multiple active projects
- Fear of lost progress or abandoned work when projects accumulate

**Workflow Disruption:**
- AI agents wait idle when developers forget they were blocked needing input
- Projects stall at critical stages because developer doesn't realize action needed
- Lost momentum when "I'll come back to this" becomes "what was I doing here?"

**Trust Requirement:**
- Dashboard must be 95%+ accurate or users won't trust it
- Wrong stage detection is worse than no detection - creates doubt
- Refresh reliability is critical - stale data destroys utility

### Why Existing Solutions Fall Short

**Current Workarounds:**
- **Manual tracking** (notes, spreadsheets) - requires discipline, becomes stale immediately
- **Claude dashboard** - shows sessions but not workflow methodology state
- **Git status** - shows code changes, not "am I at PRD stage or Implementation?"
- **File system navigation** - buried .bmad folders don't surface human-readable insights
- **Memory alone** - fails as soon as project count exceeds working memory limits

**Gap Analysis:**
- No tool understands vibe coding methodology workflows (BMAD-Method, Speckit)
- No automatic state detection from project artifacts (.bmad/, specs/)
- No intelligent filtering (everything shown = overwhelming noise)
- No proactive signaling ("this project needs your attention NOW")
- No hibernation model (stale projects clutter active workspace)
- No solution maintains accuracy without manual user corrections

### Proposed Solution

**Vibe Dashboard** reads workflow state directly from vibe coding methodology artifacts (.bmad folders, Speckit specs/) and presents a real-time CLI dashboard showing:

**For Each Active Project:**
- Current workflow stage detected from artifacts (Brainstorming → PRD → Epics → Stories → Implementation)
- Detection reasoning shown transparently ("Found prd.md, no epics/ → Stage: Planning")
- Last activity timestamp and file location
- Project description/name (human-readable, not just paths)
- Status indicators: ✨ recent progress, ⚡ active, ⏸️ agent waiting, 🤷 uncertain, 💤 hibernating

**Smart Two-State System:**
- **Active projects** (dashboard default view) - touched within threshold (7-14 days configurable)
- **Hibernated projects** (separate view via `[h]` key) - automatically hidden to reduce cognitive load
- Seamless auto-promotion: starting work on hibernated project brings it back to dashboard
- Always visible count: "3 active, 2 hibernated" prevents data-loss panic

**Developer-Friendly Add Flow:**
```bash
# Primary flow: already in project directory
$ cd ~/my-awesome-project
$ vibe add .
→ ✓ Added! Detected: BMAD-Method
→ Tracking stage: Planning (prd.md found)

# Dashboard shows it immediately
$ vibe
PROJECT: my-awesome-project
  ✨ Planning Stage (2m ago)
  └─ Detected: prd.md exists, no epics yet
```

**Handling Detection Uncertainty:**
When heuristic is uncertain, show uncertainty transparently rather than guess wrong:
```
PROJECT B: 🤷 Recent activity in stories/ (today)
  └─ Unable to determine exact stage
  └─ Press [d] for details | [r] to refresh
```

**Dual-Method Conflict Resolution:**
When both BMAD and Speckit detected:
```
⚠️ Found both BMAD (.bmad/) and Speckit (specs/)
Which methodology are you actively using?
[1] BMAD-Method
[2] Speckit  
[3] Track both separately
```

**Dashboard as Environment:**
Users don't "check" the dashboard occasionally - they **live in it**. It's always visible in terminal, serving as the command center. When it signals action needed (agent waiting, project stuck), user acts. Otherwise, they trust the system and focus on current work.

**Interactive Controls:**
```
[a] Add project  [h] Show hibernated  [r] Refresh  [d] Details  [?] Help  [q] Quit
```

**North Star Vision:**
> "I will be simply watching `vibe` dashboard all the time and when this dashboard shows that it requires my action, then I go there and make some action for the project. And then come back to the dashboard. This is my life now."

### Key Differentiators

**1. Artifacts Are Always Truth**
Unlike manual tracking tools, Vibe Dashboard NEVER asks users to maintain state. It reads directly from .bmad/ and specs/ artifacts. When uncertain, shows uncertainty transparently. No `set-stage` commands - refresh forces re-scan from artifacts.

**2. 95% Detection Accuracy Target**
Rigorous heuristic development with golden path test suite (20 real BMAD/Speckit projects with known stages). Detection must be trusted or tool becomes useless. Manual overrides defeat the purpose.

**3. Real-Time Refresh Reliability**
- fsnotify file watcher for instant updates when artifacts change
- Manual `[r]` refresh keyboard shortcut forces re-scan
- Visual "last updated X seconds ago" timestamp shows data freshness
- Works on local filesystems with debouncing for rapid changes

**4. Methodology-Aware Intelligence**
Unlike generic file watchers or git status, Vibe Dashboard understands structured vibe coding workflows. It knows the difference between "PRD stage" and "Implementation" by parsing BMAD artifacts and Speckit specs.

**5. Automatic Hibernation Model**
Borrowed from nature: two-state system (active vs hibernated) mirrors natural human mental model. No manual organization, tagging, or project management overhead. System learns from activity patterns.

**6. Agent Waiting State Detection**
Killer feature: track when AI coding agent is blocked waiting for human input. Dashboard shows distinct indicator when agent idle, preventing lost workflow momentum and forgotten blocked tasks.

**7. Signal Over Noise Philosophy**
Reverse thinking breakthrough: don't show everything (overwhelming), show only what needs attention (actionable). Smart filtering keeps active projects limited to working memory capacity (5-7 items).

**8. Production-Grade Performance**
Go + Bubble Tea stack proven by k9s, kubectl, docker CLI. Single binary deployment, cross-platform, near-instant startup, real-time updates via goroutine-based file monitoring (fsnotify).

**9. Extensible Plugin Architecture**
Hexagonal architecture with interface-based method detection. Built-in support for BMAD-Method and Speckit, designed for future vibe coding methodologies without core changes.

**Why Now:**
- AI coding agents now mainstream (Cursor, Copilot, Claude, Windsurf)
- Vibe coding methodologies emerging (BMAD-Method, Speckit)
- Developers managing more simultaneous projects than ever
- CLI renaissance: k9s proves developers love terminal dashboards
- Go/Bubble Tea ecosystem mature and production-proven
- Open source community values tools that respect developer workflow

**Hard to Copy:**
- Deep understanding of vibe coding methodology workflows
- 95% accurate heuristic-based detection without agent API cooperation
- Natural two-state hibernation UX (non-obvious design)
- Artifacts-as-truth principle (no manual state maintenance)
- Method-agnostic plugin architecture for extensibility
- Transparent uncertainty handling (show "don't know" vs guess wrong)

**Open Source Strategy:**
- MIT license for maximum adoption
- GitHub stars as primary success metric
- Community-driven feature requests and contributions
- Buy-me-a-coffee for optional support
- Focus on developer love, not monetization

---

## Target Users

### Primary User: The "Super-Human" Freelancer

**Meet Jeff - Senior Developer turned Vibe Coding Freelancer**

Jeff is a skilled senior developer who spent years at a big corporation before watching LLM agents and vibe coding methods blossom into something transformative. He made the leap to freelancing and now occasionally teaches vibe coding methodologies, calling himself a "super-human" because he can accomplish what seemed impossible before AI coding agents.

**Current Reality:**
- Juggles 3-6 client projects simultaneously
- Experiments with personal idea projects in between
- Uses Claude Code primarily but eager to try new vibe coding tools (Cursor, Windsurf, Copilot CLI)
- Works across multiple BMAD-Method and Speckit projects

**The Pain Point:**
Jeff realized he's become the bottleneck in his own projects. Despite AI agents being ready to work, **he struggles to remember which vibe coding stage each project is at**. The problem hits hardest after sleep - waking up, he needs to investigate where each project left off. Was that client project at PRD stage or Implementation? Did he finish Epic breakdown on that other one?

The mental overhead of tracking 3-6 projects across different vibe coding stages drains energy before any actual work begins.

**Success Vision:**
Jeff's ideal morning: **The first thing he launches is `vibe`.** When the dashboard appears, he instantly knows what to do next. No investigation, no mental reconstruction - just clarity. He sees which projects need his attention, which agents are waiting, and exactly where each project sits in its methodology workflow.

**"Aha!" Moment:**
The moment Jeff opens `vibe` after a good night's sleep and sees:
```
PROJECT-CLIENT-A: ⏸️ WAITING - Agent needs PRD approval (16h)
PROJECT-CLIENT-B: ✨ Implementation (2d ago)
PROJECT-PERSONAL: 🤷 Activity in epics/ - Stage unclear
```

He doesn't have to think. He knows exactly where to dive in first.

---

### Secondary User: The Vibe Coding Learner

**Meet Sam - Developer Learning Vibe Coding Methods**

Sam recently discovered that vibe coding is incredibly useful after realizing that "just tossing prompts to the coding agent goes nowhere." He's actively learning BMAD-Method and Speckit, frequently referring to official documentation to understand what comes next in the workflow.

**Current Reality:**
- Learning vibe coding methodologies while working on projects
- Constantly checking documentation: "What stage comes after Product Brief?"
- Uncertain about whether he's following the methodology correctly
- Needs guidance on proper vibe coding workflow structure

**The Pain Point:**
Sam knows the stages exist (Brainstorming → Product Brief → PRD → Epics → Stories → Implementation) but forgets where he is and what comes next. He's learning the methodology while trying to execute it, which means constant context switching between docs and actual work.

**How Vibe Dashboard Helps:**
When Sam runs `vibe`, the dashboard shows him:
- Current stage detected automatically (no need to remember)
- What the methodology expects at this stage
- Clear signal when he's ready to move forward

Sam doesn't just get project state - he gets **implicit methodology coaching**. The dashboard reinforces the vibe coding workflow structure by making it visible and concrete.

**Success Moment:**
Sam realizes he's internalized the vibe coding workflow without studying documentation - the dashboard trained him through daily use.

---

### User Journey

**Discovery:**
- Jeff hears about Vibe Dashboard from another freelancer managing multiple vibe coding projects
- Sam discovers it mentioned in BMAD-Method or Speckit documentation as a helpful tool
- Both are attracted by the promise: "Never ask 'where was I?' again"

**First Experience:**
```bash
# Jeff is already in a project directory
$ cd ~/client-project-alpha
$ vibe add .
→ ✓ Added! Detected: BMAD-Method
→ Tracking stage: PRD (prd.md found, no epics yet)
→ Run `vibe` to see dashboard

$ vibe
┌─ ACTIVE PROJECTS (1) ─────────────────────────┐
│ client-project-alpha                           │
│   ✨ PRD Stage (5m ago)                        │
│   └─ Detected: prd.md exists, no epics yet    │
└────────────────────────────────────────────────┘

[a] Add project  [r] Refresh  [?] Help  [q] Quit
```

**"Aha!" Moment:**
- **Jeff:** Wakes up next morning, types `vibe`, and **instantly knows** which project needs attention without mental reconstruction
- **Sam:** Dashboard shows "PRD Stage" and he realizes "Oh! I'm past Product Brief but haven't broken into Epics yet" - methodology becomes concrete

**Daily Routine:**
- Jeff keeps `vibe` dashboard visible in a terminal split while working
- When dashboard signals ⏸️ agent waiting, he switches to that project
- After sleep/breaks, first command is `vibe` - instant context restoration
- Dashboard becomes his external working memory

**Long-term Integration:**
- Jeff's morning ritual: Coffee → `vibe` → Start working on flagged project
- Sam stops checking methodology docs - dashboard shows where he is
- Both trust the 95% detection accuracy - refresh `[r]` when needed
- Hibernation feature prevents overwhelming dashboard as project count grows

**Viral Moment:**
Jeff screenshots his `vibe` dashboard showing 5 active projects with perfect stage detection. Posts on Twitter: "This is how I manage 5 vibe coding projects without losing my mind." GitHub stars roll in from other freelancers facing the same pain.

---


## Success Metrics

### User Success Metrics

**Jeff's Morning Test (Primary Success Indicator):**
- **Target:** `vibe` is the first command Jeff types after opening terminal in the morning
- **Measurement:** User behavior tracking - does dashboard become default startup command?
- **Success Threshold:** Jeff keeps dashboard open in terminal split throughout workday

**Context Reconstruction Time:**
- **Current State:** 5+ minutes investigating .bmad folders and remembering project state
- **Target State:** Under 10 seconds from `vibe` command to "I know what to do next"
- **Measurement:** Time from dashboard display to user taking action on a project

**Zero "What Was I Doing?" Moments:**
- **Behavioral Indicator:** User never manually navigates to .bmad/ or specs/ folders to check state
- **Success Threshold:** Dashboard transparency eliminates need for artifact investigation
- **Trust Signal:** Users rely on dashboard detection, only use `[r]` refresh when truly needed

**Stage Detection Accuracy:**
- **Target:** 95%+ accuracy on real BMAD-Method and Speckit projects
- **Measurement:** Golden path test suite (20 real projects with known stages) passes consistently
- **User Trust:** Users don't report incorrect stage detection as blocking issue

**Dashboard Performance:**
- **Render Speed:** Dashboard displays in <100ms for up to 20 projects
- **Refresh Reliability:** File changes detected within 5-10 seconds (configurable via fsnotify)
- **Cross-platform:** Works reliably on Linux, macOS, Windows

**Hibernation Adoption:**
- **Natural Behavior:** Users don't manually manage project visibility
- **Success Threshold:** Active projects stay limited to 5-7 (working memory capacity)
- **Panic Prevention:** Zero reports of users feeling "projects disappeared"

**Sam's Learning Success:**
- **Implicit Coaching:** Sam stops checking BMAD/Speckit documentation for "what stage comes next"
- **Methodology Internalization:** Dashboard visibility teaches workflow structure through daily use
- **Confidence Growth:** Sam knows where he is in methodology without external reference

---

### Business Objectives

**Open Source Community Growth:**
- **GitHub Stars Target:**
  - Month 1: 100 stars (early adopters)
  - Month 6: 500 stars (growing community)
  - Year 1: 1,000+ stars (established tool)
- **Success Indicator:** Organic growth through user recommendations and social proof

**Active User Adoption:**
- **Month 3:** 50 active users (validation phase)
- **Month 6:** 200 active users (product-market fit)
- **Year 1:** 1,000+ active users (community established)
- **Definition:** Active = running `vibe` daily for 1+ week

**Community Engagement:**
- **Issue/PR Activity:** Healthy stream of bug reports, feature requests, contributions
- **Documentation:** Community creates tutorials, blog posts, integration guides
- **Social Proof:** Users share screenshots/testimonials showing multi-project management

**Viral Moment Success:**
- **Target:** Jeff-style testimonials - "Managing X projects with vibe dashboard"
- **Platform:** Twitter, Reddit (r/programming), HackerNews upvotes
- **Success Threshold:** At least one viral post driving significant GitHub traffic

---

### Key Performance Indicators

**Technical KPIs:**
- **Detection Accuracy:** 95%+ on golden path test suite (critical trust metric)
- **Performance:** <100ms dashboard render for 20 projects
- **Refresh Reliability:** File changes detected within 5-10 seconds
- **Cross-platform Success:** Works on Linux, macOS, Windows without platform-specific bugs
- **Zero Config:** No setup required beyond `vibe add .` command

**Adoption KPIs:**
- **Daily Active Users (DAU):** Users running `vibe` command daily
- **Retention Rate:** % of users still active 30 days after first use
- **Project Tracking:** Average number of projects tracked per user (target: 3-6)
- **Session Duration:** Dashboard kept visible throughout work sessions

**Quality KPIs:**
- **Issue Resolution Time:** Critical bugs fixed within 48 hours
- **False Positive Rate:** <5% incorrect stage detection reports
- **User Satisfaction:** Based on GitHub stars, testimonials, retention

**Community KPIs:**
- **GitHub Activity:**
  - Stars growth rate (month-over-month)
  - Fork count (community interest in customization)
  - Contributor count (community participation)
- **Social Mentions:** Twitter, Reddit, blog posts mentioning vibe dashboard
- **Integration Requests:** Users asking for additional vibe coding method support

**Viral Success Indicators:**
- **Social Media Reach:** Posts showing multi-project management get 100+ likes/retweets
- **Organic Traffic:** GitHub stars spike after social media mentions
- **Word of Mouth:** New users report "found via friend recommendation"
- **Testimonials:** Users publicly share: "This changed how I work with AI agents"

---

### Success Validation

**3-Month Validation:**
- 50+ active users managing 3+ projects each
- 95%+ detection accuracy on test suite
- Zero critical bugs reported
- Community engagement: 5+ GitHub issues/PRs
- At least one positive social media mention

**12-Month Success:**
- 1,000+ GitHub stars
- Active community contributing features
- Multiple blog posts/tutorials created by users
- Jeff-style viral testimonials appearing organically
- Tool mentioned in BMAD-Method/Speckit documentation
- Other vibe coding tools requesting integration

**Long-term North Star:**
- Vibe Dashboard becomes default tool recommended in vibe coding method documentation
- "How do you manage multiple vibe coding projects?" → "Use vibe dashboard"
- Community maintains and extends tool beyond original creator
- New vibe coding methodologies ask for vibe dashboard support

---


## MVP Scope

### Core Features (4-6 Weeks)

**Essential Dashboard (Week 1-2):**
- `vibe` command displays dashboard of all active projects
- **Per-Project Display:**
  - Project name (human-readable from directory name)
  - Current workflow stage (detected from Speckit artifacts)
  - Last modified timestamp
  - Detection reasoning (transparent: "Found spec.md, no plan.md → Stage: Specify")
- **Interactive Controls:**
  - `[a]` Add project
  - `[h]` Show hibernated (future)
  - `[r]` Force refresh/re-scan
  - `[d]` Project details
  - `[?]` Help
  - `[q]` Quit
- **Visual Indicators:**
  - ✨ Recent activity (today)
  - ⚡ Active work (this week)
  - 🤷 Stage uncertain (show activity location)

**Project Management:**
- `vibe add .` - Add project from current directory
- `vibe add <path>` - Add project from specified path
- Project state stored in SQLite (~/.vibe/global.db)
- Auto-detect Speckit methodology (check for .specify/, .speckit/, specs/)

**Speckit Stage Detection (Primary Focus):**
- Detect Speckit folder structure correctly
- Parse `specs/NNN-feature-name/` directories
- Identify stage based on artifact existence:
  - spec.md exists → "Specify" stage
  - plan.md exists → "Plan" stage  
  - tasks.md exists → "Tasks" stage
  - implement.md exists → "Implement" stage
- Show most recent spec when multiple exist
- Handle uncertainty transparently (show "Activity in specs/")

**Real-Time Refresh:**
- fsnotify file watcher monitors Speckit artifacts
- Debouncing for rapid changes (group updates within 5-10 seconds)
- Manual `[r]` refresh forces immediate re-scan
- Visual "last updated Xs ago" timestamp

**Agent Waiting State Detection (Killer Feature):**
- **Heuristic-based detection** for Speckit projects:
  - Spec file modified >1 hour ago + no recent plan.md changes → ⏸️ "Waiting for planning"
  - Tasks defined but no implementation activity >1 hour → ⏸️ "Waiting to implement"
  - Show clear indicator: `⏸️ WAITING - Agent needs your input (2h)`
- **Manual refresh** allows user to clear waiting state if inaccurate
- **Transparent messaging** when heuristic uncertain

**Small Manual Changes Handling:**
- File watcher detects ANY changes to tracked artifacts
- Timestamp updates to "last modified Xs ago"
- Stage re-detection runs automatically
- If stage remains same but timestamp updated, user knows "I touched it recently"
- Dashboard shows freshness, not just absolute stage

**Architecture (Extensibility-First):**
- **Hexagonal architecture** with plugin-based method detection
- **Interface-based design:** MethodDetector interface for all vibe coding methods
- **Speckit detector:** Implemented as plugin (not hardcoded into core)
- **Ready for expansion:** BMAD, custom methods plug in without core changes
- **SQLite storage:** Method-agnostic project state management

**Cross-Platform Support:**
- Single binary deployment (Go compiled)
- Works on Linux, macOS, Windows
- Local SQLite storage (~/.vibe/)

---

### Out of Scope for MVP

**Deferred to Phase 2 (Post-MVP):**
- ❌ Auto-hibernation system (not essential for core problem)
- ❌ `vibe hibernated` command (no hibernation in MVP)
- ❌ `vibe scan` auto-discovery (too complex, time-consuming, deferred to far future)
- ❌ BMAD-Method detection (focus Speckit first, add BMAD later via plugin)
- ❌ Dual-method conflict resolution (only Speckit in MVP)
- ❌ Progress metrics and daily recap
- ❌ Fuzzy search across projects
- ❌ Per-project notes
- ❌ Configurable hibernation thresholds

**Explicitly NOT Building:**
- ❌ Web interface
- ❌ Cloud sync
- ❌ Team/collaboration features
- ❌ Manual stage override commands (artifacts are truth)
- ❌ Complex analytics or insights
- ❌ Integration with specific coding agents (Cursor, Claude, etc.)

**Rationale:**
- **Focus:** Solve Jeff's morning problem - "where am I?" across 3-6 Speckit projects
- **Speed:** Ship working MVP in 4-6 weeks, validate with real users
- **Learn:** Test 95% detection accuracy goal with Speckit before adding BMAD complexity
- **Iterate:** Build hibernation and other features based on actual user feedback

---

### MVP Success Criteria

**Technical Validation:**
- ✅ Dashboard renders <100ms for 20 projects
- ✅ 95%+ stage detection accuracy on 20 real Speckit projects (golden path test suite)
- ✅ File changes detected within 5-10 seconds
- ✅ Works on Linux, macOS, Windows without critical bugs
- ✅ Zero config - `vibe add .` just works

**User Validation (Jeff Test):**
- ✅ Jeff adds 3-6 Speckit projects
- ✅ Jeff's morning test: types `vibe` and knows what to do next in <10 seconds
- ✅ Agent waiting state correctly identifies at least one blocked project
- ✅ Jeff keeps dashboard open in terminal split for full workday
- ✅ Jeff doesn't manually check specs/ folders anymore

**Adoption Validation:**
- ✅ 10+ early adopters using MVP daily for 1 week
- ✅ Detection accuracy <5% false positive reports
- ✅ At least one user testimonial: "This changed how I work"
- ✅ GitHub repo gets 50+ stars from organic sharing

**Go/No-Go Decision:**
- **Proceed to Phase 2 if:** Users validate problem solved, detection accuracy ≥95%, retention >70% at 2 weeks
- **Pivot if:** Detection accuracy <90%, users report trust issues, retention <50%
- **Kill if:** No adoption after 3 months, fundamental architecture problems

---

### Future Vision (Post-MVP)

**Phase 2: Enhanced Intelligence (Month 2-3)**
- BMAD-Method detection support (plug in via MethodDetector interface)
- Dual-method conflict resolution (both BMAD + Speckit in same project)
- Auto-hibernation system (active vs dormant projects)
- `vibe hibernated` command
- Improved agent waiting state detection (more sophisticated heuristics)
- Configurable refresh intervals

**Phase 3: Advanced Features (Month 4-6)**
- Progress metrics and daily recap ("Good job! You completed 3 specs this week")
- Fuzzy search across projects
- Per-project notes (lightweight context)
- `vibe recent` command for morning startup
- Additional vibe coding method support (community-requested)

**Phase 4: Ecosystem (Month 6-12)**
- Plugin architecture for custom method detectors (community contributions)
- Community-contributed method detectors
- Optional cloud sync (CRDTs for multi-device)
- Web dashboard companion (read-only view)
- Integration hooks for other tools

**Long-Term North Star (Year 2+):**
- Vibe Dashboard becomes default tool in vibe coding method documentation
- Active community maintains and extends tool
- New vibe coding methodologies request official support
- "How do you manage multiple vibe coding projects?" → "vibe dashboard"
- Tool mentioned alongside BMAD-Method, Speckit as essential workflow tool

**Expansion Opportunities:**
- Team features (optional): shared project visibility
- Advanced analytics: methodology adherence tracking
- AI coaching: "Your specs typically take 2 days, this one is at 5 days"
- Cross-project insights: "Pattern detected: you always stall at tasks stage"

---

### MVP Timeline

**Week 1-2: Core Foundation**
- Go + Cobra + Bubble Tea setup
- Basic dashboard UI with interactive controls
- `vibe add` command + SQLite storage
- Speckit folder detection
- Simple stage detection (artifact existence)
- Plugin interface design (MethodDetector)

**Week 3-4: Intelligence Layer**
- fsnotify real-time file watching
- Stage detection heuristics (95% accuracy target)
- Agent waiting state detection
- Transparent uncertainty handling
- Golden path test suite (20 Speckit projects)

**Week 5-6: Polish & Validation**
- Cross-platform testing (Linux, macOS, Windows)
- Performance optimization (<100ms render)
- Documentation and README
- Early adopter testing with real users
- Bug fixes based on feedback

**Week 6: MVP Launch**
- GitHub release v1.0.0
- Announce on Twitter, Reddit (r/programming), HackerNews
- Share in BMAD/Speckit communities
- Collect user feedback for Phase 2 prioritization

---

