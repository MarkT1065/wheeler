package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"stonks/internal/web"
	"syscall"
	"time"
)

func main() {
	// Create web server (it will handle database initialization)
	server, err := web.NewServer()
	if err != nil {
		log.Fatalf("Failed to create web server: %v", err)
	}

	// Setup routes
	server.SetupTestRoutes()

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: http.DefaultServeMux,
	}

	// Start server in background goroutine
	go func() {
		log.Printf("🚀 Wheeler web application starting on http://localhost:8080")
		log.Printf("   📈 Dashboard:    http://localhost:8080/")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start web server: %v", err)
		}
	}()

	// Create stop channel and wait for shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a shutdown signal
	<-stop
	log.Println("Shutdown signal received, gracefully shutting down...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server gracefully shut down")
	}
	
	// Close database connection
	if err := server.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	} else {
		log.Println("Database connection closed")
	}
}
