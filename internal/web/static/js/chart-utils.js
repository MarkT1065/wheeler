// Chart color palette - bright theme colors matching Overall Allocation
export const CHART_COLORS = [
    '#4fc3f7',  // Bright Cyan (primary theme color)
    '#ff6b6b',  // Bright Red-Pink (critical/tech)
    '#2ecc71',  // Bright Green (success)
    '#3498db',  // Bright Blue (info)
    '#f39c12',  // Bright Orange (warning)
    '#9b59b6',  // Bright Purple (finance)
    '#1abc9c',  // Bright Turquoise
    '#e74c3c',  // Bright Red (danger)
    '#ffce56',  // Bright Yellow (treasuries color)
    '#29b6f6',  // Light Blue
    '#e67e22',  // Dark Orange
    '#8e44ad'   // Dark Purple
];

// Month names
export const MONTH_NAMES = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

// Get consistent color for a symbol
export function getSymbolColor(symbol, sortedSymbols) {
    const index = sortedSymbols.indexOf(symbol);
    return CHART_COLORS[index % CHART_COLORS.length];
}

// Sort ticker data by value (descending)
export function sortTickerData(labels, data) {
    const combined = labels.map((label, index) => ({
        ticker: label,
        amount: data[index]
    }));
    
    combined.sort((a, b) => b.amount - a.amount);
    
    return {
        labels: combined.map(item => item.ticker),
        data: combined.map(item => item.amount)
    };
}

// Format currency value as string
export function formatCurrencyValue(value) {
    return '$' + Math.round(Math.abs(value)).toLocaleString();
}

// Format currency and update element
export function formatCurrency(value, element) {
    const formattedValue = formatCurrencyValue(value);
    if (value < 0) {
        element.innerHTML = '-' + formattedValue;
    } else {
        element.innerHTML = formattedValue;
    }
    element.style.color = '#27ae60';
}

// Format percentage value
export function formatPercentage(value, element) {
    const formattedValue = value.toFixed(1) + '%';
    element.innerHTML = formattedValue;
    element.style.color = '#27ae60';
}

// Default pie chart options
export const PIE_CHART_OPTIONS = {
    responsive: true,
    maintainAspectRatio: false,
    interaction: {
        intersect: false
    },
    animation: {
        duration: 750
    },
    plugins: {
        legend: {
            position: 'right',
            align: 'start',
            maxWidth: 200,
            labels: {
                boxWidth: 12,
                font: {
                    size: 14
                },
                padding: 10,
                color: '#e0e0e0',
                usePointStyle: true,
                sort: function(a, b) {
                    return a.text.localeCompare(b.text);
                }
            }
        },
        tooltip: {
            enabled: true,
            mode: 'nearest',
            callbacks: {
                label: function(context) {
                    return context.label + ': $' + context.parsed.toLocaleString();
                }
            }
        },
        datalabels: {
            display: true,
            color: '#e0e0e0',
            font: {
                size: 12,
                weight: 'bold'
            },
            formatter: function(value) {
                if (value < 500) return '';
                return '$' + Math.round(value).toLocaleString();
            }
        }
    },
    layout: {
        padding: {
            top: 10,
            bottom: 10,
            left: 10,
            right: 120
        }
    }
};

// Default datalabels configuration for bar charts
export function createDatalabelsConfig(minValue = 500) {
    return {
        display: true,
        color: '#e0e0e0',
        font: {
            size: 12,
            weight: 'bold'
        },
        formatter: function(value) {
            if (value < minValue) return '';
            return '$' + Math.round(value).toLocaleString();
        }
    };
}

// Default tooltip configuration for financial charts
export const FINANCIAL_TOOLTIP = {
    callbacks: {
        label: function(context) {
            return context.label + ': $' + context.parsed.toLocaleString();
        }
    }
};

// Clone and merge options (for customizing base options)
export function mergeChartOptions(baseOptions, customOptions) {
    return JSON.parse(JSON.stringify(Object.assign({}, baseOptions, customOptions)));
}
