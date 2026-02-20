package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	qrterminal "github.com/mdp/qrterminal/v3"

	"github.com/tafli/calling-parents/internal/activitylog"
	"github.com/tafli/calling-parents/internal/auth"
	"github.com/tafli/calling-parents/internal/children"
	"github.com/tafli/calling-parents/internal/config"
	"github.com/tafli/calling-parents/internal/message"
	"github.com/tafli/calling-parents/internal/network"
	"github.com/tafli/calling-parents/internal/version"
)

//go:embed all:web
var webFS embed.FS

func main() {
	log.Printf("calling-parents %s", version.Info())

	// Determine config file path: flag > default "config.toml".
	configPath := "config.toml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, result, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if result.Created {
		log.Printf("Created default config file: %s (edit it to customize settings)", configPath)
	}
	if len(result.MergedKeys) > 0 {
		log.Printf("Config updated: added new keys %v (backup: %s)", result.MergedKeys, result.BackupPath)
	}

	// Resolve auth token: use env var or generate a random one.
	token := cfg.AuthToken
	if token == "" {
		var err error
		token, err = auth.GenerateToken()
		if err != nil {
			log.Fatalf("failed to generate auth token: %v", err)
		}
		log.Println("Generated random auth token (set auth_token in config.toml to use a fixed one)")
	}

	lanURL := network.LanURL(cfg.ListenAddr) + "#token=" + token

	log.Printf("ProPresenter API: %s", cfg.ProPresenterURL())
	log.Printf("Message template: %s", cfg.MessageName)
	if cfg.AutoClearSeconds > 0 {
		log.Printf("Auto-clear after %d seconds", cfg.AutoClearSeconds)
	} else {
		log.Println("Auto-clear disabled")
	}
	log.Printf("Listening on %s", cfg.ListenAddr)

	fmt.Println()
	fmt.Println("Open this URL on the phone:")
	fmt.Println(lanURL)
	fmt.Println()
	qrterminal.GenerateWithConfig(lanURL, qrterminal.Config{
		Level:          qrterminal.L,
		Writer:         os.Stdout,
		HalfBlocks:     true,
		BlackChar:      qrterminal.BLACK_BLACK,
		WhiteBlackChar: qrterminal.WHITE_BLACK,
		WhiteChar:      qrterminal.WHITE_WHITE,
		BlackWhiteChar: qrterminal.BLACK_WHITE,
		QuietZone:      1,
	})
	fmt.Println()

	// Activity logger (optional)
	var logger *activitylog.Logger
	if cfg.ActivityLog != "" {
		var err error
		logger, err = activitylog.New(cfg.ActivityLog)
		if err != nil {
			log.Fatalf("failed to open activity log: %v", err)
		}
		defer logger.Close()
		log.Printf("Activity log: %s", cfg.ActivityLog)
	}

	// Children store
	childStore, err := children.NewStore(cfg.ChildrenFile)
	if err != nil {
		log.Fatalf("failed to load children: %v", err)
	}
	log.Printf("Loaded %d children from %s", len(childStore.Names()), cfg.ChildrenFile)

	mux := http.NewServeMux()

	// Version endpoint (no auth required)
	mux.HandleFunc("/version", version.HandleVersion())

	// Children endpoint
	mux.Handle("/children", childStore)

	// Message endpoints: send, clear, test connection
	msgHandler := message.New(cfg.ProPresenterURL(), cfg.MessageName, cfg.AutoClearSeconds, logger)
	mux.HandleFunc("/message/send", msgHandler.HandleSend)
	mux.HandleFunc("/message/clear", msgHandler.HandleClear)
	mux.HandleFunc("/message/test", msgHandler.HandleTest)
	mux.HandleFunc("/message/config", msgHandler.HandleConfig)

	// Static PWA files
	webContent, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("failed to create sub filesystem: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(webContent)))

	// Wrap mux with auth middleware: protect /message/ and /children.
	protectedPrefixes := []string{"/message/", "/children"}
	handler := auth.Middleware(token, protectedPrefixes)(mux)

	if err := http.ListenAndServe(cfg.ListenAddr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
