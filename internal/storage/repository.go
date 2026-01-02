package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ddahon/workfromearth/internal/scraping"
)

type Repository struct {
	db *DB
}

func NewRepository(db *DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveJob(job scraping.Job, companyID int64) error {
	query := `
		INSERT INTO jobs (id, title, company, company_id, description, job_url, salary_range, published_at, updated_at)
		VALUES ($1, $2, (SELECT name FROM companies WHERE id = $3), $3, $4, $5, $6, $7, datetime('now'))
		ON CONFLICT (job_url) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			salary_range = EXCLUDED.salary_range,
			published_at = EXCLUDED.published_at,
			company_id = EXCLUDED.company_id,
			updated_at = datetime('now')
	`

	id := uuid.New().String()
	_, err := r.db.Exec(
		query,
		id,
		job.Title,
		companyID,
		job.Description,
		job.Url,
		job.SalaryRange,
		job.PublishedAt,
	)
	if err != nil {
		return fmt.Errorf("saving job: %w", err)
	}

	return nil
}

func (r *Repository) SaveJobs(jobs []scraping.Job, companyID int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO jobs (id, title, company, company_id, description, job_url, salary_range, published_at, updated_at)
		VALUES ($1, $2, (SELECT name FROM companies WHERE id = $3), $3, $4, $5, $6, $7, datetime('now'))
		ON CONFLICT (job_url) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			salary_range = EXCLUDED.salary_range,
			published_at = EXCLUDED.published_at,
			company_id = EXCLUDED.company_id,
			updated_at = datetime('now')
	`)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, job := range jobs {
		id := uuid.New().String()
		if _, err := stmt.Exec(id, job.Title, companyID, job.Description, job.Url, job.SalaryRange, job.PublishedAt); err != nil {
			return fmt.Errorf("executing statement: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

func (r *Repository) GetCompanies() ([]scraping.Company, error) {
	query := `SELECT id, name, site_url, careers_url, ats_type, ats_url, scraped_at, created_at, updated_at FROM companies`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("querying companies: %w", err)
	}
	defer rows.Close()

	var companies []scraping.Company
	for rows.Next() {
		var c scraping.Company
		var siteURL, careersURL, atsType, atsURL, scrapedAt sql.NullString

		err := rows.Scan(
			&c.ID,
			&c.Name,
			&siteURL,
			&careersURL,
			&atsType,
			&atsURL,
			&scrapedAt,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning company: %w", err)
		}

		c.SiteURL = siteURL.String
		c.CareersURL = careersURL.String
		c.ATSType = atsType.String
		c.ATSUrl = atsURL.String
		c.ScrapedAt = scrapedAt.String

		companies = append(companies, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating companies: %w", err)
	}

	return companies, nil
}

func (r *Repository) SaveCompany(company scraping.Company) (int64, error) {
	var query string
	var result sql.Result
	var err error

	if company.ID == 0 {
		// Insert new company - let SQLite auto-generate the ID
		query = `
			INSERT INTO companies (name, site_url, careers_url, ats_type, ats_url, scraped_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, datetime('now'))
		`
		result, err = r.db.Exec(
			query,
			company.Name,
			company.SiteURL,
			company.CareersURL,
			company.ATSType,
			company.ATSUrl,
			company.ScrapedAt,
		)
		if err != nil {
			return 0, fmt.Errorf("saving company: %w", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("getting last insert id: %w", err)
		}
		return id, nil
	} else {
		// Update existing company
		query = `
			UPDATE companies SET
				name = $1,
				site_url = $2,
				careers_url = $3,
				ats_type = $4,
				ats_url = $5,
				scraped_at = $6,
				updated_at = datetime('now')
			WHERE id = $7
		`
		_, err = r.db.Exec(
			query,
			company.Name,
			company.SiteURL,
			company.CareersURL,
			company.ATSType,
			company.ATSUrl,
			company.ScrapedAt,
			company.ID,
		)
		if err != nil {
			return 0, fmt.Errorf("updating company: %w", err)
		}
		return company.ID, nil
	}
}

func (r *Repository) UpdateScrapedAt(companyID int64) error {
	query := `UPDATE companies SET scraped_at = datetime('now'), updated_at = datetime('now') WHERE id = $1`
	_, err := r.db.Exec(query, companyID)
	if err != nil {
		return fmt.Errorf("updating scraped_at: %w", err)
	}
	return nil
}

func (r *Repository) GetCompanyByURL(url string) (*scraping.Company, error) {
	query := `SELECT id, name, site_url, careers_url, ats_type, ats_url, scraped_at, created_at, updated_at FROM companies WHERE careers_url = $1 OR ats_url = $1`
	row := r.db.QueryRow(query, url)

	var c scraping.Company
	var siteURL, careersURL, atsType, atsURL, scrapedAt sql.NullString

	err := row.Scan(
		&c.ID,
		&c.Name,
		&siteURL,
		&careersURL,
		&atsType,
		&atsURL,
		&scrapedAt,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no company found with URL: %s", url)
		}
		return nil, fmt.Errorf("scanning company: %w", err)
	}

	c.SiteURL = siteURL.String
	c.CareersURL = careersURL.String
	c.ATSType = atsType.String
	c.ATSUrl = atsURL.String
	c.ScrapedAt = scrapedAt.String

	return &c, nil
}

func (r *Repository) GetAllJobs() ([]scraping.Job, error) {
	query := `
		SELECT 
			j.id, 
			j.title, 
			j.description, 
			j.job_url, 
			j.salary_range,
			j.published_at,
			j.created_at, 
			j.updated_at,
			c.id as company_id,
			c.name as company_name,
			c.site_url as company_site_url,
			c.careers_url as company_careers_url,
			c.ats_type as company_ats_type,
			c.ats_url as company_ats_url,
			c.scraped_at as company_scraped_at,
			c.created_at as company_created_at,
			c.updated_at as company_updated_at
		FROM jobs j
		LEFT JOIN companies c ON j.company_id = c.id
		ORDER BY j.published_at IS NULL, j.published_at DESC, j.created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("querying jobs: %w", err)
	}
	defer rows.Close()

	var jobs []scraping.Job
	for rows.Next() {
		var j scraping.Job
		var createdAt, updatedAt sql.NullString
		var companyID sql.NullInt64
		var companyName, companySiteURL, companyCareersURL, companyATSType, companyATSUrl sql.NullString
		var companyScrapedAt, companyCreatedAt, companyUpdatedAt sql.NullString

		err := rows.Scan(
			&j.ID,
			&j.Title,
			&j.Description,
			&j.Url,
			&j.SalaryRange,
			&j.PublishedAt,
			&createdAt,
			&updatedAt,
			&companyID,
			&companyName,
			&companySiteURL,
			&companyCareersURL,
			&companyATSType,
			&companyATSUrl,
			&companyScrapedAt,
			&companyCreatedAt,
			&companyUpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning job: %w", err)
		}

		if createdAt.Valid && createdAt.String != "" {
			formats := []string{
				"2006-01-02 15:04:05.000",
				"2006-01-02 15:04:05",
				time.RFC3339,
			}
			for _, format := range formats {
				if t, err := time.Parse(format, createdAt.String); err == nil {
					j.CreatedAt = t
					break
				}
			}
		}
		if updatedAt.Valid && updatedAt.String != "" {
			formats := []string{
				"2006-01-02 15:04:05.000",
				"2006-01-02 15:04:05",
				time.RFC3339,
			}
			for _, format := range formats {
				if t, err := time.Parse(format, updatedAt.String); err == nil {
					j.UpdatedAt = t
					break
				}
			}
		}

		if companyID.Valid && companyName.Valid {
			j.Company = &scraping.Company{
				ID:         companyID.Int64,
				Name:       companyName.String,
				SiteURL:    companySiteURL.String,
				CareersURL: companyCareersURL.String,
				ATSType:    companyATSType.String,
				ATSUrl:     companyATSUrl.String,
				ScrapedAt:  companyScrapedAt.String,
				CreatedAt:  companyCreatedAt.String,
				UpdatedAt:  companyUpdatedAt.String,
			}
		}

		jobs = append(jobs, j)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating jobs: %w", err)
	}

	return jobs, nil
}

