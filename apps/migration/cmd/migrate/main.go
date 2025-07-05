package main

import (
	"log"

	"github.com/k8s-study/migration/internal/config"
	"github.com/k8s-study/migration/internal/migration"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	migrationService, err := migration.NewService(cfg.DatabaseURL, "migrations")
	if err != nil {
		log.Fatalf("Failed to create migration service: %v", err)
	}
	defer migrationService.Close()

	log.Println("Running database migrations...")
	if err := migrationService.Up(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	version, dirty, err := migrationService.Status()
	if err != nil {
		log.Printf("Warning: Failed to get migration status: %v", err)
	} else {
		log.Printf("Migration completed successfully. Current version: %d, Dirty: %v", version, dirty)
	}
}
