package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

type AtsKind int16

const (
	Unknown AtsKind = iota
	Ashby
)

type Source struct {
	Url string
	Ats AtsKind
}

func main() {
	sources := readSources("./source_urls")
	fmt.Println(sources)
	// TODO: for ashby, get endpoint https://developers.ashbyhq.com/docs/public-job-posting-api
}

func readSources(path string) []Source {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("opening file %v: %v", path, err)
	}
	s := bufio.NewScanner(f)
	var res []Source
	for s.Scan() {
		l := s.Text()
		source := Source{
			Url: l,
			Ats: parseAtsFromUrl(l),
		}
		res = append(res, source)
	}

	return res
}

func parseAtsFromUrl(url string) AtsKind {
	if strings.Contains(url, "ashbyhq") {
		return Ashby
	}
	return Unknown
}
