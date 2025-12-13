package sqlite

// projectColumns lists all columns for SELECT queries (DRY)
const projectColumns = `id, name, path, display_name, detected_method, current_stage,
       confidence, detection_reasoning, is_favorite, state, notes,
       last_activity_at, created_at, updated_at`

// insertOrReplaceProjectSQL upserts a project by ID
const insertOrReplaceProjectSQL = `
INSERT OR REPLACE INTO projects (` + projectColumns + `)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

// selectByIDSQL retrieves a project by its unique identifier
const selectByIDSQL = `SELECT ` + projectColumns + ` FROM projects WHERE id = ?`

// selectByPathSQL retrieves a project by its canonical path
const selectByPathSQL = `SELECT ` + projectColumns + ` FROM projects WHERE path = ?`

// selectAllSQL retrieves all projects ordered by name
const selectAllSQL = `SELECT ` + projectColumns + ` FROM projects ORDER BY name`

// selectActiveSQL retrieves only active projects ordered by name
const selectActiveSQL = `SELECT ` + projectColumns + ` FROM projects WHERE state = 'active' ORDER BY name`

// selectHibernatedSQL retrieves only hibernated projects ordered by name
const selectHibernatedSQL = `SELECT ` + projectColumns + ` FROM projects WHERE state = 'hibernated' ORDER BY name`

// deleteByIDSQL removes a project by its unique identifier
const deleteByIDSQL = `DELETE FROM projects WHERE id = ?`

// updateStateSQL updates only the state and updated_at fields
const updateStateSQL = `UPDATE projects SET state = ?, updated_at = ? WHERE id = ?`
