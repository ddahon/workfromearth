SHELL = /bin/bash

.PHONY: scrape server server-watch templ-generate help

help:
	@echo "Available targets:"
	@echo "  scraper         - Build the scraper binary"
	@echo "  server         - Generate templ files and build the server binary"
	@echo "  server-watch   - Watch for templ changes and run server with hot reload"
	@echo "  templ-generate - Generate Go code from templ templates"

scraper:
	@go build -o bin/scraper ./cmd/scraper

server: templ-generate
	@CGO_ENABLED=0 GOOS=linux go build -o bin/server ./cmd/server

server-watch:
	@templ generate --watch --proxy="http://localhost:8080" --cmd="go run cmd/server/main.go ./server.config.yml" ./cmd/server/views/...

templ-generate:
	@templ generate ./cmd/server/views/...

