package providers

import (
	"context"
	"fmt"
	"stonks/internal/polygon"
	"time"
)

// PolygonProvider implements the Provider interface using Polygon.io
type PolygonProvider struct {
	client *polygon.Client
}

// NewPolygonProvider creates a new Polygon.io provider
func NewPolygonProvider(apiKey string) *PolygonProvider {
	return &PolygonProvider{
		client: polygon.NewClient(apiKey),
	}
}

func newPolygonProviderForTest(client *polygon.Client) *PolygonProvider {
	return &PolygonProvider{client: client}
}

// GetQuote retrieves the current or most recent quote for a symbol
func (p *PolygonProvider) GetQuote(ctx context.Context, symbol string) (*Quote, error) {
	polygonQuote, err := p.client.GetPreviousClose(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("polygon: %w", err)
	}

	return &Quote{
		Symbol: polygonQuote.Results.Symbol,
		// Polygon's /prev aggregate exposes the prior session's close as its
		// only close value. It is therefore both the returned price and the
		// best available previous-close value without another market-data call.
		Price:         polygonQuote.Results.Price,
		PreviousClose: polygonQuote.Results.Price,
		Open:          polygonQuote.Results.Open,
		High:          polygonQuote.Results.High,
		Low:           polygonQuote.Results.Low,
		Volume:        polygonQuote.Results.Volume,
		Timestamp:     time.Unix(0, polygonQuote.Results.Timestamp*int64(time.Millisecond)),
	}, nil
}

// GetTickerDetails retrieves detailed information about a ticker
func (p *PolygonProvider) GetTickerDetails(ctx context.Context, symbol string) (*TickerDetails, error) {
	details, err := p.client.GetTickerDetails(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("polygon: %w", err)
	}

	return &TickerDetails{
		Symbol:      details.Results.Symbol,
		Name:        details.Results.Name,
		Market:      details.Results.Market,
		Type:        details.Results.Type,
		Active:      details.Results.Active,
		Currency:    details.Results.CurrencyName,
		Description: details.Results.Description,
		Homepage:    details.Results.HomepageURL,
		MarketCap:   details.Results.MarketCap,
		Employees:   details.Results.TotalEmployees,
	}, nil
}

// GetDividends retrieves dividend history for a symbol
func (p *PolygonProvider) GetDividends(ctx context.Context, symbol string, limit int) ([]*Dividend, error) {
	dividends, err := p.client.GetDividends(ctx, symbol, limit)
	if err != nil {
		return nil, fmt.Errorf("polygon: %w", err)
	}

	result := make([]*Dividend, 0, len(dividends.Results))
	for _, div := range dividends.Results {
		result = append(result, &Dividend{
			Symbol:          div.Ticker,
			CashAmount:      div.CashAmount,
			DeclarationDate: div.DeclarationDate,
			ExDividendDate:  div.ExDividendDate,
			PayDate:         div.PayDate,
			RecordDate:      div.RecordDate,
			DividendType:    div.DividendType,
			Frequency:       div.Frequency,
		})
	}

	return result, nil
}

// ValidateConnection tests if the provider connection is working
func (p *PolygonProvider) ValidateConnection(ctx context.Context) error {
	return p.client.IsValidAPIKey(ctx)
}

// Name returns the provider's name
func (p *PolygonProvider) Name() string {
	return "Polygon.io"
}
