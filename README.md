# Wheeler - Advanced Options Trading Portfolio System

A sophisticated web-based portfolio tracking system built with Go, specializing in options trading strategies (particularly the "wheel strategy"), Treasury collateral management, and comprehensive portfolio analytics with interactive Chart.js visualizations.

⚠️ Disclaimer

This tool is for educational and research purposes only. Wheeler does not provide investment advice, recommendations, or financial guidance. All information provided is for informational purposes only and should not be considered as investment advice. Always consult with a qualified financial advisor before making any investment decisions. Past performance does not guarantee future results. Trading and investing involve risk of loss.

## Features

### Core Portfolio Management
- **Modern Web Dashboard**: Responsive interface with real-time charts and metrics
- **Wheel Strategy Support**: Complete lifecycle tracking for cash-secured puts and covered calls
- **Treasury Collateral Management**: Automatic adjustment of Treasury positions based on option assignments
- **Multiple Database Support**: Create separate databases for different portfolios or testing environments
- **Test Data Generation**: One-click import of realistic trading scenarios

### Trading and Analytics
- **Comprehensive Options Tracking**: Put/Call options with full lifecycle management (opening → assignment/expiration)
- **Long Stock Positions**: Entry/exit tracking with cost basis and P&L calculations
- **Dividend Management**: Payment recording and yield analysis
- **Treasury Securities**: Bond management with yields, maturity tracking, and interest payments
- **Performance Analytics**: Monthly breakdowns, allocation analysis, and risk metrics

### User Experience
- **Interactive Charts**: Real-time pie charts for positions, allocations, and risk exposure
- **Tabbed Help System**: Wheeler Help and Tutorial with realistic trading examples
- **CSV Import Tools**: Bulk import for options, stocks, and dividends
- **Database Management**: Create, switch, backup, and delete databases via web interface

## Quick Start

```bash
# Clone and navigate to project
cd wheeler

# Run the web application
go run main.go

# Open your browser to:
# http://localhost:8080
```

### Getting Started with Test Data

1. Navigate to **Help** → **Tutorial** tab
2. Click **"Generate Test Data"** to import a complete wheel strategy example
3. Explore the dashboard to see realistic portfolio tracking in action
4. View the tutorial content to understand the trading strategy

## Prerequisites

- Go 1.19 or later
- SQLite3 (automatically included via Go driver)
- Modern web browser

## Application Overview

### Main Pages

- **Dashboard** (`/`) - Portfolio overview with charts and performance metrics
- **Monthly** (`/monthly`) - Month-by-month performance analysis
- **Options** (`/options`) - Detailed options positions and trading history
- **Treasuries** (`/treasuries`) - Treasury securities collateral management
- **Symbol Pages** (`/symbol/{SYMBOL}`) - Individual stock analysis and history
- **Import** (`/import`) - CSV data import tools
- **Admin** (`/backup`) - Database management and backups
- **Help** (`/help`) - Wheeler Help and Tutorial system

### Wheel Strategy Implementation

Wheeler specializes in tracking sophisticated options strategies:

- **Cash-Secured Puts**: Sell puts with Treasury collateral, track assignments
- **Covered Calls**: Sell calls against stock positions, track exercises  
- **Stock Assignments**: Automatic conversion of expired puts to stock positions
- **Premium Collection**: Track income from option premiums across all strategies
- **Position Scaling**: Support for increasing position sizes over time
- **Treasury Integration**: Automatic collateral adjustment based on option assignments

### Key Features

1. **Real-time Calculations**: Automatic P&L, allocation, and risk metrics
2. **Portfolio Allocation Charts**: Visual representation of stocks vs. treasuries vs. cash
3. **Options Risk Visualization**: Put exposure and covered call tracking
4. **Treasury Collateral Management**: Dynamic balance adjustment based on option activity
5. **Multiple Portfolio Support**: Separate databases for different trading accounts
6. **Historical Analysis**: Monthly performance breakdowns and trend analysis

## Database Management

Wheeler supports multiple SQLite databases:

### Database Operations
- **Current Database**: Active database tracked in `./data/currentdb`
- **Storage Location**: All `.db` files stored in `./data/` directory
- **Create/Switch/Delete**: Full database lifecycle management via web interface
- **Backup System**: Manual backups to `./data/backups/` with timestamps

### Database Schema
- **Symbols Table**: Stock symbols with prices, dividends, P/E ratios (`symbols.symbol` PK)
- **Options Table**: Put/Call tracking with integer IDs (`options.id` PK)
- **Long Positions Table**: Stock holdings with entry/exit tracking (`long_positions.id` PK)
- **Dividends Table**: Payment records (`dividends.id` PK)
- **Treasuries Table**: Securities with CUSPID, yields, maturity (`treasuries.cuspid` PK)
- **Foreign Key Relationships**: All tables reference `symbols.symbol` for data integrity

## API Endpoints

Wheeler provides comprehensive RESTful APIs:

