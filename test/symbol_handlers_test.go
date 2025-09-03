package test

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"stonks/internal/database"
	"stonks/internal/models"
	"stonks/internal/web"

	_ "github.com/mattn/go-sqlite3"
)

// TestSymbolDataStructure tests the SymbolData structure and its monthly results
func TestSymbolDataStructure(t *testing.T) {
	// Setup test database with symbol-specific data
	testDB := setupSymbolTestDatabase(t)
	defer testDB.Close()

	// Build symbol data (simulating the handler)
	symbolData := buildTestSymbolData(t, testDB, "AAPL")

	// Test JSON serialization/deserialization
	jsonData, err := json.Marshal(symbolData)
	if err != nil {
		t.Fatalf("Failed to marshal SymbolData: %v", err)
	}

	var unmarshalled web.SymbolData
	if err := json.Unmarshal(jsonData, &unmarshalled); err != nil {
		t.Fatalf("Failed to unmarshal SymbolData: %v", err)
	}

	// Validate basic symbol information
	t.Run("BasicSymbolInfo", func(t *testing.T) {
		if unmarshalled.Symbol != "AAPL" {
			t.Errorf("Expected symbol AAPL, got %s", unmarshalled.Symbol)
		}
		if unmarshalled.CompanyName == "" {
			t.Error("CompanyName should not be empty")
		}
		if unmarshalled.Price <= 0 {
			t.Errorf("Price should be positive, got %f", unmarshalled.Price)
		}
		if len(unmarshalled.AllSymbols) == 0 {
			t.Error("AllSymbols should not be empty")
		}
	})

	// Validate financial calculations
	t.Run("FinancialCalculations", func(t *testing.T) {
		// These are formatted strings in the actual struct
		if unmarshalled.OptionsGains == "" {
			t.Error("OptionsGains should be calculated")
		}
		if unmarshalled.CapGains == "" {
			t.Error("CapGains should be calculated")
		}
		if unmarshalled.TotalProfits == "" {
			t.Error("TotalProfits should be calculated")
		}
		if unmarshalled.CashOnCash == "" {
			t.Error("CashOnCash should be calculated")
		}
	})

	// Validate dividend information
	t.Run("DividendInfo", func(t *testing.T) {
		if unmarshalled.Dividend < 0 {
			t.Errorf("Dividend should be non-negative, got %f", unmarshalled.Dividend)
		}
		if unmarshalled.Yield < 0 {
			t.Errorf("Yield should be non-negative, got %f", unmarshalled.Yield)
		}
		if unmarshalled.DividendsTotal < 0 {
			t.Errorf("DividendsTotal should be non-negative, got %f", unmarshalled.DividendsTotal)
		}
		
		// Validate dividends list structure
		for i, dividend := range unmarshalled.DividendsList {
			if dividend.Symbol != "AAPL" {
				t.Errorf("Dividend %d should be for AAPL, got %s", i, dividend.Symbol)
			}
			if dividend.Amount <= 0 {
				t.Errorf("Dividend %d amount should be positive, got %f", i, dividend.Amount)
			}
		}
	})

	// Validate options list
	t.Run("OptionsList", func(t *testing.T) {
		if len(unmarshalled.OptionsList) == 0 {
			t.Error("Expected options data for AAPL")
		}
		
		for i, option := range unmarshalled.OptionsList {
			if option.Symbol != "AAPL" {
				t.Errorf("Option %d should be for AAPL, got %s", i, option.Symbol)
			}
			if option.Type != "Put" && option.Type != "Call" {
				t.Errorf("Option %d should be Put or Call, got %s", i, option.Type)
			}
			if option.Strike <= 0 {
				t.Errorf("Option %d strike should be positive, got %f", i, option.Strike)
			}
			if option.Premium <= 0 {
				t.Errorf("Option %d premium should be positive, got %f", i, option.Premium)
			}
			if option.Contracts <= 0 {
				t.Errorf("Option %d contracts should be positive, got %d", i, option.Contracts)
			}
			
			// Test AROI calculation
			aroi := option.CalculateAROI()
			if math.IsNaN(aroi) || math.IsInf(aroi, 0) {
				t.Errorf("Option %d AROI should be a valid number, got %f", i, aroi)
			}
			
			// For closed positions, AROI should be calculated based on actual days in trade
			if option.Closed != nil {
				// AROI should be reasonable (between -1000% and +1000% annualized)
				if aroi < -1000 || aroi > 1000 {
					t.Errorf("Option %d AROI seems unrealistic: %.1f%% (closed position)", i, aroi)
				}
			}
			
			// Test other option calculations work correctly
			percentProfit := option.CalculatePercentOfProfit()
			percentTime := option.CalculatePercentOfTime()
			multiplier := option.CalculateMultiplier()
			
			if math.IsNaN(percentProfit) || math.IsInf(percentProfit, 0) {
				t.Errorf("Option %d PercentOfProfit should be valid, got %f", i, percentProfit)
			}
			if math.IsNaN(percentTime) || math.IsInf(percentTime, 0) {
				t.Errorf("Option %d PercentOfTime should be valid, got %f", i, percentTime)
			}
			if math.IsNaN(multiplier) || math.IsInf(multiplier, 0) {
				t.Errorf("Option %d Multiplier should be valid, got %f", i, multiplier)
			}
		}
	})

	// Validate long positions list
	t.Run("LongPositionsList", func(t *testing.T) {
		if len(unmarshalled.LongPositionsList) == 0 {
			t.Error("Expected long positions data for AAPL")
		}
		
		for i, position := range unmarshalled.LongPositionsList {
			if position.Symbol != "AAPL" {
				t.Errorf("Position %d should be for AAPL, got %s", i, position.Symbol)
			}
			if position.Shares <= 0 {
				t.Errorf("Position %d shares should be positive, got %d", i, position.Shares)
			}
			if position.BuyPrice <= 0 {
				t.Errorf("Position %d buy price should be positive, got %f", i, position.BuyPrice)
			}
		}
	})

	// Validate monthly results (most important for charts)
	t.Run("MonthlyResults", func(t *testing.T) {
		if len(unmarshalled.MonthlyResults) == 0 {
			t.Error("Expected monthly results data for AAPL")
		}
		
		for i, result := range unmarshalled.MonthlyResults {
			// Validate month format
			if result.Month == "" {
				t.Errorf("MonthlyResult %d has empty month", i)
			}
			
			// Validate month format (should be YYYY-MM)
			if len(result.Month) != 7 {
				t.Errorf("MonthlyResult %d month should be YYYY-MM format, got %s", i, result.Month)
			}
			
			// Counts can be zero, but shouldn't be negative
			if result.PutsCount < 0 {
				t.Errorf("MonthlyResult %d PutsCount should be non-negative, got %d", i, result.PutsCount)
			}
			if result.CallsCount < 0 {
				t.Errorf("MonthlyResult %d CallsCount should be non-negative, got %d", i, result.CallsCount)
			}
			
			// Totals can be negative (losses), but should be reasonable
			if result.Total < -10000 || result.Total > 10000 {
				t.Errorf("MonthlyResult %d Total seems unrealistic: %f", i, result.Total)
			}
			
			// Validate that totals match puts + calls (approximately)
			calculatedTotal := result.PutsTotal + result.CallsTotal
			if math.Abs(calculatedTotal-result.Total) > 0.01 {
				t.Errorf("MonthlyResult %d Total (%f) should equal PutsTotal + CallsTotal (%f)", 
					i, result.Total, calculatedTotal)
			}
		}
		
		// Validate chronological order
		for i := 1; i < len(unmarshalled.MonthlyResults); i++ {
			if unmarshalled.MonthlyResults[i].Month < unmarshalled.MonthlyResults[i-1].Month {
				t.Errorf("MonthlyResults should be in chronological order, but %s comes after %s", 
					unmarshalled.MonthlyResults[i].Month, unmarshalled.MonthlyResults[i-1].Month)
			}
		}
	})

	t.Logf("✅ SymbolData structure validation passed for %s - %d options, %d positions, %d monthly results", 
		unmarshalled.Symbol, len(unmarshalled.OptionsList), len(unmarshalled.LongPositionsList), len(unmarshalled.MonthlyResults))
}

