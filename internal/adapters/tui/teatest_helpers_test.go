// Package tui provides terminal user interface components for vibe-dash.
//
// # Teatest Helper Functions
//
// This file provides reusable test helper functions for behavioral testing
// of vibe-dash TUI using the teatest framework. It was created as part of
// Story 9.2 to establish a consistent testing foundation.
//
// ## Background
//
// Story 9.1 evaluated teatest and created a proof-of-concept in teatest_poc_test.go.
// This file builds upon that PoC to provide:
//   - Consistent test setup with sensible defaults
//   - Terminal size presets for common scenarios
//   - Resize simulation helpers
//   - Golden file path utilities
//
// ## Usage Example
//
//	func TestMyFeature(t *testing.T) {
//	    tm := NewTeatestModel(t, WithTermSizePreset(TermSizeNarrow))
//	    defer tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))
//
//	    // Perform actions...
//	    ResizeTerminal(tm, TermWidthWide, TermHeightStandard)
//
//	    // Verify output...
//	}
//
// ## Relationship to Story 9.1 PoC
//
// The helpers in this file wrap the patterns established in teatest_poc_test.go:
//   - Model creation with mock repository (lines 44-81)
//   - Color profile forcing for determinism (line 103)
//   - Wait patterns for ready state (lines 107-116)
//   - Key send with small delays (lines 119-124)
package tui

import (
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/muesli/termenv"

	"github.com/JeiKeiLim/vibe-dash/internal/core/domain"
	"github.com/JeiKeiLim/vibe-dash/internal/core/ports"
)

// ============================================================================
// Test Initialization
// ============================================================================
//
// NOTE: We do NOT use a global init() to set environment variables, as this
// would affect other tests in the package. Instead, we call
// lipgloss.SetColorProfile(termenv.Ascii) within NewTeatestModel() to ensure
// deterministic output for each test model.
//
// For CI environments, the env vars NO_COLOR=1, FORCE_COLOR=0, and TERM=dumb
// are set in .github/workflows/ci.yml (see Story 9.2 AC4).

// ============================================================================
// Terminal Size Constants and Presets (AC2)
// ============================================================================

// Terminal width constants for common scenarios.
// These values are derived from actual usage patterns and model.go thresholds.
const (
	// TermWidthStandard is the traditional terminal width (80 columns).
	TermWidthStandard = 80

	// TermWidthNarrow triggers narrow/vertical-only layout (40 columns).
	TermWidthNarrow = 40

	// TermWidthWide enables wide layout features (160 columns).
	TermWidthWide = 160

	// TermWidthUltraWide goes beyond typical max_content_width (200 columns).
	TermWidthUltraWide = 200

	// TermWidthMinimum is the smallest viable terminal width (20 columns).
	TermWidthMinimum = 20
)

// Terminal height constants for common scenarios.
const (
	// TermHeightStandard is the traditional terminal height (24 rows).
	TermHeightStandard = 24

	// TermHeightTall enables auto-open detail panel (40 rows).
	// This exceeds HeightThresholdTall (35) from model.go.
	TermHeightTall = 40

	// TermHeightMinimum is the smallest viable terminal height (10 rows).
	TermHeightMinimum = 10
)

// Terminal size presets for common test scenarios.
// Each preset is a [2]int where [0] is width and [1] is height.
var (
	// TermSizeStandard is the traditional 80x24 terminal.
	// Use this for baseline testing of standard terminal behavior.
	TermSizeStandard = [2]int{TermWidthStandard, TermHeightStandard}

	// TermSizeNarrow triggers narrow/mobile view behavior.
	// Use this to test vertical-only layout and narrow warnings.
	TermSizeNarrow = [2]int{TermWidthNarrow, TermHeightStandard}

	// TermSizeWide enables ultra-wide layout features.
	// Use this to test content capping and wide layout behavior.
	TermSizeWide = [2]int{TermWidthWide, TermHeightStandard}

	// TermSizeTall enables auto-open detail panel.
	// Use this to test height-based layout changes.
	TermSizeTall = [2]int{TermWidthStandard, TermHeightTall}

	// TermSizeMinimum is the smallest viable terminal.
	// Use this to test graceful degradation.
	TermSizeMinimum = [2]int{TermWidthMinimum, TermHeightMinimum}

	// TermSizeUltraWide goes beyond typical max_content_width (200x30).
	// Uses height 30 as a middle ground between standard (24) and tall (40).
	// Use this to test behavior when terminal exceeds content limits.
	TermSizeUltraWide = [2]int{TermWidthUltraWide, 30}
)

// ============================================================================
// Teatest Configuration Types (AC1)
// ============================================================================

// teatestConfig holds configuration for NewTeatestModel.
type teatestConfig struct {
	width    int
	height   int
	repo     ports.ProjectRepository
	projects []*domain.Project
}

// defaultTeatestConfig returns sensible defaults for teatest setup.
// Default terminal size: 80x24
// Default timeout: 3 seconds (via teatest.WithFinalTimeout)
func defaultTeatestConfig() *teatestConfig {
	return &teatestConfig{
		width:    TermWidthStandard,
		height:   TermHeightStandard,
		repo:     nil, // Will use teatestMockRepository if nil
		projects: []*domain.Project{},
	}
}

