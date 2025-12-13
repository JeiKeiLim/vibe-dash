package sqlite

// SchemaVersion is the current schema version for migrations
const SchemaVersion = 1

// CreateSchemaVersionTableSQL creates the schema_version table for tracking migrations
const CreateSchemaVersionTableSQL = `
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL
);`

// CreateProjectsTableSQL creates the projects table with all required columns
const CreateProjectsTableSQL = `
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,
    display_name TEXT,
    detected_method TEXT,
    current_stage TEXT,
    confidence TEXT,
    detection_reasoning TEXT,
    is_favorite INTEGER DEFAULT 0,
    state TEXT DEFAULT 'active',
    notes TEXT,
    last_activity_at TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);`

// CreateIndexesSQL creates indexes for common queries
const CreateIndexPathSQL = `CREATE INDEX IF NOT EXISTS idx_projects_path ON projects(path);`
const CreateIndexStateSQL = `CREATE INDEX IF NOT EXISTS idx_projects_state ON projects(state);`
