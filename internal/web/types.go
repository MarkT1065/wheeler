package web

import (
	"stonks/internal/models"
	"time"
)

// Data transfer objects and request/response types for web handlers

type PremiumData struct {
	PutPremium  float64 `json:"putPremium"`
	CallPremium float64 `json:"callPremium"`
}

type SymbolUpdateRequest struct {
	Price          *float64 `json:"price,omitempty"`
	Dividend       *float64 `json:"dividend,omitempty"`
	ExDividendDate *string  `json:"ex_dividend_date,omitempty"`
	PERatio        *float64 `json:"pe_ratio,omitempty"`
}

type TreasuryUpdateRequest struct {
	Purchased    string   `json:"purchased"`
	Maturity     string   `json:"maturity"`
	Amount       float64  `json:"amount"`
	Yield        float64  `json:"yield"`
	BuyPrice     float64  `json:"buyPrice"`
	CurrentValue *float64 `json:"currentValue,omitempty"`
	ExitPrice    *float64 `json:"exitPrice,omitempty"`
}

type ImportResponse struct {
	Success       bool   `json:"success"`
	ImportedCount int    `json:"imported_count"`
	SkippedCount  int    `json:"skipped_count"`
	Error         string `json:"error,omitempty"`
	Details       string `json:"details,omitempty"`
}

type CSVOptionRecord struct {
	Symbol     string
	Opened     string
	Closed     string
	Type       string
	Strike     string
	Expiration string
	Premium    string
	Contracts  string
	ExitPrice  string
	Commission string
}

type CSVStockRecord struct {
	Symbol     string
	Purchased  string
	ClosedDate string
	Shares     string
	BuyPrice   string
	ExitPrice  string
}

type CSVDividendRecord struct {
	Symbol       string
	DateReceived string
	Amount       string
}

// DashboardData holds data for the dashboard template
type DashboardData struct {
	Symbols         []string        `json:"symbols"`
	SymbolSummaries []SymbolSummary `json:"symbolSummaries"`
	LongByTicker    []ChartData     `json:"longByTicker"`
	PutsByTicker    []ChartData     `json:"putsByTicker"`
	TotalAllocation []ChartData     `json:"totalAllocation"`
	Totals          DashboardTotals `json:"totals"`
	CurrentDB       string          `json:"currentDB"`
}

type SymbolSummary struct {
	Ticker       string  `json:"ticker"`
	CurrentPrice float64 `json:"currentPrice"`
	LongAmount   float64 `json:"longAmount"`
	PutExposed   float64 `json:"putExposed"`
	Puts         float64 `json:"puts"`
	Calls        float64 `json:"calls"`
	CapGains     float64 `json:"capGains"`
	Dividends    float64 `json:"dividends"`
	Net          float64 `json:"net"`
	CashOnCash   float64 `json:"cashOnCash"`
}

type ChartData struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
	Color string  `json:"color"`
}

type DashboardTotals struct {
	TotalLong         float64 `json:"totalLong"`
	TotalPuts         float64 `json:"totalPuts"`
	TotalTreasuries   float64 `json:"totalTreasuries"`
	TotalPutPremiums  float64 `json:"totalPutPremiums"`
	TotalCallPremiums float64 `json:"totalCallPremiums"`
	TotalCapGains     float64 `json:"totalCapGains"`
	TotalDividends    float64 `json:"totalDividends"`
	TotalNet          float64 `json:"totalNet"`
	OverallCashOnCash float64 `json:"overallCashOnCash"`
	PutROI            float64 `json:"putROI"`
	LongROI           float64 `json:"longROI"`
	GrandTotal        float64 `json:"grandTotal"`
}

// MonthlyData holds data for the monthly template
type MonthlyData struct {
	Symbols                  []string                      `json:"symbols"`
	PutsData                 MonthlyOptionData             `json:"putsData"`
	CallsData                MonthlyOptionData             `json:"callsData"`
	CapGainsData             MonthlyFinancialData          `json:"capGainsData"`
	DividendsData            MonthlyFinancialData          `json:"dividendsData"`
	TableData                []MonthlyTableRow             `json:"tableData"`
	TotalsByMonth            []MonthlyTotal                `json:"totalsByMonth"`
	MonthlyPremiumsBySymbol  []MonthlyPremiumsBySymbol     `json:"monthlyPremiumsBySymbol"`
	GrandTotal               float64                       `json:"grandTotal"`
	CurrentDB                string                        `json:"currentDB"`
}

type MonthlyOptionData struct {
	ByMonth  []MonthlyChartData `json:"byMonth"`
	ByTicker []TickerChartData  `json:"byTicker"`
}

type MonthlyFinancialData struct {
	ByMonth  []MonthlyChartData `json:"byMonth"`
	ByTicker []TickerChartData  `json:"byTicker"`
}

type MonthlyChartData struct {
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
}

type TickerChartData struct {
	Ticker string  `json:"ticker"`
	Amount float64 `json:"amount"`
}

type MonthlyTableRow struct {
	Ticker string      `json:"ticker"`
	Total  float64     `json:"total"`
	Months [12]float64 `json:"months"` // Jan-Dec
}

