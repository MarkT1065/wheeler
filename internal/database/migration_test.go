package database

import (
	"context"
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

	t.Run("provider settings migration applied", func(t *testing.T) {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = '20260822220000_add_provider_settings'").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query provider settings migration: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected provider settings migration to be applied, got count=%d", count)
		}

		for _, name := range []string{"DATA_PROVIDER_TYPE", "GOOGLE_FINANCE_EXCHANGE"} {
			var value string
			if err := db.QueryRow("SELECT value FROM settings WHERE name = ?", name).Scan(&value); err != nil {
				t.Errorf("Expected setting %s to exist: %v", name, err)
			}
		}
	})

	t.Run("legacy Google exchange migration applied", func(t *testing.T) {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = '20260823211500_mark_google_exchange_legacy'").Scan(&count)
		if err != nil || count != 1 {
			t.Fatalf("expected legacy Google exchange migration, count=%d err=%v", count, err)
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
			"idx_options_direction",
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

	t.Run("options table has direction column", func(t *testing.T) {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('options') WHERE name='direction'").Scan(&count); err != nil || count != 1 {
			t.Fatalf("expected options.direction column, count=%d err=%v", count, err)
		}
	})

	t.Run("option direction defaults to Short and permits opposite legs", func(t *testing.T) {
		if _, err := db.ExecContext(context.Background(), "INSERT INTO symbols (symbol) VALUES ('NASDAQ:SPREAD')"); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts)
			VALUES ('NASDAQ:SPREAD', 'Put', '2026-08-01', 100, '2026-09-01', 2, 1)`); err != nil {
			t.Fatal(err)
		}
		var direction string
		if err := db.QueryRow("SELECT direction FROM options WHERE symbol = 'NASDAQ:SPREAD'").Scan(&direction); err != nil {
			t.Fatal(err)
		}
		if direction != "Short" {
			t.Fatalf("default direction = %q", direction)
		}
		if _, err := db.Exec(`INSERT INTO options (symbol, type, direction, opened, strike, expiration, premium, contracts)
			VALUES ('NASDAQ:SPREAD', 'Put', 'Long', '2026-08-01', 100, '2026-09-01', 2, 1)`); err != nil {
			t.Fatalf("opposite spread leg should be allowed: %v", err)
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
		if count != 4 {
			t.Errorf("Expected 4 migration records after re-running migrations, got %d", count)
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
