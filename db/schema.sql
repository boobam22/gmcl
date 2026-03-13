CREATE TABLE IF NOT EXISTS t_versions (
    id TEXT PRIMARY KEY,
    type TEXT,
    time TEXT,
    release_time TEXT,
    url TEXT,
    is_installed BOOLEAN DEFAULT FALSE,
    last_started_at TEXT
);

CREATE TABLE IF NOT EXISTS t_files (
    path TEXT PRIMARY KEY,
    sha TEXT,
    size INTEGER,
    kind TEXT,
    ref_count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS t_version_files (
    version_id TEXT NOT NULL,
    path TEXT NOT NULL,
    PRIMARY KEY(version_id, path),
    FOREIGN KEY(version_id) REFERENCES t_versions(id) ON DELETE CASCADE,
    FOREIGN KEY(path) REFERENCES t_files(path) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_versions_installed ON t_versions(is_installed);
CREATE INDEX IF NOT EXISTS idx_files_kind ON t_files(kind);
