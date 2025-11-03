# Wheeler V2 Data Model Specification

This document defines the enhanced data model for Wheeler V2, a financial trading portfolio system with multi-account support and comprehensive transaction tracking.

## Entity Relationship Diagram

```
┌──────────────────────────────────┐
│          ACCOUNT                 │
├──────────────────────────────────┤
│ PK  id                  INTEGER  │
│     name                TEXT     │
│     account_type        TEXT     │  -- 'CASH', 'MARGIN', 'IRA'
│     balance             REAL     │
│     cash_balance        REAL     │
│     created_at          DATETIME │
│     updated_at          DATETIME │
└──────────────────────────────────┘
                  │
                  │ 1:N
                  ▼
┌──────────────────────────────────┐
│         TRANSACTION              │
├──────────────────────────────────┤
│ PK  id                  INTEGER  │
│ FK  account_id          INTEGER  │──┐
│     asset_type          TEXT     │  │  -- 'STOCK', 'OPTION', 'TREASURY', 'DIVIDEND'
│     asset_id            INTEGER  │  │  -- Polymorphic FK to asset tables
│     trade_type          TEXT     │  │  -- 'BUY_TO_OPEN', 'SELL_TO_CLOSE', 
│     transaction_date    DATE     │  │  --  'SELL_TO_OPEN', 'BUY_TO_CLOSE',
│     quantity            INTEGER  │  │  --  'ASSIGNED', 'EXPIRED', 'RECEIVE', 'INTEREST'
│     price               REAL     │  │
│     total_amount        REAL     │  │  -- quantity * price (signed: + credit, - debit)
│     commission          REAL     │  │
│     net_amount          REAL     │  │  -- total_amount - commission
│     notes               TEXT     │  │
│     created_at          DATETIME │  │
│     updated_at          DATETIME │  │
└──────────────────────────────────┘  │
         │                             │
         │                             │
         ▼ (polymorphic)               │
    ┌────────────────┐                │
    │  asset_type +  │                │
    │  asset_id      │                │
    └────────────────┘                │
         │                             │
    ┌────┴────┬────────┬───────┐      │
    │         │        │       │      │
    ▼         ▼        ▼       ▼      │
┌─────────┐ ┌──────┐ ┌────┐ ┌─────┐  │
│ STOCK   │ │OPTION│ │TREA│ │ DIV │  │
└─────────┘ └──────┘ └────┘ └─────┘  │
                                      │
┌──────────────────────────────────┐  │
│            SYMBOL                │  │
├──────────────────────────────────┤  │
│ PK  symbol              TEXT     │  │
│     name                TEXT     │  │
│     price               REAL     │  │
│     dividend_yield      REAL     │  │
│     ex_dividend_date    DATE     │  │
│     pe_ratio            REAL     │  │
│     sector              TEXT     │  │
│     created_at          DATETIME │  │
│     updated_at          DATETIME │  │
└──────────────────────────────────┘  │
         │                             │
         │ 1:N                         │
         ▼                             │
┌──────────────────────────────────┐  │
│            STOCK                 │  │
├──────────────────────────────────┤  │
│ PK  id                  INTEGER  │  │
│ FK  account_id          INTEGER  │◄─┤
│ FK  symbol              TEXT     │  │
│ FK  tx_id               INTEGER  │  │  -- References opening transaction
│     shares              INTEGER  │  │
│     cost_basis          REAL     │  │  -- Total cost (shares * price + commission)
│     avg_price           REAL     │  │  -- Average price per share
│     opened_date         DATE     │  │
│     closed_date         DATE     │  │
│     status              TEXT     │  │  -- 'OPEN', 'CLOSED'
│     notes               TEXT     │  │
│     created_at          DATETIME │  │
│     updated_at          DATETIME │  │
└──────────────────────────────────┘  │
                                      │
┌──────────────────────────────────┐  │
│            OPTION                │  │
├──────────────────────────────────┤  │
│ PK  id                  INTEGER  │  │
│ FK  account_id          INTEGER  │◄─┤
│ FK  symbol              TEXT     │  │
│ FK  tx_id               INTEGER  │  │  -- References opening transaction
│     option_type         TEXT     │  │  -- 'PUT', 'CALL'
│     strike              REAL     │  │
│     expiration          DATE     │  │
│     contracts           INTEGER  │  │
│     premium_received    REAL     │  │  -- If sold to open
│     premium_paid        REAL     │  │  -- If bought to close
│     opened_date         DATE     │  │
│     closed_date         DATE     │  │
│     status              TEXT     │  │  -- 'OPEN', 'CLOSED', 'ASSIGNED', 'EXPIRED'
│     assignment_tx_id    INTEGER  │  │  -- References assignment transaction if applicable
│     current_price       REAL     │  │
│     notes               TEXT     │  │
│     created_at          DATETIME │  │
│     updated_at          DATETIME │  │
└──────────────────────────────────┘  │
                                      │
┌──────────────────────────────────┐  │
│          TREASURY                │  │
├──────────────────────────────────┤  │
│ PK  id                  INTEGER  │  │
│ FK  account_id          INTEGER  │◄─┤
│ UK  cuspid              TEXT     │  │
│ FK  tx_id               INTEGER  │  │  -- References purchase transaction
│     amount              REAL     │  │  -- Face value
│     yield               REAL     │  │
│     buy_price           REAL     │  │  -- Purchase price
│     current_value       REAL     │  │
│     purchased_date      DATE     │  │
│     maturity_date       DATE     │  │
│     sold_date           DATE     │  │
│     status              TEXT     │  │  -- 'HELD', 'SOLD', 'MATURED'
│     notes               TEXT     │  │
│     created_at          DATETIME │  │
│     updated_at          DATETIME │  │
└──────────────────────────────────┘  │
                                      │
┌──────────────────────────────────┐  │
│          DIVIDEND                │  │
├──────────────────────────────────┤  │
│ PK  id                  INTEGER  │  │
│ FK  account_id          INTEGER  │◄─┘
│ FK  symbol              TEXT     │
│ FK  stock_id            INTEGER  │  -- References stock position
│ FK  tx_id               INTEGER  │  -- References dividend receipt transaction
│     payment_date        DATE     │
│     ex_dividend_date    DATE     │
│     shares              INTEGER  │  -- Shares held at payment
│     amount_per_share    REAL     │
│     total_amount        REAL     │  -- shares * amount_per_share
│     dividend_type       TEXT     │  -- 'CASH', 'QUALIFIED', 'SPECIAL'
│     notes               TEXT     │
│     created_at          DATETIME │
│     updated_at          DATETIME │
└──────────────────────────────────┘
```

