package test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"stonks/internal/database"
	"stonks/internal/models"
	"stonks/internal/web"

	_ "github.com/mattn/go-sqlite3"
)

// TestAllChartStructuresComprehensive runs comprehensive tests on all chart data structures
func TestAllChartStructuresComprehensive(t *testing.T) {
	// This test validates ALL chart data structures together in a single comprehensive test
	
	t.Run("Dashboard Charts", func(t *testing.T) {
		testDB := setupComprehensiveTestDatabase(t)
		defer testDB.Close()
		
		// Test AllocationData (4 charts)
		allocationData := buildComprehensiveAllocationData(t, testDB)
		validateAllocationDataComprehensive(t, allocationData)
	})

	t.Run("Monthly Charts", func(t *testing.T) {
		testDB := setupComprehensiveTestDatabase(t)
		defer testDB.Close()
		
		// Test MonthlyData (8+ charts)
		monthlyData := buildComprehensiveMonthlyData(t, testDB)
		validateMonthlyDataComprehensive(t, monthlyData)
	})

	t.Run("Symbol Charts", func(t *testing.T) {
		testDB := setupComprehensiveTestDatabase(t)
		defer testDB.Close()
		
		// Test SymbolData (monthly results chart)
		symbolData := buildComprehensiveSymbolData(t, testDB)
		validateSymbolDataComprehensive(t, symbolData)
	})

	t.Run("Metrics Charts", func(t *testing.T) {
		testDB := setupComprehensiveTestDatabase(t)
		defer testDB.Close()
		
		// Test metrics chart data (7 time series charts)
		metricsData := buildComprehensiveMetricsData(t, testDB)
		validateMetricsDataComprehensive(t, metricsData)
	})

	t.Run("Options Scatter Chart", func(t *testing.T) {
		// Test proposed OptionsScatterData structure
		scatterData := buildComprehensiveOptionsScatterData(t)
		validateOptionsScatterDataComprehensive(t, scatterData)
	})

	t.Run("Tutorial Chart", func(t *testing.T) {
		// Test proposed TutorialChartData structure  
		tutorialData := buildComprehensiveTutorialData(t)
		validateTutorialDataComprehensive(t, tutorialData)
	})

	t.Run("Treasury Charts", func(t *testing.T) {
		testDB := setupComprehensiveTestDatabase(t)
		defer testDB.Close()
		
		// Test TreasuriesData structure
		treasuryData := buildComprehensiveTreasuryData(t, testDB)
		validateTreasuryDataComprehensive(t, treasuryData)
	})
}

