package providers

import (
	"context"
	"testing"
)

// MockProvider implements Provider interface for testing
type MockProvider struct {
	name          string
	quoteFunc     func(ctx context.Context, symbol string) (*Quote, error)
	detailsFunc   func(ctx context.Context, symbol string) (*TickerDetails, error)
	dividendsFunc func(ctx context.Context, symbol string, limit int) ([]*Dividend, error)
	validateFunc  func(ctx context.Context) error
}

func (m *MockProvider) GetQuote(ctx context.Context, symbol string) (*Quote, error) {
	if m.quoteFunc != nil {
		return m.quoteFunc(ctx, symbol)
	}
	return &Quote{Symbol: symbol, Price: 100.0}, nil
}

func (m *MockProvider) GetTickerDetails(ctx context.Context, symbol string) (*TickerDetails, error) {
	if m.detailsFunc != nil {
		return m.detailsFunc(ctx, symbol)
	}
	return &TickerDetails{Symbol: symbol, Name: "Test Company"}, nil
}

func (m *MockProvider) GetDividends(ctx context.Context, symbol string, limit int) ([]*Dividend, error) {
	if m.dividendsFunc != nil {
		return m.dividendsFunc(ctx, symbol, limit)
	}
	return []*Dividend{{Symbol: symbol, CashAmount: 1.0}}, nil
}

func (m *MockProvider) ValidateConnection(ctx context.Context) error {
	if m.validateFunc != nil {
		return m.validateFunc(ctx)
	}
	return nil
}

func (m *MockProvider) Name() string {
	if m.name != "" {
		return m.name
	}
	return "MockProvider"
}

var _ Provider = (*MockProvider)(nil)

func TestProviderInterface(t *testing.T) {
	ctx := context.Background()
	mock := &MockProvider{name: "TestProvider"}

	// Test Name
	if name := mock.Name(); name != "TestProvider" {
		t.Errorf("Expected name 'TestProvider', got '%s'", name)
	}

	// Test GetQuote
	quote, err := mock.GetQuote(ctx, "AAPL")
	if err != nil {
		t.Fatalf("GetQuote failed: %v", err)
	}
	if quote.Symbol != "AAPL" {
		t.Errorf("Expected symbol 'AAPL', got '%s'", quote.Symbol)
	}
	if quote.Price != 100.0 {
		t.Errorf("Expected price 100.0, got %.2f", quote.Price)
	}

	// Test with custom quote function
	mock.quoteFunc = func(ctx context.Context, symbol string) (*Quote, error) {
		return &Quote{Symbol: symbol, Price: 150.5}, nil
	}
	quote, err = mock.GetQuote(ctx, "TSLA")
	if err != nil {
		t.Fatalf("GetQuote with custom func failed: %v", err)
	}
	if quote.Price != 150.5 {
		t.Errorf("Expected custom price 150.5, got %.2f", quote.Price)
	}

	// Test GetTickerDetails
	details, err := mock.GetTickerDetails(ctx, "AAPL")
	if err != nil {
		t.Fatalf("GetTickerDetails failed: %v", err)
	}
	if details.Symbol != "AAPL" {
		t.Errorf("Expected symbol 'AAPL', got '%s'", details.Symbol)
	}

	// Test GetDividends
	dividends, err := mock.GetDividends(ctx, "AAPL", 10)
	if err != nil {
		t.Fatalf("GetDividends failed: %v", err)
	}
	if len(dividends) != 1 {
		t.Errorf("Expected 1 dividend, got %d", len(dividends))
	}

	// Test ValidateConnection
	if err := mock.ValidateConnection(ctx); err != nil {
		t.Errorf("ValidateConnection failed: %v", err)
	}
}