## Relationships

```
Account 1 ────── N Transaction    (account_id FK)
Account 1 ────── N Stock           (account_id FK)
Account 1 ────── N Option          (account_id FK)
Account 1 ────── N Treasury        (account_id FK)
Account 1 ────── N Dividend        (account_id FK)

Symbol  1 ────── N Stock           (symbol FK)
Symbol  1 ────── N Option          (symbol FK)
Symbol  1 ────── N Dividend        (symbol FK)

Transaction 1 ── 0..1 Stock        (tx_id FK - opening transaction)
Transaction 1 ── 0..1 Option       (tx_id FK - opening transaction)
Transaction 1 ── 0..1 Treasury     (tx_id FK - purchase transaction)
Transaction 1 ── 0..1 Dividend     (tx_id FK - receipt transaction)

Stock       1 ── N Dividend        (stock_id FK)
```

## Transaction Flow Examples

### 0. Open Account
```
Account:     id=1, name='Trading Account', account_type='CASH', balance=$0, cash_balance=$0
```

### 1. Deposit Cash
```
Transaction: trade_type='RECEIVE', asset_type='CASH', net_amount=+$10,000
Account:     cash_balance=$10,000, balance=$10,000
```

### 2. Sell Put to Open
```
Transaction: trade_type='SELL_TO_OPEN', asset_type='OPTION', asset_id=option.id, net_amount=+$150
Option:      status='OPEN', premium_received=$150
Account:     cash_balance=$10,150, balance=$10,150
```

### 3. Put Assignment
```
Transaction: trade_type='ASSIGNED', asset_type='OPTION', asset_id=option.id, net_amount=$0
Option:      status='ASSIGNED', assignment_tx_id=stock_tx.id
Transaction: trade_type='BUY_TO_OPEN', asset_type='STOCK', asset_id=stock.id, net_amount=-$5,000
Stock:       status='OPEN', shares=100, cost_basis=$5,000
Account:     cash_balance=$5,150, balance=$10,150 (Cash $5,150 + Stock $5,000)
```

