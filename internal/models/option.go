package models

import (
	"database/sql"
	"fmt"
	"math"
	"time"
)

// Commission constants
const OptionCommissionPerContract = 0.65

type OptionService struct {
	db *sql.DB
}

func NewOptionService(db *sql.DB) *OptionService {
	return &OptionService{db: db}
}

func (s *OptionService) Create(symbol, optionType string, opened time.Time, strike float64, expiration time.Time, premium float64, contracts int) (*Option, error) {
	// Automatically calculate opening commission: $0.65 per contract
	openingCommission := OptionCommissionPerContract * float64(contracts)
	return s.CreateWithCommission(symbol, optionType, opened, strike, expiration, premium, contracts, openingCommission)
}

func (s *OptionService) CreateWithCommission(symbol, optionType string, opened time.Time, strike float64, expiration time.Time, premium float64, contracts int, commission float64) (*Option, error) {
	if optionType != "Put" && optionType != "Call" {
		return nil, fmt.Errorf("option type must be 'Put' or 'Call'")
	}

	query := `INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?) 
			  RETURNING id, symbol, type, opened, closed, strike, expiration, premium, contracts, exit_price, commission, current_price, created_at, updated_at`

	var option Option
	err := s.db.QueryRow(query, symbol, optionType, opened, strike, expiration, premium, contracts, commission).Scan(
		&option.ID, &option.Symbol, &option.Type, &option.Opened, &option.Closed, &option.Strike,
		&option.Expiration, &option.Premium, &option.Contracts, &option.ExitPrice, &option.Commission,
		&option.CurrentPrice, &option.CreatedAt, &option.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create option: %w", err)
	}

	return &option, nil
}

func (s *OptionService) GetBySymbol(symbol string) ([]*Option, error) {
	query := `SELECT o.id, o.symbol, o.type, o.opened, o.closed, o.strike, o.expiration, o.premium,
			  o.contracts, o.exit_price, o.commission, o.current_price,
			  COALESCE(s.currency, 'USD') as currency,
			  o.created_at, o.updated_at
			  FROM options o
			  LEFT JOIN symbols s ON o.symbol = s.symbol
			  WHERE o.symbol = ? ORDER BY o.expiration DESC, o.opened DESC`

	rows, err := s.db.Query(query, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}
	defer rows.Close()

	var options []*Option
	for rows.Next() {
		var option Option
		if err := rows.Scan(&option.ID, &option.Symbol, &option.Type, &option.Opened, &option.Closed,
			&option.Strike, &option.Expiration, &option.Premium, &option.Contracts,
			&option.ExitPrice, &option.Commission, &option.CurrentPrice, &option.Currency,
			&option.CreatedAt, &option.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan option: %w", err)
		}
		options = append(options, &option)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating options: %w", err)
	}

	return options, nil
}

func (s *OptionService) GetAll() ([]*Option, error) {
	query := `SELECT o.id, o.symbol, o.type, o.opened, o.closed, o.strike, o.expiration, o.premium,
			  o.contracts, o.exit_price, o.commission, o.current_price,
			  COALESCE(s.currency, 'USD') as currency,
			  o.created_at, o.updated_at
			  FROM options o
			  LEFT JOIN symbols s ON o.symbol = s.symbol
			  ORDER BY o.expiration DESC, o.opened DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}
	defer rows.Close()

	var options []*Option
	for rows.Next() {
		var option Option
		if err := rows.Scan(&option.ID, &option.Symbol, &option.Type, &option.Opened, &option.Closed,
			&option.Strike, &option.Expiration, &option.Premium, &option.Contracts,
			&option.ExitPrice, &option.Commission, &option.CurrentPrice, &option.Currency,
			&option.CreatedAt, &option.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan option: %w", err)
		}
		options = append(options, &option)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating options: %w", err)
	}

	return options, nil
}

func (s *OptionService) GetOpen() ([]*Option, error) {
	query := `SELECT id, symbol, type, opened, closed, strike, expiration, premium, contracts, exit_price, commission, current_price, created_at, updated_at 
			  FROM options WHERE closed IS NULL ORDER BY expiration ASC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get open options: %w", err)
	}
	defer rows.Close()

	var options []*Option
	for rows.Next() {
		var option Option
		if err := rows.Scan(&option.ID, &option.Symbol, &option.Type, &option.Opened, &option.Closed,
			&option.Strike, &option.Expiration, &option.Premium, &option.Contracts,
			&option.ExitPrice, &option.Commission, &option.CurrentPrice, &option.CreatedAt, &option.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan option: %w", err)
		}
		options = append(options, &option)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating options: %w", err)
	}

	return options, nil
}

