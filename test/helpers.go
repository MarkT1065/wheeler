package test

import (
	"strings"
	"testing"

	"stonks/internal/database"
)

const testDBName = database.IntegrationTestDatabaseName

func getTestDBPath() string {
	return database.IntegrationTestDatabasePath()
}

func setupTestDB(t *testing.T) *database.DB {
	testDBPath := getTestDBPath()
	db, err := database.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	resetTestDatabase(t, db)

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func resetTestDatabase(t *testing.T, db *database.DB) {
	t.Helper()

	tables := []string{
		"transactions_new",
		"trades",
		"option_history",
		"price_history",
		"index_history",
		"options",
		"long_positions",
		"dividends",
		"treasuries",
		"metrics",
		"symbols",
	}

	for _, table := range tables {
		if _, err := db.Exec("DELETE FROM " + table); err != nil && !strings.Contains(err.Error(), "no such table") {
			t.Fatalf("Failed to reset %s: %v", table, err)
		}
	}
}

func getSharedTestDB(t *testing.T) *database.DB {
	testDBPath := getTestDBPath()

	db, err := database.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to open shared test database: %v", err)
	}

	return db
}
