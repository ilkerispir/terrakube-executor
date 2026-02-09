package main

import (
	"log"
	"os"

	"github.com/ilkerispir/terrakube-executor/internal/config"
	"github.com/ilkerispir/terrakube-executor/internal/core"
	"github.com/ilkerispir/terrakube-executor/internal/mode/batch"
	"github.com/ilkerispir/terrakube-executor/internal/mode/online"
	"github.com/ilkerispir/terrakube-executor/internal/status"
	"github.com/ilkerispir/terrakube-executor/internal/storage"
)

func main() {
	log.Println("Terrakube Executor Go - Starting...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	statusService := status.NewStatusService(cfg)
	storageService, err := storage.NewStorageService(cfg.StorageType)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	processor := core.NewJobProcessor(cfg, statusService, storageService)

	if cfg.Mode == "BATCH" {
		if cfg.EphemeralJobData == nil {
			log.Fatal("Batch mode selected but no job data provided")
		}
		batch.AdjustAndExecute(cfg.EphemeralJobData, processor)
	} else {
		// Default to Online
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		online.StartServer(port, processor)
	}
}
