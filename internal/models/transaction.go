package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// Transaction represents a financial transaction using the Universal Transaction CSV format
type Transaction struct {
	ID              int        `json:"id"`
	TransactionType string     `json:"transaction_type"` // STOCK, OPTION, DIVIDEND
	Symbol          string     `json:"symbol"`
	Date            time.Time  `json:"date"`
	Action          string     `json:"action"` // BUY, SELL, SELL_TO_OPEN, BUY_TO_CLOSE, ASSIGNED, EXPIRED, RECEIVE
	Quantity        *int       `json:"quantity,omitempty"`        // Number of shares/contracts (null for dividends)
	Price           *float64   `json:"price,omitempty"`           // Price per share/option premium (null for dividends)
	Strike          *float64   `json:"strike,omitempty"`          // Strike price (options only)
	Expiration      *time.Time `json:"expiration,omitempty"`      // Expiration date (options only)
	OptionType      *string    `json:"option_type,omitempty"`     // Put/Call (options only)
	Amount          *float64   `json:"amount,omitempty"`          // Direct monetary amount (dividends)
	Commission      float64    `json:"commission"`                // Transaction fees
	Notes           *string    `json:"notes,omitempty"`           // Free-form description
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// TransactionService handles CRUD operations for transactions
type TransactionService struct {
	db *sql.DB
}

// NewTransactionService creates a new transaction service
func NewTransactionService(db *sql.DB) *TransactionService {
	return &TransactionService{db: db}
}

// Create creates a new transaction
func (s *TransactionService) Create(transactionType, symbol string, date time.Time, action string, 
	quantity *int, price *float64, strike *float64, expiration *time.Time, optionType *string, 
	amount *float64, commission float64, notes *string) (*Transaction, error) {
	
	log.Printf("[TRANSACTION SERVICE] Create: Creating transaction %s %s %s %s", 
		transactionType, symbol, date.Format("2006-01-02"), action)

	// Validate required fields based on transaction type
	if err := s.validateTransaction(transactionType, action, quantity, price, strike, expiration, optionType, amount); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	now := time.Now()
	query := `
		INSERT INTO transactions (
			transaction_type, symbol, date, action, quantity, price, strike, expiration, 
			option_type, amount, commission, notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := s.db.Exec(query, transactionType, symbol, date, action, quantity, price, strike, 
		expiration, optionType, amount, commission, notes, now, now)
	if err != nil {
		log.Printf("[TRANSACTION SERVICE] Create: Failed to execute query: %v", err)
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("[TRANSACTION SERVICE] Create: Failed to get last insert ID: %v", err)
		return nil, fmt.Errorf("failed to get transaction ID: %w", err)
	}

	transaction := &Transaction{
		ID:              int(id),
		TransactionType: transactionType,
		Symbol:          symbol,
		Date:            date,
		Action:          action,
		Quantity:        quantity,
		Price:           price,
		Strike:          strike,
		Expiration:      expiration,
		OptionType:      optionType,
		Amount:          amount,
		Commission:      commission,
		Notes:           notes,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	log.Printf("[TRANSACTION SERVICE] Create: Successfully created transaction with ID %d", id)
	return transaction, nil
}

// GetByID retrieves a transaction by ID
func (s *TransactionService) GetByID(id int) (*Transaction, error) {
	log.Printf("[TRANSACTION SERVICE] GetByID: Retrieving transaction ID %d", id)

	query := `
		SELECT id, transaction_type, symbol, date, action, quantity, price, strike, expiration,
		       option_type, amount, commission, notes, created_at, updated_at
		FROM transactions WHERE id = ?`

	row := s.db.QueryRow(query, id)
	transaction, err := s.scanTransaction(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction with ID %d not found", id)
		}
		log.Printf("[TRANSACTION SERVICE] GetByID: Failed to scan transaction: %v", err)
		return nil, fmt.Errorf("failed to retrieve transaction: %w", err)
	}

	log.Printf("[TRANSACTION SERVICE] GetByID: Successfully retrieved transaction %d", id)
	return transaction, nil
}

// GetBySymbol retrieves all transactions for a specific symbol
func (s *TransactionService) GetBySymbol(symbol string) ([]*Transaction, error) {
	log.Printf("[TRANSACTION SERVICE] GetBySymbol: Retrieving transactions for symbol %s", symbol)

	query := `
		SELECT id, transaction_type, symbol, date, action, quantity, price, strike, expiration,
		       option_type, amount, commission, notes, created_at, updated_at
		FROM transactions WHERE symbol = ? ORDER BY date DESC, id DESC`

	rows, err := s.db.Query(query, symbol)
	if err != nil {
		log.Printf("[TRANSACTION SERVICE] GetBySymbol: Failed to execute query: %v", err)
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		transaction, err := s.scanTransaction(rows)
		if err != nil {
			log.Printf("[TRANSACTION SERVICE] GetBySymbol: Failed to scan transaction: %v", err)
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[TRANSACTION SERVICE] GetBySymbol: Row iteration error: %v", err)
		return nil, fmt.Errorf("failed to iterate transactions: %w", err)
	}

	log.Printf("[TRANSACTION SERVICE] GetBySymbol: Successfully retrieved %d transactions for %s", len(transactions), symbol)
	return transactions, nil
}

// GetAll retrieves all transactions
func (s *TransactionService) GetAll() ([]*Transaction, error) {
	log.Printf("[TRANSACTION SERVICE] GetAll: Retrieving all transactions")

	query := `
		SELECT id, transaction_type, symbol, date, action, quantity, price, strike, expiration,
		       option_type, amount, commission, notes, created_at, updated_at
		FROM transactions ORDER BY date DESC, id DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		log.Printf("[TRANSACTION SERVICE] GetAll: Failed to execute query: %v", err)
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		transaction, err := s.scanTransaction(rows)
		if err != nil {
			log.Printf("[TRANSACTION SERVICE] GetAll: Failed to scan transaction: %v", err)
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[TRANSACTION SERVICE] GetAll: Row iteration error: %v", err)
		return nil, fmt.Errorf("failed to iterate transactions: %w", err)
	}

	log.Printf("[TRANSACTION SERVICE] GetAll: Successfully retrieved %d transactions", len(transactions))
	return transactions, nil
}

// UpdateByID updates a transaction by ID
func (s *TransactionService) UpdateByID(id int, transactionType, symbol string, date time.Time, action string,
	quantity *int, price *float64, strike *float64, expiration *time.Time, optionType *string,
	amount *float64, commission float64, notes *string) (*Transaction, error) {

	log.Printf("[TRANSACTION SERVICE] UpdateByID: Updating transaction ID %d", id)

	// Validate the updated transaction
	if err := s.validateTransaction(transactionType, action, quantity, price, strike, expiration, optionType, amount); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	now := time.Now()
	query := `
		UPDATE transactions SET 
			transaction_type = ?, symbol = ?, date = ?, action = ?, quantity = ?, price = ?,
			strike = ?, expiration = ?, option_type = ?, amount = ?, commission = ?, notes = ?, updated_at = ?
		WHERE id = ?`

	result, err := s.db.Exec(query, transactionType, symbol, date, action, quantity, price, strike,
		expiration, optionType, amount, commission, notes, now, id)
	if err != nil {
		log.Printf("[TRANSACTION SERVICE] UpdateByID: Failed to execute query: %v", err)
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[TRANSACTION SERVICE] UpdateByID: Failed to get rows affected: %v", err)
		return nil, fmt.Errorf("failed to get update result: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("transaction with ID %d not found", id)
	}

	log.Printf("[TRANSACTION SERVICE] UpdateByID: Successfully updated transaction %d", id)
	return s.GetByID(id)
}

// DeleteByID deletes a transaction by ID
func (s *TransactionService) DeleteByID(id int) error {
	log.Printf("[TRANSACTION SERVICE] DeleteByID: Deleting transaction ID %d", id)

	query := `DELETE FROM transactions WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		log.Printf("[TRANSACTION SERVICE] DeleteByID: Failed to execute query: %v", err)
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[TRANSACTION SERVICE] DeleteByID: Failed to get rows affected: %v", err)
		return fmt.Errorf("failed to get delete result: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction with ID %d not found", id)
	}

	log.Printf("[TRANSACTION SERVICE] DeleteByID: Successfully deleted transaction %d", id)
	return nil
}

// GetByTypeAndSymbol retrieves transactions filtered by type and symbol
func (s *TransactionService) GetByTypeAndSymbol(transactionType, symbol string) ([]*Transaction, error) {
	log.Printf("[TRANSACTION SERVICE] GetByTypeAndSymbol: Retrieving %s transactions for %s", transactionType, symbol)

	query := `
		SELECT id, transaction_type, symbol, date, action, quantity, price, strike, expiration,
		       option_type, amount, commission, notes, created_at, updated_at
		FROM transactions WHERE transaction_type = ? AND symbol = ? ORDER BY date DESC, id DESC`

	rows, err := s.db.Query(query, transactionType, symbol)
	if err != nil {
		log.Printf("[TRANSACTION SERVICE] GetByTypeAndSymbol: Failed to execute query: %v", err)
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		transaction, err := s.scanTransaction(rows)
		if err != nil {
			log.Printf("[TRANSACTION SERVICE] GetByTypeAndSymbol: Failed to scan transaction: %v", err)
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[TRANSACTION SERVICE] GetByTypeAndSymbol: Row iteration error: %v", err)
		return nil, fmt.Errorf("failed to iterate transactions: %w", err)
	}

	log.Printf("[TRANSACTION SERVICE] GetByTypeAndSymbol: Successfully retrieved %d transactions", len(transactions))
	return transactions, nil
}

// scanTransaction is a helper function to scan database rows into Transaction struct
func (s *TransactionService) scanTransaction(scanner interface {
	Scan(dest ...interface{}) error
}) (*Transaction, error) {
	var transaction Transaction
	var dateStr, createdAtStr, updatedAtStr string
	var expirationStr sql.NullString

	err := scanner.Scan(
		&transaction.ID,
		&transaction.TransactionType,
		&transaction.Symbol,
		&dateStr,
		&transaction.Action,
		&transaction.Quantity,
		&transaction.Price,
		&transaction.Strike,
		&expirationStr,
		&transaction.OptionType,
		&transaction.Amount,
		&transaction.Commission,
		&transaction.Notes,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		return nil, err
	}

	// Parse dates - handle both DATE and DATETIME formats from SQLite
	transaction.Date, err = parseFlexibleDate(dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %w", err)
	}

	if expirationStr.Valid {
		expiration, err := parseFlexibleDate(expirationStr.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse expiration: %w", err)
		}
		transaction.Expiration = &expiration
	}

	transaction.CreatedAt, err = parseFlexibleTimestamp(createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	transaction.UpdatedAt, err = parseFlexibleTimestamp(updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &transaction, nil
}

// validateTransaction validates transaction data based on transaction type and action
func (s *TransactionService) validateTransaction(transactionType, action string, quantity *int, price *float64,
	strike *float64, expiration *time.Time, optionType *string, amount *float64) error {

	// Validate transaction type
	validTypes := map[string]bool{"STOCK": true, "OPTION": true, "DIVIDEND": true}
	if !validTypes[transactionType] {
		return fmt.Errorf("invalid transaction type: %s", transactionType)
	}

	// Validate action
	validActions := map[string]bool{
		"BUY": true, "SELL": true, "SELL_TO_OPEN": true, "BUY_TO_CLOSE": true,
		"ASSIGNED": true, "EXPIRED": true, "RECEIVE": true,
	}
	if !validActions[action] {
		return fmt.Errorf("invalid action: %s", action)
	}

	// Validate based on transaction type
	switch transactionType {
	case "STOCK":
		if action != "BUY" && action != "SELL" {
			return fmt.Errorf("invalid action %s for STOCK transaction", action)
		}
		if quantity == nil || *quantity <= 0 {
			return fmt.Errorf("quantity must be positive for STOCK transactions")
		}
		if price == nil || *price <= 0 {
			return fmt.Errorf("price must be positive for STOCK transactions")
		}

	case "OPTION":
		validOptionActions := map[string]bool{
			"SELL_TO_OPEN": true, "BUY_TO_CLOSE": true, "ASSIGNED": true, "EXPIRED": true,
		}
		if !validOptionActions[action] {
			return fmt.Errorf("invalid action %s for OPTION transaction", action)
		}
		if quantity == nil || *quantity <= 0 {
			return fmt.Errorf("quantity must be positive for OPTION transactions")
		}
		if strike == nil || *strike <= 0 {
			return fmt.Errorf("strike must be positive for OPTION transactions")
		}
		if expiration == nil {
			return fmt.Errorf("expiration is required for OPTION transactions")
		}
		if optionType == nil || (*optionType != "Put" && *optionType != "Call") {
			return fmt.Errorf("option_type must be 'Put' or 'Call' for OPTION transactions")
		}
		// Price validation for options (not required for ASSIGNED/EXPIRED)
		if (action == "SELL_TO_OPEN" || action == "BUY_TO_CLOSE") && (price == nil || *price < 0) {
			return fmt.Errorf("price must be non-negative for %s actions", action)
		}

	case "DIVIDEND":
		if action != "RECEIVE" {
			return fmt.Errorf("invalid action %s for DIVIDEND transaction", action)
		}
		if amount == nil || *amount <= 0 {
			return fmt.Errorf("amount must be positive for DIVIDEND transactions")
		}
	}

	return nil
}

// parseFlexibleDate parses date strings in various formats that SQLite might return
func parseFlexibleDate(dateStr string) (time.Time, error) {
	// Try different date formats
	formats := []string{
		"2006-01-02",                      // Standard DATE format
		"2006-01-02T15:04:05Z",           // ISO 8601 with timezone
		"2006-01-02T15:04:05",            // ISO 8601 without timezone
		"2006-01-02 15:04:05",            // SQLite DATETIME format
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			// Return just the date part (strip time)
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// parseFlexibleTimestamp parses timestamp strings in various formats that SQLite might return
func parseFlexibleTimestamp(timestampStr string) (time.Time, error) {
	// Try different timestamp formats
	formats := []string{
		"2006-01-02 15:04:05",                    // SQLite DATETIME format
		"2006-01-02T15:04:05",                   // ISO 8601 without timezone
		"2006-01-02T15:04:05Z",                  // ISO 8601 with UTC
		"2006-01-02T15:04:05.999999999Z",        // ISO 8601 with nanoseconds and UTC
		"2006-01-02T15:04:05.999999999-07:00",   // ISO 8601 with nanoseconds and timezone
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, timestampStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", timestampStr)
}