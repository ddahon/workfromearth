package storage

import (
	"fmt"

	"github.com/ddahon/workfromearth/internal/scraping"
	"github.com/google/uuid"
)

type Repository struct {
	db *DB
}

func NewRepository(db *DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveJob(job scraping.Job, company string) error {
	query := `
		INSERT INTO jobs (id, title, company, description, job_url, salary_range, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, datetime('now'))
		ON CONFLICT (job_url) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			salary_range = EXCLUDED.salary_range,
			updated_at = datetime('now')
	`

	id := uuid.New().String()
	_, err := r.db.Exec(query, id, job.Title, company, job.Description, job.Url, job.SalaryRange)
	if err != nil {
		return fmt.Errorf("saving job: %w", err)
	}

	return nil
}

func (r *Repository) SaveJobs(jobs []scraping.Job, company string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO jobs (id, title, company, description, job_url, salary_range, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, datetime('now'))
		ON CONFLICT (job_url) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			salary_range = EXCLUDED.salary_range,
			updated_at = datetime('now')
	`)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, job := range jobs {
		id := uuid.New().String()
		if _, err := stmt.Exec(id, job.Title, company, job.Description, job.Url, job.SalaryRange); err != nil {
			return fmt.Errorf("executing statement: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}
