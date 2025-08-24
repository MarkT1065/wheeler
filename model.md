# Data Model

## Overview
This document defines the data model for Wheeler, an advanced options trading portfolio system specializing in the "wheel strategy" and Treasury collateral management. The model consists of five main entities with defined relationships and attributes, implemented in SQLite with proper indexing and constraints for modern web application patterns.

## Entities

### Symbols
Represents a stock's unique exchange ticker and fundamental information.

**Primary Key:** symbol (TEXT)

**Attributes:**
- symbol (TEXT) - Unique stock ticker identifier (e.g., "AAPL", "MSFT")
- price (REAL) - Current stock price (default: 0.0)
- dividend (REAL) - Current dividend yield (default: 0.0)
- ex_dividend_date (DATE) - Last ex-dividend date
- pe_ratio (REAL) - Price-to-earnings ratio
- created_at (DATETIME) - Record creation timestamp (default: CURRENT_TIMESTAMP)
- updated_at (DATETIME) - Record update timestamp (default: CURRENT_TIMESTAMP)

### Long Positions
Represents long stock positions, often resulting from put option assignments in wheel strategy trading.

**Primary Key:** id (INTEGER AUTOINCREMENT)
**Unique Constraint:** (symbol, opened, shares, buy_price) - Prevents duplicate entries

**Attributes:**
- id (INTEGER) - Auto-incrementing primary key for web-friendly operations
- symbol (TEXT) - Foreign key to symbols table
- opened (DATE) - Date position was opened
- closed (DATE) - Date position was closed (null if still open)
- shares (INTEGER) - Number of shares held
- buy_price (REAL) - Price per share at purchase
- exit_price (REAL) - Price per share at sale (null if still open)
- created_at (DATETIME) - Record creation timestamp (default: CURRENT_TIMESTAMP)
- updated_at (DATETIME) - Record update timestamp (default: CURRENT_TIMESTAMP)

**Constraints:**
- symbol must reference existing symbol in symbols table
- shares must be positive integer
- buy_price must be positive
- Unique constraint on (symbol, opened, shares, buy_price)

### Options
Represents options positions (cash-secured puts and covered calls) central to wheel strategy trading.

**Primary Key:** id (INTEGER AUTOINCREMENT)
**Unique Constraint:** (symbol, type, opened, strike, expiration, premium, contracts) - Prevents duplicate entries

**Attributes:**
- id (INTEGER) - Auto-incrementing primary key for web-friendly operations
- symbol (TEXT) - Foreign key to symbols table
- type (TEXT) - Option type: "Put" or "Call" (CHECK constraint enforced)
- opened (DATE) - Date option was sold/opened
- closed (DATE) - Date option was closed (null if still open)
- strike (REAL) - Strike price of the option
- expiration (DATE) - Option expiration date
- premium (REAL) - Premium received when selling the option
- contracts (INTEGER) - Number of option contracts
- exit_price (REAL) - Price paid to close position (null if still open)
- created_at (DATETIME) - Record creation timestamp (default: CURRENT_TIMESTAMP)
- updated_at (DATETIME) - Record update timestamp (default: CURRENT_TIMESTAMP)

**Wheel Strategy Context:**
- **Cash-Secured Puts**: Backed by Treasury collateral, convert to stock positions on assignment
- **Covered Calls**: Sold against existing stock positions, generate premium income
- **Assignment Tracking**: Options that reach expiration ITM trigger collateral adjustments

**Constraints:**
- symbol must reference existing symbol in symbols table
- type must be either "Put" or "Call"
- contracts must be positive integer
- premium and strike must be positive
- Unique constraint on (symbol, type, opened, strike, expiration, premium, contracts)

### Dividends
Represents dividend payments received from stock holdings, complementing wheel strategy income.

**Primary Key:** id (INTEGER AUTOINCREMENT)
**Unique Constraint:** (symbol, received, amount) - Prevents duplicate entries

**Attributes:**
- id (INTEGER) - Auto-incrementing primary key for web-friendly operations
- symbol (TEXT) - Foreign key to symbols table
- received (DATE) - Date dividend was received
- amount (REAL) - Dividend amount received
- created_at (DATETIME) - Record creation timestamp (default: CURRENT_TIMESTAMP)

**Constraints:**
- symbol must reference existing symbol in symbols table
- amount must be positive
- Unique constraint on (symbol, received, amount)

### Treasuries
Represents U.S. Treasury securities used as cash collateral for options trading in the wheel strategy.

**Primary Key:** cuspid (TEXT) - Natural primary key using CUSIP identifier

**Attributes:**
- cuspid (TEXT) - Unique CUSIP identifier for the treasury security
- purchased (DATE) - Date treasury was purchased
- maturity (DATE) - Treasury maturity date
- amount (REAL) - Face value amount of the treasury (dynamically adjusted for collateral)
- yield (REAL) - Treasury yield at purchase
- buy_price (REAL) - Price paid for the treasury
- current_value (REAL) - Current market value (null if not updated)
- exit_price (REAL) - Sale price if sold (null if still held)
- created_at (DATETIME) - Record creation timestamp (default: CURRENT_TIMESTAMP)
- updated_at (DATETIME) - Record update timestamp (default: CURRENT_TIMESTAMP)

