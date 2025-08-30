package models

import (
	"database/sql"
	"fmt"
	"time"
)

// MetricType represents the type of metric being tracked
type MetricType string

const (

	// TreasuryValue is the total value of open treasury transactions
	TreasuryValue MetricType = "treasury_value"

	// LongValue is the total value of all open long_positions
	LongValue MetricType = "long_value"

	// LongCount is the count of open long_positions
	LongCount MetricType = "long_count"

	// PutExposure is the total value of all open Put options
	PutExposure MetricType = "put_exposure"

	// OpenPutPremium is the total of all open Put option premiums
	OpenPutPremium MetricType = "open_put_premium"

	// OpenPutCount is the count of open Put options
	OpenPutCount MetricType = "open_put_count"

	//OpenCallPremium is the total of all open Call option premiums
	OpenCallPremium MetricType = "open_call_premium"

	// OpenCallCount is the count of open Call options
	OpenCallCount MetricType = "open_call_count"
)

type Metric struct {
	ID      int        `json:"id"`
	Created time.Time  `json:"created"`
	Type    MetricType `json:"type"`
	Value   float64    `json:"value"`
}

type MetricService struct {
	db *sql.DB
}

func NewMetricService(db *sql.DB) *MetricService {
	return &MetricService{db: db}
}

func (ms *MetricService) Create(metricType MetricType, value float64) (*Metric, error) {
	if metricType == "" {
		return nil, fmt.Errorf("metric type cannot be empty")
	}

	query := `INSERT INTO metrics (type, value) VALUES (?, ?) RETURNING id, created, type, value`
	var metric Metric
	err := ms.db.QueryRow(query, string(metricType), value).Scan(&metric.ID, &metric.Created, &metric.Type, &metric.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric: %w", err)
	}

	return &metric, nil
}

func (ms *MetricService) GetByID(id int) (*Metric, error) {
	query := `SELECT id, created, type, value FROM metrics WHERE id = ?`
	var metric Metric
	err := ms.db.QueryRow(query, id).Scan(&metric.ID, &metric.Created, &metric.Type, &metric.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("metric not found")
		}
		return nil, fmt.Errorf("failed to get metric: %w", err)
	}

	return &metric, nil
}

func (ms *MetricService) GetAll() ([]*Metric, error) {
	query := `SELECT id, created, type, value FROM metrics ORDER BY created DESC`
	rows, err := ms.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*Metric
	for rows.Next() {
		var metric Metric
		if err := rows.Scan(&metric.ID, &metric.Created, &metric.Type, &metric.Value); err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}
		metrics = append(metrics, &metric)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metrics: %w", err)
	}

	return metrics, nil
}

func (ms *MetricService) GetByType(metricType MetricType) ([]*Metric, error) {
	query := `SELECT id, created, type, value FROM metrics WHERE type = ? ORDER BY created DESC`
	rows, err := ms.db.Query(query, string(metricType))
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics by type: %w", err)
	}
	defer rows.Close()

	var metrics []*Metric
	for rows.Next() {
		var metric Metric
		if err := rows.Scan(&metric.ID, &metric.Created, &metric.Type, &metric.Value); err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}
		metrics = append(metrics, &metric)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metrics: %w", err)
	}

	return metrics, nil
}

func (ms *MetricService) Update(id int, value float64) (*Metric, error) {
	query := `UPDATE metrics SET value = ? WHERE id = ? RETURNING id, created, type, value`
	var metric Metric
	err := ms.db.QueryRow(query, value, id).Scan(&metric.ID, &metric.Created, &metric.Type, &metric.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("metric not found")
		}
		return nil, fmt.Errorf("failed to update metric: %w", err)
	}

	return &metric, nil
}

func (ms *MetricService) Delete(id int) error {
	query := `DELETE FROM metrics WHERE id = ?`
	result, err := ms.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete metric: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("metric not found")
	}

	return nil
}

// CreateSnapshot creates a complete snapshot of all metric types
func (ms *MetricService) CreateSnapshot(metrics map[MetricType]float64) ([]*Metric, error) {
	var createdMetrics []*Metric

	for metricType, value := range metrics {
		metric, err := ms.Create(metricType, value)
		if err != nil {
			return nil, fmt.Errorf("failed to create metric %s: %w", metricType, err)
		}
		createdMetrics = append(createdMetrics, metric)
	}

	return createdMetrics, nil
}

