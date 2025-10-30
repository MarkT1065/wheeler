package models

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"
)

type Symbol struct {
	Symbol         string     `json:"symbol"`
	Price          float64    `json:"price"`
	Dividend       float64    `json:"dividend"`
	ExDividendDate *time.Time `json:"ex_dividend_date"`
	PERatio        *float64   `json:"pe_ratio"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CalculateYield calculates the annualized dividend yield percentage
func (s *Symbol) CalculateYield() float64 {
	if s.Price == 0 {
		return 0
	}
	return (s.Dividend * 4) / s.Price * 100 // Convert to percentage
}

type LongPosition struct {
	ID        int        `json:"id"`
	Symbol    string     `json:"symbol"`
	Opened    time.Time  `json:"opened"`
	Closed    *time.Time `json:"closed"`
	Shares    int        `json:"shares"`
	BuyPrice  float64    `json:"buy_price"`
	ExitPrice *float64   `json:"exit_price"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (lp *LongPosition) CalculateYield(quarterlyDividend float64) float64 {
	if lp.BuyPrice == 0 {
		return 0
	}
	return (quarterlyDividend * 4) / lp.BuyPrice * 100
}

func (lp *LongPosition) CalculateProfitLoss(currentPrice float64) float64 {
	exitPrice := currentPrice
	if lp.ExitPrice != nil {
		exitPrice = *lp.ExitPrice
	}
	return (exitPrice - lp.BuyPrice) * float64(lp.Shares)
}

func (lp *LongPosition) CalculateROI(currentPrice float64) float64 {
	if lp.BuyPrice == 0 {
		return 0
	}
	exitPrice := currentPrice
	if lp.ExitPrice != nil {
		exitPrice = *lp.ExitPrice
	}
	return ((exitPrice - lp.BuyPrice) / lp.BuyPrice) * 100
}

func (lp *LongPosition) CalculateAmount() float64 {
	return lp.BuyPrice * float64(lp.Shares)
}

func (lp *LongPosition) CalculateTotalInvested() float64 {
	// For now, same as amount (but could include fees, etc. in future)
	return lp.BuyPrice * float64(lp.Shares)
}

func (lp *LongPosition) GetExitPriceValue() float64 {
	if lp.ExitPrice == nil {
		return 0.0
	}
	return *lp.ExitPrice
}

type Option struct {
	ID         int        `json:"id"`
	Symbol     string     `json:"symbol"`
	Type       string     `json:"type"`
	Opened     time.Time  `json:"opened"`
	Closed     *time.Time `json:"closed"`
	Strike     float64    `json:"strike"`
	Expiration time.Time  `json:"expiration"`
	Premium    float64    `json:"premium"`
	Contracts  int        `json:"contracts"`
	ExitPrice  *float64   `json:"exit_price"`
	Commission float64    `json:"commission"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (o *Option) CalculatePercentOTM(currentPrice float64) float64 {
	if currentPrice <= 0 {
		return 0
	}
	switch o.Type {
	case "Put":
		if currentPrice >= o.Strike {
			return math.Abs((currentPrice - o.Strike) / currentPrice * 100)
		}
	case "Call":
		if currentPrice <= o.Strike {
			return math.Abs((o.Strike - currentPrice) / currentPrice * 100)
		}
	}
	return 0
}

func (o *Option) CalculateDTE() int {
	if o.Expiration.Before(o.Opened) {
		return 0
	}
	return int(o.Expiration.Sub(o.Opened).Hours() / 24)
}

func (o *Option) CalculateDTC() int {
	if o.Closed == nil {
		return 0
	}
	return int(o.Closed.Sub(o.Opened).Hours() / 24)
}

func (o *Option) CalculateDaysRemaining() int {
	now := time.Now()
	// Use ceiling to avoid off-by-one error - tomorrow should show 1 day, not 0
	return int(math.Ceil(o.Expiration.Sub(now).Hours() / 24))
}

func (o *Option) CalculateTotalProfit() float64 {
	exitPrice := 0.0
	if o.ExitPrice != nil {
		exitPrice = *o.ExitPrice
	}
	profit := math.Floor((o.Premium - exitPrice) * float64(o.Contracts) * 100)

	// Commission is now stored as total in database
	// For open/expired positions: commission = per_contract * contracts
	// For closed positions (buy-to-close): commission = per_contract * contracts * 2
	return profit - o.Commission
}

func (o *Option) CalculatePercentOfProfit() float64 {
	if o.Premium == 0 {
		return 0
	}

	// MaxProfit = (Premium × Contracts × 100) - Total Commission
	// Commission is now stored as total in database
	// For open positions: commission is already the total opening commission
	// For closed positions: commission is already doubled (opening + closing)
	maxProfit := (o.Premium * float64(o.Contracts) * 100) - o.Commission
	actualProfit := o.CalculateTotalProfit()

	if maxProfit <= 0 {
		return 0
	}

	return (actualProfit / maxProfit) * 100
}

func (o *Option) CalculatePercentOfTime() float64 {
	totalDays := o.Expiration.Sub(o.Opened).Hours() / 24
	if totalDays <= 0 {
		return 0
	}

	// Use today's date for open positions, actual closed date for closed positions
	var endDate time.Time
	if o.Closed == nil {
		endDate = time.Now()
	} else {
		endDate = *o.Closed
	}

	usedDays := endDate.Sub(o.Opened).Hours() / 24

	// If 0 days used, assume 1 minimum
	if usedDays < 1 {
		usedDays = 1
	}

	// Don't allow percentage to exceed 100% even if we're past expiration
	percentOfTime := (usedDays / totalDays) * 100
	if percentOfTime > 100 {
		return 100
	}
	if percentOfTime < 0 {
		return 0
	}

	return percentOfTime
}

func (o *Option) CalculateMultiplier() float64 {
	percentTime := o.CalculatePercentOfTime()
	if percentTime == 0 {
		return 0
	}
	return o.CalculatePercentOfProfit() / percentTime
}

// CalculateAROI calculates the Annualized Return on Investment (AROI) for the option
// This extrapolates the profit to an annual basis based on time in trade
func (o *Option) CalculateAROI() float64 {
	// Calculate days the trade has been active
	var endDate time.Time
	if o.Closed == nil {
		endDate = time.Now()
	} else {
		endDate = *o.Closed
	}

	daysInTrade := endDate.Sub(o.Opened).Hours() / 24
	if daysInTrade <= 0 {
		daysInTrade = 1 // Minimum 1 day to avoid division by zero
	}

	// Calculate total profit
	profit := o.CalculateTotalProfit()

	// Calculate the capital base (exposure for puts, long value for calls)
	var capitalBase float64
	switch o.Type {
	case "Put":
		// For puts, use strike * contracts * 100 as the exposure/capital at risk
		capitalBase = o.Strike * float64(o.Contracts) * 100
	case "Call":
		// For calls, we need the underlying stock value, but we don't have current price here
		// Use strike as approximation for now - this should be enhanced with current price
		capitalBase = o.Strike * float64(o.Contracts) * 100
	}

	if capitalBase <= 0 {
		return 0
	}

	// Calculate return percentage for the period
	periodReturn := (profit / capitalBase) * 100

	// Annualize the return: (period return) * (365.25 days per year / days in trade)
	aroi := periodReturn * (365.25 / daysInTrade)

	return aroi
}

func (o *Option) GetExitPriceValue() float64 {
	if o.ExitPrice != nil {
		return *o.ExitPrice
	}
	return 0.0
}

// IsProfit returns true if the option generated a profit
func (o *Option) IsProfit() bool {
	return o.CalculateTotalProfit() > 0
}

// IsLoss returns true if the option generated a loss
func (o *Option) IsLoss() bool {
	return o.CalculateTotalProfit() < 0
}

// IsOpen returns true if the option position is still open
func (o *Option) IsOpen() bool {
	return o.Closed == nil
}

// IsClosed returns true if the option position is closed
func (o *Option) IsClosed() bool {
	return o.Closed != nil
}

// GetProfitLossClass returns CSS class for profit/loss styling
func (o *Option) GetProfitLossClass() string {
	if o.IsLoss() {
		return "negative"
	} else if o.IsProfit() {
		return "positive"
	}
	return ""
}

// GetFormattedTotalProfit returns formatted total profit with appropriate styling class
func (o *Option) GetFormattedTotalProfit() string {
	profit := o.CalculateTotalProfit()
	if profit < 0 {
		return fmt.Sprintf("<span class=\"negative\">$%.2f</span>", profit)
	}
	return fmt.Sprintf("$%.2f", profit)
}

// GetFormattedPercentOfProfit returns formatted percent of profit with appropriate styling
func (o *Option) GetFormattedPercentOfProfit() string {
	percent := o.CalculatePercentOfProfit()
	if percent < 0 {
		return fmt.Sprintf("<span class=\"negative\">%.2f%%</span>", percent)
	}
	return fmt.Sprintf("%.2f%%", percent)
}

type Dividend struct {
	ID        int       `json:"id"`
	Symbol    string    `json:"symbol"`
	Received  time.Time `json:"received"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

type SymbolService struct {
	db *sql.DB
}

func NewSymbolService(db *sql.DB) *SymbolService {
	return &SymbolService{db: db}
}

func (s *SymbolService) Create(symbol string) (*Symbol, error) {
	symbol = strings.TrimSpace(strings.ToUpper(symbol))
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	query := `INSERT INTO symbols (symbol) VALUES (?) RETURNING symbol, price, dividend, ex_dividend_date, pe_ratio, created_at, updated_at`
	var sym Symbol
	err := s.db.QueryRow(query, symbol).Scan(&sym.Symbol, &sym.Price, &sym.Dividend, &sym.ExDividendDate, &sym.PERatio, &sym.CreatedAt, &sym.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create symbol: %w", err)
	}

	return &sym, nil
}

func (s *SymbolService) GetBySymbol(symbol string) (*Symbol, error) {
	query := `SELECT symbol, price, dividend, ex_dividend_date, pe_ratio, created_at, updated_at FROM symbols WHERE symbol = ?`
	var sym Symbol
	err := s.db.QueryRow(query, symbol).Scan(&sym.Symbol, &sym.Price, &sym.Dividend, &sym.ExDividendDate, &sym.PERatio, &sym.CreatedAt, &sym.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("symbol not found")
		}
		return nil, fmt.Errorf("failed to get symbol: %w", err)
	}

	return &sym, nil
}

func (s *SymbolService) GetAll() ([]*Symbol, error) {
	query := `SELECT symbol, price, dividend, ex_dividend_date, pe_ratio, created_at, updated_at FROM symbols ORDER BY symbol`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbols: %w", err)
	}
	defer rows.Close()

	var symbols []*Symbol
	for rows.Next() {
		var symbol Symbol
		if err := rows.Scan(&symbol.Symbol, &symbol.Price, &symbol.Dividend, &symbol.ExDividendDate, &symbol.PERatio, &symbol.CreatedAt, &symbol.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan symbol: %w", err)
		}
		symbols = append(symbols, &symbol)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating symbols: %w", err)
	}

	return symbols, nil
}

