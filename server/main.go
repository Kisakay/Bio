package main

import (
	"log"
	"net/http"
	"time"

	"kisakay/server/internal/config"
	"kisakay/server/internal/lastfm"
	"kisakay/server/internal/views"
	"kisakay/server/internal/web"
)

func main() {
	cfg := config.Load()

	viewStore, err := views.NewStore(cfg.ViewStorePath)
	if err != nil {
		log.Fatalf("unable to initialize view store: %v", err)
	}

	lastfmClient := lastfm.NewClient(cfg.LastfmAPIKey, cfg.LastfmUser, &http.Client{
		Timeout: 10 * time.Second,
	})

	server := web.NewServer(cfg, lastfmClient, viewStore)
	addr := ":" + cfg.Port

	log.Printf("API listening on http://127.0.0.1%s", addr)

	if err := http.ListenAndServe(addr, server.Handler()); err != nil {
		log.Fatal(err)
	}
}
