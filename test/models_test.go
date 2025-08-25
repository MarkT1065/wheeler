package test

import (
	"os"
	"testing"
	"time"

	"stonks/internal/database"
	"stonks/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

func TestTransactionModel(t *testing.T) {
	// Setup test database
	testDB := setupTestDatabaseForModels(t)
	defer testDB.Close()

	// Create services
	symbolService := models.NewSymbolService(testDB.DB)
	transactionService := models.NewTransactionService(testDB.DB)

	// Create a test symbol first
	testSymbol := "AAPL"
	_, err := symbolService.Create(testSymbol)
	if err != nil {
		t.Fatalf("Failed to create test symbol: %v", err)
	}

	t.Run("CreateStockTransaction", func(t *testing.T) {
		testCreateStockTransaction(t, transactionService)
	})

	t.Run("CreateOptionTransaction", func(t *testing.T) {
		testCreateOptionTransaction(t, transactionService)
	})

	t.Run("CreateDividendTransaction", func(t *testing.T) {
		testCreateDividendTransaction(t, transactionService)
	})

	t.Run("ValidationTests", func(t *testing.T) {
		testValidationErrors(t, transactionService)
	})

	t.Run("CRUDOperations", func(t *testing.T) {
		testCRUDOperations(t, transactionService)
	})

	t.Run("QueryOperations", func(t *testing.T) {
		testQueryOperations(t, transactionService)
	})
}

func testCreateStockTransaction(t *testing.T, service *models.TransactionService) {
	// Test BUY transaction
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	quantity := 100
	price := 150.50
	commission := 2.50

	transaction, err := service.Create("STOCK", "AAPL", date, "BUY", &quantity, &price, nil, nil, nil, nil, commission, nil)
	if err != nil {
		t.Fatalf("Failed to create STOCK BUY transaction: %v", err)
	}

	// Verify transaction fields
	if transaction.ID == 0 {
		t.Error("Transaction ID should be set")
	}
	if transaction.TransactionType != "STOCK" {
		t.Errorf("Expected transaction type STOCK, got %s", transaction.TransactionType)
	}
	if transaction.Symbol != "AAPL" {
		t.Errorf("Expected symbol AAPL, got %s", transaction.Symbol)
	}
	if transaction.Action != "BUY" {
		t.Errorf("Expected action BUY, got %s", transaction.Action)
	}
	if transaction.Quantity == nil || *transaction.Quantity != quantity {
		t.Errorf("Expected quantity %d, got %v", quantity, transaction.Quantity)
	}
	if transaction.Price == nil || *transaction.Price != price {
		t.Errorf("Expected price %f, got %v", price, transaction.Price)
	}
	if transaction.Commission != commission {
		t.Errorf("Expected commission %f, got %f", commission, transaction.Commission)
	}
	if transaction.Strike != nil {
		t.Error("Strike should be nil for STOCK transactions")
	}
	if transaction.Expiration != nil {
		t.Error("Expiration should be nil for STOCK transactions")
	}
	if transaction.OptionType != nil {
		t.Error("OptionType should be nil for STOCK transactions")
	}
	if transaction.Amount != nil {
		t.Error("Amount should be nil for STOCK transactions")
	}

	t.Logf("✅ Successfully created STOCK BUY transaction with ID %d", transaction.ID)

	// Test SELL transaction
	sellQuantity := 50
	sellPrice := 155.75
	_, err = service.Create("STOCK", "AAPL", date.AddDate(0, 0, 30), "SELL", &sellQuantity, &sellPrice, nil, nil, nil, nil, 2.50, nil)
	if err != nil {
		t.Fatalf("Failed to create STOCK SELL transaction: %v", err)
	}

	t.Logf("✅ Successfully created STOCK SELL transaction")
}

func testCreateOptionTransaction(t *testing.T, service *models.TransactionService) {
	// Test SELL_TO_OPEN (cash-secured put)
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	expiration := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	quantity := 2
	premium := 5.50
	strike := 150.00
	optionType := "Put"
	commission := 2.10

	transaction, err := service.Create("OPTION", "AAPL", date, "SELL_TO_OPEN", &quantity, &premium, &strike, &expiration, &optionType, nil, commission, nil)
	if err != nil {
		t.Fatalf("Failed to create OPTION SELL_TO_OPEN transaction: %v", err)
	}

	// Verify option-specific fields
	if transaction.TransactionType != "OPTION" {
		t.Errorf("Expected transaction type OPTION, got %s", transaction.TransactionType)
	}
	if transaction.Action != "SELL_TO_OPEN" {
		t.Errorf("Expected action SELL_TO_OPEN, got %s", transaction.Action)
	}
	if transaction.Strike == nil || *transaction.Strike != strike {
		t.Errorf("Expected strike %f, got %v", strike, transaction.Strike)
	}
	if transaction.Expiration == nil || !transaction.Expiration.Equal(expiration) {
		t.Errorf("Expected expiration %v, got %v", expiration, transaction.Expiration)
	}
	if transaction.OptionType == nil || *transaction.OptionType != optionType {
		t.Errorf("Expected option type %s, got %v", optionType, transaction.OptionType)
	}
	if transaction.Amount != nil {
		t.Error("Amount should be nil for OPTION transactions")
	}

	t.Logf("✅ Successfully created OPTION SELL_TO_OPEN transaction with ID %d", transaction.ID)

	// Test BUY_TO_CLOSE
	closePrice := 3.25
	_, err = service.Create("OPTION", "AAPL", date.AddDate(0, 0, 20), "BUY_TO_CLOSE", &quantity, &closePrice, &strike, &expiration, &optionType, nil, commission, nil)
	if err != nil {
		t.Fatalf("Failed to create OPTION BUY_TO_CLOSE transaction: %v", err)
	}

	t.Logf("✅ Successfully created OPTION BUY_TO_CLOSE transaction")

	// Test ASSIGNED transaction (no price required)
	_, err = service.Create("OPTION", "AAPL", expiration, "ASSIGNED", &quantity, nil, &strike, &expiration, &optionType, nil, 0, nil)
	if err != nil {
		t.Fatalf("Failed to create OPTION ASSIGNED transaction: %v", err)
	}

	t.Logf("✅ Successfully created OPTION ASSIGNED transaction")

	// Test EXPIRED transaction (no price required)
	callType := "Call"
	_, err = service.Create("OPTION", "AAPL", expiration, "EXPIRED", &quantity, nil, &strike, &expiration, &callType, nil, 0, nil)
	if err != nil {
		t.Fatalf("Failed to create OPTION EXPIRED transaction: %v", err)
	}

	t.Logf("✅ Successfully created OPTION EXPIRED transaction")
}

func testCreateDividendTransaction(t *testing.T, service *models.TransactionService) {
	// Test RECEIVE dividend
	date := time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC)
	amount := 25.50
	notes := "Quarterly dividend payment"

	transaction, err := service.Create("DIVIDEND", "AAPL", date, "RECEIVE", nil, nil, nil, nil, nil, &amount, 0, &notes)
	if err != nil {
		t.Fatalf("Failed to create DIVIDEND RECEIVE transaction: %v", err)
	}

	// Verify dividend-specific fields
	if transaction.TransactionType != "DIVIDEND" {
		t.Errorf("Expected transaction type DIVIDEND, got %s", transaction.TransactionType)
	}
	if transaction.Action != "RECEIVE" {
		t.Errorf("Expected action RECEIVE, got %s", transaction.Action)
	}
	if transaction.Amount == nil || *transaction.Amount != amount {
		t.Errorf("Expected amount %f, got %v", amount, transaction.Amount)
	}
	if transaction.Notes == nil || *transaction.Notes != notes {
		t.Errorf("Expected notes '%s', got %v", notes, transaction.Notes)
	}
	if transaction.Quantity != nil {
		t.Error("Quantity should be nil for DIVIDEND transactions")
	}
	if transaction.Price != nil {
		t.Error("Price should be nil for DIVIDEND transactions")
	}
	if transaction.Strike != nil {
		t.Error("Strike should be nil for DIVIDEND transactions")
	}
	if transaction.Expiration != nil {
		t.Error("Expiration should be nil for DIVIDEND transactions")
	}
	if transaction.OptionType != nil {
		t.Error("OptionType should be nil for DIVIDEND transactions")
	}

	t.Logf("✅ Successfully created DIVIDEND RECEIVE transaction with ID %d", transaction.ID)
}

