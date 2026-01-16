package tui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"

	"github.com/JeiKeiLim/vibe-dash/internal/adapters/filesystem"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/components"
	"github.com/JeiKeiLim/vibe-dash/internal/adapters/tui/statsview"
	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/emoji"
	"github.com/JeiKeiLim/vibe-dash/internal/shared/project"
)

// Story 12.2 AC2: Timeout for double-key detection (e.g., 'gg' for jump to top)
const ggTimeoutMs = 500

// metricsRecorderInterface defines the contract for recording stage transitions (Story 16.2).
// Used to decouple TUI from concrete metrics implementation for testability.
type metricsRecorderInterface interface {
	OnDetection(ctx context.Context, projectID string, newStage domain.Stage)
}

// metricsReaderInterface defines the contract for reading metrics data (Story 16.4, 16.5).
// Used to decouple TUI from concrete metrics repository for testability.
// Supports both sparklines (timestamps) and breakdown (full transitions).
type metricsReaderInterface interface {
	GetTransitionTimestamps(ctx context.Context, projectID string, since time.Time) []statsview.Transition
	GetFullTransitions(ctx context.Context, projectID string, since time.Time) []statsview.FullTransition
}

// Model represents the main TUI application state.
type Model struct {
	width            int  // Terminal width (from WindowSizeMsg)
	height           int  // Terminal height (from WindowSizeMsg)
	ready            bool // True after first WindowSizeMsg received
	showHelp         bool // Toggle help overlay
	hasPendingResize bool // True when resize is pending (debounce)
	pendingWidth     int  // Buffered width for debounced resize
	pendingHeight    int  // Buffered height for debounced resize

	// Path validation state
	viewMode          viewMode
	invalidProjects   []InvalidProject
	currentInvalidIdx int
	validationError   string // Error message to display in validation dialog

	// Project list state (Story 3.1)
	projects    []*domain.Project
	projectList components.ProjectListModel

	// Detail panel state (Story 3.3)
	showDetailPanel bool
	detailPanel     components.DetailPanelModel

	// Status bar state (Story 3.4)
	statusBar components.StatusBarModel

	// Refresh state (Story 3.6)
	isRefreshing    bool
	refreshTotal    int
	refreshProgress int
	refreshError    string

	// Note editing state (Story 3.7)
	isEditingNote  bool
	noteInput      textinput.Model // From charmbracelet/bubbles
	originalNote   string          // For cancel restoration
	noteEditTarget *domain.Project // Project being edited
	noteFeedback   string          // "✓ Note saved" or error message

	// Remove confirmation state (Story 3.9)
	isConfirmingRemove bool
	confirmTarget      *domain.Project // Project pending removal

	// Dependencies (injected)
	repository       ports.ProjectRepository
	detectionService ports.Detector        // Optional - may be nil if not wired
	waitingDetector  ports.WaitingDetector // Story 4.5: Optional - for WAITING indicator

	// Story 4.6: File watcher for real-time dashboard updates
	fileWatcher          ports.FileWatcher
	eventCh              <-chan ports.FileEvent
	watchCtx             context.Context
	watchCancel          context.CancelFunc
	fileWatcherAvailable bool // false if watcher failed to start

	// Story 7.2: Config warning state
	configWarning     string    // Config error message to display
	configWarningTime time.Time // When warning was set (for auto-clearing)

	// Story 7.3: Corrupted projects state
	corruptedProjects []string // Names of projects with corrupted databases

	// Story 7.4: Loading state for initial project load
	isLoading bool

	// Story 8.4: Pending projects waiting for ready state (race condition fix)
	pendingProjects []*domain.Project

	// Story 8.6: Layout configuration
	detailLayout string // "vertical" (left/right) or "horizontal" (top/bottom)

	// Story 8.7: Config for help overlay display
	config *ports.Config

	// Story 8.10: Max content width from config (0 = unlimited)
	maxContentWidth int

	// Story 8.11: Periodic stage re-detection interval (0 = disabled)
	stageRefreshInterval int

	// Story 9.5-2: Grace period for restart race condition
	lastWatcherRestart time.Time

	// Story 9.5-2 fix: Track if stage refresh timer has been started (prevent duplicates)
	stageTimerStarted bool

	// Story 11.2: Hibernation service for auto-hibernation
	hibernationService      ports.HibernationService
	hibernationTimerStarted bool // Prevent duplicate hourly timers

	// Story 11.3: State service for auto-activation on file events
	stateService ports.StateActivator

	// Story 11.4: Hibernated projects view state
	hibernatedProjects     []*domain.Project
	hibernatedList         components.ProjectListModel
	activeSelectedIdx      int    // Preserve selection when switching views
	justActivatedProjectID string // Track which project to select after activation (AC3)

	// Story 12.1: Log viewer state
	logReaderRegistry  ports.LogReaderRegistry
	currentLogReader   ports.LogReader     // Active log reader for the current project
	currentLogProject  *domain.Project     // Project whose logs are being viewed
	showSessionPicker  bool                // Whether session picker overlay is visible
	logSessions        []domain.LogSession // Available sessions for session picker
	sessionPickerIndex int                 // Selected index in session picker

	// Story 12.1: Text view state (for displaying jq-formatted logs)
	textViewContent []string // Lines of text to display
	textViewTitle   string   // Title for the text view header
	textViewScroll  int      // Current scroll position (line offset)

	// Story 12.1: Flash message state (AC8)
	flashMessage     string
	flashMessageTime time.Time

	// Story 12.1: Live tailing state (AC3, AC4)
	currentSessionPath string // Path to session file being tailed
	logAutoScroll      bool   // True = auto-scroll to latest, false = paused
	logLastOffset      int64  // Last read offset for incremental reads
	logTailActive      bool   // Whether tailing is active

	// Story 12.2 AC2: Double-key detection for 'gg' jump to top
	lastKeyPress string    // Last key pressed in text view
	lastKeyTime  time.Time // Time of last key press

	// Story 12.2 AC3: Search state for log viewer
	searchMode    bool   // Whether search mode is active
	searchQuery   string // Current search query (after Enter)
	searchInput   string // Text being typed (before Enter)
	searchIndex   int    // Current match index (0-based)
	searchMatches []int  // Line numbers with matches

	// Story 16.2: Metrics recorder for stage transition tracking
	metricsRecorder metricsRecorderInterface

	// Story 16.3: Stats view state
	statsViewScroll       int // Scroll position in stats view
	statsActiveProjectIdx int // Saved dashboard selection for restoration

	// Story 16.4: Metrics reader for stats view sparklines (optional, graceful nil)
	metricsReader metricsReaderInterface

	// Story 16.5: Stats View breakdown state
	statsBreakdownProject   *domain.Project           // Currently selected project for breakdown (nil = list view)
	statsBreakdownDurations []statsview.StageDuration // Cached durations for display

	// Story 16.6: Stats View date range state
	statsDateRange statsview.DateRange // Current date range preset (initialized on Stats View entry)
}

// resizeTickMsg is used for resize debouncing.
type resizeTickMsg struct{}

// Message types for async validation operations
type validationCompleteMsg struct {
	invalidProjects []InvalidProject
}

type validationErrorMsg struct {
	err error
}

type deleteProjectMsg struct {
	projectID string
	err       error
}

type moveProjectMsg struct {
	projectID string
	newPath   string
	err       error
}

type keepProjectMsg struct {
	projectID string
	err       error
}

// ProjectsLoadedMsg is sent when projects are loaded from the repository.
type ProjectsLoadedMsg struct {
	projects []*domain.Project
	err      error
}

// refreshCompleteMsg signals refresh is complete (Story 3.6).
type refreshCompleteMsg struct {
	refreshedCount int
	failedCount    int
	err            error // Only set if ALL projects failed
}

// clearRefreshMsgMsg signals to clear the refresh completion message (Story 3.6).
// Sent after 3-second timer expires.
type clearRefreshMsgMsg struct{}

// noteSavedMsg signals note was saved successfully (Story 3.7).
type noteSavedMsg struct {
	projectID string
	newNote   string
}

// noteSaveErrorMsg signals note save failed (Story 3.7).
type noteSaveErrorMsg struct {
	err error
}

// clearNoteFeedbackMsg signals to clear note feedback message (Story 3.7).
type clearNoteFeedbackMsg struct{}

// favoriteSavedMsg signals favorite was toggled successfully (Story 3.8).
type favoriteSavedMsg struct {
	projectID  string
	isFavorite bool
}

// favoriteSaveErrorMsg signals favorite save failed (Story 3.8).
type favoriteSaveErrorMsg struct {
	err error
}

// clearFavoriteFeedbackMsg signals to clear favorite feedback message (Story 3.8).
type clearFavoriteFeedbackMsg struct{}

// removeConfirmedMsg signals project removal was confirmed (Story 3.9).
type removeConfirmedMsg struct {
	projectID   string
	projectName string
	err         error
}

// removeConfirmTimeoutMsg signals confirmation dialog timed out (Story 3.9).
type removeConfirmTimeoutMsg struct{}

// clearRemoveFeedbackMsg signals to clear remove feedback message (Story 3.9).
type clearRemoveFeedbackMsg struct{}

// tickMsg is sent every 5 seconds for responsive waiting detection (Story 8.2, AC2).
// Originally 60 seconds for Story 4.2 AC4, reduced for 24/7 reliability.
type tickMsg time.Time

// stageRefreshTickMsg triggers periodic stage re-detection (Story 8.11).
type stageRefreshTickMsg time.Time

// hibernationCompleteMsg signals auto-hibernation check is complete (Story 11.2).
type hibernationCompleteMsg struct {
	count int
	err   error
}

// hibernationTickMsg triggers hourly auto-hibernation check (Story 11.2).
type hibernationTickMsg time.Time

// hibernatedProjectsLoadedMsg signals hibernated projects loaded (Story 11.4).
type hibernatedProjectsLoadedMsg struct {
	projects []*domain.Project
	err      error
}

// projectActivatedMsg signals a project was activated (Story 11.4).
type projectActivatedMsg struct {
	projectID   string
	projectName string
	err         error
}

// stateToggledMsg signals a project state was toggled via H key (Story 11.7).
type stateToggledMsg struct {
	projectID   string
	projectName string
	action      string // "hibernated" or "activated"
	err         error
}

// fileEventMsg wraps file system events for Bubble Tea message passing (Story 4.6).
type fileEventMsg struct {
	Path      string
	Operation ports.FileOperation
	Timestamp time.Time
}

// fileWatcherErrorMsg signals a file watcher error (Story 4.6).
type fileWatcherErrorMsg struct {
	err error
}

// watcherWarningMsg signals partial file watcher failures (Story 7.1).
// Sent when some (but not all) paths fail to watch during startup.
type watcherWarningMsg struct {
	failedPaths []string
	totalPaths  int
}

// configWarningMsg signals configuration loading errors (Story 7.2).
// Sent when config file has syntax errors or invalid values.
type configWarningMsg struct {
	warning string
}

// clearConfigWarningMsg signals to clear the config warning after timeout (Story 7.2).
type clearConfigWarningMsg struct{}

// projectCorruptionMsg signals project database corruption (Story 7.3).
// Sent when projects are skipped due to corrupted state.db files.
type projectCorruptionMsg struct {
	projects []string // Names of corrupted projects
}

// Story 12.1: Log session picker message types
// logSessionsLoadedMsg signals sessions have been loaded for picker.
type logSessionsLoadedMsg struct {
	sessions []domain.LogSession
	err      error
}

// textViewContentMsg signals jq-formatted content is ready to display.
type textViewContentMsg struct {
	title       string
	content     string
	sessionPath string // For live tailing (AC4)
	fileSize    int64  // Initial file size for offset tracking
	err         error
}

// flashMsg triggers a flash message display.
type flashMsg struct {
	text string
}

// clearFlashMsg clears the flash message.
type clearFlashMsg struct{}

// logTailTickMsg triggers live tailing poll for new log entries (AC4: 2s interval).
type logTailTickMsg time.Time

// logNewEntriesMsg signals new entries were found during tailing.
type logNewEntriesMsg struct {
	newContent string
	newOffset  int64 // Updated file offset for next poll
	err        error
}

// NewModel creates a new Model with default values.
// The repository parameter is used for project persistence operations.
func NewModel(repo ports.ProjectRepository) Model {
	defaults := ports.NewConfig()
	return Model{
		ready:           false,
		showHelp:        false,
		showDetailPanel: false, // Default closed, set based on height in resizeTickMsg
		viewMode:        viewModeNormal,
		repository:      repo,
		statusBar:       components.NewStatusBarModel(0), // Width set in resizeTickMsg
		detailLayout:    "horizontal",                    // Story 8.6: Default layout mode
		maxContentWidth: defaults.MaxContentWidth,        // Story 8.10: Default from config
	}
}

// SetDetectionService sets the detection service for refresh operations.
// This is optional - if not set, refresh will show "Detection service not available".
func (m *Model) SetDetectionService(svc ports.Detector) {
	m.detectionService = svc
}

// SetWaitingDetector sets the waiting detector for WAITING indicators (Story 4.5).
// This is optional - if not set, waiting indicators will not be shown.
func (m *Model) SetWaitingDetector(detector ports.WaitingDetector) {
	m.waitingDetector = detector
}

// SetFileWatcher sets the file watcher for real-time updates (Story 4.6).
// This is optional - if not set, file watching is disabled.
func (m *Model) SetFileWatcher(watcher ports.FileWatcher) {
	m.fileWatcher = watcher
	m.fileWatcherAvailable = true // Assume available until proven otherwise
}

// SetDetailLayout configures the detail panel layout mode (Story 8.6).
// Supports "horizontal" (default, stacked top/bottom) and "vertical" (side-by-side).
func (m *Model) SetDetailLayout(layout string) {
	if layout == "horizontal" || layout == "vertical" {
		m.detailLayout = layout
	} else {
		m.detailLayout = "horizontal" // Fallback to default
	}
}

// isHorizontalLayout returns true if detail panel should be below project list (Story 8.6).
func (m Model) isHorizontalLayout() bool {
	return m.detailLayout == "horizontal"
}

