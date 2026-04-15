package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"kisakay/server/internal/config"
	"kisakay/server/internal/views"
)

type legacyPayload struct {
	Hashes []string `json:"hashes"`
}

func main() {
	cfg := config.Load()

	var (
		sourcePath = flag.String("source", filepath.Join("server-data", "views.json"), "Path to the legacy views.json file")
		dbPath     = flag.String("db", cfg.ViewDatabasePath, "Path to the SQLite views database")
		username   = flag.String("username", "", "Profile username that should receive the imported views")
	)

	flag.Parse()

	if *username == "" && flag.NArg() > 0 {
		*username = flag.Arg(0)
	}

	normalizedUsername, err := views.NormalizeUsername(*username)
	if err != nil {
		log.Fatalf("invalid username: %v", err)
	}

	payload, err := readLegacyPayload(*sourcePath)
	if err != nil {
		log.Fatalf("unable to read legacy payload: %v", err)
	}

	store, err := views.NewStore(*dbPath)
	if err != nil {
		log.Fatalf("unable to open view database: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("unable to close view database: %v", err)
		}
	}()

	imported, total, err := store.ImportLegacy(context.Background(), normalizedUsername, payload.Hashes)
	if err != nil {
		log.Fatalf("unable to import legacy views: %v", err)
	}

	fmt.Printf(
		"Imported %d legacy views for %s into %s. Total views for %s: %d\n",
		imported,
		normalizedUsername,
		*dbPath,
		normalizedUsername,
		total,
	)
}

func readLegacyPayload(path string) (legacyPayload, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return legacyPayload{}, err
	}

	if len(data) == 0 {
		return legacyPayload{}, nil
	}

	var payload legacyPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return legacyPayload{}, err
	}

	return payload, nil
}