func testValidationErrors(t *testing.T, service *models.TransactionService) {
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	quantity := 100
	price := 150.50

	// Test invalid transaction type
	_, err := service.Create("INVALID", "AAPL", date, "BUY", &quantity, &price, nil, nil, nil, nil, 0, nil)
	if err == nil {
		t.Error("Expected error for invalid transaction type")
	}

	// Test invalid action
	_, err = service.Create("STOCK", "AAPL", date, "INVALID", &quantity, &price, nil, nil, nil, nil, 0, nil)
	if err == nil {
		t.Error("Expected error for invalid action")
	}

	// Test STOCK transaction without quantity
	_, err = service.Create("STOCK", "AAPL", date, "BUY", nil, &price, nil, nil, nil, nil, 0, nil)
	if err == nil {
		t.Error("Expected error for STOCK transaction without quantity")
	}

	// Test STOCK transaction without price
	_, err = service.Create("STOCK", "AAPL", date, "BUY", &quantity, nil, nil, nil, nil, nil, 0, nil)
	if err == nil {
		t.Error("Expected error for STOCK transaction without price")
	}

	// Test OPTION transaction without required fields
	_, err = service.Create("OPTION", "AAPL", date, "SELL_TO_OPEN", &quantity, &price, nil, nil, nil, nil, 0, nil)
	if err == nil {
		t.Error("Expected error for OPTION transaction without strike")
	}

	// Test DIVIDEND transaction without amount
	_, err = service.Create("DIVIDEND", "AAPL", date, "RECEIVE", nil, nil, nil, nil, nil, nil, 0, nil)
	if err == nil {
		t.Error("Expected error for DIVIDEND transaction without amount")
	}

	t.Logf("✅ All validation tests passed")
}

