package scraping

import (
	"fmt"
)

type Scraper interface {
	Scrape() ([]Job, error)
}

type UnknownScraper struct {
	CompanyName string
	Url         string
}

func (s UnknownScraper) Scrape() ([]Job, error) {
	return []Job{}, nil
}

func CompanyToScraper(company Company) (Scraper, error) {
	switch company.ATSType {
	case "ashby":
		return NewAshbyScraper(company.ATSUrl), nil
	case "greenhouse":
		return NewGreenhouseScraper(company.ATSUrl), nil
	case "recruitee":
		return NewRecruiteeScraper(company.ATSUrl), nil
	case "custom", "unknown", "":
		return UnknownScraper{
			Url: company.ATSUrl,
		}, nil
	default:
		return nil, fmt.Errorf("unknown ATS type: %s", company.ATSType)
	}
}