// ComprehensiveSnapshot creates historical snapshots for each day going back the specified number of days
func (ms *MetricService) ComprehensiveSnapshot(days int) error {
	if days <= 0 {
		return fmt.Errorf("days must be positive")
	}

	// Get today's date and calculate the start date
	today := time.Now()
	
	// For each day in the range, calculate and upsert all metrics
	for i := 0; i < days; i++ {
		targetDate := today.AddDate(0, 0, -i)
		
		// Calculate and upsert treasury value
		treasuryValue, err := ms.calculateTreasuryValueForDate(targetDate)
		if err != nil {
			return fmt.Errorf("failed to calculate treasury value for %s: %w", targetDate.Format("2006-01-02"), err)
		}
		if err = ms.upsertMetricForDate(TreasuryValue, treasuryValue, targetDate); err != nil {
			return fmt.Errorf("failed to upsert treasury metric for %s: %w", targetDate.Format("2006-01-02"), err)
		}

		// Calculate and upsert long value
		longValue, err := ms.calculateLongValueForDate(targetDate)
		if err != nil {
			return fmt.Errorf("failed to calculate long value for %s: %w", targetDate.Format("2006-01-02"), err)
		}
		if err = ms.upsertMetricForDate(LongValue, longValue, targetDate); err != nil {
			return fmt.Errorf("failed to upsert long value metric for %s: %w", targetDate.Format("2006-01-02"), err)
		}

		// Calculate and upsert long count
		longCount, err := ms.calculateLongCountForDate(targetDate)
		if err != nil {
			return fmt.Errorf("failed to calculate long count for %s: %w", targetDate.Format("2006-01-02"), err)
		}
		if err = ms.upsertMetricForDate(LongCount, longCount, targetDate); err != nil {
			return fmt.Errorf("failed to upsert long count metric for %s: %w", targetDate.Format("2006-01-02"), err)
		}

		// Calculate and upsert put exposure
		putExposure, err := ms.calculatePutExposureForDate(targetDate)
		if err != nil {
			return fmt.Errorf("failed to calculate put exposure for %s: %w", targetDate.Format("2006-01-02"), err)
		}
		if err = ms.upsertMetricForDate(PutExposure, putExposure, targetDate); err != nil {
			return fmt.Errorf("failed to upsert put exposure metric for %s: %w", targetDate.Format("2006-01-02"), err)
		}

		// Calculate and upsert open put premium
		openPutPremium, err := ms.calculateOpenPutPremiumForDate(targetDate)
		if err != nil {
			return fmt.Errorf("failed to calculate open put premium for %s: %w", targetDate.Format("2006-01-02"), err)
		}
		if err = ms.upsertMetricForDate(OpenPutPremium, openPutPremium, targetDate); err != nil {
			return fmt.Errorf("failed to upsert open put premium metric for %s: %w", targetDate.Format("2006-01-02"), err)
		}

		// Calculate and upsert open put count
		openPutCount, err := ms.calculateOpenPutCountForDate(targetDate)
		if err != nil {
			return fmt.Errorf("failed to calculate open put count for %s: %w", targetDate.Format("2006-01-02"), err)
		}
		if err = ms.upsertMetricForDate(OpenPutCount, openPutCount, targetDate); err != nil {
			return fmt.Errorf("failed to upsert open put count metric for %s: %w", targetDate.Format("2006-01-02"), err)
		}

		// Calculate and upsert open call premium
		openCallPremium, err := ms.calculateOpenCallPremiumForDate(targetDate)
		if err != nil {
			return fmt.Errorf("failed to calculate open call premium for %s: %w", targetDate.Format("2006-01-02"), err)
		}
		if err = ms.upsertMetricForDate(OpenCallPremium, openCallPremium, targetDate); err != nil {
			return fmt.Errorf("failed to upsert open call premium metric for %s: %w", targetDate.Format("2006-01-02"), err)
		}

		// Calculate and upsert open call count
		openCallCount, err := ms.calculateOpenCallCountForDate(targetDate)
		if err != nil {
			return fmt.Errorf("failed to calculate open call count for %s: %w", targetDate.Format("2006-01-02"), err)
		}
		if err = ms.upsertMetricForDate(OpenCallCount, openCallCount, targetDate); err != nil {
			return fmt.Errorf("failed to upsert open call count metric for %s: %w", targetDate.Format("2006-01-02"), err)
		}
	}

	return nil
}

