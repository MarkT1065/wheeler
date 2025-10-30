package web

import (
	"encoding/json"
	"context"
	"log"
	"net/http"
	"sort"
	"stonks/internal/models"
	"strconv"
	"strings"
	"time"
)


// symbolHandler serves the symbol-specific analysis view
func (s *Server) symbolHandler(w http.ResponseWriter, r *http.Request) {
	// Extract symbol from URL path
	symbol := r.URL.Path[len("/symbol/"):]
	if symbol == "" {
		log.Printf("[SYMBOL] ERROR: Empty symbol in URL path")
		http.NotFound(w, r)
		return
	}

	log.Printf("[SYMBOL] ===== Starting symbol handler for: %s =====", symbol)

	log.Printf("[SYMBOL] Step 1: Getting all symbols from database")
	symbols, err := s.symbolService.GetDistinctSymbols()
	if err != nil {
		log.Printf("[SYMBOL] ERROR: Failed to get symbols: %v", err)
		symbols = []string{}
	} else {
		log.Printf("[SYMBOL] Successfully retrieved %d symbols: %v", len(symbols), symbols)
	}

	// Get symbol data from database
	log.Printf("[SYMBOL] Step 2: Getting symbol data for %s", symbol)
	symbolData, err := s.symbolService.GetBySymbol(symbol)
	currentPrice := "0.00"
	lastUpdate := time.Now().Format("01/02/2006")
	var price float64
	var dividend float64
	var exDividendDate *time.Time
	var peRatio *float64

	var yield float64
	var peRatioValue float64
	var hasPERatio bool

	if err == nil && symbolData != nil {
		log.Printf("[SYMBOL] Found symbol data for %s: Price=%.2f, Dividend=%.2f, PERatio=%v", symbol, symbolData.Price, symbolData.Dividend, symbolData.PERatio)
		currentPrice = strconv.FormatFloat(symbolData.Price, 'f', 2, 64)
		lastUpdate = symbolData.UpdatedAt.Format("01/02/2006")
		price = symbolData.Price
		dividend = symbolData.Dividend
		exDividendDate = symbolData.ExDividendDate
		peRatio = symbolData.PERatio

		// Handle P/E ratio safely
		if symbolData.PERatio != nil {
			peRatioValue = *symbolData.PERatio
			hasPERatio = true
			log.Printf("[SYMBOL] P/E Ratio for %s: %.2f", symbol, peRatioValue)
		} else {
			log.Printf("[SYMBOL] No P/E Ratio for %s", symbol)
		}

		// Calculate yield using Symbol method
		yield = symbolData.CalculateYield()
		log.Printf("[SYMBOL] Calculated yield for %s: %.2f%%", symbol, yield)
	} else {
		log.Printf("[SYMBOL] ERROR: No symbol data found for %s, error: %v", symbol, err)
	}

	// Get dividends for this symbol
	log.Printf("[SYMBOL] Step 3: Getting dividends for %s", symbol)
	dividendsList, err := s.dividendService.GetBySymbol(symbol)
	if err != nil {
		log.Printf("[SYMBOL] ERROR: Failed to get dividends for %s: %v", symbol, err)
		dividendsList = []*models.Dividend{}
	} else {
		log.Printf("[SYMBOL] Retrieved %d dividends for %s", len(dividendsList), symbol)
	}

	// Calculate total dividends
	var dividendsTotal float64
	for _, dividend := range dividendsList {
		dividendsTotal += dividend.Amount
	}
	log.Printf("[SYMBOL] Total dividends for %s: $%.2f", symbol, dividendsTotal)

	// Get options for this symbol
	log.Printf("[SYMBOL] Step 4: Getting options for %s", symbol)
	optionsList, err := s.optionService.GetBySymbol(symbol)
	if err != nil {
		log.Printf("[SYMBOL] ERROR: Failed to get options for %s: %v", symbol, err)
		optionsList = []*models.Option{}
	} else {
		log.Printf("[SYMBOL] Retrieved %d options for %s", len(optionsList), symbol)

		// Sort options: open positions first, ordered by days remaining ascending
		sort.Slice(optionsList, func(i, j int) bool {
			optI, optJ := optionsList[i], optionsList[j]

			// Open positions come before closed positions
			if optI.Closed == nil && optJ.Closed != nil {
				return true // i comes before j
			}
			if optI.Closed != nil && optJ.Closed == nil {
				return false // j comes before i
			}

			// If both are open or both are closed, sort by remaining days (ascending)
			if optI.Closed == nil && optJ.Closed == nil {
				// Both open: sort by days remaining (ascending - closest expiration first)
				return optI.CalculateDaysRemaining() < optJ.CalculateDaysRemaining()
			} else {
				// Both closed: sort by expiration date descending (newest first)
				return optI.Expiration.After(optJ.Expiration)
			}
		})
		log.Printf("[SYMBOL] Sorted %d options for %s (open positions first)", len(optionsList), symbol)
	}

	// Get long positions for this symbol
	log.Printf("[SYMBOL] Step 5: Getting long positions for %s", symbol)
	longPositionsList, err := s.longPositionService.GetBySymbol(symbol)
	if err != nil {
		log.Printf("[SYMBOL] ERROR: Failed to get long positions for %s: %v", symbol, err)
		longPositionsList = []*models.LongPosition{}
	} else {
		log.Printf("[SYMBOL] Retrieved %d long positions for %s", len(longPositionsList), symbol)
	}

	// Calculate Options Gains - sum of total profit from all closed options for this symbol
	log.Printf("[SYMBOL] Step 6: Calculating options gains for %s", symbol)
	var optionsGains float64
	allOptionsCount := 0
	for _, option := range optionsList {
		profit := option.CalculateTotalProfit()
		optionsGains += profit
		allOptionsCount++
		status := "closed"
		if option.IsOpen() {
			status = "open"
		}
		log.Printf("[SYMBOL] Option %d for %s (%s): Profit=%.2f", option.ID, symbol, status, profit)
	}
	log.Printf("[SYMBOL] Total options gains for %s: $%.2f (%d total options)", symbol, optionsGains, allOptionsCount)

	// Calculate Total Commissions - sum of all commissions paid for this symbol
	log.Printf("[SYMBOL] Step 6b: Calculating total commissions for %s", symbol)
	var totalCommissions float64
	for _, option := range optionsList {
		totalCommissions += option.Commission
	}
	log.Printf("[SYMBOL] Total commissions for %s: $%.2f", symbol, totalCommissions)

	// Calculate Cap Gains - sum of profit/loss from all closed long positions for this symbol
	log.Printf("[SYMBOL] Step 7: Calculating cap gains for %s", symbol)
	var capGains float64
	closedPositionsCount := 0
	for _, position := range longPositionsList {
		if position.Closed != nil && position.ExitPrice != nil {
			profit := position.CalculateProfitLoss(*position.ExitPrice)
			capGains += profit
			closedPositionsCount++
			log.Printf("[SYMBOL] Position %d for %s: Profit=%.2f", position.ID, symbol, profit)
		}
	}
	log.Printf("[SYMBOL] Total cap gains for %s: $%.2f (%d closed positions)", symbol, capGains, closedPositionsCount)

	// Calculate total invested in all long positions (for Cash on Cash calculation)
	log.Printf("[SYMBOL] Step 8: Calculating total invested for %s", symbol)
	var totalInvested float64
	for _, position := range longPositionsList {
		invested := position.CalculateTotalInvested()
		totalInvested += invested
		status := "open"
		if position.Closed != nil {
			status = "closed"
		}
		log.Printf("[SYMBOL] Position %d for %s (%s): Invested=%.2f", position.ID, symbol, status, invested)
	}
	log.Printf("[SYMBOL] Total invested for %s: $%.2f (%d positions)", symbol, totalInvested, len(longPositionsList))

	// Calculate Total Profits - sum of Options Gains + Cap Gains + Dividends
	totalProfits := optionsGains + capGains + dividendsTotal
	log.Printf("[SYMBOL] Total profits for %s: $%.2f (Options: $%.2f + Cap: $%.2f + Div: $%.2f)", symbol, totalProfits, optionsGains, capGains, dividendsTotal)

	// Calculate Cash on Cash - Total Profits / Total Invested (all positions)
	var cashOnCash float64
	if totalInvested > 0 {
		cashOnCash = (totalProfits / totalInvested) * 100
	}
	log.Printf("[SYMBOL] Cash on Cash for %s: %.2f%% (Total Profits: $%.2f / Total Invested: $%.2f)", symbol, cashOnCash, totalProfits, totalInvested)

	// Build monthly results for this symbol
	log.Printf("[SYMBOL] Step 9: Building monthly results for %s", symbol)
	monthlyResults := s.buildSymbolMonthlyResults(optionsList)
	log.Printf("[SYMBOL] Built %d monthly results for %s", len(monthlyResults), symbol)

	log.Printf("[SYMBOL] Step 10: Creating template data for %s", symbol)
	data := SymbolData{
		Symbol:            symbol,
		AllSymbols:        symbols,
		CompanyName:       getCompanyName(symbol),
		CurrentPrice:      currentPrice,
		LastUpdate:        lastUpdate,
		Price:             price,
		Dividend:          dividend,
		ExDividendDate:    exDividendDate,
		PERatio:           peRatio,
		PERatioValue:      peRatioValue,
		HasPERatio:        hasPERatio,
		Yield:             yield,
		OptionsGains:      strconv.FormatFloat(optionsGains, 'f', 2, 64),
		CapGains:          strconv.FormatFloat(capGains, 'f', 2, 64),
		Dividends:         strconv.FormatFloat(dividendsTotal, 'f', 2, 64),
		TotalProfits:      strconv.FormatFloat(totalProfits, 'f', 2, 64),
		CashOnCash:        strconv.FormatFloat(cashOnCash, 'f', 2, 64),
		TotalCommissions:  totalCommissions,
		DividendsList:     dividendsList,
		DividendsTotal:    dividendsTotal,
		OptionsList:       optionsList,
		LongPositionsList: longPositionsList,
		MonthlyResults:    monthlyResults,
		CurrentDB:         s.getCurrentDatabaseName(),
		ActivePage:        "symbol",
	}

	log.Printf("[SYMBOL] Step 11: Template data created successfully for %s", symbol)
	log.Printf("[SYMBOL] Data summary: Price=%.2f, OptionsGains=%s, CapGains=%s, Dividends=%s, TotalProfits=%s, CashOnCash=%s%%",
		data.Price, data.OptionsGains, data.CapGains, data.Dividends, data.TotalProfits, data.CashOnCash)
	log.Printf("[SYMBOL] Data counts: %d dividends, %d options, %d long positions, %d monthly results",
		len(data.DividendsList), len(data.OptionsList), len(data.LongPositionsList), len(data.MonthlyResults))

	log.Printf("[SYMBOL] Step 12: Rendering template for %s", symbol)
	s.renderTemplate(w, "symbol.html", data)
	log.Printf("[SYMBOL] ===== Completed symbol handler for: %s =====", symbol)
}

