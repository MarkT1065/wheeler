package test

import (
	"stonks/internal/database"
	"stonks/internal/models"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TestWheelerMarketStory tests a complete wheel strategy story with buy/sell transactions
// Each step has measurable expectations that tell the story of a Wheeler trader in the market
func TestWheelerMarketStory(t *testing.T) {
	// Setup test database
	testDB := setupTestDatabaseForStory(t)
	defer testDB.Close()

	// Create services
	symbolService := models.NewSymbolService(testDB.DB)
	transactionService := models.NewTransactionService(testDB.DB)

	// Create test symbols
	symbols := []string{"AAPL", "MSFT", "NVDA"}
	for _, symbol := range symbols {
		_, err := symbolService.Create(symbol)
		if err != nil {
			t.Fatalf("Failed to create symbol %s: %v", symbol, err)
		}
	}

	// Wheeler's Market Story - Chapter 1: Building the Foundation
	t.Run("Chapter1_BuildingTheFoundation", func(t *testing.T) {
		testChapter1BuildingFoundation(t, transactionService)
	})

	// Wheeler's Market Story - Chapter 2: The First Assignment
	t.Run("Chapter2_TheFirstAssignment", func(t *testing.T) {
		testChapter2FirstAssignment(t, transactionService)
	})

	// Wheeler's Market Story - Chapter 3: Scaling Success
	t.Run("Chapter3_ScalingSuccess", func(t *testing.T) {
		testChapter3ScalingSuccess(t, transactionService)
	})

	// Wheeler's Market Story - Chapter 4: Dividend Harvest
	t.Run("Chapter4_DividendHarvest", func(t *testing.T) {
		testChapter4DividendHarvest(t, transactionService)
	})

	// Wheeler's Market Story - Chapter 5: Market Rally
	t.Run("Chapter5_MarketRally", func(t *testing.T) {
		testChapter5MarketRally(t, transactionService)
	})

	// Wheeler's Market Story - Chapter 6: Final Tally
	t.Run("Chapter6_FinalTally", func(t *testing.T) {
		testChapter6FinalTally(t, transactionService)
	})
}

// Chapter 1: Wheeler starts with cash-secured puts, collecting premium while waiting for assignment
func testChapter1BuildingFoundation(t *testing.T, service *models.TransactionService) {
	t.Log("📖 Chapter 1: Building the Foundation")
	t.Log("Wheeler enters the market with $50,000 cash, selling cash-secured puts on quality stocks")

	// Wheeler's first move: Sell cash-secured puts on AAPL, MSFT, NVDA
	// Market is at reasonable levels, Wheeler wants to own these at lower strikes

	// January 15, 2024: Sell AAPL $180 Put for March expiration
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	expiration := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	quantity := 1
	premium := 8.50 // Collect $850 premium
	strike := 180.00
	optionType := "Put"

	_, err := service.Create("OPTION", "AAPL", date, "SELL_TO_OPEN", &quantity, &premium, &strike, &expiration, &optionType, nil, 1.05, nil)
	if err != nil {
		t.Fatalf("Failed to sell AAPL put: %v", err)
	}

	// January 15, 2024: Sell MSFT $350 Put for March expiration
	premium = 12.75 // Collect $1,275 premium
	strike = 350.00
	_, err = service.Create("OPTION", "MSFT", date, "SELL_TO_OPEN", &quantity, &premium, &strike, &expiration, &optionType, nil, 1.05, nil)
	if err != nil {
		t.Fatalf("Failed to sell MSFT put: %v", err)
	}

	// January 15, 2024: Sell NVDA $450 Put for March expiration
	premium = 22.50 // Collect $2,250 premium
	strike = 450.00
	_, err = service.Create("OPTION", "NVDA", date, "SELL_TO_OPEN", &quantity, &premium, &strike, &expiration, &optionType, nil, 1.05, nil)
	if err != nil {
		t.Fatalf("Failed to sell NVDA put: %v", err)
	}

	// Verify Chapter 1 expectations
	allTransactions, err := service.GetAll()
	if err != nil {
		t.Fatalf("Failed to get transactions: %v", err)
	}

	// Assert: Wheeler has sold 3 puts and collected premium
	optionCount := 0
	totalPremiumCollected := 0.0
	totalCashRequired := 0.0 // Cash collateral required

	for _, tx := range allTransactions {
		if tx.TransactionType == "OPTION" && tx.Action == "SELL_TO_OPEN" {
			optionCount++
			if tx.Price != nil {
				totalPremiumCollected += (*tx.Price * float64(*tx.Quantity) * 100) // Premium * contracts * 100
			}
			if tx.Strike != nil {
				totalCashRequired += (*tx.Strike * float64(*tx.Quantity) * 100) // Strike * contracts * 100
			}
		}
	}

	// Story expectations: Wheeler sold 3 puts, collected $4,375 premium, secured with $98,000 cash
	expectedPremium := (8.50 + 12.75 + 22.50) * 100 // $4,375
	expectedCashRequired := (180.00 + 350.00 + 450.00) * 100 // $98,000

	if optionCount != 3 {
		t.Errorf("Expected Wheeler to sell 3 puts, got %d", optionCount)
	}
	if totalPremiumCollected != expectedPremium {
		t.Errorf("Expected premium collected %.2f, got %.2f", expectedPremium, totalPremiumCollected)
	}
	if totalCashRequired != expectedCashRequired {
		t.Errorf("Expected cash required %.2f, got %.2f", expectedCashRequired, totalCashRequired)
	}

	t.Logf("✅ Wheeler successfully sold 3 cash-secured puts")
	t.Logf("   Premium collected: $%.2f", totalPremiumCollected)
	t.Logf("   Cash collateral required: $%.2f", totalCashRequired)
	t.Logf("   Wheeler's yield on collateral: %.2f%%", (totalPremiumCollected/totalCashRequired)*100)
}

// Chapter 2: NVDA drops below strike, Wheeler gets assigned and owns the stock
func testChapter2FirstAssignment(t *testing.T, service *models.TransactionService) {
	t.Log("📖 Chapter 2: The First Assignment")
	t.Log("NVDA drops to $440, Wheeler's put gets assigned - now owns 100 shares at $450")

	// March 15, 2024: NVDA put assignment - Wheeler buys stock at strike
	assignmentDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	quantity := 1
	strike := 450.00
	optionType := "Put"

	// Assignment transaction: Put expires ITM, Wheeler must buy stock
	_, err := service.Create("OPTION", "NVDA", assignmentDate, "ASSIGNED", &quantity, nil, &strike, &assignmentDate, &optionType, nil, 0, nil)
	if err != nil {
		t.Fatalf("Failed to record NVDA put assignment: %v", err)
	}

	// Stock acquisition from assignment
	shares := 100
	assignmentPrice := 450.00
	_, err = service.Create("STOCK", "NVDA", assignmentDate, "BUY", &shares, &assignmentPrice, nil, nil, nil, nil, 0, nil)
	if err != nil {
		t.Fatalf("Failed to record NVDA stock acquisition: %v", err)
	}

	// AAPL and MSFT puts expire worthless (stocks stayed above strike)
	// AAPL $180 put expires worthless
	quantity = 1
	strike = 180.00
	_, err = service.Create("OPTION", "AAPL", assignmentDate, "EXPIRED", &quantity, nil, &strike, &assignmentDate, &optionType, nil, 0, nil)
	if err != nil {
		t.Fatalf("Failed to record AAPL put expiration: %v", err)
	}

	// MSFT $350 put expires worthless
	strike = 350.00
	_, err = service.Create("OPTION", "MSFT", assignmentDate, "EXPIRED", &quantity, nil, &strike, &assignmentDate, &optionType, nil, 0, nil)
	if err != nil {
		t.Fatalf("Failed to record MSFT put expiration: %v", err)
	}

	// Verify Chapter 2 expectations
	allTransactions, err := service.GetAll()
	if err != nil {
		t.Fatalf("Failed to get transactions: %v", err)
	}

	// Count assignments vs expirations
	assignments := 0
	expirations := 0
	stockPurchases := 0
	totalStockCost := 0.0

	for _, tx := range allTransactions {
		if tx.TransactionType == "OPTION" {
			if tx.Action == "ASSIGNED" {
				assignments++
			} else if tx.Action == "EXPIRED" {
				expirations++
			}
		} else if tx.TransactionType == "STOCK" && tx.Action == "BUY" {
			stockPurchases++
			if tx.Price != nil && tx.Quantity != nil {
				totalStockCost += (*tx.Price * float64(*tx.Quantity))
			}
		}
	}

	// Story expectations: 1 assignment (NVDA), 2 expirations (AAPL, MSFT), 1 stock purchase ($45,000)
	if assignments != 1 {
		t.Errorf("Expected 1 assignment, got %d", assignments)
	}
	if expirations != 2 {
		t.Errorf("Expected 2 expirations, got %d", expirations)
	}
	if stockPurchases != 1 {
		t.Errorf("Expected 1 stock purchase, got %d", stockPurchases)
	}

	expectedStockCost := 450.00 * 100 // $45,000
	if totalStockCost != expectedStockCost {
		t.Errorf("Expected stock cost %.2f, got %.2f", expectedStockCost, totalStockCost)
	}

	t.Logf("✅ Wheeler's first assignment completed")
	t.Logf("   NVDA: Assigned at $450/share (100 shares = $45,000)")
	t.Logf("   AAPL & MSFT puts: Expired worthless (kept premium)")
	t.Logf("   Wheeler now owns: 100 NVDA shares")
}

// Chapter 3: Wheeler scales up, selling covered calls on NVDA and more cash-secured puts
func testChapter3ScalingSuccess(t *testing.T, service *models.TransactionService) {
	t.Log("📖 Chapter 3: Scaling Success")
	t.Log("Wheeler sells covered call on NVDA position and scales up with more puts")

	// March 20, 2024: Sell covered call on NVDA position
	date := time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC)
	expiration := time.Date(2024, 4, 19, 0, 0, 0, 0, time.UTC)
	quantity := 1
	premium := 18.75 // Collect $1,875 for covered call
	strike := 480.00 // Above current price, hoping for assignment
	optionType := "Call"

	_, err := service.Create("OPTION", "NVDA", date, "SELL_TO_OPEN", &quantity, &premium, &strike, &expiration, &optionType, nil, 1.05, nil)
	if err != nil {
		t.Fatalf("Failed to sell NVDA covered call: %v", err)
	}

	// March 20, 2024: Sell more cash-secured puts (scaling strategy)
	optionType = "Put"

	// Sell AAPL $175 Put (lower strike after learning market behavior)
	strike = 175.00
	premium = 6.25
	_, err = service.Create("OPTION", "AAPL", date, "SELL_TO_OPEN", &quantity, &premium, &strike, &expiration, &optionType, nil, 1.05, nil)
	if err != nil {
		t.Fatalf("Failed to sell AAPL put: %v", err)
	}

	// Sell MSFT $340 Put
	strike = 340.00
	premium = 9.50
	_, err = service.Create("OPTION", "MSFT", date, "SELL_TO_OPEN", &quantity, &premium, &strike, &expiration, &optionType, nil, 1.05, nil)
	if err != nil {
		t.Fatalf("Failed to sell MSFT put: %v", err)
	}

	// Verify Chapter 3 expectations
	// Count new options sold in this chapter
	marchTransactions, err := service.GetBySymbol("NVDA")
	if err != nil {
		t.Fatalf("Failed to get NVDA transactions: %v", err)
	}

	// Find the covered call
	coveredCallFound := false
	for _, tx := range marchTransactions {
		if tx.TransactionType == "OPTION" && tx.Action == "SELL_TO_OPEN" && 
		   tx.OptionType != nil && *tx.OptionType == "Call" && 
		   tx.Date.Month() == time.March {
			coveredCallFound = true
			expectedCallPremium := 18.75
			if tx.Price == nil || *tx.Price != expectedCallPremium {
				t.Errorf("Expected NVDA call premium %.2f, got %.2f", expectedCallPremium, *tx.Price)
			}
		}
	}

	if !coveredCallFound {
		t.Error("Expected to find NVDA covered call sold in March")
	}

	// Count total option positions Wheeler now has
	allTransactions, err := service.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all transactions: %v", err)
	}

	activeOptions := 0
	totalNewPremium := 0.0

	for _, tx := range allTransactions {
		if tx.TransactionType == "OPTION" && tx.Action == "SELL_TO_OPEN" && 
		   tx.Date.Month() == time.March {
			activeOptions++
			if tx.Price != nil {
				totalNewPremium += (*tx.Price * 100) // Premium per contract
			}
		}
	}

	// Story expectations: 3 new options sold in March, $3,450 premium collected
	expectedNewOptions := 3
	expectedNewPremium := (18.75 + 6.25 + 9.50) * 100 // $3,450

	if activeOptions != expectedNewOptions {
		t.Errorf("Expected %d new options in March, got %d", expectedNewOptions, activeOptions)
	}
	if totalNewPremium != expectedNewPremium {
		t.Errorf("Expected new premium %.2f, got %.2f", expectedNewPremium, totalNewPremium)
	}

	t.Logf("✅ Wheeler successfully scaled the strategy")
	t.Logf("   NVDA covered call: $%.2f premium collected", 18.75*100)
	t.Logf("   New cash-secured puts: $%.2f premium collected", (6.25+9.50)*100)
	t.Logf("   Total March premium: $%.2f", totalNewPremium)
}

