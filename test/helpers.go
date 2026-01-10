package test

import (
	"os"
	"testing"

	"stonks/internal/database"
)

const testDBName = "integration_test.db"

func getTestDBPath() string {
	return "./data/" + testDBName
}

func setupTestDB(t *testing.T) *database.DB {
	testDBPath := getTestDBPath()
	
	if err := os.Remove(testDBPath); err != nil && !os.IsNotExist(err) {
		t.Logf("Note: Could not delete existing test database: %v", err)
	}
	
	db, err := database.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	
	t.Cleanup(func() {
		db.Close()
	})
	
	return db
}

func getSharedTestDB(t *testing.T) *database.DB {
	testDBPath := getTestDBPath()
	
	db, err := database.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to open shared test database: %v", err)
	}
	
	return db
}
