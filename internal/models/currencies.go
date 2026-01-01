package models

// Currency represents a currency with ISO code and full name
type Currency struct {
	Code string // ISO 4217 currency code (e.g., USD, EUR)
	Name string // Full English name (e.g., US Dollar, Euro)
}

// Currencies is a list of supported currencies in the application
// The default currency is USD (US Dollar)
var Currencies = []Currency{
	{Code: "USD", Name: "US Dollar"},
	{Code: "EUR", Name: "Euro"},
	{Code: "GBP", Name: "British Pound"},
	{Code: "JPY", Name: "Japanese Yen"},
	{Code: "CHF", Name: "Swiss Franc"},
	{Code: "CAD", Name: "Canadian Dollar"},
	{Code: "AUD", Name: "Australian Dollar"},
	{Code: "NZD", Name: "New Zealand Dollar"},
	{Code: "CNY", Name: "Chinese Yuan"},
	{Code: "HKD", Name: "Hong Kong Dollar"},
	{Code: "SGD", Name: "Singapore Dollar"},
	{Code: "INR", Name: "Indian Rupee"},
	{Code: "MXN", Name: "Mexican Peso"},
	{Code: "BRL", Name: "Brazilian Real"},
	{Code: "KRW", Name: "South Korean Won"},
	{Code: "SEK", Name: "Swedish Krona"},
	{Code: "NOK", Name: "Norwegian Krone"},
	{Code: "DKK", Name: "Danish Krone"},
	{Code: "PLN", Name: "Polish Zloty"},
	{Code: "RUB", Name: "Russian Ruble"},
	{Code: "ZAR", Name: "South African Rand"},
	{Code: "TRY", Name: "Turkish Lira"},
}

// DefaultCurrency is the default currency for symbols that don't have one specified
const DefaultCurrency = "USD"

// GetCurrencyName returns the full name of a currency by its code
func GetCurrencyName(code string) string {
	for _, currency := range Currencies {
		if currency.Code == code {
			return currency.Name
		}
	}
	return "Unknown"
}

// IsValidCurrency checks if a currency code is valid
func IsValidCurrency(code string) bool {
	for _, currency := range Currencies {
		if currency.Code == code {
			return true
		}
	}
	return false
}
