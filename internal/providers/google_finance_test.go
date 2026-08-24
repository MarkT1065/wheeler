package providers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

const googleFinanceFixture = `<!doctype html>
<html><head><title>Apple Inc (AAPL) Stock Price &amp; News - Google Finance</title></head>
<body><meta itemprop="price" content="189.42"><span>· USD</span></body></html>`

type googleFinanceRoundTripper struct {
	body       string
	statusCode int
}

type googleFinanceValidationRoundTripper struct{}

func (googleFinanceValidationRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Path != "/finance/" {
		return nil, fmt.Errorf("validation must not use an exchange-specific quote path: %s", request.URL.Path)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("Google Finance")),
		Header:     make(http.Header),
		Request:    request,
	}, nil
}

func (t googleFinanceRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Path != "/finance/quote/AAPL:NASDAQ" && request.URL.Path != "/finance/quote/MISSING:NASDAQ" {
		return nil, fmt.Errorf("unexpected path: %s", request.URL.Path)
	}
	statusCode := t.statusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(t.body)),
		Header:     make(http.Header),
		Request:    request,
	}, nil
}

func TestGoogleFinanceProviderQuoteAndDetails(t *testing.T) {
	provider := newGoogleFinanceProviderForTest(&http.Client{Transport: googleFinanceRoundTripper{body: googleFinanceFixture}}, "http://google.test", "NASDAQ")
	quote, err := provider.GetQuote(context.Background(), "AAPL:NASDAQ")
	if err != nil {
		t.Fatalf("GetQuote failed: %v", err)
	}
	if quote.Symbol != "AAPL" || quote.Price != 189.42 {
		t.Fatalf("unexpected quote: %+v", quote)
	}

	details, err := provider.GetTickerDetails(context.Background(), "AAPL:NASDAQ")
	if err != nil {
		t.Fatalf("GetTickerDetails failed: %v", err)
	}
	if details.Name != "Apple Inc" || details.Currency != "USD" || details.Market != "NASDAQ" {
		t.Fatalf("unexpected details: %+v", details)
	}
}

func TestGoogleFinanceProviderMissingPrice(t *testing.T) {
	provider := newGoogleFinanceProviderForTest(&http.Client{Transport: googleFinanceRoundTripper{body: "<html><title>Not found</title></html>"}}, "http://google.test", "NASDAQ")
	_, err := provider.GetQuote(context.Background(), "AAPL:NASDAQ")
	if err == nil || !strings.Contains(err.Error(), "current price was not found") {
		t.Fatalf("expected missing-price error, got %v", err)
	}
}

func TestGoogleFinanceProviderUnknownSymbol(t *testing.T) {
	provider := newGoogleFinanceProviderForTest(&http.Client{Transport: googleFinanceRoundTripper{statusCode: http.StatusNotFound}}, "http://google.test", "")
	_, err := provider.GetQuote(context.Background(), "MISSING:NASDAQ")
	if err == nil || !strings.Contains(err.Error(), `unknown symbol "MISSING:NASDAQ"`) {
		t.Fatalf("expected unknown-symbol error, got %v", err)
	}
}

func TestGoogleFinanceProviderDividendsAreExplicitlyUnsupported(t *testing.T) {
	provider := NewGoogleFinanceProvider("NASDAQ")
	_, err := provider.GetDividends(context.Background(), "AAPL", 10)
	if !errors.Is(err, errGoogleFinanceDividendsUnsupported) {
		t.Fatalf("expected unsupported dividend error, got %v", err)
	}
}

func TestGoogleFinanceProviderValidationDoesNotDependOnConfiguredExchange(t *testing.T) {
	provider := newGoogleFinanceProviderForTest(&http.Client{Transport: googleFinanceValidationRoundTripper{}}, "http://google.test", "NYSE")
	if err := provider.ValidateConnection(context.Background()); err != nil {
		t.Fatalf("ValidateConnection failed for NYSE provider: %v", err)
	}
}
