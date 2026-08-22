# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is "Wheeler" - a comprehensive financial portfolio tracking system built with Go. The project specializes in tracking sophisticated options trading strategies, particularly the "wheel strategy" (cash-secured puts, covered calls, and stock assignments), along with comprehensive portfolio management including Treasury securities collateral management.

## Applications

The project contains one main application:

1. **Web Dashboard** (`main.go`) - Modern web interface for comprehensive portfolio tracking

## Development Commands

### Quick Start
```bash
go run main.go                    # Start web server on :8080
make build                        # Build to bin/wheeler with CGO enabled
make test                         # Run all tests (requires CGO for SQLite)
```

### Database Operations
- **Multiple databases**: Current DB tracked in `./data/currentdb` file
- **Test data**: Click "Generate Test Data" in Help → Tutorial (loads `wheel_strategy_example.sql`)
- **Backups**: Manual backup via Admin page saves to `./data/backups/`

### Testing
- Unit tests alongside code files (`*_test.go`)
- Integration tests in `test/` folder with consistent naming pattern
- Integration tests require live Polygon.io API key: `internal/polygon/live_integration_test.go`
- Test databases automatically cleaned: `rm -f test_*.db`

## Dependencies

- Go 1.19+ with modules support
- SQLite3 (`github.com/mattn/go-sqlite3`) - requires CGO
- Chart.js for interactive visualizations
- Standard library: `database/sql`, `net/http`, `html/template`
- No GTK or external UI framework dependencies for web dashboard

## Architecture Overview

Wheeler follows a layered architecture with clear separation of concerns:

### Service Layer Pattern (`internal/models/`)
Each domain entity has a dedicated service struct wrapping the `*sql.DB`:
- **OptionService**: Manages Put/Call lifecycle with automatic commission calculation ($0.65/contract)
- **LongPositionService**: Tracks stock holdings from put assignments to call assignments
- **SymbolService**: Maintains stock metadata (price, dividend, P/E ratio)
- **TreasuryService**: Handles collateral with dynamic balance adjustments
- **DividendService**: Records dividend payments and yield calculations
- **MetricService**: Aggregates portfolio performance and P&L calculations

Services use **ID-based CRUD** operations (`GetByID`, `Update`, `Delete`) for web-friendly REST patterns. Compound key fallbacks exist for backward compatibility but are deprecated.

### Database Design (`internal/database/`)
**Hybrid Primary Key Strategy**:
- **Transactional tables** use `INTEGER PRIMARY KEY AUTOINCREMENT` for easier HTTP operations: `options.id`, `long_positions.id`, `dividends.id`
- **Reference tables** use natural keys: `symbols.symbol`, `treasuries.cuspid`, `settings.name`
- **Unique constraints** prevent duplicate business records (e.g., same option opened twice)

**Schema Management**:
- `schema.sql` is the single source of truth (embedded via `//go:embed`) and defines the base schema
- Incremental schema changes are managed via embedded SQL migrations in `internal/database/migrations/*.sql`, which are executed at startup
- SQLite WAL mode with foreign keys enabled: `?_busy_timeout=10000&_journal_mode=WAL&_foreign_keys=on`

### Market Data Providers (`internal/providers/`)
- Defines a common interface and types for market data providers (quotes, aggregates, ticker metadata)
- Application code depends on this abstraction rather than any specific vendor
- New providers should implement this interface and be registered via the provider layer

### Polygon Provider (`internal/polygon/`)
- Default implementation of the market data provider interface using the Polygon.io REST API
- `client.go`: HTTP client wrapper around Polygon.io endpoints
- `service.go`: Business logic layer for fetching quotes, aggregates, and ticker details via the provider interface

### Web Layer (`internal/web/`)
**Handler Organization**:
- `server.go`: Server struct initialization, routing, template loading with custom functions
- `*_handlers.go`: Domain-specific handlers (dashboard, options, symbols, etc.)
- `types.go`: Web-specific data structures for JSON API responses
- `utility_handlers.go`: Shared helper functions

**Template Pattern**:
- HTML templates in `templates/` use Go templating with custom functions (`groupByExpiration`, `add`, `mul`)
- Shared component: `_symbol_modal.html` for reusable symbol entry
- Chart.js for interactive visualizations (scatter plots, pie charts with click navigation)

## Web Dashboard Features

The modern web interface provides comprehensive portfolio tracking:

### Main Pages
- **Dashboard** (`/`) - Portfolio overview with charts and performance metrics
- **Monthly** (`/monthly`) - Month-by-month performance analysis
- **Options** (`/options`) - Detailed options positions and trading
- **Treasuries** (`/treasuries`) - Treasury securities collateral management
- **Symbol Pages** (`/symbol/{SYMBOL}`) - Individual stock analysis and history
- **Help** (`/help`) - Wheeler Help and Tutorial (with test data generation)
- **Admin** (`/backup`) - Database management and backups
- **Import** (`/import`) - CSV data import tools
- **Settings** (`/settings`) - Polygon.io API configuration

