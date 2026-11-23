-- ============================================================================
-- ADD HISTORY TABLES
-- ============================================================================
-- Adds tables for storing historical price data from Polygon/Massive API:
-- - price_history: Daily OHLCV bars for tracked stock symbols
-- - option_history: Daily OHLCV bars for option contracts
-- - index_history: Daily OHLC bars for market indices (SPX, NDX, DJI, VIX)
-- ============================================================================

CREATE TABLE IF NOT EXISTS price_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    bar_date DATE NOT NULL,
    open REAL NOT NULL,
    high REAL NOT NULL,
    low REAL NOT NULL,
    close REAL NOT NULL,
    volume REAL DEFAULT 0,
    vwap REAL DEFAULT 0,
    transactions INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (symbol) REFERENCES symbols(symbol)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_price_history_symbol_date
    ON price_history(symbol, bar_date);
CREATE INDEX IF NOT EXISTS idx_price_history_bar_date
    ON price_history(bar_date);

CREATE TABLE IF NOT EXISTS option_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    option_ticker TEXT NOT NULL,
    underlying_symbol TEXT NOT NULL,
    contract_type TEXT NOT NULL CHECK (contract_type IN ('call', 'put')),
    strike_price REAL NOT NULL,
    expiration_date DATE NOT NULL,
    bar_date DATE NOT NULL,
    open REAL NOT NULL,
    high REAL NOT NULL,
    low REAL NOT NULL,
    close REAL NOT NULL,
    volume REAL DEFAULT 0,
    vwap REAL DEFAULT 0,
    transactions INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (underlying_symbol) REFERENCES symbols(symbol)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_option_history_ticker_date
    ON option_history(option_ticker, bar_date);
CREATE INDEX IF NOT EXISTS idx_option_history_underlying
    ON option_history(underlying_symbol);
CREATE INDEX IF NOT EXISTS idx_option_history_bar_date
    ON option_history(bar_date);
CREATE INDEX IF NOT EXISTS idx_option_history_expiration
    ON option_history(expiration_date);

CREATE TABLE IF NOT EXISTS index_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    index_ticker TEXT NOT NULL,
    bar_date DATE NOT NULL,
    open REAL NOT NULL,
    high REAL NOT NULL,
    low REAL NOT NULL,
    close REAL NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_index_history_ticker_date
    ON index_history(index_ticker, bar_date);
CREATE INDEX IF NOT EXISTS idx_index_history_bar_date
    ON index_history(bar_date);

INSERT OR IGNORE INTO schema_migrations (version)
VALUES ('20260602000001_add_history_tables');
