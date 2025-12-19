CREATE TABLE IF NOT EXISTS companies (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    site_url TEXT,
    careers_url TEXT,
    ats_type TEXT,
    ats_url TEXT,
    scraped_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_companies_ats_type ON companies(ats_type);
CREATE INDEX IF NOT EXISTS idx_companies_name ON companies(name);

