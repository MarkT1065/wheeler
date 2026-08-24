package providers

import (
	"context"
	"fmt"
	"log"
	"os"
	"stonks/internal/markets"
	"stonks/internal/models"
	"strings"
	"time"
)

// Service provides market data integration for Wheeler using provider abstraction
type Service struct {
	symbolService  *models.SymbolService
	settingService *models.SettingService
}

const (
	providerTypePolygon  = "polygon"
	providerTypeGoogle   = "google_finance"
	providerTypeYFinance = "yfinance"
	// bulkPriceUpdateLimit keeps Polygon's 12-second free-tier pacing safely
	// within the five-minute web request deadline. The remainder is reported
	// to callers so it can be refreshed in a later request.
	bulkPriceUpdateLimit = 20
)

// BulkPriceUpdateLimit is the maximum number of symbols processed by one bulk
// price request. Callers can use it to explain why a later request is needed.
func BulkPriceUpdateLimit() int {
	return bulkPriceUpdateLimit
}

// ProviderProfile describes the configuration and capabilities of a market
// data provider. Provider IDs are stored per database in DATA_PROVIDER_TYPE.
type ProviderProfile struct {
	ID                string        `json:"id"`
	DisplayName       string        `json:"display_name"`
	RequiresAPIKey    bool          `json:"requires_api_key"`
	SupportsDividends bool          `json:"supports_dividends"`
	RateLimitDelay    time.Duration `json:"-"`
}

var providerProfiles = map[string]ProviderProfile{
	providerTypePolygon: {
		ID: providerTypePolygon, DisplayName: "Polygon.io", RequiresAPIKey: true,
		SupportsDividends: true, RateLimitDelay: 12 * time.Second,
	},
	providerTypeGoogle: {
		ID: providerTypeGoogle, DisplayName: "Google Finance",
		SupportsDividends: false, RateLimitDelay: 2 * time.Second,
	},
	providerTypeYFinance: {
		ID: providerTypeYFinance, DisplayName: "Yahoo Finance (yfinance)",
		SupportsDividends: true, RateLimitDelay: 2 * time.Second,
	},
}

// CanonicalProviderType normalizes supported aliases. The empty value keeps
// the historical default of Polygon. Unsupported values return ok=false.
func CanonicalProviderType(raw string) (providerType string, ok bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", providerTypePolygon:
		return providerTypePolygon, true
	case "google", "googlefinance", providerTypeGoogle:
		return providerTypeGoogle, true
	case "yahoo", "yahoo_finance", providerTypeYFinance:
		return providerTypeYFinance, true
	default:
		return "", false
	}
}

func normalizeProviderType(raw string) string {
	if providerType, ok := CanonicalProviderType(raw); ok {
		return providerType
	}
	return strings.ToLower(strings.TrimSpace(raw))
}

// ProviderProfiles returns the supported profiles in stable UI order.
func ProviderProfiles() []ProviderProfile {
	return []ProviderProfile{
		providerProfiles[providerTypePolygon],
		providerProfiles[providerTypeGoogle],
		providerProfiles[providerTypeYFinance],
	}
}

// BulkPriceUpdateResult captures aggregate information from a bulk price update run
type BulkPriceUpdateResult struct {
	Updated int
	Failed  int
	Skipped int
	Errors  []string
}

// DividendFetchRecord captures dividend data returned for a symbol
type DividendFetchRecord struct {
	Symbol    string
	Dividends []*DividendInfo
}

// BulkDividendFetchResult summarizes dividend history fetches across many symbols
type BulkDividendFetchResult struct {
	Processed int
	Skipped   int
	Results   []DividendFetchRecord
	Errors    []string
}

// NewService creates a new provider service with the default provider
func NewService(symbolService *models.SymbolService, settingService *models.SettingService) *Service {
	return &Service{
		symbolService:  symbolService,
		settingService: settingService,
	}
}

// getProvider returns the configured provider with API key
func (s *Service) getProvider() (Provider, error) {
	// Determine which data provider to use, defaulting to Polygon for backwards compatibility
	providerType := normalizeProviderType(s.settingService.GetValue("DATA_PROVIDER_TYPE"))

	switch providerType {
	case providerTypePolygon:
		apiKey := s.settingService.GetValue("POLYGON_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("Polygon API key not configured - please set your API key in Settings")
		}

		// Optionally log masked API key for debugging when DEBUG_PROVIDERS is enabled
		if os.Getenv("DEBUG_PROVIDERS") == "1" {
			var maskedKey string
			if len(apiKey) > 6 {
				maskedKey = apiKey[:3] + "..." + apiKey[len(apiKey)-3:]
			} else {
				maskedKey = "***"
			}
			log.Printf("[PROVIDER] Using %s provider with API key: %s (DEBUG_PROVIDERS=1 - do not enable in production)", strings.ToUpper(providerType), maskedKey)
		}

		return NewPolygonProvider(apiKey), nil
	case providerTypeGoogle:
		return NewGoogleFinanceProvider(""), nil
	case providerTypeYFinance:
		return NewYFinanceProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported data provider type: %s", providerType)
	}
}

