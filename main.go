package main

import (
	"fmt"
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
	sources := scraping.ReadSources("./source_urls")
	fmt.Println(sources)
	for _, src := range sources {
		jobs, err := src.Scrape()
		if err != nil {
			log.Println(err)
			continue
		}
		err = repo.SaveJobs(jobs, src.Company())
		if err != nil {
			log.Printf("saving %v jobs: %v\n", len(jobs), err)
			continue
		}
	}
}