// Chapter 4: Wheeler collects dividends while holding NVDA
func testChapter4DividendHarvest(t *testing.T, service *models.TransactionService) {
	t.Log("📖 Chapter 4: Dividend Harvest")
	t.Log("Wheeler's NVDA position pays dividends while waiting for options to expire")

	// March 25, 2024: NVDA pays quarterly dividend
	dividendDate := time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC)
	dividendAmount := 16.00 // NVDA quarterly dividend on 100 shares

	_, err := service.Create("DIVIDEND", "NVDA", dividendDate, "RECEIVE", nil, nil, nil, nil, nil, &dividendAmount, 0, nil)
	if err != nil {
		t.Fatalf("Failed to record NVDA dividend: %v", err)
	}

	// Verify Chapter 4 expectations
	dividendTransactions, err := service.GetByTypeAndSymbol("DIVIDEND", "NVDA")
	if err != nil {
		t.Fatalf("Failed to get NVDA dividends: %v", err)
	}

	// Count dividend payments
	totalDividends := 0.0
	dividendCount := 0

	for _, tx := range dividendTransactions {
		if tx.Action == "RECEIVE" {
			dividendCount++
			if tx.Amount != nil {
				totalDividends += *tx.Amount
			}
		}
	}

	// Story expectations: 1 dividend payment of $16.00
	if dividendCount != 1 {
		t.Errorf("Expected 1 dividend payment, got %d", dividendCount)
	}
	if totalDividends != 16.00 {
		t.Errorf("Expected dividend amount %.2f, got %.2f", 16.00, totalDividends)
	}

	t.Logf("✅ Wheeler collected dividend income")
	t.Logf("   NVDA dividend: $%.2f", totalDividends)
	t.Logf("   Quarterly yield on NVDA position: %.3f%%", (totalDividends/(450.00*100))*100)
}

