package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"stonks/internal/models"
)

// monthlyHandler serves the monthly performance view
func (s *Server) monthlyHandler(w http.ResponseWriter, r *http.Request) {
	symbols, err := s.symbolService.GetDistinctSymbols()
	if err != nil {
		symbols = []string{}
	}

	// Get all options for calculations
	options, err := s.optionService.GetAll()
	if err != nil {
		options = []*models.Option{}
	}

	// Create options index for advanced filtering
	optionsIndex, err := s.optionService.Index()
	if err != nil {
		optionsIndex = make(map[string]interface{})
	}

	// Get all dividends for calculations
	dividends, err := s.dividendService.GetAll()
	if err != nil {
		dividends = []*models.Dividend{}
	}

	// Get all long positions for capital gains calculations
	longPositions, err := s.longPositionService.GetAll()
	if err != nil {
		longPositions = []*models.LongPosition{}
	}

	// Build monthly data
	data := s.buildMonthlyData(symbols, options, dividends, longPositions, optionsIndex)

	s.renderTemplate(w, "monthly.html", data)
}

// buildMonthlyData creates comprehensive monthly financial data based on transaction dates
func (s *Server) buildMonthlyData(symbols []string, options []*models.Option, dividends []*models.Dividend, longPositions []*models.LongPosition, optionsIndex map[string]interface{}) MonthlyData {
	// Initialize data structures
	putsByMonth := make(map[int]float64)          // month -> total
	callsByMonth := make(map[int]float64)         // month -> total
	putsByTicker := make(map[string]float64)      // ticker -> total
	callsByTicker := make(map[string]float64)     // ticker -> total
	capGainsByMonth := make(map[int]float64)      // month -> total
	dividendsByMonth := make(map[int]float64)     // month -> total
	capGainsByTicker := make(map[string]float64)  // ticker -> total
	dividendsByTicker := make(map[string]float64) // ticker -> total

	// Ticker -> YearMonth (yyyy-mm) -> Amount for table
	tickerMonthData := make(map[string]map[string]float64)
	
	// Track all unique year-months
	yearMonthSet := make(map[string]bool)
	
	// Symbol -> Month -> Premium for stacked bar chart
	symbolMonthPremiums := make(map[string][12]float64)
	

	// Process all options (both open and closed)
	for _, option := range options {
		// Calculate profit/loss for all options (premium realized at open)
		totalPremium := option.CalculateTotalProfit()

		// Get the month from the opened date (when premium was realized)
		month := int(option.Opened.Month()) - 1 // 0-11 for array indexing
		
		// Get yyyy-mm for table
		yearMonth := fmt.Sprintf("%04d-%02d", option.Opened.Year(), option.Opened.Month())
		yearMonthSet[yearMonth] = true

		// Aggregate by month and type
		if option.Type == "Put" {
			putsByMonth[month] += totalPremium
			putsByTicker[option.Symbol] += totalPremium
		} else if option.Type == "Call" {
			callsByMonth[month] += totalPremium
			callsByTicker[option.Symbol] += totalPremium
		}

		// Aggregate for table data (both puts and calls combined)
		if data, exists := tickerMonthData[option.Symbol]; exists {
			data[yearMonth] += totalPremium
		} else {
			newData := make(map[string]float64)
			newData[yearMonth] = totalPremium
			tickerMonthData[option.Symbol] = newData
		}

		// Aggregate for stacked chart data (premium only)
		if data, exists := symbolMonthPremiums[option.Symbol]; exists {
			data[month] += totalPremium
			symbolMonthPremiums[option.Symbol] = data
		} else {
			var newData [12]float64
			newData[month] = totalPremium
			symbolMonthPremiums[option.Symbol] = newData
		}
	}

	// Process all dividends (based on received date)
	for _, dividend := range dividends {
		amount := dividend.Amount

		// Get the month from the received date
		month := int(dividend.Received.Month()) - 1 // 0-11 for array indexing
		
		// Get yyyy-mm for table
		yearMonth := fmt.Sprintf("%04d-%02d", dividend.Received.Year(), dividend.Received.Month())
		yearMonthSet[yearMonth] = true

		// Aggregate by month and ticker
		dividendsByMonth[month] += amount
		dividendsByTicker[dividend.Symbol] += amount

		// Aggregate for table data
		if data, exists := tickerMonthData[dividend.Symbol]; exists {
			data[yearMonth] += amount
		} else {
			newData := make(map[string]float64)
			newData[yearMonth] = amount
			tickerMonthData[dividend.Symbol] = newData
		}
	}

	// Process capital gains only from closed long positions (realized gains only)
	for _, position := range longPositions {
		if position.Closed != nil {
			// Calculate profit/loss for closed position
			profit := (position.GetExitPriceValue() - position.BuyPrice) * float64(position.Shares)

			// Get the month from the closed date
			month := int(position.Closed.Month()) - 1 // 0-11 for array indexing
			
			// Get yyyy-mm for table
			yearMonth := fmt.Sprintf("%04d-%02d", position.Closed.Year(), position.Closed.Month())
			yearMonthSet[yearMonth] = true

			// Aggregate by month and ticker
			capGainsByMonth[month] += profit
			capGainsByTicker[position.Symbol] += profit

			// Aggregate for table data
			if data, exists := tickerMonthData[position.Symbol]; exists {
				data[yearMonth] += profit
			} else {
				newData := make(map[string]float64)
				newData[yearMonth] = profit
				tickerMonthData[position.Symbol] = newData
			}
		}
	}

	// Build chart data for Puts by month
	putsMonthChart := make([]MonthlyChartData, 12)
	monthNames := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	for i := 0; i < 12; i++ {
		putsMonthChart[i] = MonthlyChartData{
			Month:  monthNames[i],
			Amount: putsByMonth[i],
		}
	}

	// Build chart data for Calls by month
	callsMonthChart := make([]MonthlyChartData, 12)
	for i := 0; i < 12; i++ {
		callsMonthChart[i] = MonthlyChartData{
			Month:  monthNames[i],
			Amount: callsByMonth[i],
		}
	}

	// Build ticker chart data for Puts
	putsTickerChart := []TickerChartData{}
	for ticker, amount := range putsByTicker {
		if amount != 0 {
			putsTickerChart = append(putsTickerChart, TickerChartData{
				Ticker: ticker,
				Amount: amount,
			})
		}
	}

	// Build ticker chart data for Calls
	callsTickerChart := []TickerChartData{}
	for ticker, amount := range callsByTicker {
		if amount != 0 {
			callsTickerChart = append(callsTickerChart, TickerChartData{
				Ticker: ticker,
				Amount: amount,
			})
		}
	}

	// Build chart data for Capital Gains by month
	capGainsMonthChart := make([]MonthlyChartData, 12)
	for i := 0; i < 12; i++ {
		capGainsMonthChart[i] = MonthlyChartData{
			Month:  monthNames[i],
			Amount: capGainsByMonth[i],
		}
	}

	// Build ticker chart data for Capital Gains
	capGainsTickerChart := []TickerChartData{}
	for ticker, amount := range capGainsByTicker {
		if amount != 0 {
			capGainsTickerChart = append(capGainsTickerChart, TickerChartData{
				Ticker: ticker,
				Amount: amount,
			})
		}
	}

	// Build chart data for Dividends by month
	dividendsMonthChart := make([]MonthlyChartData, 12)
	for i := 0; i < 12; i++ {
		dividendsMonthChart[i] = MonthlyChartData{
			Month:  monthNames[i],
			Amount: dividendsByMonth[i],
		}
	}

	// Build ticker chart data for Dividends
	dividendsTickerChart := []TickerChartData{}
	for ticker, amount := range dividendsByTicker {
		if amount != 0 {
			dividendsTickerChart = append(dividendsTickerChart, TickerChartData{
				Ticker: ticker,
				Amount: amount,
			})
		}
	}

	// Convert year-month set to sorted slice
	yearMonths := make([]string, 0, len(yearMonthSet))
	for ym := range yearMonthSet {
		yearMonths = append(yearMonths, ym)
	}
	sort.Strings(yearMonths) // Sort ascending
	
	// Create formatted labels ("2025 Jan", etc.)
	monthLabels := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	formattedLabels := make([]string, len(yearMonths))
	for i, ym := range yearMonths {
		// Parse yyyy-mm
		var year, month int
		fmt.Sscanf(ym, "%04d-%02d", &year, &month)
		formattedLabels[i] = fmt.Sprintf("%d %s", year, monthLabels[month-1])
	}
	
	// Build table data
	tableData := []MonthlyTableRow{}
	for ticker, monthValues := range tickerMonthData {
		// Calculate total for this ticker
		total := 0.0
		for _, amount := range monthValues {
			total += amount
		}

		if total != 0 {
			tableData = append(tableData, MonthlyTableRow{
				Ticker:      ticker,
				Total:       total,
				MonthValues: monthValues,
			})
		}
	}

	// Calculate totals by year-month for the table footer
	totalsByYearMonth := make(map[string]float64)
	grandTotal := 0.0
	
	// Aggregate totals from all data sources
	for ticker, monthValues := range tickerMonthData {
		_ = ticker // Unused, just iterating
		for ym, amount := range monthValues {
			totalsByYearMonth[ym] += amount
			grandTotal += amount
		}
	}
	
	// Also build traditional monthly totals for charts (Jan-Dec)
	totalsByMonth := []MonthlyTotal{}
	for i := 0; i < 12; i++ {
		total := putsByMonth[i] + callsByMonth[i] + capGainsByMonth[i] + dividendsByMonth[i]
		totalsByMonth = append(totalsByMonth, MonthlyTotal{
			Month:  monthNames[i],
			Amount: total,
		})
	}

	// Build monthly premiums by symbol data for stacked chart
	monthlyPremiumsBySymbol := make([]MonthlyPremiumsBySymbol, 12)
	
	for i := 0; i < 12; i++ {
		symbolData := []SymbolPremiumData{}
		
		// Get all symbols that have premiums in this month
		for symbol, monthData := range symbolMonthPremiums {
			if monthData[i] > 0 {
				symbolData = append(symbolData, SymbolPremiumData{
					Symbol: symbol,
					Amount: monthData[i],
				})
			}
		}
		
		monthlyPremiumsBySymbol[i] = MonthlyPremiumsBySymbol{
			Month:   monthNames[i],
			Symbols: symbolData,
		}
	}
	
	// Convert options index to JSON for template
	indexJSON, err := json.Marshal(optionsIndex)
	if err != nil {
		log.Printf("[MONTHLY PAGE] ERROR: Failed to marshal options index to JSON: %v", err)
		indexJSON = []byte("{}")
	}

	return MonthlyData{
		Symbols:    symbols,
		AllSymbols: symbols, // For navigation compatibility
		PutsData: MonthlyOptionData{
			ByMonth:  putsMonthChart,
			ByTicker: putsTickerChart,
		},
		CallsData: MonthlyOptionData{
			ByMonth:  callsMonthChart,
			ByTicker: callsTickerChart,
		},
		CapGainsData: MonthlyFinancialData{
			ByMonth:  capGainsMonthChart,
			ByTicker: capGainsTickerChart,
		},
		DividendsData: MonthlyFinancialData{
			ByMonth:  dividendsMonthChart,
			ByTicker: dividendsTickerChart,
		},
		TableData:               tableData,
		TableYearMonths:         yearMonths,
		TableMonthLabels:        formattedLabels,
		TableTotalsByMonth:      totalsByYearMonth,
		TotalsByMonth:           totalsByMonth,
		MonthlyPremiumsBySymbol: monthlyPremiumsBySymbol,
		OptionsIndex:            optionsIndex,
		OptionsIndexJSON:        template.JS(string(indexJSON)),
		GrandTotal:              grandTotal,
		CurrentDB:               s.getCurrentDatabaseName(),
		ActivePage:              "monthly",
	}
}