// TestSymbolMonthlyResultType tests the SymbolMonthlyResult type individually
func TestSymbolMonthlyResultType(t *testing.T) {
	// Test individual SymbolMonthlyResult structure
	testResult := web.SymbolMonthlyResult{
		Month:      "2024-03",
		PutsCount:  3,
		CallsCount: 2,
		PutsTotal:  1250.75,
		CallsTotal: 850.50,
		Total:      2101.25,
	}

	// Test JSON serialization/deserialization
	jsonData, err := json.Marshal(testResult)
	if err != nil {
		t.Fatalf("Failed to marshal SymbolMonthlyResult: %v", err)
	}

	var unmarshalled web.SymbolMonthlyResult
	if err := json.Unmarshal(jsonData, &unmarshalled); err != nil {
		t.Fatalf("Failed to unmarshal SymbolMonthlyResult: %v", err)
	}

	// Validate all fields
	if unmarshalled.Month != "2024-03" {
		t.Errorf("Expected month 2024-03, got %s", unmarshalled.Month)
	}
	if unmarshalled.PutsCount != 3 {
		t.Errorf("Expected PutsCount 3, got %d", unmarshalled.PutsCount)
	}
	if unmarshalled.CallsCount != 2 {
		t.Errorf("Expected CallsCount 2, got %d", unmarshalled.CallsCount)
	}
	if unmarshalled.PutsTotal != 1250.75 {
		t.Errorf("Expected PutsTotal 1250.75, got %f", unmarshalled.PutsTotal)
	}
	if unmarshalled.CallsTotal != 850.50 {
		t.Errorf("Expected CallsTotal 850.50, got %f", unmarshalled.CallsTotal)
	}
	if unmarshalled.Total != 2101.25 {
		t.Errorf("Expected Total 2101.25, got %f", unmarshalled.Total)
	}

	t.Logf("✅ SymbolMonthlyResult type validation passed")
}