// Chapter 5: Market rallies, NVDA called away at profit
func testChapter5MarketRally(t *testing.T, service *models.TransactionService) {
	t.Log("📖 Chapter 5: Market Rally")
	t.Log("NVDA rallies to $485, Wheeler's covered call gets assigned - stock sold for profit")

	// April 19, 2024: NVDA covered call assigned, other puts expire worthless
	assignmentDate := time.Date(2024, 4, 19, 0, 0, 0, 0, time.UTC)
	quantity := 1
	strike := 480.00
	optionType := "Call"

	// NVDA call assignment
	_, err := service.Create("OPTION", "NVDA", assignmentDate, "ASSIGNED", &quantity, nil, &strike, &assignmentDate, &optionType, nil, 0, nil)
	if err != nil {
		t.Fatalf("Failed to record NVDA call assignment: %v", err)
	}

	// Stock sale from call assignment
	shares := 100
	salePrice := 480.00 // Assigned at strike price
	_, err = service.Create("STOCK", "NVDA", assignmentDate, "SELL", &shares, &salePrice, nil, nil, nil, nil, 0, nil)
	if err != nil {
		t.Fatalf("Failed to record NVDA stock sale: %v", err)
	}

	// April puts expire worthless (market stayed strong)
	optionType = "Put"

	// AAPL $175 put expires worthless
	strike = 175.00
	_, err = service.Create("OPTION", "AAPL", assignmentDate, "EXPIRED", &quantity, nil, &strike, &assignmentDate, &optionType, nil, 0, nil)
	if err != nil {
		t.Fatalf("Failed to record AAPL put expiration: %v", err)
	}

	// MSFT $340 put expires worthless
	strike = 340.00
	_, err = service.Create("OPTION", "MSFT", assignmentDate, "EXPIRED", &quantity, nil, &strike, &assignmentDate, &optionType, nil, 0, nil)
	if err != nil {
		t.Fatalf("Failed to record MSFT put expiration: %v", err)
	}

	// Verify Chapter 5 expectations - calculate NVDA total profit
	nvdaTransactions, err := service.GetBySymbol("NVDA")
	if err != nil {
		t.Fatalf("Failed to get NVDA transactions: %v", err)
	}

	// Calculate complete NVDA trade profit
	var stockBuyPrice, stockSellPrice float64
	var putPremium, callPremium float64
	var dividend float64

	for _, tx := range nvdaTransactions {
		if tx.TransactionType == "STOCK" {
			if tx.Action == "BUY" && tx.Price != nil {
				stockBuyPrice = *tx.Price
			} else if tx.Action == "SELL" && tx.Price != nil {
				stockSellPrice = *tx.Price
			}
		} else if tx.TransactionType == "OPTION" && tx.Action == "SELL_TO_OPEN" && tx.Price != nil {
			if tx.OptionType != nil && *tx.OptionType == "Put" {
				putPremium = *tx.Price
			} else if tx.OptionType != nil && *tx.OptionType == "Call" {
				callPremium = *tx.Price
			}
		} else if tx.TransactionType == "DIVIDEND" && tx.Amount != nil {
			dividend = *tx.Amount
		}
	}

	// NVDA complete trade P&L calculation
	stockProfit := (stockSellPrice - stockBuyPrice) * 100 // $30/share * 100 = $3,000
	totalPremium := (putPremium + callPremium) * 100      // ($22.50 + $18.75) * 100 = $4,125
	totalNVDAProfit := stockProfit + totalPremium + dividend // $3,000 + $4,125 + $16 = $7,141

	// Story expectations
	expectedStockProfit := (480.00 - 450.00) * 100 // $3,000
	expectedTotalPremium := (22.50 + 18.75) * 100  // $4,125
	expectedDividend := 16.00
	expectedTotalProfit := expectedStockProfit + expectedTotalPremium + expectedDividend // $7,141

	if stockProfit != expectedStockProfit {
		t.Errorf("Expected NVDA stock profit %.2f, got %.2f", expectedStockProfit, stockProfit)
	}
	if totalPremium != expectedTotalPremium {
		t.Errorf("Expected NVDA total premium %.2f, got %.2f", expectedTotalPremium, totalPremium)
	}
	if totalNVDAProfit != expectedTotalProfit {
		t.Errorf("Expected NVDA total profit %.2f, got %.2f", expectedTotalProfit, totalNVDAProfit)
	}

	t.Logf("✅ Wheeler's NVDA trade completed profitably")
	t.Logf("   Stock profit: $%.2f (bought $450, sold $480)", stockProfit)
	t.Logf("   Premium collected: $%.2f (put + call)", totalPremium)
	t.Logf("   Dividend income: $%.2f", dividend)
	t.Logf("   Total NVDA profit: $%.2f", totalNVDAProfit)
	t.Logf("   Return on investment: %.1f%%", (totalNVDAProfit/(450.00*100))*100)
}

