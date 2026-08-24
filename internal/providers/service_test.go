package providers

import "testing"

func TestNormalizeProviderType(t *testing.T) {
	tests := map[string]string{
		"":               providerTypePolygon,
		" polygon ":      providerTypePolygon,
		"google":         providerTypeGoogle,
		"googlefinance":  providerTypeGoogle,
		"google_finance": providerTypeGoogle,
		"yahoo":          providerTypeYFinance,
		"yahoo_finance":  providerTypeYFinance,
		"YFINANCE":       providerTypeYFinance,
		"custom":         "custom",
	}
	for input, want := range tests {
		if got := normalizeProviderType(input); got != want {
			t.Errorf("normalizeProviderType(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCanonicalProviderTypeRejectsUnsupportedValues(t *testing.T) {
	if _, ok := CanonicalProviderType("unsupported"); ok {
		t.Fatal("unsupported provider type was accepted")
	}
}

func TestProviderProfiles(t *testing.T) {
	profiles := ProviderProfiles()
	if len(profiles) != 3 {
		t.Fatalf("expected three provider profiles, got %d", len(profiles))
	}
	if profiles[0].ID != providerTypePolygon || !profiles[0].RequiresAPIKey || !profiles[0].SupportsDividends {
		t.Fatalf("unexpected Polygon profile: %+v", profiles[0])
	}
	if profiles[1].ID != providerTypeGoogle || profiles[1].SupportsDividends {
		t.Fatalf("unexpected Google Finance profile: %+v", profiles[1])
	}
	if profiles[2].ID != providerTypeYFinance || profiles[2].RequiresAPIKey || !profiles[2].SupportsDividends {
		t.Fatalf("unexpected Yahoo Finance profile: %+v", profiles[2])
	}
}

func TestBulkPriceUpdateLimitFitsPolygonRequestBudget(t *testing.T) {
	if bulkPriceUpdateLimit >= 25 {
		t.Fatalf("bulk price update limit %d can exceed the five-minute Polygon request budget", bulkPriceUpdateLimit)
	}
}

func TestBulkPriceUpdateLimitIsExposed(t *testing.T) {
	if got := BulkPriceUpdateLimit(); got != bulkPriceUpdateLimit {
		t.Fatalf("BulkPriceUpdateLimit() = %d, want %d", got, bulkPriceUpdateLimit)
	}
}

func TestConvertDividendsToInfoSkipsNilEntries(t *testing.T) {
	dividends := convertDividendsToInfo([]*Dividend{nil, {Symbol: "AAPL", CashAmount: 0.25}})
	if len(dividends) != 1 || dividends[0].Symbol != "AAPL" {
		t.Fatalf("unexpected converted dividends: %+v", dividends)
	}
}
