package main

import (
	"log"
	"os"

	"scheduler-service/internal/cron"
	"scheduler-service/internal/repository"
	"scheduler-service/internal/router"
)

func main() {
	// Initialize Crunchy Data Postgres database
	repository.InitDB()

	// Start active cron runners
	cron.StartCronEngine()

	// Setup router
	r := router.SetupRouter()

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082" // Distinct from config-service (8080) and analyzer-service (8081)
	}

	log.Printf("Scheduler Service starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start scheduler-service: %v", err)
	}
}