- `GET/PUT /api/symbols/{symbol}` - Symbol operations and price updates
- `GET/POST/PUT/DELETE /api/options` - Options management with lifecycle tracking
- `GET/POST/PUT/DELETE /api/long-positions` - Stock position management
- `GET/POST/PUT/DELETE /api/dividends` - Dividend tracking and calculations
- `GET/POST/PUT/DELETE /api/treasuries/{cuspid}` - Treasury operations
- `GET /api/allocation-data` - Portfolio allocation data for charts
- `POST /api/generate-test-data` - Test data generation for tutorials

## Project Structure

```
wheeler/
├── main.go                           # Web application entry point
├── ofx_parser.go                     # OFX financial data parser utility
├── model.md                          # Data model specification
├── CLAUDE.md                         # Development guidance
├── README.md                         # This documentation
├── go.mod                           # Go module dependencies
├── data/                            # Database storage directory
│   ├── currentdb                    # Current database tracker
│   ├── *.db                         # SQLite database files
│   └── backups/                     # Database backup directory
├── internal/
│   ├── database/
│   │   ├── db.go                    # Database connection and setup
│   │   ├── schema.sql               # Complete SQLite schema
│   │   └── wheel_strategy_example.sql # Test data for tutorials
│   ├── models/
│   │   ├── symbol.go                # Symbol entity and service
│   │   ├── option.go                # Options tracking service
│   │   ├── long_position.go         # Stock position management
│   │   ├── dividend.go              # Dividend payment tracking
│   │   └── treasury.go              # Treasury securities management
│   └── web/
│       ├── server.go                # Web server and routing
│       ├── handlers.go              # Main page handlers
│       ├── import_handlers.go       # Import/backup/database handlers
│       ├── templates/               # HTML templates
│       │   ├── dashboard.html       # Main dashboard with charts
│       │   ├── monthly.html         # Monthly performance analysis
│       │   ├── options.html         # Options trading interface
│       │   ├── treasuries.html      # Treasury management
│       │   ├── symbol.html          # Individual symbol analysis
│       │   ├── help.html            # Tabbed help system
│       │   ├── backup.html          # Database management
│       │   └── import.html          # CSV import tools
│       └── static/                  # Static web assets
│           ├── css/styles.css       # Application styling
│           └── js/                  # JavaScript modules
```

## Development

### Building and Running
```bash
# Build the application
go build .

# Run with development logging
go run main.go

# Access the application
open http://localhost:8080
```

### Testing with Real Data
```bash
# Generate test data via web interface (recommended)
# Navigate to Help → Tutorial → Generate Test Data

# Or manually load test data
sqlite3 data/wheeler.db < internal/database/wheel_strategy_example.sql
```

### Architecture

Wheeler follows modern web application patterns:

- **Service Layer Architecture**: Each model has a dedicated service for CRUD operations
- **RESTful API Design**: Clean HTTP endpoints using integer IDs for web-friendly operations
- **Server-Side Rendering**: HTML templating with real-time data injection
- **Interactive Frontend**: Chart.js visualizations with AJAX data loading
- **Responsive Design**: Mobile-friendly interface using CSS Grid and Flexbox
- **Multiple Database Support**: Dynamic database switching for portfolio management

## Advanced Usage

### Wheel Strategy Workflow

1. **Setup Treasury Collateral**: Add Treasury securities as cash collateral
2. **Sell Cash-Secured Puts**: Create put positions with Treasury backing
3. **Assignment Handling**: Treasury balances automatically adjust on assignment
4. **Stock Management**: Assigned shares become long positions
5. **Covered Calls**: Sell calls against stock positions
6. **Performance Tracking**: Monitor premium income and portfolio growth

### Portfolio Management

- **Multiple Environments**: Create separate databases for live trading vs. paper trading
- **Risk Management**: Monitor put exposure and covered call obligations
- **Performance Analysis**: Track monthly performance and allocation changes
- **Treasury Optimization**: Balance cash collateral with yield optimization

### Data Import and Export

- **CSV Import**: Bulk import options, stocks, and dividends via web interface
- **OFX Parser**: Parse broker files with `go run ofx_parser.go [directory] [output]`
- **Database Backups**: Create timestamped backups via Admin → Database
- **Data Portability**: SQLite files can be copied between systems

## Financial Concepts

Wheeler implements sophisticated financial tracking:

- **Options Greeks**: While not calculated, position data supports external analysis
- **Collateral Management**: Treasury securities automatically adjust based on option activity
- **Yield Tracking**: Monitor income from dividends, interest, and option premiums
- **Risk Metrics**: Put exposure visualization and covered call obligation tracking
- **Portfolio Allocation**: Dynamic allocation including stocks, treasuries, and cash

## Getting Help

Wheeler includes comprehensive documentation:

- **Built-in Tutorial**: Navigate to Help → Tutorial for interactive examples
- **Test Data**: Generate realistic trading scenarios to explore features
- **Code Documentation**: See `CLAUDE.md` for development guidance
- **Data Model**: See `model.md` for database schema details

## License

This project is open source and available under the MIT License.