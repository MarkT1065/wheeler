package providers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

const yahooChartFixture = `{"chart":{"result":[{"meta":{"currency":"USD","symbol":"AAPL","exchangeName":"NMS","fullExchangeName":"NasdaqGS","instrumentType":"EQUITY","regularMarketTime":1724328000,"regularMarketPrice":189.42,"regularMarketDayHigh":190.00,"regularMarketDayLow":187.00,"regularMarketVolume":123456,"chartPreviousClose":188.50},"timestamp":[1724241600,1724328000],"indicators":{"quote":[{"open":[187.0,188.0],"high":[189.0,190.0],"low":[186.0,187.0],"close":[188.5,189.42],"volume":[100,123456]}]},"events":{"dividends":{"1723507200":{"amount":0.25},"1715731200":{"amount":0.24}}}}],"error":null}}`

type yahooRoundTripper struct{ body string }

func (t yahooRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Path != "/v8/finance/chart/AAPL" {
		return nil, fmt.Errorf("unexpected path: %s", request.URL.Path)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(t.body)),
		Header:     make(http.Header),
		Request:    request,
	}, nil
}

func TestYFinanceProviderQuoteAndDetails(t *testing.T) {
	provider := newYFinanceProviderForTest(&http.Client{Transport: yahooRoundTripper{body: yahooChartFixture}}, "http://yahoo.test")

	quote, err := provider.GetQuote(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("GetQuote failed: %v", err)
	}
	if quote.Symbol != "AAPL" || quote.Price != 189.42 || quote.PreviousClose != 188.50 {
		t.Fatalf("unexpected quote: %+v", quote)
	}

	details, err := provider.GetTickerDetails(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("GetTickerDetails failed: %v", err)
	}
	if details.Symbol != "AAPL" || details.Market != "NasdaqGS" || details.Currency != "USD" {
		t.Fatalf("unexpected details: %+v", details)
	}
}

func TestParseInt64RejectsTrailingCharacters(t *testing.T) {
	if _, err := parseInt64("1723507200invalid"); err == nil {
		t.Fatal("expected invalid epoch to be rejected")
	}
}

func TestYFinanceProviderDividendsNewestFirstAndLimited(t *testing.T) {
	provider := newYFinanceProviderForTest(&http.Client{Transport: yahooRoundTripper{body: yahooChartFixture}}, "http://yahoo.test")
	dividends, err := provider.GetDividends(context.Background(), "AAPL", 1)
	if err != nil {
		t.Fatalf("GetDividends failed: %v", err)
	}
	if len(dividends) != 1 || dividends[0].CashAmount != 0.25 || dividends[0].ExDividendDate != "2024-08-13" {
		t.Fatalf("unexpected dividends: %+v", dividends)
	}
}

func TestYFinanceProviderUsesMostRecentNonNullClose(t *testing.T) {
	fixture := strings.Replace(yahooChartFixture, `"regularMarketPrice":189.42`, `"regularMarketPrice":0`, 1)
	fixture = strings.Replace(fixture, `"close":[188.5,189.42]`, `"close":[188.5,189.42,null]`, 1)
	provider := newYFinanceProviderForTest(&http.Client{Transport: yahooRoundTripper{body: fixture}}, "http://yahoo.test")

	quote, err := provider.GetQuote(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("GetQuote failed: %v", err)
	}
	if quote.Price != 189.42 || quote.PreviousClose != 188.5 {
		t.Fatalf("unexpected quote after trailing null close: %+v", quote)
	}
}

func TestYFinanceProviderHTTPError(t *testing.T) {
	client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusTooManyRequests, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
	provider := newYFinanceProviderForTest(client, "http://yahoo.test")
	if _, err := provider.GetQuote(context.Background(), "AAPL"); err == nil || !strings.Contains(err.Error(), "HTTP 429") {
		t.Fatalf("expected HTTP 429 error, got %v", err)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
