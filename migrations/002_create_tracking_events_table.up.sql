CREATE TABLE IF NOT EXISTS tracking_events (
    id                 UUID         NOT NULL DEFAULT gen_random_uuid(),
    event              VARCHAR(100) NOT NULL,
    screen             VARCHAR(100) NOT NULL,
    user_id            BIGINT       NOT NULL,
    batch_id           BIGINT,
    properties         JSONB        DEFAULT '{}',
    meta_data          JSONB        DEFAULT '{}',
    occurred_at           TIMESTAMPTZ  NOT NULL,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, occurred_at)
);

-- Convert to hypertable with 14-day chunks
SELECT create_hypertable(
    'tracking_events',
    'occurred_at',
    chunk_time_interval => INTERVAL '14 days',
    if_not_exists => TRUE
);

-- Indexes
CREATE INDEX idx_tracking_events_event ON tracking_events (event);
CREATE INDEX idx_tracking_events_screen ON tracking_events (screen);
CREATE INDEX idx_tracking_events_user_id ON tracking_events (user_id);
CREATE INDEX idx_tracking_events_batch_id ON tracking_events (batch_id);
CREATE INDEX idx_tracking_events_occurred_at ON tracking_events (occurred_at);

-- Auto-delete chunks older than 6 months
SELECT add_retention_policy('tracking_events', INTERVAL '6 months', if_not_exists => TRUE);
