package test

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestGetDashboard tests the dashboard endpoint returns valid HTML
func TestGetDashboard(t *testing.T) {
	// Create HTTP client
	client := &http.Client{}

	// Make request to running server
	resp, err := client.Get("http://localhost:8081/")
	if err != nil {
		t.Fatalf("Failed to make request to test server: %v", err)
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Dashboard returned wrong status code: got %v want %v", resp.StatusCode, http.StatusOK)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	bodyStr := string(body)

	// Basic HTML validation
	if !strings.Contains(bodyStr, "<!DOCTYPE html>") {
		t.Error("Response does not appear to be valid HTML (missing DOCTYPE)")
	}

	if !strings.Contains(bodyStr, "<title>") {
		t.Error("Response missing HTML title tag")
	}

	// Wheeler-specific content checks
	expectedContent := []string{
		"Wheeler",
		"Dashboard",
	}

	for _, content := range expectedContent {
		if !strings.Contains(bodyStr, content) {
			t.Errorf("Dashboard HTML missing expected content: %s", content)
		}
	}

	// Check for error indicators
	errorStrings := []string{
		"Internal server error",
		"Template execution failed",
		"ERROR:",
	}

	for _, errorStr := range errorStrings {
		if strings.Contains(bodyStr, errorStr) {
			t.Errorf("Dashboard HTML contains error: %s", errorStr)
		}
	}

	t.Logf("✅ Dashboard test passed - HTML contains expected Wheeler content")
}