### 4. Sell Call to Open
```
Transaction: trade_type='SELL_TO_OPEN', asset_type='OPTION', asset_id=option.id, net_amount=+$200
Option:      status='OPEN', premium_received=$200
Account:     cash_balance=$5,350, balance=$10,350
```

### 5. Dividend Received
```
Transaction: trade_type='RECEIVE', asset_type='DIVIDEND', asset_id=dividend.id, net_amount=+$50
Dividend:    total_amount=$50
Account:     cash_balance=$5,400, balance=$10,400
```

### 6. Account Value Calculation
```
Cash Balance:        account.cash_balance
Stock Value:         SUM(stock.shares * current_price) WHERE account_id = X
Option Value:        SUM(option.current_value) WHERE account_id = X AND status = 'OPEN'
Treasury Value:      SUM(treasury.current_value) WHERE account_id = X AND status = 'HELD'

Total Account Value: Cash + Stock + Option + Treasury = account.balance

Alternatively (Transaction-Centric):
SUM(transactions.net_amount WHERE account_id = X) = Current Account Value
```

## Database Indexes

```sql
CREATE INDEX idx_transaction_account_id      ON transaction(account_id);
CREATE INDEX idx_transaction_asset           ON transaction(asset_type, asset_id);
CREATE INDEX idx_transaction_date            ON transaction(transaction_date);
CREATE INDEX idx_stock_account_symbol        ON stock(account_id, symbol);
CREATE INDEX idx_stock_status                ON stock(status);
CREATE INDEX idx_option_account_symbol       ON option(account_id, symbol);
CREATE INDEX idx_option_expiration           ON option(expiration);
CREATE INDEX idx_option_status               ON option(status);
CREATE INDEX idx_treasury_account            ON treasury(account_id);
CREATE INDEX idx_treasury_maturity           ON treasury(maturity_date);
CREATE INDEX idx_dividend_account_symbol     ON dividend(account_id, symbol);
CREATE INDEX idx_dividend_payment_date       ON dividend(payment_date);
```

## Key Design Principles

### 1. Multi-Account Support
All asset tables (Stock, Option, Treasury, Dividend) reference an Account via `account_id` foreign key. This enables tracking multiple portfolios or accounts within a single database. Each account maintains its own cash balance via the `cash_balance` attribute.

### 2. Transaction-Centric Accounting
The Transaction table serves as the source of truth for all financial movements. Account value is calculated as the sum of all transaction net amounts for that account.

### 3. Polymorphic Asset References
Transactions use `asset_type` (enum) and `asset_id` (integer) to reference different asset tables. This flexible design supports diverse transaction types while maintaining referential integrity.

### 4. Bidirectional Transaction-Asset Linking
- Transactions reference assets via `asset_type` + `asset_id`
- Assets reference their opening transaction via `tx_id`
- This dual linking enables both transaction→asset and asset→transaction queries

### 5. Symbol Independence
The Symbol table stands independent without account_id, representing universal market data. Multiple accounts can hold the same symbol.

### 6. Status Tracking
Assets maintain explicit status fields ('OPEN', 'CLOSED', 'ASSIGNED', 'EXPIRED', etc.) enabling lifecycle management and filtering.

### 7. Audit Trail
All tables include `created_at` and `updated_at` timestamps for comprehensive audit trails.

## Trade Type Enumeration

```
RECEIVE              - Cash deposit or dividend/interest receipt
BUY_TO_OPEN          - Open long stock position
SELL_TO_CLOSE        - Close long stock position
SELL_TO_OPEN         - Sell option (put/call)
BUY_TO_CLOSE         - Buy to close option position
ASSIGNED             - Option assignment
EXPIRED              - Option expiration
INTEREST             - Treasury interest payment
WITHDRAW             - Cash withdrawal
```

## Asset Type Enumeration

```
CASH                 - Cash deposit/withdrawal
STOCK                - Stock position
OPTION               - Option contract
TREASURY             - Treasury security
DIVIDEND             - Dividend payment
```

## Option Type Enumeration

```
PUT                  - Put option
CALL                 - Call option
```