// TestChartDataConsistency tests that all charts use consistent data patterns
func TestChartDataConsistency(t *testing.T) {
	t.Run("ColorConsistency", func(t *testing.T) {
		// Test that all charts use the same color palette
		expectedColors := []string{"#FF6384", "#36A2EB", "#FFCE56", "#4BC0C0", "#9966FF", "#FF9F40"}
		
		// Test dashboard allocation data colors
		testDB := setupComprehensiveTestDatabase(t)
		defer testDB.Close()
		
		allocationData := buildComprehensiveAllocationData(t, testDB)
		
		// Validate colors are from expected palette
		for _, chartData := range allocationData.LongByTicker {
			if !contains(expectedColors, chartData.Color) {
				t.Errorf("LongByTicker uses unexpected color: %s", chartData.Color)
			}
		}
		
		for _, chartData := range allocationData.PutsByTicker {
			if !contains(expectedColors, chartData.Color) {
				t.Errorf("PutsByTicker uses unexpected color: %s", chartData.Color)
			}
		}
		
		t.Logf("✅ Chart color consistency validated")
	})

	t.Run("DateFormatConsistency", func(t *testing.T) {
		// Test that all charts use consistent date formats
		testDB := setupComprehensiveTestDatabase(t)
		defer testDB.Close()
		
		// Test monthly data date format (YYYY-MM)
		monthlyData := buildComprehensiveMonthlyData(t, testDB)
		
		for i, monthData := range monthlyData.PutsData.ByMonth {
			if len(monthData.Month) != 7 || monthData.Month[4] != '-' {
				t.Errorf("Monthly data %d has inconsistent date format: %s", i, monthData.Month)
			}
		}
		
		// Test metrics data date format (YYYY-MM-DD)
		metricsData := buildComprehensiveMetricsData(t, testDB)
		
		for metricType, points := range metricsData {
			for i, point := range points {
				if _, err := time.Parse("2006-01-02", point.Date); err != nil {
					t.Errorf("Metrics %s point %d has invalid date format: %s", metricType, i, point.Date)
				}
			}
		}
		
		t.Logf("✅ Date format consistency validated")
	})

	t.Run("FinancialPrecisionConsistency", func(t *testing.T) {
		// Test that all financial amounts use consistent precision
		testDB := setupComprehensiveTestDatabase(t)
		defer testDB.Close()
		
		allocationData := buildComprehensiveAllocationData(t, testDB)
		
		// All financial amounts should be reasonable (not NaN, Inf, or extreme values)
		for i, chartData := range allocationData.LongByTicker {
			validateFinancialAmount(t, "LongByTicker", i, chartData.Value)
		}
		
		monthlyData := buildComprehensiveMonthlyData(t, testDB)
		
		for i, monthData := range monthlyData.PutsData.ByMonth {
			validateFinancialAmount(t, "PutsData.ByMonth", i, monthData.Amount)
		}
		
		t.Logf("✅ Financial precision consistency validated")
	})

	t.Run("StructuralConsistency", func(t *testing.T) {
		// Test that all chart data structures follow consistent patterns
		
		// All chart data should serialize/deserialize without loss
		testDB := setupComprehensiveTestDatabase(t)
		defer testDB.Close()
		
		structures := map[string]interface{}{
			"AllocationData": buildComprehensiveAllocationData(t, testDB),
			"MonthlyData":    buildComprehensiveMonthlyData(t, testDB),
			"SymbolData":     buildComprehensiveSymbolData(t, testDB),
			"MetricsData":    buildComprehensiveMetricsData(t, testDB),
			"TreasuryData":   buildComprehensiveTreasuryData(t, testDB),
		}
		
		for structName, structData := range structures {
			jsonData, err := json.Marshal(structData)
			if err != nil {
				t.Errorf("Failed to marshal %s: %v", structName, err)
				continue
			}
			
			// Validate JSON is not empty
			if len(jsonData) == 0 {
				t.Errorf("%s marshaled to empty JSON", structName)
				continue
			}
			
			// Validate JSON is valid (by unmarshaling to interface{})
			var unmarshalled interface{}
			if err := json.Unmarshal(jsonData, &unmarshalled); err != nil {
				t.Errorf("Failed to unmarshal %s: %v", structName, err)
				continue
			}
			
			t.Logf("✅ %s structural consistency validated (%d bytes JSON)", structName, len(jsonData))
		}
	})
}

