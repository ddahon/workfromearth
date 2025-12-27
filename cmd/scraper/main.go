package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/ddahon/workfromearth/internal/scraping"
	"github.com/ddahon/workfromearth/internal/storage"
)

func main() {
	urlFlag := flag.String("url", "", "URL to scrape (searches in database by careers_url or ats_url)")
	flag.Parse()

	db, err := storage.NewDB("./db.sqlite")
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

	log.Printf("Found %d companies\n", len(companies))

	for _, company := range companies {
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
