package providers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"stonks/internal/polygon"
)

type requestTrace struct {
	Method string
	URL    string
	Status int
}

type recordingTransport struct {
	handler func(*http.Request) (*http.Response, error)
	calls   []requestTrace
}

func (t *recordingTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := t.handler(request)
	if response != nil {
		t.calls = append(t.calls, requestTrace{
			Method: request.Method,
			URL:    sanitizeRequestURL(request.URL),
			Status: response.StatusCode,
		})
	}
	return response, err
}

func sanitizeRequestURL(value *url.URL) string {
	copyURL := *value
	query := copyURL.Query()
	for _, key := range []string{"apikey", "api_key", "token"} {
		if query.Has(key) {
			query.Set(key, "REDACTED")
		}
	}
	copyURL.RawQuery = query.Encode()
	return copyURL.String()
}

func traceResponse(status int, body string) (*http.Response, error) {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}

func assertAndLogTrace(t *testing.T, provider, operation string, trace requestTrace, wantURL, result string) {
	t.Helper()
	if trace.Method != http.MethodGet || trace.URL != wantURL || trace.Status != http.StatusOK {
		t.Fatalf("unexpected %s %s trace: %+v, want GET %s status 200", provider, operation, trace, wantURL)
	}
	t.Logf("TRACE provider=%s operation=%s method=%s url=%s status=%d result=%s", provider, operation, trace.Method, trace.URL, trace.Status, result)
}

func TestProviderRequestTraces(t *testing.T) {
	t.Run("google finance", func(t *testing.T) {
		transport := &recordingTransport{handler: func(request *http.Request) (*http.Response, error) {
			switch request.URL.Path {
			case "/finance/":
				return traceResponse(http.StatusOK, "Google Finance")
			case "/finance/quote/AAPL:NASDAQ":
				return traceResponse(http.StatusOK, googleFinanceFixture)
			default:
				return nil, fmt.Errorf("unexpected Google Finance path: %s", request.URL.Path)
			}
		}}
		provider := newGoogleFinanceProviderForTest(&http.Client{Transport: transport}, "http://google.test", "NASDAQ")

		quote, err := provider.GetQuote(context.Background(), "AAPL:NASDAQ")
		if err != nil {
			t.Fatal(err)
		}
		assertAndLogTrace(t, "google_finance", "quote", transport.calls[0], "http://google.test/finance/quote/AAPL:NASDAQ?hl=en", "price:189.42")
		if quote.Price != 189.42 {
			t.Fatalf("unexpected quote: %+v", quote)
		}

		details, err := provider.GetTickerDetails(context.Background(), "AAPL:NASDAQ")
		if err != nil {
			t.Fatal(err)
		}
		assertAndLogTrace(t, "google_finance", "details", transport.calls[1], "http://google.test/finance/quote/AAPL:NASDAQ?hl=en", "symbol:AAPL")
		if details.Name != "Apple Inc" {
			t.Fatalf("unexpected details: %+v", details)
		}

		if err := provider.ValidateConnection(context.Background()); err != nil {
			t.Fatal(err)
		}
		assertAndLogTrace(t, "google_finance", "validate_connection", transport.calls[2], "http://google.test/finance/?hl=en", "connected")

		before := len(transport.calls)
		if _, err := provider.GetDividends(context.Background(), "AAPL", 1); err != errGoogleFinanceDividendsUnsupported {
			t.Fatalf("expected unsupported dividends error, got %v", err)
		}
		if len(transport.calls) != before {
			t.Fatal("Google Finance dividends must not make an HTTP request")
		}
	})

	t.Run("yfinance", func(t *testing.T) {
		transport := &recordingTransport{handler: func(request *http.Request) (*http.Response, error) {
			if request.URL.Path != "/v8/finance/chart/AAPL" {
				return nil, fmt.Errorf("unexpected Yahoo Finance path: %s", request.URL.Path)
			}
			return traceResponse(http.StatusOK, yahooChartFixture)
		}}
		provider := newYFinanceProviderForTest(&http.Client{Transport: transport}, "http://yahoo.test")

		quote, err := provider.GetQuote(context.Background(), "AAPL")
		if err != nil {
			t.Fatal(err)
		}
		assertAndLogTrace(t, "yfinance", "quote", transport.calls[0], "http://yahoo.test/v8/finance/chart/AAPL?events=div%2Csplits&interval=1d&range=5d", "price:189.42")
		if quote.Price != 189.42 {
			t.Fatalf("unexpected quote: %+v", quote)
		}

		if _, err := provider.GetTickerDetails(context.Background(), "AAPL"); err != nil {
			t.Fatal(err)
		}
		assertAndLogTrace(t, "yfinance", "details", transport.calls[1], "http://yahoo.test/v8/finance/chart/AAPL?events=div%2Csplits&interval=1d&range=5d", "symbol:AAPL")

		dividends, err := provider.GetDividends(context.Background(), "AAPL", 1)
		if err != nil {
			t.Fatal(err)
		}
		assertAndLogTrace(t, "yfinance", "dividends", transport.calls[2], "http://yahoo.test/v8/finance/chart/AAPL?events=div%2Csplits&interval=1d&range=10y", "count:1")
		if len(dividends) != 1 {
			t.Fatalf("unexpected dividends: %+v", dividends)
		}

		if err := provider.ValidateConnection(context.Background()); err != nil {
			t.Fatal(err)
		}
		assertAndLogTrace(t, "yfinance", "validate_connection", transport.calls[3], "http://yahoo.test/v8/finance/chart/AAPL?events=div%2Csplits&interval=1d&range=5d", "connected")
	})

	t.Run("polygon", func(t *testing.T) {
		transport := &recordingTransport{handler: func(request *http.Request) (*http.Response, error) {
			switch request.URL.Path {
			case "/v2/aggs/ticker/AAPL/prev":
				return traceResponse(http.StatusOK, `{"status":"OK","results":[{"T":"AAPL","c":189.42,"h":190,"l":187,"o":188,"v":123,"t":1724328000000}]}`)
			case "/v3/reference/tickers/AAPL":
				return traceResponse(http.StatusOK, `{"status":"OK","results":{"ticker":"AAPL","name":"Apple Inc.","market":"stocks","type":"CS","active":true,"currency_name":"USD"}}`)
			case "/v3/reference/dividends":
				return traceResponse(http.StatusOK, `{"status":"OK","results":[{"ticker":"AAPL","cash_amount":0.25,"ex_dividend_date":"2024-08-13"}]}`)
			case "/v1/marketstatus/now":
				return traceResponse(http.StatusOK, `{}`)
			default:
				return nil, fmt.Errorf("unexpected Polygon path: %s", request.URL.Path)
			}
		}}
		client := polygon.NewClientWithBaseURL("secret-key", "http://polygon.test", &http.Client{Transport: transport})
		provider := newPolygonProviderForTest(client)

		quote, err := provider.GetQuote(context.Background(), "AAPL")
		if err != nil {
			t.Fatal(err)
		}
		assertAndLogTrace(t, "polygon", "quote", transport.calls[0], "http://polygon.test/v2/aggs/ticker/AAPL/prev?adjusted=true&apikey=REDACTED", "price:189.42")
		if quote.Price != 189.42 || quote.PreviousClose != 189.42 {
			t.Fatalf("unexpected quote: %+v", quote)
		}

		if _, err := provider.GetTickerDetails(context.Background(), "AAPL"); err != nil {
			t.Fatal(err)
		}
		assertAndLogTrace(t, "polygon", "details", transport.calls[1], "http://polygon.test/v3/reference/tickers/AAPL?apikey=REDACTED", "symbol:AAPL")

		dividends, err := provider.GetDividends(context.Background(), "AAPL", 1)
		if err != nil {
			t.Fatal(err)
		}
		assertAndLogTrace(t, "polygon", "dividends", transport.calls[2], "http://polygon.test/v3/reference/dividends?apikey=REDACTED&limit=1&ticker=AAPL", "count:1")
		if len(dividends) != 1 {
			t.Fatalf("unexpected dividends: %+v", dividends)
		}

		if err := provider.ValidateConnection(context.Background()); err != nil {
			t.Fatal(err)
		}
		assertAndLogTrace(t, "polygon", "validate_connection", transport.calls[3], "http://polygon.test/v1/marketstatus/now?apikey=REDACTED", "connected")
	})
}

