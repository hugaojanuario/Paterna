CREATE TABLE IF NOT EXISTS alert_rules (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    condition  TEXT NOT NULL CHECK (condition IN ('cpu_high','mem_high','container_down','oom_kill')),
    threshold  REAL,
    enabled    INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
