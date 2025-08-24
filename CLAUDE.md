# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is "Wheeler" - a comprehensive financial portfolio tracking system built with Go. The project specializes in tracking sophisticated options trading strategies, particularly the "wheel strategy" (cash-secured puts, covered calls, and stock assignments), along with comprehensive portfolio management including Treasury securities collateral management.

## Applications

The project contains one main application:

1. **Web Dashboard** (`main.go`) - Modern web interface for comprehensive portfolio tracking

## Development Commands

### Web Dashboard (Primary Application)
- **Run web dashboard**: `go run main.go` (starts on http://localhost:8080)
- **Build web dashboard**: `go build . && ./wheeler`

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
- `/api/polygon/*` - Polygon.io API integration endpoints

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

- Never commit API keys for market data providers
- Validate all financial calculations with appropriate precision
- Follow security best practices for handling financial data
- Use prepared statements for SQL operations to prevent injection
- Validate form inputs before database operations

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