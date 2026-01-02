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
	if *atsTypeFlag != "" && *atsURLFlag == "" {
		log.Fatal("Error: -atsurl is required when -atstype is provided")
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
		CareersURL: *careersURLFlag,
		ATSType:    *atsTypeFlag,
		ATSUrl:     *atsURLFlag,
	}

	if err := repo.SaveCompany(company); err != nil {
		log.Fatalf("Error saving company: %v", err)
	}

	fmt.Printf("Successfully created company: %s (ID: %s)\n", company.Name, company.ID)
}
