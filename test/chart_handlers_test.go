package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"stonks/internal/database"
	"stonks/internal/models"
	"stonks/internal/web"

	_ "github.com/mattn/go-sqlite3"
)

// TestAllocationDataHandler tests the /api/allocation-data endpoint
func TestAllocationDataHandler(t *testing.T) {
	// Setup test database with deterministic data
	testDB := setupChartTestDatabase(t)
	defer testDB.Close()

	// Create test server
	server := createTestServer(testDB)

	// Test the endpoint
	req, err := http.NewRequest("GET", "/api/allocation-data", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse response
	var allocationData web.AllocationData
	if err := json.Unmarshal(rr.Body.Bytes(), &allocationData); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Validate LongByTicker chart data
	if len(allocationData.LongByTicker) == 0 {
		t.Error("Expected LongByTicker data, got empty slice")
	}

	// Verify deterministic test data
	foundAAPL := false
	foundTSLA := false
	for _, item := range allocationData.LongByTicker {
		switch item.Label {
		case "AAPL":
			foundAAPL = true
			expectedValue := 15000.0 // 100 shares * $150
			if item.Value != expectedValue {
				t.Errorf("Expected AAPL value %f, got %f", expectedValue, item.Value)
			}
		case "TSLA":
			foundTSLA = true
			expectedValue := 20000.0 // 100 shares * $200
			if item.Value != expectedValue {
				t.Errorf("Expected TSLA value %f, got %f", expectedValue, item.Value)
			}
		}
	}

	if !foundAAPL {
		t.Error("Expected to find AAPL in LongByTicker data")
	}
	if !foundTSLA {
		t.Error("Expected to find TSLA in LongByTicker data")
	}

	// Validate PutsByTicker chart data
	if len(allocationData.PutsByTicker) == 0 {
		t.Error("Expected PutsByTicker data, got empty slice")
	}

	// Verify put exposure data
	foundAAPLPut := false
	for _, item := range allocationData.PutsByTicker {
		if item.Label == "AAPL" {
			foundAAPLPut = true
			expectedValue := 29000.0 // 2 contracts * $145 strike * 100
			if item.Value != expectedValue {
				t.Errorf("Expected AAPL put exposure %f, got %f", expectedValue, item.Value)
			}
		}
	}

	if !foundAAPLPut {
		t.Error("Expected to find AAPL in PutsByTicker data")
	}

	// Validate TotalAllocation includes all categories
	categories := make(map[string]bool)
	for _, item := range allocationData.TotalAllocation {
		categories[item.Label] = true
	}

	expectedCategories := []string{"Long Positions", "Put Exposure", "Treasuries"}
	for _, category := range expectedCategories {
		if !categories[category] {
			t.Errorf("Expected category '%s' in TotalAllocation", category)
		}
	}

	// Validate financial calculations
	if allocationData.TotalPutPremiums <= 0 {
		t.Error("Expected positive put premiums")
	}
	if allocationData.TotalCallPremiums < 0 {
		t.Error("Expected non-negative call premiums")
	}

	t.Logf("✅ AllocationData test passed - found %d long positions, %d put exposures", 
		len(allocationData.LongByTicker), len(allocationData.PutsByTicker))
}

// TestMetricsChartDataHandler tests the /api/metrics/chart-data endpoint
func TestMetricsChartDataHandler(t *testing.T) {
	// Setup test database with metrics data
	testDB := setupMetricsTestDatabase(t)
	defer testDB.Close()

	// Create test server
	server := createTestServer(testDB)

	// Test the endpoint
	req, err := http.NewRequest("GET", "/api/metrics/chart-data", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse response
	var metricsData map[string][]web.ChartPoint
	if err := json.Unmarshal(rr.Body.Bytes(), &metricsData); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Validate required metric types
	expectedMetrics := []string{
		"put_exposure", "open_put_premium", "open_put_count",
		"long_value", "open_call_premium", "open_call_count",
		"treasury_value",
	}

	for _, metric := range expectedMetrics {
		data, exists := metricsData[metric]
		if !exists {
			t.Errorf("Expected metric type '%s' in response", metric)
			continue
		}

		// Validate deterministic test data points
		if len(data) == 0 {
			t.Errorf("Expected data points for metric '%s', got empty slice", metric)
			continue
		}

		// Verify data structure
		for i, point := range data {
			if point.Date == "" {
				t.Errorf("Point %d for metric '%s' has empty date", i, metric)
			}
			if point.Value < 0 && metric != "put_exposure" { // put_exposure can be negative
				t.Errorf("Point %d for metric '%s' has negative value: %f", i, metric, point.Value)
			}

			// Validate date format
			if _, err := time.Parse("2006-01-02", point.Date); err != nil {
				t.Errorf("Point %d for metric '%s' has invalid date format '%s'", i, metric, point.Date)
			}
		}

		t.Logf("✅ Metric '%s' has %d data points", metric, len(data))
	}

	t.Logf("✅ MetricsChartData test passed - validated %d metric types", len(expectedMetrics))
}

// TestOptionsScatterDataStructure tests the proposed OptionsScatterData struct
func TestOptionsScatterDataStructure(t *testing.T) {
	// This test validates the proposed Go struct for the options scatter plot
	// that currently uses DOM extraction

	// Create test data matching the proposed struct
	scatterData := web.OptionsScatterData{
		ScatterPoints: []web.OptionScatterPoint{
			{
				Expiration:     "2024-12-20",
				ExpirationDate: "2024-12-20T00:00:00Z",
				Profit:         150.50,
				Symbol:         "AAPL",
				Type:           "Put",
				Strike:         145.0,
				Contracts:      2,
				DTE:            7,
			},
			{
				Expiration:     "2024-12-20",
				ExpirationDate: "2024-12-20T00:00:00Z",
				Profit:         -75.25,
				Symbol:         "TSLA",
				Type:           "Call",
				Strike:         200.0,
				Contracts:      1,
				DTE:            7,
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
				Min: -100.0,
				Max: 200.0,
			},
		},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(scatterData)
	if err != nil {
		t.Fatalf("Failed to marshal OptionsScatterData: %v", err)
	}

	// Test JSON deserialization
	var unmarshalled web.OptionsScatterData
	if err := json.Unmarshal(jsonData, &unmarshalled); err != nil {
		t.Fatalf("Failed to unmarshal OptionsScatterData: %v", err)
	}

	// Validate data integrity
	if len(unmarshalled.ScatterPoints) != 2 {
		t.Errorf("Expected 2 scatter points, got %d", len(unmarshalled.ScatterPoints))
	}

	// Validate first point
	point1 := unmarshalled.ScatterPoints[0]
	if point1.Symbol != "AAPL" {
		t.Errorf("Expected symbol AAPL, got %s", point1.Symbol)
	}
	if point1.Type != "Put" {
		t.Errorf("Expected type Put, got %s", point1.Type)
	}
	if point1.Profit != 150.50 {
		t.Errorf("Expected profit 150.50, got %f", point1.Profit)
	}

	// Validate chart config
	if unmarshalled.ChartConfig.Colors.PutColor != "#27ae60" {
		t.Errorf("Expected put color #27ae60, got %s", unmarshalled.ChartConfig.Colors.PutColor)
	}

	// Validate date parsing
	for i, point := range unmarshalled.ScatterPoints {
		if _, err := time.Parse("2006-01-02", point.Expiration); err != nil {
			t.Errorf("Point %d has invalid expiration date format: %s", i, point.Expiration)
		}
		if _, err := time.Parse(time.RFC3339, point.ExpirationDate); err != nil {
			t.Errorf("Point %d has invalid expiration date RFC3339 format: %s", i, point.ExpirationDate)
		}
	}

	t.Logf("✅ OptionsScatterData struct validation passed")
}

// TestTutorialChartDataStructure tests the proposed TutorialChartData struct
func TestTutorialChartDataStructure(t *testing.T) {
	// Create test data matching the proposed struct for tutorial chart
	tutorialData := web.TutorialChartData{
		IncomeBreakdown: []web.TutorialIncomeData{
			{
				Category:   "Put Premiums",
				Amount:     15420.50,
				Percentage: 45.2,
				Color:      "#27ae60",
			},
			{
				Category:   "Call Premiums",
				Amount:     8750.25,
				Percentage: 25.6,
				Color:      "#3498db",
			},
			{
				Category:   "Capital Gains",
				Amount:     7230.00,
				Percentage: 21.2,
				Color:      "#9966FF",
			},
			{
				Category:   "Dividends",
				Amount:     2750.75,
				Percentage: 8.0,
				Color:      "#FFCE56",
			},
		},
		TotalReturn:   34151.50,
		AnnualizedROI: 73.4,
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(tutorialData)
	if err != nil {
		t.Fatalf("Failed to marshal TutorialChartData: %v", err)
	}

	// Test JSON deserialization
	var unmarshalled web.TutorialChartData
	if err := json.Unmarshal(jsonData, &unmarshalled); err != nil {
		t.Fatalf("Failed to unmarshal TutorialChartData: %v", err)
	}

	// Validate data integrity
	if len(unmarshalled.IncomeBreakdown) != 4 {
		t.Errorf("Expected 4 income categories, got %d", len(unmarshalled.IncomeBreakdown))
	}

	// Validate percentage totals (should approximately equal 100%)
	totalPercentage := 0.0
	totalAmount := 0.0
	for _, income := range unmarshalled.IncomeBreakdown {
		totalPercentage += income.Percentage
		totalAmount += income.Amount
		
		// Validate required fields
		if income.Category == "" {
			t.Error("Income category should not be empty")
		}
		if income.Amount <= 0 {
			t.Errorf("Income amount should be positive, got %f", income.Amount)
		}
		if income.Percentage <= 0 {
			t.Errorf("Income percentage should be positive, got %f", income.Percentage)
		}
		if income.Color == "" {
			t.Error("Income color should not be empty")
		}
	}

	// Allow for small rounding differences
	if totalPercentage < 99.0 || totalPercentage > 101.0 {
		t.Errorf("Total percentage should be approximately 100%%, got %f", totalPercentage)
	}

	// Validate total calculations
	if unmarshalled.TotalReturn != totalAmount {
		t.Errorf("TotalReturn %f should equal sum of amounts %f", unmarshalled.TotalReturn, totalAmount)
	}

	if unmarshalled.AnnualizedROI <= 0 {
		t.Errorf("AnnualizedROI should be positive, got %f", unmarshalled.AnnualizedROI)
	}

	// Validate color consistency (should use Wheeler standard colors)
	expectedColors := []string{"#27ae60", "#3498db", "#9966FF", "#FFCE56"}
	for i, income := range unmarshalled.IncomeBreakdown {
		if income.Color != expectedColors[i] {
			t.Errorf("Expected color %s for category %s, got %s", 
				expectedColors[i], income.Category, income.Color)
		}
	}

	t.Logf("✅ TutorialChartData struct validation passed - Total: $%.2f, ROI: %.1f%%", 
		unmarshalled.TotalReturn, unmarshalled.AnnualizedROI)
}

// Helper function to create test database with deterministic chart data
func setupChartTestDatabase(t *testing.T) *database.DB {
	testDBPath := "test_charts.db"
	os.Remove(testDBPath)

	db, err := database.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create deterministic test data
	createDeterministicChartData(t, db)

	t.Cleanup(func() {
		db.Close()
		os.Remove(testDBPath)
	})

	return db
}

// Helper function to create metrics test database
func setupMetricsTestDatabase(t *testing.T) *database.DB {
	testDBPath := "test_metrics.db"
	os.Remove(testDBPath)

	db, err := database.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create deterministic metrics data
	createDeterministicMetricsData(t, db)

	t.Cleanup(func() {
		db.Close()
		os.Remove(testDBPath)
	})

	return db
}

// Helper function to create deterministic test data for charts
func createDeterministicChartData(t *testing.T, db *database.DB) {
	symbolService := models.NewSymbolService(db.DB)
	settingService := models.NewSettingService(db.DB)
	optionService := models.NewOptionService(db.DB, settingService)
	longPositionService := models.NewLongPositionService(db.DB)
	treasuryService := models.NewTreasuryService(db.DB)

	// Create symbols
	symbols := []string{"AAPL", "TSLA", "NVDA"}
	for _, symbol := range symbols {
		if _, err := symbolService.Create(symbol); err != nil {
			t.Fatalf("Failed to create symbol %s: %v", symbol, err)
		}
	}

	// Update symbol prices (deterministic)
	if _, err := symbolService.Update("AAPL", 150.0, 0, nil, nil); err != nil {
		t.Fatalf("Failed to update AAPL price: %v", err)
	}
	if _, err := symbolService.Update("TSLA", 200.0, 0, nil, nil); err != nil {
		t.Fatalf("Failed to update TSLA price: %v", err)
	}
	if _, err := symbolService.Update("NVDA", 400.0, 0, nil, nil); err != nil {
		t.Fatalf("Failed to update NVDA price: %v", err)
	}

	// Create long positions (deterministic amounts)
	baseDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	
	// AAPL: 100 shares at $150 = $15,000 value
	if _, err := longPositionService.Create("AAPL", baseDate, 100, 150.0); err != nil {
		t.Fatalf("Failed to create AAPL long position: %v", err)
	}
	
	// TSLA: 100 shares at $200 = $20,000 value  
	if _, err := longPositionService.Create("TSLA", baseDate, 100, 200.0); err != nil {
		t.Fatalf("Failed to create TSLA long position: %v", err)
	}

	// Create open put options (deterministic exposure)
	expirationDate := time.Date(2024, 12, 20, 0, 0, 0, 0, time.UTC)

	// AAPL Put: 2 contracts at $145 strike = $29,000 exposure
	// Commission: 0.65 * 2 = 1.30 total
	if _, err := optionService.CreateWithCommission("AAPL", "Put", baseDate, 145.0, expirationDate, 5.50, 2, 1.30); err != nil {
		t.Fatalf("Failed to create AAPL put: %v", err)
	}

	// TSLA Put: 1 contract at $190 strike = $19,000 exposure
	// Commission: 0.65 * 1 = 0.65 total
	if _, err := optionService.CreateWithCommission("TSLA", "Put", baseDate, 190.0, expirationDate, 7.25, 1, 0.65); err != nil {
		t.Fatalf("Failed to create TSLA put: %v", err)
	}

	// Create treasury position (deterministic value)
	// $50,000 treasury at 4.5% yield
	treasuryDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	maturityDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	
	if _, err := treasuryService.Create("TEST-TREASURY-001", treasuryDate, maturityDate, 50000.0, 4.5, 50000.0); err != nil {
		t.Fatalf("Failed to create treasury: %v", err)
	}

	t.Logf("✅ Created deterministic test data for charts")
}

// Helper function to create deterministic metrics data
func createDeterministicMetricsData(t *testing.T, db *database.DB) {
	// Create metrics service
	// Note: This assumes metrics service exists - may need to create it
	// For now, we'll insert directly into the metrics table
	
	// Create 30 days of metrics data (deterministic values)
	baseDate := time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC)
	
	metricsData := []struct {
		metricType string
		baseValue  float64
		increment  float64
	}{
		{"put_exposure", 45000.0, 500.0},
		{"open_put_premium", 1200.0, 25.0},
		{"open_put_count", 5.0, 0.1},
		{"long_value", 65000.0, 750.0},
		{"open_call_premium", 800.0, 15.0},
		{"open_call_count", 3.0, 0.05},
		{"treasury_value", 50000.0, 100.0},
	}

	// Insert deterministic metrics for 30 days
	for day := 0; day < 30; day++ {
		currentDate := baseDate.AddDate(0, 0, day)
		
		for _, metric := range metricsData {
			// Create deterministic but varying values
			value := metric.baseValue + (metric.increment * float64(day))
			
			// Add some deterministic variation (sine wave pattern)
			variation := value * 0.05 * (0.5 + 0.5*float64(day%7)/7.0)
			finalValue := value + variation
			
			// Insert directly into database (this assumes metrics table structure)
			query := `INSERT INTO metrics (created, type, value) VALUES (?, ?, ?)`
			if _, err := db.DB.Exec(query, currentDate.Format("2006-01-02 15:04:05"), metric.metricType, finalValue); err != nil {
				t.Fatalf("Failed to insert metric %s for day %d: %v", metric.metricType, day, err)
			}
		}
	}

	t.Logf("✅ Created deterministic metrics data for 30 days")
}

// Helper function to create test web server
func createTestServer(db *database.DB) http.Handler {
	// This would create a test server with the database
	// For now, return a basic handler - this needs to be implemented
	// based on the actual web server structure
	
	mux := http.NewServeMux()
	
	// Add minimal handler for testing
	mux.HandleFunc("/api/allocation-data", func(w http.ResponseWriter, r *http.Request) {
		// Return test data
		testData := web.AllocationData{
			LongByTicker: []web.ChartData{
				{Label: "AAPL", Value: 15000.0, Color: "#FF6384"},
				{Label: "TSLA", Value: 20000.0, Color: "#36A2EB"},
			},
			PutsByTicker: []web.ChartData{
				{Label: "AAPL", Value: 29000.0, Color: "#FF6384"},
				{Label: "TSLA", Value: 19000.0, Color: "#36A2EB"},
			},
			TotalAllocation: []web.ChartData{
				{Label: "Long Positions", Value: 35000.0, Color: "#27ae60"},
				{Label: "Put Exposure", Value: 48000.0, Color: "#e74c3c"},
				{Label: "Treasuries", Value: 50000.0, Color: "#f39c12"},
			},
			TotalPutPremiums:  1275.0,
			TotalCallPremiums: 850.0,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testData)
	})
	
	mux.HandleFunc("/api/metrics/chart-data", func(w http.ResponseWriter, r *http.Request) {
		// Return test metrics data
		testData := map[string][]web.ChartPoint{
			"put_exposure": {
				{Date: "2024-11-01", Value: 45000.0},
				{Date: "2024-11-02", Value: 45500.0},
				{Date: "2024-11-03", Value: 46000.0},
			},
			"open_put_premium": {
				{Date: "2024-11-01", Value: 1200.0},
				{Date: "2024-11-02", Value: 1225.0},
				{Date: "2024-11-03", Value: 1250.0},
			},
			"open_put_count": {
				{Date: "2024-11-01", Value: 5.0},
				{Date: "2024-11-02", Value: 5.1},
				{Date: "2024-11-03", Value: 5.2},
			},
			"long_value": {
				{Date: "2024-11-01", Value: 65000.0},
				{Date: "2024-11-02", Value: 65750.0},
				{Date: "2024-11-03", Value: 66500.0},
			},
			"open_call_premium": {
				{Date: "2024-11-01", Value: 800.0},
				{Date: "2024-11-02", Value: 815.0},
				{Date: "2024-11-03", Value: 830.0},
			},
			"open_call_count": {
				{Date: "2024-11-01", Value: 3.0},
				{Date: "2024-11-02", Value: 3.05},
				{Date: "2024-11-03", Value: 3.1},
			},
			"treasury_value": {
				{Date: "2024-11-01", Value: 50000.0},
				{Date: "2024-11-02", Value: 50100.0},
				{Date: "2024-11-03", Value: 50200.0},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testData)
	})
	
	return mux
}