## Account Type Enumeration

```
CASH                 - Cash account
MARGIN               - Margin account
IRA                  - Individual Retirement Account
```

## Dividend Type Enumeration

```
CASH                 - Regular cash dividend
QUALIFIED            - Qualified dividend (tax treatment)
SPECIAL              - Special/one-time dividend
```

## Status Enumeration

### Stock Status
```
OPEN                 - Active position
CLOSED               - Position closed
```

### Option Status
```
OPEN                 - Active option
CLOSED               - Closed via buy-to-close
ASSIGNED             - Option assigned
EXPIRED              - Expired worthless
```

### Treasury Status
```
HELD                 - Currently held
SOLD                 - Sold before maturity
MATURED              - Reached maturity
```

## Migration Plan

### File-Based Migration System

Wheeler V2 uses a file-based migration system for database schema evolution. All migrations are stored in `internal/database/migrations/` and executed in lexicographical order.

**Migration Guidelines:**
- All migrations MUST be idempotent (safe to run multiple times)
- Use `CREATE TABLE IF NOT EXISTS`, `ALTER TABLE IF NOT EXISTS` patterns
- Check for column/index existence before adding
- Migrations are named with timestamp prefix: `YYYYMMDDHHMMSS_description.sql`
- Once applied to production, migrations should never be modified

**Migration Execution:**
```go
// internal/database/db.go
func (db *DB) runMigrations() error {
    migrationFiles, err := migrationFS.ReadDir("migrations")
    if err != nil {
        return fmt.Errorf("failed to read migrations directory: %w", err)
    }

    for _, file := range migrationFiles {
        if !strings.HasSuffix(file.Name(), ".sql") {
            continue
        }

        content, err := migrationFS.ReadFile(filepath.Join("migrations", file.Name()))
        if err != nil {
            return fmt.Errorf("failed to read migration %s: %w", file.Name(), err)
        }

        if _, err := db.Exec(string(content)); err != nil {
            return fmt.Errorf("failed to execute migration %s: %w", file.Name(), err)
        }
    }

    return nil
}
```

### V2 Migration Files

#### `internal/database/migrations/20250103000001_create_accounts.sql`
```sql
-- Create accounts table for multi-account support
CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    account_type TEXT NOT NULL CHECK (account_type IN ('CASH', 'MARGIN', 'IRA')),
    balance REAL DEFAULT 0.0,
    cash_balance REAL DEFAULT 0.0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create default account for existing data
INSERT OR IGNORE INTO accounts (id, name, account_type, balance, cash_balance)
VALUES (1, 'Default Account', 'CASH', 0.0, 0.0);
```

#### `internal/database/migrations/20250103000002_create_transactions.sql`
```sql
-- Create transactions table for transaction-centric accounting
CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL,
    asset_type TEXT NOT NULL CHECK (asset_type IN ('CASH', 'STOCK', 'OPTION', 'TREASURY', 'DIVIDEND')),
    asset_id INTEGER,
    trade_type TEXT NOT NULL CHECK (trade_type IN ('RECEIVE', 'BUY_TO_OPEN', 'SELL_TO_CLOSE', 'SELL_TO_OPEN', 'BUY_TO_CLOSE', 'ASSIGNED', 'EXPIRED', 'INTEREST', 'WITHDRAW')),
    transaction_date DATE NOT NULL,
    quantity INTEGER,
    price REAL,
    total_amount REAL NOT NULL,
    commission REAL DEFAULT 0.0,
    net_amount REAL NOT NULL,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (account_id) REFERENCES accounts(id)
);

CREATE INDEX IF NOT EXISTS idx_transaction_account_id ON transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_transaction_asset ON transactions(asset_type, asset_id);
CREATE INDEX IF NOT EXISTS idx_transaction_date ON transactions(transaction_date);
```

