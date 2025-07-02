-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       email VARCHAR(255) UNIQUE NOT NULL,
                       username VARCHAR(255) UNIQUE NOT NULL,
                       full_name VARCHAR(255) NOT NULL,
                       password VARCHAR(255) NOT NULL,
                       profile_picture TEXT,
                       bio TEXT,
                       location VARCHAR(255),
                       website VARCHAR(255),
                       is_verified BOOLEAN DEFAULT FALSE,
                       is_premium BOOLEAN DEFAULT FALSE,
                       premium_until TIMESTAMP,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                       deleted_at TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
