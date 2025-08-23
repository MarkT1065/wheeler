package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type DividendService struct {
	db *sql.DB
}

func NewDividendService(db *sql.DB) *DividendService {
	return &DividendService{db: db}
}

func (s *DividendService) Create(symbol string, received time.Time, amount float64) (*Dividend, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("dividend amount must be positive")
	}

	query := `INSERT INTO dividends (symbol, received, amount) 
			  VALUES (?, ?, ?) 
			  RETURNING id, symbol, received, amount, created_at`

	var dividend Dividend
	err := s.db.QueryRow(query, symbol, received, amount).Scan(
		&dividend.ID, &dividend.Symbol, &dividend.Received, &dividend.Amount, &dividend.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create dividend: %w", err)
	}

	return &dividend, nil
}

func (s *DividendService) GetBySymbol(symbol string) ([]*Dividend, error) {
	query := `SELECT id, symbol, received, amount, created_at 
			  FROM dividends WHERE symbol = ? ORDER BY received DESC`

	rows, err := s.db.Query(query, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get dividends: %w", err)
	}
	defer rows.Close()

	var dividends []*Dividend
	for rows.Next() {
		var dividend Dividend
		if err := rows.Scan(&dividend.ID, &dividend.Symbol, &dividend.Received, &dividend.Amount, &dividend.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan dividend: %w", err)
		}
		dividends = append(dividends, &dividend)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating dividends: %w", err)
	}

	return dividends, nil
}

func (s *DividendService) GetAll() ([]*Dividend, error) {
	query := `SELECT id, symbol, received, amount, created_at 
			  FROM dividends ORDER BY received DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get dividends: %w", err)
	}
	defer rows.Close()

	var dividends []*Dividend
	for rows.Next() {
		var dividend Dividend
		if err := rows.Scan(&dividend.ID, &dividend.Symbol, &dividend.Received, &dividend.Amount, &dividend.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan dividend: %w", err)
		}
		dividends = append(dividends, &dividend)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating dividends: %w", err)
	}

	return dividends, nil
}

func (s *DividendService) GetByDateRange(symbol string, startDate, endDate time.Time) ([]*Dividend, error) {
	query := `SELECT id, symbol, received, amount, created_at 
			  FROM dividends 
			  WHERE symbol = ? AND received BETWEEN ? AND ? 
			  ORDER BY received DESC`

	rows, err := s.db.Query(query, symbol, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get dividends by date range: %w", err)
	}
	defer rows.Close()

	var dividends []*Dividend
	for rows.Next() {
		var dividend Dividend
		if err := rows.Scan(&dividend.ID, &dividend.Symbol, &dividend.Received, &dividend.Amount, &dividend.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan dividend: %w", err)
		}
		dividends = append(dividends, &dividend)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating dividends: %w", err)
	}

	return dividends, nil
}

func (s *DividendService) GetTotalForSymbol(symbol string) (float64, error) {
	query := `SELECT COALESCE(SUM(amount), 0) FROM dividends WHERE symbol = ?`

	var total float64
	err := s.db.QueryRow(query, symbol).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total dividends: %w", err)
	}

	return total, nil
}

func (s *DividendService) Delete(symbol string, received time.Time, amount float64) error {
	// Use ABS() function to handle floating-point precision issues
	// Allow for a small epsilon (0.001) in the comparison
	query := `DELETE FROM dividends WHERE symbol = ? AND received = ? AND ABS(amount - ?) < 0.001`
	result, err := s.db.Exec(query, symbol, received, amount)
	if err != nil {
		return fmt.Errorf("failed to delete dividend: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("dividend not found (symbol: %s, received: %s, amount: %.2f)", symbol, received.Format("2006-01-02"), amount)
	}

	return nil
}

func (s *DividendService) DeleteByID(id int) error {
	query := `DELETE FROM dividends WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete dividend: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("dividend not found with ID: %d", id)
	}

	return nil
}

func (s *DividendService) DeleteBySymbol(symbol string) error {
	query := `DELETE FROM dividends WHERE symbol = ?`
	result, err := s.db.Exec(query, symbol)
	if err != nil {
		return fmt.Errorf("failed to delete dividends for symbol %s: %w", symbol, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	log.Printf("Deleted %d dividends for symbol: %s", rowsAffected, symbol)
	return nil
}
