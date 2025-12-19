-- SQLite migration: Create jobs table
-- Matches the Job struct: Url, Description, Title, SalaryRange
-- Also includes company (passed separately) and metadata fields

CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    company TEXT NOT NULL,
    description TEXT,
    job_url TEXT NOT NULL UNIQUE,
    salary_range TEXT,
    published_at TEXT,
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_jobs_company ON jobs(company);
CREATE INDEX IF NOT EXISTS idx_jobs_job_url ON jobs(job_url);
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at DESC);
