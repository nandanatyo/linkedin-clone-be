
-- +goose Up
-- +goose StatementBegin
CREATE TABLE comments (
                          id SERIAL PRIMARY KEY,
                          user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                          post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
                          content TEXT NOT NULL,
                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                          deleted_at TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_comments_user_id ON comments(user_id);
CREATE INDEX idx_comments_post_id ON comments(post_id);
CREATE INDEX idx_comments_created_at ON comments(created_at);
CREATE INDEX idx_comments_deleted_at ON comments(deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS comments;
-- +goose StatementEnd
