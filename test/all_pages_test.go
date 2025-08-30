package test

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestAllPageRendering tests every HTML page in the Wheeler application
func TestAllPageRendering(t *testing.T) {
	client := &http.Client{}

	// Define all main HTML pages to test
	pages := []struct {
		name string
		url  string
	}{
		{"Dashboard", "http://localhost:8081/"},
		{"Monthly", "http://localhost:8081/monthly"},
		{"Options", "http://localhost:8081/options"},
		{"Treasuries", "http://localhost:8081/treasuries"},
		{"Metrics", "http://localhost:8081/metrics"},
		{"Help", "http://localhost:8081/help"},
		{"Settings", "http://localhost:8081/settings"},
		{"Import", "http://localhost:8081/import"},
		{"Backup", "http://localhost:8081/backup"},
	}

	// Test each main page
	for _, page := range pages {
		t.Run(page.name, func(t *testing.T) {
			resp, err := client.Get(page.url)
			if err != nil {
				t.Fatalf("Failed to request %s page: %v", page.name, err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != http.StatusOK {
				t.Errorf("%s page returned status %d, expected %d", page.name, resp.StatusCode, http.StatusOK)
			}

			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read %s page response: %v", page.name, err)
			}
			bodyStr := string(body)

			// Basic HTML validation
			if !strings.Contains(bodyStr, "<!DOCTYPE html>") {
				t.Errorf("%s page missing DOCTYPE declaration", page.name)
			}

			if !strings.Contains(bodyStr, "<title>") {
				t.Errorf("%s page missing title tag", page.name)
			}

			// Check for error indicators
			errorStrings := []string{
				"Internal server error",
				"Internal Server Error", 
				"Template execution failed",
				"ERROR:",
				"500 Internal Server Error",
				"template:",
				"undefined",
			}

			for _, errorStr := range errorStrings {
				if strings.Contains(bodyStr, errorStr) {
					t.Errorf("%s page contains error: %s", page.name, errorStr)
				}
			}

			// Check for Wheeler-specific content
			if !strings.Contains(bodyStr, "Wheeler") {
				t.Errorf("%s page missing Wheeler branding", page.name)
			}

			t.Logf("✅ %s page rendered successfully", page.name)
		})
	}
}

// TestSymbolPagesRendering tests symbol-specific pages
func TestSymbolPagesRendering(t *testing.T) {
	client := &http.Client{}

	// Test symbol pages - these require actual symbols in the database
	// We'll test with common symbols that might be in test data
	testSymbols := []string{"AAPL", "MSFT", "NVDA", "TSLA", "VZ"}

	for _, symbol := range testSymbols {
		t.Run("Symbol_"+symbol, func(t *testing.T) {
			url := "http://localhost:8081/symbol/" + symbol
			resp, err := client.Get(url)
			if err != nil {
				t.Fatalf("Failed to request symbol page for %s: %v", symbol, err)
			}
			defer resp.Body.Close()

			// For symbol pages, we expect either 200 (symbol exists) or 404 (symbol doesn't exist)
			// Both are valid responses - we just want to ensure no server errors
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
				t.Errorf("Symbol page %s returned unexpected status %d", symbol, resp.StatusCode)
			}

			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read symbol page response for %s: %v", symbol, err)
			}
			bodyStr := string(body)

			// Check for server errors (500 errors indicate real problems)
			serverErrorStrings := []string{
				"Internal server error",
				"Internal Server Error",
				"Template execution failed", 
				"500 Internal Server Error",
			}

			for _, errorStr := range serverErrorStrings {
				if strings.Contains(bodyStr, errorStr) {
					t.Errorf("Symbol page %s contains server error: %s", symbol, errorStr)
				}
			}

			if resp.StatusCode == http.StatusOK {
				t.Logf("✅ Symbol page %s rendered successfully", symbol)
			} else {
				t.Logf("ℹ️ Symbol page %s returned %d (symbol may not exist in test database)", symbol, resp.StatusCode)
			}
		})
	}
}

// TestAddOptionAndTreasuryPages tests the add forms (these might redirect or show forms)
func TestAddFormsRendering(t *testing.T) {
	client := &http.Client{}
	
	// Don't follow redirects to see original response
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	addPages := []struct {
		name string
		url  string
	}{
		{"AddOption", "http://localhost:8081/add-option"},
		{"AddTreasury", "http://localhost:8081/add-treasury"},
	}

	for _, page := range addPages {
		t.Run(page.name, func(t *testing.T) {
			resp, err := client.Get(page.url)
			if err != nil {
				t.Fatalf("Failed to request %s page: %v", page.name, err)
			}
			defer resp.Body.Close()

			// These pages might redirect (302) or show forms (200) - both are fine
			// We just don't want server errors (500)
			if resp.StatusCode >= 500 {
				t.Errorf("%s page returned server error status %d", page.name, resp.StatusCode)
			}

			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read %s page response: %v", page.name, err)
			}
			bodyStr := string(body)

			// Check for server errors
			serverErrorStrings := []string{
				"Internal server error",
				"Internal Server Error",
				"Template execution failed",
				"500 Internal Server Error",
			}

			for _, errorStr := range serverErrorStrings {
				if strings.Contains(bodyStr, errorStr) {
					t.Errorf("%s page contains server error: %s", page.name, errorStr)
				}
			}

			t.Logf("✅ %s page handled successfully (status: %d)", page.name, resp.StatusCode)
		})
	}
}