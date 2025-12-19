package scraping

import (
	"fmt"
	"regexp"
)

type GreenhouseScraper struct {
	CompanyName string
	Url         string
}

type GreenhouseResponse struct {
	Jobs []GreenhouseJob `json:"jobs"`
}

type GreenhouseJob struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Location    *Location `json:"location"`
	Content     string    `json:"content"`
	UpdatedAt   string    `json:"updated_at"`
	AbsoluteURL string    `json:"absolute_url"`
}

type Location struct {
	Name string `json:"name"`
}

func (s GreenhouseScraper) Company() string {
	return s.CompanyName
}

func (s GreenhouseScraper) Scrape() ([]Job, error) {
	// https://boards-api.greenhouse.io/v1/boards/{company}/jobs
	endpoint := "https://boards-api.greenhouse.io/v1/boards/" + s.CompanyName + "/jobs"

	var greenhouseResp GreenhouseResponse
	if err := FetchJSON(endpoint, &greenhouseResp); err != nil {
		return nil, err
	}

	jobs := make([]Job, 0, len(greenhouseResp.Jobs))
	for _, greenhouseJob := range greenhouseResp.Jobs {
		job := Job{
			Title:       greenhouseJob.Title,
			Url:         greenhouseJob.AbsoluteURL,
			Description: greenhouseJob.Content,
			PublishedAt: greenhouseJob.UpdatedAt,
			SalaryRange: "",
		}
		jobs = append(jobs, job)
	}

	LogScrapeResult(s.Url, len(jobs))
	return jobs, nil
}

func parseGreenhouseSource(url string) (Scraper, error) {
	re, err := regexp.Compile(`(?:boards|job-boards)\.greenhouse\.io/([^/]+)`)
	if err != nil {
		return nil, fmt.Errorf("compiling regexp for greenhouse: %v", err)
	}
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		return nil, fmt.Errorf("cannot extract company from greenhouse URL: %v", url)
	}
	company := matches[1]
	return GreenhouseScraper{
		CompanyName: company,
		Url:         url,
	}, nil
}