// TeatestOption is a functional option for configuring NewTeatestModel.
type TeatestOption func(*teatestConfig)

// WithTermSize sets custom terminal dimensions for the test.
//
// Example:
//
//	tm := NewTeatestModel(t, WithTermSize(120, 30))
func WithTermSize(width, height int) TeatestOption {
	return func(c *teatestConfig) {
		c.width = width
		c.height = height
	}
}

// WithTermSizePreset uses a predefined terminal size preset.
// Available presets: TermSizeStandard, TermSizeNarrow, TermSizeWide,
// TermSizeTall, TermSizeMinimum, TermSizeUltraWide.
//
// Example:
//
//	tm := NewTeatestModel(t, WithTermSizePreset(TermSizeNarrow))
func WithTermSizePreset(preset [2]int) TeatestOption {
	return func(c *teatestConfig) {
		c.width = preset[0]
		c.height = preset[1]
	}
}

// WithRepository sets a custom repository implementation.
// If not set, a default mock repository is used.
//
// Example:
//
//	customRepo := &myMockRepository{...}
//	tm := NewTeatestModel(t, WithRepository(customRepo))
func WithRepository(repo ports.ProjectRepository) TeatestOption {
	return func(c *teatestConfig) {
		c.repo = repo
	}
}

// WithProjects sets the projects that the mock repository will return.
// This option is ignored if a custom repository is provided via WithRepository.
//
// Example:
//
//	projects := []*domain.Project{
//	    {ID: "1", Name: "test-project", Path: "/test/path"},
//	}
//	tm := NewTeatestModel(t, WithProjects(projects))
func WithProjects(projects []*domain.Project) TeatestOption {
	return func(c *teatestConfig) {
		c.projects = projects
	}
}

// ============================================================================
// Test Model Creation (AC1)
// ============================================================================

// NewTeatestModel creates a new teatest.TestModel with sensible defaults.
//
// Default configuration:
//   - Terminal size: 80x24 (standard)
//   - Color profile: ASCII (deterministic output)
//   - Mock repository: teatestMockRepository with empty projects
//   - Wait timeout: Use teatest.WithFinalTimeout(3*time.Second) when calling WaitFinished
//
// The returned TestModel should be used with:
//   - tm.Send() to send key events
//   - tm.WaitFinished() to wait for program termination
//   - tm.FinalOutput() to capture output after WaitFinished
//   - tm.FinalModel() to access final model state
//
// Example:
//
//	func TestMyFeature(t *testing.T) {
//	    tm := NewTeatestModel(t, WithTermSizePreset(TermSizeNarrow))
//
//	    // Send navigation key
//	    tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
//	    time.Sleep(50 * time.Millisecond)
//
//	    // Quit and capture output
//	    tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
//	    tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
//
//	    out, _ := io.ReadAll(tm.FinalOutput(t))
//	    // Assert on output...
//	}
func NewTeatestModel(t *testing.T, opts ...TeatestOption) *teatest.TestModel {
	t.Helper()

	// Force ASCII color profile for deterministic output in golden file tests.
	// This is called per-test rather than in init() to avoid affecting other tests.
	lipgloss.SetColorProfile(termenv.Ascii)

	cfg := defaultTeatestConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	// Use provided repository or create default mock
	repo := cfg.repo
	if repo == nil {
		repo = &teatestMockRepository{projects: cfg.projects}
	}

	// Create model with repository to avoid Init() panic
	// (See teatest_poc_test.go TestTeatest_BasicModelInitialization comments)
	m := NewModel(repo)

	// Create teatest model with configured terminal size
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(cfg.width, cfg.height))

	return tm
}

// ============================================================================
// Terminal Resize Simulation (AC5)
// ============================================================================

// ResizeTerminal sends a WindowSizeMsg to simulate terminal resize.
// A small delay is added after sending to allow the model to process the resize.
//
// This is useful for testing responsive layout behavior.
//
// Example:
//
//	tm := NewTeatestModel(t, WithTermSizePreset(TermSizeStandard))
//
//	// Resize to narrow view
//	ResizeTerminal(tm, TermWidthNarrow, TermHeightStandard)
//
//	// Verify narrow layout behavior...
func ResizeTerminal(tm *teatest.TestModel, width, height int) {
	tm.Send(tea.WindowSizeMsg{Width: width, Height: height})
	// Allow time for resize message to be processed.
	// See teatest_poc_test.go lines 120-124 for rationale.
	time.Sleep(50 * time.Millisecond)
}

// ============================================================================
// Golden File Utilities (AC3)
// ============================================================================

// GoldenFilePath returns the path to a golden file for the given category and test name.
// Golden files are stored in internal/adapters/tui/testdata/golden/{category}/{testName}.golden
//
// Available categories (created in Task 3):
//   - navigation: Navigation-related test outputs
//   - layout: Layout and responsive design outputs
//   - resize: Terminal resize behavior outputs
//
// Example:
//
//	path := GoldenFilePath("navigation", "TestNavigation_Basic")
//	// Returns: testdata/golden/navigation/TestNavigation_Basic.golden
func GoldenFilePath(category, testName string) string {
	return filepath.Join("testdata", "golden", category, testName+".golden")
}
