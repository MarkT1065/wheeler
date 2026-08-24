// Package markets defines Wheeler's supported market identifiers and their
// provider-specific ticker formats.
package markets

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	NASDAQ   = "NASDAQ"
	NYSE     = "NYSE"
	NYSEARCA = "NYSEARCA"
	BVMF     = "BVMF"
)

var tickerPattern = regexp.MustCompile(`^[A-Z0-9.\-]{1,10}$`)

// Definition describes a market supported by Wheeler. ID is also the Google
// Finance exchange code and the first part of Wheeler's canonical symbol key.
type Definition struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	YahooSuffix string `json:"-"`
	Polygon     bool   `json:"polygon_supported"`
}

var definitions = map[string]Definition{
	NASDAQ:   {ID: NASDAQ, DisplayName: "NASDAQ", Polygon: true},
	NYSE:     {ID: NYSE, DisplayName: "NYSE", Polygon: true},
	NYSEARCA: {ID: NYSEARCA, DisplayName: "NYSE Arca", Polygon: true},
	BVMF:     {ID: BVMF, DisplayName: "B3 (Brasil)", YahooSuffix: ".SA"},
}

// Instrument is Wheeler's provider-neutral instrument identity.
type Instrument struct {
	Market string
	Ticker string
}

func (i Instrument) Key() string { return i.Market + ":" + i.Ticker }

// Definitions returns the supported markets in stable UI order.
func Definitions() []Definition {
	return []Definition{definitions[NASDAQ], definitions[NYSE], definitions[NYSEARCA], definitions[BVMF]}
}

func Get(id string) (Definition, bool) {
	definition, ok := definitions[strings.ToUpper(strings.TrimSpace(id))]
	return definition, ok
}

func New(market, ticker string) (Instrument, error) {
	definition, ok := Get(market)
	if !ok {
		return Instrument{}, fmt.Errorf("unsupported market: %q", strings.TrimSpace(market))
	}
	ticker = strings.ToUpper(strings.TrimSpace(ticker))
	if !tickerPattern.MatchString(ticker) {
		return Instrument{}, fmt.Errorf("invalid ticker: %q", ticker)
	}
	return Instrument{Market: definition.ID, Ticker: ticker}, nil
}

// Parse accepts only the canonical MARKET:TICKER key. Bare legacy tickers are
// intentionally rejected so callers cannot issue ambiguous provider requests.
func Parse(key string) (Instrument, error) {
	parts := strings.Split(strings.ToUpper(strings.TrimSpace(key)), ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return Instrument{}, fmt.Errorf("market selection required; use MARKET:TICKER")
	}
	return New(parts[0], parts[1])
}

func IsLegacyKey(key string) bool {
	return !strings.Contains(strings.TrimSpace(key), ":")
}

// ProviderSymbol converts a canonical key to the external ticker expected by
// a provider. providerType uses Wheeler's canonical provider identifiers.
func ProviderSymbol(providerType string, instrument Instrument) (string, error) {
	definition, ok := Get(instrument.Market)
	if !ok {
		return "", fmt.Errorf("unsupported market: %q", instrument.Market)
	}
	switch providerType {
	case "google_finance":
		return instrument.Ticker + ":" + definition.ID, nil
	case "yfinance":
		return instrument.Ticker + definition.YahooSuffix, nil
	case "polygon":
		if !definition.Polygon {
			return "", fmt.Errorf("Polygon.io does not support market %s", definition.ID)
		}
		return instrument.Ticker, nil
	default:
		return "", fmt.Errorf("unsupported data provider type: %s", providerType)
	}
}
