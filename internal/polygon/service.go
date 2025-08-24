package polygon

import (
	"context"
	"fmt"
	"log"
	"stonks/internal/models"
	"strings"
	"time"
)

// Service provides Polygon.io integration for Wheeler
type Service struct {
	client         *Client
	symbolService  *models.SymbolService
	settingService *models.SettingService
}

// NewService creates a new Polygon service
func NewService(symbolService *models.SymbolService, settingService *models.SettingService) *Service {
	return &Service{
		symbolService:  symbolService,
		settingService: settingService,
	}
}

// getClient returns a Polygon client with the current API key
func (s *Service) getClient() (*Client, error) {
	apiKey := s.settingService.GetValue("POLYGON_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("Polygon API key not configured - please set your API key in Settings")
	}

	// Log masked API key for debugging (show first 3 and last 3 characters)
	var maskedKey string
	if len(apiKey) > 6 {
		maskedKey = apiKey[:3] + "..." + apiKey[len(apiKey)-3:]
	} else {
		maskedKey = "***"
	}
	log.Printf("[POLYGON] Using API key: %s", maskedKey)

	return NewClient(apiKey), nil
}

// UpdateSymbolPrice updates a single symbol's price from Polygon.io
func (s *Service) UpdateSymbolPrice(ctx context.Context, symbol string) error {
	client, err := s.getClient()
	if err != nil {
		return fmt.Errorf("failed to get Polygon client: %w", err)
	}

	log.Printf("[POLYGON] Updating price for symbol: %s", symbol)

	// Get current price from Polygon
	quote, err := client.GetPreviousClose(ctx, symbol)
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
		quote.Results.Price,
		currentSymbol.Dividend,
		currentSymbol.ExDividendDate,
		currentSymbol.PERatio,
	)
	if err != nil {
		return fmt.Errorf("failed to update symbol price: %w", err)
	}

	log.Printf("[POLYGON] Updated %s price to $%.2f", symbol, quote.Results.Price)
	return nil
}

// UpdateAllSymbolPrices updates prices for all symbols in the database
func (s *Service) UpdateAllSymbolPrices(ctx context.Context) error {
	symbols, err := s.symbolService.GetDistinctSymbols()
	if err != nil {
		return fmt.Errorf("failed to get symbols: %w", err)
	}

	log.Printf("[POLYGON] Starting bulk price update for %d symbols", len(symbols))

	var updated, failed int
	for _, symbol := range symbols {
		if err := s.UpdateSymbolPrice(ctx, symbol); err != nil {
			log.Printf("[POLYGON] Failed to update %s: %v", symbol, err)
			failed++
		} else {
			updated++
		}

		// Rate limiting: Free tier allows 5 requests per minute
		time.Sleep(12 * time.Second)
	}

	log.Printf("[POLYGON] Bulk price update complete: %d updated, %d failed", updated, failed)
	return nil
}

// FetchSymbolDetails gets detailed information about a symbol from Polygon
func (s *Service) FetchSymbolDetails(ctx context.Context, symbol string) (*SymbolInfo, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get Polygon client: %w", err)
	}

	details, err := client.GetTickerDetails(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticker details: %w", err)
	}

	quote, err := client.GetPreviousClose(ctx, symbol)
	if err != nil {
		log.Printf("[POLYGON] Warning: failed to get current price for %s: %v", symbol, err)
		// Continue without current price
	}

	info := &SymbolInfo{
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
	}

	if quote != nil {
		info.CurrentPrice = quote.Results.Price
		info.PreviousClose = quote.Results.PreviousClose
		info.High = quote.Results.High
		info.Low = quote.Results.Low
		info.Volume = quote.Results.Volume
	}

	return info, nil
}

// FetchDividendHistory gets recent dividend history for a symbol
func (s *Service) FetchDividendHistory(ctx context.Context, symbol string, limit int) ([]*DividendInfo, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get Polygon client: %w", err)
	}

	if limit <= 0 {
		limit = 10
	}

	dividends, err := client.GetDividends(ctx, symbol, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get dividends: %w", err)
	}

	var result []*DividendInfo
	for _, div := range dividends.Results {
		info := &DividendInfo{
			Symbol:          div.Ticker,
			CashAmount:      div.CashAmount,
			DeclarationDate: div.DeclarationDate,
			ExDividendDate:  div.ExDividendDate,
			PayDate:         div.PayDate,
			RecordDate:      div.RecordDate,
			DividendType:    div.DividendType,
			Frequency:       div.Frequency,
		}
		result = append(result, info)
	}

	return result, nil
}

// TestConnection validates the API key and connection
func (s *Service) TestConnection(ctx context.Context) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}

	return client.IsValidAPIKey(ctx)
}

// GetAPIKeyStatus returns information about the current API key configuration
func (s *Service) GetAPIKeyStatus() *APIKeyStatus {
	apiKey := s.settingService.GetValue("POLYGON_API_KEY")
	
	status := &APIKeyStatus{
		Configured: apiKey != "",
		Masked:     "",
	}

	if status.Configured {
		// Mask the API key for display (show first 3 and last 3 characters)
		if len(apiKey) > 6 {
			status.Masked = apiKey[:3] + "..." + apiKey[len(apiKey)-3:]
		} else {
			status.Masked = strings.Repeat("*", len(apiKey))
		}
	}

	return status
}

// SymbolInfo represents enriched symbol information from Polygon
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
	Volume        float64   `json:"volume"`
}

// DividendInfo represents dividend information from Polygon
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

// APIKeyStatus represents the status of the Polygon API key
type APIKeyStatus struct {
	Configured bool   `json:"configured"`
	Masked     string `json:"masked"`
	Valid      bool   `json:"valid"`
	Error      string `json:"error,omitempty"`
}