package polygon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client represents a Polygon.io API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Polygon.io API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://api.polygon.io",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// StockQuote represents a stock quote response from Polygon.io
type StockQuote struct {
	Status string `json:"status"`
	Results struct {
		Symbol           string  `json:"T"`
		Price            float64 `json:"c"`
		High             float64 `json:"h"`
		Low              float64 `json:"l"`
		Open             float64 `json:"o"`
		Volume           float64   `json:"v"`
		PreviousClose    float64 `json:"pc"`
		Change           float64 `json:"change,omitempty"`
		ChangePercent    float64 `json:"changep,omitempty"`
		MarketStatus     string  `json:"market_status,omitempty"`
		Timestamp        int64   `json:"t"`
	} `json:"results"`
	RequestID string `json:"request_id"`
}

// TickerDetails represents detailed ticker information
type TickerDetails struct {
	Status string `json:"status"`
	Results struct {
		Symbol              string  `json:"ticker"`
		Name                string  `json:"name"`
		Market              string  `json:"market"`
		Locale              string  `json:"locale"`
		PrimaryExchange     string  `json:"primary_exchange"`
		Type                string  `json:"type"`
		Active              bool    `json:"active"`
		CurrencyName        string  `json:"currency_name"`
		CIK                 string  `json:"cik"`
		CompositeFigi       string  `json:"composite_figi"`
		ShareClassFigi      string  `json:"share_class_figi"`
		MarketCap           float64 `json:"market_cap"`
		PhoneNumber         string  `json:"phone_number"`
		Address             Address `json:"address"`
		Description         string  `json:"description"`
		SicCode             string  `json:"sic_code"`
		SicDescription      string  `json:"sic_description"`
		TickerRoot          string  `json:"ticker_root"`
		HomepageURL         string  `json:"homepage_url"`
		TotalEmployees      int     `json:"total_employees"`
		ListDate            string  `json:"list_date"`
		Branding            Branding `json:"branding"`
		ShareClassSharesOutstanding int64 `json:"share_class_shares_outstanding"`
		WeightedSharesOutstanding   int64 `json:"weighted_shares_outstanding"`
	} `json:"results"`
	RequestID string `json:"request_id"`
}

type Address struct {
	Address1   string `json:"address1"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
}

type Branding struct {
	LogoURL    string `json:"logo_url"`
	IconURL    string `json:"icon_url"`
}

type OptionSnapshot struct {
	Status  string `json:"status"`
	Results struct {
		BreakEvenPrice   float64 `json:"break_even_price"`
		Day              DayData `json:"day"`
		Details          OptionDetails `json:"details"`
		Greeks           Greeks `json:"greeks"`
		ImpliedVolatility float64 `json:"implied_volatility"`
		LastQuote        Quote `json:"last_quote"`
		LastTrade        Trade `json:"last_trade"`
		OpenInterest     float64 `json:"open_interest"`
		UnderlyingAsset  UnderlyingAsset `json:"underlying_asset"`
	} `json:"results"`
	RequestID string `json:"request_id"`
}

type DayData struct {
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"change_percent"`
	Close         float64 `json:"close"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Open          float64 `json:"open"`
	PreviousClose float64 `json:"previous_close"`
	Volume        float64 `json:"volume"`
	VWAP          float64 `json:"vwap"`
}

type OptionDetails struct {
	ContractType   string  `json:"contract_type"`
	ExerciseStyle  string  `json:"exercise_style"`
	ExpirationDate string  `json:"expiration_date"`
	SharesPerContract float64 `json:"shares_per_contract"`
	StrikePrice    float64 `json:"strike_price"`
	Ticker         string  `json:"ticker"`
}

type Greeks struct {
	Delta float64 `json:"delta"`
	Gamma float64 `json:"gamma"`
	Theta float64 `json:"theta"`
	Vega  float64 `json:"vega"`
}

type Quote struct {
	Ask            float64 `json:"ask"`
	AskSize        float64 `json:"ask_size"`
	Bid            float64 `json:"bid"`
	BidSize        float64 `json:"bid_size"`
	LastUpdated    int64   `json:"last_updated"`
	Midpoint       float64 `json:"midpoint"`
	Timeframe      string  `json:"timeframe"`
}

type Trade struct {
	Conditions      []int   `json:"conditions"`
	Exchange        int     `json:"exchange"`
	Price           float64 `json:"price"`
	SipTimestamp    int64   `json:"sip_timestamp"`
	Size            float64 `json:"size"`
}

type UnderlyingAsset struct {
	ChangeToBreakEven float64 `json:"change_to_break_even"`
	LastUpdated       int64   `json:"last_updated"`
	Price             float64 `json:"price"`
	Ticker            string  `json:"ticker"`
	Timeframe         string  `json:"timeframe"`
}

// DividendData represents dividend information
type DividendData struct {
	Status string `json:"status"`
	Results []struct {
		CashAmount      float64 `json:"cash_amount"`
		DeclarationDate string  `json:"declaration_date"`
		DividendType    string  `json:"dividend_type"`
		ExDividendDate  string  `json:"ex_dividend_date"`
		Frequency       int     `json:"frequency"`
		PayDate         string  `json:"pay_date"`
		RecordDate      string  `json:"record_date"`
		Ticker          string  `json:"ticker"`
	} `json:"results"`
	Count     int    `json:"count"`
	RequestID string `json:"request_id"`
}

// GetLastQuote fetches the last quote for a stock symbol
func (c *Client) GetLastQuote(ctx context.Context, symbol string) (*StockQuote, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("polygon API key not configured")
	}

	endpoint := fmt.Sprintf("/v2/last/nbbo/%s", url.PathEscape(symbol))
	url := fmt.Sprintf("%s%s?apikey=%s", c.baseURL, endpoint, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("unauthorized: invalid or missing Polygon API key (status 401)")
		} else if resp.StatusCode == http.StatusForbidden {
			return nil, fmt.Errorf("forbidden: API key may not have access to this endpoint (status 403)")
		}
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var quote StockQuote
	if err := json.NewDecoder(resp.Body).Decode(&quote); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if quote.Status != "OK" {
		return nil, fmt.Errorf("API returned status: %s", quote.Status)
	}

	return &quote, nil
}

// GetPreviousClose gets the previous trading day's close price for a symbol
func (c *Client) GetPreviousClose(ctx context.Context, symbol string) (*StockQuote, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("polygon API key not configured")
	}

	endpoint := fmt.Sprintf("/v2/aggs/ticker/%s/prev", url.PathEscape(symbol))
	url := fmt.Sprintf("%s%s?adjusted=true&apikey=%s", c.baseURL, endpoint, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("unauthorized: invalid or missing Polygon API key (status 401)")
		} else if resp.StatusCode == http.StatusForbidden {
			return nil, fmt.Errorf("forbidden: API key may not have access to this endpoint (status 403)")
		}
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var result struct {
		Status string `json:"status"`
		Results []struct {
			Symbol        string  `json:"T"`
			Volume        float64   `json:"v"`
			VolumeWeighted float64 `json:"vw"`
			Open          float64 `json:"o"`
			Close         float64 `json:"c"`
			High          float64 `json:"h"`
			Low           float64 `json:"l"`
			Timestamp     int64   `json:"t"`
			Transactions  int     `json:"n"`
		} `json:"results"`
		RequestID string `json:"request_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Status != "OK" || len(result.Results) == 0 {
		return nil, fmt.Errorf("API returned status: %s or no results", result.Status)
	}

	// Convert to StockQuote format
	quote := &StockQuote{
		Status: result.Status,
		RequestID: result.RequestID,
	}
	quote.Results.Symbol = result.Results[0].Symbol
	quote.Results.Price = result.Results[0].Close
	quote.Results.High = result.Results[0].High
	quote.Results.Low = result.Results[0].Low
	quote.Results.Open = result.Results[0].Open
	quote.Results.Volume = result.Results[0].Volume
	quote.Results.Timestamp = result.Results[0].Timestamp

	return quote, nil
}