#### `internal/database/migrations/20250103000003_add_account_id_to_assets.sql`
```sql
-- Add account_id to long_positions (stock)
ALTER TABLE long_positions ADD COLUMN account_id INTEGER DEFAULT 1;
ALTER TABLE long_positions ADD COLUMN tx_id INTEGER;
ALTER TABLE long_positions ADD COLUMN cost_basis REAL;
ALTER TABLE long_positions ADD COLUMN avg_price REAL;
ALTER TABLE long_positions ADD COLUMN status TEXT DEFAULT 'OPEN';
ALTER TABLE long_positions ADD COLUMN notes TEXT;

-- Backfill cost_basis and avg_price
UPDATE long_positions SET cost_basis = shares * buy_price WHERE cost_basis IS NULL;
UPDATE long_positions SET avg_price = buy_price WHERE avg_price IS NULL;
UPDATE long_positions SET status = CASE WHEN closed IS NULL THEN 'OPEN' ELSE 'CLOSED' END WHERE status IS NULL;

-- Add account_id to options
ALTER TABLE options ADD COLUMN account_id INTEGER DEFAULT 1;
ALTER TABLE options ADD COLUMN tx_id INTEGER;
ALTER TABLE options ADD COLUMN premium_received REAL;
ALTER TABLE options ADD COLUMN premium_paid REAL;
ALTER TABLE options ADD COLUMN status TEXT DEFAULT 'OPEN';
ALTER TABLE options ADD COLUMN assignment_tx_id INTEGER;
ALTER TABLE options ADD COLUMN notes TEXT;

-- Backfill premium fields based on option type
UPDATE options SET premium_received = premium WHERE type IN ('Put', 'Call') AND premium_received IS NULL;
UPDATE options SET status = CASE WHEN closed IS NULL THEN 'OPEN' ELSE 'CLOSED' END WHERE status IS NULL;

-- Add account_id to dividends
ALTER TABLE dividends ADD COLUMN account_id INTEGER DEFAULT 1;
ALTER TABLE dividends ADD COLUMN stock_id INTEGER;
ALTER TABLE dividends ADD COLUMN tx_id INTEGER;
ALTER TABLE dividends ADD COLUMN ex_dividend_date DATE;
ALTER TABLE dividends ADD COLUMN shares INTEGER;
ALTER TABLE dividends ADD COLUMN amount_per_share REAL;
ALTER TABLE dividends ADD COLUMN total_amount REAL;
ALTER TABLE dividends ADD COLUMN dividend_type TEXT DEFAULT 'CASH';
ALTER TABLE dividends ADD COLUMN notes TEXT;
ALTER TABLE dividends ADD COLUMN updated_at DATETIME DEFAULT CURRENT_TIMESTAMP;

-- Backfill total_amount from amount
UPDATE dividends SET total_amount = amount WHERE total_amount IS NULL;

-- Add account_id to treasuries
ALTER TABLE treasuries ADD COLUMN id INTEGER;
ALTER TABLE treasuries ADD COLUMN account_id INTEGER DEFAULT 1;
ALTER TABLE treasuries ADD COLUMN tx_id INTEGER;
ALTER TABLE treasuries ADD COLUMN status TEXT DEFAULT 'HELD';
ALTER TABLE treasuries ADD COLUMN notes TEXT;

-- Backfill treasury status
UPDATE treasuries SET status = CASE WHEN exit_price IS NOT NULL THEN 'SOLD' ELSE 'HELD' END WHERE status IS NULL;
```

#### `internal/database/migrations/20250103000004_add_symbol_enhancements.sql`
```sql
-- Add symbol table enhancements
ALTER TABLE symbols ADD COLUMN name TEXT;
ALTER TABLE symbols ADD COLUMN dividend_yield REAL;
ALTER TABLE symbols ADD COLUMN sector TEXT;

-- Backfill dividend_yield from existing dividend field
UPDATE symbols SET dividend_yield = dividend WHERE dividend_yield IS NULL;
```

#### `internal/database/migrations/20250103000005_create_indexes.sql`
```sql
-- Create performance indexes for V2 schema
CREATE INDEX IF NOT EXISTS idx_long_positions_account_id ON long_positions(account_id);
CREATE INDEX IF NOT EXISTS idx_long_positions_status ON long_positions(status);
CREATE INDEX IF NOT EXISTS idx_options_account_id ON options(account_id);
CREATE INDEX IF NOT EXISTS idx_options_status ON options(status);
CREATE INDEX IF NOT EXISTS idx_dividends_account_id ON dividends(account_id);
CREATE INDEX IF NOT EXISTS idx_treasuries_account_id ON treasuries(account_id);
```

### Migration from V1 to V2

