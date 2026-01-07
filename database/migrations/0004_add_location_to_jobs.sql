-- SQLite migration: Add location column to jobs table

ALTER TABLE jobs ADD COLUMN location TEXT;

CREATE INDEX IF NOT EXISTS idx_jobs_location ON jobs(location);