**Wheel Strategy Integration:**
- **Cash Collateral**: Treasury amounts automatically decrease when puts are assigned
- **Collateral Recovery**: Treasury amounts increase when calls are assigned or puts expire worthless
- **Interest Income**: Quarterly interest payments recorded as new Treasury entries
- **Yield Optimization**: Balance collateral needs with Treasury yields and maturities

**Constraints:**
- cuspid must be unique
- amount, yield, and buy_price must be positive
- maturity must be after purchased date

### Settings
Represents application configuration settings stored as name-value pairs for dynamic system configuration.

**Primary Key:** name (TEXT) - Natural primary key using setting name

**Attributes:**
- name (TEXT) - Unique setting name identifier (e.g., "POLYGON_API_KEY", "AUTO_UPDATE_PRICES")
- value (TEXT) - Setting value stored as text (can be parsed for different data types)
- description (TEXT) - Human-readable description of the setting purpose
- created_at (DATETIME) - Record creation timestamp (default: CURRENT_TIMESTAMP)
- updated_at (DATETIME) - Record update timestamp (default: CURRENT_TIMESTAMP)

**Common Settings:**
- **POLYGON_API_KEY**: API key for Polygon.io stock market data integration
- **AUTO_UPDATE_INTERVAL**: Minutes between automatic price updates
- **DEFAULT_CURRENCY**: Base currency for portfolio calculations
- **ENABLE_NOTIFICATIONS**: Enable/disable system notifications

**Constraints:**
- name must be unique
- name cannot be null or empty
- value can be null (for boolean false or unset values)

## Relationships

```
Symbols (1) ←→ (Many) Long Positions (via symbol FK)
Symbols (1) ←→ (Many) Options (via symbol FK)
Symbols (1) ←→ (Many) Dividends (via symbol FK)
Treasuries (Independent entity - no FK relationships)
Settings (Independent entity - no FK relationships)
```

### Primary Key Strategy

Wheeler uses a hybrid primary key approach optimized for modern web applications:

**Transactional Tables (Auto-increment IDs):**
- options.id, long_positions.id, dividends.id
- Web-friendly integer IDs for easy HTTP CRUD operations
- Unique constraints on business keys prevent duplicate records

**Reference Tables (Natural Keys):**
- symbols.symbol (stock ticker), treasuries.cuspid (bond identifier), settings.name (configuration key)
- Business identifiers as primary keys for reference data

## Database Indexes

### Performance Optimization Indexes
- `idx_symbols_symbol` - Primary key index on symbols.symbol
- `idx_long_positions_symbol` - Foreign key index on long_positions.symbol
- `idx_long_positions_opened` - Query optimization for date ranges
- `idx_options_symbol` - Foreign key index on options.symbol
- `idx_options_expiration` - Query optimization for expiration dates
- `idx_options_type` - Query optimization for Put/Call filtering
- `idx_dividends_symbol` - Foreign key index on dividends.symbol
- `idx_dividends_received` - Query optimization for date ranges
- `idx_treasuries_cuspid` - Primary key index on treasuries.cuspid
- `idx_treasuries_maturity` - Query optimization for maturity dates
- `idx_treasuries_purchased` - Query optimization for purchase dates
- `idx_settings_name` - Primary key index on settings.name

## Data Constraints & Business Rules

### Referential Integrity
- All symbol references in long_positions, options, and dividends must exist in symbols table
- Foreign key constraints enforced at database level

### Data Validation
- Dates must be in valid date format (YYYY-MM-DD)
- Numeric values must be non-negative where applicable
- Option type must be either "Put" or "Call" (CHECK constraint)
- Compound primary keys ensure unique position identification

### Business Logic Constraints
- Long positions can have null closed date (open positions)
- Options can have null exit_price (open positions)
- Treasuries can have null current_value and exit_price (held positions)
- All monetary values stored as REAL type for precision

## Schema Evolution Notes

### Timestamp Tracking
- All tables include created_at and updated_at fields for audit trail
- Timestamps automatically set to CURRENT_TIMESTAMP on record creation
- updated_at should be manually updated on record modifications

### Web Application Optimization
- Integer primary keys provide clean URLs and easy REST API operations
- Unique constraints on business keys prevent duplicate data entry
- Foreign key constraints maintain referential integrity
- Auto-increment IDs avoid compound key complexity in web forms

### Wheel Strategy Data Flow

**Option Assignment Process:**
1. Cash-secured put expires ITM → Treasury amount decreases (collateral used)
2. New long position created with assigned shares
3. Covered call sold against new stock position
4. Call assignment → Treasury amount increases (cash received)

**Treasury Collateral Management:**
- Put assignments reduce Treasury balances (cash used for stock purchase)
- Call assignments increase Treasury balances (stock sold for cash)
- Interest payments add new Treasury entries quarterly
- Treasury table independent of symbols (bonds vs. stocks)