package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kisakay/server/internal/config"
	"kisakay/server/internal/lastfm"
	"kisakay/server/internal/views"
	"kisakay/server/internal/web"
)

func main() {
	cfg := config.Load()

	viewStore, err := views.NewStore(cfg.ViewDatabasePath)
	if err != nil {
		log.Fatalf("unable to initialize view store: %v", err)
	}
	defer func() {
		if err := viewStore.Close(); err != nil {
			log.Printf("unable to flush view store: %v", err)
		}
	}()

	lastfmStore, err := lastfm.NewStore(cfg.ViewDatabasePath)
	if err != nil {
		log.Fatalf("unable to initialize lastfm store: %v", err)
	}
	defer func() {
		if err := lastfmStore.Close(); err != nil {
			log.Printf("unable to flush lastfm store: %v", err)
		}
	}()

	lastfmClient := lastfm.NewClient(cfg.LastfmAPIKey, cfg.LastfmUser, &http.Client{
		Timeout: 10 * time.Second,
	})
	lastfmService := lastfm.NewService(lastfmClient, lastfmStore, lastfm.DefaultCacheTTL)

	server := web.NewServer(cfg, lastfmService, viewStore)
	addr := ":" + cfg.Port
	httpServer := &http.Server{
		Addr:    addr,
		Handler: server.Handler(),
	}

	log.Printf("API listening on http://127.0.0.1%s", addr)

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-shutdownCtx.Done():
		stop()
		drainCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(drainCtx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}

		if err := <-serverErrCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}
}
