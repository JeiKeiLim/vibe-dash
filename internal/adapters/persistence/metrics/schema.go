package metrics

// SchemaVersion is the current schema version for metrics database migrations
const SchemaVersion = 1

// CreateSchemaVersionTableSQL creates the schema_version table for tracking migrations
const CreateSchemaVersionTableSQL = `
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL
);`

// CreateStageTransitionsTableSQL creates the stage_transitions table for recording stage changes
const CreateStageTransitionsTableSQL = `
CREATE TABLE IF NOT EXISTS stage_transitions (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    from_stage TEXT NOT NULL,
    to_stage TEXT NOT NULL,
    transitioned_at TEXT NOT NULL
);`

// CreateIndexProjectSQL creates an index on project_id for efficient lookups
const CreateIndexProjectSQL = `CREATE INDEX IF NOT EXISTS idx_stage_transitions_project ON stage_transitions(project_id);`

// CreateIndexTimeSQL creates an index on transitioned_at for time-range queries
const CreateIndexTimeSQL = `CREATE INDEX IF NOT EXISTS idx_stage_transitions_time ON stage_transitions(transitioned_at);`