// GetActivePositionSymbols returns only symbols with open positions (long or options)
// Use this for high-priority quick updates that focus on active trading
func (s *SymbolService) GetActivePositionSymbols() ([]string, error) {
	query := `
		SELECT DISTINCT symbol FROM (
			-- Symbols with open long positions
			SELECT symbol FROM long_positions WHERE closed IS NULL
			UNION
			-- Symbols with open options positions
			SELECT symbol FROM options WHERE closed IS NULL
		) ORDER BY symbol
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active position symbols: %w", err)
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, fmt.Errorf("failed to scan symbol: %w", err)
		}
		symbols = append(symbols, symbol)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating symbols: %w", err)
	}

	return symbols, nil
}

func (s *SymbolService) Update(symbol string, price float64, dividend float64, exDividendDate *time.Time, peRatio *float64) (*Symbol, error) {
	symbol = strings.TrimSpace(strings.ToUpper(symbol))
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	query := `UPDATE symbols SET price = ?, dividend = ?, ex_dividend_date = ?, pe_ratio = ?, updated_at = CURRENT_TIMESTAMP WHERE symbol = ? RETURNING symbol, price, dividend, ex_dividend_date, pe_ratio, created_at, updated_at`
	var sym Symbol
	err := s.db.QueryRow(query, price, dividend, exDividendDate, peRatio, symbol).Scan(&sym.Symbol, &sym.Price, &sym.Dividend, &sym.ExDividendDate, &sym.PERatio, &sym.CreatedAt, &sym.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("symbol not found")
		}
		return nil, fmt.Errorf("failed to update symbol: %w", err)
	}

	return &sym, nil
}

func (s *SymbolService) Delete(symbol string) error {
	query := `DELETE FROM symbols WHERE symbol = ?`
	result, err := s.db.Exec(query, symbol)
	if err != nil {
		return fmt.Errorf("failed to delete symbol: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("symbol not found")
	}

	return nil
}

func (s *SymbolService) GetDistinctSymbols() ([]string, error) {
	query := `
		SELECT DISTINCT symbol FROM (
			SELECT symbol FROM symbols
			UNION
			SELECT symbol FROM options
			UNION
			SELECT symbol FROM long_positions
			UNION
			SELECT symbol FROM dividends
		) ORDER BY symbol
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct symbols: %w", err)
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, fmt.Errorf("failed to scan symbol: %w", err)
		}
		symbols = append(symbols, symbol)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating symbols: %w", err)
	}

	return symbols, nil
}