// buildSymbolMonthlyResults creates monthly aggregation for a specific symbol's options based on opened date
func (s *Server) buildSymbolMonthlyResults(options []*models.Option) []SymbolMonthlyResult {
	log.Printf("[MONTHLY] Building monthly results for %d options", len(options))
	monthMap := make(map[string]*SymbolMonthlyResult)
	months := []string{"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}

	// Initialize all months
	for _, month := range months {
		monthMap[month] = &SymbolMonthlyResult{
			Month:           month,
			PutsCount:       0,
			CallsCount:      0,
			PutsTotal:       0.0,
			CallsTotal:      0.0,
			Total:           0.0,
			CommissionTotal: 0.0,
		}
	}
	log.Printf("[MONTHLY] Initialized %d months", len(monthMap))

	// Process all options (both open and closed)
	processedCount := 0
	for _, option := range options {
		processedCount++
		monthName := months[option.Opened.Month()-1]
		monthData := monthMap[monthName]

		profit := option.CalculateTotalProfit()
		status := "open"
		if option.Closed != nil {
			status = "closed"
		}
		log.Printf("[MONTHLY] Processing option %d for %s: %s in %s (%s), profit=%.2f", option.ID, option.Symbol, option.Type, monthName, status, profit)

		if option.Type == "Put" {
			monthData.PutsCount++
			monthData.PutsTotal += profit
		} else {
			monthData.CallsCount++
			monthData.CallsTotal += profit
		}
		monthData.Total = monthData.PutsTotal + monthData.CallsTotal
		monthData.CommissionTotal += option.Commission
		log.Printf("[MONTHLY] Month %s now has: Puts=%d (%.2f), Calls=%d (%.2f), Total=%.2f, Commissions=%.2f",
			monthName, monthData.PutsCount, monthData.PutsTotal, monthData.CallsCount, monthData.CallsTotal, monthData.Total, monthData.CommissionTotal)
	}
	log.Printf("[MONTHLY] Processed %d options out of %d total", processedCount, len(options))

	// Convert map to slice in month order
	var results []SymbolMonthlyResult
	for _, month := range months {
		monthData := *monthMap[month]
		results = append(results, monthData)
		if monthData.PutsCount > 0 || monthData.CallsCount > 0 || monthData.Total != 0.0 {
			log.Printf("[MONTHLY] Adding non-zero month %s: Puts=%d (%.2f), Calls=%d (%.2f), Total=%.2f",
				monthData.Month, monthData.PutsCount, monthData.PutsTotal, monthData.CallsCount, monthData.CallsTotal, monthData.Total)
		}
	}

	log.Printf("[MONTHLY] Built %d monthly results", len(results))
	return results
}

// getCompanyName returns a company name for the symbol (static for now)
func getCompanyName(symbol string) string {
	companies := map[string]string{
		"VZ":   "Verizon Communications Inc",
		"AAPL": "Apple Inc",
		"MSFT": "Microsoft Corporation",
		"KO":   "The Coca-Cola Company",
		"CVX":  "Chevron Corporation",
		"FDD":  "First Trust STOXX European Select Dividend Index Fund",
	}

	if name, exists := companies[symbol]; exists {
		return name
	}
	return symbol + " Inc"
}

// updateSymbolHandler handles PUT requests to update symbol data
func (s *Server) updateSymbolHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract symbol from URL path
	symbol := strings.TrimPrefix(r.URL.Path, "/api/symbols/")
	if symbol == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		return
	}

	var updateReq SymbolUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get existing symbol or create if it doesn't exist
	existingSymbol, err := s.symbolService.GetBySymbol(symbol)
	if err != nil {
		// If symbol doesn't exist, create it
		existingSymbol, err = s.symbolService.Create(symbol)
		if err != nil {
			http.Error(w, "Failed to create symbol", http.StatusInternalServerError)
			return
		}
	}

	// Use existing values if not provided in update
	price := existingSymbol.Price
	dividend := existingSymbol.Dividend
	exDividendDate := existingSymbol.ExDividendDate
	var peRatio *float64 = existingSymbol.PERatio

	if updateReq.Price != nil {
		price = *updateReq.Price
	}
	if updateReq.Dividend != nil {
		dividend = *updateReq.Dividend
	}
	if updateReq.ExDividendDate != nil {
		if *updateReq.ExDividendDate == "" {
			exDividendDate = nil
		} else {
			parsedDate, err := time.Parse("2006-01-02", *updateReq.ExDividendDate)
			if err != nil {
				http.Error(w, "Invalid ex-dividend date format", http.StatusBadRequest)
				return
			}
			exDividendDate = &parsedDate
		}
	}
	if updateReq.PERatio != nil {
		peRatio = updateReq.PERatio
	}

	// Update the symbol
	updatedSymbol, err := s.symbolService.Update(symbol, price, dividend, exDividendDate, peRatio)
	if err != nil {
		http.Error(w, "Failed to update symbol", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedSymbol)
}

