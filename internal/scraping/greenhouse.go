package scraping

type GreenhouseScraper struct {
	Url string
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

func NewGreenhouseScraper(atsURL string) GreenhouseScraper {
	return GreenhouseScraper{
		Url: atsURL,
	}
}

func (s GreenhouseScraper) Scrape() ([]Job, error) {
	// ats_url should contain the full JSON API endpoint
	var greenhouseResp GreenhouseResponse
	if err := FetchJSON(s.Url, &greenhouseResp); err != nil {
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
