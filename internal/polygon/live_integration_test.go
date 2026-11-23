package polygon

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// TestLivePolygonIntegration tests against the real Polygon.io API
// This test only runs when POLYGON_API_KEY is explicitly set in the test environment.
// It never reads credentials or data from the production Wheeler database.
func TestLivePolygonIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live integration test in short mode")
	}

	apiKey := os.Getenv("POLYGON_API_KEY")
	if apiKey == "" {
		t.Skip("POLYGON_API_KEY is not set - skipping live integration test")
	}

	// Mask API key for logging
	maskedKey := maskAPIKey(apiKey)
	t.Logf("Running live integration tests with API key: %s", maskedKey)

	client := NewClient(apiKey)

	// Run tests with generous timeout for API calls
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("TestAPIKeyValidation", func(t *testing.T) {
		err := client.IsValidAPIKey(ctx)
		if err != nil {
			t.Errorf("API key validation failed: %v", err)
		} else {
			t.Logf("✅ API key is valid")
		}
	})

	t.Run("TestGetPreviousClose", func(t *testing.T) {
		quote, err := client.GetPreviousClose(ctx, "AAPL")
		if err != nil {
			t.Errorf("GetPreviousClose failed: %v", err)
			return
		}

		if quote == nil {
			t.Error("Expected quote but got nil")
			return
		}

		if quote.Status != "OK" {
			t.Errorf("Expected status OK, got %s", quote.Status)
		}

		if quote.Results.Symbol != "AAPL" {
			t.Errorf("Expected symbol AAPL, got %s", quote.Results.Symbol)
		}

		if quote.Results.Price <= 0 {
			t.Errorf("Expected positive price, got %f", quote.Results.Price)
		}

		t.Logf("✅ AAPL previous close: $%.2f", quote.Results.Price)
	})

	t.Run("TestGetLastQuote", func(t *testing.T) {
		quote, err := client.GetLastQuote(ctx, "MSFT")
		if err != nil {
			// Free API keys don't have access to /v2/last/nbbo endpoint
			if strings.Contains(err.Error(), "forbidden") || strings.Contains(err.Error(), "status 403") {
				t.Skipf("GetLastQuote endpoint not available with current API key (likely free tier): %v", err)
				return
			}
			t.Errorf("GetLastQuote failed: %v", err)
			return
		}

		if quote == nil {
			t.Error("Expected quote but got nil")
			return
		}

		if quote.Status != "OK" {
			t.Errorf("Expected status OK, got %s", quote.Status)
		}

		if quote.Results.Symbol != "MSFT" {
			t.Errorf("Expected symbol MSFT, got %s", quote.Results.Symbol)
		}

		if quote.Results.Price <= 0 {
			t.Errorf("Expected positive price, got %f", quote.Results.Price)
		}

		t.Logf("✅ MSFT last quote: $%.2f", quote.Results.Price)
	})

	t.Run("TestGetTickerDetails", func(t *testing.T) {
		details, err := client.GetTickerDetails(ctx, "GOOGL")
		if err != nil {
			t.Errorf("GetTickerDetails failed: %v", err)
			return
		}

		if details == nil {
			t.Error("Expected details but got nil")
			return
		}

		if details.Status != "OK" {
			t.Errorf("Expected status OK, got %s", details.Status)
		}

		if details.Results.Symbol != "GOOGL" {
			t.Errorf("Expected symbol GOOGL, got %s", details.Results.Symbol)
		}

		if details.Results.Name == "" {
			t.Error("Expected non-empty company name")
		}

		t.Logf("✅ GOOGL details: %s (%s)", details.Results.Name, details.Results.Market)
	})

	t.Run("TestGetDividends", func(t *testing.T) {
		dividends, err := client.GetDividends(ctx, "KO", 5)
		if err != nil {
			t.Errorf("GetDividends failed: %v", err)
			return
		}

		if dividends == nil {
			t.Error("Expected dividends but got nil")
			return
		}

		if dividends.Status != "OK" {
			t.Errorf("Expected status OK, got %s", dividends.Status)
		}

		if len(dividends.Results) > 0 {
			firstDiv := dividends.Results[0]
			if firstDiv.Ticker != "KO" {
				t.Errorf("Expected ticker KO, got %s", firstDiv.Ticker)
			}
			if firstDiv.CashAmount <= 0 {
				t.Errorf("Expected positive cash amount, got %f", firstDiv.CashAmount)
			}
			t.Logf("✅ KO recent dividend: $%.3f on %s", firstDiv.CashAmount, firstDiv.ExDividendDate)
		} else {
			t.Log("✅ KO dividends query successful (no results)")
		}
	})

	t.Run("TestSpecialCharacterSymbol", func(t *testing.T) {
		// Test with Berkshire Hathaway Class A which has a dot in the symbol
		quote, err := client.GetPreviousClose(ctx, "BRK.A")
		if err != nil {
			t.Errorf("Special character symbol test failed: %v", err)
			return
		}

		if quote == nil {
			t.Error("Expected quote but got nil")
			return
		}

		if quote.Results.Symbol != "BRK.A" {
			t.Errorf("Expected symbol BRK.A, got %s", quote.Results.Symbol)
		}

		t.Logf("✅ BRK.A (special character) price: $%.2f", quote.Results.Price)
	})

	t.Run("TestGetOptionSnapshot", func(t *testing.T) {
		underlyingAsset := "AAPL"
		optionContract := "O:AAPL250117C00150000"

		snapshot, err := client.GetOptionSnapshot(ctx, underlyingAsset, optionContract)
		if err != nil {
			if strings.Contains(err.Error(), "forbidden") || strings.Contains(err.Error(), "status 403") {
				t.Skipf("Option snapshot endpoint not available with current API key (likely free tier): %v", err)
				return
			}
			t.Errorf("GetOptionSnapshot failed: %v", err)
			return
		}

		if snapshot == nil {
			t.Error("Expected snapshot but got nil")
			return
		}

		if snapshot.Status != "OK" {
			t.Errorf("Expected status OK, got %s", snapshot.Status)
		}

		if snapshot.Results.LastTrade.Price <= 0 {
			t.Errorf("Expected positive last trade price, got %f", snapshot.Results.LastTrade.Price)
		}

		t.Logf("✅ Option snapshot - Last Trade Price: $%.2f, Day Close: $%.2f",
			snapshot.Results.LastTrade.Price,
			snapshot.Results.Day.Close)
	})

	// Rate limiting test - Polygon free tier has 5 requests per minute
	t.Log("⏱️  Respecting API rate limits (free tier: 5 requests/minute)")
}

// maskAPIKey masks an API key for safe logging
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 6 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:3] + "..." + apiKey[len(apiKey)-3:]
}