// calculateTreasuryValueForDate calculates total treasury value as of a specific date
func (ms *MetricService) calculateTreasuryValueForDate(date time.Time) (float64, error) {
	// Query for treasuries that were active on the given date
	// Since treasuries don't have a sold_date field, we need to handle this differently:
	// - Include treasuries that were purchased on or before the target date
	// - Only include treasuries that haven't been sold (exit_price IS NULL)
	// - For historical accuracy, treasuries with exit_price should be excluded from current calculations
	//   but included in historical dates before they were sold (assuming sold at maturity for historical data)
	
	// For current date calculations, only include unsold treasuries
	if date.Format("2006-01-02") == time.Now().Format("2006-01-02") {
		query := `
			SELECT COALESCE(SUM(amount), 0) as total_value
			FROM treasuries 
			WHERE date(purchased) <= date(?) 
			AND exit_price IS NULL
		`
		dateStr := date.Format("2006-01-02")
		var totalValue float64
		err := ms.db.QueryRow(query, dateStr).Scan(&totalValue)
		if err != nil {
			return 0, fmt.Errorf("failed to calculate current treasury value: %w", err)
		}
		return totalValue, nil
	}
	
	// For historical dates, use the original logic
	query := `
		SELECT COALESCE(SUM(amount), 0) as total_value
		FROM treasuries 
		WHERE date(purchased) <= date(?) 
		AND (exit_price IS NULL OR date(maturity) > date(?))
	`
	
	dateStr := date.Format("2006-01-02")
	var totalValue float64
	err := ms.db.QueryRow(query, dateStr, dateStr).Scan(&totalValue)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate treasury value: %w", err)
	}
	
	return totalValue, nil
}

// calculateLongValueForDate calculates total long position value as of a specific date
func (ms *MetricService) calculateLongValueForDate(date time.Time) (float64, error) {
	// Query for long positions that were active on the given date
	// Active means: opened <= date AND (closed IS NULL OR closed > date)
	// Value = shares * buy_price
	query := `
		SELECT COALESCE(SUM(shares * buy_price), 0) as total_value
		FROM long_positions 
		WHERE date(opened) <= date(?) 
		AND (closed IS NULL OR date(closed) > date(?))
	`
	
	dateStr := date.Format("2006-01-02")
	var totalValue float64
	err := ms.db.QueryRow(query, dateStr, dateStr).Scan(&totalValue)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate long value: %w", err)
	}
	
	return totalValue, nil
}

// calculateLongCountForDate calculates total count of long positions as of a specific date
func (ms *MetricService) calculateLongCountForDate(date time.Time) (float64, error) {
	// Query for long positions that were active on the given date
	// Active means: opened <= date AND (closed IS NULL OR closed > date)
	query := `
		SELECT COALESCE(COUNT(*), 0) as total_count
		FROM long_positions 
		WHERE date(opened) <= date(?) 
		AND (closed IS NULL OR date(closed) > date(?))
	`
	
	dateStr := date.Format("2006-01-02")
	var totalCount int64
	err := ms.db.QueryRow(query, dateStr, dateStr).Scan(&totalCount)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate long count: %w", err)
	}
	
	return float64(totalCount), nil
}

// calculatePutExposureForDate calculates total put option exposure as of a specific date
func (ms *MetricService) calculatePutExposureForDate(date time.Time) (float64, error) {
	// Query for put options that were active on the given date
	// Active means: opened <= date AND (closed IS NULL OR closed > date) AND type = 'Put'
	// Exposure = strike * contracts * 100 (standard option contract multiplier)
	query := `
		SELECT COALESCE(SUM(strike * contracts * 100), 0) as total_exposure
		FROM options 
		WHERE date(opened) <= date(?) 
		AND (closed IS NULL OR date(closed) > date(?))
		AND type = 'Put'
	`
	
	dateStr := date.Format("2006-01-02")
	var totalExposure float64
	err := ms.db.QueryRow(query, dateStr, dateStr).Scan(&totalExposure)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate put exposure: %w", err)
	}
	
	return totalExposure, nil
}

