-- Add option position direction without changing existing option records.
ALTER TABLE options ADD COLUMN direction TEXT NOT NULL DEFAULT 'Short'
    CHECK (direction IN ('Short', 'Long'));

CREATE INDEX IF NOT EXISTS idx_options_direction ON options(direction);

DROP INDEX IF EXISTS idx_options_unique;
CREATE UNIQUE INDEX IF NOT EXISTS idx_options_unique
    ON options(symbol, type, direction, opened, strike, expiration, premium, contracts);

INSERT OR IGNORE INTO schema_migrations (version)
VALUES ('20260824000000_add_option_direction');