type MonthlyTotal struct {
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
}

// MonthlyPremiumsBySymbol holds data for stacked bar chart showing monthly premiums by symbol
type MonthlyPremiumsBySymbol struct {
	Month   string             `json:"month"`
	Symbols []SymbolPremiumData `json:"symbols"`
}

type SymbolPremiumData struct {
	Symbol string  `json:"symbol"`
	Amount float64 `json:"amount"`
}

// TreasuriesData holds data for the treasuries template
type TreasuriesData struct {
	Symbols    []string           `json:"symbols"`
	Treasuries []*models.Treasury `json:"treasuries"`
	Summary    TreasuriesSummary  `json:"summary"`
	CurrentDB  string             `json:"currentDB"`
}

type TreasuriesSummary struct {
	TotalAmount     float64 `json:"totalAmount"`
	TotalBuyPrice   float64 `json:"totalBuyPrice"`
	TotalProfitLoss float64 `json:"totalProfitLoss"`
	TotalInterest   float64 `json:"totalInterest"`
	AverageReturn   float64 `json:"averageReturn"`
	ActivePositions int     `json:"activePositions"`
}

type OptionsData struct {
	Symbols        []string                   `json:"symbols"`
	OptionsSummary []*models.OptionSummary    `json:"options_summary"`
	OpenPositions  []*models.OpenPositionData `json:"open_positions"`
	SummaryTotals  *models.OptionSummary      `json:"summary_totals"`
	CurrentDB      string                     `json:"currentDB"`
}

// SymbolMonthlyResult represents monthly results for a specific symbol
type SymbolMonthlyResult struct {
	Month      string  `json:"month"`
	PutsCount  int     `json:"putsCount"`
	CallsCount int     `json:"callsCount"`
	PutsTotal  float64 `json:"putsTotal"`
	CallsTotal float64 `json:"callsTotal"`
	Total      float64 `json:"total"`
}

// SymbolData holds data for the symbol-specific template
type SymbolData struct {
	Symbol            string                 `json:"symbol"`
	AllSymbols        []string               `json:"allSymbols"`
	CompanyName       string                 `json:"companyName"`
	CurrentPrice      string                 `json:"currentPrice"`
	LastUpdate        string                 `json:"lastUpdate"`
	Price             float64                `json:"price"`
	Dividend          float64                `json:"dividend"`
	ExDividendDate    *time.Time             `json:"exDividendDate"`
	PERatio           *float64               `json:"peRatio"`
	PERatioValue      float64                `json:"peRatioValue"`
	HasPERatio        bool                   `json:"hasPERatio"`
	Yield             float64                `json:"yield"`
	OptionsGains      string                 `json:"optionsGains"`
	CapGains          string                 `json:"capGains"`
	Dividends         string                 `json:"dividends"`
	TotalProfits      string                 `json:"totalProfits"`
	CashOnCash        string                 `json:"cashOnCash"`
	DividendsList     []*models.Dividend     `json:"dividendsList"`
	DividendsTotal    float64                `json:"dividendsTotal"`
	OptionsList       []*models.Option       `json:"optionsList"`
	LongPositionsList []*models.LongPosition `json:"longPositionsList"`
	MonthlyResults    []SymbolMonthlyResult  `json:"monthlyResults"`
	CurrentDB         string                 `json:"currentDB"`
}

type OptionRequest struct {
	ID         *int     `json:"id,omitempty"`
	Symbol     string   `json:"symbol"`
	Type       string   `json:"type"`
	Strike     float64  `json:"strike"`
	Expiration string   `json:"expiration"`
	Premium    float64  `json:"premium"`
	Contracts  int      `json:"contracts"`
	Opened     string   `json:"opened"`
	Closed     *string  `json:"closed,omitempty"`
	ExitPrice  *float64 `json:"exit_price,omitempty"`
	Commission float64  `json:"commission,omitempty"`
}

type DividendRequest struct {
	ID           *int    `json:"id,omitempty"`
	Symbol       string  `json:"symbol"`
	Amount       float64 `json:"amount"`
	DateReceived string  `json:"date_received"`
	Received     string  `json:"received"`
}

type LongPositionRequest struct {
	ID        *int     `json:"id,omitempty"`        // For updates
	Symbol    string   `json:"symbol"`
	Shares    int      `json:"shares"`
	BuyPrice  float64  `json:"buy_price"`
	Purchased string   `json:"purchased"`
	Opened    string   `json:"opened"`
	Closed    *string  `json:"closed,omitempty"`
	ExitPrice *float64 `json:"exit_price,omitempty"`
}

type AllocationData struct {
	LongByTicker      []ChartData `json:"longByTicker"`
	PutsByTicker      []ChartData `json:"putsByTicker"`
	TotalAllocation   []ChartData `json:"totalAllocation"`
	PutROI            float64     `json:"putROI"`
	LongROI           float64     `json:"longROI"`
	TotalPutPremiums  float64     `json:"totalPutPremiums"`
	TotalCallPremiums float64     `json:"totalCallPremiums"`
}