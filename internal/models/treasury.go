package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Treasury struct {
	CUSPID       string     `json:"cuspid"`
	Purchased    time.Time  `json:"purchased"`
	Maturity     time.Time  `json:"maturity"`
	Amount       float64    `json:"amount"`
	Yield        float64    `json:"yield"`
	BuyPrice     float64    `json:"buy_price"`
	CurrentValue *float64   `json:"current_value"`
	ExitPrice    *float64   `json:"exit_price"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (t *Treasury) CalculateProfitLoss() float64 {
	// Use exit price if bond was sold
	if t.ExitPrice != nil {
		return *t.ExitPrice - t.BuyPrice
	}
	// Use current market value if available
	if t.CurrentValue != nil {
		return *t.CurrentValue - t.BuyPrice
	}
	// Fall back to face value vs purchase price
	return t.Amount - t.BuyPrice
}

func (t *Treasury) CalculateROI() float64 {
	if t.BuyPrice == 0 {
		return 0
	}
	profitLoss := t.CalculateProfitLoss()
	return (profitLoss / t.BuyPrice) * 100
}

func (t *Treasury) CalculateTotalInvested() float64 {
	return t.Amount
}

func (t *Treasury) CalculateInterest() float64 {
	// Interest earned is the difference between Exit Price and Buy Price
	// Only calculate if bond has been sold (Exit Price is set)
	if t.ExitPrice == nil {
		return 0.0 // No interest earned until bond is sold
	}
	return *t.ExitPrice - t.BuyPrice
}

// GetCurrentValue returns the current value as a float64, or 0.0 if nil
func (t *Treasury) GetCurrentValue() float64 {
	if t.CurrentValue == nil {
		return 0.0
	}
	return *t.CurrentValue
}

// GetExitPrice returns the exit price as a float64, or 0.0 if nil
func (t *Treasury) GetExitPrice() float64 {
	if t.ExitPrice == nil {
		return 0.0
	}
	return *t.ExitPrice
}

// HasCurrentValue returns true if current value is set
func (t *Treasury) HasCurrentValue() bool {
	return t.CurrentValue != nil
}

// HasExitPrice returns true if exit price is set
func (t *Treasury) HasExitPrice() bool {
	return t.ExitPrice != nil
}

// CalculateDaysRemaining calculates days remaining until maturity
func (t *Treasury) CalculateDaysRemaining() int {
	if t.ExitPrice != nil {
		return 0
	}
	now := time.Now()
	duration := t.Maturity.Sub(now)
	return int(duration.Hours() / 24)
}

type TreasuryService struct {
	db *sql.DB
}

func NewTreasuryService(db *sql.DB) *TreasuryService {
	return &TreasuryService{db: db}
}

func (s *TreasuryService) Create(cuspid string, purchased, maturity time.Time, amount, yield, buyPrice float64) (*Treasury, error) {
	log.Printf("[TREASURY SERVICE] Create: Starting creation for CUSPID=%s", cuspid)
	log.Printf("[TREASURY SERVICE] Create: Parameters - Purchased=%v, Maturity=%v, Amount=%.2f, Yield=%.3f, BuyPrice=%.2f", 
		purchased, maturity, amount, yield, buyPrice)
	
	if cuspid == "" {
		log.Printf("[TREASURY SERVICE] Create: ERROR - CUSPID cannot be empty")
		return nil, fmt.Errorf("CUSPID cannot be empty")
	}

	query := `INSERT INTO treasuries (cuspid, purchased, maturity, amount, yield, buy_price) 
			  VALUES (?, ?, ?, ?, ?, ?) 
			  RETURNING cuspid, purchased, maturity, amount, yield, buy_price, current_value, exit_price, created_at, updated_at`
	
	log.Printf("[TREASURY SERVICE] Create: Executing SQL query for CUSPID=%s", cuspid)
	log.Printf("[TREASURY SERVICE] Create: SQL = %s", query)
	
	var treasury Treasury
	err := s.db.QueryRow(query, cuspid, purchased, maturity, amount, yield, buyPrice).Scan(
		&treasury.CUSPID, &treasury.Purchased, &treasury.Maturity, &treasury.Amount,
		&treasury.Yield, &treasury.BuyPrice, &treasury.CurrentValue, &treasury.ExitPrice,
		&treasury.CreatedAt, &treasury.UpdatedAt,
	)
	if err != nil {
		log.Printf("[TREASURY SERVICE] Create: ERROR - SQL execution failed for CUSPID=%s: %v", cuspid, err)
		log.Printf("[TREASURY SERVICE] Create: Query parameters were: [%s, %v, %v, %.2f, %.3f, %.2f]", 
			cuspid, purchased, maturity, amount, yield, buyPrice)
		return nil, fmt.Errorf("failed to create treasury: %w", err)
	}

	log.Printf("[TREASURY SERVICE] Create: Successfully created treasury for CUSPID=%s", cuspid)
	log.Printf("[TREASURY SERVICE] Create: Created treasury data - Amount=%.2f, Yield=%.3f, BuyPrice=%.2f, CreatedAt=%v", 
		treasury.Amount, treasury.Yield, treasury.BuyPrice, treasury.CreatedAt)

	return &treasury, nil
}

// CreateFull creates a new treasury with all fields including optional current value and exit price
func (s *TreasuryService) CreateFull(cuspid string, purchased, maturity time.Time, amount, yield, buyPrice float64, currentValue, exitPrice *float64) (*Treasury, error) {
	log.Printf("[TREASURY SERVICE] CreateFull: Starting creation for CUSPID=%s", cuspid)
	log.Printf("[TREASURY SERVICE] CreateFull: Parameters - Purchased=%v, Maturity=%v, Amount=%.2f, Yield=%.3f, BuyPrice=%.2f", 
		purchased, maturity, amount, yield, buyPrice)
	if currentValue != nil {
		log.Printf("[TREASURY SERVICE] CreateFull: CurrentValue=%.2f", *currentValue)
	} else {
		log.Printf("[TREASURY SERVICE] CreateFull: CurrentValue=nil")
	}
	if exitPrice != nil {
		log.Printf("[TREASURY SERVICE] CreateFull: ExitPrice=%.2f", *exitPrice)
	} else {
		log.Printf("[TREASURY SERVICE] CreateFull: ExitPrice=nil")
	}
	
	if cuspid == "" {
		log.Printf("[TREASURY SERVICE] CreateFull: ERROR - CUSPID cannot be empty")
		return nil, fmt.Errorf("CUSPID cannot be empty")
	}

	query := `INSERT INTO treasuries (cuspid, purchased, maturity, amount, yield, buy_price, current_value, exit_price) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?) 
			  RETURNING cuspid, purchased, maturity, amount, yield, buy_price, current_value, exit_price, created_at, updated_at`
	
	log.Printf("[TREASURY SERVICE] CreateFull: Executing SQL query for CUSPID=%s", cuspid)
	log.Printf("[TREASURY SERVICE] CreateFull: SQL = %s", query)
	
	var treasury Treasury
	err := s.db.QueryRow(query, cuspid, purchased, maturity, amount, yield, buyPrice, currentValue, exitPrice).Scan(
		&treasury.CUSPID, &treasury.Purchased, &treasury.Maturity, &treasury.Amount,
		&treasury.Yield, &treasury.BuyPrice, &treasury.CurrentValue, &treasury.ExitPrice,
		&treasury.CreatedAt, &treasury.UpdatedAt,
	)
	if err != nil {
		log.Printf("[TREASURY SERVICE] CreateFull: ERROR - SQL execution failed for CUSPID=%s: %v", cuspid, err)
		log.Printf("[TREASURY SERVICE] CreateFull: Query parameters were: [%s, %v, %v, %.2f, %.3f, %.2f, %v, %v]", 
			cuspid, purchased, maturity, amount, yield, buyPrice, currentValue, exitPrice)
		return nil, fmt.Errorf("failed to create treasury: %w", err)
	}

	log.Printf("[TREASURY SERVICE] CreateFull: Successfully created treasury for CUSPID=%s", cuspid)
	log.Printf("[TREASURY SERVICE] CreateFull: Created treasury data - Amount=%.2f, Yield=%.3f, BuyPrice=%.2f, CreatedAt=%v", 
		treasury.Amount, treasury.Yield, treasury.BuyPrice, treasury.CreatedAt)

	return &treasury, nil
}

func (s *TreasuryService) GetAll() ([]*Treasury, error) {
	log.Printf("[TREASURY SERVICE] GetAll: Starting to retrieve all treasuries")
	
	query := `SELECT cuspid, purchased, maturity, amount, yield, buy_price, current_value, exit_price, created_at, updated_at 
			  FROM treasuries ORDER BY maturity DESC, purchased DESC`
	
	log.Printf("[TREASURY SERVICE] GetAll: Executing SQL query")
	log.Printf("[TREASURY SERVICE] GetAll: SQL = %s", query)
	
	rows, err := s.db.Query(query)
	if err != nil {
		log.Printf("[TREASURY SERVICE] GetAll: ERROR - SQL query failed: %v", err)
		return nil, fmt.Errorf("failed to get treasuries: %w", err)
	}
	defer rows.Close()

	var treasuries []*Treasury
	rowCount := 0
	for rows.Next() {
		var treasury Treasury
		if err := rows.Scan(&treasury.CUSPID, &treasury.Purchased, &treasury.Maturity, &treasury.Amount,
			&treasury.Yield, &treasury.BuyPrice, &treasury.CurrentValue, &treasury.ExitPrice,
			&treasury.CreatedAt, &treasury.UpdatedAt); err != nil {
			log.Printf("[TREASURY SERVICE] GetAll: ERROR - Failed to scan row %d: %v", rowCount, err)
			return nil, fmt.Errorf("failed to scan treasury: %w", err)
		}
		treasuries = append(treasuries, &treasury)
		rowCount++
		log.Printf("[TREASURY SERVICE] GetAll: Scanned treasury %d - CUSPID=%s, Amount=%.2f", 
			rowCount, treasury.CUSPID, treasury.Amount)
	}

	if err := rows.Err(); err != nil {
		log.Printf("[TREASURY SERVICE] GetAll: ERROR - Row iteration error: %v", err)
		return nil, fmt.Errorf("error iterating treasuries: %w", err)
	}

	log.Printf("[TREASURY SERVICE] GetAll: Successfully retrieved %d treasuries", len(treasuries))
	return treasuries, nil
}

func (s *TreasuryService) GetTotalOpenValue() (float64, error) {
	log.Printf("[TREASURY SERVICE] GetTotalOpenValue: Starting to calculate total open treasury value")
	
	query := `SELECT COALESCE(SUM(amount), 0) FROM treasuries WHERE exit_price IS NULL`
	
	log.Printf("[TREASURY SERVICE] GetTotalOpenValue: Executing SQL query")
	log.Printf("[TREASURY SERVICE] GetTotalOpenValue: SQL = %s", query)
	
	var total float64
	err := s.db.QueryRow(query).Scan(&total)
	if err != nil {
		log.Printf("[TREASURY SERVICE] GetTotalOpenValue: ERROR - SQL query failed: %v", err)
		return 0, fmt.Errorf("failed to get total open treasury value: %w", err)
	}
	
	log.Printf("[TREASURY SERVICE] GetTotalOpenValue: Successfully calculated total = $%.2f", total)
	return total, nil
}

func (s *TreasuryService) GetByCUSPID(cuspid string) (*Treasury, error) {
	log.Printf("[TREASURY SERVICE] GetByCUSPID: Starting to retrieve treasury for CUSPID=%s", cuspid)
	
	query := `SELECT cuspid, purchased, maturity, amount, yield, buy_price, current_value, exit_price, created_at, updated_at 
			  FROM treasuries WHERE cuspid = ?`
	
	log.Printf("[TREASURY SERVICE] GetByCUSPID: Executing SQL query for CUSPID=%s", cuspid)
	log.Printf("[TREASURY SERVICE] GetByCUSPID: SQL = %s", query)
	
	var treasury Treasury
	err := s.db.QueryRow(query, cuspid).Scan(&treasury.CUSPID, &treasury.Purchased, &treasury.Maturity,
		&treasury.Amount, &treasury.Yield, &treasury.BuyPrice, &treasury.CurrentValue, &treasury.ExitPrice,
		&treasury.CreatedAt, &treasury.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[TREASURY SERVICE] GetByCUSPID: ERROR - Treasury not found for CUSPID=%s", cuspid)
			return nil, fmt.Errorf("treasury not found")
		}
		log.Printf("[TREASURY SERVICE] GetByCUSPID: ERROR - SQL query failed for CUSPID=%s: %v", cuspid, err)
		return nil, fmt.Errorf("failed to get treasury: %w", err)
	}

	log.Printf("[TREASURY SERVICE] GetByCUSPID: Successfully retrieved treasury for CUSPID=%s", cuspid)
	log.Printf("[TREASURY SERVICE] GetByCUSPID: Treasury data - Amount=%.2f, Yield=%.3f, BuyPrice=%.2f", 
		treasury.Amount, treasury.Yield, treasury.BuyPrice)

	return &treasury, nil
}

func (s *TreasuryService) Update(cuspid string, currentValue, exitPrice *float64) (*Treasury, error) {
	query := `UPDATE treasuries SET current_value = ?, exit_price = ?, updated_at = CURRENT_TIMESTAMP 
			  WHERE cuspid = ? 
			  RETURNING cuspid, purchased, maturity, amount, yield, buy_price, current_value, exit_price, created_at, updated_at`
	
	var treasury Treasury
	err := s.db.QueryRow(query, currentValue, exitPrice, cuspid).Scan(&treasury.CUSPID, &treasury.Purchased,
		&treasury.Maturity, &treasury.Amount, &treasury.Yield, &treasury.BuyPrice, &treasury.CurrentValue,
		&treasury.ExitPrice, &treasury.CreatedAt, &treasury.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("treasury not found")
		}
		return nil, fmt.Errorf("failed to update treasury: %w", err)
	}

	return &treasury, nil
}

// UpdateFull updates all editable fields of a treasury
func (s *TreasuryService) UpdateFull(cuspid string, purchased, maturity time.Time, amount, yield, buyPrice float64, currentValue, exitPrice *float64) (*Treasury, error) {
	log.Printf("[TREASURY SERVICE] UpdateFull: Starting full update for CUSPID=%s", cuspid)
	log.Printf("[TREASURY SERVICE] UpdateFull: Parameters - Purchased=%v, Maturity=%v, Amount=%.2f, Yield=%.3f, BuyPrice=%.2f", 
		purchased, maturity, amount, yield, buyPrice)
	if currentValue != nil {
		log.Printf("[TREASURY SERVICE] UpdateFull: CurrentValue=%.2f", *currentValue)
	} else {
		log.Printf("[TREASURY SERVICE] UpdateFull: CurrentValue=nil")
	}
	if exitPrice != nil {
		log.Printf("[TREASURY SERVICE] UpdateFull: ExitPrice=%.2f", *exitPrice)
	} else {
		log.Printf("[TREASURY SERVICE] UpdateFull: ExitPrice=nil")
	}
	
	query := `UPDATE treasuries SET purchased = ?, maturity = ?, amount = ?, yield = ?, buy_price = ?, current_value = ?, exit_price = ?, updated_at = CURRENT_TIMESTAMP 
			  WHERE cuspid = ? 
			  RETURNING cuspid, purchased, maturity, amount, yield, buy_price, current_value, exit_price, created_at, updated_at`
	
	log.Printf("[TREASURY SERVICE] UpdateFull: Executing SQL query for CUSPID=%s", cuspid)
	log.Printf("[TREASURY SERVICE] UpdateFull: SQL = %s", query)
	
	var treasury Treasury
	err := s.db.QueryRow(query, purchased, maturity, amount, yield, buyPrice, currentValue, exitPrice, cuspid).Scan(
		&treasury.CUSPID, &treasury.Purchased, &treasury.Maturity, &treasury.Amount,
		&treasury.Yield, &treasury.BuyPrice, &treasury.CurrentValue, &treasury.ExitPrice,
		&treasury.CreatedAt, &treasury.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[TREASURY SERVICE] UpdateFull: ERROR - Treasury not found for CUSPID=%s", cuspid)
			return nil, fmt.Errorf("treasury not found")
		}
		log.Printf("[TREASURY SERVICE] UpdateFull: ERROR - SQL execution failed for CUSPID=%s: %v", cuspid, err)
		log.Printf("[TREASURY SERVICE] UpdateFull: Query parameters were: [%v, %v, %.2f, %.3f, %.2f, %v, %v, %s]", 
			purchased, maturity, amount, yield, buyPrice, currentValue, exitPrice, cuspid)
		return nil, fmt.Errorf("failed to update treasury: %w", err)
	}

	log.Printf("[TREASURY SERVICE] UpdateFull: Successfully updated treasury for CUSPID=%s", cuspid)
	log.Printf("[TREASURY SERVICE] UpdateFull: Updated treasury data - Amount=%.2f, Yield=%.3f, BuyPrice=%.2f, UpdatedAt=%v", 
		treasury.Amount, treasury.Yield, treasury.BuyPrice, treasury.UpdatedAt)

	return &treasury, nil
}

func (s *TreasuryService) Delete(cuspid string) error {
	log.Printf("[TREASURY SERVICE] Delete: Starting deletion for CUSPID=%s", cuspid)
	
	query := `DELETE FROM treasuries WHERE cuspid = ?`
	
	log.Printf("[TREASURY SERVICE] Delete: Executing SQL query for CUSPID=%s", cuspid)
	log.Printf("[TREASURY SERVICE] Delete: SQL = %s", query)
	
	result, err := s.db.Exec(query, cuspid)
	if err != nil {
		log.Printf("[TREASURY SERVICE] Delete: ERROR - SQL execution failed for CUSPID=%s: %v", cuspid, err)
		return fmt.Errorf("failed to delete treasury: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[TREASURY SERVICE] Delete: ERROR - Failed to get rows affected for CUSPID=%s: %v", cuspid, err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	log.Printf("[TREASURY SERVICE] Delete: SQL execution completed for CUSPID=%s, rows affected: %d", cuspid, rowsAffected)

	if rowsAffected == 0 {
		log.Printf("[TREASURY SERVICE] Delete: ERROR - Treasury not found for CUSPID=%s (no rows affected)", cuspid)
		return fmt.Errorf("treasury not found")
	}

	log.Printf("[TREASURY SERVICE] Delete: Successfully deleted treasury for CUSPID=%s", cuspid)
	return nil
}

// CalculateSummary calculates summary statistics for a collection of treasuries
func (s *TreasuryService) CalculateSummary(treasuries []*Treasury) *TreasurySummary {
	summary := &TreasurySummary{}
	
	if len(treasuries) == 0 {
		return summary
	}

	var totalYield float64
	activeCount := 0

	for _, t := range treasuries {
		summary.TotalAmount += t.Amount
		summary.TotalBuyPrice += t.BuyPrice
		summary.TotalProfitLoss += t.CalculateProfitLoss()
		summary.TotalInterest += t.CalculateInterest()

		if t.ExitPrice == nil {
			activeCount++
			totalYield += t.Yield
		}
	}

	summary.ActivePositions = activeCount
	// Average Return = (Total Interest Earned / Total Buy Prices) * 100
	if summary.TotalBuyPrice > 0 {
		summary.AverageReturn = (summary.TotalInterest / summary.TotalBuyPrice) * 100
	}

	return summary
}

type TreasurySummary struct {
	TotalAmount     float64 `json:"total_amount"`
	TotalBuyPrice   float64 `json:"total_buy_price"`
	TotalProfitLoss float64 `json:"total_profit_loss"`
	TotalInterest   float64 `json:"total_interest"`
	AverageReturn   float64 `json:"average_return"`
	ActivePositions int     `json:"active_positions"`
}