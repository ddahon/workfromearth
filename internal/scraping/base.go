package scraping

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type Scraper interface {
	Scrape() ([]Job, error)
	Company() string
}

type ScraperParser func(url string) (Scraper, error)

var scraperRegistry = map[string]ScraperParser{
	"ashbyhq":       parseAshbySource,
	"greenhouse.io": parseGreenhouseSource,
}

type UnknownScraper struct {
	Url string
}

func (s UnknownScraper) Company() string {
	return ""
}

func (s UnknownScraper) Scrape() ([]Job, error) {
	return []Job{}, nil
}

func ReadSources(path string) []Scraper {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("opening file %v: %v", path, err)
	}
	s := bufio.NewScanner(f)
	var res []Scraper
	for s.Scan() {
		url := s.Text()
		scraper, err := parseSource(url)
		if err != nil {
			log.Printf("failed to parse source %v: %v\n", url, err)
		}
		res = append(res, scraper)
	}

	return res
}

func parseSource(url string) (Scraper, error) {
	for pattern, parser := range scraperRegistry {
		if strings.Contains(url, pattern) {
			return parser(url)
		}
	}
	return UnknownScraper{url}, nil
}
