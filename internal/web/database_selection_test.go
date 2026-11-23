package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"stonks/internal/database"
)

func TestProductionRejectsIntegrationTestDatabaseSelection(t *testing.T) {
	form := url.Values{"database": {database.IntegrationTestDatabaseName}}
	request := httptest.NewRequest(http.MethodPost, "/database/set-current", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()

	server := &Server{dbName: database.ProductionDatabaseName}
	server.handleSetCurrentDatabase(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusForbidden, response.Body.String())
	}
	if got := server.getCurrentDatabaseName(); got != database.ProductionDatabaseName {
		t.Fatalf("database changed to %q", got)
	}
}

func TestServerReportsExplicitDatabaseName(t *testing.T) {
	server := &Server{dbName: database.IntegrationTestDatabaseName}
	if got := server.getCurrentDatabaseName(); got != database.IntegrationTestDatabaseName {
		t.Fatalf("getCurrentDatabaseName() = %q, want %q", got, database.IntegrationTestDatabaseName)
	}
}
