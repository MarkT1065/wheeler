package database

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	// Check if currency column exists in symbols table
	var hasCurrency bool
	err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('symbols') WHERE name = 'currency'").Scan(&hasCurrency)
	if err != nil {
		return fmt.Errorf("failed to check for currency column: %w", err)
	}

	if !hasCurrency {
		_, err := db.Exec("ALTER TABLE symbols ADD COLUMN currency TEXT DEFAULT 'USD'")
		if err != nil {
			return fmt.Errorf("failed to add currency column: %w", err)
		}
	}

	return nil
}

func (db *DB) Close() error {
	return db.DB.Close()
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