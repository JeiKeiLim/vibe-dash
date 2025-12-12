# Architectural Patterns

**Shard 2 of 9 - Technical Research**

---

## Overview

This shard explores architectural patterns for building CLI applications with web expansion capabilities, including plugin systems, state management, and local-first design with optional cloud sync.

---

## CLI-to-Web Application Architecture

Modern applications increasingly require both CLI and web interfaces sharing the same business logic. Several proven patterns emerged in 2025 for this dual-interface architecture.

### 1. Layered (N-Tier) Architecture

**Pattern:**
- Presentation Layer: Separate CLI and Web UI
- Business Logic Layer: Shared service classes/libraries
- Data Access Layer: Common data persistence

**Implementation:**
- Business logic resides in separate libraries
- Both CLI and web consume the same service layer
- Common in ASP.NET, Spring frameworks

**Sources:**
- https://learn.microsoft.com/en-us/dotnet/architecture/modern-web-apps-azure/common-web-application-architectures
- https://learn.microsoft.com/en-us/dotnet/standard/commandline/design-guidance

**Pros:**
- ✅ Clear separation of concerns
- ✅ Easy to test business logic independently
- ✅ Straightforward to understand

**Cons:**
- ❌ Can become monolithic
- ❌ Tight coupling between layers

---

### 2. Hexagonal (Ports and Adapters) Architecture ⭐ RECOMMENDED

**Pattern:**
- Core business logic in center (domain layer)
- Adapters implement interfaces for CLI, Web, API
- "Ports" define boundaries between core and adapters

**Implementation:**
- CLI and web are different "adapters"
- Both call the same central core logic
- Core is agnostic to how it's invoked

**Sources:**
- https://www.digitalplatformarchitect.com/patterns
- https://askai.glarity.app/search/How-can-I-build-a-command-line-interface--CLI--and-a-web-user-interface--UI--with-a-single-library

**Pros:**
- ✅ Maximum flexibility and testability
- ✅ Clean dependency management (domain has no external deps)
- ✅ Ideal for automation (CLI) + user workflows (web)
- ✅ Easy to add new interfaces without core changes

**Cons:**
- ❌ More complex initial setup
- ❌ Requires careful interface design

---

### 3. Microservices/Service-Oriented Architecture

**Pattern:**
- Core business logic exposed as internal APIs
- CLI and web both call same APIs (HTTP/gRPC)
- Services can scale independently

**Implementation:**
- Business logic runs as services
- CLI uses API client libraries
- Web frontend uses same APIs
- Local or cloud deployment

**Sources:**
- https://www.storifyagency.com/web-application-architecture-your-2025/
- https://acropolium.com/blog/modern-web-app-architecture/

**Pros:**
- ✅ Highly scalable
- ✅ Language-agnostic (CLI and web can use different stacks)
- ✅ Cloud-native deployment ready

**Cons:**
- ❌ Network overhead for local CLI use
- ❌ More complex operational requirements
- ❌ Overkill for smaller applications

---

### 4. Shared Library Pattern

**Pattern:**
- Business logic extracted into shared library/package
- CLI app and web app both reference same library
- Simplest approach for tightly integrated codebases

**Implementation:**
- DLL/.so/package containing core logic
- CLI binary links library
- Web server imports library
- Common in Python, Node.js, .NET

**Sources:**
- https://developersvoice.com/blog/architecture/dot-net-maui-for-architects/

**Pros:**
- ✅ Simple to implement
- ✅ No network overhead
- ✅ Direct function calls

**Cons:**
- ❌ Must be same language/runtime
- ❌ Versioning can be complex
- ❌ Less flexible than service-oriented

---

## Recommended Architecture for Vibe Coding Dashboard

### Hybrid Approach: Hexagonal + Shared Library

