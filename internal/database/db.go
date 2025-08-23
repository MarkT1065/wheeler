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
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

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