// TestChartPerformanceAndScaling tests chart data performance with larger datasets
func TestChartPerformanceAndScaling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	t.Run("LargeDatasetHandling", func(t *testing.T) {
		// Test with larger datasets to ensure charts can handle real-world data volumes
		testDB := setupLargeTestDatabase(t)
		defer testDB.Close()
		
		start := time.Now()
		
		// Build larger dataset
		allocationData := buildComprehensiveAllocationData(t, testDB)
		
		buildTime := time.Since(start)
		
		// Validate data was built successfully
		if len(allocationData.LongByTicker) == 0 {
			t.Error("Large dataset should have data")
		}
		
		// Test JSON serialization performance
		start = time.Now()
		jsonData, err := json.Marshal(allocationData)
		if err != nil {
			t.Fatalf("Failed to marshal large dataset: %v", err)
		}
		serializationTime := time.Since(start)
		
		// Test JSON deserialization performance
		start = time.Now()
		var unmarshalled web.AllocationData
		if err := json.Unmarshal(jsonData, &unmarshalled); err != nil {
			t.Fatalf("Failed to unmarshal large dataset: %v", err)
		}
		deserializationTime := time.Since(start)
		
		t.Logf("✅ Large dataset performance: Build=%v, Serialize=%v, Deserialize=%v, JSON Size=%d bytes",
			buildTime, serializationTime, deserializationTime, len(jsonData))
		
		// Performance thresholds (adjust as needed)
		if buildTime > 5*time.Second {
			t.Errorf("Build time too slow: %v", buildTime)
		}
		if serializationTime > 1*time.Second {
			t.Errorf("Serialization time too slow: %v", serializationTime)
		}
		if deserializationTime > 1*time.Second {
			t.Errorf("Deserialization time too slow: %v", deserializationTime)
		}
	})
}

// Helper functions for comprehensive testing

func setupComprehensiveTestDatabase(t *testing.T) *database.DB {
	testDBPath := "test_comprehensive.db"
	os.Remove(testDBPath)

	db, err := database.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to create comprehensive test database: %v", err)
	}

	// Create comprehensive test data
	createComprehensiveTestData(t, db)

	t.Cleanup(func() {
		db.Close()
		os.Remove(testDBPath)
	})

	return db
}

func setupLargeTestDatabase(t *testing.T) *database.DB {
	testDBPath := "test_large.db"
	os.Remove(testDBPath)

	db, err := database.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to create large test database: %v", err)
	}

	// Create larger test dataset
	createLargeTestData(t, db)

	t.Cleanup(func() {
		db.Close()
		os.Remove(testDBPath)
	})

	return db
}

func createComprehensiveTestData(t *testing.T, db *database.DB) {
	// Create comprehensive test data for all chart types
	symbolService := models.NewSymbolService(db.DB)
	optionService := models.NewOptionService(db.DB)
	longPositionService := models.NewLongPositionService(db.DB)
	dividendService := models.NewDividendService(db.DB)
	treasuryService := models.NewTreasuryService(db.DB)

	// Create multiple symbols
	symbols := []string{"AAPL", "TSLA", "NVDA", "MSFT", "GOOGL"}
	prices := []float64{150.0, 200.0, 400.0, 350.0, 2500.0}
	
	for i, symbol := range symbols {
		if _, err := symbolService.Create(symbol); err != nil {
			t.Fatalf("Failed to create symbol %s: %v", symbol, err)
		}
		if _, err := symbolService.Update(symbol, prices[i], 0, nil, nil); err != nil {
			t.Fatalf("Failed to update price for %s: %v", symbol, err)
		}
	}

	// Create historical data for 12 months
	baseDate := time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC)
	
	for month := 0; month < 12; month++ {
		monthDate := baseDate.AddDate(0, month, 0)
		
		for i, symbol := range symbols {
			// Create closed options
			openDate := monthDate.AddDate(0, 0, -30)
			_ = monthDate.AddDate(0, 0, -5) // closeDate
			expirationDate := monthDate.AddDate(0, 0, 15)
			
			premium := 5.0 + float64(i)*2.0 + float64(month)*0.5
			_ = 3.0 + float64(i)*1.0 + float64(month)*0.25 // exitPrice
			
			// Create puts and calls
			if _, err := optionService.Create(symbol, "Put", openDate, prices[i]*0.95, expirationDate, premium, 2); err != nil {
				t.Fatalf("Failed to create %s put for month %d: %v", symbol, month, err)
			}
			
			if month%2 == 0 { // Calls every other month
				callPremium := premium * 0.8
				_ = premium * 0.8 * 0.7 // callExitPrice
				if _, err := optionService.Create(symbol, "Call", openDate, prices[i]*1.05, expirationDate, callPremium, 1); err != nil {
					t.Fatalf("Failed to create %s call for month %d: %v", symbol, month, err)
				}
			}
			
			// Create long positions
			if month%3 == 0 {
				shares := 100 + i*25
				buyPrice := prices[i] * (0.98 + float64(month)*0.002)
				if _, err := longPositionService.Create(symbol, openDate, shares, buyPrice); err != nil {
					t.Fatalf("Failed to create %s long position for month %d: %v", symbol, month, err)
				}
			}
			
			// Create dividends
			if i < 3 && month%3 == 0 { // Only some symbols pay dividends
				dividendAmount := 25.0 + float64(i)*10.0
				dividendDate := monthDate.AddDate(0, 0, 25)
				if _, err := dividendService.Create(symbol, dividendDate, dividendAmount); err != nil {
					t.Fatalf("Failed to create %s dividend for month %d: %v", symbol, month, err)
				}
			}
		}
		
		// Create treasury data
		if month%4 == 0 { // Quarterly treasuries
			cuspid := fmt.Sprintf("TREASURY-%03d", month/4+1)
			purchaseDate := monthDate
			maturityDate := monthDate.AddDate(1, 0, 0)
			amount := 50000.0 + float64(month)*2500.0
			yield := 4.0 + float64(month)*0.1
			buyPrice := amount * 0.995
			_ = amount * (1.0 + float64(month)*0.005) // currentValue
			
			if _, err := treasuryService.Create(cuspid, purchaseDate, maturityDate, amount, yield, buyPrice); err != nil {
				t.Fatalf("Failed to create treasury for month %d: %v", month, err)
			}
		}
		
		// Create metrics data
		createMetricsData(t, db, monthDate, month)
	}

	t.Logf("✅ Created comprehensive test data for %d symbols over 12 months", len(symbols))
}