// SetConfig stores config for help overlay display and max content width (Story 8.7, 8.10, 8.11).
// Config passed as parameter to avoid cli→tui→cli import cycle.
// Nil-safe: stores defaults if cfg is nil.
func (m *Model) SetConfig(cfg *ports.Config) {
	if cfg == nil {
		cfg = ports.NewConfig()
	}
	m.config = cfg
	m.maxContentWidth = cfg.MaxContentWidth                  // Story 8.10
	m.stageRefreshInterval = cfg.StageRefreshIntervalSeconds // Story 8.11
}

// SetHibernationService sets the hibernation service for auto-hibernation (Story 11.2).
// This is optional - if not set, auto-hibernation is disabled.
func (m *Model) SetHibernationService(svc ports.HibernationService) {
	m.hibernationService = svc
}

// SetStateService sets the StateService for auto-activation on file events (Story 11.3).
// This is optional - if not set, auto-activation is disabled.
func (m *Model) SetStateService(svc ports.StateActivator) {
	m.stateService = svc
}

// SetLogReaderRegistry sets the log reader registry for log viewing (Story 12.1).
// This is optional - if not set, log viewing shows "not available" message.
func (m *Model) SetLogReaderRegistry(registry ports.LogReaderRegistry) {
	m.logReaderRegistry = registry
}

// SetMetricsRecorder sets the metrics recorder for stage transition tracking (Story 16.2).
// This is optional - if not set, metrics recording is disabled.
func (m *Model) SetMetricsRecorder(recorder metricsRecorderInterface) {
	m.metricsRecorder = recorder
}

// SetMetricsReader sets the metrics reader for stats view sparklines (Story 16.4).
// This is optional - if not set, sparklines show empty (graceful degradation).
func (m *Model) SetMetricsReader(reader metricsReaderInterface) {
	m.metricsReader = reader
}

// isProjectWaiting wraps WaitingDetector.IsWaiting for component callbacks.
// Uses context.Background() since Bubble Tea Render() doesn't provide ctx.
// Story 4.5: Returns false if detector is nil.
func (m Model) isProjectWaiting(p *domain.Project) bool {
	if m.waitingDetector == nil {
		return false
	}
	return m.waitingDetector.IsWaiting(context.Background(), p)
}

// getWaitingDuration wraps WaitingDetector.WaitingDuration for component callbacks.
// Uses context.Background() since Bubble Tea Render() doesn't provide ctx.
// Story 4.5: Returns 0 if detector is nil.
func (m Model) getWaitingDuration(p *domain.Project) time.Duration {
	if m.waitingDetector == nil {
		return 0
	}
	return m.waitingDetector.WaitingDuration(context.Background(), p)
}

// getAgentState wraps WaitingDetector.AgentState for component callbacks.
// Uses context.Background() since Bubble Tea Render() doesn't provide ctx.
// Story 15.7: Returns empty state if detector is nil.
func (m Model) getAgentState(p *domain.Project) domain.AgentState {
	if m.waitingDetector == nil {
		return domain.AgentState{}
	}
	return m.waitingDetector.AgentState(context.Background(), p)
}

// getProjectActivity fetches activity counts for sparkline generation (Story 16.4).
// Returns nil if metricsReader is not available (graceful degradation).
// buckets: number of time buckets for the sparkline (typically 7-14).
// Story 16.6: Uses m.statsDateRange for date range filtering.
func (m Model) getProjectActivity(projectID string, buckets int) []int {
	if m.metricsReader == nil {
		return nil // Graceful degradation
	}
	ctx := context.Background()
	now := time.Now()
	since := m.statsDateRange.Since() // Story 16.6: Use selected date range

	transitions := m.metricsReader.GetTransitionTimestamps(ctx, projectID, since)
	if len(transitions) == 0 {
		return nil
	}

	// Extract timestamps for bucketing
	timestamps := make([]time.Time, len(transitions))
	for i, t := range transitions {
		timestamps[i] = t.TransitionedAt
	}

	// Story 16.6: Calculate time range for bucketing
	timeRange := m.statsDateRange.Duration()
	if timeRange == 0 {
		// All time: calculate from earliest transition to now
		timeRange = statsview.CalculateTimeRangeFromTimestamps(timestamps, now)
	}

	return statsview.BucketActivityCounts(timestamps, buckets, timeRange, now)
}

// getStageBreakdown fetches stage durations for a project (Story 16.5).
// Returns nil if metricsReader is not available (graceful degradation).
// Story 16.6: Uses m.statsDateRange for date range filtering.
func (m Model) getStageBreakdown(projectID string) []statsview.StageDuration {
	if m.metricsReader == nil {
		return nil
	}
	ctx := context.Background()
	now := time.Now()
	since := m.statsDateRange.Since() // Story 16.6: Use selected date range
	transitions := m.metricsReader.GetFullTransitions(ctx, projectID, since)
	if len(transitions) == 0 {
		return nil
	}
	return statsview.CalculateFromFullTransitions(transitions, now)
}

// shouldShowDetailPanelByDefault returns true if detail panel should be open by default
// based on terminal height. Per AC7: >= HeightThresholdTall (35) rows = open, otherwise closed.
func shouldShowDetailPanelByDefault(height int) bool {
	return height >= HeightThresholdTall
}

// isNarrowWidth returns true if terminal width is in narrow range (60-79).
// Used for Story 3.10 AC2 narrow warning display.
func isNarrowWidth(width int) bool {
	return width >= MinWidth && width < 80
}

// isWideWidth returns true if terminal width exceeds config max_content_width.
// Story 3.10 AC4: Content capping and centering.
// Story 8.10: Uses config-based maxContentWidth (0 = unlimited, always returns false).
func (m Model) isWideWidth() bool {
	if m.maxContentWidth == 0 {
		return false // Unlimited width mode
	}
	return m.width > m.maxContentWidth
}

// statusBarHeight returns the status bar height based on terminal height.
// Returns 1 for condensed mode (height < 20), 2 for normal mode.
func statusBarHeight(height int) int {
	if height < MinHeight {
		return 1
	}
	return 2
}

// Init implements tea.Model. Returns commands to validate project paths and start periodic tick.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.checkAutoHibernationCmd(), // Story 11.2: Run FIRST before validation
		m.validatePathsCmd(),
		tickCmd(), // Start periodic timestamp refresh (Story 4.2, AC4)
	)
}

// checkAutoHibernationCmd creates a command that checks for auto-hibernation (Story 11.2).
// Returns nil if hibernation service is not set.
func (m Model) checkAutoHibernationCmd() tea.Cmd {
	if m.hibernationService == nil {
		return nil
	}
	return func() tea.Msg {
		count, err := m.hibernationService.CheckAndHibernate(context.Background())
		return hibernationCompleteMsg{count: count, err: err}
	}
}

// validatePathsCmd creates a command that validates all project paths.
func (m Model) validatePathsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		invalid, err := ValidateProjectPaths(ctx, m.repository)
		if err != nil {
			return validationErrorMsg{err}
		}
		return validationCompleteMsg{invalid}
	}
}

// loadProjectsCmd creates a command that loads active projects from the repository.
// Hibernated projects are loaded separately via loadHibernatedProjectsCmd.
func (m Model) loadProjectsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		projects, err := m.repository.FindActive(ctx)
		return ProjectsLoadedMsg{projects: projects, err: err}
	}
}

// tickCmd returns a command that ticks every 5 seconds for responsive waiting detection (Story 8.2).
// Story 4.2 AC4 originally used 1-minute interval, but 24/7 usage requires faster updates.
func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// stageRefreshTickCmd returns command for periodic stage detection (Story 8.11).
// Story 9.5-2 fix: Only starts timer once - subsequent calls return nil.
// The timer reschedules itself via stageRefreshTickMsg handler using rescheduleStageTimer.
func (m *Model) stageRefreshTickCmd() tea.Cmd {
	if m.stageRefreshInterval == 0 {
		return nil
	}
	// Only start timer once - it reschedules itself via stageRefreshTickMsg
	if m.stageTimerStarted {
		return nil
	}
	m.stageTimerStarted = true
	return m.rescheduleStageTimer()
}

// rescheduleStageTimer creates the actual timer tick. Used by stageRefreshTickMsg handler
// to reschedule after each tick. Does not check stageTimerStarted flag.
func (m Model) rescheduleStageTimer() tea.Cmd {
	if m.stageRefreshInterval == 0 {
		return nil
	}
	return tea.Tick(time.Duration(m.stageRefreshInterval)*time.Second, func(t time.Time) tea.Msg {
		return stageRefreshTickMsg(t)
	})
}

// hibernationTickCmd returns command for hourly hibernation check (Story 11.2, AC3).
// Only starts timer once - subsequent calls return nil.
// The timer reschedules itself via hibernationTickMsg handler.
func (m *Model) hibernationTickCmd() tea.Cmd {
	if m.hibernationService == nil {
		return nil
	}
	// Only start timer once - it reschedules itself via hibernationTickMsg
	if m.hibernationTimerStarted {
		return nil
	}
	m.hibernationTimerStarted = true
	return m.rescheduleHibernationTimer()
}

// rescheduleHibernationTimer creates the actual hourly timer tick (Story 11.2).
// Used by hibernationTickMsg handler to reschedule after each tick.
func (m Model) rescheduleHibernationTimer() tea.Cmd {
	if m.hibernationService == nil {
		return nil
	}
	return tea.Tick(time.Hour, func(t time.Time) tea.Msg {
		return hibernationTickMsg(t)
	})
}

// logTailTickCmd returns command for log tailing poll (Story 12.1, AC4: 2s interval).
func (m Model) logTailTickCmd() tea.Cmd {
	if !m.logTailActive {
		return nil
	}
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return logTailTickMsg(t)
	})
}

// pollLogEntriesCmd creates a command to read new entries from the log file.
func (m Model) pollLogEntriesCmd() tea.Cmd {
	sessionPath := m.currentSessionPath
	lastOffset := m.logLastOffset
	return func() tea.Msg {
		if sessionPath == "" {
			return logNewEntriesMsg{err: fmt.Errorf("no session path")}
		}

		file, err := os.Open(sessionPath)
		if err != nil {
			return logNewEntriesMsg{err: err}
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			return logNewEntriesMsg{err: err}
		}

		// No new data
		if info.Size() <= lastOffset {
			return logNewEntriesMsg{newOffset: lastOffset}
		}

		// Seek to last known position
		if _, err := file.Seek(lastOffset, 0); err != nil {
			return logNewEntriesMsg{err: err}
		}

		// Read new content
		newContent := make([]byte, info.Size()-lastOffset)
		_, err = file.Read(newContent)
		if err != nil {
			return logNewEntriesMsg{err: err}
		}

		return logNewEntriesMsg{
			newContent: string(newContent),
			newOffset:  info.Size(),
		}
	}
}

// loadHibernatedProjectsCmd loads hibernated projects from repository (Story 11.4).
func (m Model) loadHibernatedProjectsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		projects, err := m.repository.FindHibernated(ctx)
		return hibernatedProjectsLoadedMsg{projects: projects, err: err}
	}
}

// activateProjectCmd activates a hibernated project (Story 11.4).
func (m Model) activateProjectCmd(projectID, projectName string) tea.Cmd {
	if m.stateService == nil {
		return func() tea.Msg {
			return projectActivatedMsg{
				projectID: projectID,
				err:       errors.New("state service not available"),
			}
		}
	}
	return func() tea.Msg {
		ctx := context.Background()
		err := m.stateService.Activate(ctx, projectID)
		return projectActivatedMsg{projectID: projectID, projectName: projectName, err: err}
	}
}

// stateToggleCmd toggles project state via H key (Story 11.7).
// If hibernate=true, calls Hibernate(); otherwise calls Activate().
// Uses 5-second timeout to prevent TUI freeze on DB issues.
func (m Model) stateToggleCmd(projectID, projectName string, hibernate bool) tea.Cmd {
	if m.stateService == nil {
		return func() tea.Msg {
			return stateToggledMsg{
				projectID:   projectID,
				projectName: projectName,
				err:         errors.New("state service not available"),
			}
		}
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var err error
		var action string
		if hibernate {
			err = m.stateService.Hibernate(ctx, projectID)
			action = "hibernated"
		} else {
			err = m.stateService.Activate(ctx, projectID)
			action = "activated"
		}
		return stateToggledMsg{projectID: projectID, projectName: projectName, action: action, err: err}
	}
}

// waitForNextFileEventCmd waits for the next file event from the stored channel (Story 4.6).
// Returns nil if channel is not set or context is cancelled.
func (m Model) waitForNextFileEventCmd() tea.Cmd {
	if m.eventCh == nil {
		return nil
	}
	return func() tea.Msg {
		select {
		case <-m.watchCtx.Done():
			return nil // Context cancelled, stop listening
		case event, ok := <-m.eventCh:
			if !ok {
				return fileWatcherErrorMsg{err: fmt.Errorf("watcher channel closed")}
			}
			return fileEventMsg{
				Path:      event.Path,
				Operation: event.Operation,
				Timestamp: event.Timestamp,
			}
		}
	}
}

// startRefresh initiates async refresh of all projects (Story 3.6).
// Story 11.2 AC3: Runs hibernation check before refresh.
func (m Model) startRefresh() (tea.Model, tea.Cmd) {
	m.isRefreshing = true
	m.refreshTotal = len(m.projects)
	m.refreshProgress = 0
	m.refreshError = ""
	m.statusBar.SetRefreshing(true, 0, m.refreshTotal)

	// Story 11.2: Run hibernation check before refresh (AC3)
	// The refreshProjectsCmd will run after hibernation completes or immediately if service is nil
	if m.hibernationService != nil {
		return m, tea.Batch(m.checkAutoHibernationCmd(), m.refreshProjectsCmd())
	}
	return m, m.refreshProjectsCmd()
}