### Dashboard Components
- **Long by Symbol Chart** - Pie chart of current stock positions
- **Put Exposure by Symbol Chart** - Options risk exposure visualization
- **Total Allocation Chart** - Complete portfolio allocation (stocks + treasuries)
- **Watchlist Summary Table** - Real-time performance metrics and P&L
- **Quick Actions** - Add symbols, options, and positions

### Key Features
- **Wheel Strategy Support** - Full lifecycle tracking of cash-secured puts and covered calls
- **Treasury Collateral Management** - Automatic adjustment of Treasury positions based on option assignments
- **Multiple Database Support** - Create separate databases for different portfolios or testing
- **Test Data Generation** - One-click import of realistic wheel strategy trading history
- **Comprehensive Import Tools** - CSV import for options, stocks, and dividends
- **Real-time Calculations** - Automatic P&L, allocation, and risk calculations
- **Polygon.io Integration** - Live market data integration with API key management
- **Interactive Charts** - Click-to-navigate functionality on scatter plots and pie charts
- **Responsive Design** - Modern web interface with Chart.js visualizations

### API Endpoints
- `/api/symbols/{symbol}` - Symbol CRUD operations and price updates
- `/api/options` - Options management (create, update, assign, close)
- `/api/long-positions` - Stock position lifecycle management
- `/api/dividends` - Dividend tracking and yield calculations
- `/api/treasuries/{cuspid}` - Treasury operations and interest tracking
- `/api/allocation-data` - Portfolio allocation data for charts
- `/api/generate-test-data` - Test data generation for tutorials
- `/api/settings` - Application settings management
- `/api/provider/*` - Market data provider endpoints (connection tests, price refresh, dividend sync)

## Financial Domain Context

Wheeler specializes in sophisticated options trading strategies:

### Wheel Strategy Implementation
- **Cash-Secured Puts** - Sell puts with cash collateral, track assignments
- **Covered Calls** - Sell calls against stock positions, track exercises
- **Stock Assignments** - Automatic conversion of expired puts to stock positions
- **Premium Collection** - Track income from option premiums across all strategies
- **Position Scaling** - Support for increasing position sizes over time

### Treasury Collateral Management
- **Cash Management** - Treasury securities as collateral for options positions
- **Collateral Adjustment** - Automatic Treasury balance changes on option assignments
- **Interest Tracking** - Quarterly interest payments on Treasury holdings
- **Yield Optimization** - Track yields and maturities across Treasury positions

### Portfolio Components
- **Stock Symbols** - Current prices, dividend yields, P/E ratios, watchlist tracking
- **Options Positions** - Complete lifecycle from opening to assignment/expiration
- **Long Stock Holdings** - Entry/exit tracking with cost basis and P&L
- **Dividend Tracking** - Payment recording and yield analysis
- **Performance Analytics** - Monthly breakdowns, allocation analysis, risk metrics
- **Market Data Integration** - Live price updates via Polygon.io API

## Security Considerations

- Never commit API keys; use `settings` table for Polygon.io key storage
- Financial precision: Use `REAL` type in SQLite, not integers for monetary values
- Input validation: All form inputs validated before database operations
- SQL injection prevention: Always use prepared statements, never string concatenation

## Key Patterns & Conventions

### Financial Calculations
```go
// Options commission is always $0.65 per contract
const OptionCommissionPerContract = 0.65

// P&L calculation includes commission deductions
profit := (premium - exitPrice) * float64(contracts) * 100 - commission
```

### Service Method Signatures
```go
// ID-based operations (preferred)
func (s *OptionService) GetByID(id int) (*Option, error)
func (s *OptionService) Update(id int, updates map[string]interface{}) error

// Compound key fallbacks (legacy compatibility)
func (s *OptionService) GetByCompoundKey(symbol, type string, opened time.Time, ...) (*Option, error)
```

### Database Query Patterns
```go
// Always use prepared statements
query := `UPDATE options SET exit_price = ?, closed = ? WHERE id = ?`
_, err := s.db.Exec(query, exitPrice, time.Now(), optionID)

// RETURNING clause for SQLite INSERT/UPDATE
query := `INSERT INTO symbols (symbol, price) VALUES (?, ?) RETURNING symbol, price, created_at`
err := s.db.QueryRow(query, symbol, price).Scan(&sym.Symbol, &sym.Price, &sym.CreatedAt)
```

