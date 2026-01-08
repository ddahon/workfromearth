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
	atsTypeFlag := flag.String("atstype", "", "ATS type (e.g., greenhouse, ashby, recruitee, lever)")
	atsURLFlag := flag.String("atsurl", "", "ATS URL (required if atstype is provided)")
	slugFlag := flag.String("slug", "", "Company slug for auto-filling careersUrl and atsUrl")
	dbPathFlag := flag.String("db", "./db.sqlite", "Path to database file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [name] [siteurl] [atstype] [slug]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Create a new company in the database.\n\n")
		fmt.Fprintf(os.Stderr, "Positional arguments (optional, overrides flags):\n")
		fmt.Fprintf(os.Stderr, "  name      Company name\n")
		fmt.Fprintf(os.Stderr, "  siteurl   Company website URL\n")
		fmt.Fprintf(os.Stderr, "  atstype   ATS type (e.g., greenhouse, ashby, recruitee, lever)\n")
		fmt.Fprintf(os.Stderr, "  slug      Company slug for auto-filling careersUrl and atsUrl\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Check for positional arguments
	args := flag.Args()
	var name, siteURL, atsType, slug string

	if len(args) >= 1 {
		name = args[0]
	}
	if len(args) >= 2 {
		siteURL = args[1]
	}
	if len(args) >= 3 {
		atsType = args[2]
	}
	if len(args) >= 4 {
		slug = args[3]
	}

	if name == "" {
		name = *nameFlag
	}
	if siteURL == "" {
		siteURL = *siteURLFlag
	}
	if atsType == "" {
		atsType = *atsTypeFlag
	}
	if slug == "" {
		slug = *slugFlag
	}

	if name == "" {
		log.Fatal("Error: name is required (use -name flag or provide as first positional argument)")
	}
	if siteURL == "" {
		log.Fatal("Error: siteurl is required (use -siteurl flag or provide as second positional argument)")
	}

	careersURL := *careersURLFlag
	atsURL := *atsURLFlag

	if slug != "" {
		if atsType == "" {
			log.Fatal("Error: atstype is required when slug is provided")
		}

		switch atsType {
		case "ashby":
			if careersURL == "" {
				careersURL = fmt.Sprintf("https://jobs.ashbyhq.com/%s", slug)
			}
			if atsURL == "" {
				atsURL = fmt.Sprintf("https://api.ashbyhq.com/posting-api/job-board/%s?includeCompensation=true", slug)
			}
		case "greenhouse":
			if careersURL == "" {
				careersURL = fmt.Sprintf("https://job-boards.greenhouse.io/%s", slug)
			}
			if atsURL == "" {
				atsURL = fmt.Sprintf("https://boards-api.greenhouse.io/v1/boards/%s/jobs?content=true", slug)
			}
		case "lever":
			if careersURL == "" {
				careersURL = fmt.Sprintf("https://jobs.lever.co/%s", slug)
			}
			if atsURL == "" {
				atsURL = fmt.Sprintf("https://api.lever.co/v0/postings/%s?mode=json", slug)
			}
		}
	}

	if atsType != "" && atsURL == "" {
		log.Fatal("Error: atsurl is required when atstype is provided (or use slug to auto-fill)")
	}

	db, err := storage.NewDB(*dbPathFlag)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)

	company := scraping.Company{
		Name:       name,
		SiteURL:    siteURL,
		CareersURL: careersURL,
		ATSType:    atsType,
		ATSUrl:     atsURL,
	}

	id, err := repo.SaveCompany(company)
	if err != nil {
		log.Fatalf("Error saving company: %v", err)
	}

	fmt.Printf("Successfully created company: %s (ID: %d)\n", company.Name, id)
}
