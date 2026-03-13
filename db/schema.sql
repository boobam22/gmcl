CREATE TABLE IF NOT EXISTS t_versions (
    id TEXT PRIMARY KEY,
    type TEXT,
    time TEXT,
    release_time TEXT,
    url TEXT,
    is_installed BOOLEAN
);

CREATE TABLE IF NOT EXISTS t_assets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sha TEXT,
    size INTEGER,
    path TEXT
);

CREATE TABLE IF NOT EXISTS t_libraries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sha TEXT,
    size INTEGER,
    path TEXT
);

CREATE TABLE IF NOT EXISTS t_natives (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sha TEXT,
    size INTEGER,
    path TEXT,
    type TEXT
);