func (s *Service) providerType() string {
	return normalizeProviderType(s.settingService.GetValue("DATA_PROVIDER_TYPE"))
}

func (s *Service) resolveProviderSymbol(symbol string) (string, error) {
	instrument, err := markets.Parse(symbol)
	if err != nil {
		return "", err
	}
	return markets.ProviderSymbol(s.providerType(), instrument)
}

// ActiveProfile returns the current provider profile. Unsupported persisted
// values return a zero profile so callers can report configuration errors.
func (s *Service) ActiveProfile() ProviderProfile {
	return providerProfiles[normalizeProviderType(s.settingService.GetValue("DATA_PROVIDER_TYPE"))]
}

// UpdateSymbolPrice updates a single symbol's price from the configured provider
func (s *Service) UpdateSymbolPrice(ctx context.Context, symbol string) error {
	providerSymbol, err := s.resolveProviderSymbol(symbol)
	if err != nil {
		return err
	}
	provider, err := s.getProvider()
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	return s.updateSymbolPriceWithProvider(ctx, provider, symbol, providerSymbol)
}

// updateSymbolPriceWithProvider updates a single symbol's price using a pre-fetched provider
func (s *Service) updateSymbolPriceWithProvider(ctx context.Context, provider Provider, symbol, providerSymbol string) error {
	log.Printf("[PROVIDER] Updating price for symbol: %s using %s", symbol, provider.Name())

	// Get current price from provider
	quote, err := provider.GetQuote(ctx, providerSymbol)
	if err != nil {
		return fmt.Errorf("failed to get quote for %s: %w", symbol, err)
	}

	// Get current symbol data
	currentSymbol, err := s.symbolService.GetBySymbol(symbol)
	if err != nil {
		return fmt.Errorf("failed to get current symbol data: %w", err)
	}

	// Update symbol with new price
	_, err = s.symbolService.Update(
		symbol,
		quote.Price,
		currentSymbol.Dividend,
		currentSymbol.ExDividendDate,
		currentSymbol.PERatio,
	)
	if err != nil {
		return fmt.Errorf("failed to update symbol price: %w", err)
	}

	log.Printf("[PROVIDER] Updated %s price to $%.2f", symbol, quote.Price)
	return nil
}

// UpdateAllSymbolPrices updates prices for all symbols in the database
func (s *Service) UpdateAllSymbolPrices(ctx context.Context) (*BulkPriceUpdateResult, error) {
	symbols, err := s.symbolService.GetPrioritizedSymbols()
	if err != nil {
		return nil, fmt.Errorf("failed to get prioritized symbols: %w", err)
	}

	log.Printf("[PROVIDER] Starting prioritized bulk price update for %d symbols", len(symbols))
	return s.UpdateSymbolPrices(ctx, symbols)
}

// UpdateSymbolPrices updates prices for the provided symbol list using a shared provider instance
func (s *Service) UpdateSymbolPrices(ctx context.Context, symbols []string) (*BulkPriceUpdateResult, error) {
	normalized := s.normalizeSymbols(symbols)
	if len(normalized) == 0 {
		return &BulkPriceUpdateResult{}, nil
	}

	provider, err := s.getProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	log.Printf("[PROVIDER] Starting price update for %d symbols using %s", len(normalized), provider.Name())
	result, runErr := s.runPriceUpdates(ctx, provider, normalized)
	if runErr != nil {
		return result, runErr
	}

	log.Printf("[PROVIDER] Price update complete: %d updated, %d failed", result.Updated, result.Failed)
	return result, nil
}

// FetchSymbolDetails gets detailed information about a symbol from the provider
func (s *Service) FetchSymbolDetails(ctx context.Context, symbol string) (*SymbolInfo, error) {
	providerSymbol, err := s.resolveProviderSymbol(symbol)
	if err != nil {
		return nil, err
	}
	provider, err := s.getProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	details, err := provider.GetTickerDetails(ctx, providerSymbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticker details: %w", err)
	}

	quote, err := provider.GetQuote(ctx, providerSymbol)
	if err != nil {
		log.Printf("[PROVIDER] Warning: failed to get current price for %s: %v", symbol, err)
		// Continue without current price
	}

	info := &SymbolInfo{
		Symbol:      symbol,
		Name:        details.Name,
		Market:      details.Market,
		Type:        details.Type,
		Active:      details.Active,
		Currency:    details.Currency,
		Description: details.Description,
		Homepage:    details.Homepage,
		MarketCap:   details.MarketCap,
		Employees:   details.Employees,
	}

	if quote != nil {
		info.CurrentPrice = quote.Price
		info.PreviousClose = quote.PreviousClose
		info.High = quote.High
		info.Low = quote.Low
		info.Volume = quote.Volume
	}

	return info, nil
}

