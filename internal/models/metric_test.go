package models

import (
	"stonks/internal/database"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestMetricService_ComprehensiveSnapshot(t *testing.T) {
	// Setup test database
	testDB, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer testDB.Close()

	// Create services
	metricService := NewMetricService(testDB.DB)
	treasuryService := NewTreasuryService(testDB.DB)
	symbolService := NewSymbolService(testDB.DB)
	positionService := NewLongPositionService(testDB.DB)
	settingService := NewSettingService(testDB.DB)
	optionService := NewOptionService(testDB.DB, settingService)

	// Create test treasury data with specific dates spanning several months
	testDate1 := time.Now().AddDate(0, -4, 0) // 4 months ago
	testDate2 := time.Now().AddDate(0, -2, 0) // 2 months ago  
	testDate3 := time.Now().AddDate(0, 0, -30) // 30 days ago

	// Treasury 1: Active for all 3 dates (purchased before testDate1)
	maturityDate1 := testDate1.AddDate(0, 3, 0) // 3 months later
	_, err = treasuryService.Create("TEST001", testDate1.AddDate(0, 0, -1), maturityDate1, 1000.0, 2.5, 1000.0)
	if err != nil {
		t.Fatalf("Failed to create test treasury 1: %v", err)
	}

	// Treasury 2: Active starting from testDate2 (purchased on testDate2)
	maturityDate2 := testDate2.AddDate(0, 6, 0) // 6 months later
	_, err = treasuryService.Create("TEST002", testDate2, maturityDate2, 2000.0, 3.0, 2000.0)
	if err != nil {
		t.Fatalf("Failed to create test treasury 2: %v", err)
	}

	// Treasury 3: Exited before testDate3 (purchased before testDate1, will be updated to have exit price)
	maturityDate3 := testDate1.AddDate(0, 12, 0) // 12 months later
	_, err = treasuryService.Create("TEST003", testDate1.AddDate(0, 0, -2), maturityDate3, 500.0, 2.8, 500.0)
	if err != nil {
		t.Fatalf("Failed to create test treasury 3: %v", err)
	}

	// Set exit price for Treasury 3 to simulate it being sold on testDate2
	exitPrice := 500.0
	_, err = treasuryService.Update("TEST003", nil, &exitPrice)
	if err != nil {
		t.Fatalf("Failed to update treasury 3 with exit price: %v", err)
	}

	// Create symbols for long positions
	_, err = symbolService.Create("AAPL")
	if err != nil {
		t.Fatalf("Failed to create AAPL symbol: %v", err)
	}
	_, err = symbolService.Create("TSLA")
	if err != nil {
		t.Fatalf("Failed to create TSLA symbol: %v", err)
	}

	// Create test long positions with specific dates
	// Position 1: AAPL opened before testDate1, still active (100 shares * $150 = $15000)
	_, err = positionService.Create("AAPL", testDate1.AddDate(0, 0, -1), 100, 150.0)
	if err != nil {
		t.Fatalf("Failed to create AAPL position: %v", err)
	}

	// Position 2: TSLA opened on testDate2, still active (50 shares * $250 = $12500)
	_, err = positionService.Create("TSLA", testDate2, 50, 250.0)
	if err != nil {
		t.Fatalf("Failed to create TSLA position: %v", err)
	}

	// Position 3: AAPL opened before testDate1, closed on testDate2 (25 shares * $150 = $3750)
	closedPosition, err := positionService.Create("AAPL", testDate1.AddDate(0, 0, -2), 25, 150.0)
	if err != nil {
		t.Fatalf("Failed to create closed AAPL position: %v", err)
	}
	// Close the position on testDate2
	err = positionService.CloseByID(closedPosition.ID, testDate2, 155.0)
	if err != nil {
		t.Fatalf("Failed to close AAPL position: %v", err)
	}

	// Create test options with specific dates
	expirationDate := testDate3.AddDate(0, 1, 0) // 1 month after testDate3
	
	// Put option 1: AAPL opened before testDate1, still active (strike $140, 2 contracts, exposure = 140 * 2 * 100 = $28000)
	// Commission: 0.65 * 2 = 1.30 total
	_, err = optionService.CreateWithCommission("AAPL", "Put", testDate1.AddDate(0, 0, -1), 140.0, expirationDate, 3.50, 2, 1.30)
	if err != nil {
		t.Fatalf("Failed to create AAPL put option: %v", err)
	}

	// Put option 2: TSLA opened on testDate2, still active (strike $230, 1 contract, exposure = 230 * 1 * 100 = $23000)
	// Commission: 0.65 * 1 = 0.65 total
	_, err = optionService.CreateWithCommission("TSLA", "Put", testDate2, 230.0, expirationDate, 5.00, 1, 0.65)
	if err != nil {
		t.Fatalf("Failed to create TSLA put option: %v", err)
	}

	// Put option 3: AAPL opened before testDate1, closed on testDate2 (strike $135, 1 contract, exposure = 135 * 1 * 100 = $13500)
	// Commission: 0.65 * 1 = 0.65 total (will be doubled on close)
	closedPutOption, err := optionService.CreateWithCommission("AAPL", "Put", testDate1.AddDate(0, 0, -2), 135.0, expirationDate, 2.75, 1, 0.65)
	if err != nil {
		t.Fatalf("Failed to create closed AAPL put option: %v", err)
	}
	// Close the put option on testDate2
	err = optionService.CloseByID(closedPutOption.ID, testDate2, 1.50)
	if err != nil {
		t.Fatalf("Failed to close AAPL put option: %v", err)
	}

	// Create test call options with specific dates
	// Call option 1: AAPL opened before testDate1, still active (premium $2.25, 3 contracts, premium = 2.25 * 3 * 100 = $675)
	// Commission: 0.65 * 3 = 1.95 total
	_, err = optionService.CreateWithCommission("AAPL", "Call", testDate1.AddDate(0, 0, -1), 160.0, expirationDate, 2.25, 3, 1.95)
	if err != nil {
		t.Fatalf("Failed to create AAPL call option: %v", err)
	}

	// Call option 2: TSLA opened on testDate2, still active (premium $4.75, 2 contracts, premium = 4.75 * 2 * 100 = $950)
	// Commission: 0.65 * 2 = 1.30 total
	_, err = optionService.CreateWithCommission("TSLA", "Call", testDate2, 270.0, expirationDate, 4.75, 2, 1.30)
	if err != nil {
		t.Fatalf("Failed to create TSLA call option: %v", err)
	}

	// Call option 3: AAPL opened before testDate1, closed on testDate2 (premium $1.85, 1 contract, premium = 1.85 * 1 * 100 = $185)
	// Commission: 0.65 * 1 = 0.65 total (will be doubled on close)
	closedCallOption, err := optionService.CreateWithCommission("AAPL", "Call", testDate1.AddDate(0, 0, -2), 155.0, expirationDate, 1.85, 1, 0.65)
	if err != nil {
		t.Fatalf("Failed to create closed AAPL call option: %v", err)
	}
	// Close the call option on testDate2
	err = optionService.CloseByID(closedCallOption.ID, testDate2, 0.95)
	if err != nil {
		t.Fatalf("Failed to close AAPL call option: %v", err)
	}

	// Test ComprehensiveSnapshot for a 150-day range (should include our test dates spanning months)
	err = metricService.ComprehensiveSnapshot(150)
	if err != nil {
		t.Fatalf("ComprehensiveSnapshot failed: %v", err)
	}

	// Verify metrics were created for TreasuryValue
	treasuryMetrics, err := metricService.GetByType(TreasuryValue)
	if err != nil {
		t.Fatalf("Failed to get treasury metrics: %v", err)
	}

	if len(treasuryMetrics) == 0 {
		t.Fatal("No treasury metrics were created")
	}

	// We should have 150 treasury value metrics (one for each day in the range)
	if len(treasuryMetrics) != 150 {
		t.Errorf("Expected 150 treasury metrics, got %d", len(treasuryMetrics))
	}

	// Find metrics for our specific test dates and verify values
	metricsByDate := make(map[string]*Metric)
	for _, metric := range treasuryMetrics {
		dateKey := metric.Created.Format("2006-01-02")
		metricsByDate[dateKey] = metric
	}

	// Expected values with corrected Treasury logic (only active treasuries, exit_price IS NULL):
	// testDate1: 1000 (Treasury 1 active, Treasury 3 excluded due to exit_price, Treasury 2 not yet purchased)
	// testDate2: 1000 + 2000 = 3000 (Treasury 1 + Treasury 2 active, Treasury 3 excluded due to exit_price)
	// testDate3: 1000 + 2000 = 3000 (Treasury 1 + Treasury 2 active, Treasury 3 excluded due to exit_price)

	testDate1Key := testDate1.Format("2006-01-02")
	if metric, exists := metricsByDate[testDate1Key]; exists {
		expectedValue := 1000.0 // Treasury 1 (1000) active, Treasury 3 excluded due to exit_price, Treasury 2 not yet purchased
		if metric.Value != expectedValue {
			t.Errorf("Expected treasury value %f for %s, got %f", expectedValue, testDate1Key, metric.Value)
		}
	} else {
		t.Errorf("Missing treasury metric for date %s", testDate1Key)
	}

	testDate2Key := testDate2.Format("2006-01-02")
	if metric, exists := metricsByDate[testDate2Key]; exists {
		expectedValue := 3000.0 // Treasury 1 (1000) + Treasury 2 (2000) active, Treasury 3 excluded due to exit_price
		if metric.Value != expectedValue {
			t.Errorf("Expected treasury value %f for %s, got %f", expectedValue, testDate2Key, metric.Value)
		}
	} else {
		t.Errorf("Missing treasury metric for date %s", testDate2Key)
	}

	testDate3Key := testDate3.Format("2006-01-02")
	if metric, exists := metricsByDate[testDate3Key]; exists {
		expectedValue := 3000.0 // Treasury 1 (1000) + Treasury 2 (2000) active, Treasury 3 excluded due to exit_price
		if metric.Value != expectedValue {
			t.Errorf("Expected treasury value %f for %s, got %f", expectedValue, testDate3Key, metric.Value)
		}
	} else {
		t.Errorf("Missing treasury metric for date %s", testDate3Key)
	}

	// Verify Long Value metrics were created
	longValueMetrics, err := metricService.GetByType(LongValue)
	if err != nil {
		t.Fatalf("Failed to get long value metrics: %v", err)
	}

	if len(longValueMetrics) != 150 {
		t.Errorf("Expected 150 long value metrics, got %d", len(longValueMetrics))
	}

	// Build map for long value metrics by date
	longValueByDate := make(map[string]*Metric)
	for _, metric := range longValueMetrics {
		dateKey := metric.Created.Format("2006-01-02")
		longValueByDate[dateKey] = metric
	}

	// Expected long values with new historical logic (include positions active on that date):
	// testDate1: 15000 + 3750 = 18750 (AAPL 100 shares + AAPL 25 shares, both active on testDate1)
	// testDate2: 15000 + 12500 = 27500 (AAPL 100 shares + TSLA 50 shares, closed position was closed on testDate2 so excluded)
	// testDate3: 15000 + 12500 = 27500 (AAPL 100 shares + TSLA 50 shares)

	if metric, exists := longValueByDate[testDate1Key]; exists {
		expectedValue := 18750.0 // AAPL 100*150 + AAPL 25*150 (both active on testDate1)
		if metric.Value != expectedValue {
			t.Errorf("Expected long value %f for %s, got %f", expectedValue, testDate1Key, metric.Value)
		}
	} else {
		t.Errorf("Missing long value metric for date %s", testDate1Key)
	}

	if metric, exists := longValueByDate[testDate2Key]; exists {
		expectedValue := 27500.0 // AAPL 100 * 150 + TSLA 50 * 250
		if metric.Value != expectedValue {
			t.Errorf("Expected long value %f for %s, got %f", expectedValue, testDate2Key, metric.Value)
		}
	} else {
		t.Errorf("Missing long value metric for date %s", testDate2Key)
	}

	if metric, exists := longValueByDate[testDate3Key]; exists {
		expectedValue := 27500.0 // AAPL 100 * 150 + TSLA 50 * 250
		if metric.Value != expectedValue {
			t.Errorf("Expected long value %f for %s, got %f", expectedValue, testDate3Key, metric.Value)
		}
	} else {
		t.Errorf("Missing long value metric for date %s", testDate3Key)
	}

	// Verify Long Count metrics were created
	longCountMetrics, err := metricService.GetByType(LongCount)
	if err != nil {
		t.Fatalf("Failed to get long count metrics: %v", err)
	}

	if len(longCountMetrics) != 150 {
		t.Errorf("Expected 150 long count metrics, got %d", len(longCountMetrics))
	}

	// Build map for long count metrics by date
	longCountByDate := make(map[string]*Metric)
	for _, metric := range longCountMetrics {
		dateKey := metric.Created.Format("2006-01-02")
		longCountByDate[dateKey] = metric
	}

	// Expected long counts with new historical logic (include positions active on that date):
	// testDate1: 2 (AAPL position + closed AAPL position, both active on testDate1)
	// testDate2: 2 (AAPL position + TSLA position, closed position was closed on testDate2 so excluded)
	// testDate3: 2 (AAPL position + TSLA position)

	if metric, exists := longCountByDate[testDate1Key]; exists {
		expectedValue := 2.0 // AAPL position + closed AAPL position (both active on testDate1)
		if metric.Value != expectedValue {
			t.Errorf("Expected long count %f for %s, got %f", expectedValue, testDate1Key, metric.Value)
		}
	} else {
		t.Errorf("Missing long count metric for date %s", testDate1Key)
	}

	if metric, exists := longCountByDate[testDate2Key]; exists {
		expectedValue := 2.0 // AAPL + TSLA positions
		if metric.Value != expectedValue {
			t.Errorf("Expected long count %f for %s, got %f", expectedValue, testDate2Key, metric.Value)
		}
	} else {
		t.Errorf("Missing long count metric for date %s", testDate2Key)
	}

	if metric, exists := longCountByDate[testDate3Key]; exists {
		expectedValue := 2.0 // AAPL + TSLA positions
		if metric.Value != expectedValue {
			t.Errorf("Expected long count %f for %s, got %f", expectedValue, testDate3Key, metric.Value)
		}
	} else {
		t.Errorf("Missing long count metric for date %s", testDate3Key)
	}

	// Verify Put Exposure metrics were created
	putExposureMetrics, err := metricService.GetByType(PutExposure)
	if err != nil {
		t.Fatalf("Failed to get put exposure metrics: %v", err)
	}

	if len(putExposureMetrics) != 150 {
		t.Errorf("Expected 150 put exposure metrics, got %d", len(putExposureMetrics))
	}

	// Build map for put exposure metrics by date
	putExposureByDate := make(map[string]*Metric)
	for _, metric := range putExposureMetrics {
		dateKey := metric.Created.Format("2006-01-02")
		putExposureByDate[dateKey] = metric
	}

	// Expected put exposures with new historical logic (include puts active on that date):
	// testDate1: 28000 + 13500 = 41500 (AAPL put + closed AAPL put, both active on testDate1)
	// testDate2: 28000 + 23000 = 51000 (AAPL put + TSLA put, closed put was closed on testDate2 so excluded)
	// testDate3: 28000 + 23000 = 51000 (AAPL put + TSLA put)

	if metric, exists := putExposureByDate[testDate1Key]; exists {
		expectedValue := 41500.0 // AAPL put 140*2*100 + closed AAPL put 135*1*100
		if metric.Value != expectedValue {
			t.Errorf("Expected put exposure %f for %s, got %f", expectedValue, testDate1Key, metric.Value)
		}
	} else {
		t.Errorf("Missing put exposure metric for date %s", testDate1Key)
	}

	if metric, exists := putExposureByDate[testDate2Key]; exists {
		expectedValue := 51000.0 // AAPL put 140*2*100 + TSLA put 230*1*100
		if metric.Value != expectedValue {
			t.Errorf("Expected put exposure %f for %s, got %f", expectedValue, testDate2Key, metric.Value)
		}
	} else {
		t.Errorf("Missing put exposure metric for date %s", testDate2Key)
	}

	if metric, exists := putExposureByDate[testDate3Key]; exists {
		expectedValue := 51000.0 // AAPL put 140*2*100 + TSLA put 230*1*100
		if metric.Value != expectedValue {
			t.Errorf("Expected put exposure %f for %s, got %f", expectedValue, testDate3Key, metric.Value)
		}
	} else {
		t.Errorf("Missing put exposure metric for date %s", testDate3Key)
	}

	// Verify OpenPutPremium metrics were created
	openPutPremiumMetrics, err := metricService.GetByType(OpenPutPremium)
	if err != nil {
		t.Fatalf("Failed to get open put premium metrics: %v", err)
	}

	if len(openPutPremiumMetrics) != 150 {
		t.Errorf("Expected 150 open put premium metrics, got %d", len(openPutPremiumMetrics))
	}

	// Build map for open put premium metrics by date
	openPutPremiumByDate := make(map[string]*Metric)
	for _, metric := range openPutPremiumMetrics {
		dateKey := metric.Created.Format("2006-01-02")
		openPutPremiumByDate[dateKey] = metric
	}

	// Expected open put premiums with new historical logic (include puts active on that date):
	// testDate1: 700 + 275 = 975 (AAPL put 3.50*2*100 + closed AAPL put 2.75*1*100, both active on testDate1)
	// testDate2: 700 + 500 = 1200 (AAPL put 3.50*2*100 + TSLA put 5.00*1*100, closed put was closed on testDate2 so excluded)
	// testDate3: 700 + 500 = 1200 (AAPL put + TSLA put)

	if metric, exists := openPutPremiumByDate[testDate1Key]; exists {
		expectedValue := 975.0 // AAPL put 3.50*2*100 + closed AAPL put 2.75*1*100
		if metric.Value != expectedValue {
			t.Errorf("Expected open put premium %f for %s, got %f", expectedValue, testDate1Key, metric.Value)
		}
	} else {
		t.Errorf("Missing open put premium metric for date %s", testDate1Key)
	}

	if metric, exists := openPutPremiumByDate[testDate2Key]; exists {
		expectedValue := 1200.0 // AAPL put 3.50*2*100 + TSLA put 5.00*1*100
		if metric.Value != expectedValue {
			t.Errorf("Expected open put premium %f for %s, got %f", expectedValue, testDate2Key, metric.Value)
		}
	} else {
		t.Errorf("Missing open put premium metric for date %s", testDate2Key)
	}

	if metric, exists := openPutPremiumByDate[testDate3Key]; exists {
		expectedValue := 1200.0 // AAPL put 3.50*2*100 + TSLA put 5.00*1*100
		if metric.Value != expectedValue {
			t.Errorf("Expected open put premium %f for %s, got %f", expectedValue, testDate3Key, metric.Value)
		}
	} else {
		t.Errorf("Missing open put premium metric for date %s", testDate3Key)
	}

	// Verify OpenPutCount metrics were created
	openPutCountMetrics, err := metricService.GetByType(OpenPutCount)
	if err != nil {
		t.Fatalf("Failed to get open put count metrics: %v", err)
	}

	if len(openPutCountMetrics) != 150 {
		t.Errorf("Expected 150 open put count metrics, got %d", len(openPutCountMetrics))
	}

	// Build map for open put count metrics by date
	openPutCountByDate := make(map[string]*Metric)
	for _, metric := range openPutCountMetrics {
		dateKey := metric.Created.Format("2006-01-02")
		openPutCountByDate[dateKey] = metric
	}

	// Expected open put counts with new historical logic (include puts active on that date):
	// testDate1: 2 (AAPL put + closed AAPL put, both active on testDate1)
	// testDate2: 2 (AAPL put + TSLA put, closed put was closed on testDate2 so excluded)
	// testDate3: 2 (AAPL put + TSLA put)

	if metric, exists := openPutCountByDate[testDate1Key]; exists {
		expectedValue := 2.0 // AAPL put + closed AAPL put (both active on testDate1)
		if metric.Value != expectedValue {
			t.Errorf("Expected open put count %f for %s, got %f", expectedValue, testDate1Key, metric.Value)
		}
	} else {
		t.Errorf("Missing open put count metric for date %s", testDate1Key)
	}

	if metric, exists := openPutCountByDate[testDate2Key]; exists {
		expectedValue := 2.0 // AAPL put + TSLA put
		if metric.Value != expectedValue {
			t.Errorf("Expected open put count %f for %s, got %f", expectedValue, testDate2Key, metric.Value)
		}
	} else {
		t.Errorf("Missing open put count metric for date %s", testDate2Key)
	}

	if metric, exists := openPutCountByDate[testDate3Key]; exists {
		expectedValue := 2.0 // AAPL put + TSLA put
		if metric.Value != expectedValue {
			t.Errorf("Expected open put count %f for %s, got %f", expectedValue, testDate3Key, metric.Value)
		}
	} else {
		t.Errorf("Missing open put count metric for date %s", testDate3Key)
	}

	// Verify OpenCallPremium metrics were created
	openCallPremiumMetrics, err := metricService.GetByType(OpenCallPremium)
	if err != nil {
		t.Fatalf("Failed to get open call premium metrics: %v", err)
	}

	if len(openCallPremiumMetrics) != 150 {
		t.Errorf("Expected 150 open call premium metrics, got %d", len(openCallPremiumMetrics))
	}

	// Build map for open call premium metrics by date
	openCallPremiumByDate := make(map[string]*Metric)
	for _, metric := range openCallPremiumMetrics {
		dateKey := metric.Created.Format("2006-01-02")
		openCallPremiumByDate[dateKey] = metric
	}

	// Expected open call premiums with new historical logic (include calls active on that date):
	// testDate1: 675 + 185 = 860 (AAPL call 2.25*3*100 + closed AAPL call 1.85*1*100, both active on testDate1)
	// testDate2: 675 + 950 = 1625 (AAPL call 2.25*3*100 + TSLA call 4.75*2*100, closed call was closed on testDate2 so excluded)
	// testDate3: 675 + 950 = 1625 (AAPL call + TSLA call)

	if metric, exists := openCallPremiumByDate[testDate1Key]; exists {
		expectedValue := 860.0 // AAPL call 2.25*3*100 + closed AAPL call 1.85*1*100
		if metric.Value != expectedValue {
			t.Errorf("Expected open call premium %f for %s, got %f", expectedValue, testDate1Key, metric.Value)
		}
	} else {
		t.Errorf("Missing open call premium metric for date %s", testDate1Key)
	}

	if metric, exists := openCallPremiumByDate[testDate2Key]; exists {
		expectedValue := 1625.0 // AAPL call 2.25*3*100 + TSLA call 4.75*2*100
		if metric.Value != expectedValue {
			t.Errorf("Expected open call premium %f for %s, got %f", expectedValue, testDate2Key, metric.Value)
		}
	} else {
		t.Errorf("Missing open call premium metric for date %s", testDate2Key)
	}

	if metric, exists := openCallPremiumByDate[testDate3Key]; exists {
		expectedValue := 1625.0 // AAPL call 2.25*3*100 + TSLA call 4.75*2*100
		if metric.Value != expectedValue {
			t.Errorf("Expected open call premium %f for %s, got %f", expectedValue, testDate3Key, metric.Value)
		}
	} else {
		t.Errorf("Missing open call premium metric for date %s", testDate3Key)
	}

	// Verify OpenCallCount metrics were created
	openCallCountMetrics, err := metricService.GetByType(OpenCallCount)
	if err != nil {
		t.Fatalf("Failed to get open call count metrics: %v", err)
	}

	if len(openCallCountMetrics) != 150 {
		t.Errorf("Expected 150 open call count metrics, got %d", len(openCallCountMetrics))
	}

	// Build map for open call count metrics by date
	openCallCountByDate := make(map[string]*Metric)
	for _, metric := range openCallCountMetrics {
		dateKey := metric.Created.Format("2006-01-02")
		openCallCountByDate[dateKey] = metric
	}

	// Expected open call counts with new historical logic (include calls active on that date):
	// testDate1: 2 (AAPL call + closed AAPL call, both active on testDate1)
	// testDate2: 2 (AAPL call + TSLA call, closed call was closed on testDate2 so excluded)
	// testDate3: 2 (AAPL call + TSLA call)

	if metric, exists := openCallCountByDate[testDate1Key]; exists {
		expectedValue := 2.0 // AAPL call + closed AAPL call (both active on testDate1)
		if metric.Value != expectedValue {
			t.Errorf("Expected open call count %f for %s, got %f", expectedValue, testDate1Key, metric.Value)
		}
	} else {
		t.Errorf("Missing open call count metric for date %s", testDate1Key)
	}

	if metric, exists := openCallCountByDate[testDate2Key]; exists {
		expectedValue := 2.0 // AAPL call + TSLA call
		if metric.Value != expectedValue {
			t.Errorf("Expected open call count %f for %s, got %f", expectedValue, testDate2Key, metric.Value)
		}
	} else {
		t.Errorf("Missing open call count metric for date %s", testDate2Key)
	}

	if metric, exists := openCallCountByDate[testDate3Key]; exists {
		expectedValue := 2.0 // AAPL call + TSLA call
		if metric.Value != expectedValue {
			t.Errorf("Expected open call count %f for %s, got %f", expectedValue, testDate3Key, metric.Value)
		}
	} else {
		t.Errorf("Missing open call count metric for date %s", testDate3Key)
	}
}