// calculateOpenPutPremiumForDate calculates total premium value of open put options as of a specific date
func (ms *MetricService) calculateOpenPutPremiumForDate(date time.Time) (float64, error) {
	// Query for put options that were active on the given date
	// Active means: opened <= date AND (closed IS NULL OR closed > date) AND type = 'Put'
	// Premium value = premium * contracts * 100 (standard option contract multiplier)
	query := `
		SELECT COALESCE(SUM(premium * contracts * 100), 0) as total_premium
		FROM options 
		WHERE date(opened) <= date(?) 
		AND (closed IS NULL OR date(closed) > date(?))
		AND type = 'Put'
	`
	
	dateStr := date.Format("2006-01-02")
	var totalPremium float64
	err := ms.db.QueryRow(query, dateStr, dateStr).Scan(&totalPremium)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate open put premium: %w", err)
	}
	
	return totalPremium, nil
}

// calculateOpenPutCountForDate calculates total count of open put options as of a specific date
func (ms *MetricService) calculateOpenPutCountForDate(date time.Time) (float64, error) {
	// Query for put options that were active on the given date
	// Active means: opened <= date AND (closed IS NULL OR closed > date) AND type = 'Put'
	query := `
		SELECT COALESCE(COUNT(*), 0) as total_count
		FROM options 
		WHERE date(opened) <= date(?) 
		AND (closed IS NULL OR date(closed) > date(?))
		AND type = 'Put'
	`
	
	dateStr := date.Format("2006-01-02")
	var totalCount int64
	err := ms.db.QueryRow(query, dateStr, dateStr).Scan(&totalCount)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate open put count: %w", err)
	}
	
	return float64(totalCount), nil
}

// calculateOpenCallPremiumForDate calculates total premium value of open call options as of a specific date
func (ms *MetricService) calculateOpenCallPremiumForDate(date time.Time) (float64, error) {
	// Query for call options that were active on the given date
	// Active means: opened <= date AND (closed IS NULL OR closed > date) AND type = 'Call'
	// Premium value = premium * contracts * 100 (standard option contract multiplier)
	query := `
		SELECT COALESCE(SUM(premium * contracts * 100), 0) as total_premium
		FROM options 
		WHERE date(opened) <= date(?) 
		AND (closed IS NULL OR date(closed) > date(?))
		AND type = 'Call'
	`
	
	dateStr := date.Format("2006-01-02")
	var totalPremium float64
	err := ms.db.QueryRow(query, dateStr, dateStr).Scan(&totalPremium)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate open call premium: %w", err)
	}
	
	return totalPremium, nil
}

// calculateOpenCallCountForDate calculates total count of open call options as of a specific date
func (ms *MetricService) calculateOpenCallCountForDate(date time.Time) (float64, error) {
	// Query for call options that were active on the given date
	// Active means: opened <= date AND (closed IS NULL OR closed > date) AND type = 'Call'
	query := `
		SELECT COALESCE(COUNT(*), 0) as total_count
		FROM options 
		WHERE date(opened) <= date(?) 
		AND (closed IS NULL OR date(closed) > date(?))
		AND type = 'Call'
	`
	
	dateStr := date.Format("2006-01-02")
	var totalCount int64
	err := ms.db.QueryRow(query, dateStr, dateStr).Scan(&totalCount)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate open call count: %w", err)
	}
	
	return float64(totalCount), nil
}

// upsertMetricForDate inserts or updates a metric for a specific date
func (ms *MetricService) upsertMetricForDate(metricType MetricType, value float64, date time.Time) error {
	// First, try to find an existing metric for this date and type
	dateStr := date.Format("2006-01-02")
	
	var existingID int
	checkQuery := `SELECT id FROM metrics WHERE type = ? AND date(created) = date(?)`
	err := ms.db.QueryRow(checkQuery, string(metricType), dateStr).Scan(&existingID)
	
	if err == sql.ErrNoRows {
		// No existing metric, insert new one with the target date
		// Set time to noon for consistent historical snapshots
		dateWithTime := time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, date.Location())
		insertQuery := `INSERT INTO metrics (created, type, value) VALUES (?, ?, ?)`
		_, err = ms.db.Exec(insertQuery, dateWithTime, string(metricType), value)
		if err != nil {
			return fmt.Errorf("failed to insert metric: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check existing metric: %w", err)
	} else {
		// Update existing metric
		updateQuery := `UPDATE metrics SET value = ? WHERE id = ?`
		_, err = ms.db.Exec(updateQuery, value, existingID)
		if err != nil {
			return fmt.Errorf("failed to update existing metric: %w", err)
		}
	}
	
	return nil
}
