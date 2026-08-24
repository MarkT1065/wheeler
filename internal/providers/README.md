# Data Provider System

The Wheeler application uses a provider abstraction layer to support multiple market data sources. This design allows for easy integration of different data providers while maintaining a consistent interface.

## Architecture

### Provider Interface

All market data providers must implement the `Provider` interface defined in `internal/providers/provider.go`:

```go
type Provider interface {
    GetQuote(ctx context.Context, symbol string) (*Quote, error)
    GetTickerDetails(ctx context.Context, symbol string) (*TickerDetails, error)
    GetDividends(ctx context.Context, symbol string, limit int) ([]*Dividend, error)
    ValidateConnection(ctx context.Context) error
    Name() string
}
```

### Current Providers

- **Polygon.io** (`PolygonProvider`) - Default provider for stock market data
- **Google Finance** (`GoogleFinanceProvider`) - Public quote-page provider; no API key required
- **Yahoo Finance via yfinance** (`YFinanceProvider`) - Native Go client; no API key required

Google Finance is accessed through its public quote page because it does not
provide a supported public REST API. It supports quotes and basic ticker
details, but not dividend history. Wheeler stores symbols as `MARKET:TICKER`,
such as `BVMF:PETR4`, and resolves them to `PETR4:BVMF` for Google Finance. The
legacy `GOOGLE_FINANCE_EXCHANGE` setting is not used for new symbols.

The yfinance provider receives the Yahoo Finance symbol resolved from Wheeler's
canonical key, such as `AAPL` or `PETR4.SA`, and calls Yahoo Finance's chart
endpoint directly from Go. It does not require Python or an API key. Set the `DATA_PROVIDER_TYPE` setting to
`yfinance`. The `YFINANCE_BASE_URL` environment variable can override the Yahoo
endpoint for a proxy or tests.

## Adding a New Provider

To add a new market data provider:

### 1. Create Provider Implementation

Create a new file in `internal/providers/` (e.g., `alphavantage.go`):

```go
package providers

import (
    "context"
    "net/http"
    "time"
)

type AlphaVantageProvider struct {
    apiKey string
    client *http.Client
}

func NewAlphaVantageProvider(apiKey string) *AlphaVantageProvider {
    return &AlphaVantageProvider{
        apiKey: apiKey,
        client: &http.Client{Timeout: 30 * time.Second},
    }
}

// Implement all Provider interface methods
func (p *AlphaVantageProvider) GetQuote(ctx context.Context, symbol string) (*Quote, error) {
    // Implementation...
}

func (p *AlphaVantageProvider) GetTickerDetails(ctx context.Context, symbol string) (*TickerDetails, error) {
    // Implementation...
}

func (p *AlphaVantageProvider) GetDividends(ctx context.Context, symbol string, limit int) ([]*Dividend, error) {
    // Implementation...
}

func (p *AlphaVantageProvider) ValidateConnection(ctx context.Context) error {
    // Implementation...
}

func (p *AlphaVantageProvider) Name() string {
    return "Alpha Vantage"
}
```

### 2. Update Service Configuration

Modify `internal/providers/service.go` to support provider selection:

```go
func (s *Service) getProvider() (Provider, error) {
    providerType := s.settingService.GetValue("DATA_PROVIDER_TYPE")
    
    switch providerType {
    case "alphavantage":
        apiKey := s.settingService.GetValue("ALPHAVANTAGE_API_KEY")
        return NewAlphaVantageProvider(apiKey), nil
    case "polygon":
        fallthrough
    default:
        apiKey := s.settingService.GetValue("POLYGON_API_KEY")
        return NewPolygonProvider(apiKey), nil
    }
}
```

### 3. Add Settings Management

Add settings for the new provider:
- API key storage
- Provider selection UI
- Configuration validation

### 4. Write Tests

Provider unit tests must inject an HTTP client or use an `httptest.Server`.
They must not call a provider's production endpoint, require credentials, or
consume an API quota. See `request_trace_test.go` for offline quote, details,
dividend, and validation examples. Keep any intentional live checks separate
and explicitly opt-in.

### 5. Update Documentation

- Update this README with provider-specific information
- Add any rate limiting considerations
- Document API key requirements

## Provider Standards

All providers should:
1. **Handle errors gracefully** - Return descriptive errors for API failures
2. **Respect rate limits** - Implement appropriate throttling
3. **Cache when appropriate** - Avoid redundant API calls
4. **Log operations** - Use structured logging for debugging
5. **Validate inputs** - Check symbols and parameters before API calls
6. **Context awareness** - Support cancellation via context

## Rate Limiting

Rate limiting is enforced centrally in the service layer via the shared `waitForRateLimit` / `runPriceUpdates` path, which throttles both price and dividend fetches.

- Polygon.io Free Tier: 5 requests/minute (enforced via the shared rate-limiting pipeline)
- When adding a new provider, document its limits and either reuse the shared rate limiter or add provider-specific safeguards where necessary

## Error Handling

Providers should wrap errors with context:
```go
return nil, fmt.Errorf("polygon: failed to fetch quote: %w", err)
```

This helps identify which provider caused an error in logs.

## Testing

Run provider tests:
```bash
go test ./internal/providers/
```

For integration tests with live APIs, set environment variables:
```bash
export POLYGON_API_KEY="your-key"
go test ./internal/providers/ -v
```
