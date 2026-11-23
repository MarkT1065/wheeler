package models

import (
	"database/sql"
	"fmt"
	"time"
)

type OptionBar struct {
	ID               int       `json:"id"`
	OptionTicker     string    `json:"option_ticker"`
	UnderlyingSymbol string    `json:"underlying_symbol"`
	ContractType     string    `json:"contract_type"`
	StrikePrice      float64   `json:"strike_price"`
	ExpirationDate   time.Time `json:"expiration_date"`
	BarDate          time.Time `json:"bar_date"`
	Open             float64   `json:"open"`
	High             float64   `json:"high"`
	Low              float64   `json:"low"`
	Close            float64   `json:"close"`
	Volume           float64   `json:"volume"`
	VWAP             float64   `json:"vwap"`
	Transactions     int       `json:"transactions"`
	CreatedAt        time.Time `json:"created_at"`
}

type OptionHistoryService struct {
	db *sql.DB
}

func NewOptionHistoryService(db *sql.DB) *OptionHistoryService {
	return &OptionHistoryService{db: db}
}

func (s *OptionHistoryService) Upsert(bar *OptionBar) error {
	query := `
		INSERT INTO option_history (option_ticker, underlying_symbol, contract_type, strike_price, expiration_date, bar_date, open, high, low, close, volume, vwap, transactions)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(option_ticker, bar_date) DO UPDATE SET
			open = excluded.open,
			high = excluded.high,
			low = excluded.low,
			close = excluded.close,
			volume = excluded.volume,
			vwap = excluded.vwap,
			transactions = excluded.transactions
	`
	_, err := s.db.Exec(query, bar.OptionTicker, bar.UnderlyingSymbol, bar.ContractType, bar.StrikePrice, bar.ExpirationDate.Format("2006-01-02"), bar.BarDate.Format("2006-01-02"), bar.Open, bar.High, bar.Low, bar.Close, bar.Volume, bar.VWAP, bar.Transactions)
	if err != nil {
		return fmt.Errorf("failed to upsert option bar: %w", err)
	}
	return nil
}

func (s *OptionHistoryService) GetByOptionTicker(optionTicker string, from, to time.Time) ([]*OptionBar, error) {
	query := `
		SELECT id, option_ticker, underlying_symbol, contract_type, strike_price, expiration_date, bar_date, open, high, low, close, volume, vwap, transactions, created_at
		FROM option_history
		WHERE option_ticker = ? AND bar_date >= ? AND bar_date <= ?
		ORDER BY bar_date ASC
	`
	return s.queryBars(query, optionTicker, from, to)
}

func (s *OptionHistoryService) GetByUnderlying(symbol string, from, to time.Time) ([]*OptionBar, error) {
	query := `
		SELECT id, option_ticker, underlying_symbol, contract_type, strike_price, expiration_date, bar_date, open, high, low, close, volume, vwap, transactions, created_at
		FROM option_history
		WHERE underlying_symbol = ? AND bar_date >= ? AND bar_date <= ?
		ORDER BY bar_date ASC
	`
	return s.queryBars(query, symbol, from, to)
}

func (s *OptionHistoryService) GetLatestDate(optionTicker string) (time.Time, error) {
	query := `SELECT COALESCE(MAX(DATE(bar_date)), '2000-01-01') FROM option_history WHERE option_ticker = ?`
	var dateStr string
	if err := s.db.QueryRow(query, optionTicker).Scan(&dateStr); err != nil {
		return time.Time{}, fmt.Errorf("failed to get latest date: %w", err)
	}
	return time.Parse("2006-01-02", dateStr)
}

func (s *OptionHistoryService) DeleteExpiredOlderThan(days int) (int64, error) {
	query := `DELETE FROM option_history WHERE expiration_date < date('now', ?)`
	result, err := s.db.Exec(query, fmt.Sprintf("-%d days", days))
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired option bars: %w", err)
	}
	return result.RowsAffected()
}

func (s *OptionHistoryService) queryBars(query string, key interface{}, from, to time.Time) ([]*OptionBar, error) {
	rows, err := s.db.Query(query, key, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to query option history: %w", err)
	}
	defer rows.Close()

	var bars []*OptionBar
	for rows.Next() {
		var bar OptionBar
		if err := rows.Scan(&bar.ID, &bar.OptionTicker, &bar.UnderlyingSymbol, &bar.ContractType, &bar.StrikePrice, &bar.ExpirationDate, &bar.BarDate, &bar.Open, &bar.High, &bar.Low, &bar.Close, &bar.Volume, &bar.VWAP, &bar.Transactions, &bar.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan option bar: %w", err)
		}
		bars = append(bars, &bar)
	}
	return bars, rows.Err()
}
