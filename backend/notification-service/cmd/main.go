package main

import (
	"log"
	"os"

	"notification-service/internal/router"
)

func main() {
	// Setup router
	r := router.SetupRouter()

	// Determine port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083" // Distinct from config-service (8080), analyzer-service (8081), scheduler-service (8082)
	}

	log.Printf("Notification Service starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start notification-service: %v", err)
	}
}
