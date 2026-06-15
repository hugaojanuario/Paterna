CREATE TABLE IF NOT EXISTS metrics_history (
    id             TEXT PRIMARY KEY,
    container_id   TEXT NOT NULL,
    container_name TEXT NOT NULL,
    cpu_percent    REAL NOT NULL,
    memory_mb      REAL NOT NULL,
    collected_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