// refreshProjectsCmd creates a command that rescans all projects (Story 3.6).
func (m Model) refreshProjectsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		var refreshedCount, failedCount int

		for _, project := range m.projects {
			select {
			case <-ctx.Done():
				return refreshCompleteMsg{refreshedCount, failedCount, ctx.Err()}
			default:
			}

			// Run detection with coexistence awareness (Story 14.5)
			winner, allResults, err := m.detectionService.DetectWithCoexistenceSelection(ctx, project.Path)
			if err != nil {
				slog.Debug("refresh detection failed", "project", project.Name, "error", err)
				failedCount++
				continue
			}

			// Reload project from DB to get current state (prevents overwriting CLI changes)
			currentProject, findErr := m.repository.FindByID(ctx, project.ID)
			if findErr != nil {
				slog.Debug("refresh find failed", "project", project.Name, "error", findErr)
				failedCount++
				continue
			}

			// Determine primary result and populate coexistence fields (Story 14.5)
			var primary *domain.DetectionResult
			if winner != nil {
				primary = winner
				// Clear any previous coexistence warning
				currentProject.CoexistenceWarning = false
				currentProject.CoexistenceMessage = ""
				currentProject.SecondaryMethod = ""
				currentProject.SecondaryStage = domain.StageUnknown
			} else if len(allResults) > 0 {
				// Tie case - use first as primary (already sorted by most recent timestamp)
				primary = allResults[0]
				currentProject.CoexistenceWarning = primary.HasCoexistenceWarning()
				currentProject.CoexistenceMessage = primary.CoexistenceMessage
				if len(allResults) > 1 {
					currentProject.SecondaryMethod = allResults[1].Method
					currentProject.SecondaryStage = allResults[1].Stage
				}
			} else {
				// No methodology detected - use unknown
				unknownResult := domain.NewDetectionResult("unknown", domain.StageUnknown, domain.ConfidenceUncertain, "No methodology detected")
				primary = &unknownResult
				currentProject.CoexistenceWarning = false
				currentProject.CoexistenceMessage = ""
				currentProject.SecondaryMethod = ""
				currentProject.SecondaryStage = domain.StageUnknown
			}

			// Story 16.2: Record stage detection for metrics (recorder internally tracks previous stage)
			if m.metricsRecorder != nil {
				m.metricsRecorder.OnDetection(ctx, currentProject.Path, primary.Stage)
			}

			// Update ONLY detection fields, preserve state/hibernation/favorites
			currentProject.DetectedMethod = primary.Method
			currentProject.CurrentStage = primary.Stage
			currentProject.Confidence = primary.Confidence
			currentProject.DetectionReasoning = primary.Reasoning
			currentProject.UpdatedAt = time.Now()

			if err := m.repository.Save(ctx, currentProject); err != nil {
				slog.Debug("refresh save failed", "project", project.Name, "error", err)
				failedCount++
				continue
			}

			// Update in-memory project with current DB state
			*project = *currentProject

			refreshedCount++
		}

		var resultErr error
		if refreshedCount == 0 && failedCount > 0 {
			resultErr = fmt.Errorf("all projects failed to refresh")
		}

		return refreshCompleteMsg{refreshedCount, failedCount, resultErr}
	}
}