func testCRUDOperations(t *testing.T, service *models.TransactionService) {
	// Create a transaction
	date := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	quantity := 150
	price := 145.00
	notes := "Test transaction"

	transaction, err := service.Create("STOCK", "AAPL", date, "BUY", &quantity, &price, nil, nil, nil, nil, 1.50, &notes)
	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// Test GetByID
	retrieved, err := service.GetByID(transaction.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve transaction: %v", err)
	}
	if retrieved.ID != transaction.ID {
		t.Errorf("Expected ID %d, got %d", transaction.ID, retrieved.ID)
	}
	if retrieved.Symbol != "AAPL" {
		t.Errorf("Expected symbol AAPL, got %s", retrieved.Symbol)
	}

	// Test UpdateByID
	newPrice := 147.50
	newNotes := "Updated test transaction"
	updated, err := service.UpdateByID(transaction.ID, "STOCK", "AAPL", date, "BUY", &quantity, &newPrice, nil, nil, nil, nil, 1.50, &newNotes)
	if err != nil {
		t.Fatalf("Failed to update transaction: %v", err)
	}
	if updated.Price == nil || *updated.Price != newPrice {
		t.Errorf("Expected updated price %f, got %v", newPrice, updated.Price)
	}
	if updated.Notes == nil || *updated.Notes != newNotes {
		t.Errorf("Expected updated notes '%s', got %v", newNotes, updated.Notes)
	}

	// Test DeleteByID
	err = service.DeleteByID(transaction.ID)
	if err != nil {
		t.Fatalf("Failed to delete transaction: %v", err)
	}

	// Verify deletion
	_, err = service.GetByID(transaction.ID)
	if err == nil {
		t.Error("Expected error when retrieving deleted transaction")
	}

	t.Logf("✅ All CRUD operations passed")
}

func testQueryOperations(t *testing.T, service *models.TransactionService) {
	// Create several test transactions
	date := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	
	// Stock transactions
	quantity := 100
	price := 160.00
	_, err := service.Create("STOCK", "AAPL", date, "BUY", &quantity, &price, nil, nil, nil, nil, 2.00, nil)
	if err != nil {
		t.Fatalf("Failed to create test stock transaction: %v", err)
	}

	// Option transaction
	expiration := time.Date(2024, 4, 15, 0, 0, 0, 0, time.UTC)
	optionQuantity := 1
	premium := 7.50
	strike := 165.00
	optionType := "Call"
	_, err = service.Create("OPTION", "AAPL", date, "SELL_TO_OPEN", &optionQuantity, &premium, &strike, &expiration, &optionType, nil, 1.05, nil)
	if err != nil {
		t.Fatalf("Failed to create test option transaction: %v", err)
	}

	// Dividend transaction
	amount := 30.00
	_, err = service.Create("DIVIDEND", "AAPL", date.AddDate(0, 0, 15), "RECEIVE", nil, nil, nil, nil, nil, &amount, 0, nil)
	if err != nil {
		t.Fatalf("Failed to create test dividend transaction: %v", err)
	}

	// Test GetBySymbol
	transactions, err := service.GetBySymbol("AAPL")
	if err != nil {
		t.Fatalf("Failed to get transactions by symbol: %v", err)
	}
	if len(transactions) < 3 {
		t.Errorf("Expected at least 3 transactions, got %d", len(transactions))
	}

	// Test GetByTypeAndSymbol
	stockTransactions, err := service.GetByTypeAndSymbol("STOCK", "AAPL")
	if err != nil {
		t.Fatalf("Failed to get stock transactions: %v", err)
	}
	stockCount := 0
	for _, tx := range stockTransactions {
		if tx.TransactionType == "STOCK" {
			stockCount++
		}
	}
	if stockCount == 0 {
		t.Error("Expected at least one STOCK transaction")
	}

	optionTransactions, err := service.GetByTypeAndSymbol("OPTION", "AAPL")
	if err != nil {
		t.Fatalf("Failed to get option transactions: %v", err)
	}
	optionCount := 0
	for _, tx := range optionTransactions {
		if tx.TransactionType == "OPTION" {
			optionCount++
		}
	}
	if optionCount == 0 {
		t.Error("Expected at least one OPTION transaction")
	}

	// Test GetAll
	allTransactions, err := service.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all transactions: %v", err)
	}
	if len(allTransactions) == 0 {
		t.Error("Expected at least some transactions")
	}

	t.Logf("✅ All query operations passed - found %d total transactions", len(allTransactions))
}

func setupTestDatabaseForModels(t *testing.T) *database.DB {
	// Create a unique test database file
	testDBPath := "test_transactions.db"

	// Remove existing test database
	os.Remove(testDBPath)

	// Create new database
	db, err := database.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Clean up function
	t.Cleanup(func() {
		db.Close()
		os.Remove(testDBPath)
	})

	return db
}