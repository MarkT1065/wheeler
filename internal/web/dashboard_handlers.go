package web

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"stonks/internal/models"
)

// dashboardHandler serves the TraderVue-style dashboard
func (s *Server) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	symbols, err := s.symbolService.GetDistinctSymbols()
	if err != nil {
		log.Printf("[DASHBOARD] Error getting symbols: %v", err)
		symbols = []string{}
	}

	// Sort symbols alphabetically for consistent legend colors across charts
	sort.Strings(symbols)

	log.Printf("[DASHBOARD] Found %d symbols for navigation: %v", len(symbols), symbols)

	// Build comprehensive dashboard data
	data, err := s.buildDashboardData(symbols)
	if err != nil {
		log.Printf("Error building dashboard data: %v", err)
		// Fallback to basic data structure
		data = DashboardData{
			Symbols:   symbols,
			CurrentDB: s.getCurrentDatabaseName(),
		}
	}

	s.renderTemplate(w, "dashboard.html", data)
}

// buildDashboardData creates comprehensive dashboard data
func (s *Server) buildDashboardData(symbols []string) (DashboardData, error) {
	// Get all data
	options, _ := s.optionService.GetAll()
	longPositions, _ := s.longPositionService.GetAll()
	dividends, _ := s.dividendService.GetAll()
	treasuries, _ := s.treasuryService.GetAll()

	// Build symbol summaries
	symbolSummaries := s.buildSymbolSummaries(symbols, options, longPositions, dividends)

	// Build chart data
	longByTicker := s.buildLongByTickerChart(longPositions)
	putsByTicker := s.buildPutsByTickerChart(options)
	totalAllocation := s.buildTotalAllocationChart(longPositions, options, treasuries)

	// Calculate totals
	totals := s.calculateDashboardTotals(symbolSummaries, treasuries)

	log.Printf("[DASHBOARD] Building dashboard data with %d symbols: %v", len(symbols), symbols)
	log.Printf("[DASHBOARD] Built %d symbol summaries", len(symbolSummaries))

	return DashboardData{
		Symbols:         symbols,
		SymbolSummaries: symbolSummaries,
		LongByTicker:    longByTicker,
		PutsByTicker:    putsByTicker,
		TotalAllocation: totalAllocation,
		Totals:          totals,
		CurrentDB:       s.getCurrentDatabaseName(),
	}, nil
}

func (s *Server) buildSymbolSummaries(symbols []string, options []*models.Option, longPositions []*models.LongPosition, dividends []*models.Dividend) []SymbolSummary {
	summaryMap := make(map[string]*SymbolSummary)

	// Initialize all symbols with current prices from database
	for _, symbol := range symbols {
		currentPrice := 0.0
		if symbolData, err := s.symbolService.GetBySymbol(symbol); err == nil {
			currentPrice = symbolData.Price
		}

		summaryMap[symbol] = &SymbolSummary{
			Ticker:       symbol,
			CurrentPrice: currentPrice,
		}
	}

	// Process long positions
	for _, pos := range longPositions {
		if summary, exists := summaryMap[pos.Symbol]; exists {
			// Only count current open positions for LongAmount
			if pos.Closed == nil {
				summary.LongAmount += pos.CalculateAmount()
			}
			// Only count realized gains from closed positions
			if pos.ExitPrice != nil && pos.Closed != nil {
				summary.CapGains += pos.CalculateProfitLoss(*pos.ExitPrice)
			}
		}
	}

	// Process options
	for _, opt := range options {
		if summary, exists := summaryMap[opt.Symbol]; exists {
			if opt.Type == "Put" {
				// Count put exposure for all open puts
				if opt.Closed == nil {
					summary.PutExposed += opt.Strike * float64(opt.Contracts) * 100
				}
				// Count premium for all puts (closed and open)
				premium := opt.CalculateTotalProfitWithCurrentPrice(summary.CurrentPrice)
				summary.Puts += premium
			} else {
				// Count premium for all calls (closed and open)
				premium := opt.CalculateTotalProfitWithCurrentPrice(summary.CurrentPrice)
				summary.Calls += premium
			}
		}
	}

	// Process dividends
	for _, div := range dividends {
		if summary, exists := summaryMap[div.Symbol]; exists {
			summary.Dividends += div.Amount
		}
	}

	// Calculate net and cash on cash, filter out symbols with no trade data
	var summaries []SymbolSummary
	for _, summary := range summaryMap {
		summary.Net = summary.CapGains + summary.Puts + summary.Calls + summary.Dividends
		if summary.LongAmount > 0 {
			summary.CashOnCash = (summary.Net / summary.LongAmount) * 100
		}

		// Only include symbols that have some trade activity
		if summary.LongAmount > 0 || summary.PutExposed > 0 || summary.Puts != 0 ||
			summary.Calls != 0 || summary.CapGains != 0 || summary.Dividends > 0 {
			summaries = append(summaries, *summary)
		}
	}

	return summaries
}

