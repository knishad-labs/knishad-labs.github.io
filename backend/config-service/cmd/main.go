package main

import (
	"log"
	"os"

	"config-service/internal/repository"
	"config-service/internal/router"
)

func main() {
	// Initialize Crunchy Data Postgres database & GORM migration
	repository.InitDB()

	// Setup router
	r := router.SetupRouter()

	// Determine port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Config Service on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
