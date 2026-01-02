package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ddahon/workfromearth/internal/scraping"
	"github.com/ddahon/workfromearth/internal/storage"
)

func main() {
	nameFlag := flag.String("name", "", "Company name (required)")
	siteURLFlag := flag.String("siteurl", "", "Company website URL (required)")
	careersURLFlag := flag.String("careersurl", "", "Company careers page URL")
	atsTypeFlag := flag.String("atstype", "", "ATS type (e.g., greenhouse, ashby, recruitee)")
	atsURLFlag := flag.String("atsurl", "", "ATS URL (required if atstype is provided)")
	slugFlag := flag.String("slug", "", "Company slug for auto-filling careersUrl and atsUrl")
	dbPathFlag := flag.String("db", "./db.sqlite", "Path to database file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Create a new company in the database.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *nameFlag == "" {
		log.Fatal("Error: -name is required")
	}
	if *siteURLFlag == "" {
		log.Fatal("Error: -siteurl is required")
	}

	careersURL := *careersURLFlag
	atsURL := *atsURLFlag

	if *slugFlag != "" {
		if *atsTypeFlag == "" {
			log.Fatal("Error: -atstype is required when -slug is provided")
		}

		switch *atsTypeFlag {
		case "ashby":
			if careersURL == "" {
				careersURL = fmt.Sprintf("https://jobs.ashbyhq.com/%s", *slugFlag)
			}
			if atsURL == "" {
				atsURL = fmt.Sprintf("https://api.ashbyhq.com/posting-api/job-board/%s?includeCompensation=true", *slugFlag)
			}
		case "greenhouse":
			if careersURL == "" {
				careersURL = fmt.Sprintf("https://job-boards.greenhouse.io/%s", *slugFlag)
			}
			if atsURL == "" {
				atsURL = fmt.Sprintf("https://boards-api.greenhouse.io/v1/boards/%s/jobs?content=true", *slugFlag)
			}
		}
	}

	if *atsTypeFlag != "" && atsURL == "" {
		log.Fatal("Error: -atsurl is required when -atstype is provided (or use -slug to auto-fill)")
	}

	db, err := storage.NewDB(*dbPathFlag)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)

	company := scraping.Company{
		Name:       *nameFlag,
		SiteURL:    *siteURLFlag,
		CareersURL: careersURL,
		ATSType:    *atsTypeFlag,
		ATSUrl:     atsURL,
	}

	id, err := repo.SaveCompany(company)
	if err != nil {
		log.Fatalf("Error saving company: %v", err)
	}

	fmt.Printf("Successfully created company: %s (ID: %d)\n", company.Name, id)
}