// Update implements tea.Model. Handles messages and returns updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Route to note editing handler when in note editing mode (Story 3.7)
		if m.isEditingNote {
			return m.handleNoteEditingKeyMsg(msg)
		}
		// Route to remove confirmation handler when in confirmation mode (Story 3.9)
		if m.isConfirmingRemove {
			return m.handleRemoveConfirmationKeyMsg(msg)
		}
		// Route to validation handler when in validation mode
		if m.viewMode == viewModeValidation {
			return m.handleValidationKeyMsg(msg)
		}
		// Story 12.1: Route to text view handler when in text view mode
		if m.viewMode == viewModeTextView {
			return m.handleTextViewKeyMsg(msg)
		}
		// Story 16.3: Route to stats view handler when in stats view mode
		if m.viewMode == viewModeStats {
			return m.handleStatsViewKeyMsg(msg)
		}
		// Story 12.1: Route to session picker handler when showing
		if m.showSessionPicker {
			return m.handleSessionPickerKeyMsg(msg)
		}
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		// Debounce resize events (50ms per UX spec)
		m.hasPendingResize = true
		m.pendingWidth = msg.Width
		m.pendingHeight = msg.Height
		return m, tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
			return resizeTickMsg{}
		})

	case resizeTickMsg:
		if m.hasPendingResize {
			wasReady := m.ready
			m.width = m.pendingWidth
			m.height = m.pendingHeight
			m.ready = true
			m.hasPendingResize = false

			// Set initial detail panel visibility only on first ready (AC4, AC5, AC6)
			if !wasReady {
				m.showDetailPanel = shouldShowDetailPanelByDefault(m.height)
				m.detailPanel.SetVisible(m.showDetailPanel)
			}

			// FIRST: Set condensed mode (affects status bar height) - Story 3.10 AC5
			m.statusBar.SetCondensed(m.height < MinHeight)

			// Calculate effective width for components (cap at maxContentWidth - Story 3.10 AC4, 8.10)
			effectiveWidth := m.width
			if m.isWideWidth() {
				effectiveWidth = m.maxContentWidth
			}

			// Update status bar width with effective width (Story 3.4, 3.10)
			m.statusBar.SetWidth(effectiveWidth)

			// Calculate content height using helper (Story 3.10 AC5)
			contentHeight := m.height - statusBarHeight(m.height)

			// Story 8.4: Process pending projects now that we have valid dimensions
			// Race condition fix: ProjectsLoadedMsg may arrive before WindowSizeMsg,
			// causing components to be created with width=0. Pending projects are
			// processed here after m.ready=true ensures correct dimensions.
			if m.pendingProjects != nil {
				m.projects = m.pendingProjects
				m.pendingProjects = nil

				// Create components with correct dimensions
				m.projectList = components.NewProjectListModel(m.projects, effectiveWidth, contentHeight)
				m.projectList.SetDelegateWaitingCallbacks(m.isProjectWaiting, m.getWaitingDuration)

				m.detailPanel = components.NewDetailPanelModel(effectiveWidth, contentHeight)
				m.detailPanel.SetProject(m.projectList.SelectedProject())
				m.detailPanel.SetVisible(m.showDetailPanel)
				m.detailPanel.SetWaitingCallbacks(m.isProjectWaiting, m.getWaitingDuration)
				m.detailPanel.SetAgentStateCallback(m.getAgentState) // Story 15.7

				// Update status bar counts
				active, hibernated, waiting := components.CalculateCountsWithWaiting(m.projects, m.isProjectWaiting)
				m.statusBar.SetCounts(active, hibernated, waiting)

				// Code review fix M3: Set height hint for pendingProjects block (before early return)
				if m.isHorizontalLayout() && m.showDetailPanel && contentHeight < HorizontalDetailThreshold {
					m.statusBar.SetHeightHint("[d] Detail hidden - insufficient height")
				} else {
					m.statusBar.SetHeightHint("")
				}

				// Story 4.6: Start file watcher (code review H1: use helper)
				watcherCmd := m.startFileWatcherForProjects()

				// Story 8.11: Start periodic stage refresh timer
				stageCmd := m.stageRefreshTickCmd()

				// Story 11.2: Start hourly hibernation timer (AC3)
				hibernationCmd := m.hibernationTickCmd()

				if watcherCmd != nil || stageCmd != nil || hibernationCmd != nil {
					return m, tea.Batch(watcherCmd, stageCmd, hibernationCmd)
				}
			}

			// Story 8.4: Update existing component dimensions (use zero-value check instead of len guard)
			if m.projectList.Width() > 0 {
				m.projectList.SetSize(effectiveWidth, contentHeight)
				m.detailPanel.SetSize(effectiveWidth, contentHeight)
			}

			// Story 8.12: Set height hint when detail is auto-hidden in horizontal mode (AC1)
			if m.isHorizontalLayout() && m.showDetailPanel && contentHeight < HorizontalDetailThreshold {
				m.statusBar.SetHeightHint("[d] Detail hidden - insufficient height")
			} else {
				m.statusBar.SetHeightHint("")
			}
		}
		return m, nil

	case validationCompleteMsg:
		// Handle validation result from Init()
		if len(msg.invalidProjects) > 0 {
			m.viewMode = viewModeValidation
			m.invalidProjects = msg.invalidProjects
			m.currentInvalidIdx = 0
			return m, nil
		}
		// Story 7.4 AC6: Set loading state before loading projects
		m.isLoading = true
		m.statusBar.SetLoading(true)
		// No invalid projects, load projects
		return m, m.loadProjectsCmd()

	case validationErrorMsg:
		// Log error and continue to normal view (non-fatal)
		slog.Error("Path validation failed", "error", msg.err)
		return m, nil

	case deleteProjectMsg:
		// Handle async delete completion
		if msg.err != nil {
			m.validationError = "Failed to delete: " + msg.err.Error()
			return m, nil
		}
		m.validationError = "" // Clear error on success
		m, cmd := m.advanceToNextInvalid()
		return m, cmd

	case moveProjectMsg:
		// Handle async move completion
		if msg.err != nil {
			m.validationError = "Failed to move: " + msg.err.Error()
			return m, nil
		}
		m.validationError = "" // Clear error on success
		m, cmd := m.advanceToNextInvalid()
		return m, cmd

	case keepProjectMsg:
		// Handle async keep completion
		if msg.err != nil {
			m.validationError = "Failed to keep: " + msg.err.Error()
			return m, nil
		}
		m.validationError = "" // Clear error on success
		m, cmd := m.advanceToNextInvalid()
		return m, cmd

	case ProjectsLoadedMsg:
		// Story 7.4 AC6: Clear loading state
		m.isLoading = false
		m.statusBar.SetLoading(false)

		// Handle project loading result
		if msg.err != nil {
			slog.Error("Failed to load projects", "error", msg.err)
			m.projects = nil
			return m, nil
		}

		// Story 8.4: Defer component creation until ready (race condition fix)
		// If m.ready is false, WindowSizeMsg hasn't been processed yet, so m.width may be 0.
		// Race condition: ProjectsLoadedMsg may arrive before WindowSizeMsg.
		if !m.ready {
			m.pendingProjects = msg.projects
			return m, nil
		}

		m.projects = msg.projects
		if len(m.projects) > 0 {
			// Story 8.4, 8.10: Use effectiveWidth instead of raw m.width to cap at maxContentWidth
			effectiveWidth := m.width
			if m.isWideWidth() {
				effectiveWidth = m.maxContentWidth
			}
			contentHeight := m.height - statusBarHeight(m.height)

			// Hotfix: Preserve selection index across refresh
			prevIndex := m.projectList.Index()

			m.projectList = components.NewProjectListModel(m.projects, effectiveWidth, contentHeight)

			// Story 11.4: Select just-activated project (AC3)
			if m.justActivatedProjectID != "" {
				for i, p := range m.projects {
					if p.ID == m.justActivatedProjectID {
						m.projectList.SelectByIndex(i)
						break
					}
				}
				m.justActivatedProjectID = "" // Clear after use
			} else if prevIndex >= 0 && prevIndex < len(m.projects) {
				// Restore selection (clamp to valid range)
				m.projectList.SelectByIndex(prevIndex)
			}

			// Story 4.5: Wire waiting callbacks to project list delegate
			m.projectList.SetDelegateWaitingCallbacks(m.isProjectWaiting, m.getWaitingDuration)

			// Initialize detail panel with selected project
			m.detailPanel = components.NewDetailPanelModel(effectiveWidth, contentHeight)
			m.detailPanel.SetProject(m.projectList.SelectedProject())
			m.detailPanel.SetVisible(m.showDetailPanel)

			// Story 4.5: Wire waiting callbacks to detail panel
			m.detailPanel.SetWaitingCallbacks(m.isProjectWaiting, m.getWaitingDuration)
			m.detailPanel.SetAgentStateCallback(m.getAgentState) // Story 15.7

			// Update status bar counts (Story 3.4, 4.5)
			active, hibernated, waiting := components.CalculateCountsWithWaiting(m.projects, m.isProjectWaiting)
			m.statusBar.SetCounts(active, hibernated, waiting)

			// Story 4.6: Start file watcher (code review H1: use helper)
			watcherCmd := m.startFileWatcherForProjects()

			// Story 8.11: Start periodic stage refresh timer
			stageCmd := m.stageRefreshTickCmd()

			// Story 11.2: Start hourly hibernation timer (AC3)
			hibernationCmd := m.hibernationTickCmd()

			if watcherCmd != nil || stageCmd != nil || hibernationCmd != nil {
				return m, tea.Batch(watcherCmd, stageCmd, hibernationCmd)
			}
		}
		return m, nil

	case refreshCompleteMsg:
		// Handle refresh completion (Story 3.6)
		m.isRefreshing = false
		m.statusBar.SetRefreshing(false, 0, 0)
		if msg.err != nil {
			m.refreshError = msg.err.Error()
			m.statusBar.SetRefreshComplete("✗ Refresh failed")
			// Note: stageRefreshTickCmd is NOT called here - managed by stageRefreshTickMsg handler only
			return m, nil
		}
		m.refreshError = ""
		// Story 7.4 AC5: Show failure count if any
		var resultMsg string
		if msg.failedCount > 0 {
			resultMsg = fmt.Sprintf("✓ Scanned %d projects (%d failed)", msg.refreshedCount, msg.failedCount)
		} else {
			resultMsg = fmt.Sprintf("✓ Scanned %d projects", msg.refreshedCount)
		}
		m.statusBar.SetRefreshComplete(resultMsg)

		// Story 7.1 AC3: Attempt watcher recovery if it was previously unavailable
		if !m.fileWatcherAvailable && m.fileWatcher != nil && len(m.projects) > 0 {
			paths := make([]string, len(m.projects))
			for i, p := range m.projects {
				paths[i] = p.Path
			}

			// Cancel old context if any
			if m.watchCancel != nil {
				m.watchCancel()
			}

			m.watchCtx, m.watchCancel = context.WithCancel(context.Background())
			m.lastWatcherRestart = time.Now() // Story 9.5-2: Grace period for recovery
			eventCh, err := m.fileWatcher.Watch(m.watchCtx, paths)
			if err == nil {
				m.fileWatcherAvailable = true
				m.eventCh = eventCh
				m.statusBar.SetWatcherWarning("") // Clear warning on successful recovery
				slog.Debug("file watcher recovered on refresh")

				// Check for partial failures
				failedPaths := m.fileWatcher.GetFailedPaths()
				if len(failedPaths) > 0 {
					// Partial recovery - still some failures
					return m, tea.Batch(
						m.loadProjectsCmd(),
						tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
							return clearRefreshMsgMsg{}
						}),
						m.waitForNextFileEventCmd(),
						func() tea.Msg {
							return watcherWarningMsg{
								failedPaths: failedPaths,
								totalPaths:  len(paths),
							}
						},
					)
				}

				return m, tea.Batch(
					m.loadProjectsCmd(),
					tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
						return clearRefreshMsgMsg{}
					}),
					m.waitForNextFileEventCmd(),
				)
			}
			// Recovery failed - keep watcher as unavailable
			slog.Debug("file watcher recovery failed", "error", err)
		}

		// Reload projects and start timer to clear message
		// Note: stageRefreshTickCmd is NOT called here - it's managed by stageRefreshTickMsg handler only
		return m, tea.Batch(
			m.loadProjectsCmd(),
			tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return clearRefreshMsgMsg{}
			}),
		)

	case clearRefreshMsgMsg:
		// Clear refresh completion message after 3 seconds (Story 3.6)
		m.statusBar.SetRefreshComplete("")
		return m, nil

	case noteSavedMsg:
		// Update local project state (Story 3.7)
		for _, p := range m.projects {
			if p.ID == msg.projectID {
				p.Notes = msg.newNote
				break
			}
		}
		// Update detail panel
		m.detailPanel.SetProject(m.projectList.SelectedProject())
		// Set feedback message and display in status bar (AC2: feedback shows "✓ Note saved")
		m.noteFeedback = "✓ Note saved"
		m.statusBar.SetRefreshComplete(m.noteFeedback)
		// Clear after 3 seconds
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNoteFeedbackMsg{}
		})

	case noteSaveErrorMsg:
		// Handle note save failure (Story 3.7)
		m.noteFeedback = "✗ Failed to save note"
		m.statusBar.SetRefreshComplete(m.noteFeedback)
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNoteFeedbackMsg{}
		})

	case clearNoteFeedbackMsg:
		// Clear note feedback message (Story 3.7)
		m.noteFeedback = ""
		m.statusBar.SetRefreshComplete("")
		return m, nil

	case favoriteSavedMsg:
		// Story 8.5: Capture selected project ID BEFORE re-sort
		selectedID := ""
		if selected := m.projectList.SelectedProject(); selected != nil {
			selectedID = selected.ID
		}

		// Update local project state (Story 3.8)
		for _, p := range m.projects {
			if p.ID == msg.projectID {
				p.IsFavorite = msg.isFavorite
				break
			}
		}

		// Story 8.5: Re-sort list (triggers SortByName via SetProjects)
		m.projectList.SetProjects(m.projects)

		// Story 8.5: Restore selection by ID (project may have moved position)
		if selectedID != "" {
			found := false
			for i, p := range m.projectList.Projects() {
				if p.ID == selectedID {
					m.projectList.SelectByIndex(i)
					found = true
					break
				}
			}
			// Edge case: If project was removed from list, select first item
			if !found && m.projectList.Len() > 0 {
				m.projectList.SelectByIndex(0)
			}
		}

		// Update detail panel with (possibly moved) selection
		m.detailPanel.SetProject(m.projectList.SelectedProject())
		// Set feedback message (Story 8.9: emoji fallback, code review M3: use emoji.EmptyStar)
		var feedback string
		if msg.isFavorite {
			feedback = emoji.Star() + " Favorited"
		} else {
			feedback = emoji.EmptyStar() + " Unfavorited"
		}
		m.statusBar.SetRefreshComplete(feedback)
		// Clear after 3 seconds
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearFavoriteFeedbackMsg{}
		})

	case favoriteSaveErrorMsg:
		m.statusBar.SetRefreshComplete("✗ Failed to toggle favorite")
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearFavoriteFeedbackMsg{}
		})

	case clearFavoriteFeedbackMsg:
		m.statusBar.SetRefreshComplete("")
		return m, nil

	case removeConfirmedMsg:
		// Handle async delete completion (Story 3.9, 11.4 AC5)
		if msg.err != nil {
			m.statusBar.SetRefreshComplete("✗ Failed to remove: " + msg.projectName)
			return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return clearRemoveFeedbackMsg{}
			})
		}

		// Story 11.4: Stay in hibernated view and reload if we were viewing hibernated
		if m.viewMode == viewModeHibernated {
			m.statusBar.SetRefreshComplete("✓ Removed: " + msg.projectName)
			return m, tea.Batch(
				m.loadHibernatedProjectsCmd(),
				tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
					return clearRemoveFeedbackMsg{}
				}),
			)
		}

		// Remove from local projects slice
		var newProjects []*domain.Project
		for _, p := range m.projects {
			if p.ID != msg.projectID {
				newProjects = append(newProjects, p)
			}
		}
		m.projects = newProjects
		// Update project list component
		m.projectList.SetProjects(newProjects)
		// Update detail panel with new selection
		m.detailPanel.SetProject(m.projectList.SelectedProject())
		// Update status bar counts (Story 4.5)
		active, hibernated, waiting := components.CalculateCountsWithWaiting(m.projects, m.isProjectWaiting)
		m.statusBar.SetCounts(active, hibernated, waiting)
		// Show success feedback
		m.statusBar.SetRefreshComplete("✓ Removed: " + msg.projectName)
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearRemoveFeedbackMsg{}
		})

	case removeConfirmTimeoutMsg:
		// Auto-cancel confirmation after 30 seconds (Story 3.9, AC4)
		if m.isConfirmingRemove {
			m.isConfirmingRemove = false
			m.confirmTarget = nil
		}
		return m, nil

	case clearRemoveFeedbackMsg:
		m.statusBar.SetRefreshComplete("")
		return m, nil

	case tickMsg:
		// Periodic tick to trigger timestamp recalculation (Story 4.2, AC4).
		// The View() method calls FormatRelativeTime which recalculates on each render.

		// Epic 4 Hotfix H5: Recalculate waiting counts on each tick.
		// Without this, status bar would show stale count even as projects transition to WAITING.
		if len(m.projects) > 0 {
			active, hibernated, waiting := components.CalculateCountsWithWaiting(m.projects, m.isProjectWaiting)
			m.statusBar.SetCounts(active, hibernated, waiting)
		}

		return m, tickCmd()

	case stageRefreshTickMsg:
		// Story 8.11: Periodic stage re-detection
		// Skip if disabled, already refreshing, or no projects - but always reschedule
		if m.stageRefreshInterval == 0 || m.isRefreshing || len(m.projects) == 0 {
			return m, m.rescheduleStageTimer()
		}
		// Start refresh and reschedule timer
		model, cmd := m.startRefresh()
		return model, tea.Batch(cmd, m.rescheduleStageTimer())

	case fileEventMsg:
		// Story 4.6: Handle file system event
		m.handleFileEvent(msg)
		// Re-subscribe to wait for next event
		return m, m.waitForNextFileEventCmd()

	case fileWatcherErrorMsg:
		// Story 9.5-2: Ignore transient errors within 500ms of watcher restart
		// Zero-value check: !m.lastWatcherRestart.IsZero() ensures app-startup errors are handled
		if !m.lastWatcherRestart.IsZero() && time.Since(m.lastWatcherRestart) < 500*time.Millisecond {
			slog.Debug("ignoring transient watcher error", "error", msg.err, "elapsed_ms", time.Since(m.lastWatcherRestart).Milliseconds())
			return m, nil
		}

		// Story 4.6: Handle genuine file watcher error (AC3, Story 8.9: emoji fallback)
		slog.Warn("file watcher error", "error", msg.err)
		m.fileWatcherAvailable = false
		m.eventCh = nil // Clear to allow recovery on next refresh
		m.statusBar.SetWatcherWarning(emoji.Warning() + " File watching unavailable")
		return m, nil

	case watcherWarningMsg:
		// Story 7.1: Handle partial file watcher failures (AC1, AC2, Story 8.9: emoji fallback)
		if len(msg.failedPaths) > 0 && len(msg.failedPaths) < msg.totalPaths {
			// Partial failure - show first failed project name (AC1)
			// Code review fix M1: Show count if multiple failures
			warningText := fmt.Sprintf("%s File watching unavailable for: %s", emoji.Warning(), filepath.Base(msg.failedPaths[0]))
			if len(msg.failedPaths) > 1 {
				warningText += fmt.Sprintf(" (+%d more)", len(msg.failedPaths)-1)
			}
			m.statusBar.SetWatcherWarning(warningText)
		} else if len(msg.failedPaths) == msg.totalPaths {
			// Complete failure (AC2)
			m.statusBar.SetWatcherWarning(emoji.Warning() + " File watching unavailable. Use [r] to refresh.")
			m.fileWatcherAvailable = false
		}
		return m, nil

	case configWarningMsg:
		// Story 7.2: Handle config loading errors (AC6)
		m.configWarning = msg.warning
		m.configWarningTime = time.Now()
		m.statusBar.SetConfigWarning(msg.warning)
		// Auto-clear after 10 seconds
		return m, tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
			return clearConfigWarningMsg{}
		})

	case clearConfigWarningMsg:
		// Story 7.2: Clear config warning after 10 seconds (AC6)
		if time.Since(m.configWarningTime) >= 10*time.Second {
			m.configWarning = ""
			m.statusBar.SetConfigWarning("")
		}
		return m, nil

	case projectCorruptionMsg:
		// Story 7.3: Handle project database corruption (AC7, Story 8.9: emoji fallback)
		m.corruptedProjects = msg.projects
		if len(msg.projects) > 0 {
			var warning string
			if len(msg.projects) == 1 {
				warning = fmt.Sprintf("%s %s: corrupted (vdash reset %s)", emoji.Warning(), msg.projects[0], msg.projects[0])
			} else {
				warning = fmt.Sprintf("%s %d projects corrupted (vdash reset --all)", emoji.Warning(), len(msg.projects))
			}
			m.statusBar.SetWatcherWarning(warning)
		}
		return m, nil

	case hibernationCompleteMsg:
		// Story 11.2: Handle auto-hibernation check completion (AC4: silent transition)
		if msg.err != nil {
			slog.Warn("auto-hibernation check failed", "error", msg.err)
		} else if msg.count > 0 {
			slog.Debug("auto-hibernated projects", "count", msg.count)
		}
		// Reload projects to update counts (silent - AC4)
		return m, m.loadProjectsCmd()

	case hibernationTickMsg:
		// Story 11.2: Hourly auto-hibernation check (AC3)
		if m.hibernationService == nil {
			return m, m.rescheduleHibernationTimer()
		}
		return m, tea.Batch(
			m.checkAutoHibernationCmd(),
			m.rescheduleHibernationTimer(),
		)

	case hibernatedProjectsLoadedMsg:
		// Story 11.4: Handle hibernated projects loaded (AC2)
		if msg.err != nil {
			slog.Error("failed to load hibernated projects", "error", msg.err)
			return m, nil
		}
		m.hibernatedProjects = msg.projects
		// Sort by LastActivityAt descending (most recent first)
		sort.Slice(m.hibernatedProjects, func(i, j int) bool {
			return m.hibernatedProjects[i].LastActivityAt.After(m.hibernatedProjects[j].LastActivityAt)
		})
		// Create list component - uses same component as active list
		effectiveWidth := m.width
		if m.isWideWidth() {
			effectiveWidth = m.maxContentWidth
		}
		contentHeight := m.height - statusBarHeight(m.height)
		m.hibernatedList = components.NewProjectListModel(m.hibernatedProjects, effectiveWidth, contentHeight)
		// Update status bar with hibernated count for AC7
		m.statusBar.SetHibernatedViewCount(len(m.hibernatedProjects))
		return m, nil

	case projectActivatedMsg:
		// Story 11.4: Handle project activation (AC3)
		// Handle race condition: project may have been auto-activated by Story 11.3
		if msg.err != nil {
			if errors.Is(msg.err, domain.ErrInvalidStateTransition) {
				// Project already active - reload hibernated list silently
				slog.Debug("project already active, reloading hibernated list", "project_id", msg.projectID)
				return m, m.loadHibernatedProjectsCmd()
			}
			slog.Warn("failed to activate project", "error", msg.err)
			return m, nil
		}
		slog.Debug("project activated", "project_id", msg.projectID, "project_name", msg.projectName)
		// Track for post-load selection (AC3)
		m.justActivatedProjectID = msg.projectID
		// Switch to active view and reload
		m.viewMode = viewModeNormal
		m.statusBar.SetInHibernatedView(false)
		return m, m.loadProjectsCmd()

	case stateToggledMsg:
		// Story 11.7: Handle state toggle result (AC3, AC4, AC5)
		if msg.err != nil {
			if errors.Is(msg.err, domain.ErrFavoriteCannotHibernate) {
				// AC3: Favorite protection - show error feedback
				m.statusBar.SetRefreshComplete("Cannot hibernate favorite project")
				return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
					return clearRemoveFeedbackMsg{} // Reuse existing clear msg
				})
			}
			if errors.Is(msg.err, domain.ErrInvalidStateTransition) {
				// Idempotent case - reload appropriate list silently
				if msg.action == "hibernated" {
					return m, m.loadProjectsCmd()
				}
				return m, m.loadHibernatedProjectsCmd()
			}
			// General error
			slog.Warn("failed to toggle project state", "error", msg.err)
			m.statusBar.SetRefreshComplete("✗ State change failed")
			return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return clearRemoveFeedbackMsg{}
			})
		}
		// AC4: Success feedback
		var feedback string
		if msg.action == "hibernated" {
			feedback = fmt.Sprintf("✓ Hibernated: %s", msg.projectName)
		} else {
			feedback = fmt.Sprintf("✓ Activated: %s", msg.projectName)
		}
		m.statusBar.SetRefreshComplete(feedback)
		// AC5: Reload appropriate list
		var reloadCmd tea.Cmd
		if msg.action == "hibernated" {
			reloadCmd = m.loadProjectsCmd()
		} else {
			reloadCmd = m.loadHibernatedProjectsCmd()
		}
		return m, tea.Batch(
			reloadCmd,
			tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return clearRemoveFeedbackMsg{}
			}),
		)

	// Story 12.1: Log session picker message handlers
	case logSessionsLoadedMsg:
		if !m.showSessionPicker {
			return m, nil // Discard - picker was closed
		}
		if msg.err != nil {
			slog.Warn("failed to load log sessions", "error", msg.err)
			m.showSessionPicker = false
			return m, func() tea.Msg {
				return flashMsg{text: "Failed to load sessions"}
			}
		}
		m.logSessions = msg.sessions
		if len(msg.sessions) == 0 {
			m.showSessionPicker = false
			return m, func() tea.Msg {
				return flashMsg{text: "No sessions found"}
			}
		}
		return m, nil

	case textViewContentMsg:
		// jq output is ready - display in text view
		if msg.err != nil {
			slog.Warn("failed to format log content", "error", msg.err)
			return m, func() tea.Msg {
				return flashMsg{text: "Failed to format logs: " + msg.err.Error()}
			}
		}
		m.viewMode = viewModeTextView
		m.textViewTitle = msg.title
		m.textViewContent = strings.Split(msg.content, "\n")

		// AC4: Store session path and offset for live tailing
		m.currentSessionPath = msg.sessionPath
		m.logLastOffset = msg.fileSize

		// AC3: Auto-scroll to latest by default
		contentHeight := m.height - statusBarHeight(m.height) - 2
		maxScroll := len(m.textViewContent) - contentHeight
		if maxScroll < 0 {
			maxScroll = 0
		}
		m.textViewScroll = maxScroll // Start at bottom
		m.logAutoScroll = true       // Enable auto-scroll

		// AC4: Start live tailing (2s interval)
		m.logTailActive = true
		return m, m.logTailTickCmd()

	case logTailTickMsg:
		// AC4: Poll for new log entries
		if !m.logTailActive || m.viewMode != viewModeTextView || m.currentSessionPath == "" {
			return m, nil
		}
		return m, tea.Batch(
			m.pollLogEntriesCmd(),
			m.logTailTickCmd(), // Reschedule next tick
		)

	case logNewEntriesMsg:
		// New entries found - append to content
		if msg.err != nil {
			slog.Debug("log poll failed", "error", msg.err)
			return m, nil
		}

		// Update offset for next poll
		m.logLastOffset = msg.newOffset

		if msg.newContent == "" {
			return m, nil // No new content
		}

		// Append new lines
		newLines := strings.Split(msg.newContent, "\n")
		m.textViewContent = append(m.textViewContent, newLines...)

		// AC3: Auto-scroll to latest if enabled
		if m.logAutoScroll {
			contentHeight := m.height - statusBarHeight(m.height) - 2
			maxScroll := len(m.textViewContent) - contentHeight
			if maxScroll < 0 {
				maxScroll = 0
			}
			m.textViewScroll = maxScroll
		}
		return m, nil

	case flashMsg:
		// AC8: Show flash message for no-logs case
		m.flashMessage = msg.text
		m.flashMessageTime = time.Now()
		m.statusBar.SetRefreshComplete(msg.text)
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg { // AC8: 2 seconds
			return clearFlashMsg{}
		})

	case clearFlashMsg:
		m.flashMessage = ""
		m.statusBar.SetRefreshComplete("")
		return m, nil
	}

	return m, nil
}

