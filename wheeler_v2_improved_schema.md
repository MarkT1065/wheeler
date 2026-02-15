# Wheeler V2 Transaction-Centric Data Model

This schema redesign makes **transaction** the central event log while moving asset-specific attributes directly onto asset tables, eliminating the need for `transaction_*` detail tables.

## Core Design Principles

1. **Transaction as Central Event Log**: Every financial event is recorded as a transaction
2. **Asset Tables Hold State**: Stocks, options, treasuries maintain their own attributes and current state
3. **No transaction_* Tables**: Asset-specific attributes live on the asset entity, not in separate detail tables
4. **Strong Referential Integrity**: All foreign keys enforced at database level
5. **Sign Convention**: Consistent accounting - positive = money in, negative = money out

## Core Philosophy

**Transaction** = Immutable append-only event log of all financial activity
**Assets** = Current state of holdings (stocks, options, treasuries)
**No Polymorphism** = Transaction references assets directly via nullable FKs

---

## Entity Relationship Diagram

```
┌──────────────────────────────────┐
│          ACCOUNT                 │
│  Central portfolio container     │
├──────────────────────────────────┤
│ PK  id                  INTEGER  │
│     name                TEXT     │
│     created_at          DATETIME │
│     updated_at          DATETIME │
└──────────────────────────────────┘
                │
                │ 1:N
                ▼
┌──────────────────────────────────┐
│         TRANSACTION              │  ← CENTER OF THE MODEL
│  Every financial event           │
├──────────────────────────────────┤
│ PK  id                  INTEGER  │
│ FK  account_id          INTEGER  │
│ FK  trade_id            INTEGER  │  Optional strategy grouping
│ FK  stock_id            INTEGER  │  Nullable - links to affected asset
│ FK  option_id           INTEGER  │  Nullable - links to affected asset
│ FK  treasury_id         INTEGER  │  Nullable - links to affected asset
│     transaction_type    TEXT     │  'STOCK_BUY', 'OPTION_SELL_OPEN', etc.
│     transaction_date    DATE     │
│     amount              REAL     │  Signed: + = credit, - = debit
│     commission          REAL     │  Always positive or zero
│     net_amount          REAL     │  amount - commission (signed)
│     notes               TEXT     │
│     created_at          DATETIME │
│     updated_at          DATETIME │
└──────────────────────────────────┘
         │              │              │
         ▼              ▼              ▼
    ┌────────┐    ┌─────────┐    ┌──────────┐
    │ STOCK  │    │ OPTION  │    │ TREASURY │
    │  Asset │    │  Asset  │    │  Asset   │
    │  State │    │  State  │    │  State   │
    └────────┘    └─────────┘    └──────────┘
         │
         ▼
    ┌─────────┐
    │ SYMBOL  │
    │ Ref Data│
    └─────────┘
```

---

## Schema Definition

