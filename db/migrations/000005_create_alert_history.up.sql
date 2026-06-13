CREATE TABLE IF NOT EXISTS alert_history (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rule_id        UUID NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    container_id   VARCHAR(255) NOT NULL,
    container_name VARCHAR(255) NOT NULL,
    message        TEXT NOT NULL,
    sent_at        TIMESTAMP NOT NULL DEFAULT NOW()
);
