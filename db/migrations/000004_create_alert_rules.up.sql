CREATE TYPE alert_condition AS ENUM ('cpu_high', 'mem_high', 'container_down', 'oom_kill');

CREATE TABLE IF NOT EXISTS alert_rules (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(255) NOT NULL,
    condition  alert_condition NOT NULL,
    threshold  NUMERIC(10, 2),
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
