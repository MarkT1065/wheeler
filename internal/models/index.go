package models

import (
	"reflect"
	"sort"
	"strconv"
	"time"
)

// Index creates a nested index structure for all options
func Index(options []*Option) (map[string]interface{}, error) {
	index := map[string]interface{}{
		"id":         make(map[string]*Option),
		"symbol":     make(map[string][]*Option),
		"type":       make(map[string][]*Option),
		"opened":     make(map[int]map[time.Month][]*Option),
		"expiration": make(map[int]map[time.Month]map[int][]*Option),
		"open":       make([]*Option, 0),
		"closed":     make(map[int]map[time.Month][]*Option),
	}

	for _, option := range options {
		// Index by ID (as string key)
		idIndex := index["id"].(map[string]*Option)
		idIndex[strconv.Itoa(option.ID)] = option

		// Index by Symbol
		symbolIndex := index["symbol"].(map[string][]*Option)
		symbolIndex[option.Symbol] = append(symbolIndex[option.Symbol], option)

		// Index by Type
		typeIndex := index["type"].(map[string][]*Option)
		typeIndex[option.Type] = append(typeIndex[option.Type], option)

		// Index by Opened (year -> month -> options)
		openedIndex := index["opened"].(map[int]map[time.Month][]*Option)
		openYear := option.Opened.Year()
		openMonth := option.Opened.Month()

		if openedIndex[openYear] == nil {
			openedIndex[openYear] = make(map[time.Month][]*Option)
		}
		openedIndex[openYear][openMonth] = append(openedIndex[openYear][openMonth], option)

		// Index by Expiration (year -> month -> day -> options)
		expirationIndex := index["expiration"].(map[int]map[time.Month]map[int][]*Option)
		expYear := option.Expiration.Year()
		expMonth := option.Expiration.Month()
		expDay := option.Expiration.Day()

		if expirationIndex[expYear] == nil {
			expirationIndex[expYear] = make(map[time.Month]map[int][]*Option)
		}
		if expirationIndex[expYear][expMonth] == nil {
			expirationIndex[expYear][expMonth] = make(map[int][]*Option)
		}
		expirationIndex[expYear][expMonth][expDay] = append(expirationIndex[expYear][expMonth][expDay], option)

		// Index open positions (no closed date)
		if option.Closed == nil {
			openList := index["open"].([]*Option)
			index["open"] = append(openList, option)
		} else {
			// Index by Closed (year -> month -> options)
			closedIndex := index["closed"].(map[int]map[time.Month][]*Option)
			closedYear := option.Closed.Year()
			closedMonth := option.Closed.Month()

			if closedIndex[closedYear] == nil {
				closedIndex[closedYear] = make(map[time.Month][]*Option)
			}
			closedIndex[closedYear][closedMonth] = append(closedIndex[closedYear][closedMonth], option)
		}
	}

	return index, nil
}

// Compare compares two index structures for equality
func Compare(idx1, idx2 map[string]interface{}) bool {
	// Check if both maps have the same keys
	if len(idx1) != len(idx2) {
		return false
	}
	
	for key := range idx1 {
		if _, exists := idx2[key]; !exists {
			return false
		}
	}

	// Compare each index type
	for key := range idx1 {
		switch key {
		case "id":
			if !compareIDIndex(idx1[key].(map[string]*Option), idx2[key].(map[string]*Option)) {
				return false
			}
		case "symbol", "type":
			if !compareStringSliceIndex(idx1[key].(map[string][]*Option), idx2[key].(map[string][]*Option)) {
				return false
			}
		case "opened", "closed":
			if !compareTimeIndex(idx1[key].(map[int]map[time.Month][]*Option), idx2[key].(map[int]map[time.Month][]*Option)) {
				return false
			}
		case "expiration":
			if !compareExpirationIndex(idx1[key].(map[int]map[time.Month]map[int][]*Option), idx2[key].(map[int]map[time.Month]map[int][]*Option)) {
				return false
			}
		case "open":
			if !compareOpenIndex(idx1[key].([]*Option), idx2[key].([]*Option)) {
				return false
			}
		}
	}
	
	return true
}

// compareIDIndex compares ID-based indexes
func compareIDIndex(idx1, idx2 map[string]*Option) bool {
	if len(idx1) != len(idx2) {
		return false
	}
	
	for key, option1 := range idx1 {
		option2, exists := idx2[key]
		if !exists {
			return false
		}
		if !compareOptions(option1, option2) {
			return false
		}
	}
	
	return true
}

