package main

import (
	"aecheck-data-process/internal/db/postgres"
	"aecheck-data-process/internal/logic"
	"aecheck-data-process/internal/logic/batch"
	"aecheck-data-process/internal/logic/common"
	"aecheck-data-process/internal/logic/database"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	if logic.IsLocalEnv() {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Failed to load .env file: %v", err)
		}
	}
	common.InitLogger()

	// Initialize database with pgxpool
	pool, err := postgres.InitFromEnv()
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	// Create database service
	dbService := database.NewService(pool)

	wikiURL := "https://anothereden.wiki/w/Renri"
	dryrun := true

	batch.CompareCharacter(wikiURL, dbService)
	batch.UpdateCharacter(wikiURL, dryrun, dbService)
}