func (s *OptionService) Close(symbol, optionType string, opened time.Time, strike float64, expiration time.Time, premium float64, contracts int, closed time.Time, exitPrice float64) error {
	// Calculate closing commission: $0.65 per contract
	closingCommission := OptionCommissionPerContract * float64(contracts)

	query := `UPDATE options 
			  SET closed = ?, exit_price = ?, commission = commission + ?, updated_at = CURRENT_TIMESTAMP 
			  WHERE symbol = ? AND type = ? AND opened = ? AND strike = ? AND expiration = ? AND premium = ? AND contracts = ?`

	result, err := s.db.Exec(query, closed, exitPrice, closingCommission, symbol, optionType, opened, strike, expiration, premium, contracts)
	if err != nil {
		return fmt.Errorf("failed to close option: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("option not found")
	}

	return nil
}

func (s *OptionService) Delete(symbol, optionType string, opened time.Time, strike float64, expiration time.Time, premium float64, contracts int) error {
	query := `DELETE FROM options WHERE symbol = ? AND type = ? AND opened = ? AND strike = ? AND expiration = ? AND premium = ? AND contracts = ?`
	result, err := s.db.Exec(query, symbol, optionType, opened, strike, expiration, premium, contracts)
	if err != nil {
		return fmt.Errorf("failed to delete option: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("option not found")
	}

	return nil
}

// GetByID retrieves an option by its ID
func (s *OptionService) GetByID(id int) (*Option, error) {
	query := `SELECT id, symbol, type, opened, closed, strike, expiration, premium, contracts, exit_price, commission, current_price, created_at, updated_at 
			  FROM options WHERE id = ?`

	var option Option
	err := s.db.QueryRow(query, id).Scan(
		&option.ID, &option.Symbol, &option.Type, &option.Opened, &option.Closed,
		&option.Strike, &option.Expiration, &option.Premium, &option.Contracts,
		&option.ExitPrice, &option.Commission, &option.CurrentPrice, &option.CreatedAt, &option.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("option not found")
		}
		return nil, fmt.Errorf("failed to get option: %w", err)
	}

	return &option, nil
}

// UpdateByID updates an option by its ID
func (s *OptionService) UpdateByID(id int, symbol, optionType string, opened time.Time, strike float64, expiration time.Time, premium float64, contracts int, commission float64, closed *time.Time, exitPrice *float64) (*Option, error) {
	if optionType != "Put" && optionType != "Call" {
		return nil, fmt.Errorf("option type must be 'Put' or 'Call'")
	}

	query := `UPDATE options 
			  SET symbol = ?, type = ?, opened = ?, strike = ?, expiration = ?, premium = ?, contracts = ?, commission = ?, closed = ?, exit_price = ?, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = ? 
			  RETURNING id, symbol, type, opened, closed, strike, expiration, premium, contracts, exit_price, commission, current_price, created_at, updated_at`

	var option Option
	err := s.db.QueryRow(query, symbol, optionType, opened, strike, expiration, premium, contracts, commission, closed, exitPrice, id).Scan(
		&option.ID, &option.Symbol, &option.Type, &option.Opened, &option.Closed,
		&option.Strike, &option.Expiration, &option.Premium, &option.Contracts,
		&option.ExitPrice, &option.Commission, &option.CurrentPrice, &option.CreatedAt, &option.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("option not found")
		}
		return nil, fmt.Errorf("failed to update option: %w", err)
	}

	return &option, nil
}

