# Wheeler - It's better than a spreadsheet! 

A web-based portfolio tracking system built with Go, specializing in options trading strategies (particularly the "wheel strategy"), Treasury collateral management, and comprehensive portfolio analytics with fancy visualizations.

⚠️ Disclaimer

This tool is for educational and research purposes only. Wheeler does not provide investment advice, recommendations, or financial guidance. All information provided is for informational purposes only and should not be considered as investment advice. Always consult with a qualified financial advisor before making any investment decisions. Past performance does not guarantee future results. Trading and investing involve risk of loss.

You shouldn't really use Wheeler for anything at any time.

## The Wheel

The Wheel is the "triple income" strategy: option premiums, capital gains, and dividends.

Wheeler helps track options and trades. Once tracked in a database, it's easy to present that data in many meaningful ways.

### Dashboard

The Dashboard shows overall progress and total Longs, Put Exposure, and Treasuries for visually measuring risk.

![Dashboard](./screenshots/dashboard.png)

### Monthly

The Monthly view shows gains over time by income type by month with the goal of doing more of what works well.  The user can toggle between cumulative and monthly views.

![Monthly](./screenshots/monthly.png)

### Options

The Options view shows what trades are nearing expiration.

![Options](./screenshots/options.png)

### Treasuries

The Treasuries view manages any bonds and bills used for collateral.

![Treasuries](./screenshots/treasuries.png)

### Symbols

The Symbols view is a total return view of one symbol, including Options, Stock, and Dividends.

![Symbol](./screenshots/symbol.png)

## Managing Wheeler

### Import

Wheeler's simple data model allows CSV import of Options, Stocks, and Dividends.
 
![Import](./screenshots/import.png)

### Database

The Database view manages the Wheeler datastore. SQLite is used and it's a single file.

![Database](./screenshots/database.png)

### Polygon

The Polygon view allows configuration of Polygon.io API and sync'ing of data. The free tier is used to get current price and other data.

![Polygon](./screenshots/polygon.png)


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

## Docker Setup

Wheeler can be run with Docker Compose for easy deployment.

### Quick Start with Docker

```bash
# Clone and navigate to project
cd wheeler

# Build and start Wheeler
docker compose up -d

# Open your browser to:
# http://localhost:8077
```

### Unraid Setup

Wheeler is designed to be Unraid-compatible:

1. **Create app directory** (if using custom path):
   ```bash
   mkdir -p /mnt/user/appdata/wheeler
   ```

2. **Edit docker-compose.yml** to customize the volume path:
   ```yaml
   volumes:
     - /mnt/user/appdata/wheeler:/app/data
   ```

3. **Start Wheeler**:
   ```bash
   docker compose up -d
   ```

4. **Access Wheeler** at `http://your-unraid-ip:8077`

### Docker Configuration

| Setting | Value | Description |
|---------|-------|-------------|
| External Port | 8077 | Web interface port (avoids conflict with SABnzbd on 8080) |
| Internal Port | 8080 | Go web server port |
| Data Volume | `/app/data` | SQLite database and backups |
| Restart Policy | `unless-stopped` | Auto-restart on failure |

### Customization

Edit `docker-compose.yml` to customize:

- **Port**: Change `8077:8080` to use a different external port
- **Volume**: Change `./data:/app/data` to your preferred data path
- **Environment**: Uncomment and set `POLYGON_API_KEY` for market data

### Stopping Wheeler

```bash
docker compose down
```

## Application Overview

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
├── model.md                          # Data model specification
├── CLAUDE.md                         # Development guidance
├── README.md                         # This documentation
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
├── screenshots/                     # Application screenshots
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
│   │   ├── option.go                # Options tracking service
│   │   ├── long_position.go         # Stock position management
│   │   ├── dividend.go              # Dividend payment tracking
│   │   ├── treasury.go              # Treasury securities management
│   │   └── setting.go               # Application settings
│   ├── polygon/                     # Polygon.io API integration
│   │   ├── client.go                # API client
│   │   ├── service.go               # Service layer
│   │   └── live_integration_test.go # Integration tests
│   └── web/
│       ├── server.go                # Web server and routing
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
│       ├── utility_handlers.go      # Utility functions
│       ├── types.go                 # Web data types and structures
│       ├── templates/               # HTML templates
│       │   ├── _symbol_modal.html   # Shared symbol modal component
│       │   ├── dashboard.html       # Main dashboard with charts
│       │   ├── monthly.html         # Monthly performance analysis
│       │   ├── options.html         # Options trading interface
│       │   ├── treasuries.html      # Treasury management
│       │   ├── symbol.html          # Individual symbol analysis
│       │   ├── help.html            # Tabbed help system
│       │   ├── backup.html          # Database management
│       │   ├── import.html          # CSV import tools
│       │   └── settings.html        # Polygon.io configuration
│       └── static/                  # Static web assets
│           ├── assets/              # Static asset files
│           ├── css/
│           │   └── styles.css       # Application styling
│           └── js/                  # JavaScript modules
│               ├── navigation.js    # Navigation functionality
│               ├── symbol-modal.js  # Symbol modal interactions
│               └── table-sort.js    # Table sorting functionality
```

## Getting Help

Wheeler includes comprehensive documentation:

- **Built-in Tutorial**: Navigate to Help → Tutorial for interactive examples
- **Test Data**: Generate realistic trading scenarios to explore features
- **Code Documentation**: See `CLAUDE.md` for development guidance
- **Data Model**: See `model.md` for database schema details

## License

This project is open source and available under the MIT License.