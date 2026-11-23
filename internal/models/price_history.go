package models

import (
	"database/sql"
	"fmt"
	"time"
)

type PriceBar struct {
	ID           int       `json:"id"`
	Symbol       string    `json:"symbol"`
	BarDate      time.Time `json:"bar_date"`
	Open         float64   `json:"open"`
	High         float64   `json:"high"`
	Low          float64   `json:"low"`
	Close        float64   `json:"close"`
	Volume       float64   `json:"volume"`
	VWAP         float64   `json:"vwap"`
	Transactions int       `json:"transactions"`
	CreatedAt    time.Time `json:"created_at"`
}

type PriceHistoryService struct {
	db *sql.DB
}

func NewPriceHistoryService(db *sql.DB) *PriceHistoryService {
	return &PriceHistoryService{db: db}
}

func (s *PriceHistoryService) Upsert(bar *PriceBar) error {
	query := `
		INSERT INTO price_history (symbol, bar_date, open, high, low, close, volume, vwap, transactions)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(symbol, bar_date) DO UPDATE SET
			open = excluded.open,
			high = excluded.high,
			low = excluded.low,
			close = excluded.close,
			volume = excluded.volume,
			vwap = excluded.vwap,
			transactions = excluded.transactions
	`
	_, err := s.db.Exec(query, bar.Symbol, bar.BarDate.Format("2006-01-02"), bar.Open, bar.High, bar.Low, bar.Close, bar.Volume, bar.VWAP, bar.Transactions)
	if err != nil {
		return fmt.Errorf("failed to upsert price bar: %w", err)
	}
	return nil
}

func (s *PriceHistoryService) GetBySymbol(symbol string, from, to time.Time) ([]*PriceBar, error) {
	query := `
		SELECT id, symbol, bar_date, open, high, low, close, volume, vwap, transactions, created_at
		FROM price_history
		WHERE symbol = ? AND bar_date >= ? AND bar_date <= ?
		ORDER BY bar_date ASC
	`
	rows, err := s.db.Query(query, symbol, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to query price history: %w", err)
	}
	defer rows.Close()

	var bars []*PriceBar
	for rows.Next() {
		var bar PriceBar
		if err := rows.Scan(&bar.ID, &bar.Symbol, &bar.BarDate, &bar.Open, &bar.High, &bar.Low, &bar.Close, &bar.Volume, &bar.VWAP, &bar.Transactions, &bar.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan price bar: %w", err)
		}
		bars = append(bars, &bar)
	}
	return bars, rows.Err()
}

func (s *PriceHistoryService) GetLatestDate(symbol string) (time.Time, error) {
	query := `SELECT COALESCE(MAX(DATE(bar_date)), '2000-01-01') FROM price_history WHERE symbol = ?`
	var dateStr string
	if err := s.db.QueryRow(query, symbol).Scan(&dateStr); err != nil {
		return time.Time{}, fmt.Errorf("failed to get latest date: %w", err)
	}
	return time.Parse("2006-01-02", dateStr)
}

func (s *PriceHistoryService) DeleteOlderThan(days int) (int64, error) {
	query := `DELETE FROM price_history WHERE bar_date < date('now', ?)`
	result, err := s.db.Exec(query, fmt.Sprintf("-%d days", days))
	if err != nil {
		return 0, fmt.Errorf("failed to delete old price bars: %w", err)
	}
	return result.RowsAffected()
}
