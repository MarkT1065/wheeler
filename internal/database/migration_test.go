package database

import (
	"os"
	"testing"
)

func TestMigrationSystem(t *testing.T) {
	testDBPath := "test_migrations.db"
	defer os.Remove(testDBPath)

	db, err := NewDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	t.Run("schema_migrations table exists", func(t *testing.T) {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query for schema_migrations table: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected schema_migrations table to exist, got count=%d", count)
		}
	})

	t.Run("baseline migration applied", func(t *testing.T) {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = '20250111000001_baseline_v1_schema'").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query schema_migrations: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected baseline migration to be applied, got count=%d", count)
		}
	})

	t.Run("all expected tables exist", func(t *testing.T) {
		expectedTables := []string{
			"schema_migrations",
			"symbols",
			"long_positions",
			"options",
			"dividends",
			"treasuries",
			"settings",
			"metrics",
		}

		for _, table := range expectedTables {
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
			if err != nil {
				t.Fatalf("Failed to check for table %s: %v", table, err)
			}
			if count != 1 {
				t.Errorf("Expected table %s to exist", table)
			}
		}
	})

	t.Run("all expected indexes exist", func(t *testing.T) {
		expectedIndexes := []string{
			"idx_long_positions_symbol",
			"idx_long_positions_opened",
			"idx_options_symbol",
			"idx_options_expiration",
			"idx_options_type",
			"idx_dividends_symbol",
			"idx_dividends_received",
			"idx_treasuries_maturity",
			"idx_treasuries_purchased",
			"idx_metrics_created",
			"idx_metrics_type",
			"idx_options_unique",
			"idx_dividends_unique",
		}

		for _, index := range expectedIndexes {
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?", index).Scan(&count)
			if err != nil {
				t.Fatalf("Failed to check for index %s: %v", index, err)
			}
			if count != 1 {
				t.Errorf("Expected index %s to exist", index)
			}
		}
	})

	t.Run("options table has current_price column", func(t *testing.T) {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('options') WHERE name='current_price'").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check for current_price column: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected options.current_price column to exist")
		}
	})

	t.Run("migrations are idempotent", func(t *testing.T) {
		// Run migrations again - should not fail
		err := db.runMigrations()
		if err != nil {
			t.Errorf("Migrations should be idempotent, but failed on second run: %v", err)
		}

		// Check migration count didn't increase
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query schema_migrations: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected only 1 migration record after re-running migrations, got %d", count)
		}
	})
}

func TestCurrentDatabaseManagement(t *testing.T) {
	// Clean up any existing test files
	os.Remove("./data/currentdb")
	defer os.Remove("./data/currentdb")

	t.Run("GetCurrentDatabase creates default", func(t *testing.T) {
		dbName, err := GetCurrentDatabase()
		if err != nil {
			t.Fatalf("Failed to get current database: %v", err)
		}
		if dbName != "wheeler.db" {
			t.Errorf("Expected default database to be 'wheeler.db', got %s", dbName)
		}
	})

	t.Run("SetCurrentDatabase and read back", func(t *testing.T) {
		testDBName := "test_portfolio.db"
		err := SetCurrentDatabase(testDBName)
		if err != nil {
			t.Fatalf("Failed to set current database: %v", err)
		}

		dbName, err := GetCurrentDatabase()
		if err != nil {
			t.Fatalf("Failed to get current database: %v", err)
		}
		if dbName != testDBName {
			t.Errorf("Expected database name %s, got %s", testDBName, dbName)
		}
	})
}
