package test

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"

	"stonks/internal/database"
	"stonks/internal/models"
	"stonks/internal/web"

	_ "github.com/mattn/go-sqlite3"
)

// TestTreasuriesDataStructure tests the TreasuriesData structure
func TestTreasuriesDataStructure(t *testing.T) {
	// Setup test database with treasury data
	testDB := setupTreasuryTestDatabase(t)
	defer testDB.Close()

	// Build treasury data (simulating the handler)
	treasuryData := buildTestTreasuryData(t, testDB)

	// Test JSON serialization/deserialization
	jsonData, err := json.Marshal(treasuryData)
	if err != nil {
		t.Fatalf("Failed to marshal TreasuriesData: %v", err)
	}

	var unmarshalled web.TreasuriesData
	if err := json.Unmarshal(jsonData, &unmarshalled); err != nil {
		t.Fatalf("Failed to unmarshal TreasuriesData: %v", err)
	}

	// Validate basic structure
	t.Run("BasicStructure", func(t *testing.T) {
		if len(unmarshalled.Symbols) == 0 {
			t.Error("Expected symbols list")
		}
		if len(unmarshalled.Treasuries) == 0 {
			t.Error("Expected treasuries data")
		}
		if unmarshalled.CurrentDB == "" {
			t.Error("CurrentDB should not be empty")
		}
	})

	// Validate treasury records
	t.Run("TreasuryRecords", func(t *testing.T) {
		for i, treasury := range unmarshalled.Treasuries {
			if treasury.CUSPID == "" {
				t.Errorf("Treasury %d has empty CUSPID", i)
			}
			
			// CUSPID should follow standard format (for test data)
			if len(treasury.CUSPID) < 5 {
				t.Errorf("Treasury %d CUSPID seems too short: %s", i, treasury.CUSPID)
			}
			
			if treasury.Amount <= 0 {
				t.Errorf("Treasury %d amount should be positive, got %f", i, treasury.Amount)
			}
			
			if treasury.Yield <= 0 || treasury.Yield > 20.0 {
				t.Errorf("Treasury %d yield seems unrealistic: %f", i, treasury.Yield)
			}
			
			if treasury.BuyPrice <= 0 {
				t.Errorf("Treasury %d buy price should be positive, got %f", i, treasury.BuyPrice)
			}
			
			// Dates should be valid
			if treasury.Purchased.IsZero() {
				t.Errorf("Treasury %d purchased date should not be zero", i)
			}
			
			if treasury.Maturity.IsZero() {
				t.Errorf("Treasury %d maturity date should not be zero", i)
			}
			
			// Maturity should be after purchase
			if treasury.Maturity.Before(treasury.Purchased) {
				t.Errorf("Treasury %d maturity (%v) should be after purchase (%v)", 
					i, treasury.Maturity, treasury.Purchased)
			}
			
			// Test profit/loss calculation method
			profitLoss := treasury.CalculateProfitLoss()
			if profitLoss != profitLoss { // Check for NaN
				t.Errorf("Treasury %d profit/loss calculation returned NaN", i)
			}
		}
	})

	// Validate treasury summary
	t.Run("TreasurySummary", func(t *testing.T) {
		summary := unmarshalled.Summary
		
		if summary.TotalAmount <= 0 {
			t.Errorf("TotalAmount should be positive, got %f", summary.TotalAmount)
		}
		
		if summary.TotalBuyPrice <= 0 {
			t.Errorf("TotalBuyPrice should be positive, got %f", summary.TotalBuyPrice)
		}
		
		if summary.TotalInterest < 0 {
			t.Errorf("TotalInterest should be non-negative, got %f", summary.TotalInterest)
		}
		
		if summary.AverageReturn < 0 || summary.AverageReturn > 20.0 {
			t.Errorf("AverageReturn seems unrealistic: %f", summary.AverageReturn)
		}
		
		if summary.ActivePositions < 0 {
			t.Errorf("ActivePositions should be non-negative, got %d", summary.ActivePositions)
		}
		
		// ActivePositions should match number of treasuries
		if summary.ActivePositions != len(unmarshalled.Treasuries) {
			t.Errorf("ActivePositions (%d) should match number of treasuries (%d)", 
				summary.ActivePositions, len(unmarshalled.Treasuries))
		}
		
		// Validate calculated totals match individual records
		calculatedAmount := 0.0
		calculatedBuyPrice := 0.0
		calculatedProfitLoss := 0.0
		
		for _, treasury := range unmarshalled.Treasuries {
			calculatedAmount += treasury.Amount
			calculatedBuyPrice += treasury.BuyPrice
			calculatedProfitLoss += treasury.CalculateProfitLoss()
		}
		
		if math.Abs(summary.TotalAmount-calculatedAmount) > 0.01 {
			t.Errorf("Summary TotalAmount (%f) doesn't match calculated (%f)", 
				summary.TotalAmount, calculatedAmount)
		}
		
		if math.Abs(summary.TotalBuyPrice-calculatedBuyPrice) > 0.01 {
			t.Errorf("Summary TotalBuyPrice (%f) doesn't match calculated (%f)", 
				summary.TotalBuyPrice, calculatedBuyPrice)
		}
		
		if math.Abs(summary.TotalProfitLoss-calculatedProfitLoss) > 0.01 {
			t.Errorf("Summary TotalProfitLoss (%f) doesn't match calculated (%f)", 
				summary.TotalProfitLoss, calculatedProfitLoss)
		}
	})

	t.Logf("✅ TreasuriesData structure validation passed - %d treasuries, Total: $%.2f, Interest: $%.2f", 
		len(unmarshalled.Treasuries), unmarshalled.Summary.TotalAmount, unmarshalled.Summary.TotalInterest)
}

