package polygon

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"stonks/internal/database"
	"stonks/internal/models"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TestLivePolygonIntegration tests against the real Polygon.io API
// This test only runs if a valid API key is configured in the Wheeler database
func TestLivePolygonIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live integration test in short mode")
	}

	// Read the current database path
	currentDBPath, err := getCurrentDatabasePath()
	if err != nil {
		t.Skipf("Could not determine current database: %v", err)
	}

	// Connect to the Wheeler database
	dbWrapper, err := database.NewDB(currentDBPath)
	if err != nil {
		t.Skipf("Could not connect to Wheeler database: %v", err)
	}
	defer dbWrapper.DB.Close()

	// Get the API key from settings
	settingService := models.NewSettingService(dbWrapper.DB)
	apiKey := settingService.GetValue("POLYGON_API_KEY")
	
	if apiKey == "" {
		t.Skip("No Polygon API key configured in Wheeler database - skipping live integration test")
	}

	// Mask API key for logging
	maskedKey := maskAPIKey(apiKey)
	t.Logf("Running live integration tests with API key: %s", maskedKey)

	// Create Polygon client and service
	client := NewClient(apiKey)
	symbolService := models.NewSymbolService(dbWrapper.DB)
	service := NewService(symbolService, settingService)

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

	t.Run("TestServiceConnection", func(t *testing.T) {
		err := service.TestConnection(ctx)
		if err != nil {
			t.Errorf("Service connection test failed: %v", err)
		} else {
			t.Logf("✅ Service connection successful")
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

	t.Run("TestServiceSymbolPriceUpdate", func(t *testing.T) {
		// Ensure TSLA symbol exists in database
		_, err := symbolService.Create("TSLA")
		if err != nil && !strings.Contains(err.Error(), "UNIQUE constraint failed") {
			t.Errorf("Failed to create test symbol: %v", err)
			return
		}

		// Test service-level price update
		err = service.UpdateSymbolPrice(ctx, "TSLA")
		if err != nil {
			t.Errorf("Service price update failed: %v", err)
			return
		}

		// Verify the price was updated in the database
		symbol, err := symbolService.GetBySymbol("TSLA")
		if err != nil {
			t.Errorf("Failed to retrieve updated symbol: %v", err)
			return
		}

		if symbol.Price <= 0 {
			t.Errorf("Expected positive updated price, got %f", symbol.Price)
		}

		t.Logf("✅ TSLA price updated in database: $%.2f", symbol.Price)
	})

	t.Run("TestServiceFetchSymbolDetails", func(t *testing.T) {
		info, err := service.FetchSymbolDetails(ctx, "NVDA")
		if err != nil {
			t.Errorf("Service fetch symbol details failed: %v", err)
			return
		}

		if info == nil {
			t.Error("Expected symbol info but got nil")
			return
		}

		if info.Symbol != "NVDA" {
			t.Errorf("Expected symbol NVDA, got %s", info.Symbol)
		}

		if info.Name == "" {
			t.Error("Expected non-empty company name")
		}

		if info.CurrentPrice <= 0 {
			t.Errorf("Expected positive current price, got %f", info.CurrentPrice)
		}

		t.Logf("✅ NVDA details: %s, Current Price: $%.2f, Market Cap: $%.2f", info.Name, info.CurrentPrice, info.MarketCap)
	})

	t.Run("TestServiceFetchDividendHistory", func(t *testing.T) {
		dividends, err := service.FetchDividendHistory(ctx, "JNJ", 3)
		if err != nil {
			t.Errorf("Service fetch dividend history failed: %v", err)
			return
		}

		if dividends == nil {
			t.Error("Expected dividends but got nil")
			return
		}

		if len(dividends) > 0 {
			firstDiv := dividends[0]
			if firstDiv.Symbol != "JNJ" {
				t.Errorf("Expected symbol JNJ, got %s", firstDiv.Symbol)
			}
			if firstDiv.CashAmount <= 0 {
				t.Errorf("Expected positive cash amount, got %f", firstDiv.CashAmount)
			}
			t.Logf("✅ JNJ recent dividend: $%.3f (Ex-Date: %s, Pay Date: %s)", firstDiv.CashAmount, firstDiv.ExDividendDate, firstDiv.PayDate)
		} else {
			t.Log("✅ JNJ dividend history query successful (no results)")
		}
	})

	// Rate limiting test - Polygon free tier has 5 requests per minute
	t.Log("⏱️  Respecting API rate limits (free tier: 5 requests/minute)")
}

// getCurrentDatabasePath reads the current database path from ./data/currentdb
func getCurrentDatabasePath() (string, error) {
	currentDBFile := "../../data/currentdb"
	
	// Check if the currentdb file exists
	if _, err := os.Stat(currentDBFile); os.IsNotExist(err) {
		return "", fmt.Errorf("currentdb file not found at %s", currentDBFile)
	}

	// Read the database filename
	content, err := ioutil.ReadFile(currentDBFile)
	if err != nil {
		return "", fmt.Errorf("failed to read currentdb file: %w", err)
	}

	dbName := strings.TrimSpace(string(content))
	if dbName == "" {
		return "", fmt.Errorf("currentdb file is empty")
	}

	// Construct the full path to the database
	dbPath := filepath.Join("../../data", dbName)
	
	// Verify the database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return "", fmt.Errorf("database file not found at %s", dbPath)
	}

	return dbPath, nil
}

// maskAPIKey masks an API key for safe logging
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 6 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:3] + "..." + apiKey[len(apiKey)-3:]
}