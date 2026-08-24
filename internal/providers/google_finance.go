package providers

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// GoogleFinanceProvider reads the public Google Finance quote page.
//
// Google Finance does not currently expose a supported public market-data API,
// so this provider is intentionally limited to the public quote page and may
// require maintenance if Google's HTML changes.
type GoogleFinanceProvider struct {
	baseURL    string
	httpClient *http.Client
}

var errGoogleFinanceDividendsUnsupported = errors.New("google finance does not provide dividend history through the public quote page")

var (
	googleFinancePricePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?is)<meta[^>]+itemprop=["']price["'][^>]+content=["']([0-9,.]+)["']`),
		regexp.MustCompile(`(?is)<[^>]+data-last-price=["']([0-9,.]+)["']`),
		regexp.MustCompile(`(?is)<[^>]+class=["'][^"']*YMlKec[^"']*["'][^>]*>\s*\$?\s*([0-9,.]+)`),
	}
	googleFinanceTitlePattern    = regexp.MustCompile(`(?is)<title[^>]*>\s*(.*?)\s*</title>`)
	googleFinanceWhitespace      = regexp.MustCompile(`\s+`)
	googleFinanceCurrencyPattern = regexp.MustCompile(`(?i)(?:·|&middot;)\s*([A-Z]{3})`)
)

// NewGoogleFinanceProvider creates a provider. The parameter is retained for
// source compatibility with legacy callers; market selection now comes from
// each canonical MARKET:TICKER symbol.
func NewGoogleFinanceProvider(_ string) *GoogleFinanceProvider {
	return &GoogleFinanceProvider{
		baseURL:    "https://www.google.com",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func newGoogleFinanceProviderForTest(client *http.Client, baseURL, exchange string) *GoogleFinanceProvider {
	p := NewGoogleFinanceProvider(exchange)
	p.httpClient = client
	p.baseURL = strings.TrimRight(baseURL, "/")
	return p
}

func (p *GoogleFinanceProvider) GetQuote(ctx context.Context, symbol string) (*Quote, error) {
	page, err := p.fetchPage(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("google finance: %w", err)
	}

	price, err := extractGoogleFinancePrice(page)
	if err != nil {
		return nil, fmt.Errorf("google finance: %w", err)
	}

	return &Quote{
		Symbol: normalizeGoogleFinanceSymbol(symbol),
		Price:  price,
		// The public page parser currently extracts only the current price.
		// Leave fields we did not observe at zero instead of duplicating data.
		Timestamp: time.Now(),
	}, nil
}

func (p *GoogleFinanceProvider) GetTickerDetails(ctx context.Context, symbol string) (*TickerDetails, error) {
	page, err := p.fetchPage(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("google finance: %w", err)
	}

	ticker := normalizeGoogleFinanceSymbol(symbol)
	name := extractGoogleFinanceTitle(page)
	if name == "" {
		name = ticker
	}

	return &TickerDetails{
		Symbol:   ticker,
		Name:     name,
		Market:   googleFinanceMarket(symbol),
		Type:     "CS",
		Active:   true,
		Currency: extractGoogleFinanceCurrency(page),
	}, nil
}

// GetDividends cannot provide a history from the public quote page. Returning
// an explicit error prevents the application from treating missing data as a
// zero-dividend history.
func (p *GoogleFinanceProvider) GetDividends(_ context.Context, _ string, _ int) ([]*Dividend, error) {
	return nil, errGoogleFinanceDividendsUnsupported
}

func (p *GoogleFinanceProvider) ValidateConnection(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(p.baseURL, "/")+"/finance/?hl=en", nil)
	if err != nil {
		return fmt.Errorf("google finance: create connection request: %w", err)
	}
	request.Header.Set("Accept", "text/html,application/xhtml+xml")
	request.Header.Set("User-Agent", "Wheeler/1.0 (+https://github.com/phbrgnomo/wheeler)")

	response, err := p.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("google finance: execute connection request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("google finance: connection page returned HTTP %d", response.StatusCode)
	}
	_, err = io.Copy(io.Discard, io.LimitReader(response.Body, 1024))
	if err != nil {
		return fmt.Errorf("google finance: read connection response: %w", err)
	}
	return nil
}

func (p *GoogleFinanceProvider) Name() string { return "Google Finance" }

func (p *GoogleFinanceProvider) fetchPage(ctx context.Context, symbol string) (string, error) {
	parts := strings.SplitN(strings.TrimSpace(symbol), ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("symbol must use TICKER:MARKET (for example AAPL:NASDAQ)")
	}
	ticker := strings.ToUpper(strings.TrimSpace(parts[0]))
	exchange := strings.ToUpper(strings.TrimSpace(parts[1]))
	if ticker == "" || exchange == "" {
		return "", fmt.Errorf("symbol must use TICKER:MARKET (for example AAPL:NASDAQ)")
	}

	path := "/finance/quote/" + url.PathEscape(ticker+":"+exchange)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(p.baseURL, "/")+path+"?hl=en", nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("Accept", "text/html,application/xhtml+xml")
	request.Header.Set("User-Agent", "Wheeler/1.0 (+https://github.com/phbrgnomo/wheeler)")

	response, err := p.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("execute request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("unknown symbol %q", symbol)
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("quote page returned HTTP %d", response.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	return html.UnescapeString(string(body)), nil
}

func extractGoogleFinancePrice(page string) (float64, error) {
	for _, pattern := range googleFinancePricePatterns {
		match := pattern.FindStringSubmatch(page)
		if len(match) < 2 {
			continue
		}
		value, err := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", ""), 64)
		if err == nil && value > 0 {
			return value, nil
		}
	}
	return 0, errors.New("current price was not found in the quote page")
}

func extractGoogleFinanceTitle(page string) string {
	match := googleFinanceTitlePattern.FindStringSubmatch(page)
	if len(match) < 2 {
		return ""
	}
	title := strings.TrimSpace(googleFinanceWhitespace.ReplaceAllString(match[1], " "))
	if index := strings.Index(title, " ("); index > 0 {
		return title[:index]
	}
	return strings.TrimSuffix(title, " Stock Price & News - Google Finance")
}

func extractGoogleFinanceCurrency(page string) string {
	match := googleFinanceCurrencyPattern.FindStringSubmatch(page)
	if len(match) >= 2 {
		return match[1]
	}
	return ""
}

func normalizeGoogleFinanceSymbol(symbol string) string {
	parts := strings.SplitN(strings.TrimSpace(symbol), ":", 2)
	if len(parts) == 2 {
		return strings.ToUpper(strings.TrimSpace(parts[0]))
	}
	return strings.ToUpper(strings.TrimSpace(symbol))
}

func googleFinanceMarket(symbol string) string {
	parts := strings.SplitN(strings.TrimSpace(symbol), ":", 2)
	if len(parts) == 2 {
		return strings.ToUpper(strings.TrimSpace(parts[1]))
	}
	return ""
}