// advanceToNextInvalid moves to the next invalid project or switches to normal view.
// Returns the model and a command to load projects if validation is complete.
func (m Model) advanceToNextInvalid() (Model, tea.Cmd) {
	m.currentInvalidIdx++
	if m.currentInvalidIdx >= len(m.invalidProjects) {
		m.viewMode = viewModeNormal
		// Load projects after validation is complete
		return m, m.loadProjectsCmd()
	}
	return m, nil
}

// handleValidationKeyMsg processes keyboard input in validation mode.
func (m Model) handleValidationKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Allow quit even in validation mode
	switch msg.String() {
	case KeyQuit, KeyForceQuit:
		// Story 4.6: Clean up file watcher on quit (AC10)
		if m.watchCancel != nil {
			m.watchCancel()
		}
		if m.fileWatcher != nil {
			m.fileWatcher.Close()
		}
		return m, tea.Quit
	}

	if m.currentInvalidIdx >= len(m.invalidProjects) {
		// All handled, switch to normal view
		m.viewMode = viewModeNormal
		return m, nil
	}

	currentProject := m.invalidProjects[m.currentInvalidIdx].Project

	switch strings.ToLower(msg.String()) {
	case "d":
		return m, m.deleteProjectCmd(currentProject.ID)
	case "m":
		return m, m.moveProjectCmd(currentProject)
	case "k":
		return m, m.keepProjectCmd(currentProject)
	}

	return m, nil
}

// deleteProjectCmd creates a command to delete a project (AC2).
func (m Model) deleteProjectCmd(projectID string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := m.repository.Delete(ctx, projectID)
		return deleteProjectMsg{projectID: projectID, err: err}
	}
}

// moveProjectCmd creates a command to move a project to current directory (AC3).
func (m Model) moveProjectCmd(project *domain.Project) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return moveProjectMsg{projectID: project.ID, err: err}
		}

		// Validate new path exists
		canonicalPath, err := filesystem.CanonicalPath(cwd)
		if err != nil {
			return moveProjectMsg{projectID: project.ID, err: err}
		}

		// Delete old project entry (ID is path-based, so it will change)
		// Ignore delete errors - old entry may not exist
		_ = m.repository.Delete(ctx, project.ID)

		// Update project with new path
		project.Path = canonicalPath
		project.ID = domain.GenerateID(canonicalPath) // ID is path-based
		project.PathMissing = false
		project.UpdatedAt = time.Now()

		// NOTE: detectionService is optional until Epic 2 detection is wired.
		// The Move action works without re-detection.

		// Save to repository
		if err := m.repository.Save(ctx, project); err != nil {
			return moveProjectMsg{projectID: project.ID, err: err}
		}

		return moveProjectMsg{projectID: project.ID, newPath: canonicalPath}
	}
}

// keepProjectCmd creates a command to keep a project with PathMissing flag (AC4).
func (m Model) keepProjectCmd(project *domain.Project) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Set PathMissing flag
		project.PathMissing = true
		project.UpdatedAt = time.Now()

		// Save to repository
		if err := m.repository.Save(ctx, project); err != nil {
			return keepProjectMsg{projectID: project.ID, err: err}
		}

		return keepProjectMsg{projectID: project.ID}
	}
}

// handleKeyMsg processes keyboard input.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If help is showing, any key closes it (except '?' which toggles)
	if m.showHelp {
		if msg.String() == KeyHelp {
			m.showHelp = false
			return m, nil
		}
		// Any other key closes help
		m.showHelp = false
		return m, nil
	}

	switch msg.String() {
	case KeyQuit:
		// Story 4.6: Clean up file watcher on quit (AC10)
		if m.watchCancel != nil {
			m.watchCancel()
		}
		if m.fileWatcher != nil {
			m.fileWatcher.Close()
		}
		return m, tea.Quit
	case KeyForceQuit:
		// Story 4.6: Clean up file watcher on force quit (AC10)
		if m.watchCancel != nil {
			m.watchCancel()
		}
		if m.fileWatcher != nil {
			m.fileWatcher.Close()
		}
		return m, tea.Quit
	case KeyHelp:
		m.showHelp = !m.showHelp
		return m, nil
	case KeyEscape:
		// Story 11.4: Return from hibernated view (AC4)
		if m.viewMode == viewModeHibernated {
			m.viewMode = viewModeNormal
			m.statusBar.SetInHibernatedView(false)
			// Restore active selection
			if m.activeSelectedIdx >= 0 && m.activeSelectedIdx < len(m.projects) {
				m.projectList.SelectByIndex(m.activeSelectedIdx)
			}
			return m, nil
		}
		// No-op in normal mode
		return m, nil
	case KeyHibernated:
		// Story 11.4: Toggle hibernated view (AC1, AC4)
		if m.viewMode == viewModeHibernated {
			// Return to active view
			m.viewMode = viewModeNormal
			m.statusBar.SetInHibernatedView(false)
			// Restore active selection
			if m.activeSelectedIdx >= 0 && m.activeSelectedIdx < len(m.projects) {
				m.projectList.SelectByIndex(m.activeSelectedIdx)
			}
			return m, nil
		}
		// Enter hibernated view
		m.activeSelectedIdx = m.projectList.Index()
		m.viewMode = viewModeHibernated
		m.statusBar.SetInHibernatedView(true)
		return m, m.loadHibernatedProjectsCmd()
	case KeyShiftEnter: // Shift+Enter
		// Story 12.1: Show session picker before opening logs
		if m.viewMode == viewModeNormal && len(m.projects) > 0 {
			return m.handleShiftEnterForSessionPicker()
		}
		return m, nil
	case "enter":
		// Story 11.4: Wake hibernated project (AC3)
		if m.viewMode == viewModeHibernated && len(m.hibernatedProjects) > 0 {
			selected := m.hibernatedList.SelectedProject()
			if selected != nil {
				return m, m.activateProjectCmd(selected.ID, project.EffectiveName(selected))
			}
			return m, nil
		}
		// Story 12.1: Open most recent log session with jq
		if m.viewMode == viewModeNormal {
			return m.handleEnterForLogs()
		}
		return m, nil
	case KeyDetail:
		// Story 11.4: Detail panel works in hibernated view too (AC9)
		m.showDetailPanel = !m.showDetailPanel
		m.detailPanel.SetVisible(m.showDetailPanel)
		// Update detail panel with current selection in hibernated view
		if m.viewMode == viewModeHibernated {
			m.detailPanel.SetProject(m.hibernatedList.SelectedProject())
		}
		return m, nil
	case KeyRefresh:
		if m.isRefreshing {
			return m, nil // Ignore if already refreshing
		}
		if m.detectionService == nil {
			// No detection service - show message and return
			m.refreshError = "Detection service not available"
			return m, nil
		}
		return m.startRefresh()
	case KeyNotes:
		// Story 3.7: Open note editor for selected project
		if m.isEditingNote {
			return m, nil // Ignore if already editing
		}
		if len(m.projects) == 0 {
			return m, nil // No project to edit
		}
		return m.startNoteEditing()
	case KeyFavorite:
		// Story 3.8: Toggle favorite status for selected project
		if len(m.projects) == 0 {
			return m, nil // No project to favorite
		}
		return m.toggleFavorite()
	case KeyRemove:
		// Story 3.9, 11.4 (AC5): Start remove confirmation for selected project
		if m.isConfirmingRemove {
			return m, nil // Already confirming
		}
		// Story 11.4: Allow removal in hibernated view
		if m.viewMode == viewModeHibernated {
			if len(m.hibernatedProjects) == 0 {
				return m, nil // No project to remove
			}
			return m.startRemoveConfirmation()
		}
		if len(m.projects) == 0 {
			return m, nil // No project to remove
		}
		return m.startRemoveConfirmation()

	case KeyStateToggle:
		// Story 11.7: Toggle project state with H key (AC1, AC2, AC7)
		if m.viewMode == viewModeHibernated {
			// In hibernated view: activate selected project (AC2)
			if len(m.hibernatedProjects) == 0 {
				return m, nil // AC7: No-op when empty
			}
			selected := m.hibernatedList.SelectedProject()
			if selected == nil {
				return m, nil
			}
			return m, m.stateToggleCmd(selected.ID, project.EffectiveName(selected), false)
		}
		// In active view: hibernate selected project (AC1)
		if len(m.projects) == 0 {
			return m, nil // AC7: No-op when empty
		}
		selected := m.projectList.SelectedProject()
		if selected == nil {
			return m, nil
		}
		return m, m.stateToggleCmd(selected.ID, project.EffectiveName(selected), true)

	case KeyLogOpenView, "L":
		// Story 12.2 AC1: 'L' key opens session picker from project list (case-insensitive)
		if m.viewMode == viewModeNormal && len(m.projects) > 0 {
			return m.handleShiftEnterForSessionPicker()
		}
		return m, nil

	case KeyStats:
		// Story 16.3: Open Stats View (AC1)
		if m.showHelp || m.isEditingNote || m.isConfirmingRemove {
			return m, nil
		}
		m.enterStatsView()
		return m, nil
	}

	// Story 11.4: Forward key messages to hibernated list when in hibernated view (AC8)
	if m.viewMode == viewModeHibernated && len(m.hibernatedProjects) > 0 {
		var cmd tea.Cmd
		m.hibernatedList, cmd = m.hibernatedList.Update(msg)
		// Update detail panel with current selection (AC9)
		m.detailPanel.SetProject(m.hibernatedList.SelectedProject())
		return m, cmd
	}

	// Forward key messages to project list when in normal mode
	if len(m.projects) > 0 {
		var cmd tea.Cmd
		m.projectList, cmd = m.projectList.Update(msg)
		// Update detail panel with current selection (may have changed)
		m.detailPanel.SetProject(m.projectList.SelectedProject())
		return m, cmd
	}

	return m, nil
}

// startNoteEditing opens the note editor for the selected project (Story 3.7).
func (m Model) startNoteEditing() (tea.Model, tea.Cmd) {
	selected := m.projectList.SelectedProject()
	if selected == nil {
		return m, nil
	}

	m.isEditingNote = true
	m.noteEditTarget = selected
	m.originalNote = selected.Notes

	// Initialize text input with current note
	ti := textinput.New()
	ti.Placeholder = "Enter note..."
	ti.Focus()
	ti.CharLimit = 500 // Reasonable limit for notes
	ti.Width = m.width - 10
	ti.SetValue(selected.Notes)
	m.noteInput = ti

	return m, textinput.Blink
}

// handleNoteEditingKeyMsg processes keyboard input during note editing (Story 3.7).
func (m Model) handleNoteEditingKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		// Save note
		return m.saveNote()
	case tea.KeyEsc:
		// Cancel editing, restore original
		m.isEditingNote = false
		m.noteInput = textinput.Model{} // Clear
		return m, nil
	}

	// Forward to text input
	var cmd tea.Cmd
	m.noteInput, cmd = m.noteInput.Update(msg)
	return m, cmd
}