### HTML Template Helpers
```go
// Custom template functions registered in server.go
funcMap := template.FuncMap{
    "groupByExpiration": groupPositionsByExpiration, // Groups options by expiration date
    "add": func(a, b interface{}) interface{} { ... }, // Safe addition for mixed types
}
```

### Wheel Strategy State Transitions
```go
// 1. Sell cash-secured put → Option record created (type="Put")
// 2. Put assigned → LongPosition created, Treasury balance decreased
// 3. Sell covered call → Option record created (type="Call")
// 4. Call assigned → LongPosition closed, Treasury balance increased
```

## Important Context

### Financial Domain Knowledge
- **Option multiplier**: 1 contract = 100 shares (always multiply by 100 for cash calculations)
- **Strike semantics**: Put assignment happens when stock price < strike; Call assignment when price > strike
- **Premium collection**: Income earned upfront when selling options (both puts and calls)
- **Collateral mechanics**: Treasury amounts dynamically adjust as options are assigned

### Project-Specific Quirks
- Module name is `stonks` in `go.mod`, but project is called "Wheeler"
- Database path resolution uses `./data/currentdb` file to track active database
- CGO must be enabled for SQLite3 driver: `CGO_ENABLED=1 go build`
- Port 8080 for development, 8077 for Docker (see `docker-compose.yml`)

## Project Structure

```
wheeler/
├── main.go                           # Web dashboard application entry point
├── model.md                          # Data model specification
├── CLAUDE.md                         # Development guidance for AI assistants
├── README.md                         # Project documentation
├── LICENSE                          # MIT License
├── Makefile                         # Build automation
├── go.mod                           # Go module dependencies
├── go.sum                           # Go module checksums
├── wheeler                          # Compiled binary
├── bin/                             # Binary output directory
├── data/                            # Database storage directory
│   ├── currentdb                    # Current database tracker
│   ├── *.db                         # SQLite database files
│   └── backups/                     # Database backup directory
├── screenshots/                     # Application screenshots for documentation
│   ├── dashboard.png                # Dashboard interface
│   ├── monthly.png                  # Monthly analysis view
│   ├── options.png                  # Options trading interface
│   ├── treasuries.png               # Treasury management
│   ├── symbol.png                   # Individual symbol analysis
│   ├── import.png                   # CSV import tools
│   ├── database.png                 # Database management
│   └── polygon.png                  # Polygon.io integration
├── internal/
│   ├── database/
│   │   ├── db.go                    # Database connection and setup
│   │   ├── schema.sql               # Complete SQLite schema
│   │   └── wheel_strategy_example.sql # Test data for tutorials
│   ├── models/
│   │   ├── symbol.go                # Symbol entity and service
│   │   ├── option.go                # Options tracking with Put/Call
│   │   ├── long_position.go         # Stock position management
│   │   ├── dividend.go              # Dividend payment tracking
│   │   ├── treasury.go              # Treasury securities management
│   │   └── setting.go               # Application settings model
│   ├── polygon/                     # Polygon.io API integration
│   │   ├── client.go                # HTTP client for Polygon.io API
│   │   ├── service.go               # Service layer for market data
│   │   └── live_integration_test.go # Integration tests with live API
│   └── web/
│       ├── server.go                # Web server setup and routing
│       ├── handlers.go              # Main page handlers
│       ├── dashboard_handlers.go    # Dashboard specific handlers
│       ├── monthly_handlers.go      # Monthly analysis handlers
│       ├── options_handlers.go      # Options trading handlers
│       ├── symbol_handlers.go       # Symbol page handlers
│       ├── position_handlers.go     # Position management handlers
│       ├── treasury_handlers.go     # Treasury management handlers
│       ├── import_handlers.go       # Import/backup/database handlers
│       ├── polygon_handlers.go      # Polygon.io integration handlers
│       ├── settings_handlers.go     # Settings management handlers
│       ├── utility_handlers.go      # Utility functions and helpers
│       ├── types.go                 # Web data types and structures
│       ├── templates/               # HTML templates with Go templating
│       │   ├── _symbol_modal.html   # Shared symbol modal component
│       │   ├── dashboard.html       # Main dashboard with interactive charts
│       │   ├── monthly.html         # Monthly performance analysis
│       │   ├── options.html         # Options trading interface with scatter plots
│       │   ├── treasuries.html      # Treasury securities management
│       │   ├── symbol.html          # Individual symbol analysis with summary metrics
│       │   ├── help.html            # Tabbed help system with tutorials
│       │   ├── backup.html          # Database management interface
│       │   ├── import.html          # CSV import tools with validation
│       │   └── settings.html        # Polygon.io API configuration
│       └── static/                  # Static web assets
│           ├── assets/              # Static asset files
│           ├── css/
│           │   └── styles.css       # Application styling (dark theme)
│           └── js/                  # JavaScript modules
│               ├── navigation.js    # Navigation and sidebar functionality
│               ├── symbol-modal.js  # Symbol modal interactions
│               └── table-sort.js    # Table sorting functionality
```

