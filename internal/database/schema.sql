CREATE TABLE IF NOT EXISTS symbols (
    symbol TEXT PRIMARY KEY,
    price REAL DEFAULT 0.0,
    dividend REAL DEFAULT 0.0,
    ex_dividend_date DATE,
    pe_ratio REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS long_positions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    opened DATE NOT NULL,
    closed DATE,
    shares INTEGER NOT NULL,
    buy_price REAL NOT NULL,
    exit_price REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (symbol) REFERENCES symbols(symbol)
);

-- Options table: tracks options trades (Puts and Calls)
-- Note: commission field stores TOTAL commission paid
--   - For open/expired positions: total = per_contract_rate * contracts
--   - For closed positions (buy-to-close): total = per_contract_rate * contracts * 2
CREATE TABLE IF NOT EXISTS options (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('Put', 'Call')),
    opened DATE NOT NULL,
    closed DATE,
    strike REAL NOT NULL,
    expiration DATE NOT NULL,
    premium REAL NOT NULL,
    contracts INTEGER NOT NULL,
    exit_price REAL,
    commission REAL DEFAULT 0.0, -- Total commission (not per-contract)
    current_price REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (symbol) REFERENCES symbols(symbol)
);

CREATE TABLE IF NOT EXISTS dividends (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    received DATE NOT NULL,
    amount REAL NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (symbol) REFERENCES symbols(symbol)
);

CREATE TABLE IF NOT EXISTS treasuries (
    cuspid TEXT PRIMARY KEY,
    purchased DATE NOT NULL,
    maturity DATE NOT NULL,
    amount REAL NOT NULL,
    yield REAL NOT NULL,
    buy_price REAL NOT NULL,
    current_value REAL,
    exit_price REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS settings (
    name TEXT PRIMARY KEY,
    value TEXT,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    type TEXT NOT NULL CHECK (type IN ('treasury_value', 'long_value', 'long_count', 'put_exposure', 'open_call_premium', 'open_call_count', 'open_put_premium', 'open_put_count', 'total_value')),
    value REAL NOT NULL
);

-- Insert default POLYGON_API_KEY setting
INSERT OR IGNORE INTO settings (name, value, description)
VALUES ('POLYGON_API_KEY', '', 'API key for Polygon.io stock market data integration');

-- Insert default OPTION_COMMISSION_PER_CONTRACT setting
INSERT OR IGNORE INTO settings (name, value, description)
VALUES ('OPTION_COMMISSION_PER_CONTRACT', '0.65', 'Default commission per options contract');

-- Indexes for performance
-- Note: Primary key columns automatically have indexes, so we don't need explicit indexes for:
-- symbols(symbol), treasuries(cuspid), settings(name) - they have PRIMARY KEY
-- Foreign key and commonly queried columns:
CREATE INDEX IF NOT EXISTS idx_long_positions_symbol ON long_positions(symbol);
CREATE INDEX IF NOT EXISTS idx_long_positions_opened ON long_positions(opened);
CREATE INDEX IF NOT EXISTS idx_options_symbol ON options(symbol);
CREATE INDEX IF NOT EXISTS idx_options_expiration ON options(expiration);
CREATE INDEX IF NOT EXISTS idx_options_type ON options(type);
CREATE INDEX IF NOT EXISTS idx_dividends_symbol ON dividends(symbol);
CREATE INDEX IF NOT EXISTS idx_dividends_received ON dividends(received);
CREATE INDEX IF NOT EXISTS idx_treasuries_maturity ON treasuries(maturity);
CREATE INDEX IF NOT EXISTS idx_treasuries_purchased ON treasuries(purchased);
CREATE INDEX IF NOT EXISTS idx_metrics_created ON metrics(created);
CREATE INDEX IF NOT EXISTS idx_metrics_type ON metrics(type);

-- Unique constraints to prevent duplicate business records
-- (These replace the compound primary keys while allowing easier HTTP CRUD with integer IDs)
CREATE UNIQUE INDEX IF NOT EXISTS idx_options_unique ON options(symbol, type, opened, strike, expiration, premium, contracts);
CREATE UNIQUE INDEX IF NOT EXISTS idx_dividends_unique ON dividends(symbol, received, amount);
