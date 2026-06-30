package main

import (
	"log"
	"os"

	"analyzer-service/internal/repository"
	"analyzer-service/internal/router"
)

func main() {
	// Initialize Crunchy Data Postgres database
	repository.InitDB()

	// Setup API router
	r := router.SetupRouter()

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Distinct from config-service on port 8080
	}

	log.Printf("Analyzer Service starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start analyzer-service: %v", err)
	}
}
