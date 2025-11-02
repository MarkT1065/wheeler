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
│     created_at          DATETIME │
│     updated_at          DATETIME │
└──────────────────────────────────┘
                  │
                  │ 1:1
                  ▼
┌──────────────────────────────────┐
│            CASH                  │
├──────────────────────────────────┤
│ PK  id                  INTEGER  │
│ FK  account_id          INTEGER  │
│     amount              REAL     │
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
Account 1 ────── 1 Cash            (account_id FK)
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
Account:     id=1, name='Trading Account', account_type='CASH', balance=$0
Cash:        account_id=1, amount=$0
```

### 1. Deposit Cash
```
Transaction: trade_type='RECEIVE', asset_type='CASH', asset_id=cash.id, net_amount=+$10,000
Cash:        account_id=1, amount=$10,000
Account:     balance=$10,000
```

### 2. Sell Put to Open
```
Transaction: trade_type='SELL_TO_OPEN', asset_type='OPTION', asset_id=option.id, net_amount=+$150
Option:      status='OPEN', premium_received=$150
Cash:        account_id=1, amount=$10,150
Account:     balance=$10,150
```

### 3. Put Assignment
```
Transaction: trade_type='ASSIGNED', asset_type='OPTION', asset_id=option.id, net_amount=$0
Option:      status='ASSIGNED', assignment_tx_id=stock_tx.id
Transaction: trade_type='BUY_TO_OPEN', asset_type='STOCK', asset_id=stock.id, net_amount=-$5,000
Stock:       status='OPEN', shares=100, cost_basis=$5,000
Cash:        account_id=1, amount=$5,150
Account:     balance=$10,150 (Cash $5,150 + Stock $5,000)
```

### 4. Sell Call to Open
```
Transaction: trade_type='SELL_TO_OPEN', asset_type='OPTION', asset_id=option.id, net_amount=+$200
Option:      status='OPEN', premium_received=$200
Cash:        account_id=1, amount=$5,350
Account:     balance=$10,350
```

### 5. Dividend Received
```
Transaction: trade_type='RECEIVE', asset_type='DIVIDEND', asset_id=dividend.id, net_amount=+$50
Dividend:    total_amount=$50
Cash:        account_id=1, amount=$5,400
Account:     balance=$10,400
```

### 6. Account Value Calculation
```
Cash Balance:        cash.amount WHERE account_id = X
Stock Value:         SUM(stock.shares * current_price) WHERE account_id = X
Option Value:        SUM(option.current_value) WHERE account_id = X AND status = 'OPEN'
Treasury Value:      SUM(treasury.current_value) WHERE account_id = X AND status = 'HELD'

Total Account Value: Cash + Stock + Option + Treasury = account.balance

Alternatively (Transaction-Centric):
SUM(transactions.net_amount WHERE account_id = X) = Current Account Value
```

## Database Indexes

```sql
CREATE INDEX idx_cash_account_id             ON cash(account_id);
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
All asset tables (Cash, Stock, Option, Treasury, Dividend) reference an Account via `account_id` foreign key. This enables tracking multiple portfolios or accounts within a single database. Each account has its own cash balance tracked in the Cash table.

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
