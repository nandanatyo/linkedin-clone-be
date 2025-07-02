-- +goose Up
-- +goose StatementBegin
-- Drop and recreate the table with correct nullable fields
DROP TABLE IF EXISTS sessions;

CREATE TABLE sessions (
                          id SERIAL PRIMARY KEY,
                          user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                          refresh_token TEXT NOT NULL UNIQUE,
                          token_hash VARCHAR(255) NOT NULL,
                          status session_status DEFAULT 'active',
                          user_agent TEXT NULL, -- Make explicitly nullable
                          ip_address INET NULL, -- Make explicitly nullable
                          expires_at TIMESTAMP NOT NULL,
                          last_used_at TIMESTAMP NULL,
                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                          deleted_at TIMESTAMP NULL
);

-- Create indexes
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_user_status ON sessions(user_id, status);
CREATE INDEX idx_sessions_deleted_at ON sessions(deleted_at);

-- Partial index for active sessions
CREATE INDEX idx_sessions_active ON sessions(user_id, status, expires_at)
    WHERE status = 'active';

-- Index for cleanup operations
CREATE INDEX idx_sessions_cleanup ON sessions(status, expires_at)
    WHERE status IN ('active', 'expired');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd