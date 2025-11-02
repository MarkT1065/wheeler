# Wheeler V2 Specification Differences

Comparison between Wheeler V2 proposed data model and current schema implementation.

## Major Architectural Differences

### V2 Spec (Transaction-Centric)
- **Account table** - Multi-account support
- **Transaction table** - Source of truth for all financial movements
- **Polymorphic asset references** - Transactions link to assets via asset_type + asset_id
- **Symbol uses integer PK** - `id` as primary key with unique constraint on `ticker`
- **All assets link back to transactions** - Bidirectional relationship via `tx_id`
- **Account value calculation** - Sum of transaction net_amount per account

### Current Schema (Asset-Centric)
- **No Account table** - Single portfolio assumption
- **No Transaction table** - Implicit transaction history via asset records
- **Direct asset tracking** - Assets are standalone entities
- **Symbol uses TEXT PK** - `symbol` ticker as primary key
- **Assets are standalone** - No explicit transaction linkage
- **Asset value calculation** - Sum of current asset positions

## Table-by-Table Comparison

### SYMBOL

**V2 Spec:**
```sql
PK  id                  INTEGER
UK  ticker              TEXT
    name                TEXT
    price               REAL
    dividend_yield      REAL
    ex_dividend_date    DATE
    pe_ratio            REAL
    sector              TEXT
    created_at          DATETIME
    updated_at          DATETIME
```

**Current Schema:**
```sql
PK  symbol              TEXT
    price               REAL
    dividend            REAL
    ex_dividend_date    DATE
    pe_ratio            REAL
    created_at          DATETIME
    updated_at          DATETIME
```

**Differences:**
- V2 uses INTEGER PK with unique ticker constraint
- Current uses TEXT PK (ticker symbol)
- V2 adds `name` and `sector` fields
- V2 uses `dividend_yield`, current uses `dividend`

---

### STOCK (long_positions in current)

**V2 Spec:**
```sql
PK  id                  INTEGER
FK  account_id          INTEGER
FK  symbol_id           INTEGER
FK  tx_id               INTEGER    -- References opening transaction
    shares              INTEGER
    cost_basis          REAL       -- Total cost
    avg_price           REAL       -- Average price per share
    opened_date         DATE
    closed_date         DATE
    status              TEXT       -- 'OPEN', 'CLOSED'
    notes               TEXT
    created_at          DATETIME
    updated_at          DATETIME
```

**Current Schema:**
```sql
PK  id                  INTEGER
FK  symbol              TEXT
    opened              DATE
    closed              DATE
    shares              INTEGER
    buy_price           REAL
    exit_price          REAL
    created_at          DATETIME
    updated_at          DATETIME
```

**Differences:**
- V2 adds `account_id` FK for multi-account support
- V2 uses `symbol_id` INTEGER FK vs current TEXT FK
- V2 adds `tx_id` linking to opening transaction
- V2 includes `cost_basis` and `avg_price` for better cost tracking
- V2 adds explicit `status` field ('OPEN', 'CLOSED')
- V2 adds `notes` field
- Current has `exit_price`, V2 relies on closing transaction

---

### OPTION

**V2 Spec:**
```sql
PK  id                  INTEGER
FK  account_id          INTEGER
FK  symbol_id           INTEGER
FK  tx_id               INTEGER    -- References opening transaction
    option_type         TEXT       -- 'PUT', 'CALL'
    strike              REAL
    expiration          DATE
    contracts           INTEGER
    premium_received    REAL       -- If sold to open
    premium_paid        REAL       -- If bought to close
    opened_date         DATE
    closed_date         DATE
    status              TEXT       -- 'OPEN', 'CLOSED', 'ASSIGNED', 'EXPIRED'
    assignment_tx_id    INTEGER    -- References assignment transaction
    current_price       REAL
    notes               TEXT
    created_at          DATETIME
    updated_at          DATETIME
```

**Current Schema:**
```sql
PK  id                  INTEGER
FK  symbol              TEXT
    type                TEXT       -- 'Put', 'Call'
    opened              DATE
    closed              DATE
    strike              REAL
    expiration          DATE
    premium             REAL
    contracts           INTEGER
    exit_price          REAL
    commission          REAL
    current_price       REAL
    created_at          DATETIME
    updated_at          DATETIME
```

**Differences:**
- V2 adds `account_id` FK for multi-account support
- V2 uses `symbol_id` INTEGER FK vs current TEXT FK
- V2 adds `tx_id` linking to opening transaction
- V2 separates `premium_received` and `premium_paid` vs single `premium`
- V2 adds explicit `status` field with assignment/expiration tracking
- V2 adds `assignment_tx_id` to track option assignment lifecycle
- V2 adds `notes` field
- Current has `commission` and `exit_price` fields
- V2 uses `option_type`, current uses `type`

---

### TREASURY

**V2 Spec:**
```sql
PK  id                  INTEGER
FK  account_id          INTEGER
UK  cuspid              TEXT
FK  tx_id               INTEGER    -- References purchase transaction
    amount              REAL       -- Face value
    yield               REAL
    buy_price           REAL
    current_value       REAL
    purchased_date      DATE
    maturity_date       DATE
    sold_date           DATE
    status              TEXT       -- 'HELD', 'SOLD', 'MATURED'
    notes               TEXT
    created_at          DATETIME
    updated_at          DATETIME
```

**Current Schema:**
```sql
PK  cuspid              TEXT
    purchased           DATE
    maturity            DATE
    amount              REAL
    yield               REAL
    buy_price           REAL
    current_value       REAL
    exit_price          REAL
    created_at          DATETIME
    updated_at          DATETIME
```

**Differences:**
- V2 uses INTEGER PK with unique constraint on `cuspid`
- Current uses TEXT PK (cuspid)
- V2 adds `account_id` FK for multi-account support
- V2 adds `tx_id` linking to purchase transaction
- V2 adds `sold_date` and explicit `status` field
- V2 adds `notes` field
- Current has `exit_price` field
- V2 uses `purchased_date` and `maturity_date`, current uses `purchased` and `maturity`

---

### DIVIDEND

**V2 Spec:**
```sql
PK  id                  INTEGER
FK  account_id          INTEGER
FK  symbol_id           INTEGER
FK  stock_id            INTEGER    -- References stock position
FK  tx_id               INTEGER    -- References dividend receipt transaction
    payment_date        DATE
    ex_dividend_date    DATE
    shares              INTEGER    -- Shares held at payment
    amount_per_share    REAL
    total_amount        REAL       -- shares * amount_per_share
    dividend_type       TEXT       -- 'CASH', 'QUALIFIED', 'SPECIAL'
    notes               TEXT
    created_at          DATETIME
    updated_at          DATETIME
```

**Current Schema:**
```sql
PK  id                  INTEGER
FK  symbol              TEXT
    received            DATE
    amount              REAL
    created_at          DATETIME
```

**Differences:**
- V2 adds `account_id` FK for multi-account support
- V2 uses `symbol_id` INTEGER FK vs current TEXT FK
- V2 adds `stock_id` FK linking to stock position
- V2 adds `tx_id` linking to dividend receipt transaction
- V2 separates `payment_date` and `ex_dividend_date` (current only has `received`)
- V2 breaks down dividend into `shares`, `amount_per_share`, and `total_amount`
- Current only tracks `amount` (total)
- V2 adds `dividend_type` for tax classification
- V2 adds `notes` field
- Current has no `updated_at` field

---

### NEW IN V2: ACCOUNT

**V2 Spec Only:**
```sql
PK  id                  INTEGER
    name                TEXT
    account_type        TEXT       -- 'CASH', 'MARGIN', 'IRA'
    initial_balance     REAL
    created_at          DATETIME
    updated_at          DATETIME
```

**Purpose:**
- Multi-account portfolio tracking
- Separate portfolios within single database
- Account type classification for tax/regulatory purposes
- Foundation for transaction-centric accounting

---

### NEW IN V2: TRANSACTION

**V2 Spec Only:**
```sql
PK  id                  INTEGER
FK  account_id          INTEGER
    asset_type          TEXT       -- 'STOCK', 'OPTION', 'TREASURY', 'DIVIDEND'
    asset_id            INTEGER    -- Polymorphic FK to asset tables
    trade_type          TEXT       -- 'BUY_TO_OPEN', 'SELL_TO_CLOSE', etc.
    transaction_date    DATE
    quantity            INTEGER
    price               REAL
    total_amount        REAL       -- quantity * price (signed)
    commission          REAL
    net_amount          REAL       -- total_amount - commission
    notes               TEXT
    created_at          DATETIME
    updated_at          DATETIME
```

**Purpose:**
- Source of truth for all financial movements
- Account value = SUM(net_amount) for account
- Polymorphic references to all asset types
- Complete audit trail of trading activity
- Enables transaction-level reporting and reconciliation

---

### ONLY IN CURRENT: SETTINGS

**Current Schema Only:**
```sql
PK  name                TEXT
    value               TEXT
    description         TEXT
    created_at          DATETIME
    updated_at          DATETIME
```

**Purpose:**
- Application configuration (Polygon API key, etc.)
- Not present in V2 spec (implementation detail)

---

### ONLY IN CURRENT: METRICS

**Current Schema Only:**
```sql
PK  id                  INTEGER
    created             DATETIME
    type                TEXT       -- 'treasury_value', 'long_value', etc.
    value               REAL
```

**Purpose:**
- Time-series performance tracking
- Historical metrics snapshots
- Not present in V2 spec (could be derived from transactions)

---

## Key Conceptual Changes

### 1. Accounting Model
- **Current**: Asset ledger approach - track current state of positions
- **V2**: Transaction journal approach - track all movements, derive current state

### 2. Multi-Tenancy
- **Current**: Single portfolio assumption
- **V2**: Multiple accounts per database

### 3. Referential Integrity
- **Current**: TEXT foreign keys to symbol table
- **V2**: INTEGER foreign keys throughout for better performance and integrity

### 4. Transaction Tracking
- **Current**: Implicit via asset opened/closed dates
- **V2**: Explicit transaction records with full details

### 5. Asset Lifecycle
- **Current**: Dates only (opened, closed)
- **V2**: Status enums + transaction references for complete lifecycle

### 6. Primary Key Strategy
- **Current**: Mixed (TEXT for symbols/treasuries, INTEGER for others)
- **V2**: Consistent INTEGER PKs with unique constraints on natural keys

### 7. Commission Tracking
- **Current**: Commission field only on options
- **V2**: Commission on all transactions, rolled into net_amount

### 8. Dividend Detail
- **Current**: Total amount only
- **V2**: Shares, per-share amount, total amount, position link, tax classification

### 9. Option Assignment
- **Current**: Implicit (close option, open stock separately)
- **V2**: Explicit via assignment_tx_id linking option → stock transaction

### 10. Value Calculation
- **Current**: Sum asset values (stocks + treasuries + open options)
- **V2**: Sum transaction net_amount per account (double-entry bookkeeping)

## Migration Considerations

### Breaking Changes
1. Account table required for all assets
2. Symbol PK changes from TEXT to INTEGER
3. All asset tables require symbol_id instead of symbol
4. Transaction table must be populated from asset history
5. Status fields must be derived from current dates

### Data Transformation Required
1. Create default Account for existing portfolio
2. Add integer IDs to symbols, update all FKs
3. Generate opening transactions for all existing assets
4. Generate closing transactions for closed positions
5. Populate status fields based on opened/closed dates
6. Split option premium into received/paid based on position type
7. Expand dividends into shares + per_share_amount

### Backwards Compatibility
- Not backwards compatible - requires full migration
- Current schema cannot coexist with V2 schema
- All queries and application code must be updated

### Advantages of V2 Model
1. True double-entry accounting
2. Complete audit trail
3. Multi-account support
4. Better performance (INTEGER FKs)
5. Explicit lifecycle tracking
6. Transaction-level reporting
7. More granular dividend tracking

### Advantages of Current Model
1. Simpler schema
2. Direct asset queries
3. No polymorphic joins
4. Smaller database size
5. Faster for single portfolio use case
