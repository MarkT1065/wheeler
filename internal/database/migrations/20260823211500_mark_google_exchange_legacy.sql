-- Markets are now selected per symbol; retain this setting only for legacy
-- compatibility and make that clear to existing databases.
UPDATE settings
SET description = 'Legacy default exchange; new symbols select their market individually',
    updated_at = CURRENT_TIMESTAMP
WHERE name = 'GOOGLE_FINANCE_EXCHANGE';

INSERT OR IGNORE INTO schema_migrations (version)
VALUES ('20260823211500_mark_google_exchange_legacy');