// TestTreasuriesSummaryStructure tests the TreasuriesSummary type individually
func TestTreasuriesSummaryStructure(t *testing.T) {
	// Test individual TreasuriesSummary structure
	testSummary := web.TreasuriesSummary{
		TotalAmount:     175000.00,
		TotalBuyPrice:   175000.00,
		TotalProfitLoss: 3250.50,
		TotalInterest:   7850.25,
		AverageReturn:   4.32,
		ActivePositions: 3,
	}

	// Test JSON serialization/deserialization
	jsonData, err := json.Marshal(testSummary)
	if err != nil {
		t.Fatalf("Failed to marshal TreasuriesSummary: %v", err)
	}

	var unmarshalled web.TreasuriesSummary
	if err := json.Unmarshal(jsonData, &unmarshalled); err != nil {
		t.Fatalf("Failed to unmarshal TreasuriesSummary: %v", err)
	}

	// Validate all fields
	if unmarshalled.TotalAmount != 175000.00 {
		t.Errorf("Expected TotalAmount 175000.00, got %f", unmarshalled.TotalAmount)
	}
	if unmarshalled.TotalBuyPrice != 175000.00 {
		t.Errorf("Expected TotalBuyPrice 175000.00, got %f", unmarshalled.TotalBuyPrice)
	}
	if unmarshalled.TotalProfitLoss != 3250.50 {
		t.Errorf("Expected TotalProfitLoss 3250.50, got %f", unmarshalled.TotalProfitLoss)
	}
	if unmarshalled.TotalInterest != 7850.25 {
		t.Errorf("Expected TotalInterest 7850.25, got %f", unmarshalled.TotalInterest)
	}
	if unmarshalled.AverageReturn != 4.32 {
		t.Errorf("Expected AverageReturn 4.32, got %f", unmarshalled.AverageReturn)
	}
	if unmarshalled.ActivePositions != 3 {
		t.Errorf("Expected ActivePositions 3, got %d", unmarshalled.ActivePositions)
	}

	t.Logf("✅ TreasuriesSummary type validation passed")
}

