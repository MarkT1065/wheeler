package providers

import (
	"context"
	"time"
)

// Provider defines the interface for market data providers
type Provider interface {
	// GetQuote retrieves the current or most recent quote for a symbol
	GetQuote(ctx context.Context, symbol string) (*Quote, error)

	// GetTickerDetails retrieves detailed information about a ticker
	GetTickerDetails(ctx context.Context, symbol string) (*TickerDetails, error)

	// GetDividends retrieves dividend history for a symbol
	GetDividends(ctx context.Context, symbol string, limit int) ([]*Dividend, error)

	// ValidateConnection tests if the provider connection is working
	ValidateConnection(ctx context.Context) error
	// Name returns the provider's name
	Name() string
}

// Quote represents a stock quote from any provider
type Quote struct {
	Symbol        string
	Price         float64
	Open          float64
	High          float64
	Low           float64
	Volume        float64
	PreviousClose float64
	Timestamp     time.Time
}

// TickerDetails represents detailed ticker information from any provider
type TickerDetails struct {
	Symbol      string
	Name        string
	Market      string
	Type        string
	Active      bool
	Currency    string
	Description string
	Homepage    string
	MarketCap   float64
	Employees   int
}

// Dividend represents dividend information from any provider
type Dividend struct {
	Symbol          string
	CashAmount      float64
	DeclarationDate string
	ExDividendDate  string
	PayDate         string
	RecordDate      string
	DividendType    string
	Frequency       int
}