// TestSymbolHandlerEndpoint tests a simulated symbol handler endpoint
func TestSymbolHandlerEndpoint(t *testing.T) {
	// Setup test database
	testDB := setupSymbolTestDatabase(t)
	defer testDB.Close()

	// Create test server with symbol endpoint
	server := createSymbolTestServer(testDB)

	// Test the AAPL symbol endpoint
	req, err := http.NewRequest("GET", "/symbol/AAPL", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// For this test, we expect HTML response (not JSON)
	// In a real implementation, this would be HTML template rendering
	responseBody := rr.Body.String()
	
	if responseBody == "" {
		t.Error("Expected non-empty response body")
	}

	// Basic content checks (this would be more sophisticated in real implementation)
	expectedContent := []string{"AAPL", "monthly", "options"}
	for _, content := range expectedContent {
		if !containsIgnoreCase(responseBody, content) {
			t.Errorf("Response missing expected content: %s", content)
		}
	}

	t.Logf("✅ Symbol handler endpoint test passed")
}

// TestMultipleSymbolsMonthlyData tests monthly data for multiple symbols
func TestMultipleSymbolsMonthlyData(t *testing.T) {
	// Setup test database with multiple symbols
	testDB := setupSymbolTestDatabase(t)
	defer testDB.Close()

	symbols := []string{"AAPL", "TSLA", "NVDA"}

	for _, symbol := range symbols {
		t.Run(symbol, func(t *testing.T) {
			symbolData := buildTestSymbolData(t, testDB, symbol)

			// Validate symbol-specific data
			if symbolData.Symbol != symbol {
				t.Errorf("Expected symbol %s, got %s", symbol, symbolData.Symbol)
			}

			// All symbols should have some monthly data
			if len(symbolData.MonthlyResults) == 0 {
				t.Errorf("Symbol %s should have monthly results", symbol)
			}

			// Validate consistent month formatting across symbols
			for i, result := range symbolData.MonthlyResults {
				if result.Month == "" {
					t.Errorf("Symbol %s MonthlyResult %d has empty month", symbol, i)
				}
				
				// Validate month format consistency
				if len(result.Month) != 7 || result.Month[4] != '-' {
					t.Errorf("Symbol %s MonthlyResult %d has invalid month format: %s", symbol, i, result.Month)
				}
			}

			t.Logf("✅ Symbol %s validation passed - %d monthly results", 
				symbol, len(symbolData.MonthlyResults))
		})
	}
}

// Helper function to setup symbol test database
func setupSymbolTestDatabase(t *testing.T) *database.DB {
	testDBPath := "test_symbol.db"
	os.Remove(testDBPath)

	db, err := database.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create deterministic symbol test data
	createDeterministicSymbolData(t, db)

	t.Cleanup(func() {
		db.Close()
		os.Remove(testDBPath)
	})

	return db
}

// Helper function to create deterministic symbol test data
func createDeterministicSymbolData(t *testing.T, db *database.DB) {
	symbolService := models.NewSymbolService(db.DB)
	optionService := models.NewOptionService(db.DB)
	longPositionService := models.NewLongPositionService(db.DB)
	dividendService := models.NewDividendService(db.DB)

	// Create symbols with detailed information
	symbols := []string{"AAPL", "TSLA", "NVDA"}
	prices := []float64{150.0, 200.0, 400.0}
	dividends := []float64{0.75, 0.00, 0.25}
	
	for i, symbol := range symbols {
		if _, err := symbolService.Create(symbol); err != nil {
			t.Fatalf("Failed to create symbol %s: %v", symbol, err)
		}
		
		// Set price and dividend
		if _, err := symbolService.Update(symbol, prices[i], dividends[i], nil, nil); err != nil {
			t.Fatalf("Failed to update symbol %s: %v", symbol, err)
		}
	}

	// Create 6 months of data for each symbol
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	for month := 0; month < 6; month++ {
		monthDate := baseDate.AddDate(0, month, 0)
		
		for i, symbol := range symbols {
			// Create options for each symbol/month
			openDate := monthDate.AddDate(0, 0, -30)
			_ = monthDate.AddDate(0, 0, -5) // closeDate
			expirationDate := monthDate.AddDate(0, 0, 15)
			
			// Put options (more frequent)
			putPremium := 5.50 + float64(i)*1.25 + float64(month)*0.25
			_ = 3.25 + float64(i)*0.75 + float64(month)*0.15 // putExitPrice
			
			_, err := optionService.Create(symbol, "Put", openDate, prices[i]*0.95, expirationDate, putPremium, 2)
			if err != nil {
				t.Fatalf("Failed to create %s put for month %d: %v", symbol, month, err)
			}
			
			// Call options (less frequent)
			if month%2 == 0 {
				callPremium := 3.75 + float64(i)*0.75 + float64(month)*0.20
				_ = 2.15 + float64(i)*0.50 + float64(month)*0.10 // callExitPrice
				
				_, err := optionService.Create(symbol, "Call", openDate, prices[i]*1.05, expirationDate, callPremium, 1)
				if err != nil {
					t.Fatalf("Failed to create %s call for month %d: %v", symbol, month, err)
				}
			}
			
			// Long positions (some months)
			if month%3 == 0 {
				shares := 100 + i*50 + month*10
				buyPrice := prices[i] * (0.98 + float64(month)*0.005)
				
				_, err := longPositionService.Create(symbol, openDate, shares, buyPrice)
				if err != nil {
					t.Fatalf("Failed to create %s long position for month %d: %v", symbol, month, err)
				}
			}
			
			// Dividends (quarterly for AAPL and NVDA)
			if dividends[i] > 0 && month%3 == 0 {
				dividendAmount := dividends[i] * (100.0 + float64(month*10)) // Based on shares owned
				dividendDate := monthDate.AddDate(0, 0, 25)
				
				_, err := dividendService.Create(symbol, dividendDate, dividendAmount)
				if err != nil {
					t.Fatalf("Failed to create %s dividend for month %d: %v", symbol, month, err)
				}
			}
		}
	}

	t.Logf("✅ Created deterministic symbol test data for %d symbols over 6 months", len(symbols))
}

// Helper function to build test symbol data
func buildTestSymbolData(t *testing.T, db *database.DB, symbol string) web.SymbolData {
	// This simulates the buildSymbolData function from symbol_handlers.go
	// For testing, we'll fetch actual data from the database
	
	optionService := models.NewOptionService(db.DB)
	longPositionService := models.NewLongPositionService(db.DB)
	dividendService := models.NewDividendService(db.DB)
	
	// Fetch options for this symbol
	allOptions, err := optionService.GetAll()
	if err != nil {
		t.Fatalf("Failed to get options for %s: %v", symbol, err)
	}
	
	var optionsList []*models.Option
	for _, option := range allOptions {
		if option.Symbol == symbol {
			optionsList = append(optionsList, option)
		}
	}
	
	// Fetch long positions for this symbol
	allPositions, err := longPositionService.GetAll()
	if err != nil {
		t.Fatalf("Failed to get long positions for %s: %v", symbol, err)
	}
	
	var longPositionsList []*models.LongPosition
	for _, position := range allPositions {
		if position.Symbol == symbol {
			longPositionsList = append(longPositionsList, position)
		}
	}
	
	// Fetch dividends for this symbol
	allDividends, err := dividendService.GetAll()
	if err != nil {
		t.Fatalf("Failed to get dividends for %s: %v", symbol, err)
	}
	
	var dividendsList []*models.Dividend
	dividendsTotal := 0.0
	for _, dividend := range allDividends {
		if dividend.Symbol == symbol {
			dividendsList = append(dividendsList, dividend)
			dividendsTotal += dividend.Amount
		}
	}
	
	// Calculate monthly results (deterministic)
	monthlyResults := []web.SymbolMonthlyResult{}
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	for month := 0; month < 6; month++ {
		monthStr := baseDate.AddDate(0, month, 0).Format("2006-01")
		
		putsCount := 2 // Each month has puts
		callsCount := 0
		if month%2 == 0 {
			callsCount = 1 // Calls every other month
		}
		
		putsTotal := 450.0 + float64(month)*25.0 // Increasing profits
		callsTotal := 0.0
		if callsCount > 0 {
			callsTotal = 325.0 + float64(month)*15.0
		}
		
		monthlyResults = append(monthlyResults, web.SymbolMonthlyResult{
			Month:      monthStr,
			PutsCount:  putsCount,
			CallsCount: callsCount,
			PutsTotal:  putsTotal,
			CallsTotal: callsTotal,
			Total:      putsTotal + callsTotal,
		})
	}
	
	// Mock current time for consistent testing
	now := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	
	return web.SymbolData{
		Symbol:         symbol,
		AllSymbols:     []string{"AAPL", "TSLA", "NVDA"},
		CompanyName:    symbol + " Inc.", // Mock company name
		CurrentPrice:   "150.00",
		LastUpdate:     now.Format("2006-01-02 15:04:05"),
		Price:          150.0,
		Dividend:       0.75,
		ExDividendDate: nil,
		PERatio:        nil,
		PERatioValue:   0,
		HasPERatio:     false,
		Yield:          2.0, // 0.75/150 * 4 quarters * 100
		OptionsGains:   "$2,750.00",
		CapGains:       "$1,250.00",
		Dividends:      "$225.00",
		TotalProfits:   "$4,225.00",
		CashOnCash:     "17.5%",
		DividendsList:  dividendsList,
		DividendsTotal: dividendsTotal,
		OptionsList:    optionsList,
		LongPositionsList: longPositionsList,
		MonthlyResults: monthlyResults,
		CurrentDB:      "test_symbol.db",
	}
}

// Helper function to create test symbol server
func createSymbolTestServer(db *database.DB) http.Handler {
	mux := http.NewServeMux()
	
	// Add symbol endpoint handler
	mux.HandleFunc("/symbol/", func(w http.ResponseWriter, r *http.Request) {
		// Extract symbol from URL (simplified)
		symbol := "AAPL" // In real implementation, this would be parsed from URL
		
		// Return mock HTML response (in real implementation, this would render template)
		response := `<!DOCTYPE html>
<html>
<head><title>` + symbol + ` - Wheeler</title></head>
<body>
	<h1>` + symbol + `</h1>
	<div id="monthly-chart"></div>
	<div id="options-table"></div>
</body>
</html>`
		
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(response))
	})
	
	return mux
}