// TestTreasuryUpdateRequest tests the TreasuryUpdateRequest structure
func TestTreasuryUpdateRequest(t *testing.T) {
	// Test TreasuryUpdateRequest structure (used in API calls)
	currentValue := 101500.00
	exitPrice := 102250.00
	
	testRequest := web.TreasuryUpdateRequest{
		Purchased:    "2024-01-15",
		Maturity:     "2025-01-15",
		Amount:       100000.00,
		Yield:        4.750,
		BuyPrice:     99750.00,
		CurrentValue: &currentValue,
		ExitPrice:    &exitPrice,
	}

	// Test JSON serialization/deserialization
	jsonData, err := json.Marshal(testRequest)
	if err != nil {
		t.Fatalf("Failed to marshal TreasuryUpdateRequest: %v", err)
	}

	var unmarshalled web.TreasuryUpdateRequest
	if err := json.Unmarshal(jsonData, &unmarshalled); err != nil {
		t.Fatalf("Failed to unmarshal TreasuryUpdateRequest: %v", err)
	}

	// Validate all fields
	if unmarshalled.Purchased != "2024-01-15" {
		t.Errorf("Expected Purchased 2024-01-15, got %s", unmarshalled.Purchased)
	}
	if unmarshalled.Maturity != "2025-01-15" {
		t.Errorf("Expected Maturity 2025-01-15, got %s", unmarshalled.Maturity)
	}
	if unmarshalled.Amount != 100000.00 {
		t.Errorf("Expected Amount 100000.00, got %f", unmarshalled.Amount)
	}
	if unmarshalled.Yield != 4.750 {
		t.Errorf("Expected Yield 4.750, got %f", unmarshalled.Yield)
	}
	if unmarshalled.BuyPrice != 99750.00 {
		t.Errorf("Expected BuyPrice 99750.00, got %f", unmarshalled.BuyPrice)
	}
	
	// Test pointer fields
	if unmarshalled.CurrentValue == nil {
		t.Error("CurrentValue should not be nil")
	} else if *unmarshalled.CurrentValue != currentValue {
		t.Errorf("Expected CurrentValue %f, got %f", currentValue, *unmarshalled.CurrentValue)
	}
	
	if unmarshalled.ExitPrice == nil {
		t.Error("ExitPrice should not be nil")
	} else if *unmarshalled.ExitPrice != exitPrice {
		t.Errorf("Expected ExitPrice %f, got %f", exitPrice, *unmarshalled.ExitPrice)
	}

	// Test with nil optional fields
	testRequestNil := web.TreasuryUpdateRequest{
		Purchased:    "2024-01-15",
		Maturity:     "2025-01-15", 
		Amount:       100000.00,
		Yield:        4.750,
		BuyPrice:     99750.00,
		CurrentValue: nil,
		ExitPrice:    nil,
	}

	jsonDataNil, err := json.Marshal(testRequestNil)
	if err != nil {
		t.Fatalf("Failed to marshal TreasuryUpdateRequest with nils: %v", err)
	}

	var unmarshalledNil web.TreasuryUpdateRequest
	if err := json.Unmarshal(jsonDataNil, &unmarshalledNil); err != nil {
		t.Fatalf("Failed to unmarshal TreasuryUpdateRequest with nils: %v", err)
	}

	if unmarshalledNil.CurrentValue != nil {
		t.Error("CurrentValue should be nil")
	}
	if unmarshalledNil.ExitPrice != nil {
		t.Error("ExitPrice should be nil")
	}

	t.Logf("✅ TreasuryUpdateRequest type validation passed")
}

// TestTreasuryCalculations tests treasury business logic calculations
func TestTreasuryCalculations(t *testing.T) {
	// Setup test database with treasury
	testDB := setupTreasuryTestDatabase(t)
	defer testDB.Close()

	treasuryService := models.NewTreasuryService(testDB.DB)

	// Create test treasury with known values
	purchaseDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	maturityDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	amount := 100000.00
	yield := 4.50
	buyPrice := 99500.00
	currentValue := 101250.00
	
	treasury, err := treasuryService.CreateFull(
		"TEST-BOND-001", 
		purchaseDate, 
		maturityDate, 
		amount, 
		yield, 
		buyPrice,
		&currentValue,
		nil, // No exit price initially
	)
	if err != nil {
		t.Fatalf("Failed to create test treasury: %v", err)
	}

	t.Run("ProfitLossCalculation", func(t *testing.T) {
		// Test profit/loss calculation
		profitLoss := treasury.CalculateProfitLoss()
		expectedProfitLoss := currentValue - buyPrice // 101250.00 - 99500.00 = 1750.00
		
		if math.Abs(profitLoss-expectedProfitLoss) > 0.01 {
			t.Errorf("Expected profit/loss %f, got %f", expectedProfitLoss, profitLoss)
		}
	})

	t.Run("CurrentValueMethods", func(t *testing.T) {
		// Test HasCurrentValue method
		if !treasury.HasCurrentValue() {
			t.Error("Treasury should have current value")
		}
		
		// Test GetCurrentValue method
		retrievedCurrentValue := treasury.GetCurrentValue()
		if retrievedCurrentValue != currentValue {
			t.Errorf("Expected current value %f, got %f", currentValue, retrievedCurrentValue)
		}
	})

	t.Run("ExitPriceMethods", func(t *testing.T) {
		// Test HasExitPrice method (should be false initially)
		if treasury.HasExitPrice() {
			t.Error("Treasury should not have exit price initially")
		}
		
		// Test GetExitPrice method (should return 0 when no exit price)
		retrievedExitPrice := treasury.GetExitPrice()
		if retrievedExitPrice != 0.0 {
			t.Errorf("Expected exit price 0, got %f", retrievedExitPrice)
		}
		
		// Update with exit price
		exitPrice := 103750.00
		updatedTreasury, err := treasuryService.Update(
			treasury.CUSPID,
			&currentValue,
			&exitPrice,
		)
		if err != nil {
			t.Fatalf("Failed to update treasury with exit price: %v", err)
		}
		
		// Test after exit price is set
		if !updatedTreasury.HasExitPrice() {
			t.Error("Treasury should have exit price after update")
		}
		
		retrievedExitPrice = updatedTreasury.GetExitPrice()
		if retrievedExitPrice != exitPrice {
			t.Errorf("Expected exit price %f, got %f", exitPrice, retrievedExitPrice)
		}
		
		// Test profit/loss calculation with exit price
		profitLossWithExit := updatedTreasury.CalculateProfitLoss()
		expectedProfitLossWithExit := exitPrice - buyPrice // 103750.00 - 99500.00 = 4250.00
		
		if math.Abs(profitLossWithExit-expectedProfitLossWithExit) > 0.01 {
			t.Errorf("Expected profit/loss with exit %f, got %f", expectedProfitLossWithExit, profitLossWithExit)
		}
	})

	t.Logf("✅ Treasury calculations test passed")
}

