# Wheeler Development Instructions for AI Agents

## Project Context
Wheeler is a Go web application for tracking options trading strategies (particularly "wheel strategy"), with SQLite persistence and Polygon.io market data integration. The core domain involves sophisticated financial tracking: cash-secured puts, covered calls, stock assignments, dividends, and Treasury collateral management.

## Architecture Overview

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
- SQL migrations live in `internal/database/migrations/*.sql` and are embedded via `//go:embed`.
- Migrations are the source of truth for schema changes; add new migrations for any schema update instead of modifying existing migrations or relying on ad-hoc `CREATE TABLE IF NOT EXISTS`.
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
- Integration tests require live Polygon.io API key: `internal/polygon/live_integration_test.go`
- Test databases automatically cleaned: `rm -f test_*.db`

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

### Security Considerations
- Never commit API keys; use `settings` table for Polygon.io key storage
- Financial precision: Use `REAL` type in SQLite, not integers for monetary values
- Input validation: All form inputs validated before database operations
- SQL injection prevention: Always use prepared statements, never string concatenation

### Project-Specific Quirks
- Module name is `stonks` in `go.mod`, but project is called "Wheeler"
- Database path resolution uses `./data/currentdb` file to track active database
- CGO must be enabled for SQLite3 driver: `CGO_ENABLED=1 go build`
- Port 8080 for development, 8077 for Docker (see `docker-compose.yml`)

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

## References
- `CLAUDE.md`: Comprehensive development guidance (250+ lines)
- `model.md`: Complete database schema specification
- `README.md`: User-facing documentation and quick start
- `internal/database/schema.sql`: Authoritative database structure
