ALTER TABLE jobs ADD COLUMN company_id TEXT;

UPDATE jobs 
SET company_id = (
    SELECT id 
    FROM companies 
    WHERE companies.name = jobs.company 
    LIMIT 1
);

CREATE INDEX IF NOT EXISTS idx_jobs_company_id ON jobs(company_id);
