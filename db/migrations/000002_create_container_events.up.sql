CREATE TYPE container_event_type AS ENUM ('started', 'stopped', 'oom_killed', 'restarted');

CREATE TABLE IF NOT EXISTS container_events (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    container_id   VARCHAR(255) NOT NULL,
    container_name VARCHAR(255) NOT NULL,
    event_type     container_event_type NOT NULL,
    occurred_at    TIMESTAMP NOT NULL,
    created_at     TIMESTAMP NOT NULL DEFAULT NOW()
);
