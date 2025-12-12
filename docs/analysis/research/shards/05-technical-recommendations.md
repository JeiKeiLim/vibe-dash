# Technical Recommendations

**Shard 5 of 9 - Technical Research**

## Technology Stack: Go + Cobra + Bubble Tea

**Recommendation Summary:**
- Primary language: Go
- CLI framework: Cobra
- TUI library: Bubble Tea + Lip Gloss + Bubbles
- File monitoring: fsnotify
- Storage: SQLite
- State management: Observer pattern

**See:** [Shard 1](./01-technology-stack.md) for detailed comparison and benchmarks.

---

## Architecture: Hexagonal + Plugin System

**Core Components:**
1. CLI Adapter Layer (Bubble Tea dashboard)
2. Core Domain Logic (project management, method detection, file monitoring)
3. Data Access Layer (SQLite + file system)

**Plugin System:**
- Interface-based method detectors
- BMAD and Speckit detectors built-in
- Extensible for future methods

**See:** [Shard 2](./02-architectural-patterns.md) for complete architecture design.

---

## Implementation Blueprint

**File Monitoring:**
```go
// Use fsnotify with debouncing
type ProjectMonitor struct {
    watcher   *fsnotify.Watcher
    debouncer *Debouncer
    updates   chan ProjectUpdate
}
```

**Method Detection:**
```go
// Plugin interface
type MethodDetector interface {
    Name() string
    Detect(projectPath string) (bool, error)
    GetProjectState(projectPath string) (*ProjectState, error)
}
```

**State Storage:**
```sql
-- ~/.vibe/global.db
CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    path TEXT UNIQUE NOT NULL,
    method TEXT,
    state TEXT,
    last_activity TIMESTAMP
);
```

**See:** [Shard 3](./03-implementation-techniques.md) for complete implementation details.

---

**Status:** ✅ Complete  
**Last Updated:** 2025-12-04T07:11:52.043Z  
**See Also:** Shards 1, 2, 3, 6, 7 for complete technical specifications
