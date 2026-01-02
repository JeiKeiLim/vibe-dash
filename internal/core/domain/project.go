package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// Project represents a development project being tracked
type Project struct {
	ID                 string       // Unique identifier (path hash, 16 hex chars)
	Name               string       // Derived from directory name
	Path               string       // Canonical absolute path
	DisplayName        string       // Optional user-set nickname (FR5)
	DetectedMethod     string       // "speckit", "bmad", "unknown"
	CurrentStage       Stage        // Current workflow stage
	Confidence         Confidence   // Detection confidence level (FR12)
	DetectionReasoning string       // Human-readable detection explanation (FR11, FR26)
	IsFavorite         bool         // Always visible regardless of activity (FR30)
	State              ProjectState // Active or Hibernated (FR28-33)
	Notes              string       // User notes/memo (FR21)
	PathMissing        bool         // True if path was inaccessible at launch (FR-validation)
	LastActivityAt     time.Time    // Last file change detected (FR34-38)
	HibernatedAt       *time.Time   // When project was hibernated (nil if active)
	CreatedAt          time.Time    // When project was added
	UpdatedAt          time.Time    // Last database update
}

// GenerateID creates a deterministic ID from canonical path
// Returns first 16 characters of the SHA-256 hex digest
func GenerateID(canonicalPath string) string {
	hash := sha256.Sum256([]byte(canonicalPath))
	return hex.EncodeToString(hash[:])[:16]
}

// NewProject creates a new Project with validation
// path must be absolute (starts with /), name is derived from path if empty
func NewProject(path, name string) (*Project, error) {
	// Path: REQUIRED, must be non-empty
	if path == "" {
		return nil, ErrPathNotAccessible
	}

	// Path: must be absolute (starts with /)
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("%w: path must be absolute", ErrPathNotAccessible)
	}

	// Name: derive from path if empty (use strings split, not filepath to avoid external deps)
	if name == "" {
		// Remove trailing slashes before splitting to handle "/path/to/project/" case
		cleanPath := strings.TrimRight(path, "/")
		if cleanPath == "" {
			// Root path "/" case - use "root" as default name
			name = "root"
		} else {
			parts := strings.Split(cleanPath, "/")
			name = parts[len(parts)-1]
		}
	}

	now := time.Now()
	return &Project{
		ID:             GenerateID(path),
		Name:           name,
		Path:           path,
		State:          StateActive,
		CreatedAt:      now,
		UpdatedAt:      now,
		LastActivityAt: now,
	}, nil
}

// IsHibernated returns true if the project is in hibernated state
func (p *Project) IsHibernated() bool {
	return p.State == StateHibernated
}

// IsActive returns true if the project is in active state
func (p *Project) IsActive() bool {
	return p.State == StateActive
}

// HasDisplayName returns true if the project has a custom display name set
func (p *Project) HasDisplayName() bool {
	return p.DisplayName != ""
}

// DaysSinceHibernated returns the number of complete days since the project was hibernated.
// Returns 0 if the project is not hibernated (HibernatedAt is nil).
// Note: Uses truncating integer division, so 23.9 hours returns 0 days, not 1.
func (p *Project) DaysSinceHibernated() int {
	if p.HibernatedAt == nil {
		return 0
	}
	duration := time.Since(*p.HibernatedAt)
	return int(duration.Hours() / 24)
}

// Validate checks Project invariants. Use after modification.
func (p *Project) Validate() error {
	if p.Path == "" {
		return ErrPathNotAccessible
	}
	if !strings.HasPrefix(p.Path, "/") {
		return fmt.Errorf("%w: path must be absolute", ErrPathNotAccessible)
	}
	if p.ID == "" {
		return fmt.Errorf("project ID cannot be empty")
	}
	return nil
}