func (s *Server) buildLongByTickerChart(longPositions []*models.LongPosition) []ChartData {
	tickerAmounts := make(map[string]float64)
	colors := []string{"#FF6384", "#36A2EB", "#FFCE56", "#4BC0C0", "#9966FF", "#FF9F40"}

	for _, pos := range longPositions {
		if pos.Closed == nil { // Only include open positions
			tickerAmounts[pos.Symbol] += pos.CalculateAmount()
		}
	}

	// Sort tickers alphabetically for consistent legend colors
	var tickers []string
	for ticker := range tickerAmounts {
		tickers = append(tickers, ticker)
	}
	sort.Strings(tickers)

	var chartData []ChartData
	for i, ticker := range tickers {
		chartData = append(chartData, ChartData{
			Label: ticker,
			Value: tickerAmounts[ticker],
			Color: colors[i%len(colors)],
		})
	}

	return chartData
}

func (s *Server) buildPutsByTickerChart(options []*models.Option) []ChartData {
	putExposure := make(map[string]float64)
	colors := []string{"#FF6384", "#36A2EB", "#FFCE56", "#4BC0C0", "#9966FF", "#FF9F40"}

	for _, opt := range options {
		if opt.Type == "Put" && opt.Closed == nil { // Only include open puts
			putExposure[opt.Symbol] += opt.Strike * float64(opt.Contracts) * 100
		}
	}

	// Sort tickers alphabetically for consistent legend colors
	var tickers []string
	for ticker := range putExposure {
		tickers = append(tickers, ticker)
	}
	sort.Strings(tickers)

	var chartData []ChartData
	for i, ticker := range tickers {
		chartData = append(chartData, ChartData{
			Label: ticker,
			Value: putExposure[ticker],
			Color: colors[i%len(colors)],
		})
	}

	return chartData
}

func (s *Server) buildTotalAllocationChart(longPositions []*models.LongPosition, options []*models.Option, treasuries []*models.Treasury) []ChartData {
	var totalLong, totalPuts, totalTreasuries float64

	// Only count open long positions for current allocation
	for _, pos := range longPositions {
		if pos.Closed == nil {
			totalLong += pos.CalculateAmount()
		}
	}

	// Only count open put options for current exposure
	for _, opt := range options {
		if opt.Type == "Put" && opt.Closed == nil {
			totalPuts += opt.Strike * float64(opt.Contracts) * 100
		}
	}

	// Count all treasury holdings (bonds are typically held to maturity)
	for _, treasury := range treasuries {
		totalTreasuries += treasury.Amount
	}

	return []ChartData{
		{Label: "Long Stock", Value: totalLong, Color: "#36A2EB"},
		{Label: "Put Exposure", Value: totalPuts, Color: "#FF6384"},
		{Label: "Treasuries", Value: totalTreasuries, Color: "#FFCE56"},
	}
}

func (s *Server) calculateDashboardTotals(symbolSummaries []SymbolSummary, treasuries []*models.Treasury) DashboardTotals {
	var totalLong, totalPuts, totalPutPremiums, totalCallPremiums, totalCapGains, totalDividends, totalTreasuries float64

	// Sum from symbol summaries
	for _, summary := range symbolSummaries {
		totalLong += summary.LongAmount
		totalPuts += summary.PutExposed
		totalPutPremiums += summary.Puts
		totalCallPremiums += summary.Calls
		totalCapGains += summary.CapGains
		totalDividends += summary.Dividends
	}

	// Sum treasuries
	for _, treasury := range treasuries {
		totalTreasuries += treasury.Amount
	}

	totalNet := totalPutPremiums + totalCallPremiums + totalCapGains + totalDividends
	overallCashOnCash := 0.0
	if totalLong > 0 {
		overallCashOnCash = (totalNet / totalLong) * 100
	}

	return DashboardTotals{
		TotalLong:         totalLong,
		TotalPuts:         totalPuts,
		TotalTreasuries:   totalTreasuries,
		TotalPutPremiums:  totalPutPremiums,
		TotalCallPremiums: totalCallPremiums,
		TotalCapGains:     totalCapGains,
		TotalDividends:    totalDividends,
		TotalNet:          totalNet,
		OverallCashOnCash: overallCashOnCash,
		GrandTotal:        totalLong + totalPuts + totalTreasuries,
	}
}

