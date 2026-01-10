package test

import (
	"context"
	"log"
	"net/http"
	"os"
	"runtime"
	"stonks/internal/database"
	"stonks/internal/web"
	"testing"
	"time"
)

var testServer *http.Server

// TestMain is the entry point for all integration tests
func TestMain(m *testing.M) {
	log.Printf("Starting Wheeler integration tests using Go version %s", runtime.Version())

	// Change to project root directory for template loading
	if err := os.Chdir(".."); err != nil {
		log.Fatalf("Failed to change to project root: %v", err)
	}

	// Setup test database
	if err := setupTestDatabase(); err != nil {
		log.Fatalf("Failed to setup test database: %v", err)
	}

	// Start background server
	if err := startTestServer(); err != nil {
		log.Fatalf("Failed to start test server: %v", err)
	}

	// Wait for server to be ready
	waitForServer()

	// Run all tests
	exitCode := m.Run()

	// Cleanup
	stopTestServer()

	log.Printf("Wheeler integration tests completed")
	os.Exit(exitCode)
}

func setupTestDatabase() error {
	testDBPath := getTestDBPath()
	
	// Remove existing test database if it exists
	if err := os.Remove(testDBPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Note: Could not delete existing test database: %v", err)
	}
	
	// Create fresh test database
	if err := database.CreateNewDatabase(testDBName); err != nil {
		return err
	}
	
	// Set it as current
	return database.SetCurrentDatabase(testDBName)
}

func startTestServer() error {
	// Create web server
	server, err := web.NewServer()
	if err != nil {
		return err
	}

	// Setup routes
	server.SetupTestRoutes()

	// Create HTTP server on test port
	testServer = &http.Server{
		Addr:    ":8081",
		Handler: http.DefaultServeMux,
	}

	// Start server in background
	go func() {
		log.Printf("Starting test server on localhost:8081")
		if err := testServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Test server error: %v", err)
		}
	}()

	return nil
}

func waitForServer() {
	client := &http.Client{Timeout: 1 * time.Second}
	for i := 0; i < 30; i++ { // Wait up to 30 seconds
		resp, err := client.Get("http://localhost:8081/")
		if err == nil {
			resp.Body.Close()
			log.Printf("Test server is ready")
			return
		}
		time.Sleep(1 * time.Second)
	}
	log.Printf("Warning: Test server may not be ready")
}

func stopTestServer() {
	if testServer != nil {
		log.Printf("Shutting down test server")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		testServer.Shutdown(ctx)
	}
}