### 1. Account
```sql
CREATE TABLE account (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### 2. Symbol (Reference Data)
```sql
CREATE TABLE symbol (
    symbol TEXT PRIMARY KEY,
    name TEXT,
    current_price REAL,
    dividend_yield REAL,
    ex_dividend_date DATE,
    pe_ratio REAL,
    on_watchlist BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### 3. Transaction (Central Event Log)
```sql
CREATE TABLE transaction (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL,
    trade_id INTEGER,  -- Optional: groups related transactions into strategies
    
    -- What happened
    transaction_type TEXT NOT NULL CHECK (transaction_type IN (
        'CASH_DEPOSIT', 'CASH_WITHDRAW',
        'STOCK_BUY', 'STOCK_SELL',
        'OPTION_SELL_OPEN', 'OPTION_BUY_CLOSE', 
        'TREASURY_BUY', 'TREASURY_SELL', 'TREASURY_MATURE', 'TREASURY_INTEREST',
        'DIVIDEND_RECEIVE'
    )),
    transaction_date DATE NOT NULL,
    
    -- Financial impact (accounting)
    amount REAL NOT NULL,  -- Signed: + = credit, - = debit
    commission REAL NOT NULL DEFAULT 0 CHECK (commission >= 0),
    net_amount REAL NOT NULL,  -- amount - commission
    
    -- Links to affected assets (nullable - depends on type)
    stock_id INTEGER,
    option_id INTEGER,
    treasury_id INTEGER,
    
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE CASCADE,
    FOREIGN KEY (trade_id) REFERENCES trade(id) ON DELETE SET NULL,
    FOREIGN KEY (stock_id) REFERENCES stock(id) ON DELETE RESTRICT,
    FOREIGN KEY (option_id) REFERENCES option(id) ON DELETE RESTRICT,
    FOREIGN KEY (treasury_id) REFERENCES treasury(id) ON DELETE RESTRICT
);

CREATE INDEX idx_transaction_account_id ON transaction(account_id);
CREATE INDEX idx_transaction_trade_id ON transaction(trade_id);
CREATE INDEX idx_transaction_date ON transaction(transaction_date);
CREATE INDEX idx_transaction_type ON transaction(transaction_type);
CREATE INDEX idx_transaction_account_date ON transaction(account_id, transaction_date);
CREATE INDEX idx_transaction_stock_id ON transaction(stock_id);
CREATE INDEX idx_transaction_option_id ON transaction(option_id);
CREATE INDEX idx_transaction_treasury_id ON transaction(treasury_id);
```

### 4. Stock (Asset State)
```sql
CREATE TABLE stock (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL,
    symbol TEXT NOT NULL,
    
    -- Position tracking
    shares INTEGER NOT NULL CHECK (shares >= 0),  -- Current holding
    avg_cost_basis REAL NOT NULL,  -- Weighted average cost per share
    
    -- Lifecycle
    opened_date DATE NOT NULL,
    closed_date DATE,  -- NULL if still open
    status TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'CLOSED')),
    
    -- Origin tracking
    acquisition_type TEXT NOT NULL CHECK (acquisition_type IN (
        'PURCHASED', 'ASSIGNED_FROM_PUT', 'ASSIGNED_FROM_CALL', 'TRANSFER_IN'
    )),
    assignment_option_id INTEGER,  -- Links to option if acquired via assignment
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE CASCADE,
    FOREIGN KEY (symbol) REFERENCES symbol(symbol) ON DELETE RESTRICT,
    FOREIGN KEY (assignment_option_id) REFERENCES option(id) ON DELETE SET NULL
);

CREATE INDEX idx_stock_account_id ON stock(account_id);
CREATE INDEX idx_stock_symbol ON stock(symbol);
CREATE INDEX idx_stock_status ON stock(status);
CREATE INDEX idx_stock_opened ON stock(opened_date);
```

### 5. Option (Asset State)
```sql
CREATE TABLE option (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL,
    symbol TEXT NOT NULL,
    
    -- Option specification
    option_type TEXT NOT NULL CHECK (option_type IN ('PUT', 'CALL')),
    strike REAL NOT NULL CHECK (strike > 0),
    expiration DATE NOT NULL,
    
    -- Position tracking
    contracts INTEGER NOT NULL CHECK (contracts >= 0),  -- Current open contracts
    avg_premium_per_contract REAL NOT NULL,  -- Weighted average
    
    -- Lifecycle
    opened_date DATE NOT NULL,
    closed_date DATE,  -- NULL if still open
    status TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN (
        'OPEN', 'CLOSED', 'ASSIGNED', 'EXPIRED'
    )),
    
    -- Assignment tracking (bidirectional link)
    assignment_stock_id INTEGER,  -- Links to resulting stock position if assigned
    
    -- Current market data
    current_price REAL,
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE CASCADE,
    FOREIGN KEY (symbol) REFERENCES symbol(symbol) ON DELETE RESTRICT,
    FOREIGN KEY (assignment_stock_id) REFERENCES stock(id) ON DELETE SET NULL,
    
    -- Prevent duplicates
    UNIQUE(account_id, symbol, option_type, strike, expiration, opened_date)
);

CREATE INDEX idx_option_account_id ON option(account_id);
CREATE INDEX idx_option_symbol ON option(symbol);
CREATE INDEX idx_option_status ON option(status);
CREATE INDEX idx_option_expiration ON option(expiration);
CREATE INDEX idx_option_type ON option(option_type);
```

### 6. Treasury (Asset State)
```sql
CREATE TABLE treasury (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL,
    
    -- Treasury specification
    cuspid TEXT NOT NULL,
    face_value REAL NOT NULL CHECK (face_value > 0),
    yield REAL NOT NULL,
    maturity_date DATE NOT NULL,
    
    -- Position tracking
    current_amount REAL NOT NULL CHECK (current_amount >= 0),  -- Dynamically adjusted for collateral
    purchase_price REAL NOT NULL,
    
    -- Lifecycle
    purchased_date DATE NOT NULL,
    sold_date DATE,  -- NULL if still held
    status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN (
        'ACTIVE', 'SOLD', 'MATURED'
    )),
    
    -- Current market data
    current_value REAL,
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE CASCADE
);

CREATE INDEX idx_treasury_account_id ON treasury(account_id);
CREATE INDEX idx_treasury_status ON treasury(status);
CREATE INDEX idx_treasury_maturity ON treasury(maturity_date);
CREATE INDEX idx_treasury_purchased ON treasury(purchased_date);
```

### 7. Dividend (Separate Tracking)
```sql
CREATE TABLE dividend (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL,
    stock_id INTEGER NOT NULL,  -- Which position generated this
    
    shares_held INTEGER NOT NULL CHECK (shares_held > 0),
    amount_per_share REAL NOT NULL CHECK (amount_per_share > 0),
    total_amount REAL NOT NULL,
    
    ex_dividend_date DATE NOT NULL,
    payment_date DATE NOT NULL,
    dividend_type TEXT NOT NULL DEFAULT 'CASH' CHECK (dividend_type IN ('CASH', 'QUALIFIED', 'SPECIAL')),
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE CASCADE,
    FOREIGN KEY (stock_id) REFERENCES stock(id) ON DELETE CASCADE
);

CREATE INDEX idx_dividend_account_id ON dividend(account_id);
CREATE INDEX idx_dividend_stock_id ON dividend(stock_id);
CREATE INDEX idx_dividend_payment_date ON dividend(payment_date);
```

### 8. Trade (Optional Strategy Grouping)
```sql
CREATE TABLE trade (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL,
    symbol TEXT,
    
    strategy_type TEXT CHECK (strategy_type IN (
        'WHEEL', 'LONG_STOCK', 'COVERED_CALL', 'CASH_SECURED_PUT',
        'IRON_CONDOR', 'BUTTERFLY', 'STRADDLE', 'STRANGLE'
    )),
    
    opened_date DATE NOT NULL,
    closed_date DATE,
    status TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'CLOSED', 'PARTIALLY_CLOSED')),
    notes TEXT,
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE CASCADE,
    FOREIGN KEY (symbol) REFERENCES symbol(symbol) ON DELETE RESTRICT
);

CREATE INDEX idx_trade_account_id ON trade(account_id);
CREATE INDEX idx_trade_symbol ON trade(symbol);
CREATE INDEX idx_trade_status ON trade(status);
CREATE INDEX idx_trade_dates ON trade(opened_date, closed_date);
```

### 9. Daily Snapshot (Charting Performance)
```sql
CREATE TABLE daily_snapshot (
    account_id INTEGER NOT NULL,
    date DATE NOT NULL,
    cash_balance REAL NOT NULL,
    stock_value REAL NOT NULL,
    option_value REAL NOT NULL,
    treasury_value REAL NOT NULL,
    total_value REAL NOT NULL,
    daily_pl REAL NOT NULL,
    cumulative_pl REAL NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (account_id, date),
    FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE CASCADE
);

CREATE INDEX idx_daily_snapshot_date ON daily_snapshot(date);
```

---

## Key Design Decisions

### 1. Transaction References Assets (Not Vice Versa)
- Transaction has nullable FKs: `stock_id`, `option_id`, `treasury_id`
- Assets don't reference transactions (avoids circular dependencies)
- Transaction log is append-only history

### 2. Assets Hold Their Own Attributes
- **Stock**: `shares`, `avg_cost_basis`, `acquisition_type`
- **Option**: `contracts`, `strike`, `expiration`, `avg_premium_per_contract`
- **Treasury**: `cuspid`, `face_value`, `yield`, `current_amount`
- **No `transaction_stock`, `transaction_option`, `transaction_treasury` tables**

### 3. Position State Lives on Asset
- `shares` on stock (not computed from transactions)
- `contracts` on option (not computed from transactions)
- `current_amount` on treasury (adjusted for collateral usage)
- Assets updated when transactions occur

### 4. Bidirectional Assignment Links
- Option → Stock: `option.assignment_stock_id`
- Stock → Option: `stock.assignment_option_id`
- Allows traversal in both directions for wheel strategy tracking

---

## Sign Convention

**Accounting Rule**: 
- Positive `amount` = money INTO account (credits)
- Negative `amount` = money OUT OF account (debits)
- `commission` is always positive or zero
- `net_amount = amount - commission`

---

## Transaction Flow Examples

### Example 1: Buy Stock
```sql
-- Create stock position
INSERT INTO stock (account_id, symbol, shares, avg_cost_basis, opened_date, acquisition_type)
VALUES (1, 'AAPL', 100, 50.00, '2025-01-02', 'PURCHASED');

-- Record transaction
INSERT INTO transaction (account_id, transaction_type, transaction_date, stock_id, amount, commission, net_amount)
VALUES (1, 'STOCK_BUY', '2025-01-02', last_insert_rowid(), -5000.00, 1.00, -5001.00);
```

### Example 2: Sell Put to Open
```sql
-- Create option position
INSERT INTO option (account_id, symbol, option_type, strike, expiration, contracts, avg_premium_per_contract, opened_date)
VALUES (1, 'AAPL', 'PUT', 50.00, '2025-02-21', 1, 1.50, '2025-01-03');

-- Record transaction
INSERT INTO transaction (account_id, trade_id, transaction_type, transaction_date, option_id, amount, commission, net_amount)
VALUES (1, 1, 'OPTION_SELL_OPEN', '2025-01-03', last_insert_rowid(), 150.00, 0.65, 149.35);
```

### Example 3: Put Assignment (Wheel Strategy)
```sql
-- Update option status and close contracts
UPDATE option 
SET status = 'ASSIGNED', 
    closed_date = '2025-02-21', 
    contracts = 0 
WHERE id = 5;

-- Create stock from assignment
INSERT INTO stock (account_id, symbol, shares, avg_cost_basis, opened_date, acquisition_type, assignment_option_id)
VALUES (1, 'AAPL', 100, 50.00, '2025-02-21', 'ASSIGNED_FROM_PUT', 5);

-- Link option to resulting stock (bidirectional)
UPDATE option 
SET assignment_stock_id = last_insert_rowid() 
WHERE id = 5;

-- Record cash transaction for stock purchase
INSERT INTO transaction (account_id, trade_id, transaction_type, transaction_date, stock_id, option_id, amount, commission, net_amount)
VALUES (1, 1, 'STOCK_BUY', '2025-02-21', last_insert_rowid(), 5, -5000.00, 0, -5000.00);
```

### Example 4: Receive Dividend
```sql
-- Record dividend
INSERT INTO dividend (account_id, stock_id, shares_held, amount_per_share, total_amount, ex_dividend_date, payment_date, dividend_type)
VALUES (1, 3, 100, 0.50, 50.00, '2025-03-10', '2025-03-15', 'QUALIFIED');

-- Record transaction
INSERT INTO transaction (account_id, transaction_type, transaction_date, amount, commission, net_amount, notes)
VALUES (1, 'DIVIDEND_RECEIVE', '2025-03-15', 50.00, 0, 50.00, 'AAPL dividend - 100 shares @ $0.50');
```

### Example 5: Sell Call Against Stock (Covered Call)
```sql
-- Assume stock_id = 3 (100 shares AAPL)

-- Create option position
INSERT INTO option (account_id, symbol, option_type, strike, expiration, contracts, avg_premium_per_contract, opened_date)
VALUES (1, 'AAPL', 'CALL', 55.00, '2025-03-21', 1, 2.25, '2025-02-22');

-- Record transaction
INSERT INTO transaction (account_id, trade_id, transaction_type, transaction_date, option_id, amount, commission, net_amount)
VALUES (1, 1, 'OPTION_SELL_OPEN', '2025-02-22', last_insert_rowid(), 225.00, 0.65, 224.35);
```

### Example 6: Treasury Purchase (Cash Collateral)
```sql
-- Create treasury position
INSERT INTO treasury (account_id, cuspid, face_value, yield, maturity_date, current_amount, purchase_price, purchased_date)
VALUES (1, '912828ZG8', 10000.00, 4.5, '2025-08-15', 10000.00, 9950.00, '2025-01-05');

-- Record transaction
INSERT INTO transaction (account_id, transaction_type, transaction_date, treasury_id, amount, commission, net_amount)
VALUES (1, 'TREASURY_BUY', '2025-01-05', last_insert_rowid(), -9950.00, 0, -9950.00);
```

### Example 7: Adjust Treasury Collateral (Put Assignment)
```sql
-- When put is assigned, reduce treasury amount (collateral used for stock purchase)
UPDATE treasury 
SET current_amount = current_amount - 5000.00,
    updated_at = CURRENT_TIMESTAMP
WHERE id = 1 AND status = 'ACTIVE';

-- No transaction needed - this is just internal collateral adjustment
```

---

## Views for Analytics

### Open Stock Positions
```sql
CREATE VIEW position_stock_open AS
SELECT 
    s.id,
    s.account_id,
    s.symbol,
    sym.name AS symbol_name,
    s.shares,
    s.avg_cost_basis,
    s.shares * s.avg_cost_basis AS total_cost,
    sym.current_price,
    s.shares * sym.current_price AS current_value,
    (s.shares * sym.current_price) - (s.shares * s.avg_cost_basis) AS unrealized_pl,
    s.opened_date,
    s.acquisition_type
FROM stock s
JOIN symbol sym ON s.symbol = sym.symbol
WHERE s.status = 'OPEN' AND s.shares > 0;
```

### Open Option Positions
```sql
CREATE VIEW position_option_open AS
SELECT 
    o.id,
    o.account_id,
    o.symbol,
    o.option_type,
    o.strike,
    o.expiration,
    o.contracts,
    o.avg_premium_per_contract,
    o.contracts * o.avg_premium_per_contract * 100 AS total_premium_collected,
    o.current_price,
    o.opened_date,
    JULIANDAY(o.expiration) - JULIANDAY('now') AS days_to_expiration
FROM option
WHERE status = 'OPEN' AND contracts > 0;
```

### Active Treasury Positions
```sql
CREATE VIEW position_treasury_active AS
SELECT 
    t.id,
    t.account_id,
    t.cuspid,
    t.face_value,
    t.current_amount,
    t.yield,
    t.maturity_date,
    t.purchase_price,
    t.current_value,
    t.purchased_date,
    JULIANDAY(t.maturity_date) - JULIANDAY('now') AS days_to_maturity
FROM treasury t
WHERE t.status = 'ACTIVE';
```

### Account Cash Balance
```sql
CREATE VIEW account_cash_balance AS
SELECT 
    account_id,
    SUM(net_amount) AS cash_balance
FROM transaction
GROUP BY account_id;
```

### Monthly Performance
```sql
CREATE VIEW monthly_performance AS
SELECT 
    account_id,
    strftime('%Y-%m', transaction_date) AS month,
    SUM(CASE WHEN transaction_type = 'OPTION_SELL_OPEN' THEN net_amount ELSE 0 END) AS option_premium,
    SUM(CASE WHEN transaction_type = 'DIVIDEND_RECEIVE' THEN net_amount ELSE 0 END) AS dividend_income,
    SUM(CASE WHEN transaction_type = 'TREASURY_INTEREST' THEN net_amount ELSE 0 END) AS interest_income,
    SUM(CASE WHEN transaction_type = 'STOCK_SELL' THEN net_amount ELSE 0 END) AS stock_sale_proceeds,
    SUM(net_amount) AS total_net_amount
FROM transaction
GROUP BY account_id, strftime('%Y-%m', transaction_date)
ORDER BY month;
```

---

## Benefits of This Approach

1. **Transaction is the immutable log** - Every financial event recorded once, append-only
2. **Assets hold current state** - No need to aggregate transactions to know position size
3. **No transaction_* tables** - Attributes live where they semantically belong
4. **Clean FK relationships** - Transaction → Asset (unidirectional, no circles)
5. **Efficient queries** - "What options do I have?" → `SELECT * FROM option WHERE status = 'OPEN'`
6. **Audit trail preserved** - Transaction history intact even after assets closed
7. **Wheel strategy tracking** - Bidirectional links between options and stocks for assignments
8. **Treasury collateral** - Dynamic `current_amount` tracks collateral usage without transaction spam

---

## Key Improvements Over Original V2 Spec

1. **Eliminated transaction_* tables**: Asset attributes now live directly on asset entities
2. **Transaction as central log**: Single source of truth for all financial events
3. **Asset state separation**: Current holdings tracked separately from historical transactions
4. **Simpler queries**: No joins to polymorphic detail tables required
5. **Bidirectional assignment tracking**: Options and stocks link both ways for wheel strategy
6. **Treasury collateral model**: `current_amount` adjusts without creating transactions for internal moves

This maintains transaction-centric logging while avoiding the complexity and join overhead of polymorphic detail tables.
