-- Add default provider settings to databases created before provider support.
INSERT OR IGNORE INTO settings (name, value, description)
VALUES ('DATA_PROVIDER_TYPE', 'polygon', 'Active market data provider: polygon, google_finance (google/googlefinance), or yfinance (yahoo/yahoo_finance)');

INSERT OR IGNORE INTO settings (name, value, description)
VALUES ('GOOGLE_FINANCE_EXCHANGE', 'NASDAQ', 'Default exchange for Google Finance symbols');

INSERT OR IGNORE INTO schema_migrations (version)
VALUES ('20260822220000_add_provider_settings');
