CREATE TABLE IF NOT EXISTS alert_history (
    id             TEXT PRIMARY KEY,
    rule_id        TEXT NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    container_id   TEXT NOT NULL,
    container_name TEXT NOT NULL,
    message        TEXT NOT NULL,
    sent_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
