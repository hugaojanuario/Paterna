CREATE TABLE IF NOT EXISTS metrics_history (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    container_id   VARCHAR(255) NOT NULL,
    container_name VARCHAR(255) NOT NULL,
    cpu_percent    NUMERIC(5, 2) NOT NULL,
    memory_mb      NUMERIC(10, 2) NOT NULL,
    collected_at   TIMESTAMP NOT NULL DEFAULT NOW()
);