// compareStringSliceIndex compares string-keyed slice indexes (symbol, type)
func compareStringSliceIndex(idx1, idx2 map[string][]*Option) bool {
	if len(idx1) != len(idx2) {
		return false
	}
	
	for key, options1 := range idx1 {
		options2, exists := idx2[key]
		if !exists {
			return false
		}
		if !compareOptionSlices(options1, options2) {
			return false
		}
	}
	
	return true
}

// compareTimeIndex compares time-based indexes (opened, closed)
func compareTimeIndex(idx1, idx2 map[int]map[time.Month][]*Option) bool {
	if len(idx1) != len(idx2) {
		return false
	}
	
	for year, months1 := range idx1 {
		months2, exists := idx2[year]
		if !exists {
			return false
		}
		if len(months1) != len(months2) {
			return false
		}
		
		for month, options1 := range months1 {
			options2, exists := months2[month]
			if !exists {
				return false
			}
			if !compareOptionSlices(options1, options2) {
				return false
			}
		}
	}
	
	return true
}

// compareExpirationIndex compares the expiration index (year -> month -> day -> options)
func compareExpirationIndex(idx1, idx2 map[int]map[time.Month]map[int][]*Option) bool {
	if len(idx1) != len(idx2) {
		return false
	}
	
	for year, months1 := range idx1 {
		months2, exists := idx2[year]
		if !exists {
			return false
		}
		if len(months1) != len(months2) {
			return false
		}
		
		for month, days1 := range months1 {
			days2, exists := months2[month]
			if !exists {
				return false
			}
			if len(days1) != len(days2) {
				return false
			}
			
			for day, options1 := range days1 {
				options2, exists := days2[day]
				if !exists {
					return false
				}
				if !compareOptionSlices(options1, options2) {
					return false
				}
			}
		}
	}
	
	return true
}

// compareOpenIndex compares the open options slice
func compareOpenIndex(options1, options2 []*Option) bool {
	return compareOptionSlices(options1, options2)
}

// compareOptionSlices compares two slices of options, order-independent
func compareOptionSlices(options1, options2 []*Option) bool {
	if len(options1) != len(options2) {
		return false
	}
	
	// Sort both slices by ID for comparison
	sorted1 := make([]*Option, len(options1))
	sorted2 := make([]*Option, len(options2))
	copy(sorted1, options1)
	copy(sorted2, options2)
	
	sort.Slice(sorted1, func(i, j int) bool {
		return sorted1[i].ID < sorted1[j].ID
	})
	sort.Slice(sorted2, func(i, j int) bool {
		return sorted2[i].ID < sorted2[j].ID
	})
	
	for i, option1 := range sorted1 {
		if !compareOptions(option1, sorted2[i]) {
			return false
		}
	}
	
	return true
}

// compareOptions compares two individual options for equality
func compareOptions(opt1, opt2 *Option) bool {
	if opt1 == nil && opt2 == nil {
		return true
	}
	if opt1 == nil || opt2 == nil {
		return false
	}
	
	// Compare pointer addresses first (if they're the same object)
	if opt1 == opt2 {
		return true
	}
	
	// Compare all fields using reflection for deep equality
	return reflect.DeepEqual(opt1, opt2)
}

// FilterOptions represents filtering criteria for combined queries
type FilterOptions struct {
	Symbols     []string    `json:"symbols,omitempty"`     // Filter by specific symbols
	Types       []string    `json:"types,omitempty"`       // Filter by option types (Put/Call)
	Status      string      `json:"status,omitempty"`      // "open", "closed", or "all"
	DateRange   *DateRange  `json:"date_range,omitempty"`  // Filter by expiration date range
	OpenedRange *DateRange  `json:"opened_range,omitempty"` // Filter by opened date range
	ClosedRange *DateRange  `json:"closed_range,omitempty"` // Filter by closed date range
	StrikeRange *StrikeRange `json:"strike_range,omitempty"` // Filter by strike price range
}

type DateRange struct {
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
}

type StrikeRange struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

