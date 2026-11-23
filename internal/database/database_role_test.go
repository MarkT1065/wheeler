package database

import (
	"path/filepath"
	"testing"
)

func TestDatabaseRolePaths(t *testing.T) {
	if got, want := ProductionDatabasePath(), filepath.Join("data", ProductionDatabaseName); got != want {
		t.Fatalf("ProductionDatabasePath() = %q, want %q", got, want)
	}
	if got, want := IntegrationTestDatabasePath(), filepath.Join("data", IntegrationTestDatabaseName); got != want {
		t.Fatalf("IntegrationTestDatabasePath() = %q, want %q", got, want)
	}
	if ProductionDatabasePath() == IntegrationTestDatabasePath() {
		t.Fatal("production and integration test database paths must differ")
	}
}