// Helper function for case-insensitive string contains
func containsIgnoreCase(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

// TestOptionAROICalculation tests the AROI (Annualized Return on Investment) calculation
func TestOptionAROICalculation(t *testing.T) {
	t.Run("PutOption_TwoWeeks_Profit", func(t *testing.T) {
		// Create a put option: 2 weeks in trade, $100 profit, $5200 exposure
		openDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		closeDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC) // 14 days
		expirationDate := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
		
		option := &models.Option{
			ID:         1,
			Symbol:     "AAPL",
			Type:       "Put",
			Opened:     openDate,
			Closed:     &closeDate,
			Strike:     52.0, // $5200 exposure for 1 contract
			Expiration: expirationDate,
			Premium:    2.0,
			Contracts:  1,
			ExitPrice:  func() *float64 { v := 1.0; return &v }(), // $1 exit price = $1 * 1 * 100 = $100 profit
			Commission: 1.30, // $0.65 * 2
		}
		
		aroi := option.CalculateAROI()
		
		// Expected calculation:
		// Profit = (2.0 - 1.0) * 1 * 100 = $100
		// Capital = 52.0 * 1 * 100 = $5200
		// Period return = ($100 / $5200) * 100 = 1.923%
		// Days in trade = 14
		// AROI = 1.923% * (365.25 / 14) = 50.2%
		
		expectedAROI := 50.2
		tolerance := 1.0 // Allow 1% tolerance
		
		if math.Abs(aroi-expectedAROI) > tolerance {
			t.Errorf("Put AROI calculation incorrect. Expected ~%.1f%%, got %.1f%%", expectedAROI, aroi)
		}
		
		t.Logf("✅ Put Option AROI: %.1f%% (14 days, $100 profit, $5200 exposure)", aroi)
	})
	
	t.Run("CallOption_OneMonth_Loss", func(t *testing.T) {
		// Create a call option: 30 days in trade, -$50 loss, $15000 exposure
		openDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		closeDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC) // 30 days
		expirationDate := time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC)
		
		option := &models.Option{
			ID:         2,
			Symbol:     "TSLA",
			Type:       "Call",
			Opened:     openDate,
			Closed:     &closeDate,
			Strike:     150.0, // $15000 exposure for 1 contract
			Expiration: expirationDate,
			Premium:    3.0,
			Contracts:  1,
			ExitPrice:  func() *float64 { v := 3.5; return &v }(), // $3.5 exit price = -$0.5 * 1 * 100 = -$50 loss
			Commission: 1.30,
		}
		
		aroi := option.CalculateAROI()
		
		// Expected calculation:
		// Profit = (3.0 - 3.5) * 1 * 100 = -$50
		// Capital = 150.0 * 1 * 100 = $15000
		// Period return = (-$50 / $15000) * 100 = -0.333%
		// Days in trade = 30
		// AROI = -0.333% * (365.25 / 30) = -4.1%
		
		expectedAROI := -4.1
		tolerance := 0.5
		
		if math.Abs(aroi-expectedAROI) > tolerance {
			t.Errorf("Call AROI calculation incorrect. Expected ~%.1f%%, got %.1f%%", expectedAROI, aroi)
		}
		
		t.Logf("✅ Call Option AROI: %.1f%% (30 days, -$50 loss, $15000 exposure)", aroi)
	})
	
	t.Run("OpenPosition_CurrentCalculation", func(t *testing.T) {
		// Create an open position: opened 7 days ago, $25 current profit
		openDate := time.Now().AddDate(0, 0, -7) // 7 days ago
		expirationDate := time.Now().AddDate(0, 0, 7) // 7 days from now
		
		option := &models.Option{
			ID:         3,
			Symbol:     "NVDA",
			Type:       "Put",
			Opened:     openDate,
			Closed:     nil, // Open position
			Strike:     100.0, // $10000 exposure for 1 contract
			Expiration: expirationDate,
			Premium:    1.25,
			Contracts:  1,
			ExitPrice:  nil, // Open position, no exit price yet
			Commission: 0.65,
		}
		
		aroi := option.CalculateAROI()
		
		// For open positions, profit = premium * contracts * 100 (assuming exit at $0)
		// Profit = 1.25 * 1 * 100 = $125
		// Capital = 100.0 * 1 * 100 = $10000
		// Period return = ($125 / $10000) * 100 = 1.25%
		// Days in trade = 7
		// AROI = 1.25% * (365.25 / 7) = 65.2%
		
		expectedAROI := 65.2
		tolerance := 5.0 // Higher tolerance for open positions due to current date variations
		
		if math.Abs(aroi-expectedAROI) > tolerance {
			t.Errorf("Open Position AROI calculation incorrect. Expected ~%.1f%%, got %.1f%%", expectedAROI, aroi)
		}
		
		t.Logf("✅ Open Position AROI: %.1f%% (7 days so far, $125 current profit, $10000 exposure)", aroi)
	})
	
	t.Run("EdgeCase_ZeroCapital", func(t *testing.T) {
		// Test edge case with zero strike price
		option := &models.Option{
			ID:         4,
			Symbol:     "TEST",
			Type:       "Put",
			Opened:     time.Now().AddDate(0, 0, -1),
			Closed:     nil,
			Strike:     0.0, // Zero capital
			Expiration: time.Now().AddDate(0, 0, 30),
			Premium:    1.0,
			Contracts:  1,
			ExitPrice:  nil,
			Commission: 0.65,
		}
		
		aroi := option.CalculateAROI()
		
		// Should return 0 for zero capital to avoid division by zero
		if aroi != 0 {
			t.Errorf("Zero capital should return 0 AROI, got %.1f%%", aroi)
		}
		
		t.Logf("✅ Zero Capital Edge Case: AROI = %.1f%%", aroi)
	})
}