// saveNote saves the note to the repository (Story 3.7).
func (m Model) saveNote() (tea.Model, tea.Cmd) {
	newNote := strings.TrimSpace(m.noteInput.Value())
	m.isEditingNote = false

	return m, m.saveNoteCmd(m.noteEditTarget.ID, newNote)
}

// saveNoteCmd creates a command that saves the note to repository (Story 3.7).
func (m Model) saveNoteCmd(projectID, note string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Find project
		project, err := m.repository.FindByID(ctx, projectID)
		if err != nil {
			return noteSaveErrorMsg{err: err}
		}

		// Update note
		project.Notes = note
		project.UpdatedAt = time.Now()

		// Save
		if err := m.repository.Save(ctx, project); err != nil {
			return noteSaveErrorMsg{err: err}
		}

		return noteSavedMsg{projectID: projectID, newNote: note}
	}
}

// toggleFavorite toggles the favorite status of the selected project (Story 3.8).
func (m Model) toggleFavorite() (tea.Model, tea.Cmd) {
	selected := m.projectList.SelectedProject()
	if selected == nil {
		return m, nil
	}

	newFavorite := !selected.IsFavorite
	return m, m.saveFavoriteCmd(selected.ID, newFavorite)
}

// saveFavoriteCmd creates a command that saves the favorite status to repository (Story 3.8).
func (m Model) saveFavoriteCmd(projectID string, isFavorite bool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Find project
		project, err := m.repository.FindByID(ctx, projectID)
		if err != nil {
			return favoriteSaveErrorMsg{err: err}
		}

		// Update favorite status
		project.IsFavorite = isFavorite
		project.UpdatedAt = time.Now()

		// Save
		if err := m.repository.Save(ctx, project); err != nil {
			return favoriteSaveErrorMsg{err: err}
		}

		return favoriteSavedMsg{projectID: projectID, isFavorite: isFavorite}
	}
}

// startRemoveConfirmation opens the remove confirmation dialog (Story 3.9, 11.4 AC5).
func (m Model) startRemoveConfirmation() (tea.Model, tea.Cmd) {
	// Story 11.4: Get selected project based on current view
	var selected *domain.Project
	if m.viewMode == viewModeHibernated {
		selected = m.hibernatedList.SelectedProject()
	} else {
		selected = m.projectList.SelectedProject()
	}
	if selected == nil {
		return m, nil
	}

	m.isConfirmingRemove = true
	m.confirmTarget = selected

	// Start 30-second timeout timer (AC4)
	return m, tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return removeConfirmTimeoutMsg{}
	})
}

// handleRemoveConfirmationKeyMsg processes keyboard input during remove confirmation (Story 3.9).
func (m Model) handleRemoveConfirmationKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle Escape key by type (consistent with handleNoteEditingKeyMsg pattern)
	if msg.Type == tea.KeyEsc {
		m.isConfirmingRemove = false
		m.confirmTarget = nil
		return m, nil
	}

	// Handle character keys by string
	switch msg.String() {
	case "y", "Y":
		// Confirm removal
		m.isConfirmingRemove = false
		target := m.confirmTarget
		m.confirmTarget = nil

		// Use shared EffectiveName for display (Story 3.9 code review fix)
		projectName := project.EffectiveName(target)

		return m, m.removeProjectCmd(target.ID, projectName)

	case "n", "N":
		// Cancel removal
		m.isConfirmingRemove = false
		m.confirmTarget = nil
		return m, nil
	}

	// AC5: Ignore all other keys during confirmation
	return m, nil
}

// removeProjectCmd creates a command that deletes a project from repository (Story 3.9).
// Note: Reuses delete pattern from deleteProjectCmd but returns removeConfirmedMsg.
func (m Model) removeProjectCmd(projectID, projectName string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := m.repository.Delete(ctx, projectID)
		return removeConfirmedMsg{projectID: projectID, projectName: projectName, err: err}
	}
}

// View implements tea.Model. Renders the UI to a string.
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Check minimum terminal size
	if m.width < MinWidth || m.height < MinHeight {
		return renderTooSmallView(m.width, m.height)
	}

	// Render validation dialog when in validation mode (AC1, AC6)
	if m.viewMode == viewModeValidation && m.currentInvalidIdx < len(m.invalidProjects) {
		return renderValidationDialog(m.invalidProjects[m.currentInvalidIdx].Project, m.width, m.height, m.validationError)
	}

	// Render help overlay (overlays everything)
	if m.showHelp {
		return renderHelpOverlay(m.width, m.height, m.config)
	}

	// Render note editor dialog (overlays everything) (Story 3.7)
	if m.isEditingNote && m.noteEditTarget != nil {
		projectName := project.EffectiveName(m.noteEditTarget)
		return renderNoteEditor(projectName, m.noteInput, m.width, m.height)
	}

	// Render remove confirmation dialog (overlays everything) (Story 3.9)
	if m.isConfirmingRemove && m.confirmTarget != nil {
		projectName := project.EffectiveName(m.confirmTarget)
		return renderConfirmRemoveDialog(projectName, m.width, m.height)
	}

	// Story 12.1: Render text view (full screen log display)
	if m.viewMode == viewModeTextView {
		return m.renderTextView()
	}

	// Story 16.3: Render Stats View (full screen metrics display)
	if m.viewMode == viewModeStats {
		return m.renderStatsView()
	}

	// Story 12.1: Render session picker overlay
	if m.showSessionPicker {
		effectiveWidth := m.width
		if m.isWideWidth() {
			effectiveWidth = m.maxContentWidth
		}
		contentHeight := m.height - statusBarHeight(m.height)
		pickerContent := m.renderSessionPicker(effectiveWidth, contentHeight)
		statusBar := m.statusBar.View()
		content := lipgloss.JoinVertical(lipgloss.Left, pickerContent, statusBar)
		if m.isWideWidth() {
			return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, content)
		}
		return content
	}

	// Story 11.4: Render hibernated view (AC6)
	if m.viewMode == viewModeHibernated {
		contentHeight := m.height - statusBarHeight(m.height)
		var content string
		if len(m.hibernatedProjects) == 0 {
			content = renderHibernatedEmptyView(m.width, contentHeight)
		} else {
			content = renderHibernatedView(&m.hibernatedList, m.showDetailPanel, &m.detailPanel, m.width, contentHeight, m.maxContentWidth)
		}
		statusBar := m.statusBar.View()
		return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
	}

	// Render empty view if no projects (AC6)
	// Status bar is always visible per AC2, even in empty view
	if len(m.projects) == 0 {
		contentHeight := m.height - 2 // Reserve 2 lines for status bar
		emptyContent := renderEmptyView(m.width, contentHeight)
		statusBar := m.statusBar.View()
		return lipgloss.JoinVertical(lipgloss.Left, emptyContent, statusBar)
	}

	// Render project list (Story 3.1)
	return m.renderDashboard()
}

// renderDashboard renders the main dashboard with project list, optional detail panel, and status bar.
func (m Model) renderDashboard() string {
	// Calculate effective width (cap at maxContentWidth for wide terminals - Story 3.10 AC4, 8.10)
	// Note: isNarrowWidth (60-79) and isWideWidth (>maxContentWidth) are mutually exclusive
	effectiveWidth := m.width
	if m.isWideWidth() {
		effectiveWidth = m.maxContentWidth
	}

	// Reserve lines for status bar using helper (Story 3.10 AC5)
	contentHeight := m.height - statusBarHeight(m.height)

	// Adjust content height if narrow warning is shown (Story 3.10 AC2)
	if isNarrowWidth(m.width) {
		contentHeight-- // Reserve 1 more line for warning
	}

	// Create a copy with effective width for rendering
	renderModel := m
	renderModel.width = effectiveWidth

	mainContent := renderModel.renderMainContent(contentHeight)

	// Build output parts
	var parts []string
	parts = append(parts, mainContent)

	// Add narrow warning if applicable (Story 3.10 AC2)
	if isNarrowWidth(m.width) {
		parts = append(parts, renderNarrowWarning(effectiveWidth))
	}

	// Use m.statusBar directly - width already set in resizeTickMsg
	parts = append(parts, m.statusBar.View())

	// Join content
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	// Center content if terminal is wider than maxContentWidth (Story 3.10 AC4, 8.10)
	if m.isWideWidth() {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, content)
	}

	return content
}

// renderMainContent renders the main content area (project list and optional detail panel).
func (m Model) renderMainContent(height int) string {
	// Show hint when terminal height MinHeight-HeightThresholdTall and detail panel closed (Story 3.10 AC6)
	// IMPORTANT: Use m.height (terminal height) not height parameter (contentHeight)
	// because AC6 defines behavior based on user-visible terminal size
	if m.height >= MinHeight && m.height < HeightThresholdTall && !m.showDetailPanel {
		// Create copy with reduced height to account for hint line
		projectList := m.projectList
		projectList.SetSize(m.width, height-1) // Use receiver's capped width
		listView := projectList.View()

		hint := DimStyle.Render("Press [d] for details")
		return listView + "\n" + hint
	}

	// Full-width project list when detail panel is hidden
	if !m.showDetailPanel {
		projectList := m.projectList
		projectList.SetSize(m.width, height) // Use receiver's capped width
		return projectList.View()
	}

	// Story 8.6: Check layout mode
	if m.isHorizontalLayout() {
		return m.renderHorizontalSplit(height)
	}

	// Vertical (side-by-side) layout - existing code
	// Split layout: list (60%) | detail (40%)
	listWidth := int(float64(m.width) * 0.6)
	detailWidth := m.width - listWidth - 1 // -1 for separator space

	// Create copies with updated sizes for this render
	projectList := m.projectList
	projectList.SetSize(listWidth, height)

	detailPanel := m.detailPanel
	detailPanel.SetSize(detailWidth, height)

	// Render side by side
	listView := projectList.View()
	detailView := detailPanel.View()

	return lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView)
}

// renderHorizontalSplit renders project list above detail panel (top/bottom).
// Story 8.6: Used when config detail_layout=horizontal.
// Story 8.12: Implements height-priority algorithm - list prioritized over detail.
func (m Model) renderHorizontalSplit(height int) string {
	// Height priority: if too short, hide detail and show only list
	if height < HorizontalDetailThreshold {
		projectList := m.projectList
		projectList.SetSize(m.width, height) // Use receiver's capped width
		return projectList.View()
	}

	// Render detail panel first to get actual height (use regular border with all 4 sides)
	detailPanel := m.detailPanel
	detailPanel.SetVisible(true)
	detailPanel.SetProject(m.projectList.SelectedProject())
	detailPanel.SetHorizontalMode(false) // Use regular border (includes top)
	detailPanel.SetSize(m.width, 0)      // height doesn't matter, it renders full content
	detailView := detailPanel.View()

	// Count actual detail lines
	detailLines := strings.Count(detailView, "\n") + 1
	listHeight := height - detailLines - 1 // -1 for newline between

	// List with remaining height - use m.width (receiver's capped width), not cached width
	projectList := m.projectList
	projectList.SetSize(m.width, listHeight)
	listView := projectList.View()

	return listView + "\n" + detailView
}

// handleFileEvent processes a file system event and updates project state (Story 4.6, 11.3).
func (m *Model) handleFileEvent(msg fileEventMsg) {
	// Find project by path prefix
	project := m.findProjectByPath(msg.Path)
	if project == nil {
		slog.Debug("event path not matched to project", "path", msg.Path)
		return
	}

	// Story 11.3: Auto-activate hibernated project on file activity (AC1, AC2)
	if project.State == domain.StateHibernated && m.stateService != nil {
		ctx := context.Background()
		if err := m.stateService.Activate(ctx, project.ID); err != nil {
			// AC6: Log warning but continue (partial failure tolerance)
			// AC3: ErrInvalidStateTransition is expected during races, log at debug
			if !errors.Is(err, domain.ErrInvalidStateTransition) {
				slog.Warn("failed to auto-activate project",
					"project_id", project.ID,
					"project_name", project.Name,
					"error", err)
			}
			// Note: Don't return - still update LastActivityAt in memory
		} else {
			// AC8: Log successful activation for debugging
			slog.Debug("auto-activated hibernated project",
				"project_id", project.ID,
				"project_name", project.Name)
			// Update local state to reflect activation (AC1, AC4)
			project.State = domain.StateActive
			project.HibernatedAt = nil
		}
	}

	// Update repository (skip if nil - e.g., in tests without mocked repo)
	ctx := context.Background()
	if m.repository == nil {
		slog.Debug("repository is nil, skipping activity update", "project_id", project.ID)
		return
	}
	if err := m.repository.UpdateLastActivity(ctx, project.ID, msg.Timestamp); err != nil {
		slog.Warn("failed to update activity", "project_id", project.ID, "error", err)
		return
	}

	// Epic 4 Hotfix H3: Log successful activity update for debugging
	slog.Debug("activity updated", "project", project.Name, "path", msg.Path)

	// Update local state
	project.LastActivityAt = msg.Timestamp

	// Update detail panel if this is selected project
	if m.detailPanel.Project() != nil && m.detailPanel.Project().ID == project.ID {
		m.detailPanel.SetProject(project)
	}

	// Recalculate status bar (waiting may have cleared, counts may have changed from activation)
	active, hibernated, waiting := components.CalculateCountsWithWaiting(m.projects, m.isProjectWaiting)
	m.statusBar.SetCounts(active, hibernated, waiting)
}