func (r *Repository) SearchJobsByTitle(query string) ([]scraping.Job, error) {
	baseQuery := `
		SELECT 
			j.id, 
			j.title, 
			j.description, 
			j.job_url, 
			j.salary_range,
			j.published_at,
			j.created_at, 
			j.updated_at,
			c.id as company_id,
			c.name as company_name,
			c.site_url as company_site_url,
			c.careers_url as company_careers_url,
			c.ats_type as company_ats_type,
			c.ats_url as company_ats_url,
			c.scraped_at as company_scraped_at,
			c.created_at as company_created_at,
			c.updated_at as company_updated_at
		FROM jobs j
		LEFT JOIN companies c ON j.company_id = c.id
	`

	var rows *sql.Rows
	var err error

	if query == "" {
		rows, err = r.db.Query(baseQuery + " ORDER BY j.published_at IS NULL, j.published_at DESC, j.created_at DESC")
	} else {
		searchPattern := "%" + query + "%"
		rows, err = r.db.Query(baseQuery+" WHERE LOWER(j.title) LIKE LOWER(?) ORDER BY j.published_at IS NULL, j.published_at DESC, j.created_at DESC", searchPattern)
	}

	if err != nil {
		return nil, fmt.Errorf("querying jobs: %w", err)
	}
	defer rows.Close()

	var jobs []scraping.Job
	for rows.Next() {
		var j scraping.Job
		var createdAt, updatedAt sql.NullString
		var companyID sql.NullInt64
		var companyName, companySiteURL, companyCareersURL, companyATSType, companyATSUrl sql.NullString
		var companyScrapedAt, companyCreatedAt, companyUpdatedAt sql.NullString

		err := rows.Scan(
			&j.ID,
			&j.Title,
			&j.Description,
			&j.Url,
			&j.SalaryRange,
			&j.PublishedAt,
			&createdAt,
			&updatedAt,
			&companyID,
			&companyName,
			&companySiteURL,
			&companyCareersURL,
			&companyATSType,
			&companyATSUrl,
			&companyScrapedAt,
			&companyCreatedAt,
			&companyUpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning job: %w", err)
		}

		if createdAt.Valid && createdAt.String != "" {
			formats := []string{
				"2006-01-02 15:04:05.000",
				"2006-01-02 15:04:05",
				time.RFC3339,
			}
			for _, format := range formats {
				if t, err := time.Parse(format, createdAt.String); err == nil {
					j.CreatedAt = t
					break
				}
			}
		}
		if updatedAt.Valid && updatedAt.String != "" {
			formats := []string{
				"2006-01-02 15:04:05.000",
				"2006-01-02 15:04:05",
				time.RFC3339,
			}
			for _, format := range formats {
				if t, err := time.Parse(format, updatedAt.String); err == nil {
					j.UpdatedAt = t
					break
				}
			}
		}

		if companyID.Valid && companyName.Valid {
			j.Company = &scraping.Company{
				ID:         companyID.Int64,
				Name:       companyName.String,
				SiteURL:    companySiteURL.String,
				CareersURL: companyCareersURL.String,
				ATSType:    companyATSType.String,
				ATSUrl:     companyATSUrl.String,
				ScrapedAt:  companyScrapedAt.String,
				CreatedAt:  companyCreatedAt.String,
				UpdatedAt:  companyUpdatedAt.String,
			}
		}

		jobs = append(jobs, j)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating jobs: %w", err)
	}

	return jobs, nil
}