// deleteSymbolHandler deletes a symbol and all related data
func (s *Server) deleteSymbolHandler(w http.ResponseWriter, r *http.Request, symbol string) {
	log.Printf("[DELETE_SYMBOL] Starting deletion process for symbol: %s", symbol)

	// Delete all related data first
	log.Printf("[DELETE_SYMBOL] Deleting dividends for symbol: %s", symbol)
	if err := s.dividendService.DeleteBySymbol(symbol); err != nil {
		log.Printf("[DELETE_SYMBOL] ERROR: Failed to delete dividends for %s: %v", symbol, err)
		http.Error(w, "Failed to delete symbol dividends", http.StatusInternalServerError)
		return
	}

	log.Printf("[DELETE_SYMBOL] Deleting long positions for symbol: %s", symbol)
	if err := s.longPositionService.DeleteBySymbol(symbol); err != nil {
		log.Printf("[DELETE_SYMBOL] ERROR: Failed to delete long positions for %s: %v", symbol, err)
		http.Error(w, "Failed to delete symbol long positions", http.StatusInternalServerError)
		return
	}

	log.Printf("[DELETE_SYMBOL] Deleting options for symbol: %s", symbol)
	if err := s.optionService.DeleteBySymbol(symbol); err != nil {
		log.Printf("[DELETE_SYMBOL] ERROR: Failed to delete options for %s: %v", symbol, err)
		http.Error(w, "Failed to delete symbol options", http.StatusInternalServerError)
		return
	}

	// Finally delete the symbol itself
	log.Printf("[DELETE_SYMBOL] Deleting symbol: %s", symbol)
	if err := s.symbolService.Delete(symbol); err != nil {
		log.Printf("[DELETE_SYMBOL] ERROR: Failed to delete symbol %s: %v", symbol, err)
		http.Error(w, "Failed to delete symbol", http.StatusInternalServerError)
		return
	}

	log.Printf("[DELETE_SYMBOL] Successfully deleted symbol and all related data: %s", symbol)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Symbol deleted successfully"})
}

