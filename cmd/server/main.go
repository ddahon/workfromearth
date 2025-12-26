package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ddahon/workfromearth/cmd/server/views"
	"github.com/ddahon/workfromearth/internal/scraping"
	"github.com/ddahon/workfromearth/internal/storage"
	"github.com/spf13/viper"
)

func redirect(w http.ResponseWriter, req *http.Request) {
	// remove/add not default ports from req.Host
	host := strings.Split(req.Host, ":")[0]
	target := "https://" + host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	log.Printf("redirect to: %s", target)
	http.Redirect(w, req, target,
		// see comments below and consider the codes 308, 302, or 301
		http.StatusTemporaryRedirect)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please specify the config file path in the arguments")
	}
	getConfig(os.Args[1])

	dbPath := viper.GetString("dbPath")
	port := viper.GetString("port")
	sslEnabled := viper.GetBool("sslEnabled")

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

	if sslEnabled {
		certFile := viper.GetString("sslCertFile")
		keyFile := viper.GetString("sslKeyFile")
		go http.ListenAndServe(":8080", http.HandlerFunc(redirect))
		log.Fatal(http.ListenAndServeTLS(":"+port, certFile, keyFile, nil))
	} else {
		log.Fatal(http.ListenAndServe(":"+port, nil))
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
