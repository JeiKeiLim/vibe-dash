package metrics

import (
	"crypto/rand"
	"fmt"
)

// stageTransitionRow is the database row representation for scanning.
// Used in Story 16.2+ for query methods returning transition data.
//
//nolint:unused // Reserved for Story 16.2 - query methods
type stageTransitionRow struct {
	ID             string `db:"id"`
	ProjectID      string `db:"project_id"`
	FromStage      string `db:"from_stage"`
	ToStage        string `db:"to_stage"`
	TransitionedAt string `db:"transitioned_at"`
}

// generateUUID creates a UUID v4 using crypto/rand (no external dependencies)
func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	// Set version (4) and variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
