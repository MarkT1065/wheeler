package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"stonks/internal/models"
	"time"
)

// dividendsAPIHandler handles CRUD operations for dividends
func (s *Server) dividendsAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.createDividendHandler(w, r)
	case http.MethodDelete:
		s.deleteDividendHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// createDividendHandler handles POST requests to create new dividends
func (s *Server) createDividendHandler(w http.ResponseWriter, r *http.Request) {
	var req DividendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Parse the received date
	receivedDate, err := time.Parse("2006-01-02", req.Received)
	if err != nil {
		http.Error(w, "Invalid received date format", http.StatusBadRequest)
		return
	}

	// Validate amount
	if req.Amount <= 0 {
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return
	}

	// Create the dividend
	dividend, err := s.dividendService.Create(req.Symbol, receivedDate, req.Amount)
	if err != nil {
		http.Error(w, "Failed to create dividend", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dividend)
}

// deleteDividendHandler handles DELETE requests to remove dividends
func (s *Server) deleteDividendHandler(w http.ResponseWriter, r *http.Request) {
	var req DividendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Delete the dividend using ID if provided, otherwise use compound key
	if req.ID != nil && *req.ID > 0 {
		// Delete by ID (preferred method)
		err := s.dividendService.DeleteByID(*req.ID)
		if err != nil {
			log.Printf("Error deleting dividend by ID: %v", err)
			http.Error(w, fmt.Sprintf("Failed to delete dividend: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		// Fallback to compound key deletion
		receivedDate, err := time.Parse("2006-01-02", req.Received)
		if err != nil {
			http.Error(w, "Invalid received date format", http.StatusBadRequest)
			return
		}

		err = s.dividendService.Delete(req.Symbol, receivedDate, req.Amount)
		if err != nil {
			log.Printf("Error deleting dividend: %v", err)
			http.Error(w, fmt.Sprintf("Failed to delete dividend: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

// dividendsHandler handles GET requests for the dividends page
func (s *Server) dividendsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DIVIDENDS] Starting dividends page handler")

	// Get all open long positions (positions that can earn dividends)
	openPositions, err := s.longPositionService.GetOpenPositions()
	if err != nil {
		log.Printf("[DIVIDENDS] Error getting open positions: %v", err)
		http.Error(w, "Failed to load open positions", http.StatusInternalServerError)
		return
	}

	// Build dividend data for symbols with open positions
	// Use a map to aggregate positions by symbol
	dividendSymbolsMap := make(map[string]*DividendSymbolData)
	var totalAnnualIncome float64
	var totalYield float64
	var yieldCount int

	// Map for pie chart data
	incomeBySymbolMap := make(map[string]float64)

	for _, position := range openPositions {
		// Get symbol data for dividend and price info
		symbol, err := s.symbolService.GetBySymbol(position.Symbol)
		if err != nil {
			log.Printf("[DIVIDENDS] Error getting symbol %s: %v", position.Symbol, err)
			continue
		}

		// Check if symbol has dividend records
		dividends, err := s.dividendService.GetBySymbol(position.Symbol)
		if err != nil {
			log.Printf("[DIVIDENDS] Error getting dividends for %s: %v", position.Symbol, err)
			continue
		}

		// Only include symbols that pay dividends or have dividend data
		if symbol.Dividend > 0 || len(dividends) > 0 {
			// Calculate annual yield (quarterly dividend x 4)
			annualDividend := symbol.Dividend * 4
			var yieldPercent float64
			if symbol.Price > 0 && annualDividend > 0 {
				yieldPercent = (annualDividend / symbol.Price) * 100
			}

			// Calculate total dividend income: shares x quarterly dividend x 4
			positionAnnualIncome := float64(position.Shares) * annualDividend
			totalAnnualIncome += positionAnnualIncome

			// Check if we already have this symbol in the map
			if existing, exists := dividendSymbolsMap[position.Symbol]; exists {
				// Aggregate shares and income
				existing.Shares += position.Shares
				existing.TotalAnnualIncome += positionAnnualIncome
				existing.Positions = append(existing.Positions, position)
				incomeBySymbolMap[position.Symbol] += positionAnnualIncome
			} else {
				// Track for pie chart
				incomeBySymbolMap[position.Symbol] = positionAnnualIncome

				// Track for average yield (only count unique symbols)
				if yieldPercent > 0 {
					totalYield += yieldPercent
					yieldCount++
				}

				// Create new entry
				dividendSymbolsMap[position.Symbol] = &DividendSymbolData{
					Symbol:            position.Symbol,
					Price:             symbol.Price,
					Dividend:          symbol.Dividend,
					AnnualDividend:    annualDividend,
					YieldPercent:      yieldPercent,
					ExDividendDate:    symbol.ExDividendDate,
					DividendCount:     len(dividends),
					Shares:            position.Shares,
					TotalAnnualIncome: positionAnnualIncome,
					Positions:         []*models.LongPosition{position},
					DividendPayments:  dividends,
				}
			}
		}
	}

	// Convert map to slice and sort alphabetically
	var dividendSymbols []DividendSymbolData
	for _, data := range dividendSymbolsMap {
		dividendSymbols = append(dividendSymbols, *data)
	}
	sort.Slice(dividendSymbols, func(i, j int) bool {
		return dividendSymbols[i].Symbol < dividendSymbols[j].Symbol
	})

	// Build pie chart data for income by symbol (sorted alphabetically)
	var symbolsSorted []string
	for symbol := range incomeBySymbolMap {
		symbolsSorted = append(symbolsSorted, symbol)
	}
	sort.Strings(symbolsSorted)

	var incomeBySymbol []ChartData
	colors := []string{"#3498db", "#2ecc71", "#e74c3c", "#f39c12", "#9b59b6", "#1abc9c", "#34495e", "#e67e22"}
	for i, symbol := range symbolsSorted {
		incomeBySymbol = append(incomeBySymbol, ChartData{
			Label: symbol,
			Value: incomeBySymbolMap[symbol],
			Color: colors[i%len(colors)],
		})
	}

	// Calculate historical dividends by month and symbol
	allDividends, err := s.dividendService.GetAll()
	if err != nil {
		log.Printf("[DIVIDENDS] Error getting all dividends: %v", err)
	}

	type monthSymbolKey struct {
		month  string
		symbol string
	}

	monthlyBySymbol := make(map[monthSymbolKey]float64)
	monthsSet := make(map[string]bool)
	symbolsSet := make(map[string]bool)

	for _, div := range allDividends {
		monthKey := div.Received.Format("2006-01")
		key := monthSymbolKey{month: monthKey, symbol: div.Symbol}
		monthlyBySymbol[key] += div.Amount
		monthsSet[monthKey] = true
		symbolsSet[div.Symbol] = true
	}

	var months []string
	for month := range monthsSet {
		months = append(months, month)
	}
	sort.Strings(months)

	var symbols []string
	for symbol := range symbolsSet {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)

	var dividendsOverTime []MonthlyChartData
	var totalDividendsPaid float64
	for month := range monthsSet {
		var monthTotal float64
		for symbol := range symbolsSet {
			key := monthSymbolKey{month: month, symbol: symbol}
			monthTotal += monthlyBySymbol[key]
		}
		dividendsOverTime = append(dividendsOverTime, MonthlyChartData{
			Month:  month,
			Amount: monthTotal,
		})
		totalDividendsPaid += monthTotal
	}

	// Build upcoming ex-dividend dates
	var upcomingExDivDates []UpcomingDividendDate
	now := time.Now()
	for _, divSymbol := range dividendSymbols {
		if divSymbol.ExDividendDate != nil {
			daysUntil := int(divSymbol.ExDividendDate.Sub(now).Hours() / 24)
			// Only show upcoming dates (within next 60 days)
			if daysUntil >= 0 && daysUntil <= 60 {
				upcomingExDivDates = append(upcomingExDivDates, UpcomingDividendDate{
					Symbol:         divSymbol.Symbol,
					ExDividendDate: *divSymbol.ExDividendDate,
					DaysUntil:      daysUntil,
					Dividend:       divSymbol.Dividend,
					Shares:         divSymbol.Shares,
					ExpectedAmount: divSymbol.Dividend * float64(divSymbol.Shares),
				})
			}
		}
	}

	// Sort upcoming dates by DaysUntil (soonest first)
	sort.Slice(upcomingExDivDates, func(i, j int) bool {
		return upcomingExDivDates[i].DaysUntil < upcomingExDivDates[j].DaysUntil
	})

	// Calculate average yield
	var averageYield float64
	if yieldCount > 0 {
		averageYield = totalYield / float64(yieldCount)
	}

	// Build stacked bar chart data
	stackedData := make(map[string]map[string]float64)
	for symbol := range symbolsSet {
		stackedData[symbol] = make(map[string]float64)
		for _, month := range months {
			key := monthSymbolKey{month: month, symbol: symbol}
			stackedData[symbol][month] = monthlyBySymbol[key]
		}
	}

	data := DividendsPageData{
		PageData: PageData{
			Title:      "Dividends",
			ActivePage: "dividends",
			CurrentDB:  s.getCurrentDatabaseName(),
			AllSymbols: s.getAllSymbolsList(),
		},
		DividendSymbols:   dividendSymbols,
		IncomeBySymbol:    incomeBySymbol,
		DividendsOverTime: dividendsOverTime,
		DividendsStackedByMonth: []DividendStackedMonthData{{
			Months:  months,
			Symbols: symbols,
			Data:    stackedData,
		}},
		UpcomingExDivDates: upcomingExDivDates,
		TotalAnnualIncome:  totalAnnualIncome,
		TotalDividendsPaid: totalDividendsPaid,
		AverageYield:       averageYield,
	}

	s.renderTemplate(w, "dividends.html", data)
}

// longPositionsAPIHandler handles CRUD operations for long positions
func (s *Server) longPositionsAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.createLongPositionHandler(w, r)
	case http.MethodPut:
		s.updateLongPositionHandler(w, r)
	case http.MethodDelete:
		s.deleteLongPositionHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// createLongPositionHandler creates a new long position
func (s *Server) createLongPositionHandler(w http.ResponseWriter, r *http.Request) {
	var req LongPositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Symbol == "" || req.Opened == "" || req.Shares == 0 || req.BuyPrice == 0 {
		http.Error(w, "Symbol, opened date, shares, and buy price are required", http.StatusBadRequest)
		return
	}

	// Parse opened date
	openedDate, err := time.Parse("2006-01-02", req.Opened)
	if err != nil {
		http.Error(w, "Invalid opened date format", http.StatusBadRequest)
		return
	}

	// Create the long position
	position, err := s.longPositionService.Create(req.Symbol, openedDate, req.Shares, req.BuyPrice)
	if err != nil {
		log.Printf("Error creating long position: %v", err)
		http.Error(w, "Failed to create long position", http.StatusInternalServerError)
		return
	}

	// If closed date and/or exit price are provided, update them
	if req.Closed != nil && *req.Closed != "" {
		closedDate, err := time.Parse("2006-01-02", *req.Closed)
		if err != nil {
			http.Error(w, "Invalid closed date format", http.StatusBadRequest)
			return
		}

		var exitPrice float64
		if req.ExitPrice != nil {
			exitPrice = *req.ExitPrice
		}

		err = s.longPositionService.CloseByID(position.ID, closedDate, exitPrice)
		if err != nil {
			log.Printf("Error closing long position: %v", err)
			// Position was created, but closing failed - return success but log error
			log.Printf("Long position created but failed to close: %v", err)
		}

		// Refresh position data to get updated values
		position, _ = s.longPositionService.GetByID(position.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(position)
}

// updateLongPositionHandler updates an existing long position
func (s *Server) updateLongPositionHandler(w http.ResponseWriter, r *http.Request) {
	var req LongPositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// ID is required for updates
	if req.ID == nil {
		http.Error(w, "ID is required for updates", http.StatusBadRequest)
		return
	}

	// Parse dates
	openedDate, err := time.Parse("2006-01-02", req.Opened)
	if err != nil {
		http.Error(w, "Invalid opened date format", http.StatusBadRequest)
		return
	}

	var closedDate *time.Time
	if req.Closed != nil && *req.Closed != "" {
		parsed, err := time.Parse("2006-01-02", *req.Closed)
		if err != nil {
			http.Error(w, "Invalid closed date format", http.StatusBadRequest)
			return
		}
		closedDate = &parsed
	}

	// Update the long position
	position, err := s.longPositionService.UpdateByID(*req.ID, req.Symbol, openedDate, req.Shares, req.BuyPrice, closedDate, req.ExitPrice)
	if err != nil {
		log.Printf("Error updating long position: %v", err)
		http.Error(w, "Failed to update long position", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(position)
}

// deleteLongPositionHandler deletes a long position
func (s *Server) deleteLongPositionHandler(w http.ResponseWriter, r *http.Request) {
	var req LongPositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Try ID-based deletion first
	if req.ID != nil {
		err := s.longPositionService.DeleteByID(*req.ID)
		if err != nil {
			log.Printf("Error deleting long position by ID: %v", err)
			http.Error(w, "Failed to delete long position", http.StatusInternalServerError)
			return
		}
	} else {
		// Fallback to compound key deletion
		openedDate, err := time.Parse("2006-01-02", req.Opened)
		if err != nil {
			http.Error(w, "Invalid opened date format", http.StatusBadRequest)
			return
		}

		err = s.longPositionService.Delete(req.Symbol, openedDate, req.Shares, req.BuyPrice)
		if err != nil {
			log.Printf("Error deleting long position: %v", err)
			http.Error(w, fmt.Sprintf("Failed to delete long position: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}
