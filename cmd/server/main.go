package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	qrterminal "github.com/mdp/qrterminal/v3"

	"github.com/calling-parents/calling-parents/internal/auth"
	"github.com/calling-parents/calling-parents/internal/children"
	"github.com/calling-parents/calling-parents/internal/config"
	"github.com/calling-parents/calling-parents/internal/network"
	"github.com/calling-parents/calling-parents/internal/proxy"
)

//go:embed all:web
var webFS embed.FS

func main() {
	cfg := config.Load()

	// Resolve auth token: use env var or generate a random one.
	token := cfg.AuthToken
	if token == "" {
		var err error
		token, err = auth.GenerateToken()
		if err != nil {
			log.Fatalf("failed to generate auth token: %v", err)
		}
		log.Println("Generated random auth token (set AUTH_TOKEN env to use a fixed one)")
	}

	lanURL := network.LanURL(cfg.ListenAddr) + "#token=" + token

	log.Printf("ProPresenter API: %s", cfg.ProPresenterURL())
	log.Printf("Listening on %s", cfg.ListenAddr)

	fmt.Println()
	fmt.Println("Open this URL on the phone:")
	fmt.Println(lanURL)
	fmt.Println()
	qrterminal.GenerateWithConfig(lanURL, qrterminal.Config{
		Level:     qrterminal.L,
		Writer:    os.Stdout,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
		QuietZone: 1,
	})
	fmt.Println()

	// Children store
	childStore, err := children.NewStore(cfg.ChildrenFile)
	if err != nil {
		log.Fatalf("failed to load children: %v", err)
	}
	log.Printf("Loaded %d children from %s", len(childStore.Names()), cfg.ChildrenFile)

	mux := http.NewServeMux()

	// Children endpoint
	mux.Handle("/children", childStore)

	// API proxy: /api/* -> ProPresenter
	apiProxy := proxy.New(cfg.ProPresenterURL())
	mux.Handle("/api/", apiProxy)

	// Static PWA files
	webContent, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("failed to create sub filesystem: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(webContent)))

	// Wrap mux with auth middleware: protect /api/ and /children.
	protectedPrefixes := []string{"/api/", "/children"}
	handler := auth.Middleware(token, protectedPrefixes)(mux)

	if err := http.ListenAndServe(cfg.ListenAddr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
