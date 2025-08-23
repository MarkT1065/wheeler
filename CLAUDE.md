# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is "Wheeler" - a comprehensive financial portfolio tracking system built with Go. The project specializes in tracking sophisticated options trading strategies, particularly the "wheel strategy" (cash-secured puts, covered calls, and stock assignments), along with comprehensive portfolio management including Treasury securities collateral management.

## Applications

The project contains two main applications:

1. **Web Dashboard** (`main.go`) - Modern web interface for comprehensive portfolio tracking
2. **OFX Parser** (`ofx_parser.go`) - CLI utility for parsing financial data files from brokers

## Development Commands

### Web Dashboard (Primary Application)
- **Run web dashboard**: `go run main.go` (starts on http://localhost:8080)
- **Build web dashboard**: `go build . && ./stonks`

### OFX Parser (Financial Data Import)
- **Run OFX parser**: `go run ofx_parser.go [directory] [output_file]`
- **Parse QuickenWin files**: `go run ofx_parser.go ./data/ transactions.json`

### Database Operations
- **Load test data**: Use the Generate Test Data buttons in Help → Tutorial
- **Multiple databases**: Create/switch databases via Admin → Database
- **Database backups**: Manual backups via Admin → Database page

## Dependencies

- Go 1.19+ with modules support
- SQLite3 (`github.com/mattn/go-sqlite3`)
- No GTK dependencies required for web dashboard
- Web technologies: HTML5, CSS3, JavaScript, Chart.js

## Architecture Notes

Key architectural decisions and patterns:

- **Primary Stack**: Go + Web Frontend + SQLite database
- **Data Models**: Symbols, options, long positions, dividends, treasuries (see model.md)
- **Database Design**: 
  - INTEGER PRIMARY KEY AUTOINCREMENT for transactional data (options, long_positions, dividends)
  - Natural primary keys for reference data (symbols.symbol, treasuries.cuspid)
  - UNIQUE indexes to prevent duplicate business records
  - Foreign key relationships for data integrity
- **Web Interface**: HTML templates with Chart.js visualizations and AJAX APIs
- **Service Layer**: ID-based CRUD operations with compound key fallbacks for compatibility
- **API Design**: RESTful endpoints using integer IDs for easier HTTP operations

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

### API Endpoints
- `/api/symbols/{symbol}` - Symbol CRUD operations and price updates
- `/api/options` - Options management (create, update, assign, close)
- `/api/long-positions` - Stock position lifecycle management
- `/api/dividends` - Dividend tracking and yield calculations
- `/api/treasuries/{cuspid}` - Treasury operations and interest tracking
- `/api/allocation-data` - Portfolio allocation data for charts
- `/api/generate-test-data` - Test data generation for tutorials

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

## Security Considerations

- Never commit API keys for market data providers
- Validate all financial calculations with appropriate precision
- Follow security best practices for handling financial data
- Use prepared statements for SQL operations to prevent injection
- Validate form inputs before database operations

## Project Structure

```
stonks/
├── main.go                    # Web dashboard application entry point
├── ofx_parser.go             # OFX financial data file parser
├── model.md                   # Data model specification
├── demo_data.sql             # Sample data for testing
├── go.mod                    # Go module dependencies
├── wheeler.db                 # SQLite database file
├── internal/
│   ├── database/
│   │   ├── db.go             # Database connection and setup
│   │   └── schema.sql        # SQLite schema with relationships
│   ├── models/
│   │   ├── symbol.go         # Symbol entity and service
│   │   ├── option.go         # Options tracking with Put/Call
│   │   ├── long_position.go  # Stock position management
│   │   ├── dividend.go       # Dividend payment tracking
│   │   └── treasury.go       # Treasury securities management
│   └── web/
│       ├── server.go         # Web server setup and routing
│       ├── handlers.go       # HTTP handlers and API endpoints
│       ├── templates/        # HTML templates
│       │   ├── dashboard.html
│       │   ├── monthly.html
│       │   ├── treasuries.html
│       │   └── symbol.html
│       └── static/           # Static web assets
│           ├── css/
│           ├── js/
│           └── images/
└── README.md                 # Project documentation
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