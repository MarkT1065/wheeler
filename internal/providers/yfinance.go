package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// YFinanceProvider retrieves Yahoo Finance data directly over HTTP. The name
// is retained for compatibility with the yfinance ecosystem, but this Go
// implementation does not require Python or the yfinance package.
type YFinanceProvider struct {
	baseURL    string
	httpClient *http.Client
}

type yahooChartResponse struct {
	Chart struct {
		Result []struct {
			Meta       yahooMeta            `json:"meta"`
			Timestamp  []int64              `json:"timestamp"`
			Indicators yahooChartIndicators `json:"indicators"`
			Events     yahooChartEvents     `json:"events"`
		} `json:"result"`
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

type yahooMeta struct {
	Currency             string  `json:"currency"`
	Symbol               string  `json:"symbol"`
	ExchangeName         string  `json:"exchangeName"`
	FullExchangeName     string  `json:"fullExchangeName"`
	InstrumentType       string  `json:"instrumentType"`
	RegularMarketTime    int64   `json:"regularMarketTime"`
	RegularMarketPrice   float64 `json:"regularMarketPrice"`
	RegularMarketDayHigh float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow  float64 `json:"regularMarketDayLow"`
	RegularMarketVolume  float64 `json:"regularMarketVolume"`
	ChartPreviousClose   float64 `json:"chartPreviousClose"`
}

type yahooChartIndicators struct {
	Quote []struct {
		Open   []*float64 `json:"open"`
		High   []*float64 `json:"high"`
		Low    []*float64 `json:"low"`
		Close  []*float64 `json:"close"`
		Volume []*float64 `json:"volume"`
	} `json:"quote"`
}

type yahooChartEvents struct {
	Dividends map[string]struct {
		Amount float64 `json:"amount"`
	} `json:"dividends"`
}

// NewYFinanceProvider creates a direct Yahoo Finance HTTP provider. The
// YFINANCE_BASE_URL environment variable is useful for testing or proxies.
func NewYFinanceProvider() *YFinanceProvider {
	baseURL := os.Getenv("YFINANCE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://query1.finance.yahoo.com"
	}
	return &YFinanceProvider{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func newYFinanceProviderForTest(client *http.Client, baseURL string) *YFinanceProvider {
	provider := NewYFinanceProvider()
	provider.httpClient = client
	provider.baseURL = strings.TrimRight(baseURL, "/")
	return provider
}

func (p *YFinanceProvider) GetQuote(ctx context.Context, symbol string) (*Quote, error) {
	result, err := p.fetchChart(ctx, symbol, "5d")
	if err != nil {
		return nil, fmt.Errorf("yfinance: %w", err)
	}
	meta := result.Meta
	latest := latestYahooQuote(result)
	price := meta.RegularMarketPrice
	if price <= 0 {
		price = latest.close
	}
	if price <= 0 {
		return nil, errors.New("yfinance: quote returned no positive price")
	}
	previousClose := meta.ChartPreviousClose
	if previousClose <= 0 {
		previousClose = latest.previousClose
	}
	timestamp := time.Unix(meta.RegularMarketTime, 0)
	if meta.RegularMarketTime == 0 {
		timestamp = time.Now()
	}
	return &Quote{
		Symbol:        meta.Symbol,
		Price:         price,
		Open:          latest.open,
		High:          firstPositive(meta.RegularMarketDayHigh, latest.high),
		Low:           firstPositive(meta.RegularMarketDayLow, latest.low),
		Volume:        firstPositive(meta.RegularMarketVolume, latest.volume),
		PreviousClose: previousClose,
		Timestamp:     timestamp,
	}, nil
}

func (p *YFinanceProvider) GetTickerDetails(ctx context.Context, symbol string) (*TickerDetails, error) {
	result, err := p.fetchChart(ctx, symbol, "5d")
	if err != nil {
		return nil, fmt.Errorf("yfinance: %w", err)
	}
	meta := result.Meta
	return &TickerDetails{
		Symbol:   meta.Symbol,
		Name:     meta.Symbol,
		Market:   firstNonEmpty(meta.FullExchangeName, meta.ExchangeName),
		Type:     meta.InstrumentType,
		Active:   true,
		Currency: meta.Currency,
	}, nil
}

func (p *YFinanceProvider) GetDividends(ctx context.Context, symbol string, limit int) ([]*Dividend, error) {
	if limit <= 0 {
		limit = 10
	}
	result, err := p.fetchChart(ctx, symbol, "10y")
	if err != nil {
		return nil, fmt.Errorf("yfinance: %w", err)
	}
	items := make([]*Dividend, 0, len(result.Events.Dividends))
	for epoch, event := range result.Events.Dividends {
		seconds, parseErr := parseInt64(epoch)
		if parseErr != nil {
			continue
		}
		items = append(items, &Dividend{
			Symbol:         result.Meta.Symbol,
			CashAmount:     event.Amount,
			ExDividendDate: time.Unix(seconds, 0).UTC().Format("2006-01-02"),
			DividendType:   "cash",
		})
	}
	sortDividendsNewestFirst(items)
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (p *YFinanceProvider) ValidateConnection(ctx context.Context) error {
	_, err := p.GetQuote(ctx, "AAPL")
	return err
}

func (p *YFinanceProvider) Name() string { return "Yahoo Finance (yfinance)" }

type yahooChartResult struct {
	Meta       yahooMeta
	Timestamp  []int64
	Indicators yahooChartIndicators
	Events     yahooChartEvents
}

func (p *YFinanceProvider) fetchChart(ctx context.Context, symbol, rangeValue string) (*yahooChartResult, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, errors.New("symbol is required")
	}
	query := url.Values{}
	query.Set("range", rangeValue)
	query.Set("interval", "1d")
	query.Set("events", "div,splits")
	endpoint := fmt.Sprintf("%s/v8/finance/chart/%s?%s", p.baseURL, url.PathEscape(symbol), query.Encode())
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "Wheeler/1.0 (+https://github.com/phbrgnomo/wheeler)")
	response, err := p.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Yahoo Finance returned HTTP %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	var payload yahooChartResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if payload.Chart.Error != nil {
		return nil, fmt.Errorf("Yahoo Finance error %s: %s", payload.Chart.Error.Code, payload.Chart.Error.Description)
	}
	if len(payload.Chart.Result) == 0 {
		return nil, errors.New("Yahoo Finance returned no chart result")
	}
	result := payload.Chart.Result[0]
	return &yahooChartResult{
		Meta: result.Meta, Timestamp: result.Timestamp,
		Indicators: result.Indicators, Events: result.Events,
	}, nil
}

type yahooLatestQuote struct {
	open, high, low, close, volume, previousClose float64
}

func latestYahooQuote(result *yahooChartResult) yahooLatestQuote {
	if len(result.Indicators.Quote) == 0 {
		return yahooLatestQuote{}
	}
	quote := result.Indicators.Quote[0]
	latest := -1
	for i := len(quote.Close) - 1; i >= 0; i-- {
		if quote.Close[i] != nil {
			latest = i
			break
		}
	}
	if latest < 0 {
		return yahooLatestQuote{}
	}
	previous := -1
	for i := latest - 1; i >= 0; i-- {
		if quote.Close[i] != nil {
			previous = i
			break
		}
	}
	get := func(values []*float64, index int) float64 {
		if index >= 0 && index < len(values) && values[index] != nil {
			return *values[index]
		}
		return 0
	}
	return yahooLatestQuote{
		open: get(quote.Open, latest), high: get(quote.High, latest),
		low: get(quote.Low, latest), close: get(quote.Close, latest),
		volume: get(quote.Volume, latest), previousClose: get(quote.Close, previous),
	}
}

func firstPositive(primary, fallback float64) float64 {
	if primary > 0 {
		return primary
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func parseInt64(value string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(value), 10, 64)
}

func sortDividendsNewestFirst(items []*Dividend) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].ExDividendDate > items[j].ExDividendDate
	})
}

var _ Provider = (*YFinanceProvider)(nil)
