package markets

import "testing"

func TestProviderSymbols(t *testing.T) {
	tests := []struct {
		key, provider, want string
	}{
		{"NASDAQ:AAPL", "google_finance", "AAPL:NASDAQ"},
		{"NASDAQ:AAPL", "yfinance", "AAPL"},
		{"NYSE:IBM", "google_finance", "IBM:NYSE"},
		{"NYSE:IBM", "yfinance", "IBM"},
		{"NYSEARCA:SPY", "google_finance", "SPY:NYSEARCA"},
		{"NYSEARCA:SPY", "yfinance", "SPY"},
		{"BVMF:PETR4", "google_finance", "PETR4:BVMF"},
		{"BVMF:PETR4", "yfinance", "PETR4.SA"},
	}
	for _, tt := range tests {
		instrument, err := Parse(tt.key)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tt.key, err)
		}
		if got, err := ProviderSymbol(tt.provider, instrument); err != nil || got != tt.want {
			t.Errorf("ProviderSymbol(%s, %s) = %q, %v; want %q", tt.provider, tt.key, got, err, tt.want)
		}
	}
}

func TestRejectsLegacyAndUnsupportedSymbols(t *testing.T) {
	if _, err := Parse("AAPL"); err == nil {
		t.Fatal("expected legacy key to be rejected")
	}
	if _, err := Parse("LSE:VOD"); err == nil {
		t.Fatal("expected unsupported market to be rejected")
	}
	instrument, _ := Parse("BVMF:PETR4")
	if _, err := ProviderSymbol("polygon", instrument); err == nil {
		t.Fatal("expected BVMF to be rejected for Polygon")
	}
}
