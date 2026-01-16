package metrics

// transitionColumns lists all columns for stage_transitions table
const transitionColumns = `id, project_id, from_stage, to_stage, transitioned_at`

// insertTransitionSQL inserts a new stage transition record
const insertTransitionSQL = `
INSERT INTO stage_transitions (` + transitionColumns + `)
VALUES (?, ?, ?, ?, ?)`

// selectByProjectSQL retrieves all transitions for a project ordered by time descending
// Reserved for future use - consider removing if not needed by Story 16.6
//
//nolint:unused // Reserved for future stories
const selectByProjectSQL = `
SELECT ` + transitionColumns + ` FROM stage_transitions
WHERE project_id = ? ORDER BY transitioned_at DESC`

// selectByTimeRangeSQL retrieves transitions within a time range ordered by time descending
// Used in Story 16.4 for GetTransitionsByTimeRange method
const selectByTimeRangeSQL = `
SELECT ` + transitionColumns + ` FROM stage_transitions
WHERE transitioned_at >= ? AND transitioned_at <= ?
ORDER BY transitioned_at DESC`

// selectByProjectWithTimeSQL retrieves transitions for a project since a given time
// Used in Story 16.4 for activity sparklines
const selectByProjectWithTimeSQL = `
SELECT ` + transitionColumns + ` FROM stage_transitions
WHERE project_id = ? AND transitioned_at >= ?
ORDER BY transitioned_at ASC`
