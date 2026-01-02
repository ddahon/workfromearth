package scraping

type AshbyScraper struct {
	Url string
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

func NewAshbyScraper(atsURL string) AshbyScraper {
	return AshbyScraper{
		Url: atsURL,
	}
}

func (s AshbyScraper) Scrape() ([]Job, error) {
	var ashbyResp AshbyResponse
	if err := FetchJSON(s.Url, &ashbyResp); err != nil {
		return nil, err
	}

	jobs := make([]Job, 0, len(ashbyResp.Jobs))
	for _, ashbyJob := range ashbyResp.Jobs {
		if !ashbyJob.IsRemote {
			continue
		}
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