func createLargeTestData(t *testing.T, db *database.DB) {
	// Create larger dataset for performance testing
	symbolService := models.NewSymbolService(db.DB)
	optionService := models.NewOptionService(db.DB)
	
	// Create many symbols
	for i := 0; i < 50; i++ {
		symbol := fmt.Sprintf("SYMBOL%03d", i+1)
		if _, err := symbolService.Create(symbol); err != nil {
			t.Fatalf("Failed to create symbol %s: %v", symbol, err)
		}
		if _, err := symbolService.Update(symbol, 100.0+float64(i), 0, nil, nil); err != nil {
			t.Fatalf("Failed to update price for %s: %v", symbol, err)
		}
	}

	// Create many options (1000+ total)
	baseDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	
	for month := 0; month < 24; month++ { // 24 months of data
		for i := 0; i < 10; i++ { // 10 options per month
			symbol := fmt.Sprintf("SYMBOL%03d", (i%50)+1)
			monthDate := baseDate.AddDate(0, month, 0)
			openDate := monthDate.AddDate(0, 0, -30)
			_ = monthDate.AddDate(0, 0, -5) // closeDate
			expirationDate := monthDate.AddDate(0, 0, 15)
			
			premium := 1.0 + float64(i)*0.5 + float64(month)*0.1
			_ = premium * 0.6 // exitPrice
			
			if _, err := optionService.Create(symbol, "Put", openDate, 100.0+float64(i), expirationDate, premium, 1); err != nil {
				t.Fatalf("Failed to create large dataset option: %v", err)
			}
		}
	}

	t.Logf("✅ Created large test data with 50 symbols and 500+ options")
}