## Database Schema Design

The database schema follows modern best practices for web applications:

### Primary Key Strategy
- **Transactional Tables**: Use `INTEGER PRIMARY KEY AUTOINCREMENT` for easier HTTP CRUD operations
  - `options.id`, `long_positions.id`, `dividends.id`
- **Reference Tables**: Use natural primary keys for business identifiers  
  - `symbols.symbol` (stock ticker), `treasuries.cuspid` (bond identifier)

### Data Integrity
- **Unique Constraints**: Prevent duplicate business records via composite unique indexes
- **Foreign Keys**: Enforce referential integrity (all tables reference `symbols.symbol`)
- **Check Constraints**: Validate option types (`'Put'` or `'Call'`)

### Schema Migration
- **`internal/database/schema.sql`**: Single source of truth for database structure
- **No Migration Files**: Removed legacy migration files; schema.sql is authoritative
- **Automatic Setup**: Database tables created via `CREATE TABLE IF NOT EXISTS`

## Database Management

Wheeler supports multiple SQLite databases for different portfolios or environments:

### Database Operations
- **Current Database**: Tracked in `./data/currentdb` file
- **Database Storage**: All `.db` files stored in `./data/` directory
- **Create Database**: Admin → Database page or API endpoint
- **Switch Database**: Change active database via web interface
- **Delete Database**: Remove unused databases with confirmation
- **Backup System**: Manual backups to `./data/backups/` with timestamps

### Test Data and Tutorials
- **Wheel Strategy Example**: Complete trading history demonstrating 73% annual returns
- **Generate Test Data**: One-click import via Help → Tutorial page
- **SQL Location**: Test data stored in `internal/database/wheel_strategy_example.sql`
- **Treasury Operations**: Realistic collateral management examples
- **Data Reset**: Switch databases or delete test database to start fresh

To get started, use `go run main.go`, visit http://localhost:8080/help, switch to the Tutorial tab, and click "Generate Test Data" to see a complete wheel strategy implementation.

## Common Tasks

### Adding New Domain Entity
1. Create model struct in `internal/models/{entity}.go` with service pattern
2. Add table schema to `internal/database/schema.sql`
3. Implement CRUD methods: `Create`, `GetByID`, `Update`, `Delete`, `GetAll`
4. Add handler in `internal/web/{entity}_handlers.go` with API endpoints
5. Create HTML template in `internal/web/templates/{entity}.html`

### Adding API Endpoint
```go
// In server.go
func (s *Server) setupRoutes() {
    http.HandleFunc("/api/entity/{id}", s.handleEntityAPI)
}

// Handler pattern (in handlers.go)
func (s *Server) handleEntityAPI(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet: /* ... */
    case http.MethodPut: /* ... */
    case http.MethodDelete: /* ... */
    }
}
```

### Running Integration Tests
```bash
# Set API key via environment or settings table
export POLYGON_API_KEY="your_key_here"
go test -v ./internal/polygon/
```

### Adding a New Data Provider
1. Create a new provider implementation in `internal/providers/{provider_name}.go`
2. Implement the `Provider` interface with all required methods:
   - `GetQuote(ctx, symbol)` - Retrieve current quote data
   - `GetTickerDetails(ctx, symbol)` - Get detailed ticker information
   - `GetDividends(ctx, symbol, limit)` - Fetch dividend history
   - `ValidateConnection(ctx)` - Test provider connectivity
   - `Name()` - Return provider name
3. Update `getProvider()` in `internal/providers/service.go` to support the new provider type
4. Add settings entries for the new provider:
   - `DATA_PROVIDER_TYPE` - Set to your provider name (e.g., "alphavantage")
   - `{PROVIDER_NAME}_API_KEY` - API key for the new provider
5. Update `GetAPIKeyStatus()` to handle the new provider's API key validation
6. Add rate limiting logic in `GetRateLimitDelay()` if needed
7. Create tests in `test/providers_test.go` for the new provider

Example provider structure:
```go
type NewProvider struct {
    apiKey string
    client *http.Client
}

func NewNewProvider(apiKey string) *NewProvider {
    return &NewProvider{
        apiKey: apiKey,
        client: &http.Client{Timeout: 30 * time.Second},
    }
}

func (p *NewProvider) Name() string {
    return "NewProvider"
}

// Implement remaining Provider interface methods...
```

## References
- `.github/copilot-instructions.md`: AI agent development instructions
- `CLAUDE.md`: AI agent development instructions
- `model.md`: Complete database schema specification
- `README.md`: User-facing documentation and quick start
- `internal/database/schema.sql`: Authoritative database structure