// Helper function to setup treasury test database
func setupTreasuryTestDatabase(t *testing.T) *database.DB {
	testDBPath := "test_treasury.db"
	os.Remove(testDBPath)

	db, err := database.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create deterministic treasury test data
	createDeterministicTreasuryData(t, db)

	t.Cleanup(func() {
		db.Close()
		os.Remove(testDBPath)
	})

	return db
}

// Helper function to create deterministic treasury test data
func createDeterministicTreasuryData(t *testing.T, db *database.DB) {
	treasuryService := models.NewTreasuryService(db.DB)
	symbolService := models.NewSymbolService(db.DB)

	// Create some symbols for the treasuries data
	symbols := []string{"AAPL", "TSLA", "NVDA"}
	for _, symbol := range symbols {
		if _, err := symbolService.Create(symbol); err != nil {
			t.Fatalf("Failed to create symbol %s: %v", symbol, err)
		}
	}

	// Create deterministic treasury data
	treasuryData := []struct {
		cuspid       string
		amount       float64
		yield        float64
		buyPrice     float64
		currentValue *float64
		exitPrice    *float64
		monthsToMaturity int
	}{
		{"TREASURY-001", 50000.00, 4.25, 49750.00, floatPtr(50125.00), nil, 12},
		{"TREASURY-002", 75000.00, 4.75, 74500.00, floatPtr(75850.00), nil, 18},
		{"TREASURY-003", 100000.00, 5.00, 99250.00, floatPtr(101250.00), floatPtr(102500.00), 6}, // This one was sold
	}

	baseDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	for i, data := range treasuryData {
		purchaseDate := baseDate.AddDate(0, -3, 0) // 3 months ago
		maturityDate := baseDate.AddDate(0, data.monthsToMaturity, 0)

		_, err := treasuryService.CreateFull(
			data.cuspid,
			purchaseDate,
			maturityDate,
			data.amount,
			data.yield,
			data.buyPrice,
			data.currentValue,
			data.exitPrice,
		)
		if err != nil {
			t.Fatalf("Failed to create treasury %d: %v", i, err)
		}
	}

	t.Logf("✅ Created deterministic treasury test data for %d treasuries", len(treasuryData))
}

// Helper function to build test treasury data
func buildTestTreasuryData(t *testing.T, db *database.DB) web.TreasuriesData {
	treasuryService := models.NewTreasuryService(db.DB)
	
	// Get all treasuries from database
	treasuries, err := treasuryService.GetAll()
	if err != nil {
		t.Fatalf("Failed to get treasuries: %v", err)
	}

	// Calculate summary
	totalAmount := 0.0
	totalBuyPrice := 0.0
	totalProfitLoss := 0.0
	totalInterest := 0.0
	activePositions := len(treasuries)

	for _, treasury := range treasuries {
		totalAmount += treasury.Amount
		totalBuyPrice += treasury.BuyPrice
		totalProfitLoss += treasury.CalculateProfitLoss()
		
		// Calculate interest earned (simplified calculation)
		// In real implementation, this would be more sophisticated
		yearsHeld := 0.5 // Assume 6 months average
		totalInterest += treasury.Amount * treasury.Yield / 100.0 * yearsHeld
	}

	averageReturn := 0.0
	if totalBuyPrice > 0 {
		averageReturn = (totalInterest / totalBuyPrice) * 100.0
	}

	return web.TreasuriesData{
		Symbols:    []string{"AAPL", "TSLA", "NVDA"},
		Treasuries: treasuries,
		Summary: web.TreasuriesSummary{
			TotalAmount:     totalAmount,
			TotalBuyPrice:   totalBuyPrice,
			TotalProfitLoss: totalProfitLoss,
			TotalInterest:   totalInterest,
			AverageReturn:   averageReturn,
			ActivePositions: activePositions,
		},
		CurrentDB: "test_treasury.db",
	}
}

// Helper function to create float pointer
func floatPtr(f float64) *float64 {
	return &f
}