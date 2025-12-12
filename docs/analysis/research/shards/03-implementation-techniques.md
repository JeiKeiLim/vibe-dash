# Implementation Techniques

**Shard 3 of 9 - Technical Research**

---

## Overview

This shard covers practical implementation techniques for file system monitoring, process state detection, and workflow state extraction for vibe coding dashboards.

---

## File System Monitoring

Modern file system monitoring enables real-time detection of project changes across platforms.

### Tool Comparison (2025)

#### 1. Watchman (Facebook)

**Overview:**
- **Platforms:** Linux, Windows, macOS
- **Performance:** Handles hundreds of thousands of files
- **Features:** Event-driven, triggers for automation, detailed events

**Best Practices:**
- Configure exclusions (node_modules, .git, etc.)
- Use triggers for automatic task execution
- Limit watch scope for stability

**Sources:**
- https://itsfoss.gitlab.io/post/watchman--a-file-and-directory-watching-tool-for-changes/

---

#### 2. Chokidar (Node.js)

**Overview:**
- **Platforms:** Cross-platform (Linux, Windows, macOS)
- **Performance:** Handles thousands of files efficiently
- **Features:** Native OS events, debouncing, atomic write handling

**API Example:**
```javascript
const chokidar = require('chokidar');

const watcher = chokidar.watch('/project-path/.bmad', {
  ignored: /(^|[\/\\])\../,  // ignore dotfiles
  persistent: true,
  ignoreInitial: false
});

watcher
  .on('add', path => console.log(`File ${path} added`))
  .on('change', path => console.log(`File ${path} changed`))
  .on('unlink', path => console.log(`File ${path} removed`));
```

**Best Practices:**
- Prefer native events over polling
- Restrict scope with `ignored` option
- Handle error events
- Set reasonable recursion depth

**Sources:**
- https://github.com/paulmillr/chokidar
- https://www.w3tutorials.net/blog/nodejs-chokidar/
- https://www.linuxlinks.com/chokidar-watch-file-system-changes/

---

#### 3. fsnotify (Go) ⭐ RECOMMENDED

**Overview:**
- **Platforms:** Cross-platform (inotify, kqueue, ReadDirectoryChangesW)
- **Performance:** Straightforward, efficient for Go applications
- **Use Cases:** Config reloading, backup triggers, job automation

**API Example:**
```go
import (
    "log"
    "github.com/fsnotify/fsnotify"
)

watcher, err := fsnotify.NewWatcher()
if err != nil {
    log.Fatal(err)
}
defer watcher.Close()

err = watcher.Add("/project-path/.bmad")
if err != nil {
    log.Fatal(err)
}

for {
    select {
    case event := <-watcher.Events:
        if event.Op&fsnotify.Write == fsnotify.Write {
            log.Println("Modified file:", event.Name)
        }
    case err := <-watcher.Errors:
        log.Println("Error:", err)
    }
}
```

**Best Practices:**
- Monitor only necessary files/directories
- Validate files before acting (avoid partial writes)
- Use channels for event delivery
- Handle both events and errors

**Sources:**
- https://stackoverflow.com/questions/72741501/monitoring-an-existing-file-with-fsnotify

---

### Recommended Approach for Vibe Dashboard

**Go with fsnotify:**
- ✅ Native integration with Go/Cobra stack
- ✅ Cross-platform support
- ✅ Efficient for monitoring multiple project directories
- ✅ Clean channel-based API fits Go patterns

**Watch Strategy:**
```go
type ProjectWatcher struct {
    projects map[string]*fsnotify.Watcher
    updates  chan ProjectUpdate
}

// Watch specific artifact files, not entire directory
func (pw *ProjectWatcher) WatchProject(path string) {
    // Watch .bmad/artifacts/*.md or specs/NNN-*/*.md
    // Ignore temp files, editor backups
    // Debounce rapid changes
    // Notify dashboard on significant changes
}
```

**Debouncing Implementation:**
```go
type Debouncer struct {
    delay  time.Duration
    timers map[string]*time.Timer
    mu     sync.Mutex
}

func (d *Debouncer) Add(key string, fn func()) {
    d.mu.Lock()
    defer d.mu.Unlock()
    
    // Cancel existing timer
    if timer, exists := d.timers[key]; exists {
        timer.Stop()
    }
    
    // Create new timer
    d.timers[key] = time.AfterFunc(d.delay, func() {
        fn()
        d.mu.Lock()
        delete(d.timers, key)
        d.mu.Unlock()
    })
}
```

---

## Process State Detection

Detecting when AI coding agents are waiting for user input requires understanding process states across platforms.

### Linux Process States

**State Codes:**
- `R` - Running or runnable
- `S` - Interruptible sleep (waiting for input)
- `D` - Uninterruptible sleep (disk I/O)
- `Z` - Zombie
- `T` - Stopped

**Detection Methods:**

**1. /proc/<PID>/stat parsing:**
```bash
cat /proc/$PID/stat | awk '{print $3}'
# Returns: S (sleeping, likely waiting for input)
```

**2. strace syscall monitoring:**
```bash
strace -p $PID 2>&1 | grep "read(0,"
# read(0, ...) means reading from stdin (fd 0)
```

**3. lsof file descriptor checking:**
```bash
lsof -p $PID | grep stdin
# Check if stdin is open and process is sleeping
```

**Sources:**
- https://www.baeldung.com/linux/process-states
- https://tech-champion.com/linux/how-to-monitor-linux-processes-with-command-line-tools/