// GetPrioritizedSymbols returns symbols ordered by importance:
// 1. Symbols with open long positions (highest priority)
// 2. Symbols with open options positions
// 3. All other symbols in the database (lowest priority)
func (s *SymbolService) GetPrioritizedSymbols() ([]string, error) {
	query := `
		WITH symbol_priorities AS (
			-- Priority 1: Symbols with open long positions
			SELECT DISTINCT symbol, 1 as priority
			FROM long_positions 
			WHERE closed IS NULL
			
			UNION
			
			-- Priority 2: Symbols with open options positions
			SELECT DISTINCT symbol, 2 as priority
			FROM options 
			WHERE closed IS NULL
			
			UNION
			
			-- Priority 3: All other symbols (from any table)
			SELECT DISTINCT symbol, 3 as priority
			FROM (
				SELECT symbol FROM symbols
				UNION
				SELECT symbol FROM options
				UNION
				SELECT symbol FROM long_positions
				UNION
				SELECT symbol FROM dividends
			) all_symbols
		)
		SELECT symbol
		FROM symbol_priorities
		GROUP BY symbol
		ORDER BY MIN(priority), symbol
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get prioritized symbols: %w", err)
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, fmt.Errorf("failed to scan symbol: %w", err)
		}
		symbols = append(symbols, symbol)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating symbols: %w", err)
	}

	return symbols, nil
}