// startFileWatcherForProjects starts the file watcher for all projects if available.
// Story 8.4 code review: Extracted to eliminate duplication between ProjectsLoadedMsg
// and resizeTickMsg handlers.
// Returns a tea.Cmd to wait for the next file event, or nil if watcher unavailable.
func (m *Model) startFileWatcherForProjects() tea.Cmd {
	if m.fileWatcher == nil || !m.fileWatcherAvailable || len(m.projects) == 0 {
		return nil
	}

	// Story 9.5-2 fix: Skip restart if watcher is already running
	// Only restart on first launch or after failure (when eventCh becomes nil)
	// This prevents unnecessary restarts on every refresh/ProjectsLoadedMsg
	if m.eventCh != nil {
		return nil
	}

	// Collect project paths
	paths := make([]string, len(m.projects))
	for i, p := range m.projects {
		paths[i] = p.Path
	}

	// Story 8.13: Cancel old context BEFORE calling Watch()
	// This ensures the old waitForNextFileEventCmd exits cleanly via context.Done()
	// instead of receiving channel close and triggering a false error warning.
	if m.watchCancel != nil {
		m.watchCancel()
	}

	// Create watch context for cancellation
	m.watchCtx, m.watchCancel = context.WithCancel(context.Background())

	// Story 9.5-2: Record timestamp BEFORE Watch() for grace period check
	m.lastWatcherRestart = time.Now()

	// Start watching (Story 8.9: emoji fallback)
	eventCh, err := m.fileWatcher.Watch(m.watchCtx, paths)
	if err != nil {
		slog.Warn("failed to start file watcher", "error", err)
		m.fileWatcherAvailable = false
		m.statusBar.SetWatcherWarning(emoji.Warning() + " File watching unavailable")
		return nil
	}

	m.eventCh = eventCh
	slog.Debug("file watcher started", "project_count", len(paths))

	// Story 7.1: Check for partial failures
	failedPaths := m.fileWatcher.GetFailedPaths()
	if len(failedPaths) > 0 {
		return tea.Batch(
			m.waitForNextFileEventCmd(),
			func() tea.Msg {
				return watcherWarningMsg{
					failedPaths: failedPaths,
					totalPaths:  len(paths),
				}
			},
		)
	}

	return m.waitForNextFileEventCmd()
}

// findProjectByPath finds the project that owns the given file path (Story 4.6).
// Uses path prefix matching - returns the project whose Path is a prefix of eventPath.
func (m Model) findProjectByPath(eventPath string) *domain.Project {
	eventPath = strings.TrimSuffix(eventPath, "/")
	for _, p := range m.projects {
		projectPath := strings.TrimSuffix(p.Path, "/")
		if eventPath == projectPath || strings.HasPrefix(eventPath, projectPath+"/") {
			return p
		}
	}
	return nil
}

// Story 12.1: Log viewer methods (spawns jq external tool)

// handleEnterForLogs handles Enter key to open most recent log session with jq (AC1).
func (m Model) handleEnterForLogs() (tea.Model, tea.Cmd) {
	if len(m.projects) == 0 {
		return m, nil
	}

	selected := m.projectList.SelectedProject()
	if selected == nil {
		return m, nil
	}

	// Check if log reader registry is available
	if m.logReaderRegistry == nil {
		return m, func() tea.Msg {
			return flashMsg{text: "Log viewing not available"}
		}
	}

	// Get a reader for this project
	ctx := context.Background()
	reader := m.logReaderRegistry.GetReader(ctx, selected.Path)
	if reader == nil {
		// AC8: No logs exist for this project - show flash message
		return m, func() tea.Msg {
			return flashMsg{text: "No Claude Code logs for this project"}
		}
	}

	// List sessions to find most recent
	sessions, err := reader.ListSessions(ctx, selected.Path)
	if err != nil || len(sessions) == 0 {
		return m, func() tea.Msg {
			return flashMsg{text: "No sessions found for this project"}
		}
	}

	// Store project and reader for S key session picker (AC6)
	m.currentLogProject = selected
	m.currentLogReader = reader
	m.logSessions = sessions

	// Sessions are sorted newest-first, so first one is most recent
	mostRecent := sessions[0]
	return m, m.openLogWithJqCmd(mostRecent.Path, project.EffectiveName(selected))
}

// handleShiftEnterForSessionPicker handles Shift+Enter to show session picker (AC6).
// Can be called from normal view (Shift+Enter) or from text view (S key).
func (m Model) handleShiftEnterForSessionPicker() (tea.Model, tea.Cmd) {
	// If already in text view with a project, reuse current context
	if m.currentLogProject != nil && m.currentLogReader != nil {
		m.showSessionPicker = true
		m.sessionPickerIndex = 0
		// Use cached sessions if available, otherwise reload
		if len(m.logSessions) > 0 {
			return m, nil
		}
		return m, m.loadLogSessionsCmd(m.currentLogProject.Path)
	}

	// Normal view: get selected project
	if len(m.projects) == 0 {
		return m, nil
	}

	selected := m.projectList.SelectedProject()
	if selected == nil {
		return m, nil
	}

	// Check if log reader registry is available
	if m.logReaderRegistry == nil {
		return m, func() tea.Msg {
			return flashMsg{text: "Log viewing not available"}
		}
	}

	// Get a reader for this project
	ctx := context.Background()
	reader := m.logReaderRegistry.GetReader(ctx, selected.Path)
	if reader == nil {
		// AC8: No logs exist for this project - show flash message
		return m, func() tea.Msg {
			return flashMsg{text: "No Claude Code logs for this project"}
		}
	}

	// Store state for session picker
	m.currentLogReader = reader
	m.currentLogProject = selected
	m.showSessionPicker = true
	m.sessionPickerIndex = 0
	m.logSessions = nil // Clear previous sessions

	// Load sessions asynchronously
	return m, m.loadLogSessionsCmd(selected.Path)
}

// openLogWithJqCmd creates a command to format and display log content.
// Pipeline priority: cclv (primary) -> jq (secondary) -> raw (fallback)
func (m Model) openLogWithJqCmd(sessionPath, projectName string) tea.Cmd {
	return func() tea.Msg {
		// Extract session name from path for title
		sessionName := filepath.Base(sessionPath)
		if len(sessionName) > 40 {
			sessionName = sessionName[:37] + "..."
		}
		// Include project name in title: "ProjectName - session.jsonl"
		title := projectName + " - " + sessionName

		// Get initial file size for offset tracking (AC4: live tailing)
		info, statErr := os.Stat(sessionPath)
		var fileSize int64
		if statErr == nil {
			fileSize = info.Size()
		}

		// Primary: Try cclv (Claude Code Log Viewer) first with color support for pipeline
		cmd := exec.Command("cclv", "--color=always", sessionPath)
		output, err := cmd.Output()
		if err == nil {
			return textViewContentMsg{
				title:       title,
				content:     string(output),
				sessionPath: sessionPath,
				fileSize:    fileSize,
			}
		}
		slog.Debug("cclv not available, trying jq", "error", err)

		// Secondary: Try jq for pretty-printing
		cmd = exec.Command("jq", ".", sessionPath)
		output, err = cmd.Output()
		if err == nil {
			return textViewContentMsg{
				title:       title + " (jq)",
				content:     string(output),
				sessionPath: sessionPath,
				fileSize:    fileSize,
			}
		}
		slog.Debug("jq not available, falling back to raw display", "error", err)

		// Fallback: read raw JSONL file (AC2: display raw JSON entries)
		rawContent, readErr := os.ReadFile(sessionPath)
		if readErr != nil {
			return textViewContentMsg{
				err: fmt.Errorf("failed to read log file: %w", readErr),
			}
		}

		return textViewContentMsg{
			title:       title + " (raw)",
			content:     string(rawContent),
			sessionPath: sessionPath,
			fileSize:    fileSize,
		}
	}
}

// loadLogSessionsCmd creates a command to load available log sessions.
func (m Model) loadLogSessionsCmd(projectPath string) tea.Cmd {
	return func() tea.Msg {
		if m.currentLogReader == nil {
			return logSessionsLoadedMsg{err: fmt.Errorf("no log reader available")}
		}
		ctx := context.Background()
		sessions, err := m.currentLogReader.ListSessions(ctx, projectPath)
		return logSessionsLoadedMsg{sessions: sessions, err: err}
	}
}

// handleSessionPickerKeyMsg handles keyboard input in session picker overlay.
func (m Model) handleSessionPickerKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case KeyEscape:
		m.showSessionPicker = false
		m.currentLogReader = nil
		m.currentLogProject = nil
		return m, nil

	case KeyDown, KeyDownArrow:
		if m.sessionPickerIndex < len(m.logSessions)-1 {
			m.sessionPickerIndex++
		}
		return m, nil

	case KeyUp, KeyUpArrow:
		if m.sessionPickerIndex > 0 {
			m.sessionPickerIndex--
		}
		return m, nil

	case "enter":
		// Select session and spawn jq
		if m.sessionPickerIndex < len(m.logSessions) {
			selected := m.logSessions[m.sessionPickerIndex]
			// Capture project name before resetting (for title display)
			projectName := ""
			if m.currentLogProject != nil {
				projectName = project.EffectiveName(m.currentLogProject)
			}
			m.showSessionPicker = false
			m.currentLogReader = nil
			m.currentLogProject = nil
			return m, m.openLogWithJqCmd(selected.Path, projectName)
		}
		return m, nil
	}

	return m, nil
}

// renderSessionPicker renders the session picker overlay.
func (m Model) renderSessionPicker(width, height int) string {
	if len(m.logSessions) == 0 {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, "No sessions available")
	}

	// Calculate max visible sessions based on height
	// Reserve: title(1) + blank(1) + footer(1) + blank(1) + border(2) + padding(2) = 8 lines
	maxVisible := height - 8
	if maxVisible < 3 {
		maxVisible = 3
	}
	if maxVisible > len(m.logSessions) {
		maxVisible = len(m.logSessions)
	}

	// Calculate scroll window
	startIdx := 0
	if m.sessionPickerIndex >= maxVisible {
		startIdx = m.sessionPickerIndex - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(m.logSessions) {
		endIdx = len(m.logSessions)
		startIdx = endIdx - maxVisible
		if startIdx < 0 {
			startIdx = 0
		}
	}

	var lines []string
	lines = append(lines, "Select Session:")
	lines = append(lines, "")

	for i := startIdx; i < endIdx; i++ {
		session := m.logSessions[i]
		prefix := "  "
		if i == m.sessionPickerIndex {
			prefix = "> "
		}

		// Format: "> session-id  timestamp  (N entries)"
		displayID := session.ID
		if len(displayID) > 16 {
			displayID = displayID[:16] + "..."
		}

		timestamp := session.StartTime.Format("2006-01-02 15:04")
		entryInfo := fmt.Sprintf("(%d entries)", session.EntryCount)

		line := fmt.Sprintf("%s%s  %s  %s", prefix, displayID, timestamp, entryInfo)
		lines = append(lines, line)
	}

	// Show scroll indicator if needed
	if len(m.logSessions) > maxVisible {
		lines = append(lines, fmt.Sprintf("  (%d/%d)", m.sessionPickerIndex+1, len(m.logSessions)))
	}

	lines = append(lines, "")
	lines = append(lines, "[Enter] Select  [Esc] Cancel")

	content := strings.Join(lines, "\n")

	// Center the picker
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

// Story 16.3/16.5/16.6: handleStatsViewKeyMsg handles keyboard input in Stats View.
// Story 16.5: Supports breakdown sub-view with Enter to drill down, Esc to return.
// Story 16.6: Supports date range cycling with [ and ] keys.
func (m Model) handleStatsViewKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case KeyEscape, KeyQuit:
		// Story 16.5: If in breakdown view, return to project list
		if m.statsBreakdownProject != nil {
			m.statsBreakdownProject = nil
			m.statsBreakdownDurations = nil
			return m, nil
		}
		// Exit Stats View entirely (return to dashboard)
		m.exitStatsView()
		return m, nil
	case "enter":
		// Story 16.5: Enter breakdown view for selected project
		if m.statsBreakdownProject == nil && len(m.projects) > 0 {
			selectedIdx := m.statsViewScroll
			if selectedIdx >= 0 && selectedIdx < len(m.projects) {
				p := m.projects[selectedIdx]
				m.statsBreakdownProject = p
				m.statsBreakdownDurations = m.getStageBreakdown(p.Path)
			}
		}
		return m, nil
	case "[":
		// Story 16.6: Cycle to previous date range preset (only in project list view)
		if m.statsBreakdownProject == nil {
			m.statsDateRange = m.statsDateRange.Prev()
		}
		return m, nil
	case "]":
		// Story 16.6: Cycle to next date range preset (only in project list view)
		if m.statsBreakdownProject == nil {
			m.statsDateRange = m.statsDateRange.Next()
		}
		return m, nil
	case KeyDown, KeyDownArrow:
		// Only scroll in project list view (not breakdown)
		if m.statsBreakdownProject == nil && len(m.projects) > 0 {
			if m.statsViewScroll < len(m.projects)-1 {
				m.statsViewScroll++
			}
		}
		return m, nil
	case KeyUp, KeyUpArrow:
		// Only scroll in project list view (not breakdown)
		if m.statsBreakdownProject == nil {
			if m.statsViewScroll > 0 {
				m.statsViewScroll--
			}
		}
		return m, nil
	}
	return m, nil
}

