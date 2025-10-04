package models

import (
	"database/sql"
	"fmt"
	"time"
)

type LongPositionService struct {
	db *sql.DB
}

func NewLongPositionService(db *sql.DB) *LongPositionService {
	return &LongPositionService{db: db}
}

func (s *LongPositionService) Create(symbol string, opened time.Time, shares int, buyPrice float64) (*LongPosition, error) {
	query := `INSERT INTO long_positions (symbol, opened, shares, buy_price) 
			  VALUES (?, ?, ?, ?) 
			  RETURNING id, symbol, opened, closed, shares, buy_price, exit_price, created_at, updated_at`
	
	var position LongPosition
	err := s.db.QueryRow(query, symbol, opened, shares, buyPrice).Scan(
		&position.ID, &position.Symbol, &position.Opened, &position.Closed, &position.Shares,
		&position.BuyPrice, &position.ExitPrice, &position.CreatedAt, &position.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create long position: %w", err)
	}

	return &position, nil
}

func (s *LongPositionService) GetBySymbol(symbol string) ([]*LongPosition, error) {
	query := `SELECT id, symbol, opened, closed, shares, buy_price, exit_price, created_at, updated_at 
			  FROM long_positions WHERE symbol = ? ORDER BY opened DESC`
	
	rows, err := s.db.Query(query, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get long positions: %w", err)
	}
	defer rows.Close()

	var positions []*LongPosition
	for rows.Next() {
		var position LongPosition
		if err := rows.Scan(&position.ID, &position.Symbol, &position.Opened, &position.Closed, &position.Shares,
			&position.BuyPrice, &position.ExitPrice, &position.CreatedAt, &position.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan long position: %w", err)
		}
		positions = append(positions, &position)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating long positions: %w", err)
	}

	return positions, nil
}

func (s *LongPositionService) GetAll() ([]*LongPosition, error) {
	query := `SELECT id, symbol, opened, closed, shares, buy_price, exit_price, created_at, updated_at 
			  FROM long_positions ORDER BY opened DESC`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get long positions: %w", err)
	}
	defer rows.Close()

	var positions []*LongPosition
	for rows.Next() {
		var position LongPosition
		if err := rows.Scan(&position.ID, &position.Symbol, &position.Opened, &position.Closed, &position.Shares,
			&position.BuyPrice, &position.ExitPrice, &position.CreatedAt, &position.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan long position: %w", err)
		}
		positions = append(positions, &position)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating long positions: %w", err)
	}

	return positions, nil
}

func (s *LongPositionService) Close(symbol string, opened time.Time, shares int, buyPrice float64, closed time.Time, exitPrice float64) error {
	query := `UPDATE long_positions 
			  SET closed = ?, exit_price = ?, updated_at = CURRENT_TIMESTAMP 
			  WHERE symbol = ? AND opened = ? AND shares = ? AND buy_price = ?`
	
	result, err := s.db.Exec(query, closed, exitPrice, symbol, opened, shares, buyPrice)
	if err != nil {
		return fmt.Errorf("failed to close long position: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("long position not found")
	}

	return nil
}

func (s *LongPositionService) Delete(symbol string, opened time.Time, shares int, buyPrice float64) error {
	query := `DELETE FROM long_positions WHERE symbol = ? AND opened = ? AND shares = ? AND buy_price = ?`
	result, err := s.db.Exec(query, symbol, opened, shares, buyPrice)
	if err != nil {
		return fmt.Errorf("failed to delete long position: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("long position not found")
	}

	return nil
}

// GetByID retrieves a long position by its ID
func (s *LongPositionService) GetByID(id int) (*LongPosition, error) {
	query := `SELECT id, symbol, opened, closed, shares, buy_price, exit_price, created_at, updated_at 
			  FROM long_positions WHERE id = ?`
	
	var position LongPosition
	err := s.db.QueryRow(query, id).Scan(
		&position.ID, &position.Symbol, &position.Opened, &position.Closed, &position.Shares,
		&position.BuyPrice, &position.ExitPrice, &position.CreatedAt, &position.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("long position not found")
		}
		return nil, fmt.Errorf("failed to get long position: %w", err)
	}

	return &position, nil
}

// UpdateByID updates a long position by its ID
func (s *LongPositionService) UpdateByID(id int, symbol string, opened time.Time, shares int, buyPrice float64, closed *time.Time, exitPrice *float64) (*LongPosition, error) {
	query := `UPDATE long_positions 
			  SET symbol = ?, opened = ?, shares = ?, buy_price = ?, closed = ?, exit_price = ?, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = ? 
			  RETURNING id, symbol, opened, closed, shares, buy_price, exit_price, created_at, updated_at`
	
	var position LongPosition
	err := s.db.QueryRow(query, symbol, opened, shares, buyPrice, closed, exitPrice, id).Scan(
		&position.ID, &position.Symbol, &position.Opened, &position.Closed, &position.Shares,
		&position.BuyPrice, &position.ExitPrice, &position.CreatedAt, &position.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("long position not found")
		}
		return nil, fmt.Errorf("failed to update long position: %w", err)
	}

	return &position, nil
}

// DeleteByID deletes a long position by its ID
func (s *LongPositionService) DeleteByID(id int) error {
	query := `DELETE FROM long_positions WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete long position: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("long position not found")
	}

	return nil
}

// CloseByID closes a long position by its ID
func (s *LongPositionService) CloseByID(id int, closed time.Time, exitPrice float64) error {
	query := `UPDATE long_positions 
			  SET closed = ?, exit_price = ?, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = ?`
	
	result, err := s.db.Exec(query, closed, exitPrice, id)
	if err != nil {
		return fmt.Errorf("failed to close long position: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("long position not found")
	}

	return nil
}

func (s *LongPositionService) DeleteBySymbol(symbol string) error {
	query := `DELETE FROM long_positions WHERE symbol = ?`
	result, err := s.db.Exec(query, symbol)
	if err != nil {
		return fmt.Errorf("failed to delete long positions for symbol %s: %w", symbol, err)
	}

	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	return nil
}

// GetOpenPositions retrieves all open long positions (where closed is NULL)
func (s *LongPositionService) GetOpenPositions() ([]*LongPosition, error) {
	query := `SELECT id, symbol, opened, closed, shares, buy_price, exit_price, created_at, updated_at 
			  FROM long_positions WHERE closed IS NULL ORDER BY opened DESC`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get open long positions: %w", err)
	}
	defer rows.Close()

	var positions []*LongPosition
	for rows.Next() {
		var position LongPosition
		if err := rows.Scan(&position.ID, &position.Symbol, &position.Opened, &position.Closed, &position.Shares,
			&position.BuyPrice, &position.ExitPrice, &position.CreatedAt, &position.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan open long position: %w", err)
		}
		positions = append(positions, &position)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating open long positions: %w", err)
	}

	return positions, nil
}