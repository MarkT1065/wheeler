package models

import (
	"database/sql"
	"fmt"
	"time"
)

type IndexBar struct {
	ID          int       `json:"id"`
	IndexTicker string    `json:"index_ticker"`
	BarDate     time.Time `json:"bar_date"`
	Open        float64   `json:"open"`
	High        float64   `json:"high"`
	Low         float64   `json:"low"`
	Close       float64   `json:"close"`
	CreatedAt   time.Time `json:"created_at"`
}

type IndexHistoryService struct {
	db *sql.DB
}

func NewIndexHistoryService(db *sql.DB) *IndexHistoryService {
	return &IndexHistoryService{db: db}
}

func (s *IndexHistoryService) Upsert(bar *IndexBar) error {
	query := `
		INSERT INTO index_history (index_ticker, bar_date, open, high, low, close)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(index_ticker, bar_date) DO UPDATE SET
			open = excluded.open,
			high = excluded.high,
			low = excluded.low,
			close = excluded.close
	`
	_, err := s.db.Exec(query, bar.IndexTicker, bar.BarDate.Format("2006-01-02"), bar.Open, bar.High, bar.Low, bar.Close)
	if err != nil {
		return fmt.Errorf("failed to upsert index bar: %w", err)
	}
	return nil
}

func (s *IndexHistoryService) GetByIndex(indexTicker string, from, to time.Time) ([]*IndexBar, error) {
	query := `
		SELECT id, index_ticker, bar_date, open, high, low, close, created_at
		FROM index_history
		WHERE index_ticker = ? AND bar_date >= ? AND bar_date <= ?
		ORDER BY bar_date ASC
	`
	rows, err := s.db.Query(query, indexTicker, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to query index history: %w", err)
	}
	defer rows.Close()

	var bars []*IndexBar
	for rows.Next() {
		var bar IndexBar
		if err := rows.Scan(&bar.ID, &bar.IndexTicker, &bar.BarDate, &bar.Open, &bar.High, &bar.Low, &bar.Close, &bar.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan index bar: %w", err)
		}
		bars = append(bars, &bar)
	}
	return bars, rows.Err()
}

func (s *IndexHistoryService) GetLatestDate(indexTicker string) (time.Time, error) {
	query := `SELECT COALESCE(MAX(DATE(bar_date)), '2023-02-14') FROM index_history WHERE index_ticker = ?`
	var dateStr string
	if err := s.db.QueryRow(query, indexTicker).Scan(&dateStr); err != nil {
		return time.Time{}, fmt.Errorf("failed to get latest date: %w", err)
	}
	return time.Parse("2006-01-02", dateStr)
}