// GetByFilters returns options that match the combined filter criteria
func GetByFilters(index map[string]interface{}, filters FilterOptions) []*Option {
	var result []*Option
	
	// Start with all options if no specific filters, or get base set
	var baseOptions []*Option
	
	// If symbols are specified, start with symbol filter
	if len(filters.Symbols) > 0 {
		symbolIndex := index["symbol"].(map[string][]*Option)
		for _, symbol := range filters.Symbols {
			if options, exists := symbolIndex[symbol]; exists {
				baseOptions = append(baseOptions, options...)
			}
		}
	} else if len(filters.Types) > 0 {
		// Start with type filter if no symbols specified
		typeIndex := index["type"].(map[string][]*Option)
		for _, optionType := range filters.Types {
			if options, exists := typeIndex[optionType]; exists {
				baseOptions = append(baseOptions, options...)
			}
		}
	} else {
		// Start with all options from ID index
		idIndex := index["id"].(map[string]*Option)
		for _, option := range idIndex {
			baseOptions = append(baseOptions, option)
		}
	}
	
	// Apply additional filters
	for _, option := range baseOptions {
		if matchesFilters(option, filters) {
			result = append(result, option)
		}
	}
	
	// Remove duplicates and sort by ID
	result = removeDuplicateOptions(result)
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	
	return result
}

// matchesFilters checks if an option matches the filter criteria
func matchesFilters(option *Option, filters FilterOptions) bool {
	// Symbol filter
	if len(filters.Symbols) > 0 {
		symbolMatch := false
		for _, symbol := range filters.Symbols {
			if option.Symbol == symbol {
				symbolMatch = true
				break
			}
		}
		if !symbolMatch {
			return false
		}
	}
	
	// Type filter
	if len(filters.Types) > 0 {
		typeMatch := false
		for _, optionType := range filters.Types {
			if option.Type == optionType {
				typeMatch = true
				break
			}
		}
		if !typeMatch {
			return false
		}
	}
	
	// Status filter (open/closed)
	if filters.Status != "" && filters.Status != "all" {
		if filters.Status == "open" && option.Closed != nil {
			return false
		}
		if filters.Status == "closed" && option.Closed == nil {
			return false
		}
	}
	
	// Expiration date range filter
	if filters.DateRange != nil {
		if filters.DateRange.Start != nil && option.Expiration.Before(*filters.DateRange.Start) {
			return false
		}
		if filters.DateRange.End != nil && option.Expiration.After(*filters.DateRange.End) {
			return false
		}
	}
	
	// Opened date range filter
	if filters.OpenedRange != nil {
		if filters.OpenedRange.Start != nil && option.Opened.Before(*filters.OpenedRange.Start) {
			return false
		}
		if filters.OpenedRange.End != nil && option.Opened.After(*filters.OpenedRange.End) {
			return false
		}
	}
	
	// Closed date range filter
	if filters.ClosedRange != nil && option.Closed != nil {
		if filters.ClosedRange.Start != nil && option.Closed.Before(*filters.ClosedRange.Start) {
			return false
		}
		if filters.ClosedRange.End != nil && option.Closed.After(*filters.ClosedRange.End) {
			return false
		}
	}
	
	// Strike price range filter
	if filters.StrikeRange != nil {
		if filters.StrikeRange.Min != nil && option.Strike < *filters.StrikeRange.Min {
			return false
		}
		if filters.StrikeRange.Max != nil && option.Strike > *filters.StrikeRange.Max {
			return false
		}
	}
	
	return true
}

// removeDuplicateOptions removes duplicate options from a slice
func removeDuplicateOptions(options []*Option) []*Option {
	seen := make(map[int]bool)
	var result []*Option
	
	for _, option := range options {
		if !seen[option.ID] {
			seen[option.ID] = true
			result = append(result, option)
		}
	}
	
	return result
}

// Convenience methods for common filter combinations

// GetOpenOptionsBySymbol returns all open options for specific symbols
func GetOpenOptionsBySymbol(index map[string]interface{}, symbols ...string) []*Option {
	return GetByFilters(index, FilterOptions{
		Symbols: symbols,
		Status:  "open",
	})
}

// GetClosedOptionsByDateRange returns all closed options within a date range
func GetClosedOptionsByDateRange(index map[string]interface{}, start, end time.Time) []*Option {
	return GetByFilters(index, FilterOptions{
		Status: "closed",
		ClosedRange: &DateRange{
			Start: &start,
			End:   &end,
		},
	})
}

// GetOptionsByTypeAndSymbol returns options filtered by type and symbol
func GetOptionsByTypeAndSymbol(index map[string]interface{}, optionType, symbol string) []*Option {
	return GetByFilters(index, FilterOptions{
		Symbols: []string{symbol},
		Types:   []string{optionType},
	})
}

// GetOptionsExpiringInRange returns options expiring within a date range
func GetOptionsExpiringInRange(index map[string]interface{}, start, end time.Time) []*Option {
	return GetByFilters(index, FilterOptions{
		DateRange: &DateRange{
			Start: &start,
			End:   &end,
		},
	})
}