// FetchDividendHistory gets recent dividend history for a symbol
func (s *Service) FetchDividendHistory(ctx context.Context, symbol string, limit int) ([]*DividendInfo, error) {
	providerSymbol, err := s.resolveProviderSymbol(symbol)
	if err != nil {
		return nil, err
	}
	provider, err := s.getProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	if limit <= 0 {
		limit = 10
	}

	dividends, err := provider.GetDividends(ctx, providerSymbol, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get dividends: %w", err)
	}

	return convertDividendsToInfo(dividends), nil
}

// FetchDividendHistoryForSymbols fetches dividend history for each provided symbol
func (s *Service) FetchDividendHistoryForSymbols(ctx context.Context, symbols []string, limit int) (*BulkDividendFetchResult, error) {
	normalized := s.normalizeSymbols(symbols)
	if len(normalized) == 0 {
		return &BulkDividendFetchResult{Results: []DividendFetchRecord{}}, nil
	}

	if limit <= 0 {
		limit = 10
	}

	provider, err := s.getProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	log.Printf("[PROVIDER] Starting dividend fetch for %d symbols using %s", len(normalized), provider.Name())
	rateLimitDelay := s.GetRateLimitDelay()
	result := &BulkDividendFetchResult{Results: []DividendFetchRecord{}}
	if len(normalized) > bulkPriceUpdateLimit {
		result.Skipped = len(normalized) - bulkPriceUpdateLimit
		normalized = normalized[:bulkPriceUpdateLimit]
		log.Printf("[PROVIDER] Limiting this dividend fetch to %d symbols; %d remaining", len(normalized), result.Skipped)
	}

	providerCalls := 0
	for _, symbol := range normalized {
		result.Processed++
		providerSymbol, resolveErr := s.resolveProviderSymbol(symbol)
		if resolveErr != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", symbol, resolveErr))
			continue
		}
		if providerCalls > 0 {
			if err := s.waitForRateLimit(ctx, rateLimitDelay); err != nil {
				return result, err
			}
		}
		dividends, fetchErr := provider.GetDividends(ctx, providerSymbol, limit)
		providerCalls++
		if fetchErr != nil {
			log.Printf("[PROVIDER] Failed to fetch dividends for %s: %v", symbol, fetchErr)
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", symbol, fetchErr))
		} else {
			result.Results = append(result.Results, DividendFetchRecord{
				Symbol:    symbol,
				Dividends: convertDividendsToInfo(dividends),
			})
		}

	}

	log.Printf("[PROVIDER] Dividend fetch complete: %d symbols processed", result.Processed)
	return result, nil
}

// TestConnection validates the provider connection
func (s *Service) TestConnection(ctx context.Context) error {
	provider, err := s.getProvider()
	if err != nil {
		return err
	}

	return provider.ValidateConnection(ctx)
}

// GetAPIKeyStatus returns information about the current API key configuration
func (s *Service) GetAPIKeyStatus() *APIKeyStatus {
	providerType := normalizeProviderType(s.settingService.GetValue("DATA_PROVIDER_TYPE"))
	profile, supported := providerProfiles[providerType]
	if !supported {
		return &APIKeyStatus{Provider: providerType}
	}

	var apiKey string

	switch providerType {
	case providerTypePolygon:
		apiKey = s.settingService.GetValue("POLYGON_API_KEY")
	case providerTypeGoogle, providerTypeYFinance:
		return &APIKeyStatus{
			Provider:          profile.ID,
			DisplayName:       profile.DisplayName,
			RequiresAPIKey:    false,
			SupportsDividends: profile.SupportsDividends,
			Configured:        false,
			Ready:             true,
			Masked:            "Not required",
		}
	}

	status := &APIKeyStatus{
		Provider:          profile.ID,
		DisplayName:       profile.DisplayName,
		RequiresAPIKey:    profile.RequiresAPIKey,
		SupportsDividends: profile.SupportsDividends,
		Configured:        apiKey != "",
		Ready:             apiKey != "",
		Masked:            "",
	}

	if !status.Configured {
		return status
	}

	// Mask the API key so only a small portion is visible
	if len(apiKey) > 6 {
		status.Masked = apiKey[:3] + "..." + apiKey[len(apiKey)-3:]
	} else {
		status.Masked = strings.Repeat("*", len(apiKey))
	}

	return status
}

