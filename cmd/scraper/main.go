package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ddahon/workfromearth/internal/scraping"
	"github.com/ddahon/workfromearth/internal/storage"
	"github.com/spf13/viper"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please specify the config file path in the arguments")
	}
	getConfig(os.Args[1])

	urlFlag := flag.String("url", "", "URL to scrape (searches in database by careers_url or ats_url)")
	lastScrapedFlag := flag.Int("last_scraped", 6, "Minimum number of hours since last scrape to rescrape a company (0 = always scrape)")
	flag.Parse()

	dbPath := viper.GetString("dbPath")

	db, err := storage.NewDB(dbPath)
	if err != nil {
		log.Fatalf("opening db %v: ", err)
	}
	defer db.Close()
	repo := storage.NewRepository(db)

	// If URL is provided, scrape only that URL and print results
	if *urlFlag != "" {
		company, err := repo.GetCompanyByURL(*urlFlag)
		if err != nil {
			log.Fatalf("getting company: %v", err)
		}

		scraper, err := scraping.CompanyToScraper(*company)
		if err != nil {
			log.Fatalf("creating scraper: %v", err)
		}

		jobs, err := scraper.Scrape()
		if err != nil {
			log.Fatalf("scraping: %v", err)
		}

		// Print results as JSON
		output := map[string]interface{}{
			"company":    company.Name,
			"url":        *urlFlag,
			"ats_type":   company.ATSType,
			"jobs_count": len(jobs),
			"jobs":       jobs,
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			log.Fatalf("encoding output: %v", err)
		}

		return
	}

	// Default behavior: scrape all companies
	companies, err := repo.GetCompanies()
	if err != nil {
		log.Fatalf("getting companies: %v", err)
	}

	var companiesToScrape []scraping.Company
	now := time.Now()
	minHoursSinceScrape := time.Duration(*lastScrapedFlag) * time.Hour

	for _, company := range companies {
		shouldScrape := false

		if company.ScrapedAt == "" {
			shouldScrape = true
		} else {
			// Parse scraped_at timestamp
			formats := []string{
				"2006-01-02 15:04:05.000",
				"2006-01-02 15:04:05",
				time.RFC3339,
			}
			var scrapedAt time.Time
			parsed := false
			for _, format := range formats {
				if t, err := time.Parse(format, company.ScrapedAt); err == nil {
					scrapedAt = t
					parsed = true
					break
				}
			}

			if !parsed {
				shouldScrape = true
			} else {
				hoursSinceScrape := now.Sub(scrapedAt)
				if hoursSinceScrape >= minHoursSinceScrape {
					shouldScrape = true
				}
			}
		}

		if shouldScrape {
			companiesToScrape = append(companiesToScrape, company)
		}
	}

	log.Printf("Found %d companies, %d need scraping\n", len(companies), len(companiesToScrape))

	for _, company := range companiesToScrape {
		scraper, err := scraping.CompanyToScraper(company)
		if err != nil {
			log.Printf("creating scraper for %s: %v\n", company.Name, err)
			continue
		}

		jobs, err := scraper.Scrape()
		if err != nil {
			log.Printf("scraping %s: %v\n", company.Name, err)
			continue
		}

		err = repo.SaveJobs(jobs, company.ID)
		if err != nil {
			log.Printf("saving %v jobs for %s: %v\n", len(jobs), company.Name, err)
			continue
		}

		if err := repo.UpdateScrapedAt(company.ID); err != nil {
			log.Printf("updating scraped_at for %s: %v\n", company.Name, err)
		}
	}
}

func getConfig(path string) {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("config file not found in %v: %w", path, err))
		} else {
			panic(fmt.Errorf("error while reading config file: %w", err))
		}
	}
}