// symbolAPIHandler routes different symbol API requests
func (s *Server) symbolAPIHandler(w http.ResponseWriter, r *http.Request) {
	pathSegments := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/symbols/"), "/")
	if len(pathSegments) == 0 || pathSegments[0] == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		return
	}

	symbol := pathSegments[0]

	// Check if this is a dividends request
	if len(pathSegments) > 1 && pathSegments[1] == "dividends" {
		s.symbolDividendsHandler(w, r, symbol)
		return
	}

	// Check if this is a price update request
	if len(pathSegments) > 1 && pathSegments[1] == "update-price" {
		s.symbolUpdatePriceHandler(w, r, symbol)
		return
	}

	// Check if this is a dividend data fetch request
	if len(pathSegments) > 1 && pathSegments[1] == "fetch-dividends" {
		s.symbolFetchDividendsHandler(w, r, symbol)
		return
	}

	// Handle different HTTP methods for symbol operations
	switch r.Method {
	case http.MethodPut:
		s.updateSymbolHandler(w, r)
	case http.MethodDelete:
		s.deleteSymbolHandler(w, r, symbol)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// symbolDividendsHandler provides API data for symbol-specific dividends
func (s *Server) symbolDividendsHandler(w http.ResponseWriter, r *http.Request, symbol string) {
	dividends, err := s.dividendService.GetBySymbol(symbol)
	if err != nil {
		http.Error(w, "Failed to get dividends", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dividends)
}

// symbolUpdatePriceHandler updates a symbol's price using Polygon.io API
func (s *Server) symbolUpdatePriceHandler(w http.ResponseWriter, r *http.Request, symbol string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("[SYMBOL API] Updating price for symbol: %s", symbol)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Update symbol price using Polygon service
	err := s.polygonService.UpdateSymbolPrice(ctx, symbol)
	
	response := map[string]interface{}{
		"success": err == nil,
		"symbol":  symbol,
	}

	if err != nil {
		response["error"] = err.Error()
		log.Printf("[SYMBOL API] Failed to update price for %s: %v", symbol, err)
		w.WriteHeader(http.StatusBadRequest)
	} else {
		response["message"] = "Price updated successfully"
		log.Printf("[SYMBOL API] Successfully updated price for %s", symbol)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[SYMBOL API] Error encoding response: %v", err)
	}
}

// symbolFetchDividendsHandler fetches dividend data for a symbol from Polygon.io
func (s *Server) symbolFetchDividendsHandler(w http.ResponseWriter, r *http.Request, symbol string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("[SYMBOL API] Fetching dividends for symbol: %s", symbol)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch dividend data using Polygon service
	dividends, err := s.polygonService.FetchDividendHistory(ctx, symbol, 10)
	
	response := map[string]interface{}{
		"success": err == nil,
		"symbol":  symbol,
	}

	if err != nil {
		response["error"] = err.Error()
		log.Printf("[SYMBOL API] Failed to fetch dividends for %s: %v", symbol, err)
		w.WriteHeader(http.StatusBadRequest)
	} else {
		response["message"] = "Dividends fetched successfully"
		response["dividends"] = dividends
		response["count"] = len(dividends)
		log.Printf("[SYMBOL API] Successfully fetched %d dividends for %s", len(dividends), symbol)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[SYMBOL API] Error encoding response: %v", err)
	}
}