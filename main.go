package main

import (
	"log"
	"stonks/internal/database"
	"stonks/internal/web"
)

func main() {
	// Initialize current database
	dbPath, err := database.GetCurrentDatabasePath()
	if err != nil {
		log.Fatalf("Failed to get current database path: %v", err)
	}
	
	_, err = database.NewDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Start web server
	server, err := web.NewServer()
	if err != nil {
		log.Fatalf("Failed to create web server: %v", err)
	}

	if err := server.Start("8080"); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}
