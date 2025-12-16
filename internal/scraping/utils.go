package scraping

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

type Source interface {
	Scrape() ([]Job, error)
	Company() string
}

func ReadSources(path string) []Source {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("opening file %v: %v", path, err)
	}
	s := bufio.NewScanner(f)
	var res []Source
	for s.Scan() {
		url := s.Text()
		source, err := parseSource(url)
		if err != nil {
			log.Printf("failed to parse source %v: %v\n", url, err)
		}
		res = append(res, source)
	}

	return res
}

func parseSource(url string) (Source, error) {
	if strings.Contains(url, "ashbyhq") {
		re, err := regexp.Compile("jobs.ashbyhq.com/([^/]+)")
		if err != nil {
			return nil, fmt.Errorf("compiling regexp for ashby: %v", err)
		}
		matches := re.FindStringSubmatch(url)
		if len(matches) < 2 {
			return nil, fmt.Errorf("cannot extract company from ashby URL: %v", url)
		}
		company := matches[1]
		return AshbySource{
			CompanyName: company,
			Url:         url,
		}, nil
	}
	return UnknownSource{url}, nil
}