func TestProviderMarketRequestPaths(t *testing.T) {
	tests := []struct {
		name, googleSymbol, yahooSymbol, googlePath, yahooPath string
	}{
		{"NASDAQ", "AAPL:NASDAQ", "AAPL", "/finance/quote/AAPL:NASDAQ", "/v8/finance/chart/AAPL"},
		{"NYSE", "IBM:NYSE", "IBM", "/finance/quote/IBM:NYSE", "/v8/finance/chart/IBM"},
		{"NYSEARCA", "SPY:NYSEARCA", "SPY", "/finance/quote/SPY:NYSEARCA", "/v8/finance/chart/SPY"},
		{"BVMF", "PETR4:BVMF", "PETR4.SA", "/finance/quote/PETR4:BVMF", "/v8/finance/chart/PETR4.SA"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			googleTransport := &recordingTransport{handler: func(request *http.Request) (*http.Response, error) {
				if request.URL.Path != tt.googlePath {
					return nil, fmt.Errorf("Google path = %s, want %s", request.URL.Path, tt.googlePath)
				}
				return traceResponse(http.StatusOK, googleFinanceFixture)
			}}
			google := newGoogleFinanceProviderForTest(&http.Client{Transport: googleTransport}, "http://google.test", "")
			if _, err := google.GetQuote(context.Background(), tt.googleSymbol); err != nil {
				t.Fatal(err)
			}
			assertAndLogTrace(t, "google_finance", "quote", googleTransport.calls[0], "http://google.test"+tt.googlePath+"?hl=en", "market:"+tt.name)

			yahooTransport := &recordingTransport{handler: func(request *http.Request) (*http.Response, error) {
				if request.URL.Path != tt.yahooPath {
					return nil, fmt.Errorf("Yahoo path = %s, want %s", request.URL.Path, tt.yahooPath)
				}
				return traceResponse(http.StatusOK, yahooChartFixture)
			}}
			yahoo := newYFinanceProviderForTest(&http.Client{Transport: yahooTransport}, "http://yahoo.test")
			if _, err := yahoo.GetQuote(context.Background(), tt.yahooSymbol); err != nil {
				t.Fatal(err)
			}
			assertAndLogTrace(t, "yfinance", "quote", yahooTransport.calls[0], "http://yahoo.test"+tt.yahooPath+"?events=div%2Csplits&interval=1d&range=5d", "market:"+tt.name)
		})
	}
}
