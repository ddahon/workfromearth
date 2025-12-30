package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ddahon/workfromearth/cmd/server/views"
	"github.com/ddahon/workfromearth/internal/scraping"
	"github.com/ddahon/workfromearth/internal/storage"
	"github.com/spf13/viper"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please specify the config file path in the arguments")
	}
	getConfig(os.Args[1])

	dbPath := viper.GetString("dbPath")
	port := viper.GetString("port")

	db, err := storage.NewDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		searchQuery := r.URL.Query().Get("q")

		var jobs []scraping.Job
		var err error
		if searchQuery != "" {
			jobs, err = repo.SearchJobsByTitle(searchQuery)
		} else {
			jobs, err = repo.GetAllJobs()
		}

		if err != nil {
			log.Printf("Failed to retrieve jobs from DB: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if err := views.Index(jobs, searchQuery).Render(r.Context(), w); err != nil {
			log.Printf("Failed to respond to request: %v", err)
		}
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))
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