// Name returns the name of the configured provider
func (s *Service) Name() string {
	return normalizeProviderType(s.settingService.GetValue("DATA_PROVIDER_TYPE"))
}

// GetRateLimitDelay returns the rate limit delay based on the configured provider
func (s *Service) GetRateLimitDelay() time.Duration {
	if profile := s.ActiveProfile(); profile.ID != "" {
		return profile.RateLimitDelay
	}
	return 10 * time.Second
}

// SymbolInfo represents enriched symbol information from a provider
type SymbolInfo struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Market        string  `json:"market"`
	Type          string  `json:"type"`
	Active        bool    `json:"active"`
	Currency      string  `json:"currency"`
	Description   string  `json:"description"`
	Homepage      string  `json:"homepage"`
	MarketCap     float64 `json:"market_cap"`
	Employees     int     `json:"employees"`
	CurrentPrice  float64 `json:"current_price"`
	PreviousClose float64 `json:"previous_close"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Volume        float64 `json:"volume"`
}

// DividendInfo represents dividend information from a provider
type DividendInfo struct {
	Symbol          string  `json:"symbol"`
	CashAmount      float64 `json:"cash_amount"`
	DeclarationDate string  `json:"declaration_date"`
	ExDividendDate  string  `json:"ex_dividend_date"`
	PayDate         string  `json:"pay_date"`
	RecordDate      string  `json:"record_date"`
	DividendType    string  `json:"dividend_type"`
	Frequency       int     `json:"frequency"`
}

// APIKeyStatus represents the status of the provider API key
type APIKeyStatus struct {
	Provider          string `json:"provider"`
	DisplayName       string `json:"display_name"`
	RequiresAPIKey    bool   `json:"requires_api_key"`
	SupportsDividends bool   `json:"supports_dividends"`
	Configured        bool   `json:"configured"`
	Ready             bool   `json:"ready"`
	Masked            string `json:"masked"`
	Valid             bool   `json:"valid"`
	Error             string `json:"error,omitempty"`
}

func (s *Service) runPriceUpdates(ctx context.Context, provider Provider, symbols []string) (*BulkPriceUpdateResult, error) {
	rateLimitDelay := s.GetRateLimitDelay()
	result := &BulkPriceUpdateResult{}
	if len(symbols) > bulkPriceUpdateLimit {
		result.Skipped = len(symbols) - bulkPriceUpdateLimit
		symbols = symbols[:bulkPriceUpdateLimit]
		log.Printf("[PROVIDER] Limiting this bulk price update to %d symbols; %d remaining", len(symbols), result.Skipped)
	}

	providerCalls := 0
	for _, symbol := range symbols {
		providerSymbol, resolveErr := s.resolveProviderSymbol(symbol)
		if resolveErr != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", symbol, resolveErr))
			continue
		}
		if providerCalls > 0 {
			if err := s.waitForRateLimit(ctx, rateLimitDelay); err != nil {
				return result, err
			}
		}
		if err := s.updateSymbolPriceWithProvider(ctx, provider, symbol, providerSymbol); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", symbol, err))
			log.Printf("[PROVIDER] Failed to update %s: %v", symbol, err)
		} else {
			result.Updated++
		}
		providerCalls++
	}

	return result, nil
}

func (s *Service) waitForRateLimit(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}

func (s *Service) normalizeSymbols(symbols []string) []string {
	seen := make(map[string]struct{})
	var normalized []string

	for _, raw := range symbols {
		symbol := strings.ToUpper(strings.TrimSpace(raw))
		if symbol == "" {
			continue
		}
		if _, exists := seen[symbol]; exists {
			continue
		}
		seen[symbol] = struct{}{}
		normalized = append(normalized, symbol)
	}

	return normalized
}

func convertDividendsToInfo(dividends []*Dividend) []*DividendInfo {
	result := make([]*DividendInfo, 0, len(dividends))
	for _, div := range dividends {
		if div == nil {
			continue
		}
		result = append(result, &DividendInfo{
			Symbol:          div.Symbol,
			CashAmount:      div.CashAmount,
			DeclarationDate: div.DeclarationDate,
			ExDividendDate:  div.ExDividendDate,
			PayDate:         div.PayDate,
			RecordDate:      div.RecordDate,
			DividendType:    div.DividendType,
			Frequency:       div.Frequency,
		})
	}
	return result
}