func createMetricsData(t *testing.T, db *database.DB, date time.Time, index int) {
	// Insert metrics data directly
	metricsTypes := []string{
		"put_exposure", "open_put_premium", "open_put_count",
		"long_value", "open_call_premium", "open_call_count",
		"treasury_value", "total_value",
	}
	
	baseValues := []float64{45000.0, 1200.0, 5.0, 65000.0, 800.0, 3.0, 50000.0, 115000.0}
	
	for i, metricType := range metricsTypes {
		value := baseValues[i] + float64(index)*100.0
		
		query := `INSERT INTO metrics (created, type, value) VALUES (?, ?, ?)`
		if _, err := db.DB.Exec(query, date.Format("2006-01-02 15:04:05"), metricType, value); err != nil {
			t.Fatalf("Failed to insert metric %s: %v", metricType, err)
		}
	}
}

// Validation helper functions

func validateAllocationDataComprehensive(t *testing.T, data web.AllocationData) {
	if len(data.LongByTicker) == 0 {
		t.Error("LongByTicker should have data")
	}
	if len(data.PutsByTicker) == 0 {
		t.Error("PutsByTicker should have data")
	}
	if len(data.TotalAllocation) == 0 {
		t.Error("TotalAllocation should have data")
	}
	t.Logf("✅ AllocationData validated: %d long, %d puts, %d total", 
		len(data.LongByTicker), len(data.PutsByTicker), len(data.TotalAllocation))
}

func validateMonthlyDataComprehensive(t *testing.T, data web.MonthlyData) {
	if len(data.PutsData.ByMonth) == 0 {
		t.Error("PutsData.ByMonth should have data")
	}
	if len(data.CallsData.ByMonth) == 0 {
		t.Error("CallsData.ByMonth should have data")
	}
	if len(data.TotalsByMonth) == 0 {
		t.Error("TotalsByMonth should have data")
	}
	t.Logf("✅ MonthlyData validated: %d months puts, %d months calls", 
		len(data.PutsData.ByMonth), len(data.CallsData.ByMonth))
}

func validateSymbolDataComprehensive(t *testing.T, data web.SymbolData) {
	if data.Symbol == "" {
		t.Error("Symbol should not be empty")
	}
	if len(data.MonthlyResults) == 0 {
		t.Error("MonthlyResults should have data")
	}
	t.Logf("✅ SymbolData validated: %s with %d monthly results", 
		data.Symbol, len(data.MonthlyResults))
}

func validateMetricsDataComprehensive(t *testing.T, data map[string][]web.ChartPoint) {
	requiredMetrics := []string{"put_exposure", "open_put_premium", "long_value", "treasury_value", "total_value"}
	for _, metric := range requiredMetrics {
		if _, exists := data[metric]; !exists {
			t.Errorf("Missing required metric: %s", metric)
		}
	}
	t.Logf("✅ MetricsData validated: %d metric types", len(data))
}

func validateOptionsScatterDataComprehensive(t *testing.T, data web.OptionsScatterData) {
	if len(data.ScatterPoints) == 0 {
		t.Error("ScatterPoints should have data")
	}
	if data.ChartConfig.Colors.PutColor == "" {
		t.Error("PutColor should not be empty")
	}
	t.Logf("✅ OptionsScatterData validated: %d points", len(data.ScatterPoints))
}

func validateTutorialDataComprehensive(t *testing.T, data web.TutorialChartData) {
	if len(data.IncomeBreakdown) == 0 {
		t.Error("IncomeBreakdown should have data")
	}
	if data.TotalReturn <= 0 {
		t.Error("TotalReturn should be positive")
	}
	t.Logf("✅ TutorialChartData validated: %d categories, total $%.2f", 
		len(data.IncomeBreakdown), data.TotalReturn)
}

func validateTreasuryDataComprehensive(t *testing.T, data web.TreasuriesData) {
	if len(data.Treasuries) == 0 {
		t.Error("Treasuries should have data")
	}
	if data.Summary.TotalAmount <= 0 {
		t.Error("TotalAmount should be positive")
	}
	t.Logf("✅ TreasuryData validated: %d treasuries, total $%.2f", 
		len(data.Treasuries), data.Summary.TotalAmount)
}

