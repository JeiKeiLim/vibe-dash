package metrics

// transitionColumns lists all columns for stage_transitions table
const transitionColumns = `id, project_id, from_stage, to_stage, transitioned_at`

// insertTransitionSQL inserts a new stage transition record
const insertTransitionSQL = `
INSERT INTO stage_transitions (` + transitionColumns + `)
VALUES (?, ?, ?, ?, ?)`

// selectByProjectSQL retrieves all transitions for a project ordered by time descending
// Used in Story 16.2 for GetTransitionsByProject method
//
//nolint:unused // Reserved for Story 16.2 - Stage Transition Event Recording
const selectByProjectSQL = `
SELECT ` + transitionColumns + ` FROM stage_transitions
WHERE project_id = ? ORDER BY transitioned_at DESC`

// selectByTimeRangeSQL retrieves transitions within a time range ordered by time descending
// Used in Story 16.2+ for GetTransitionsByTimeRange method
//
//nolint:unused // Reserved for Story 16.2 - Stage Transition Event Recording
const selectByTimeRangeSQL = `
SELECT ` + transitionColumns + ` FROM stage_transitions
WHERE transitioned_at >= ? AND transitioned_at <= ?
ORDER BY transitioned_at DESC`