// handleTextViewKeyMsg handles keyboard input in text view mode.
func (m Model) handleTextViewKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	contentHeight := m.height - statusBarHeight(m.height) - 2 // -2 for header/footer
	maxScroll := len(m.textViewContent) - contentHeight
	if maxScroll < 0 {
		maxScroll = 0
	}

	switch msg.String() {
	case KeyEscape, KeyQuit:
		// Story 12.2 AC3: If in search mode, exit search mode first
		if m.searchMode {
			m.searchMode = false
			m.searchInput = ""
			m.searchQuery = ""
			m.searchIndex = 0
			m.searchMatches = nil
			return m, nil
		}
		// Return to normal view and stop tailing
		m.viewMode = viewModeNormal
		m.textViewContent = nil
		m.textViewTitle = ""
		m.textViewScroll = 0
		m.currentLogProject = nil
		m.currentLogReader = nil
		m.logSessions = nil
		m.logTailActive = false // Stop tailing
		m.currentSessionPath = ""
		m.logAutoScroll = false
		// Story 12.2 AC2: Reset double-key detection state
		m.lastKeyPress = ""
		// Story 12.2 AC3: Reset search state
		m.searchMode = false
		m.searchQuery = ""
		m.searchInput = ""
		m.searchIndex = 0
		m.searchMatches = nil
		return m, nil

	case KeyForceQuit:
		// Story 12.2 AC3: In search mode, Ctrl+C exits search instead of quitting
		if m.searchMode {
			m.searchMode = false
			m.searchInput = ""
			m.searchQuery = ""
			m.searchIndex = 0
			m.searchMatches = nil
			return m, nil
		}
		// Exit completely
		if m.watchCancel != nil {
			m.watchCancel()
		}
		if m.fileWatcher != nil {
			m.fileWatcher.Close()
		}
		return m, tea.Quit

	case KeyDown: // 'j' - scroll down (but typeable in search mode)
		// If actively typing in search mode, add to input
		if m.searchMode {
			m.searchInput += "j"
			return m, nil
		}
		m.logAutoScroll = false
		m.lastKeyPress = "" // Reset gg detection
		if m.textViewScroll < maxScroll {
			m.textViewScroll++
		}
		return m, nil

	case KeyDownArrow: // Arrow down - always scroll
		m.logAutoScroll = false
		m.lastKeyPress = "" // Reset gg detection
		if m.textViewScroll < maxScroll {
			m.textViewScroll++
		}
		return m, nil

	case KeyUp: // 'k' - scroll up (but typeable in search mode)
		// If actively typing in search mode, add to input
		if m.searchMode {
			m.searchInput += "k"
			return m, nil
		}
		m.logAutoScroll = false
		m.lastKeyPress = "" // Reset gg detection
		if m.textViewScroll > 0 {
			m.textViewScroll--
		}
		return m, nil

	case KeyUpArrow: // Arrow up - always scroll
		m.logAutoScroll = false
		m.lastKeyPress = "" // Reset gg detection
		if m.textViewScroll > 0 {
			m.textViewScroll--
		}
		return m, nil

	case "ctrl+d": // Story 12.2 AC2: Half-page down
		m.logAutoScroll = false // Pause auto-scroll
		m.lastKeyPress = ""     // Reset gg detection
		halfPage := contentHeight / 2
		m.textViewScroll += halfPage
		if m.textViewScroll > maxScroll {
			m.textViewScroll = maxScroll
		}
		return m, nil

	case "ctrl+u": // Story 12.2 AC2: Half-page up
		m.logAutoScroll = false // Pause auto-scroll
		m.lastKeyPress = ""     // Reset gg detection
		halfPage := contentHeight / 2
		m.textViewScroll -= halfPage
		if m.textViewScroll < 0 {
			m.textViewScroll = 0
		}
		return m, nil

	case "g": // Story 12.2 AC2: Double 'g' jumps to top (vim-standard)
		// If actively typing in search mode, add to input instead of navigation
		if m.searchMode {
			m.searchInput += "g"
			return m, nil
		}
		now := time.Now()
		// Check if this is the second 'g' within timeout
		if m.lastKeyPress == "g" && now.Sub(m.lastKeyTime).Milliseconds() <= ggTimeoutMs {
			m.logAutoScroll = false // Pause auto-scroll
			m.textViewScroll = 0
			m.lastKeyPress = "" // Reset
			return m, nil
		}
		// First 'g' - record and wait for second
		m.lastKeyPress = "g"
		m.lastKeyTime = now
		return m, nil

	case KeyLogJumpEnd: // AC3: Jump to end and resume auto-scroll (G)
		// If actively typing in search mode, add to input instead of navigation
		if m.searchMode {
			m.searchInput += "G"
			return m, nil
		}
		m.textViewScroll = maxScroll
		m.logAutoScroll = true // Resume auto-scroll
		m.lastKeyPress = ""    // Reset gg detection
		return m, nil

	case KeyLogSession: // AC6: Open session picker from log view (S)
		// If actively typing in search mode, add to input instead of action
		if m.searchMode {
			m.searchInput += "S"
			return m, nil
		}
		m.lastKeyPress = "" // Reset gg detection
		if m.currentLogProject != nil {
			return m.handleShiftEnterForSessionPicker()
		}
		return m, nil

	case " ": // Space - page down (but typeable in search mode)
		// If actively typing in search mode, add to input
		if m.searchMode {
			m.searchInput += " "
			return m, nil
		}
		m.logAutoScroll = false // Pause auto-scroll
		m.lastKeyPress = ""     // Reset gg detection
		m.textViewScroll += contentHeight
		if m.textViewScroll > maxScroll {
			m.textViewScroll = maxScroll
		}
		return m, nil

	case "pgdown": // Page down (not typeable, keep as-is)
		m.logAutoScroll = false // Pause auto-scroll
		m.lastKeyPress = ""     // Reset gg detection
		m.textViewScroll += contentHeight
		if m.textViewScroll > maxScroll {
			m.textViewScroll = maxScroll
		}
		return m, nil

	case "b": // Page up (but typeable in search mode)
		// If actively typing in search mode, add to input
		if m.searchMode {
			m.searchInput += "b"
			return m, nil
		}
		m.logAutoScroll = false // Pause auto-scroll
		m.lastKeyPress = ""     // Reset gg detection
		m.textViewScroll -= contentHeight
		if m.textViewScroll < 0 {
			m.textViewScroll = 0
		}
		return m, nil

	case "pgup": // Page up (not typeable, keep as-is)
		m.logAutoScroll = false // Pause auto-scroll
		m.lastKeyPress = ""     // Reset gg detection
		m.textViewScroll -= contentHeight
		if m.textViewScroll < 0 {
			m.textViewScroll = 0
		}
		return m, nil

	case "/": // Story 12.2 AC3: Enter search mode
		if !m.searchMode {
			m.searchMode = true
			m.searchInput = ""
			m.searchQuery = ""
			m.searchMatches = nil
			m.searchIndex = 0
		}
		return m, nil

	case "n": // Story 12.2 AC3: Next match
		// If actively typing in search mode, add to input instead of navigating
		if m.searchMode {
			m.searchInput += "n"
			return m, nil
		}
		if len(m.searchMatches) > 0 {
			m.searchIndex = (m.searchIndex + 1) % len(m.searchMatches)
			m.textViewScroll = m.searchMatches[m.searchIndex]
			// Center the match if possible
			if m.textViewScroll > contentHeight/2 {
				m.textViewScroll -= contentHeight / 2
			} else {
				m.textViewScroll = 0
			}
			m.logAutoScroll = false
		}
		return m, nil

	case "N": // Story 12.2 AC3: Previous match
		// If actively typing in search mode, add to input instead of navigating
		if m.searchMode {
			m.searchInput += "N"
			return m, nil
		}
		if len(m.searchMatches) > 0 {
			m.searchIndex--
			if m.searchIndex < 0 {
				m.searchIndex = len(m.searchMatches) - 1
			}
			m.textViewScroll = m.searchMatches[m.searchIndex]
			// Center the match if possible
			if m.textViewScroll > contentHeight/2 {
				m.textViewScroll -= contentHeight / 2
			} else {
				m.textViewScroll = 0
			}
			m.logAutoScroll = false
		}
		return m, nil

	case "enter": // Story 12.2 AC3: Execute search
		if m.searchMode && m.searchInput != "" {
			m.searchQuery = m.searchInput
			m.searchMatches = m.findMatches(m.searchQuery)
			m.searchIndex = 0
			m.searchMode = false // Exit input mode so n/N can navigate
			if len(m.searchMatches) > 0 {
				// Jump to first match
				m.textViewScroll = m.searchMatches[0]
				if m.textViewScroll > contentHeight/2 {
					m.textViewScroll -= contentHeight / 2
				} else {
					m.textViewScroll = 0
				}
			} else {
				// Story 12.2 AC3: Show "Pattern not found" flash message
				return m, func() tea.Msg {
					return flashMsg{text: "Pattern not found"}
				}
			}
			m.logAutoScroll = false
		}
		return m, nil

	case "backspace": // Story 12.2 AC3: Delete character in search input
		if m.searchMode && len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
		}
		return m, nil
	}

	// Story 12.2 AC3: Handle typing in search mode
	if m.searchMode {
		// Accept printable characters
		if msg.Type == tea.KeyRunes {
			m.searchInput += string(msg.Runes)
			return m, nil
		}
	}

	return m, nil
}

// findMatches returns line numbers containing the search query (case-insensitive).
func (m Model) findMatches(query string) []int {
	if query == "" {
		return nil
	}
	lowerQuery := strings.ToLower(query)
	var matches []int
	for i, line := range m.textViewContent {
		if strings.Contains(strings.ToLower(line), lowerQuery) {
			matches = append(matches, i)
		}
	}
	return matches
}

// Story 12.2 AC4: ANSI-aware string utilities for proper truncation

// stripANSI removes ANSI escape sequences from a string.
// Used for calculating visual width without color codes.
func stripANSI(s string) string {
	return ansi.Strip(s)
}

// visibleWidth returns the display width of a string excluding ANSI codes.
// Uses runewidth to handle wide characters (CJK, etc.) correctly.
func visibleWidth(s string) int {
	return runewidth.StringWidth(stripANSI(s))
}

// truncateToWidth truncates a string to fit within the given visual width,
// preserving ANSI escape sequences and adding "..." suffix.
// This ensures colored output is truncated correctly without breaking colors.
func truncateToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if visibleWidth(s) <= width {
		return s
	}

	// We need to truncate while preserving ANSI sequences
	// Track visual width as we iterate through the string
	var result strings.Builder
	visualWidth := 0
	inEscapeSeq := false
	escapeSeq := strings.Builder{}

	for _, r := range s {
		if inEscapeSeq {
			escapeSeq.WriteRune(r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				// End of escape sequence
				result.WriteString(escapeSeq.String())
				escapeSeq.Reset()
				inEscapeSeq = false
			}
			continue
		}

		if r == '\x1b' {
			// Start of escape sequence
			inEscapeSeq = true
			escapeSeq.WriteRune(r)
			continue
		}

		// Regular character - check width
		charWidth := runewidth.RuneWidth(r)
		if visualWidth+charWidth > width-3 { // -3 for "..."
			result.WriteString("...")
			// Append any pending ANSI reset to preserve terminal state
			return result.String()
		}
		result.WriteRune(r)
		visualWidth += charWidth
	}

	// If we're still in an escape sequence, flush it
	if inEscapeSeq {
		result.WriteString(escapeSeq.String())
	}

	return result.String()
}

// renderTextView renders the scrollable text view.
func (m Model) renderTextView() string {
	effectiveWidth := m.width
	if m.isWideWidth() {
		effectiveWidth = m.maxContentWidth
	}
	contentHeight := m.height - statusBarHeight(m.height) - 2 // -2 for header/footer

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("255")).
		Width(effectiveWidth).
		Render(" " + m.textViewTitle)

	// Content with scroll
	var visibleLines []string
	startLine := m.textViewScroll
	endLine := startLine + contentHeight
	if endLine > len(m.textViewContent) {
		endLine = len(m.textViewContent)
	}

	// Story 12.2 AC3: Highlight style for current match
	highlightStyle := lipgloss.NewStyle().Reverse(true)

	for i := startLine; i < endLine; i++ {
		line := m.textViewContent[i]
		// Story 12.2 AC4: Truncate using visual width (handles ANSI codes correctly)
		if visibleWidth(line) > effectiveWidth {
			line = truncateToWidth(line, effectiveWidth)
		}
		// Story 12.2 AC3: Highlight current match line
		if len(m.searchMatches) > 0 && m.searchIndex < len(m.searchMatches) {
			if i == m.searchMatches[m.searchIndex] {
				line = highlightStyle.Render(line)
			}
		}
		visibleLines = append(visibleLines, line)
	}

	// Pad to fill height
	for len(visibleLines) < contentHeight {
		visibleLines = append(visibleLines, "")
	}

	contentStr := strings.Join(visibleLines, "\n")

	// Footer with keybindings and scroll position
	maxScroll := len(m.textViewContent) - contentHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	scrollPercent := 0
	if maxScroll > 0 {
		scrollPercent = (m.textViewScroll * 100) / maxScroll
	} else if len(m.textViewContent) <= contentHeight {
		scrollPercent = 100
	}

	// Story 12.2 AC3: Show search footer when in search mode
	var footerText string
	if m.searchMode {
		// Search mode footer with input and match counter
		matchCounter := "0/0"
		if len(m.searchMatches) > 0 {
			matchCounter = fmt.Sprintf("%d/%d", m.searchIndex+1, len(m.searchMatches))
		}
		searchDisplay := m.searchInput
		// Truncate query if too long
		maxQueryLen := effectiveWidth - 40 // Reserve space for controls
		if maxQueryLen < 10 {
			maxQueryLen = 10
		}
		if len(searchDisplay) > maxQueryLen {
			searchDisplay = searchDisplay[:maxQueryLen-3] + "..."
		}
		footerText = fmt.Sprintf(" /%s_  [n/N] Next/Prev  %s  %d%%", searchDisplay, matchCounter, scrollPercent)
	} else if len(m.searchMatches) > 0 {
		// Have search results but not in input mode
		matchCounter := fmt.Sprintf("%d/%d", m.searchIndex+1, len(m.searchMatches))
		footerText = fmt.Sprintf(" [/] Search  [n/N] %s  [Esc] Clear  %d%%", matchCounter, scrollPercent)
	} else {
		footerText = fmt.Sprintf(" [j/k] Scroll  [gg/G] Top/Bottom  [/] Search  [Esc] Exit  %d%% ", scrollPercent)
	}
	footer := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("244")).
		Width(effectiveWidth).
		Render(footerText)

	// Combine
	textContent := lipgloss.JoinVertical(lipgloss.Left, header, contentStr, footer)

	// Add status bar
	statusBar := m.statusBar.View()
	combined := lipgloss.JoinVertical(lipgloss.Left, textContent, statusBar)

	if m.isWideWidth() {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top, combined)
	}
	return combined
}

// Story 16.3: enterStatsView saves dashboard selection and switches to stats view.
// Story 16.6: Initializes date range to default (30 days) on each entry (AC #7).
func (m *Model) enterStatsView() {
	m.statsActiveProjectIdx = m.projectList.Index()
	m.viewMode = viewModeStats
	m.statsViewScroll = 0
	m.statsDateRange = statsview.DefaultDateRange() // Story 16.6: Reset to default on entry
}

// Story 16.3: exitStatsView returns to normal view with bounds-checked selection restoration.
func (m *Model) exitStatsView() {
	m.viewMode = viewModeNormal
	if m.statsActiveProjectIdx >= 0 && m.statsActiveProjectIdx < len(m.projects) {
		m.projectList.SelectByIndex(m.statsActiveProjectIdx)
	}
}
