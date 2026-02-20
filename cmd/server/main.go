package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/calling-parents/calling-parents/internal/config"
	"github.com/calling-parents/calling-parents/internal/proxy"
)

//go:embed all:web
var webFS embed.FS

func main() {
	cfg := config.Load()

	log.Printf("ProPresenter API: %s", cfg.ProPresenterURL())
	log.Printf("Listening on %s", cfg.ListenAddr)

	mux := http.NewServeMux()

	// API proxy: /api/* -> ProPresenter
	apiProxy := proxy.New(cfg.ProPresenterURL())
	mux.Handle("/api/", apiProxy)

	// Static PWA files
	webContent, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("failed to create sub filesystem: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(webContent)))

	if err := http.ListenAndServe(cfg.ListenAddr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
