package metrics

import (
	"crypto/rand"
	"fmt"
	"time"
)

// stageTransitionRow is the database row representation for scanning.
type stageTransitionRow struct {
	ID             string `db:"id"`
	ProjectID      string `db:"project_id"`
	FromStage      string `db:"from_stage"`
	ToStage        string `db:"to_stage"`
	TransitionedAt string `db:"transitioned_at"`
}

// StageTransition is the public domain type for stage transition data.
// Distinct from internal stageTransitionRow - this is the exported API type.
type StageTransition struct {
	ID             string
	ProjectID      string
	FromStage      string
	ToStage        string
	TransitionedAt time.Time
}

// rowToTransition converts internal row to public domain type.
func rowToTransition(row *stageTransitionRow) StageTransition {
	t, _ := time.Parse(time.RFC3339Nano, row.TransitionedAt)
	return StageTransition{
		ID:             row.ID,
		ProjectID:      row.ProjectID,
		FromStage:      row.FromStage,
		ToStage:        row.ToStage,
		TransitionedAt: t,
	}
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