**Step 1: Run Migrations**
```bash
go run main.go  # Migrations run automatically on startup
```

**Step 2: Generate Opening Transactions**

Create opening transactions for all existing assets:

```sql
-- Generate opening transactions for existing long positions
INSERT INTO transactions (account_id, asset_type, asset_id, trade_type, transaction_date, quantity, price, total_amount, commission, net_amount)
SELECT 
    COALESCE(account_id, 1),
    'STOCK',
    id,
    'BUY_TO_OPEN',
    opened,
    shares,
    buy_price,
    -(shares * buy_price),
    0.0,
    -(shares * buy_price)
FROM long_positions
WHERE tx_id IS NULL;

-- Update long_positions with transaction references
UPDATE long_positions
SET tx_id = (
    SELECT id FROM transactions 
    WHERE asset_type = 'STOCK' 
    AND asset_id = long_positions.id 
    LIMIT 1
)
WHERE tx_id IS NULL;

-- Generate opening transactions for existing options
INSERT INTO transactions (account_id, asset_type, asset_id, trade_type, transaction_date, quantity, price, total_amount, commission, net_amount)
SELECT 
    COALESCE(account_id, 1),
    'OPTION',
    id,
    'SELL_TO_OPEN',
    opened,
    contracts,
    premium,
    premium * contracts * 100,
    COALESCE(commission, 0.0),
    (premium * contracts * 100) - COALESCE(commission, 0.0)
FROM options
WHERE tx_id IS NULL;

-- Update options with transaction references
UPDATE options
SET tx_id = (
    SELECT id FROM transactions 
    WHERE asset_type = 'OPTION' 
    AND asset_id = options.id 
    LIMIT 1
)
WHERE tx_id IS NULL;

-- Generate dividend transactions
INSERT INTO transactions (account_id, asset_type, asset_id, trade_type, transaction_date, quantity, price, total_amount, commission, net_amount)
SELECT 
    COALESCE(account_id, 1),
    'DIVIDEND',
    id,
    'RECEIVE',
    received,
    NULL,
    NULL,
    amount,
    0.0,
    amount
FROM dividends
WHERE tx_id IS NULL;

-- Update dividends with transaction references
UPDATE dividends
SET tx_id = (
    SELECT id FROM transactions 
    WHERE asset_type = 'DIVIDEND' 
    AND asset_id = dividends.id 
    LIMIT 1
)
WHERE tx_id IS NULL;
```

**Step 3: Update Account Balances**

```sql
-- Calculate and update account balance from transactions
UPDATE accounts
SET balance = (
    SELECT COALESCE(SUM(net_amount), 0)
    FROM transactions
    WHERE transactions.account_id = accounts.id
);

-- Calculate cash balance (transactions minus asset values)
UPDATE accounts
SET cash_balance = (
    SELECT 
        COALESCE(SUM(net_amount), 0) +
        COALESCE((SELECT SUM(shares * buy_price) FROM long_positions WHERE account_id = accounts.id AND status = 'OPEN'), 0)
    FROM transactions
    WHERE transactions.account_id = accounts.id
);
```

### Cleanup: Remove Deprecated Migration Code

Remove the inline migration code from `internal/database/db.go`:

```go
// REMOVE THIS SECTION:
func (db *DB) runMigrations() error {
    var hasCurrentPrice bool
    err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('options') WHERE name = 'current_price'").Scan(&hasCurrentPrice)
    if err != nil {
        return fmt.Errorf("failed to check for current_price column: %w", err)
    }

    if !hasCurrentPrice {
        _, err := db.Exec("ALTER TABLE options ADD COLUMN current_price REAL")
        if err != nil {
            return fmt.Errorf("failed to add current_price column: %w", err)
        }
    }

    return nil
}
```

Replace with file-based migration system shown above.

### Rollback Strategy

**Database Backup Before Migration:**
```bash
cp ./data/wheeler.db ./data/wheeler_backup_$(date +%Y%m%d_%H%M%S).db
```

**Rollback Process:**
1. Stop application
2. Restore backup: `cp ./data/wheeler_backup_TIMESTAMP.db ./data/wheeler.db`
3. Restart application on previous version

**Note:** V2 schema is not backwards compatible with V1 application code. Full migration to V2 requires updating all application queries and handlers.
