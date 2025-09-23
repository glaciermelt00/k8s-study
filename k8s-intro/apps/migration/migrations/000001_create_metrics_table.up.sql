CREATE TABLE
    IF NOT EXISTS metrics (
        id SERIAL PRIMARY KEY,
        workspace_id VARCHAR(255) NOT NULL,
        channel_id VARCHAR(255) NOT NULL,
        user_id VARCHAR(255) NOT NULL,
        message_count INTEGER NOT NULL DEFAULT 0,
        reaction_count INTEGER NOT NULL DEFAULT 0,
        thread_count INTEGER NOT NULL DEFAULT 0,
        recorded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

CREATE INDEX idx_metrics_workspace_id ON metrics (workspace_id);

CREATE INDEX idx_metrics_channel_id ON metrics (channel_id);

CREATE INDEX idx_metrics_user_id ON metrics (user_id);

CREATE INDEX idx_metrics_recorded_at ON metrics (recorded_at);
