package scraping

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func FetchJSON(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("getting %v: %w", url, err)
	}
	defer resp.Body.Close()

	if err := ValidateResponse(resp); err != nil {
		return err
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decoding JSON response: %w", err)
	}

	return nil
}

func ValidateResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received status %d scraping %v", resp.StatusCode, resp.Request.URL)
	}
	return nil
}

func LogScrapeResult(sourceURL string, jobCount int) {
	log.Printf("scraped %v jobs from %v", jobCount, sourceURL)
}