// Chapter 6: Wheeler tallies up total performance across all trades
func testChapter6FinalTally(t *testing.T, service *models.TransactionService) {
	t.Log("📖 Chapter 6: Final Tally")
	t.Log("Wheeler reviews the complete trading story and calculates total returns")

	// Get all transactions for final analysis
	allTransactions, err := service.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all transactions: %v", err)
	}

	// Calculate comprehensive performance metrics
	var totalPremiumCollected float64
	var totalDividends float64
	var totalStockProfitLoss float64
	var totalCommissions float64
	var activePositions int
	
	// Track by symbol for detailed analysis
	symbolMetrics := make(map[string]struct {
		premiumCollected float64
		stockPL         float64
		dividends       float64
		isActive        bool
	})

	for _, symbol := range []string{"AAPL", "MSFT", "NVDA"} {
		symbolMetrics[symbol] = struct {
			premiumCollected float64
			stockPL         float64
			dividends       float64
			isActive        bool
		}{}
	}

	stockPositions := make(map[string]struct {
		buyPrice  float64
		sellPrice float64
		hasBuy    bool
		hasSell   bool
	})

	for _, tx := range allTransactions {
		switch tx.TransactionType {
		case "OPTION":
			if tx.Action == "SELL_TO_OPEN" && tx.Price != nil {
				premium := *tx.Price * float64(*tx.Quantity) * 100
				totalPremiumCollected += premium
				
				if metrics, exists := symbolMetrics[tx.Symbol]; exists {
					metrics.premiumCollected += premium
					symbolMetrics[tx.Symbol] = metrics
				}
			}

		case "STOCK":
			pos := stockPositions[tx.Symbol]
			if tx.Action == "BUY" && tx.Price != nil {
				pos.buyPrice = *tx.Price
				pos.hasBuy = true
			} else if tx.Action == "SELL" && tx.Price != nil {
				pos.sellPrice = *tx.Price
				pos.hasSell = true
			}
			stockPositions[tx.Symbol] = pos

		case "DIVIDEND":
			if tx.Amount != nil {
				totalDividends += *tx.Amount
				
				if metrics, exists := symbolMetrics[tx.Symbol]; exists {
					metrics.dividends += *tx.Amount
					symbolMetrics[tx.Symbol] = metrics
				}
			}
		}

		totalCommissions += tx.Commission
	}

	// Calculate stock P&L
	for symbol, pos := range stockPositions {
		if pos.hasBuy && pos.hasSell {
			stockPL := (pos.sellPrice - pos.buyPrice) * 100
			totalStockProfitLoss += stockPL
			
			if metrics, exists := symbolMetrics[symbol]; exists {
				metrics.stockPL = stockPL
				symbolMetrics[symbol] = metrics
			}
		} else if pos.hasBuy && !pos.hasSell {
			activePositions++
			if metrics, exists := symbolMetrics[symbol]; exists {
				metrics.isActive = true
				symbolMetrics[symbol] = metrics
			}
		}
	}

	// Wheeler's final performance
	totalIncome := totalPremiumCollected + totalDividends + totalStockProfitLoss
	netProfit := totalIncome - totalCommissions

	// Story expectations - verify Wheeler's successful wheel strategy
	// Based on our transaction story:
	// - Total premium: $4,375 (Jan) + $3,450 (Mar) = $7,825
	// - NVDA stock profit: $3,000  
	// - NVDA dividend: $16
	// - Total commissions: ~$12 (multiple trades)
	expectedTotalPremium := 7825.00
	expectedStockPL := 3000.00
	expectedDividends := 16.00

	// Verify final results with some tolerance for calculations
	tolerance := 0.01
	if abs(totalPremiumCollected-expectedTotalPremium) > tolerance {
		t.Errorf("Expected total premium %.2f, got %.2f", expectedTotalPremium, totalPremiumCollected)
	}
	if abs(totalStockProfitLoss-expectedStockPL) > tolerance {
		t.Errorf("Expected stock P&L %.2f, got %.2f", expectedStockPL, totalStockProfitLoss)
	}
	if abs(totalDividends-expectedDividends) > tolerance {
		t.Errorf("Expected dividends %.2f, got %.2f", expectedDividends, totalDividends)
	}

	t.Logf("🎯 Wheeler's Market Story - Final Results:")
	t.Logf("   📊 Total Premium Collected: $%.2f", totalPremiumCollected)
	t.Logf("   📈 Stock Trading Profit: $%.2f", totalStockProfitLoss)
	t.Logf("   💰 Dividend Income: $%.2f", totalDividends)
	t.Logf("   💸 Total Commissions: $%.2f", totalCommissions)
	t.Logf("   🏆 Net Profit: $%.2f", netProfit)
	t.Logf("")
	t.Logf("   📋 By Symbol Performance:")

	for symbol, metrics := range symbolMetrics {
		symbolTotal := metrics.premiumCollected + metrics.stockPL + metrics.dividends
		status := "Completed"
		if metrics.isActive {
			status = "Active Position"
		}
		t.Logf("     %s: $%.2f premium + $%.2f stock P&L + $%.2f dividends = $%.2f total (%s)",
			symbol, metrics.premiumCollected, metrics.stockPL, metrics.dividends, symbolTotal, status)
	}

	t.Logf("")
	t.Logf("✅ Wheeler's wheel strategy generated %.2f%% return on deployed capital", 
		(netProfit/50000.00)*100) // Assume $50k initial capital

	// Verify the story makes sense
	if activePositions > 0 {
		t.Logf("   📌 Wheeler still has %d active position(s) that could be monetized", activePositions)
	}

	if netProfit <= 0 {
		t.Error("Wheeler's strategy should be profitable!")
	}

	// Final story validation - ensure we have the expected number of each transaction type
	transactionCounts := make(map[string]int)
	for _, tx := range allTransactions {
		key := tx.TransactionType + "_" + tx.Action
		transactionCounts[key]++
	}

	t.Logf("   🔢 Transaction Summary:")
	t.Logf("     Option Sales: %d", transactionCounts["OPTION_SELL_TO_OPEN"])
	t.Logf("     Option Assignments: %d", transactionCounts["OPTION_ASSIGNED"])
	t.Logf("     Option Expirations: %d", transactionCounts["OPTION_EXPIRED"])
	t.Logf("     Stock Purchases: %d", transactionCounts["STOCK_BUY"])
	t.Logf("     Stock Sales: %d", transactionCounts["STOCK_SELL"])
	t.Logf("     Dividend Receipts: %d", transactionCounts["DIVIDEND_RECEIVE"])

	t.Logf("📚 Wheeler's market story demonstrates successful wheel strategy execution!")
}

func setupTestDatabaseForStory(t *testing.T) *database.DB {
	// Create a unique test database file
	testDBPath := "test_wheeler_story.db"

	// Remove existing test database
	// os.Remove(testDBPath) // Commented out to allow inspection

	// Create new database
	db, err := database.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Clean up function
	t.Cleanup(func() {
		db.Close()
		// os.Remove(testDBPath) // Commented out to allow inspection
	})

	return db
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}