---

### Windows Process States

**Tools:**
- **Process Explorer:** Thread states, stack traces
- **Process Monitor (Procmon):** Real-time thread activity
- **PowerShell:** Script-based monitoring

**Detection:**
- Thread blocked on `ReadConsoleInput`
- Zero CPU usage with open stdin handle
- Stack trace shows input wait

**Sources:**
- https://learn.microsoft.com/en-us/sysinternals/downloads/procmon
- https://www.ittsystems.com/best-process-operating-system-monitoring-tools/

---

### Modern Monitoring Tools (2025)

**denet (Linux):**
- Real-time process monitor (Rust/Python)
- Tracks process state per thread
- Identifies waiting processes efficiently

**Sources:**
- https://arxiv.org/abs/2510.13818

---

### Recommended Approach for Agent Waiting Detection

**Strategy:**
- **Not Real-Time Detection:** Too complex and platform-specific
- **Heuristic-Based:** Infer from file timestamps and workflow state

**Implementation:**
```go
type AgentState struct {
    LastActivity   time.Time
    CurrentStage   string
    AwaitingInput  bool
}

func DetectAgentWaiting(projectPath string) bool {
    // Heuristic detection based on file timestamps and workflow stage
    // No reliance on agent-created markers
    
    // Check last activity timestamp
    lastModified := getLastModified(path.Join(projectPath, ".bmad/artifacts"))
    if time.Since(lastModified) > 1*time.Hour {
        // Check if at interactive stage (PRD review, etc.)
        stage := getCurrentStage(projectPath)
        if isInteractiveStage(stage) {
            return true
        }
    }
    
    return false
}
```

**Heuristic Detection Approach:**
- Infer waiting state from file timestamps
- Check workflow stage for interactive phases (review, approval)
- No agent cooperation required
- Works with existing BMAD-Method implementation

---

## Workflow State Extraction

Extracting workflow states requires parsing method-specific artifact structures.

### General Strategy

```go
type WorkflowState struct {
    Method         string    // "bmad", "speckit", etc.
    CurrentPhase   string    // Method-specific phase
    LastActivity   time.Time
    Progress       int       // 0-100
    AgentWaiting   bool
}

func ExtractWorkflowState(projectPath string) (*WorkflowState, error) {
    // 1. Detect method
    method := detectMethod(projectPath)
    
    // 2. Get method-specific detector
    detector := registry.Get(method)
    
    // 3. Parse state
    return detector.GetProjectState(projectPath)
}
```

### File Parsing Techniques

**YAML Parsing (Go):**
```go
import "gopkg.in/yaml.v3"

type WorkflowConfig struct {
    CurrentPhase string    `yaml:"current_phase"`
    LastStep     int       `yaml:"last_step"`
    StepsComplete []int    `yaml:"steps_completed"`
}

func parseWorkflowYAML(path string) (*WorkflowConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var config WorkflowConfig
    err = yaml.Unmarshal(data, &config)
    return &config, err
}
```

**Markdown Frontmatter Parsing:**
```go
import "github.com/BurntSushi/toml"

func parseFrontmatter(path string) (map[string]interface{}, error) {
    content, _ := os.ReadFile(path)
    
    // Extract frontmatter between --- markers
    // Parse as YAML or TOML
    // Return metadata map
}
```

**File Timestamp Analysis:**
```go
func getLastModifiedInDir(dirPath string) time.Time {
    var latest time.Time
    
    filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() && info.ModTime().After(latest) {
            latest = info.ModTime()
        }
        return nil
    })
    
    return latest
}
```

---

## Implementation Best Practices

### File Monitoring

1. **Watch Scope:** Monitor specific directories (.bmad, .speckit, specs/)
2. **Ignore Patterns:** Exclude temp files, editor backups, node_modules
3. **Debouncing:** Batch rapid changes (300-500ms delay)
4. **Error Handling:** Graceful degradation on permission errors
5. **Resource Management:** Close watchers when projects removed

### State Detection

1. **Caching:** Cache parsed states, refresh on file changes
2. **Fallback:** If parsing fails, use last known good state
3. **Heuristics:** Combine file timestamps, stage markers, explicit config
4. **Validation:** Sanity check extracted states

### Performance

1. **Lazy Loading:** Parse project states on-demand
2. **Background Refresh:** Update states in background goroutines
3. **Rate Limiting:** Throttle dashboard updates to 1-2 per second
4. **Memory Management:** Limit in-memory state cache size

---

## Sources and References

**File System Monitoring:**
- https://itsfoss.gitlab.io/post/watchman--a-file-and-directory-watching-tool-for-changes/
- https://github.com/paulmillr/chokidar
- https://www.w3tutorials.net/blog/nodejs-chokidar/
- https://stackoverflow.com/questions/72741501/monitoring-an-existing-file-with-fsnotify

**Process Monitoring:**
- https://www.baeldung.com/linux/process-states
- https://tech-champion.com/linux/how-to-monitor-linux-processes-with-command-line-tools/
- https://learn.microsoft.com/en-us/sysinternals/downloads/procmon
- https://arxiv.org/abs/2510.13818

---

**Shard Status:** ✅ Complete  
**Last Updated:** 2025-12-04T07:11:52.043Z  
**Confidence Level:** High (practical implementation patterns, verified)

**Previous Shard:** [02-architectural-patterns.md](./02-architectural-patterns.md)  
**Next Shard:** [04-vibe-coding-methods.md](./04-vibe-coding-methods-CORRECTED.md)