// DeleteByID deletes an option by its ID
func (s *OptionService) DeleteByID(id int) error {
	query := `DELETE FROM options WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete option: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("option not found")
	}

	return nil
}

// CloseByID closes an option by its ID
func (s *OptionService) CloseByID(id int, closed time.Time, exitPrice float64) error {
	// First get the option to find out the number of contracts for commission calculation
	option, err := s.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get option for commission calculation: %w", err)
	}

	// Calculate closing commission: $0.65 per contract
	closingCommission := OptionCommissionPerContract * float64(option.Contracts)

	query := `UPDATE options 
			  SET closed = ?, exit_price = ?, commission = commission + ?, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = ?`

	result, err := s.db.Exec(query, closed, exitPrice, closingCommission, id)
	if err != nil {
		return fmt.Errorf("failed to close option: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("option not found")
	}

	return nil
}

func (s *OptionService) DeleteBySymbol(symbol string) error {
	query := `DELETE FROM options WHERE symbol = ?`
	result, err := s.db.Exec(query, symbol)
	if err != nil {
		return fmt.Errorf("failed to delete options for symbol %s: %w", symbol, err)
	}

	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	return nil
}

// OptionSummary represents options summary data by symbol
type OptionSummary struct {
	Symbol         string  `json:"symbol"`
	TotalPositions int     `json:"total_positions"`
	PutPositions   int     `json:"put_positions"`
	CallPositions  int     `json:"call_positions"`
	TotalPremium   float64 `json:"total_premium"`
	PutPremium     float64 `json:"put_premium"`
	CallPremium    float64 `json:"call_premium"`
	NetPremium     float64 `json:"net_premium"`
}

// OpenPositionData represents an open option position with additional calculated fields
type OpenPositionData struct {
	*Option
	DaysToExpiration int       `json:"days_to_expiration"`
	Status           string    `json:"status"`
	EntryDate        time.Time `json:"entry_date"`
	Currency         string    `json:"currency"`
}

// GetOptionsSummaryBySymbol returns options summary data grouped by symbol
func (s *OptionService) GetOptionsSummaryBySymbol() ([]*OptionSummary, error) {
	query := `
		SELECT 
			symbol,
			COUNT(*) as total_positions,
			SUM(CASE WHEN type = 'Put' THEN 1 ELSE 0 END) as put_positions,
			SUM(CASE WHEN type = 'Call' THEN 1 ELSE 0 END) as call_positions,
			SUM(premium) as total_premium,
			SUM(CASE WHEN type = 'Put' THEN premium ELSE 0 END) as put_premium,
			SUM(CASE WHEN type = 'Call' THEN premium ELSE 0 END) as call_premium,
			SUM(premium) as net_premium
		FROM options 
		WHERE closed IS NULL
		GROUP BY symbol 
		ORDER BY symbol`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get options summary: %w", err)
	}
	defer rows.Close()

	var summaries []*OptionSummary
	for rows.Next() {
		var summary OptionSummary
		if err := rows.Scan(
			&summary.Symbol, &summary.TotalPositions, &summary.PutPositions, &summary.CallPositions,
			&summary.TotalPremium, &summary.PutPremium, &summary.CallPremium, &summary.NetPremium,
		); err != nil {
			return nil, fmt.Errorf("failed to scan options summary: %w", err)
		}
		summaries = append(summaries, &summary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating options summaries: %w", err)
	}

	return summaries, nil
}

// GetOpenPositionsWithDetails returns open positions with calculated fields
func (s *OptionService) GetOpenPositionsWithDetails() ([]*OpenPositionData, error) {
	query := `SELECT o.id, o.symbol, o.type, o.opened, o.closed, o.strike, o.expiration, o.premium,
			  o.contracts, o.exit_price, o.commission, o.current_price, o.created_at, o.updated_at,
			  COALESCE(s.currency, 'USD') as currency
			  FROM options o
			  LEFT JOIN symbols s ON o.symbol = s.symbol
			  WHERE o.closed IS NULL
			  ORDER BY o.expiration ASC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get open options with details: %w", err)
	}
	defer rows.Close()

	var openPositions []*OpenPositionData
	now := time.Now()

	for rows.Next() {
		var option Option
		var currency string
		if err := rows.Scan(&option.ID, &option.Symbol, &option.Type, &option.Opened, &option.Closed,
			&option.Strike, &option.Expiration, &option.Premium, &option.Contracts,
			&option.ExitPrice, &option.Commission, &option.CurrentPrice, &option.CreatedAt, &option.UpdatedAt,
			&currency); err != nil {
			return nil, fmt.Errorf("failed to scan option with details: %w", err)
		}

		// Calculate days to expiration - use ceiling to avoid off-by-one error
		// An option expiring tomorrow should show 1 day, not 0
		daysToExp := int(math.Ceil(option.Expiration.Sub(now).Hours() / 24))

		// Determine status based on days to expiration
		status := "Active"
		if daysToExp < 0 {
			status = "Expired"
		} else if daysToExp <= 7 {
			status = "Critical"
		} else if daysToExp <= 30 {
			status = "Warning"
		}

		openPosition := &OpenPositionData{
			Option:           &option,
			DaysToExpiration: daysToExp,
			Status:           status,
			EntryDate:        option.Opened,
			Currency:         currency,
		}
		openPositions = append(openPositions, openPosition)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating open positions: %w", err)
	}

	return openPositions, nil
}

// GetOptionsSummaryTotals returns aggregate totals for all options
func (s *OptionService) GetOptionsSummaryTotals() (*OptionSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_positions,
			SUM(CASE WHEN type = 'Put' THEN 1 ELSE 0 END) as put_positions,
			SUM(CASE WHEN type = 'Call' THEN 1 ELSE 0 END) as call_positions,
			SUM(premium) as total_premium,
			SUM(CASE WHEN type = 'Put' THEN premium ELSE 0 END) as put_premium,
			SUM(CASE WHEN type = 'Call' THEN premium ELSE 0 END) as call_premium,
			SUM(premium) as net_premium
		FROM options 
		WHERE closed IS NULL`

	var totals OptionSummary
	totals.Symbol = "Total"

	err := s.db.QueryRow(query).Scan(
		&totals.TotalPositions, &totals.PutPositions, &totals.CallPositions,
		&totals.TotalPremium, &totals.PutPremium, &totals.CallPremium, &totals.NetPremium,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get options summary totals: %w", err)
	}

	return &totals, nil
}

// Index creates a nested index structure for all options
func (s *OptionService) Index() (map[string]interface{}, error) {
	options, err := s.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all options: %w", err)
	}

	return Index(options)
}
