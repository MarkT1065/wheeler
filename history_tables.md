# Historical Market Data — Implementation Plan

## 1. Massive.com (formerly Polygon.io) Free Tier API Capabilities

The API base URL remains `https://api.polygon.io`. The free tier provides 5 API calls per minute with end-of-day (delayed) data. All endpoints below are confirmed available on the free tier.

### 1a. Stocks — All US Stocks + Historical Data

| Endpoint | Path | Description |
|---|---|---|
| **Aggregates (Bars)** | `GET /v2/aggs/ticker/{ticker}/range/{multiplier}/{timespan}/{from}/{to}` | Historical OHLCV bars. Timespan: minute, hour, day, week, month, quarter, year. Returns: open, high, low, close, volume, vwap, timestamp, transaction count. Limit 50000 bars per request. |
| **Previous Close** | `GET /v2/aggs/ticker/{ticker}/prev` | Previous trading day's OHLCV. Already implemented in `client.go`. |
| **Daily Open/Close** | `GET /v1/open-close/{ticker}/{date}` | Single-day OHLCV for a specific date. |
| **Ticker Details** | `GET /v3/reference/tickers/{ticker}` | Company info, market cap, employees, description, SIC code. Already implemented. |
| **Tickers List** | `GET /v3/reference/tickers` | List/search all tickers with filtering by market, type, exchange. |
| **Market Status** | `GET /v1/marketstatus/now` | Current market open/closed status. Already used for API key validation. |

### 1b. Corporate Actions — Dividends (confirmed available on free tier)

| Endpoint | Path | Description |
|---|---|---|
| **Dividends** | `GET /v3/reference/dividends` | Historical dividend records. Fields: cash_amount, ex_dividend_date, pay_date, record_date, declaration_date, distribution_type, frequency, ticker. Data back to Jan 2000. Already implemented in `client.go`. |
| **Stock Splits** | `GET /v3/reference/splits` | Historical split events. Fields: ticker, execution_date, split_from, split_to. |
| **Ticker Events** | `GET /vX/reference/tickers/{id}/events` | Timeline of corporate events for a ticker. |

### 1c. Options — All US Option Tickers + Historical Data

| Endpoint | Path | Description |
|---|---|---|
| **Options Aggregates (Bars)** | `GET /v2/aggs/ticker/{optionsTicker}/range/{multiplier}/{timespan}/{from}/{to}` | Historical OHLCV bars for an option contract. Same structure as stock aggregates. |
| **Options Previous Close** | `GET /v2/aggs/ticker/{optionsTicker}/prev` | Previous day's bar for an option contract. |
| **Options Contracts** | `GET /v3/reference/options/contracts` | List option contracts. Filter by: underlying_ticker, contract_type (call/put), expiration_date, strike_price, expired (include expired contracts). Data back to June 2014. |
| **Options Contract Detail** | `GET /v3/reference/options/contracts/{options_ticker}` | Single contract details: ticker, underlying_ticker, contract_type, strike_price, expiration_date, exercise_style, shares_per_contract, primary_exchange. |
| **Options Snapshot** | `GET /v3/snapshot/options/{underlyingAsset}/{optionContract}` | Current snapshot with greeks, IV, open interest, last quote/trade. Already implemented in `client.go`. |

**Option ticker format**: `O:{TICKER}{YYMMDD}{C/P}{STRIKE}` — e.g., `O:AAPL250620C00150000`

### 1d. Indices — Limited Index Tickers + Historical Data

