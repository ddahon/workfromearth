package main

import (
	"log"

	"github.com/ddahon/workfromearth/internal/scraping"
	"github.com/ddahon/workfromearth/internal/storage"
)

func main() {
	db, err := storage.NewDB("./db.sqlite")
	if err != nil {
		log.Fatalf("opening db %v: ", err)
	}
	defer db.Close()
	repo := storage.NewRepository(db)

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

		err = repo.SaveJobs(jobs, scraper.Company())
		if err != nil {
			log.Printf("saving %v jobs for %s: %v\n", len(jobs), company.Name, err)
			continue
		}

		if err := repo.UpdateScrapedAt(company.ID); err != nil {
			log.Printf("updating scraped_at for %s: %v\n", company.Name, err)
		}
	}
}