// GetTickerDetails fetches detailed information about a ticker
func (c *Client) GetTickerDetails(ctx context.Context, symbol string) (*TickerDetails, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("polygon API key not configured")
	}

	endpoint := fmt.Sprintf("/v3/reference/tickers/%s", url.PathEscape(symbol))
	url := fmt.Sprintf("%s%s?apikey=%s", c.baseURL, endpoint, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var details TickerDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if details.Status != "OK" {
		return nil, fmt.Errorf("API returned status: %s", details.Status)
	}

	return &details, nil
}

// GetDividends fetches dividend information for a symbol
func (c *Client) GetDividends(ctx context.Context, symbol string, limit int) (*DividendData, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("polygon API key not configured")
	}

	if limit <= 0 {
		limit = 10
	}

	endpoint := fmt.Sprintf("/v3/reference/dividends")
	params := url.Values{}
	params.Set("ticker", symbol)
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("apikey", c.apiKey)

	url := fmt.Sprintf("%s%s?%s", c.baseURL, endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var dividends DividendData
	if err := json.NewDecoder(resp.Body).Decode(&dividends); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if dividends.Status != "OK" {
		return nil, fmt.Errorf("API returned status: %s", dividends.Status)
	}

	return &dividends, nil
}

// IsValidAPIKey tests if the API key is valid by making a simple request
func (c *Client) IsValidAPIKey(ctx context.Context) error {
	if c.apiKey == "" {
		return fmt.Errorf("polygon API key not configured")
	}

	// Test with a simple request to get market status
	endpoint := "/v1/marketstatus/now"
	url := fmt.Sprintf("%s%s?apikey=%s", c.baseURL, endpoint, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute test request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid or expired Polygon.io API key")
	}
	
	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("API key does not have permission to access Polygon.io endpoints")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API test request failed with status %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetOptionSnapshot(ctx context.Context, underlyingAsset, optionContract string) (*OptionSnapshot, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("polygon API key not configured")
	}

	endpoint := fmt.Sprintf("/v3/snapshot/options/%s/%s", 
		url.PathEscape(underlyingAsset), 
		url.PathEscape(optionContract))
	url := fmt.Sprintf("%s%s?apikey=%s", c.baseURL, endpoint, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("unauthorized: invalid or missing Polygon API key (status 401)")
		} else if resp.StatusCode == http.StatusForbidden {
			return nil, fmt.Errorf("forbidden: API key may not have access to this endpoint (status 403)")
		}
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var snapshot OptionSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if snapshot.Status != "OK" {
		return nil, fmt.Errorf("API returned status: %s", snapshot.Status)
	}

	return &snapshot, nil
}