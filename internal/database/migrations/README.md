# Wheeler Database Migrations

This directory contains **file-based database migrations** for Wheeler V1 schema evolution.

## Migration System Design

- **Automatic Execution**: Migrations run automatically on application startup
- **Idempotent**: All migrations use `IF NOT EXISTS` patterns - safe to run multiple times
- **Ordered**: Migrations execute in lexicographical order by filename
- **Tracked**: `schema_migrations` table tracks which migrations have been applied
- **Zero Downtime**: All migrations are additive and backward-compatible

## File Naming Convention

```
YYYYMMDDHHMMSS_description.sql
```

Example: `20250111000001_baseline_v1_schema.sql`

## Migration Guidelines

### DO:
- ✅ Use `CREATE TABLE IF NOT EXISTS`
- ✅ Check for column existence before `ALTER TABLE ADD COLUMN`
- ✅ Use `CREATE INDEX IF NOT EXISTS`
- ✅ Insert default data with `INSERT OR IGNORE`
- ✅ Test migrations on a database copy first
- ✅ Keep migrations small and focused

### DON'T:
- ❌ Modify existing migration files after they're merged
- ❌ Use `DROP TABLE` or `DROP COLUMN` (breaks backward compatibility)
- ❌ Change existing column types (create new columns instead)
- ❌ Remove indexes that existing queries depend on

## Creating a New Migration

1. **Create file with timestamp**:
   ```bash
   touch migrations/$(date +%Y%m%d%H%M%S)_add_feature_name.sql
   ```

2. **Write idempotent SQL**:
   ```sql
   -- Add new column example
   ALTER TABLE options ADD COLUMN IF NOT EXISTS new_field REAL;
   
   -- Add new index example
   CREATE INDEX IF NOT EXISTS idx_options_new_field ON options(new_field);
   
   -- Record migration
   INSERT OR IGNORE INTO schema_migrations (version) 
   VALUES ('20250111120000_add_feature_name');
   ```

3. **Test locally**:
   ```bash
   # Backup your database first!
   cp ./data/wheeler.db ./data/wheeler_backup.db
   
   # Run application (migrations auto-apply)
   go run main.go
   ```

4. **Verify migration applied**:
   ```sql
   SELECT * FROM schema_migrations ORDER BY applied_at DESC LIMIT 5;
   ```

## Current Migrations

| Version | Description | Applied |
|---------|-------------|---------|
| `20250111000001` | Baseline V1 schema | 2025-01-11 |

## Rollback Strategy

**Wheeler migrations are additive only** - we don't support automatic rollback.

If a migration causes issues:

1. **Stop the application**
2. **Restore from backup**: 
   ```bash
   cp ./data/wheeler_backup_TIMESTAMP.db ./data/wheeler.db
   ```
3. **Fix the migration file**
4. **Restart application**

## Testing Migrations

Before merging a migration:

1. Backup production database
2. Test migration on backup copy
3. Verify queries still work
4. Verify application still runs
5. Check migration is idempotent (run twice)

## Checking Column Existence (Pattern)

For SQLite, check columns before adding:

```sql
-- Pattern: Check if column exists before adding
SELECT COUNT(*) FROM pragma_table_info('table_name') WHERE name = 'column_name';
```

Better approach - just try to add it and handle gracefully:

```sql
-- SQLite doesn't support ALTER TABLE IF NOT EXISTS, so we use a workaround
-- The migration system already tracks applied migrations via schema_migrations table
-- So we can safely add columns knowing the migration won't run twice

ALTER TABLE options ADD COLUMN new_field REAL;
```

## Migration Execution Flow

```
1. Application starts
2. db.InitSchema() called
3. schema.sql executed (creates base tables)
4. db.runMigrations() called
5. Reads migrations/*.sql in order
6. For each migration:
   - Check if already applied (schema_migrations table)
   - If not applied: execute SQL
   - Migration records itself in schema_migrations
7. Application ready
```

## Future Improvements

- [ ] Add migration rollback SQL (for manual rollback)
- [ ] Add migration validation tool
- [ ] Add migration dry-run mode
- [ ] Add schema diff tool
