package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"stonks/internal/database"
	"stonks/internal/providers"
)

func newProviderTestServer(t *testing.T, name string) (*Server, *database.DB) {
	t.Helper()
	db, err := database.NewDB(filepath.Join(t.TempDir(), name+".db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	server := &Server{}
	server.initializeServices(db.DB)
	return server, db
}

func TestProviderSettingValidation(t *testing.T) {
	server, _ := newProviderTestServer(t, "settings")

	request := httptest.NewRequest(http.MethodPut, "/api/settings/DATA_PROVIDER_TYPE", strings.NewReader(`{"value":"google","description":""}`))
	recorder := httptest.NewRecorder()
	server.updateSettingAPI(recorder, request, "DATA_PROVIDER_TYPE")
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected valid provider update, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if got := server.settingService.GetValue("DATA_PROVIDER_TYPE"); got != "google_finance" {
		t.Fatalf("expected canonical provider type, got %q", got)
	}

	request = httptest.NewRequest(http.MethodPut, "/api/settings/DATA_PROVIDER_TYPE", strings.NewReader(`{"value":"not-a-provider"}`))
	recorder = httptest.NewRecorder()
	server.updateSettingAPI(recorder, request, "DATA_PROVIDER_TYPE")
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid provider to be rejected, got %d", recorder.Code)
	}
}

func TestOptionAPIStoresLongDirection(t *testing.T) {
	server, _ := newProviderTestServer(t, "long-option")
	if _, err := server.symbolService.Create("NASDAQ:AAPL"); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/options", strings.NewReader(`{"symbol":"NASDAQ:AAPL","type":"Put","direction":"Long","opened":"2026-08-01","strike":100,"expiration":"2026-09-01","premium":2,"contracts":1,"commission":0.65}`))
	recorder := httptest.NewRecorder()
	server.optionAPIHandler(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("create long option = %d: %s", recorder.Code, recorder.Body.String())
	}
	var option struct {
		ID        int    `json:"id"`
		Direction string `json:"direction"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&option); err != nil {
		t.Fatal(err)
	}
	if option.Direction != "Long" {
		t.Fatalf("direction = %q", option.Direction)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/options/"+strconv.Itoa(option.ID), nil)
	recorder = httptest.NewRecorder()
	server.individualOptionAPIHandler(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("get long option = %d: %s", recorder.Code, recorder.Body.String())
	}
	var retrieved struct {
		Direction string `json:"direction"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&retrieved); err != nil {
		t.Fatal(err)
	}
	if retrieved.Direction != "Long" {
		t.Fatalf("retrieved direction = %q", retrieved.Direction)
	}
}

func TestProviderServiceIsRecreatedForNewDatabase(t *testing.T) {
	server, first := newProviderTestServer(t, "first")
	if err := server.settingService.SetValue("DATA_PROVIDER_TYPE", "google_finance", ""); err != nil {
		t.Fatal(err)
	}
	if got := server.providerService.Name(); got != "google_finance" {
		t.Fatalf("first database provider = %q", got)
	}

	second, err := database.NewDB(filepath.Join(t.TempDir(), "second.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = second.Close() })
	if err := server.switchServices(second.DB); err != nil {
		t.Fatal(err)
	}
	if err := server.settingService.SetValue("DATA_PROVIDER_TYPE", "yfinance", ""); err != nil {
		t.Fatal(err)
	}
	if got := server.providerService.Name(); got != "yfinance" {
		t.Fatalf("provider service retained first database configuration: %q", got)
	}
	_ = first
}

func TestGoogleFinanceDividendEndpointIsDisabled(t *testing.T) {
	server, _ := newProviderTestServer(t, "google-dividends")
	if err := server.settingService.SetValue("DATA_PROVIDER_TYPE", "google_finance", ""); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/provider/fetch-dividends", bytes.NewBufferString(`{"symbols":["AAPL"]}`))
	recorder := httptest.NewRecorder()
	server.providerFetchDividendsHandler(recorder, request)
	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected unsupported dividends response, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestProviderUpdateRejectsMalformedJSON(t *testing.T) {
	server, _ := newProviderTestServer(t, "malformed-provider-update")
	request := httptest.NewRequest(http.MethodPost, "/api/provider/update-prices", strings.NewReader(`{"all":`))
	recorder := httptest.NewRecorder()
	server.providerUpdatePricesHandler(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected malformed JSON to be rejected, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestProviderUpdateResponseExposesBulkLimit(t *testing.T) {
	server, _ := newProviderTestServer(t, "provider-update-limit")
	request := httptest.NewRequest(http.MethodPost, "/api/provider/update-prices", strings.NewReader(`{"all":true}`))
	recorder := httptest.NewRecorder()
	server.providerUpdatePricesHandler(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("provider update = %d: %s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Limit int `json:"limit"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Limit != providers.BulkPriceUpdateLimit() {
		t.Fatalf("response limit = %d, want %d", response.Limit, providers.BulkPriceUpdateLimit())
	}
}

func TestProviderSymbolValidation(t *testing.T) {
	valid, err := validateSymbols([]string{" nasdaq:aapl ", "nasdaq:msft", "NYSE:BRK.B"})
	if err != nil {
		t.Fatalf("expected valid symbols, got %v", err)
	}
	want := []string{"NASDAQ:AAPL", "NASDAQ:MSFT", "NYSE:BRK.B"}
	if strings.Join(valid, ",") != strings.Join(want, ",") {
		t.Fatalf("validated symbols = %v, want %v", valid, want)
	}

	if _, err := validateSymbols([]string{"NASDAQ:AAPL", "../../etc/passwd"}); err == nil || !strings.Contains(err.Error(), "../../etc/passwd") {
		t.Fatalf("expected invalid symbol to be identified, got %v", err)
	}
}

func TestProviderConfigurationIsAtomic(t *testing.T) {
	server, _ := newProviderTestServer(t, "provider-configuration")
	if err := server.settingService.SetValue("DATA_PROVIDER_TYPE", "yfinance", ""); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPut, "/api/provider/configuration", strings.NewReader(`{"provider_type":"polygon","polygon_api_key":""}`))
	recorder := httptest.NewRecorder()
	server.providerConfigurationAPIHandler(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected missing Polygon key to be rejected, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if got := server.settingService.GetValue("DATA_PROVIDER_TYPE"); got != "yfinance" {
		t.Fatalf("failed configuration activated provider %q", got)
	}

	request = httptest.NewRequest(http.MethodPut, "/api/provider/configuration", strings.NewReader(`{"provider_type":"google"}`))
	recorder = httptest.NewRecorder()
	server.providerConfigurationAPIHandler(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected Google configuration to succeed, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if got := server.settingService.GetValue("DATA_PROVIDER_TYPE"); got != "google_finance" {
		t.Fatalf("provider = %q, want google_finance", got)
	}
}

func TestProviderValidationCache(t *testing.T) {
	server := &Server{}
	checks := 0
	validate := func() error {
		checks++
		return nil
	}

	for range 2 {
		valid, errText := server.cachedProviderValidation("yfinance", validate)
		if !valid || errText != "" {
			t.Fatalf("unexpected validation result: valid=%v error=%q", valid, errText)
		}
	}
	if checks != 1 {
		t.Fatalf("expected cached validation to run once, ran %d times", checks)
	}

	server.invalidateProviderValidation()
	if valid, errText := server.cachedProviderValidation("yfinance", validate); !valid || errText != "" {
		t.Fatalf("unexpected validation after invalidation: valid=%v error=%q", valid, errText)
	}
	if checks != 2 {
		t.Fatalf("expected invalidated validation to run again, ran %d times", checks)
	}
}

func TestProviderValidationCacheCoalescesConcurrentRequests(t *testing.T) {
	server := &Server{}
	started := make(chan struct{})
	waiting := make(chan struct{})
	release := make(chan struct{})
	results := make(chan struct {
		valid bool
		err   string
	}, 2)
	checks := 0
	server.providerValidationWaitHook = func() { close(waiting) }
	validate := func() error {
		checks++
		close(started)
		<-release
		return nil
	}

	go func() {
		valid, errText := server.cachedProviderValidation("yfinance", validate)
		results <- struct {
			valid bool
			err   string
		}{valid, errText}
	}()
	<-started
	go func() {
		valid, errText := server.cachedProviderValidation("yfinance", validate)
		results <- struct {
			valid bool
			err   string
		}{valid, errText}
	}()

	// Do not release the first validation until the second request has entered
	// the in-flight wait path. This proves coalescing rather than a cache hit.
	<-waiting
	close(release)
	for range 2 {
		result := <-results
		if !result.valid || result.err != "" {
			t.Fatalf("unexpected validation result: %+v", result)
		}
	}
	if checks != 1 {
		t.Fatalf("expected one concurrent validation request, got %d", checks)
	}
}

func TestDatabaseSwitchDoesNotWaitForLeasedProviderServices(t *testing.T) {
	server, first := newProviderTestServer(t, "leased-first")
	_, release := server.acquireProviderServices()

	second, err := database.NewDB(filepath.Join(t.TempDir(), "leased-second.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = second.Close() })

	done := make(chan error, 1)
	go func() { done <- server.switchServices(second.DB) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("database switch waited for the leased provider request")
	}
	if err := first.DB.Ping(); err != nil {
		t.Fatalf("leased database was closed during provider request: %v", err)
	}

	release()
	if err := first.DB.Ping(); err == nil {
		t.Fatal("retired database remained open after provider request completed")
	}
}

func TestProviderConfigurationTemplateContainsAllProviderControls(t *testing.T) {
	contents, err := os.ReadFile("templates/settings.html")
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"providerTypeInput", "apiKeyField", "noApiKeyField", "Dividend history:", "Market selection required", "LegacySymbols", "savedProviderType", "Unsaved provider selection"} {
		if !strings.Contains(string(contents), value) {
			t.Errorf("settings template is missing %q", value)
		}
	}

	profileJSON, err := json.Marshal(providers.ProviderProfiles())
	if err != nil || !strings.Contains(string(profileJSON), "google_finance") || !strings.Contains(string(profileJSON), "yfinance") {
		t.Fatalf("provider profiles cannot be represented for the settings UI: %s, %v", profileJSON, err)
	}
}

func TestSymbolModalRequiresMarketSelection(t *testing.T) {
	contents, err := os.ReadFile("templates/_symbol_modal.html")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), `id="marketInput"`) || !strings.Contains(string(contents), "MARKET:TICKER") {
		t.Fatalf("symbol modal does not expose the market selection: %s", contents)
	}
}

func TestSymbolModalQualifiesLegacySymbols(t *testing.T) {
	contents, err := os.ReadFile("static/js/symbol-modal.js")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), "this.editingSymbol.includes(':')") || !strings.Contains(string(contents), "/qualify") || !strings.Contains(string(contents), "response.text()") {
		t.Fatal("symbol modal does not route legacy edits through the qualification endpoint")
	}
}

func TestSymbolProviderOperationsNormalizePathSymbols(t *testing.T) {
	server, _ := newProviderTestServer(t, "symbol-provider-normalization")
	for name, handler := range map[string]func(http.ResponseWriter, *http.Request, string){
		"update price":    server.symbolUpdatePriceHandler,
		"fetch dividends": server.symbolFetchDividendsHandler,
	} {
		t.Run(name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/symbols/nasdaq:aapl", nil)
			recorder := httptest.NewRecorder()
			handler(recorder, request, "nasdaq:aapl")

			var response struct {
				Symbol string `json:"symbol"`
			}
			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatal(err)
			}
			if response.Symbol != "NASDAQ:AAPL" {
				t.Fatalf("symbol = %q, want NASDAQ:AAPL", response.Symbol)
			}
		})
	}
}

func TestSymbolTemplateRestoresMarketAfterOptionsLoad(t *testing.T) {
	contents, err := os.ReadFile("templates/symbol.html")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), "selectedMarket") || !strings.Contains(string(contents), "marketInput.value = selectedMarket") {
		t.Fatal("symbol template does not restore the selected market after loading options")
	}
}

func TestOptionsTemplateChartParsersAccountForPositionColumn(t *testing.T) {
	contents, err := os.ReadFile("templates/options.html")
	if err != nil {
		t.Fatal(err)
	}
	template := string(contents)
	for _, expected := range []string{
		"data-symbol=\"{{.Symbol}}\"",
		"data-type=\"{{if eq .Type \"Put\"}}Put{{else}}Call{{end}}\"",
		"data-direction=\"{{.Direction}}\"",
		"data-strike=\"{{printf \"%.2f\" .Strike}}\"",
		"data-contracts=\"{{.Contracts}}\"",
		"data-total-profit=\"{{printf \"%.2f\" .CalculateTotalProfit}}\"",
		"row.querySelector('[data-type]')?.dataset.type",
		"row.querySelector('[data-symbol]')?.dataset.symbol",
		"type === 'Put' && direction === 'Short'",
		"row.querySelector('[data-total-profit]')?.dataset.totalProfit",
	} {
		if !strings.Contains(template, expected) {
			t.Errorf("options chart parser is missing %q", expected)
		}
	}
}

func TestMarketsAPIListsCanonicalMarkets(t *testing.T) {
	server, _ := newProviderTestServer(t, "markets")
	request := httptest.NewRequest(http.MethodGet, "/api/markets", nil)
	recorder := httptest.NewRecorder()
	server.marketsAPIHandler(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("markets endpoint = %d", recorder.Code)
	}
	for _, market := range []string{"NASDAQ", "NYSE", "NYSEARCA", "BVMF"} {
		if !strings.Contains(recorder.Body.String(), market) {
			t.Fatalf("markets response does not include %s: %s", market, recorder.Body.String())
		}
	}
}