// premiumDataHandler returns premium data for charts
func (s *Server) premiumDataHandler(w http.ResponseWriter, r *http.Request) {
	options, err := s.optionService.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var putPremium, callPremium float64
	for _, option := range options {
		totalPremium := option.Premium * float64(option.Contracts) * 100

		if option.Type == "Put" {
			putPremium += totalPremium
		} else if option.Type == "Call" {
			callPremium += totalPremium
		}
	}

	data := PremiumData{
		PutPremium:  putPremium,
		CallPremium: callPremium,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// allocationDataHandler returns allocation data for charts
func (s *Server) allocationDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get open treasuries (no exit price)
	treasuries, err := s.treasuryService.GetAll()
	if err != nil {
		log.Printf("[ALLOCATION API] Error getting treasuries: %v", err)
		http.Error(w, "Failed to get treasuries", http.StatusInternalServerError)
		return
	}

	var totalTreasuries float64
	for _, treasury := range treasuries {
		// Only count open positions (no exit price)
		if treasury.ExitPrice == nil {
			totalTreasuries += treasury.Amount
		}
	}

	// Get open long positions (no exit price)
	longPositions, err := s.longPositionService.GetAll()
	if err != nil {
		log.Printf("[ALLOCATION API] Error getting long positions: %v", err)
		http.Error(w, "Failed to get long positions", http.StatusInternalServerError)
		return
	}

	var totalLong float64
	longByTicker := make(map[string]float64)
	for _, pos := range longPositions {
		if pos.Closed == nil { // Only open positions
			amount := pos.CalculateAmount()
			totalLong += amount
			longByTicker[pos.Symbol] += amount
		}
	}

	// Get open put options
	options, err := s.optionService.GetAll()
	if err != nil {
		log.Printf("[ALLOCATION API] Error getting options: %v", err)
		http.Error(w, "Failed to get options", http.StatusInternalServerError)
		return
	}

	var totalPuts float64
	putsByTicker := make(map[string]float64)
	for _, opt := range options {
		if opt.Type == "Put" && opt.Closed == nil { // Only open puts
			exposure := opt.Strike * float64(opt.Contracts) * 100
			totalPuts += exposure
			putsByTicker[opt.Symbol] += exposure
		}
	}

	log.Printf("[ALLOCATION API] Calculated totals - Long: $%.2f, Puts: $%.2f, Treasuries: $%.2f", totalLong, totalPuts, totalTreasuries)

	// Build response data
	var longByTickerChart []ChartData
	var putsByTickerChart []ChartData
	colors := []string{"#FF6384", "#36A2EB", "#FFCE56", "#4BC0C0", "#9966FF", "#FF9F40"}

	// Sort tickers for consistent colors
	longTickers := make([]string, 0, len(longByTicker))
	for ticker := range longByTicker {
		longTickers = append(longTickers, ticker)
	}
	sort.Strings(longTickers)

	putTickers := make([]string, 0, len(putsByTicker))
	for ticker := range putsByTicker {
		putTickers = append(putTickers, ticker)
	}
	sort.Strings(putTickers)

	// Build chart data
	for i, ticker := range longTickers {
		longByTickerChart = append(longByTickerChart, ChartData{
			Label: ticker,
			Value: longByTicker[ticker],
			Color: colors[i%len(colors)],
		})
	}

	for i, ticker := range putTickers {
		putsByTickerChart = append(putsByTickerChart, ChartData{
			Label: ticker,
			Value: putsByTicker[ticker],
			Color: colors[i%len(colors)],
		})
	}

	totalAllocation := []ChartData{
		{Label: "Long Stock", Value: totalLong, Color: "#36A2EB"},
		{Label: "Put Exposure", Value: totalPuts, Color: "#FF6384"},
		{Label: "Treasuries", Value: totalTreasuries, Color: "#FFCE56"},
	}

	response := AllocationData{
		LongByTicker:    longByTickerChart,
		PutsByTicker:    putsByTickerChart,
		TotalAllocation: totalAllocation,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ALLOCATION API] Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}