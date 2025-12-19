package scraping

import (
	"fmt"
)

type Scraper interface {
	Scrape() ([]Job, error)
	Company() string
}

type UnknownScraper struct {
	CompanyName string
	Url         string
}

func (s UnknownScraper) Company() string {
	return s.CompanyName
}

func (s UnknownScraper) Scrape() ([]Job, error) {
	return []Job{}, nil
}

func CompanyToScraper(company Company) (Scraper, error) {
	switch company.ATSType {
	case "ashby":
		return NewAshbyScraper(company.Name, company.ATSUrl), nil
	case "greenhouse":
		return NewGreenhouseScraper(company.Name, company.ATSUrl), nil
	case "unknown", "":
		return UnknownScraper{
			CompanyName: company.Name,
			Url:         company.ATSUrl,
		}, nil
	default:
		return nil, fmt.Errorf("unknown ATS type: %s", company.ATSType)
	}
}
