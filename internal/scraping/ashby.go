package scraping

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type AshbySource struct {
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

func (source AshbySource) Company() string {
	return source.CompanyName
}

func (source AshbySource) Scrape() ([]Job, error) {
	// https://developers.ashbyhq.com/docs/public-job-posting-api
	endpoint := "https://api.ashbyhq.com/posting-api/job-board/" + source.CompanyName + "?includeCompensation=true"
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("getting %v: %v", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status %d scraping %v", resp.StatusCode, endpoint)
	}

	var ashbyResp AshbyResponse
	if err := json.NewDecoder(resp.Body).Decode(&ashbyResp); err != nil {
		return nil, fmt.Errorf("decoding JSON response: %v", err)
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
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}
