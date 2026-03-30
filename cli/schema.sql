CREATE TABLE IF NOT EXISTS t_versions (
    tag TEXT PRIMARY KEY,
    type TEXT,
    url TEXT,
    time TEXT,
    release_time TEXT,
    is_installed INTEGER,
    last_started_at TEXT
);

CREATE TABLE IF NOT EXISTS t_files (
    path TEXT PRIMARY KEY,
    size INTEGER,
    name TEXT,
    sha1 TEXT
);

CREATE TABLE IF NOT EXISTS t_version_files (
    tag TEXT,
    path TEXT,
    PRIMARY KEY (tag, path)
);