```
┌─────────────────────────────────────────┐
│           CLI Interface                  │
│    (Bubble Tea/Cobra Dashboard)         │
└────────────────┬────────────────────────┘
                 │
                 ├─────────────┐
                 │             │
         ┌───────▼──────┐ ┌───▼────────────┐
         │ CLI Adapter  │ │ Web Adapter    │
         └───────┬──────┘ └───┬────────────┘
                 │             │
                 └──────┬──────┘
                        │
         ┌──────────────▼──────────────────┐
         │      Core Domain Logic          │
         │  (Project State, Hibernation,   │
         │   File Monitoring, Method       │
         │   Detection)                    │
         └──────────────┬──────────────────┘
                        │
         ┌──────────────▼──────────────────┐
         │     Data Access Layer           │
         │  (Local SQLite, File System,    │
         │   Optional Cloud Sync)          │
         └─────────────────────────────────┘
```

**Rationale:**
- **Start:** CLI-first with shared library
- **Core Logic:** Hexagonal architecture for clean separation
- **Future Web:** Add web adapter without touching core
- **Local-First:** SQLite for state, file system for projects
- **Optional Cloud:** Add sync service later without core changes

---

## Plugin Architecture for Extensibility

Supporting multiple vibe coding methodologies (BMAD-Method, Speckit, future methods) requires extensible plugin architecture.

### Model Context Protocol (MCP) Approach

**Overview:**
- Open standard for AI model <-> tool communication
- Protocol-driven extensibility
- Hosts (main app), Clients (connection managers), Servers (feature providers)

**Sources:**
- https://www.byteplus.com/en/topic/541395?title=mcp-plugin-architecture-design-integration-guide

**Benefits:**
- ✅ Standardized communication protocol
- ✅ New methods added without core changes
- ✅ Excellent for AI-assisted tools

---

### Classic Plugin Pattern with Dependency Injection ⭐ RECOMMENDED

**Implementation Approach:**

```go
// Core interface all method detectors must implement
type MethodDetector interface {
    DetectMethod(projectPath string) (*MethodInfo, error)
    GetWorkflowStage(projectPath string) (*StageInfo, error)
    GetProjectState(projectPath string) (*ProjectState, error)
}

// Plugin registry
type PluginRegistry struct {
    detectors map[string]MethodDetector
}

// Auto-discovery via reflection or explicit registration
func (r *PluginRegistry) Register(name string, detector MethodDetector) {
    r.detectors[name] = detector
}

// BMAD Method plugin
type BMADDetector struct {}
func (b *BMADDetector) DetectMethod(path string) (*MethodInfo, error) {
    // Check for .bmad folder, parse artifacts
}

// Speckit plugin  
type SpeckitDetector struct {}
func (s *SpeckitDetector) DetectMethod(path string) (*MethodInfo, error) {
    // Check for .specify/.speckit folder and specs/ directory
}
```

**Sources:**
- https://www.c-sharpcorner.com/article/extensible-asp-net-core-systems-building-plugin-based-architectures-with-reflec/
- https://softwarepatternslexicon.com/kotlin/architectural-patterns/plugin-architecture/
- https://dev.to/devleader/plugin-architecture-design-pattern-a-beginners-guide-to-modularity-4bo8

**Key Patterns:**
1. **Interface-Based:** Define contracts for method detection
2. **Reflection/Discovery:** Scan for plugin assemblies
3. **Dependency Injection:** Load and resolve plugins dynamically
4. **Factory Pattern:** Create plugin instances without hard-coding

**Benefits:**
- ✅ Add new methods without core changes
- ✅ Each method isolated and independently testable
- ✅ Community can contribute method plugins
- ✅ Version isolation per plugin

**Best Practices (2025):**
- Clear plugin contracts (interfaces)
- Explicit versioning
- Plugin lifecycle management
- Security: digital signatures, trusted plugins only
- Documentation for plugin developers

---

## State Management Patterns

Managing state across multiple projects requires robust real-time update mechanisms.

### Observer Pattern for Dashboard Updates ⭐ RECOMMENDED

