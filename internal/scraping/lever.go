// https://hire.lever.co/developer/documentation#list-all-postings
package scraping

import (
	"fmt"
	"time"
)

type LeverScraper struct {
	Url string
}

type LeverResponse []LeverJob

type LeverJob struct {
	ID            string            `json:"id"`
	Text          string            `json:"text"`
	Categories    LeverCategories   `json:"categories"`
	Description   string            `json:"description"`
	HostedURL     string            `json:"hostedUrl"`
	ApplyURL      string            `json:"applyUrl"`
	WorkplaceType string            `json:"workplaceType"`
	SalaryRange   *LeverSalaryRange `json:"salaryRange,omitempty"`
}

type LeverCategories struct {
	Location string `json:"location"`
}

type LeverSalaryRange struct {
	Currency string   `json:"currency"`
	Interval string   `json:"interval"`
	Min      *float64 `json:"min"`
	Max      *float64 `json:"max"`
}

func NewLeverScraper(atsURL string) LeverScraper {
	return LeverScraper{
		Url: atsURL,
	}
}

func (s LeverScraper) Scrape() ([]Job, error) {
	var leverResp LeverResponse
	if err := FetchJSON(s.Url, &leverResp); err != nil {
		return nil, err
	}

	scrapeTime := time.Now().Format(time.RFC3339)
	jobs := make([]Job, 0, len(leverResp))
	for _, leverJob := range leverResp {
		if leverJob.WorkplaceType != "remote" {
			continue
		}

		salaryRange := ""
		if leverJob.SalaryRange != nil {
			parts := []string{}
			if leverJob.SalaryRange.Currency != "" {
				parts = append(parts, leverJob.SalaryRange.Currency)
			}

			var minStr, maxStr string
			if leverJob.SalaryRange.Min != nil {
				minStr = fmt.Sprintf("%.0f", *leverJob.SalaryRange.Min)
			}
			if leverJob.SalaryRange.Max != nil {
				maxStr = fmt.Sprintf("%.0f", *leverJob.SalaryRange.Max)
			}

			if minStr != "" && maxStr != "" {
				parts = append(parts, minStr+" - "+maxStr)
			} else if minStr != "" {
				parts = append(parts, minStr)
			} else if maxStr != "" {
				parts = append(parts, maxStr)
			}

			if leverJob.SalaryRange.Interval != "" {
				parts = append(parts, leverJob.SalaryRange.Interval)
			}

			if len(parts) > 0 {
				for i, part := range parts {
					if i > 0 {
						salaryRange += " "
					}
					salaryRange += part
				}
			}
		}

		job := Job{
			Title:       leverJob.Text,
			Url:         leverJob.HostedURL,
			Description: leverJob.Description,
			SalaryRange: salaryRange,
			Location:    leverJob.Categories.Location,
			PublishedAt: scrapeTime,
		}
		jobs = append(jobs, job)
	}

	LogScrapeResult(s.Url, len(jobs))
	return jobs, nil
}
