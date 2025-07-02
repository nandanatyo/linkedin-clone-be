
-- +goose Up
-- +goose StatementBegin
CREATE TYPE job_type AS ENUM ('full_time', 'part_time', 'contract', 'internship');
CREATE TYPE experience_level AS ENUM ('entry', 'mid', 'senior', 'executive');

CREATE TABLE jobs (
                      id SERIAL PRIMARY KEY,
                      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                      title VARCHAR(200) NOT NULL,
                      company VARCHAR(100) NOT NULL,
                      location VARCHAR(100) NOT NULL,
                      description TEXT NOT NULL,
                      requirements TEXT,
                      job_type job_type NOT NULL,
                      experience_level experience_level NOT NULL,
                      salary_min INTEGER,
                      salary_max INTEGER,
                      is_active BOOLEAN DEFAULT TRUE,
                      application_count INTEGER DEFAULT 0,
                      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                      deleted_at TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_jobs_user_id ON jobs(user_id);
CREATE INDEX idx_jobs_job_type ON jobs(job_type);
CREATE INDEX idx_jobs_experience_level ON jobs(experience_level);
CREATE INDEX idx_jobs_location ON jobs(location);
CREATE INDEX idx_jobs_is_active ON jobs(is_active);
CREATE INDEX idx_jobs_created_at ON jobs(created_at);
CREATE INDEX idx_jobs_deleted_at ON jobs(deleted_at);

-- Full text search index for job search
CREATE INDEX idx_jobs_search ON jobs USING gin(to_tsvector('english', title || ' ' || company || ' ' || description));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS jobs;
DROP TYPE IF EXISTS experience_level;
DROP TYPE IF EXISTS job_type;
-- +goose StatementEnd
