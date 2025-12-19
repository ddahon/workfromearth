package scraping

import (
	"fmt"
	"regexp"
)

type AshbyScraper struct {
	CompanyName string
	Url         string
}

type AshbyResponse struct {
	Jobs []AshbyJob `json:"jobs"`
}

type AshbyJob struct {
	Title          string             `json:"title"`
	Location       string             `json:"location"`
	IsRemote       bool               `json:"isRemote"`
	JobURL         string             `json:"jobUrl"`
	ApplyURL       string             `json:"applyUrl"`
	Description    string             `json:"descriptionHtml"`
	EmploymentType string             `json:"employmentType"`
	PublishedAt    string             `json:"publishedAt"`
	Compensation   *AshbyCompensation `json:"compensation,omitempty"`
}

type AshbyCompensation struct {
	ScrapeableCompensationSalarySummary string `json:"scrapeableCompensationSalarySummary"`
}

func (s AshbyScraper) Company() string {
	return s.CompanyName
}

func (s AshbyScraper) Scrape() ([]Job, error) {
	// https://developers.ashbyhq.com/docs/public-job-posting-api
	endpoint := "https://api.ashbyhq.com/posting-api/job-board/" + s.CompanyName + "?includeCompensation=true"

	var ashbyResp AshbyResponse
	if err := FetchJSON(endpoint, &ashbyResp); err != nil {
		return nil, err
	}

	jobs := make([]Job, 0, len(ashbyResp.Jobs))
	for _, ashbyJob := range ashbyResp.Jobs {
		salaryRange := ""
		if ashbyJob.Compensation != nil {
			salaryRange = ashbyJob.Compensation.ScrapeableCompensationSalarySummary
		}
		job := Job{
			Title:       ashbyJob.Title,
			Url:         ashbyJob.JobURL,
			Description: ashbyJob.Description,
			SalaryRange: salaryRange,
			PublishedAt: ashbyJob.PublishedAt,
		}
		jobs = append(jobs, job)
	}

	LogScrapeResult(s.Url, len(jobs))
	return jobs, nil
}

func parseAshbySource(url string) (Scraper, error) {
	re, err := regexp.Compile("jobs.ashbyhq.com/([^/]+)")
	if err != nil {
		return nil, fmt.Errorf("compiling regexp for ashby: %v", err)
	}
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		return nil, fmt.Errorf("cannot extract company from ashby URL: %v", url)
	}
	company := matches[1]
	return AshbyScraper{
		CompanyName: company,
		Url:         url,
	}, nil
}
