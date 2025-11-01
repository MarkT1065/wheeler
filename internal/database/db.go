package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaFS embed.FS

type DB struct {
	*sql.DB
}

func NewDB(dataSourceName string) (*DB, error) {
	// Add SQLite connection parameters for better reliability
	connStr := dataSourceName + "?_busy_timeout=10000&_journal_mode=WAL&_foreign_keys=on"
	
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	dbWrapper := &DB{DB: db}
	if err := dbWrapper.InitSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Run migrations after schema is initialized
	if err := dbWrapper.runMigrations(dataSourceName); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return dbWrapper, nil
}

func (db *DB) InitSchema() error {
	schemaSQL, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	if _, err := db.Exec(string(schemaSQL)); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	if err := db.runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func (db *DB) runMigrations() error {
	var hasCurrentPrice bool
	err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('options') WHERE name = 'current_price'").Scan(&hasCurrentPrice)
	if err != nil {
		return fmt.Errorf("failed to check for current_price column: %w", err)
	}

	if !hasCurrentPrice {
		_, err := db.Exec("ALTER TABLE options ADD COLUMN current_price REAL")
		if err != nil {
			return fmt.Errorf("failed to add current_price column: %w", err)
		}
	}

	return nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

// BackupDatabase creates a timestamped backup of the specified database
func BackupDatabase(dbPath string) (string, error) {
	// Ensure backup directory exists
	backupDir := "./data/backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	baseName := strings.TrimSuffix(filepath.Base(dbPath), ".db")
	backupFileName := fmt.Sprintf("%s.%s.db", baseName, timestamp)
	backupPath := filepath.Join(backupDir, backupFileName)

	// Copy the file
	sourceFile, err := os.Open(dbPath)
	if err != nil {
		return "", fmt.Errorf("failed to open source: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return "", fmt.Errorf("failed to copy data: %w", err)
	}

	if err := destFile.Sync(); err != nil {
		return "", fmt.Errorf("failed to sync data: %w", err)
	}

	log.Printf("Database backup created: %s", backupPath)
	return backupPath, nil
}

// runMigrations performs database migrations with automatic backup
func (db *DB) runMigrations(dbPath string) error {
	// Check if commission migration has run
	var value string
	err := db.QueryRow("SELECT value FROM settings WHERE name = 'COMMISSION_DATA_MIGRATION_V1'").Scan(&value)
	if err == nil && value == "completed" {
		// Migration already completed
		log.Println("[MIGRATION] Commission data migration already completed, skipping")
		return nil
	}

	log.Println("[MIGRATION] Commission data migration needed - converting per-contract to total commission")

	// Create backup before migration (skip for in-memory databases)
	if dbPath != ":memory:" && dbPath != "file::memory:?cache=shared" {
		backupPath, err := BackupDatabase(dbPath)
		if err != nil {
			return fmt.Errorf("failed to create pre-migration backup: %w", err)
		}
		log.Printf("[MIGRATION] Pre-migration backup created: %s", backupPath)
	} else {
		log.Println("[MIGRATION] Skipping backup for in-memory database")
	}

	// Start transaction for migration
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start migration transaction: %w", err)
	}
	defer tx.Rollback()

	// Update commission from per-contract to total
	// For open/expired positions: total = commission * contracts
	// For closed positions (buy-to-close): total = commission * contracts * 2
	updateSQL := `
		UPDATE options
		SET commission = CASE
			WHEN closed IS NOT NULL AND exit_price IS NOT NULL AND exit_price > 0
			THEN commission * contracts * 2.0
			ELSE commission * contracts
		END
		WHERE commission > 0
	`

	result, err := tx.Exec(updateSQL)
	if err != nil {
		return fmt.Errorf("failed to update commission data: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("[MIGRATION] Updated %d option records with total commission", rowsAffected)

	// Mark migration as complete
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO settings (name, value, description)
		VALUES ('COMMISSION_DATA_MIGRATION_V1', 'completed', 'Commission data migrated from per-contract to total')
	`)
	if err != nil {
		return fmt.Errorf("failed to mark migration as complete: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	log.Println("[MIGRATION] Commission data migration completed successfully")
	return nil
}

// GetCurrentDatabase reads the current database filename from ./data/currentdb
func GetCurrentDatabase() (string, error) {
	// Ensure data directory exists
	if err := os.MkdirAll("./data", 0755); err != nil {
		return "", fmt.Errorf("failed to create data directory: %w", err)
	}

	currentDBPath := "./data/currentdb"
	
	// Check if currentdb file exists
	if _, err := os.Stat(currentDBPath); os.IsNotExist(err) {
		// Create default currentdb file with wheeler.db
		if err := os.WriteFile(currentDBPath, []byte("wheeler.db"), 0644); err != nil {
			return "", fmt.Errorf("failed to create currentdb file: %w", err)
		}
		return "wheeler.db", nil
	}

	// Read the current database name
	data, err := os.ReadFile(currentDBPath)
	if err != nil {
		return "", fmt.Errorf("failed to read currentdb file: %w", err)
	}

	dbName := strings.TrimSpace(string(data))
	if dbName == "" {
		dbName = "wheeler.db"
	}

	return dbName, nil
}

// SetCurrentDatabase writes the current database filename to ./data/currentdb
func SetCurrentDatabase(dbName string) error {
	// Ensure data directory exists
	if err := os.MkdirAll("./data", 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	currentDBPath := "./data/currentdb"
	if err := os.WriteFile(currentDBPath, []byte(dbName), 0644); err != nil {
		return fmt.Errorf("failed to write currentdb file: %w", err)
	}

	return nil
}

// GetCurrentDatabasePath returns the full path to the current database
func GetCurrentDatabasePath() (string, error) {
	dbName, err := GetCurrentDatabase()
	if err != nil {
		return "", err
	}
	
	return filepath.Join("./data", dbName), nil
}

// CreateNewDatabase creates a new SQLite database in the data directory
func CreateNewDatabase(name string) error {
	// Ensure data directory exists
	if err := os.MkdirAll("./data", 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Add .db extension if not present
	if !strings.HasSuffix(name, ".db") {
		name = name + ".db"
	}

	dbPath := filepath.Join("./data", name)
	
	// Check if database already exists
	if _, err := os.Stat(dbPath); err == nil {
		return fmt.Errorf("database %s already exists", name)
	}

	// Create new database with schema
	_, err := NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	return nil
}

// ListDatabases returns a list of all .db files in the data directory
func ListDatabases() ([]string, error) {
	dataDir := "./data"
	
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	files, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}

	var databases []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(file.Name()), ".db") {
			databases = append(databases, file.Name())
		}
	}

	return databases, nil
}