**Architecture:**
```
┌──────────────┐
│ Project 1    │──────┐
└──────────────┘      │
                      │
┌──────────────┐      ▼      ┌──────────────┐
│ Project 2    │────────────>│  Dashboard   │
└──────────────┘      ▲      │  (Observer)  │
                      │      └──────────────┘
┌──────────────┐      │
│ Project 3    │──────┘
└──────────────┘
```

**Implementation Pattern:**
- **Subjects:** Project state monitors
- **Observers:** Dashboard widgets/views
- **Events:** Project state changes notify dashboard

**Sources:**
- https://www.momentslog.com/development/design-pattern/observer-pattern-in-real-time-analytics-dashboards
- https://codezup.com/observer-pattern-real-time-data-updates-modern-apps/
- https://dev.to/brdnicolas/mastering-real-time-magic-the-observer-pattern-1l0k

**Benefits:**
- ✅ Loose coupling between projects and dashboard
- ✅ Automatic synchronization
- ✅ Easy to add/remove observers at runtime
- ✅ Scales to hundreds of projects

**Modern Enhancements:**
- **Reactive Streams:** RxJS-style observables
- **Event Filtering:** Only notify relevant widgets
- **Debouncing:** Batch rapid updates
- **Memory Management:** Unsubscribe on widget close

---

### Centralized State Store

**Pattern:**
- Single source of truth for all project states
- Dashboard subscribes to global state
- State changes propagate automatically

**Implementation Options:**
- Redux/Vuex-style stores (if web later)
- Simple in-memory cache with file-backed persistence
- SQLite as state database

**Best for:**
- Cross-project dependencies
- Complex state relationships
- Undo/redo functionality

---

## Local-First Architecture with Cloud Sync

Local-first architecture treats the local device as the authoritative source of data, with cloud synchronization as a background process.

### Key Characteristics (2025)

**1. Immediate Responsiveness:**
- All user actions against local state
- Sub-100ms latency even offline
- No network dependency for core functionality

**2. Offline By Default:**
- Indefinite offline operation
- Automatic sync when connectivity returns
- No data loss during disconnection

**3. Convergence and Consistency:**
- Conflict-Free Replicated Data Types (CRDTs)
- Independent edits merge automatically
- No central arbiter required

**Sources:**
- https://debugg.ai/resources/local-first-apps-2025-crdts-replication-edge-storage-offline-sync
- https://developersvoice.com/blog/mobile/offline-first-sync-patterns/
- https://makitsol.com/offline-first/

---

### Implementation Strategies

**1. CRDTs for Data Synchronization:**
- Automatic conflict resolution
- Multiwriter collaboration support
- Guarantees convergence without central server

**Data Types:**
- Registers (Last-Write-Wins)
- Counters (increment-only)
- Sets (add/remove operations)
- Maps and Lists with CRDT semantics

**Sources:**
- https://dev.to/neon-postgres/comparing-local-first-frameworks-and-approaches-1hgn

**2. Cloud Sync Patterns:**

**WebSocket/Service Worker Background Sync:**
- Real-time or queued updates
- Delta/Outbox pattern for efficiency
- Local writes enqueued and synced when possible

**Edge Storage:**
- Cloudflare Durable Objects
- Turso SQLite at the edge
- Browser local storage (OPFS, IndexedDB)

**Sources:**
- https://alexop.dev/posts/building-local-first-apps-vue-dexie/
- https://www.enamic.io/resources/powersync-real-time-data-synchronization-2025

**3. Storage Options:**

**SQLite with OPFS (Origin Private File System):**
- Near-native mobile/desktop speeds
- Robust transactionality
- Excellent for complex, high-volume data

**IndexedDB:**
- Universal browser support
- Good for simple key-value storage
- Less suitable for relational workloads

