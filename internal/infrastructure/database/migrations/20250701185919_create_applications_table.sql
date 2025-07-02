
-- +goose Up
-- +goose StatementBegin
CREATE TYPE application_status AS ENUM ('pending', 'reviewed', 'accepted', 'rejected');

CREATE TABLE applications (
                              id SERIAL PRIMARY KEY,
                              user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                              job_id INTEGER NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
                              cover_letter TEXT,
                              resume_url TEXT,
                              status application_status DEFAULT 'pending',
                              applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                              created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                              updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                              deleted_at TIMESTAMP,
                              UNIQUE(user_id, job_id)
);

-- Create indexes
CREATE INDEX idx_applications_user_id ON applications(user_id);
CREATE INDEX idx_applications_job_id ON applications(job_id);
CREATE INDEX idx_applications_status ON applications(status);
CREATE INDEX idx_applications_applied_at ON applications(applied_at);
CREATE INDEX idx_applications_deleted_at ON applications(deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS applications;
DROP TYPE IF EXISTS application_status;
-- +goose StatementEnd