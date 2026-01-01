package models

import (
	"fmt"
	"strconv"
)

// SymbolPosition defines where the currency symbol should be placed
type SymbolPosition int

const (
	SymbolBefore SymbolPosition = iota // Symbol goes before the number (e.g., $100)
	SymbolAfter                       // Symbol goes after the number (e.g., 100€)
)

// Currency represents a currency with ISO code, full name, symbol, and position
type Currency struct {
	Code     string         // ISO 4217 currency code (e.g., USD, EUR)
	Name     string         // Full English name (e.g., US Dollar, Euro)
	Symbol   string         // Currency symbol (e.g., $, €, £)
	Position SymbolPosition // Whether symbol goes before or after the number
}

// Currencies is a list of supported currencies in the application
// The default currency is USD (US Dollar)
var Currencies = []Currency{
	{Code: "USD", Name: "US Dollar", Symbol: "$", Position: SymbolBefore},
	{Code: "EUR", Name: "Euro", Symbol: "€", Position: SymbolAfter},
	{Code: "GBP", Name: "British Pound", Symbol: "£", Position: SymbolBefore},
	{Code: "JPY", Name: "Japanese Yen", Symbol: "¥", Position: SymbolBefore},
	{Code: "CHF", Name: "Swiss Franc", Symbol: "CHF", Position: SymbolAfter},
	{Code: "CAD", Name: "Canadian Dollar", Symbol: "C$", Position: SymbolBefore},
	{Code: "AUD", Name: "Australian Dollar", Symbol: "A$", Position: SymbolBefore},
	{Code: "NZD", Name: "New Zealand Dollar", Symbol: "NZ$", Position: SymbolBefore},
	{Code: "CNY", Name: "Chinese Yuan", Symbol: "¥", Position: SymbolBefore},
	{Code: "HKD", Name: "Hong Kong Dollar", Symbol: "HK$", Position: SymbolBefore},
	{Code: "SGD", Name: "Singapore Dollar", Symbol: "S$", Position: SymbolBefore},
	{Code: "INR", Name: "Indian Rupee", Symbol: "₹", Position: SymbolBefore},
	{Code: "MXN", Name: "Mexican Peso", Symbol: "$", Position: SymbolBefore},
	{Code: "BRL", Name: "Brazilian Real", Symbol: "R$", Position: SymbolBefore},
	{Code: "KRW", Name: "South Korean Won", Symbol: "₩", Position: SymbolBefore},
	{Code: "SEK", Name: "Swedish Krona", Symbol: "kr", Position: SymbolAfter},
	{Code: "NOK", Name: "Norwegian Krone", Symbol: "kr", Position: SymbolAfter},
	{Code: "DKK", Name: "Danish Krone", Symbol: "kr", Position: SymbolAfter},
	{Code: "PLN", Name: "Polish Zloty", Symbol: "zł", Position: SymbolAfter},
	{Code: "RUB", Name: "Russian Ruble", Symbol: "₽", Position: SymbolBefore},
	{Code: "ZAR", Name: "South African Rand", Symbol: "R", Position: SymbolBefore},
	{Code: "TRY", Name: "Turkish Lira", Symbol: "₺", Position: SymbolBefore},
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

// GetCurrencySymbol returns the symbol for a currency code
func GetCurrencySymbol(code string) string {
	for _, currency := range Currencies {
		if currency.Code == code {
			return currency.Symbol
		}
	}
	return "$" // Default to dollar symbol
}

// FormatCurrency formats a value with the appropriate currency symbol
func FormatCurrency(value float64, currencyCode string) string {
	symbol := GetCurrencySymbol(currencyCode)
	currency := getCurrencyByCode(currencyCode)

	// Format the number without decimals
	formattedValue := strconv.FormatInt(int64(value), 10)

	if currency != nil && currency.Position == SymbolAfter {
		return formattedValue + symbol
	}
	return symbol + formattedValue
}

// FormatCurrencyWithDecimals formats a value with decimals and the appropriate currency symbol
func FormatCurrencyWithDecimals(value float64, currencyCode string) string {
	symbol := GetCurrencySymbol(currencyCode)
	currency := getCurrencyByCode(currencyCode)

	// Format the number with 2 decimal places
	formattedValue := fmt.Sprintf("%.2f", value)

	if currency != nil && currency.Position == SymbolAfter {
		return formattedValue + symbol
	}
	return symbol + formattedValue
}

// getCurrencyByCode returns the Currency struct for a given code
func getCurrencyByCode(code string) *Currency {
	for _, currency := range Currencies {
		if currency.Code == code {
			return &currency
		}
	}
	// Return USD as default
	for _, currency := range Currencies {
		if currency.Code == "USD" {
			return &currency
		}
	}
	return nil
}