**Performance Comparison (2025):**
- **SQLite+OPFS:** Near-native speeds, SQL support, best for multi-table operations
- **IndexedDB:** Slower for batch operations, verbose API, good for basic caching

**Sources:**
- https://www.powersync.com/blog/powersync-2025-roadmap-sqlite-web-speed-and-versatility
- https://rxdb.info/articles/localstorage-indexeddb-cookies-opfs-sqlite-wasm.html
- https://blog.logrocket.com/offline-first-frontend-apps-2025-indexeddb-sqlite/

---

### Conflict Resolution Techniques

**Last-Write-Wins (LWW):**
- Simplest approach
- Timestamp-based resolution
- Good for single-user or low-conflict scenarios

**Three-Way Merge:**
- Compare original, local, and remote versions
- Detect and resolve conflicts intelligently
- Better for collaborative scenarios

**CRDT Automatic Resolution:**
- No manual conflict handling
- Deterministic convergence
- Best for high-conflict collaborative environments

---

### Recommended Approach for Vibe Dashboard

**Phase 1 (MVP):**
- **Local-Only:** SQLite database in project .vibe directory
- **File-Based State:** Project metadata in .vibe/state.db
- **No Cloud:** Focus on local-first functionality

**Phase 2 (Optional Cloud Sync):**
- **CRDT Layer:** Add CRDT wrapper for project state
- **Sync Service:** Optional background sync to cloud
- **Conflict Resolution:** Automatic via CRDTs
- **Edge Storage:** Turso SQLite or similar

**Storage Architecture:**
```
~/.vibe/
├── config.yaml          # Global configuration
├── global.db            # SQLite: all projects registry
└── cache/               # Cached metadata

/project-path/.bmad/     # or .speckit or specs/
├── artifacts/           # Method-specific workflow files
└── .vibe/
    └── state.db         # SQLite: project-specific state
```

---

## Architecture Decision Summary

### For CLI-First Development

**Recommended Stack:**
1. **Architecture Pattern:** Hexagonal (Ports and Adapters)
2. **Plugin System:** Interface-based with dependency injection
3. **State Management:** Observer pattern with centralized SQLite store
4. **Storage:** Local-first SQLite, optional cloud sync later

**Rationale:**
- Clean separation enables future web expansion
- Plugin architecture supports multiple vibe coding methods
- Observer pattern provides real-time dashboard updates
- Local-first ensures performance and offline capability

---

## Sources and References

**Architecture Patterns:**
- https://learn.microsoft.com/en-us/dotnet/architecture/modern-web-apps-azure/common-web-application-architectures
- https://www.digitalplatformarchitect.com/patterns
- https://www.storifyagency.com/web-application-architecture-your-2025/

**Plugin Architecture:**
- https://www.byteplus.com/en/topic/541395?title=mcp-plugin-architecture-design-integration-guide
- https://www.c-sharpcorner.com/article/extensible-asp-net-core-systems-building-plugin-based-architectures-with-reflec/
- https://dev.to/devleader/plugin-architecture-design-pattern-a-beginners-guide-to-modularity-4bo8

**State Management:**
- https://www.momentslog.com/development/design-pattern/observer-pattern-in-real-time-analytics-dashboards
- https://codezup.com/observer-pattern-real-time-data-updates-modern-apps/

**Local-First:**
- https://debugg.ai/resources/local-first-apps-2025-crdts-replication-edge-storage-offline-sync
- https://www.powersync.com/blog/powersync-2025-roadmap-sqlite-web-speed-and-versatility
- https://rxdb.info/articles/localstorage-indexeddb-cookies-opfs-sqlite-wasm.html

---

**Shard Status:** ✅ Complete  
**Last Updated:** 2025-12-04T07:10:42.556Z  
**Confidence Level:** High (authoritative architectural patterns, 2025 verified)

**Previous Shard:** [01-technology-stack.md](./01-technology-stack.md)  
**Next Shard:** [03-implementation-techniques.md](./03-implementation-techniques.md)