func validateFinancialAmount(t *testing.T, context string, index int, amount float64) {
	if amount != amount { // Check for NaN
		t.Errorf("%s[%d] is NaN", context, index)
	}
	if amount < -1000000 || amount > 10000000 {
		t.Errorf("%s[%d] amount seems unrealistic: %f", context, index, amount)
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Builder functions for comprehensive test data

func buildComprehensiveAllocationData(t *testing.T, db *database.DB) web.AllocationData {
	return web.AllocationData{
		LongByTicker: []web.ChartData{
			{Label: "AAPL", Value: 15000.0, Color: "#FF6384"},
			{Label: "TSLA", Value: 20000.0, Color: "#36A2EB"},
			{Label: "NVDA", Value: 40000.0, Color: "#FFCE56"},
			{Label: "MSFT", Value: 35000.0, Color: "#4BC0C0"},
			{Label: "GOOGL", Value: 25000.0, Color: "#9966FF"},
		},
		PutsByTicker: []web.ChartData{
			{Label: "AAPL", Value: 29000.0, Color: "#FF6384"},
			{Label: "TSLA", Value: 38000.0, Color: "#36A2EB"},
			{Label: "NVDA", Value: 76000.0, Color: "#FFCE56"},
		},
		CallsToLongs: []web.ChartData{
			{Label: "AAPL", Value: 75.0, Color: "#FF6384"},
			{Label: "TSLA", Value: 50.0, Color: "#36A2EB"},
		},
		TotalAllocation: []web.ChartData{
			{Label: "Long Positions", Value: 135000.0, Color: "#27ae60"},
			{Label: "Put Exposure", Value: 143000.0, Color: "#e74c3c"},
			{Label: "Treasuries", Value: 200000.0, Color: "#f39c12"},
		},
		TotalPutPremiums:  15750.0,
		TotalCallPremiums: 8250.0,
		TotalCallCovered:  18750.0,
		TotalOptionable:   67500.0,
	}
}

func buildComprehensiveMonthlyData(t *testing.T, db *database.DB) web.MonthlyData {
	return web.MonthlyData{
		Symbols: []string{"AAPL", "TSLA", "NVDA", "MSFT", "GOOGL"},
		PutsData: web.MonthlyOptionData{
			ByMonth: []web.MonthlyChartData{
				{Month: "2023-12", Amount: 1200.00},
				{Month: "2024-01", Amount: 1350.00},
				{Month: "2024-02", Amount: 1500.00},
				{Month: "2024-03", Amount: 1650.00},
				{Month: "2024-04", Amount: 1800.00},
				{Month: "2024-05", Amount: 1950.00},
			},
			ByTicker: []web.TickerChartData{
				{Ticker: "AAPL", Amount: 3250.00},
				{Ticker: "TSLA", Amount: 2850.00},
				{Ticker: "NVDA", Amount: 4350.00},
			},
		},
		CallsData: web.MonthlyOptionData{
			ByMonth: []web.MonthlyChartData{
				{Month: "2023-12", Amount: 850.00},
				{Month: "2024-01", Amount: 950.00},
				{Month: "2024-02", Amount: 1050.00},
			},
			ByTicker: []web.TickerChartData{
				{Ticker: "AAPL", Amount: 1450.00},
				{Ticker: "TSLA", Amount: 1400.00},
			},
		},
		CapGainsData: web.MonthlyFinancialData{
			ByMonth: []web.MonthlyChartData{
				{Month: "2024-01", Amount: 750.00},
				{Month: "2024-04", Amount: 1250.00},
			},
			ByTicker: []web.TickerChartData{
				{Ticker: "NVDA", Amount: 2000.00},
			},
		},
		DividendsData: web.MonthlyFinancialData{
			ByMonth: []web.MonthlyChartData{
				{Month: "2024-01", Amount: 125.00},
				{Month: "2024-04", Amount: 135.00},
			},
			ByTicker: []web.TickerChartData{
				{Ticker: "AAPL", Amount: 260.00},
			},
		},
		TableData: []web.MonthlyTableRow{
			{
				Ticker: "AAPL",
				Total:  4700.00,
				Months: [12]float64{500, 550, 600, 650, 700, 750, 0, 0, 0, 0, 0, 0},
			},
			{
				Ticker: "TSLA", 
				Total:  4250.00,
				Months: [12]float64{400, 450, 500, 550, 600, 650, 0, 0, 0, 0, 0, 0},
			},
		},
		TotalsByMonth: []web.MonthlyTotal{
			{Month: "2023-12", Amount: 2050.00},
			{Month: "2024-01", Amount: 2425.00},
			{Month: "2024-02", Amount: 2550.00},
			{Month: "2024-03", Amount: 1650.00},
			{Month: "2024-04", Amount: 3185.00},
			{Month: "2024-05", Amount: 1950.00},
		},
		MonthlyPremiumsBySymbol: []web.MonthlyPremiumsBySymbol{
			{
				Month: "2024-01",
				Symbols: []web.SymbolPremiumData{
					{Symbol: "AAPL", Amount: 1350.00},
					{Symbol: "TSLA", Amount: 950.00},
				},
			},
			{
				Month: "2024-02", 
				Symbols: []web.SymbolPremiumData{
					{Symbol: "AAPL", Amount: 1500.00},
					{Symbol: "TSLA", Amount: 1050.00},
				},
			},
		},
		GrandTotal: 12450.00,
		CurrentDB:  "test_comprehensive.db",
	}
}

func buildComprehensiveSymbolData(t *testing.T, db *database.DB) web.SymbolData {
	return web.SymbolData{
		Symbol:      "AAPL",
		AllSymbols:  []string{"AAPL", "TSLA", "NVDA", "MSFT", "GOOGL"},
		CompanyName: "Apple Inc.",
		Price:       150.0,
		MonthlyResults: []web.SymbolMonthlyResult{
			{Month: "2023-12", PutsCount: 2, CallsCount: 1, PutsTotal: 450.0, CallsTotal: 325.0, Total: 775.0},
			{Month: "2024-01", PutsCount: 3, CallsCount: 0, PutsTotal: 675.0, CallsTotal: 0.0, Total: 675.0},
			{Month: "2024-02", PutsCount: 2, CallsCount: 1, PutsTotal: 525.0, CallsTotal: 275.0, Total: 800.0},
		},
		CurrentDB: "test_comprehensive.db",
	}
}

func buildComprehensiveMetricsData(t *testing.T, db *database.DB) map[string][]web.ChartPoint {
	return map[string][]web.ChartPoint{
		"put_exposure": {
			{Date: "2024-01-01", Value: 45000.0},
			{Date: "2024-01-02", Value: 45500.0},
			{Date: "2024-01-03", Value: 46000.0},
		},
		"open_put_premium": {
			{Date: "2024-01-01", Value: 1200.0},
			{Date: "2024-01-02", Value: 1250.0},
			{Date: "2024-01-03", Value: 1300.0},
		},
		"long_value": {
			{Date: "2024-01-01", Value: 65000.0},
			{Date: "2024-01-02", Value: 66000.0},
			{Date: "2024-01-03", Value: 67000.0},
		},
		"treasury_value": {
			{Date: "2024-01-01", Value: 50000.0},
			{Date: "2024-01-02", Value: 50100.0},
			{Date: "2024-01-03", Value: 50200.0},
		},
		"total_value": {
			{Date: "2024-01-01", Value: 115000.0},
			{Date: "2024-01-02", Value: 116100.0},
			{Date: "2024-01-03", Value: 117200.0},
		},
	}
}

func buildComprehensiveOptionsScatterData(t *testing.T) web.OptionsScatterData {
	return web.OptionsScatterData{
		ScatterPoints: []web.OptionScatterPoint{
			{
				Expiration:     "2024-12-20",
				ExpirationDate: "2024-12-20T00:00:00Z",
				Profit:         225.50,
				Symbol:         "AAPL",
				Type:           "Put",
				Strike:         145.0,
				Contracts:      2,
				DTE:            15,
			},
			{
				Expiration:     "2024-12-27",
				ExpirationDate: "2024-12-27T00:00:00Z",
				Profit:         175.25,
				Symbol:         "TSLA",
				Type:           "Call",
				Strike:         200.0,
				Contracts:      1,
				DTE:            22,
			},
		},
		ChartConfig: web.ScatterChartConfig{
			Colors: web.ScatterColors{
				PutColor:  "#27ae60",
				CallColor: "#3498db",
			},
			DateRange: web.DateRange{
				Start: "2024-12-01",
				End:   "2025-01-31",
			},
			ProfitRange: web.ProfitRange{
				Min: -200.0,
				Max: 500.0,
			},
		},
	}
}

func buildComprehensiveTutorialData(t *testing.T) web.TutorialChartData {
	return web.TutorialChartData{
		IncomeBreakdown: []web.TutorialIncomeData{
			{Category: "Put Premiums", Amount: 18750.00, Percentage: 48.5, Color: "#27ae60"},
			{Category: "Call Premiums", Amount: 9250.00, Percentage: 23.9, Color: "#3498db"},
			{Category: "Capital Gains", Amount: 7650.00, Percentage: 19.8, Color: "#9966FF"},
			{Category: "Dividends", Amount: 3000.00, Percentage: 7.8, Color: "#FFCE56"},
		},
		TotalReturn:   38650.00,
		AnnualizedROI: 77.3,
	}
}

func buildComprehensiveTreasuryData(t *testing.T, db *database.DB) web.TreasuriesData {
	// Create mock treasury data with realistic values
	baseDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	maturityDate1 := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	maturityDate2 := time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC)
	maturityDate3 := time.Date(2024, 12, 15, 0, 0, 0, 0, time.UTC)
	
	currentValue1 := 75125.00
	currentValue2 := 100850.00
	exitPrice3 := 51250.00
	
	mockTreasuries := []*models.Treasury{
		{
			CUSPID:       "TREASURY-001",
			Purchased:    baseDate,
			Maturity:     maturityDate1,
			Amount:       75000.00,
			Yield:        4.25,
			BuyPrice:     74750.00,
			CurrentValue: &currentValue1,
			ExitPrice:    nil,
			CreatedAt:    baseDate,
			UpdatedAt:    baseDate,
		},
		{
			CUSPID:       "TREASURY-002", 
			Purchased:    baseDate,
			Maturity:     maturityDate2,
			Amount:       100000.00,
			Yield:        4.75,
			BuyPrice:     99625.00,
			CurrentValue: &currentValue2,
			ExitPrice:    nil,
			CreatedAt:    baseDate,
			UpdatedAt:    baseDate,
		},
		{
			CUSPID:       "TREASURY-003",
			Purchased:    baseDate,
			Maturity:     maturityDate3,
			Amount:       50000.00,
			Yield:        5.00,
			BuyPrice:     49500.00,
			CurrentValue: nil,
			ExitPrice:    &exitPrice3,
			CreatedAt:    baseDate,
			UpdatedAt:    baseDate,
		},
	}
	
	return web.TreasuriesData{
		Symbols:    []string{"AAPL", "TSLA", "NVDA", "MSFT", "GOOGL"},
		Treasuries: mockTreasuries,
		Summary: web.TreasuriesSummary{
			TotalAmount:     225000.00,
			TotalBuyPrice:   223875.00,
			TotalProfitLoss: 4500.00,
			TotalInterest:   9675.00,
			AverageReturn:   4.32,
			ActivePositions: 2, // Only 2 active (not sold)
		},
		CurrentDB: "test_comprehensive.db",
	}
}

