package scraping

import "time"

type Job struct {
	ID          string
	Url         string
	Description string
	Title       string
	SalaryRange string
	PublishedAt string
	Company     *Company
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
