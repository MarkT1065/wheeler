// Currency symbol mapping
const currencySymbols = {
    'USD': { symbol: '$', position: 'before' },
    'EUR': { symbol: '€', position: 'after' },
    'GBP': { symbol: '£', position: 'before' },
    'JPY': { symbol: '¥', position: 'before' },
    'CHF': { symbol: 'CHF', position: 'after' },
    'CAD': { symbol: 'C$', position: 'before' },
    'AUD': { symbol: 'A$', position: 'before' },
    'NZD': { symbol: 'NZ$', position: 'before' },
    'CNY': { symbol: '¥', position: 'before' },
    'HKD': { symbol: 'HK$', position: 'before' },
    'SGD': { symbol: 'S$', position: 'before' },
    'INR': { symbol: '₹', position: 'before' },
    'MXN': { symbol: '$', position: 'before' },
    'BRL': { symbol: 'R$', position: 'before' },
    'KRW': { symbol: '₩', position: 'before' },
    'SEK': { symbol: 'kr', position: 'after' },
    'NOK': { symbol: 'kr', position: 'after' },
    'DKK': { symbol: 'kr', position: 'after' },
    'PLN': { symbol: 'zł', position: 'after' },
    'RUB': { symbol: '₽', position: 'before' },
    'ZAR': { symbol: 'R', position: 'before' },
    'TRY': { symbol: '₺', position: 'before' }
};

// Currency color mapping for charts
const currencyColors = {
    'USD': '#3498db',
    'EUR': '#27ae60',
    'GBP': '#e74c3c',
    'JPY': '#f39c12',
    'CHF': '#9b59b6',
    'CAD': '#1abc9c',
    'AUD': '#e67e22',
    'NZD': '#16a085',
    'CNY': '#d35400',
    'HKD': '#c0392b',
    'SGD': '#8e44ad',
    'INR': '#27ae60',
    'MXN': '#2c3e50',
    'BRL': '#2ecc71',
    'KRW': '#e67e22',
    'SEK': '#3498db',
    'NOK': '#2980b9',
    'DKK': '#1abc9c',
    'PLN': '#9b59b6',
    'RUB': '#c0392b',
    'ZAR': '#f39c12',
    'TRY': '#d35400',
    'default': '#95a5a6'
};

// Format currency with symbol
// value: number to format
// currencyCode: ISO currency code (e.g., 'USD', 'EUR')
// decimals: number of decimal places (default: 2)
function formatCurrencyWithSymbol(value, currencyCode, decimals = 2) {
    const currency = currencySymbols[currencyCode] || currencySymbols['USD'];
    const formatted = value.toFixed(decimals % 1 === 0 ? 0 : decimals);
    if (currency.position === 'after') {
        return formatted + currency.symbol;
    }
    return currency.symbol + formatted;
}

// Format currency with symbol for integers (no decimals)
function formatCurrencyInt(value, currencyCode) {
    return formatCurrencyWithSymbol(value, currencyCode, 0);
}
