package utils

import (
	"fmt"
	"strings"
	"time"
)

// ParseDate parses date string in YYYY-MM-DD format
func ParseDate(dateStr, fieldName string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("%s is required", fieldName)
	}
	
	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid %s format (expected YYYY-MM-DD): %w", fieldName, err)
	}
	
	return parsed, nil
}

// ParseOptionalDate parses date string or returns zero time if empty
func ParseOptionalDate(dateStr, fieldName string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}
	
	return ParseDate(dateStr, fieldName)
}

// ParseMultipleDateFormats tries multiple date formats
func ParseMultipleDateFormats(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
	}
	
	for _, format := range formats {
		if parsed, err := time.Parse(format, dateStr); err == nil {
			return parsed, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date: %s (tried multiple formats)", dateStr)
}

// ValidateFilename checks for path traversal attacks
func ValidateFilename(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}
	
	if strings.Contains(filename, "..") {
		return fmt.Errorf("filename cannot contain '..'")
	}
	
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("filename cannot contain path separators")
	}
	
	return nil
}

// ValidateRequired checks if string is non-empty
func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidatePositive checks if float is positive
func ValidatePositive(value float64, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive", fieldName)
	}
	return nil
}

// ValidateNonNegative checks if float is non-negative
func ValidateNonNegative(value float64, fieldName string) error {
	if value < 0 {
		return fmt.Errorf("%s cannot be negative", fieldName)
	}
	return nil
}
