-- +goose Up
-- +goose StatementBegin
CREATE TYPE connection_status AS ENUM ('pending', 'accepted', 'blocked');

CREATE TABLE connections (
                             id SERIAL PRIMARY KEY,
                             requester_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                             addressee_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                             status connection_status DEFAULT 'pending',
                             requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                             accepted_at TIMESTAMP,
                             created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                             updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                             deleted_at TIMESTAMP,
                             UNIQUE(requester_id, addressee_id)
);

-- Create indexes
CREATE INDEX idx_connections_requester_id ON connections(requester_id);
CREATE INDEX idx_connections_addressee_id ON connections(addressee_id);
CREATE INDEX idx_connections_status ON connections(status);
CREATE INDEX idx_connections_requested_at ON connections(requested_at);
CREATE INDEX idx_connections_deleted_at ON connections(deleted_at);

-- Index for mutual connections queries
CREATE INDEX idx_connections_user_status ON connections(requester_id, addressee_id, status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS connections;
DROP TYPE IF EXISTS connection_status;
-- +goose StatementEnd