CREATE TABLE IF NOT EXISTS container_events (
    id             TEXT PRIMARY KEY,
    container_id   TEXT NOT NULL,
    container_name TEXT NOT NULL,
    event_type     TEXT NOT NULL CHECK (event_type IN ('started','stopped','oom_killed','restarted')),
    occurred_at    DATETIME NOT NULL,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
