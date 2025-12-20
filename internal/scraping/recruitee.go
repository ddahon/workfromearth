package scraping

import (
	"encoding/json"
	"fmt"
	"strings"
)

type RecruiteeScraper struct {
	Url string
}

type RecruiteeResponse struct {
	Offers []RecruiteeOffer `json:"offers"`
}

type RecruiteeOffer struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	CareersURL  string            `json:"careers_url"`
	Locations   RecruiteeLocation `json:"location,omitempty"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	Salary      RecruiteeSalary   `json:"salary,omitempty"`
}

type RecruiteeLocation struct {
	City string
}

// UnmarshalJSON handles both string and object formats for location
func (l *RecruiteeLocation) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		l.City = str
		return nil
	}

	var obj struct {
		City string `json:"city"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	l.City = obj.City
	return nil
}

type RecruiteeSalary string

func (s *RecruiteeSalary) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = RecruiteeSalary(str)
		return nil
	}

	var obj struct {
		Min      string `json:"min"`
		Max      string `json:"max"`
		Period   string `json:"period"`
		Currency string `json:"currency"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	var parts []string

	if obj.Currency != "" {
		parts = append(parts, obj.Currency)
	}

	if obj.Min != "" && obj.Max != "" {
		parts = append(parts, fmt.Sprintf("%s - %s", obj.Min, obj.Max))
	} else if obj.Min != "" {
		parts = append(parts, obj.Min)
	} else if obj.Max != "" {
		parts = append(parts, obj.Max)
	}

	if obj.Period != "" {
		parts = append(parts, "per", obj.Period)
	}

	if len(parts) > 0 {
		*s = RecruiteeSalary(strings.Join(parts, " "))
	} else {
		*s = ""
	}

	return nil
}

func NewRecruiteeScraper(atsURL string) RecruiteeScraper {
	return RecruiteeScraper{
		Url: atsURL,
	}
}

func (s RecruiteeScraper) Scrape() ([]Job, error) {
	var recruiteeResp RecruiteeResponse
	if err := FetchJSON(s.Url, &recruiteeResp); err != nil {
		return nil, err
	}

	jobs := make([]Job, 0, len(recruiteeResp.Offers))
	for _, recruiteeOffer := range recruiteeResp.Offers {
		publishedAt := recruiteeOffer.CreatedAt
		if publishedAt == "" {
			publishedAt = recruiteeOffer.UpdatedAt
		}

		job := Job{
			Title:       recruiteeOffer.Title,
			Url:         recruiteeOffer.CareersURL,
			Description: recruiteeOffer.Description,
			PublishedAt: publishedAt,
			SalaryRange: string(recruiteeOffer.Salary),
		}
		jobs = append(jobs, job)
	}

	LogScrapeResult(s.Url, len(jobs))
	return jobs, nil
}
