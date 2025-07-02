
-- +goose Up
-- +goose StatementBegin
CREATE TABLE likes (
                       id SERIAL PRIMARY KEY,
                       user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                       post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                       deleted_at TIMESTAMP,
                       UNIQUE(user_id, post_id)
);

-- Create indexes
CREATE INDEX idx_likes_user_id ON likes(user_id);
CREATE INDEX idx_likes_post_id ON likes(post_id);
CREATE INDEX idx_likes_deleted_at ON likes(deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS likes;
-- +goose StatementEnd