| Endpoint | Path | Description |
|---|---|---|
| **Index Aggregates (Bars)** | `GET /v2/aggs/ticker/{indicesTicker}/range/{multiplier}/{timespan}/{from}/{to}` | Historical OHLC for indices. No volume data (indices don't have volume). Data available since Feb 2023. |
| **Index Previous Close** | `GET /v2/aggs/ticker/{indicesTicker}/prev` | Previous day's bar for an index. |

**Index ticker format**: `I:{INDEX}` — e.g., `I:SPX`, `I:NDX`, `I:DJI`, `I:VIX`

**Relevant indices for Wheeler**:
- `I:SPX` — S&P 500
- `I:NDX` — Nasdaq 100
- `I:DJI` — Dow Jones Industrial Average
- `I:VIX` — CBOE Volatility Index (important for options pricing context)

### Rate Limiting Strategy

Free tier: **5 calls/minute** = 1 call every 12 seconds.

The background worker must enforce this globally. With ~34 symbols, fetching daily bars for all takes ~7 minutes per cycle. Options and indices add to the queue. The worker processes a prioritized FIFO queue and sleeps 12s between calls.

---

## 2. New Models

### 2a. `internal/models/price_history.go` — Stock Price History

```go
type PriceBar struct {
    ID            int       `json:"id"`
    Symbol        string    `json:"symbol"`
    BarDate       time.Time `json:"bar_date"`
    Open          float64   `json:"open"`
    High          float64   `json:"high"`
    Low           float64   `json:"low"`
    Close         float64   `json:"close"`
    Volume        float64   `json:"volume"`
    VWAP          float64   `json:"vwap"`
    Transactions  int       `json:"transactions"`
    CreatedAt     time.Time `json:"created_at"`
}

type PriceHistoryService struct {
    db *sql.DB
}
```

**Methods**: `Upsert(bar)`, `GetBySymbol(symbol, from, to)`, `GetLatestDate(symbol) time.Time`, `DeleteOlderThan(days int)`

### 2b. `internal/models/option_history.go` — Option Contract Price History

```go
type OptionBar struct {
    ID              int       `json:"id"`
    OptionTicker    string    `json:"option_ticker"`
    UnderlyingSymbol string  `json:"underlying_symbol"`
    ContractType    string    `json:"contract_type"`
    StrikePrice     float64   `json:"strike_price"`
    ExpirationDate  time.Time `json:"expiration_date"`
    BarDate         time.Time `json:"bar_date"`
    Open            float64   `json:"open"`
    High            float64   `json:"high"`
    Low             float64   `json:"low"`
    Close           float64   `json:"close"`
    Volume          float64   `json:"volume"`
    VWAP            float64   `json:"vwap"`
    Transactions    int       `json:"transactions"`
    CreatedAt       time.Time `json:"created_at"`
}

type OptionHistoryService struct {
    db *sql.DB
}
```

**Methods**: `Upsert(bar)`, `GetByOptionTicker(ticker, from, to)`, `GetByUnderlying(symbol, from, to)`, `GetLatestDate(optionTicker) time.Time`

### 2c. `internal/models/index_history.go` — Index Price History

```go
type IndexBar struct {
    ID            int       `json:"id"`
    IndexTicker   string    `json:"index_ticker"`
    BarDate       time.Time `json:"bar_date"`
    Open          float64   `json:"open"`
    High          float64   `json:"high"`
    Low           float64   `json:"low"`
    Close         float64   `json:"close"`
    CreatedAt     time.Time `json:"created_at"`
}

type IndexHistoryService struct {
    db *sql.DB
}
```

**Methods**: `Upsert(bar)`, `GetByIndex(ticker, from, to)`, `GetLatestDate(ticker) time.Time`

**Note**: Indices have no volume or VWAP — only OHLC.

---

## 3. Database Migration

**File**: `internal/database/migrations/20260602000001_add_history_tables.sql`

```sql
-- Stock price history (daily bars)
CREATE TABLE IF NOT EXISTS price_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    bar_date DATE NOT NULL,
    open REAL NOT NULL,
    high REAL NOT NULL,
    low REAL NOT NULL,
    close REAL NOT NULL,
    volume REAL DEFAULT 0,
    vwap REAL DEFAULT 0,
    transactions INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (symbol) REFERENCES symbols(symbol)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_price_history_symbol_date
    ON price_history(symbol, bar_date);
CREATE INDEX IF NOT EXISTS idx_price_history_bar_date
    ON price_history(bar_date);

-- Option contract price history (daily bars)
CREATE TABLE IF NOT EXISTS option_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    option_ticker TEXT NOT NULL,
    underlying_symbol TEXT NOT NULL,
    contract_type TEXT NOT NULL CHECK (contract_type IN ('call', 'put')),
    strike_price REAL NOT NULL,
    expiration_date DATE NOT NULL,
    bar_date DATE NOT NULL,
    open REAL NOT NULL,
    high REAL NOT NULL,
    low REAL NOT NULL,
    close REAL NOT NULL,
    volume REAL DEFAULT 0,
    vwap REAL DEFAULT 0,
    transactions INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (underlying_symbol) REFERENCES symbols(symbol)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_option_history_ticker_date
    ON option_history(option_ticker, bar_date);
CREATE INDEX IF NOT EXISTS idx_option_history_underlying
    ON option_history(underlying_symbol);
CREATE INDEX IF NOT EXISTS idx_option_history_bar_date
    ON option_history(bar_date);
CREATE INDEX IF NOT EXISTS idx_option_history_expiration
    ON option_history(expiration_date);

-- Index price history (daily bars)
CREATE TABLE IF NOT EXISTS index_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    index_ticker TEXT NOT NULL,
    bar_date DATE NOT NULL,
    open REAL NOT NULL,
    high REAL NOT NULL,
    low REAL NOT NULL,
    close REAL NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_index_history_ticker_date
    ON index_history(index_ticker, bar_date);
CREATE INDEX IF NOT EXISTS idx_index_history_bar_date
    ON index_history(bar_date);

INSERT OR IGNORE INTO schema_migrations (version)
VALUES ('20260602000001_add_history_tables');
```

---

## 4. Background Worker — `internal/polygon/worker.go`

### Architecture

The worker is a standalone goroutine started by the `polygon.Service` at server boot. It owns all API call scheduling, rate limiting, and metrics. The web handlers no longer make API calls directly — they read data the worker has already fetched and stored.

```
main.go
  └─ web.NewServer()
       └─ polygon.NewService(...)
            └─ service.StartWorker(ctx)   // goroutine, runs until ctx cancelled
                 ├─ priceUpdateCycle()     // update symbol prices (existing)
                 ├─ dividendFetchCycle()   // fetch dividends (existing)
                 ├─ priceHistoryBackfill() // new: backfill daily bars
                 ├─ optionHistoryFetch()   // new: fetch option bars
                 └─ indexHistoryFetch()    // new: fetch index bars
```

### Worker State (in-memory, thread-safe)

```go
type WorkerStatus struct {
    Running         bool          `json:"running"`
    StartedAt       time.Time     `json:"started_at"`
    Uptime          string        `json:"uptime"`
    TotalAPICalls   int64         `json:"total_api_calls"`
    AvgResponseTime time.Duration `json:"avg_response_time"`
    LastError       string        `json:"last_error,omitempty"`
    LastErrorAt     *time.Time    `json:"last_error_at,omitempty"`
    CurrentTask     string        `json:"current_task"`
    RecentCalls     []APICallLog  `json:"recent_calls"`  // FIFO, max 20
}

type APICallLog struct {
    Endpoint     string        `json:"endpoint"`
    Symbol       string        `json:"symbol"`
    StatusCode   int           `json:"status_code"`
    ResponseTime time.Duration `json:"response_time"`
    Timestamp    time.Time     `json:"timestamp"`
    Error        string        `json:"error,omitempty"`
}
```

### Work Cycle

The worker loops continuously with prioritized tasks:

1. **Price updates** (every cycle): Update `symbols.price` for all tracked symbols via Previous Close. Prioritized: open positions first, then all others.
2. **Dividend sync** (daily): Fetch dividend history and upsert into dividends table. Compare with existing records; only insert new ones.
3. **Price history backfill** (on startup, then daily): For each symbol, check `GetLatestDate()`. Fetch daily bars from last known date to today. On first run, backfill 2 years.
4. **Option history** (daily): For each open option position, build the option ticker from symbol/type/strike/expiration, fetch daily bars for the contract's lifetime.
5. **Index history** (daily): Fetch daily bars for SPX, NDX, DJI, VIX. Backfill from Feb 2023 (data availability start).

Each task yields to the rate limiter (12s between calls). A full cycle through all tasks may take 30-60 minutes depending on symbol count. The worker then sleeps until the next cycle window (e.g., once per hour during market hours, once per day on weekends).

### Rate Limiter

```go
type RateLimiter struct {
    mu       sync.Mutex
    interval time.Duration  // 12 seconds for free tier
    lastCall time.Time
}

func (rl *RateLimiter) Wait() {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    elapsed := time.Since(rl.lastCall)
    if elapsed < rl.interval {
        time.Sleep(rl.interval - elapsed)
    }
    rl.lastCall = time.Now()
}
```

All API calls in `client.go` go through this single rate limiter, ensuring the 5 calls/minute budget is never exceeded regardless of which task is running.

---

## 5. Polygon Client Enhancements — `internal/polygon/client.go`

### New Methods

```go
// GetAggregates fetches historical OHLCV bars for a stock
// timespan: "day", "week", "month"
func (c *Client) GetAggregates(ctx context.Context, ticker, timespan, from, to string, limit int) (*AggregatesResponse, error)

// GetOptionAggregates fetches historical OHLCV bars for an option contract
func (c *Client) GetOptionAggregates(ctx context.Context, optionTicker, timespan, from, to string, limit int) (*AggregatesResponse, error)

// GetIndexAggregates fetches historical OHLC bars for an index
func (c *Client) GetIndexAggregates(ctx context.Context, indexTicker, timespan, from, to string, limit int) (*AggregatesResponse, error)

// GetOptionsContracts lists option contracts for an underlying ticker
func (c *Client) GetOptionsContracts(ctx context.Context, underlyingTicker string, params OptionsContractParams) (*OptionsContractsResponse, error)

// GetStockSplits fetches split history for a ticker
func (c *Client) GetStockSplits(ctx context.Context, ticker string) (*SplitsResponse, error)
```

All three aggregate methods hit the same `/v2/aggs/ticker/{ticker}/range/1/day/{from}/{to}` endpoint — the ticker prefix (`O:` for options, `I:` for indices) determines the asset class.

### New Response Types

```go
type AggregatesResponse struct {
    Status       string `json:"status"`
    Ticker       string `json:"ticker"`
    QueryCount   int    `json:"queryCount"`
    ResultsCount int    `json:"resultsCount"`
    Adjusted     bool   `json:"adjusted"`
    Results      []AggregateBar `json:"results"`
    RequestID    string `json:"request_id"`
}

type AggregateBar struct {
    Open         float64 `json:"o"`
    High         float64 `json:"h"`
    Low          float64 `json:"l"`
    Close        float64 `json:"c"`
    Volume       float64 `json:"v"`
    VWAP         float64 `json:"vw"`
    Timestamp    int64   `json:"t"`
    Transactions int     `json:"n"`
}

type OptionsContractsResponse struct {
    Status    string `json:"status"`
    Results   []OptionsContract `json:"results"`
    RequestID string `json:"request_id"`
    NextURL   string `json:"next_url"`
}

type OptionsContract struct {
    Ticker            string  `json:"ticker"`
    UnderlyingTicker  string  `json:"underlying_ticker"`
    ContractType      string  `json:"contract_type"`
    ExpirationDate    string  `json:"expiration_date"`
    StrikePrice       float64 `json:"strike_price"`
    ExerciseStyle     string  `json:"exercise_style"`
    SharesPerContract float64 `json:"shares_per_contract"`
    PrimaryExchange   string  `json:"primary_exchange"`
}
```

---

## 6. Extract Job Logic from Handlers

### Move FROM `internal/web/polygon_handlers.go`:

The price update loop (lines 72-113) and dividend fetch loop (lines 257-310) contain business logic that belongs in the polygon service/worker, not in HTTP handlers.

### Move TO `internal/polygon/worker.go`:

- `polygonUpdatePricesHandler` → Worker's `priceUpdateCycle()` runs this automatically
- `polygonFetchDividendsHandler` → Worker's `dividendFetchCycle()` runs this automatically

### What remains in `polygon_handlers.go`:

- `polygonTestHandler` — Keep: user-initiated connection test
- `polygonSymbolInfoHandler` — Keep: on-demand symbol lookup
- `polygonStatusHandler` — Rewrite: return `WorkerStatus` from the background worker
- Remove `polygonUpdatePricesHandler` — No longer needed (worker does this)
- Remove `polygonFetchDividendsHandler` — No longer needed (worker does this)
- Add `polygonWorkerStatusHandler` — Returns worker metrics for the Polygon page

---

## 7. Polygon Page Updates — `internal/web/templates/settings.html`

### Remove
- "Update All" button (`#updatePricesBtn`)
- The confirmation dialog and manual fetch JS

### Update "About Polygon.io API" section
Replace generic tier descriptions with what Wheeler actually consumes:

```
Data Consumed by Wheeler:
• Stock Prices — Daily OHLCV for all tracked symbols (34 symbols)
• Price History — Historical daily bars, backfilled to 2 years
• Dividends — Corporate action dividend records
• Option Contracts — Contract details for open positions
• Option History — Historical daily bars for active option contracts
• Index Data — SPX, NDX, DJI, VIX daily history (since Feb 2023)

API Usage: Free tier (5 calls/min). Background worker processes
all data automatically — no manual updates needed.
```

### Add Worker Status Section

Below the API key config and status, add a new card:

```
┌─────────────────────────────────────────────────┐
│  ⚙ Background Worker                           │
│                                                 │
│  Status: ● Running                              │
│  Uptime: 4h 23m                                 │
│  Current Task: Fetching price history for MSFT  │
│                                                 │
│  ┌─────────────────────────────────────────────┐│
│  │ Metrics                                     ││
│  │ Total API Calls:    847                     ││
│  │ Avg Response Time:  234ms                   ││
│  │ Last Error:         None                    ││
│  └─────────────────────────────────────────────┘│
│                                                 │
│  Recent API Calls (last 20):                    │
│  ┌──────────┬──────────┬──────┬───────┬───────┐│
│  │ Time     │ Endpoint │ Sym  │ Code  │ ms    ││
│  ├──────────┼──────────┼──────┼───────┼───────┤│
│  │ 14:23:01 │ prev     │ MSFT │ 200   │ 189   ││
│  │ 14:22:49 │ prev     │ VZ   │ 200   │ 245   ││
│  │ 14:22:37 │ aggs     │ PG   │ 200   │ 312   ││
│  │ 14:22:25 │ divs     │ KO   │ 200   │ 198   ││
│  │ ...      │          │      │       │       ││
│  └──────────┴──────────┴──────┴───────┴───────┘│
└─────────────────────────────────────────────────┘
```

The page polls `/api/polygon/worker-status` every 10 seconds to keep the table live.

---

## 8. Server Integration — `internal/web/server.go` and `main.go`

### `server.go` Changes

- Add `priceHistoryService`, `optionHistoryService`, `indexHistoryService` fields to `Server` struct
- Initialize new services in `NewServer()`
- Pass services to `polygon.NewService()` so the worker can write to the database
- Start worker in `NewServer()` after services are initialized:
  ```go
  ctx, cancel := context.WithCancel(context.Background())
  s.workerCancel = cancel
  s.polygonService.StartWorker(ctx)
  ```
- Stop worker in `Server.Close()`:
  ```go
  s.workerCancel()
  ```

### `main.go` Changes

- The existing shutdown signal handling already calls `server.Close()`, which will now also cancel the worker context. No changes needed to `main.go`.

### Route Changes

- Remove: `/api/polygon/update-prices` route
- Remove: `/api/polygon/fetch-dividends` route
- Add: `/api/polygon/worker-status` route → returns `WorkerStatus` JSON

---

## 9. Task Breakdown

| # | Task | Files | Depends On |
|---|---|---|---|
| 1 | Create migration for history tables | `internal/database/migrations/20260602000001_add_history_tables.sql` | — |
| 2 | Create `PriceBar` model + `PriceHistoryService` | `internal/models/price_history.go` | 1 |
| 3 | Create `OptionBar` model + `OptionHistoryService` | `internal/models/option_history.go` | 1 |
| 4 | Create `IndexBar` model + `IndexHistoryService` | `internal/models/index_history.go` | 1 |
| 5 | Add `AggregatesResponse`, `GetAggregates`, `GetOptionAggregates`, `GetIndexAggregates`, `GetOptionsContracts` to polygon client | `internal/polygon/client.go` | — |
| 6 | Create `RateLimiter` and `WorkerStatus` types | `internal/polygon/worker.go` | — |
| 7 | Implement background worker (`StartWorker`, `priceUpdateCycle`, `dividendFetchCycle`, `priceHistoryBackfill`, `optionHistoryFetch`, `indexHistoryFetch`) | `internal/polygon/worker.go` | 2, 3, 4, 5, 6 |
| 8 | Update `polygon.Service` to accept history services and expose `StartWorker`/`GetWorkerStatus` | `internal/polygon/service.go` | 7 |
| 9 | Update `Server` struct: add history services, start/stop worker | `internal/web/server.go` | 8 |
| 10 | Refactor `polygon_handlers.go`: remove update/fetch handlers, add worker-status handler | `internal/web/polygon_handlers.go` | 9 |
| 11 | Update Polygon page template: remove Update All button, add worker status section, update About section | `internal/web/templates/settings.html` | 10 |
| 12 | Update routes: remove old routes, add new worker-status route | `internal/web/server.go` | 10 |
| 13 | Build and verify: `go build .` | — | All |

---

## 10. Option Ticker Construction

To fetch historical data for Wheeler's option positions, the worker must construct Polygon-format option tickers from the data in the `options` table:

```
Format: O:{SYMBOL}{YYMMDD}{C|P}{STRIKE_PADDED}
Example: O:MSFT260508P00390000  (MSFT $390 Put expiring 2026-05-08)

Strike padding: 8 digits, strike * 1000, zero-padded
  $390.00 → 00390000
  $5.00   → 00005000
  $1327.50 → 01327500
```

The worker queries open options from the database, constructs the ticker, and fetches daily bars for each contract. This only runs for open positions (closed positions don't need price tracking).

---

## 11. Data Retention

- **Price history**: Keep all data (daily bars are small: ~34 symbols × 365 days/yr × ~40 bytes = ~500KB/year)
- **Option history**: Prune bars for expired contracts older than 1 year
- **Index history**: Keep all data (4 indices × 365 days/yr × ~30 bytes = ~44KB/year)

Total estimated database growth: < 5MB/year. No retention concerns.
