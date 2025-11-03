# Symbol Detail Page

**Page Type:** Detail Page

**Navigation:** Main Menu → Symbols → [Symbol Name] (e.g., ANET)

**Purpose:** Display comprehensive information and trading history for a single stock symbol, including options trades, stock positions, dividends, and monthly performance breakdown.

---

## V2 Changes Summary

| Component | Change Type | Description |
|-----------|-------------|-------------|
| Account Selector | **NEW** | Add account filtering dropdown between Entity Summary Panel and Content Sections |
| Data Filtering | **ENHANCED** | All content sections filter by selected account |
| Entity Summary Panel | **ENHANCED** | Metrics calculate based on selected account filter |

---

## Page Structure

### Entity Summary Panel
**Entity Title:** `[SYMBOL] - [Company Name]` (e.g., "ANET - ANET Inc")

**Metric Cards:** 12 key metrics displayed in horizontal grid
- Price: Current stock price
- Div: Annual dividend amount
- Yield: Dividend yield percentage
- P/E: Price-to-earnings ratio
- Options Gains: Total profit from options trades
- Cap Gains: Total capital gains from stock sales
- Dividends: Total dividend income received
- Total Profits: Sum of all gains (options + capital + dividends)
- Cash on Cash: Return percentage on cash deployed
- Long Value: Current value of held stock positions
- Put Exposed: Total capital at risk from open put positions

**Entity Actions:**
- "Edit" button (blue) - Opens Symbol Edit Modal
- "Delete" button (red) - Opens Symbol Delete Confirmation Modal

**Data Requirements:**
- Symbol entity (symbol, name, price, dividend, pe_ratio)
- Aggregated options P&L for this symbol
- Aggregated capital gains for this symbol
- Aggregated dividend income for this symbol
- Current long position value calculation
- Current put exposure calculation

**Business Rules:**
- Total Profits = Options Gains + Cap Gains + Dividends
- Cash on Cash = Total Profits / Total Capital Deployed
- Long Value = SUM(shares × current_price) for open positions
- Put Exposed = SUM(strike × contracts × 100) for open puts
- **🆕 V2:** All metrics filter by selected account when account filter active

---

### Account Selector
**🆕 V2 NEW COMPONENT**

**Purpose:** Allow users to filter all page data by specific account or view aggregated data across all accounts.

**Component Type:** Dropdown selector

**Position:** Between Entity Summary Panel and Content Sections (full-width card)

**Display Options:**
- Default: "-- All Accounts --"
- Account List: All user accounts (e.g., "Primary Account", "IRA Account", "Margin Account")

**Optional Enhancement:** Account Summary Card
- Account Name
- Account Type (Cash, Margin, IRA)
- Account Total Value
- Account Status (Active/Archived)

**Behavior:**
- **"-- All Accounts --" selected:** 
  - Shows aggregated data across all accounts
  - Entity Summary Panel shows combined metrics
  - Content sections show all data (no account column added to tables)
  
- **Specific Account selected:**
  - Entity Summary Panel recalculates metrics for selected account only
  - All content sections filter: WHERE account_id = [selected_account]
  - No account column needed in tables (implicit from selector)

**Data Requirements:**
- List of all accounts for dropdown: `SELECT id, name, account_type FROM accounts ORDER BY name`
- Selected account persists in session/state during page navigation

**Business Rules:**
- Account filter applies to ALL sections simultaneously
- Account selection persists when navigating to other pages
- Default to "-- All Accounts --" on initial page load
- Cannot filter by archived accounts (only show active accounts in dropdown)

**Interactions:**
- User selects account from dropdown → Page reloads/refreshes with filtered data
- Account selection shows in URL or state (for deep linking/bookmarking)

**Frontend Validation:**
- Dropdown always has valid selection (cannot be empty)

**Backend Requirements:**
- Account list API endpoint
- All data queries accept optional account_id filter parameter
- Aggregate queries handle NULL account_id (meaning all accounts)

---

### Content Sections

#### Section 1: Options
**Purpose:** Display all option trades (puts and calls) for this symbol with complete trade details and performance metrics.

**Section Actions:** 
- "+ Add" button - Opens Add Option Modal

**Components:**
- Interactive Sortable Table

**Table Columns:** (18 columns)
1. Call/Put - Type Badge ("P" or "C")
2. Date Sold - Trade opening date
3. Closed Date - Trade closing date (empty if open)
4. Strike - Option strike price
5. OTM - Out of the money amount
6. Expiration - Option expiration date
7. Remaining - Days remaining (for open positions)
8. DTE - Days to expiration
9. DTC - Days to close (for closed positions)
10. Contracts - Number of contracts
11. Premium - Premium received per contract
12. Exit Price - Price paid to close (if closed)
13. Commission - Total commission paid
14. Total - Net profit/loss for trade
15. % of Profit - Percentage return
16. % of Time - Percentage of time to expiration used
17. Multiplier - Risk-adjusted multiplier
18. Actions - Edit and Delete icons

**Row Actions:**
- Edit icon (pencil) - Opens Edit Option Modal (pre-populated)
- Delete icon (trash) - Opens Delete Confirmation Modal

**Data Requirements:**
- All options records WHERE symbol = [current_symbol]
- **🆕 V2:** Additional filter: AND account_id = [selected_account] (if account selected)
- Calculated fields: OTM, Remaining, DTE, DTC, Total P&L, % of Profit, % of Time

**Business Rules:**
- Total = (Premium - Exit Price) × Contracts × 100 - Commission
- % of Profit = (Total / (Strike × Contracts × 100)) × 100
- DTE = Days between current date and expiration
- DTC = Days between opened and closed dates
- % of Time = (DTC / (DTE at open)) × 100
- OTM calculation varies by Put/Call and current price vs strike

**Interactions:**
- Click column header → Sort table by that column (ascending/descending toggle)
- Click Edit icon → Opens Edit Option Modal with form pre-populated
- Click Delete icon → Opens Confirmation Modal, then deletes record
- Click "+ Add" → Opens Add Option Modal for new trade entry

**Empty State:** Table shows existing data (no empty state shown in current example)

**Validation:** N/A (display only, validation occurs in modals)

---

#### Section 2: Stock Positions
**Purpose:** Display all long stock positions (current and historical) for this symbol.

**Section Actions:**
- "+ Add" button - Opens Add Stock Position Modal

**Components:**
- Interactive Sortable Table

**Table Columns:** (10 columns)
1. Purchased - Date shares acquired
2. Closed Date - Date shares sold (empty if holding)
3. Yield - Dividend yield at purchase
4. Shares - Number of shares
5. Buy Price - Purchase price per share
6. Exit Price - Sale price per share (if sold)
7. Profit/Loss - Capital gain/loss
8. ROI - Return on investment percentage
9. Amount - Total purchase amount
10. Total Invested - Buy Price × Shares + Commission
11. Actions - Edit and Delete icons

**Row Actions:**
- Edit icon (pencil) - Opens Edit Stock Position Modal
- Delete icon (trash) - Opens Delete Confirmation Modal

**Data Requirements:**
- All long_positions records WHERE symbol = [current_symbol]
- **🆕 V2:** Additional filter: AND account_id = [selected_account] (if account selected)
- Calculated fields: Profit/Loss, ROI, Amount, Total Invested

**Business Rules:**
- Profit/Loss = (Exit Price - Buy Price) × Shares
- ROI = (Profit/Loss / Total Invested) × 100
- Amount = Current Price × Shares (for open positions)
- Total Invested = Buy Price × Shares + Commission

**Interactions:**
- Click column header → Sort table by that column
- Click Edit icon → Opens Edit Stock Position Modal
- Click Delete icon → Opens Confirmation Modal, then deletes record
- Click "+ Add" → Opens Add Stock Position Modal

**Empty State:** "No stock positions recorded for [SYMBOL]"

**Validation:** N/A (display only, validation occurs in modals)

---

#### Section 3: Dividends
**Purpose:** Display all dividend payments received for this symbol.

**Section Actions:**
- "+ Add" button - Opens Add Dividend Modal

**Components:**
- Interactive Sortable Table

**Table Columns:** (3 columns)
1. Received - Payment date
2. Amount - Dividend amount received
3. Actions - Delete icon

**Row Actions:**
- Delete icon (trash) - Opens Delete Confirmation Modal

**Data Requirements:**
- All dividends records WHERE symbol = [current_symbol]
- **🆕 V2:** Additional filter: AND account_id = [selected_account] (if account selected)

**Business Rules:**
- Simple display of payment date and amount
- No calculations required

**Interactions:**
- Click column header → Sort table by that column
- Click Delete icon → Opens Confirmation Modal, then deletes record
- Click "+ Add" → Opens Add Dividend Modal

**Empty State:** "No dividends recorded for [SYMBOL]"

**Validation:** N/A (display only, validation occurs in modals)

---

#### Section 4: Monthly Results
**Purpose:** Show monthly breakdown of options performance (puts vs calls) for this symbol.

**Section Actions:** None (read-only analytics)

**Components:**
- Data Table (read-only, not sortable)

**Table Columns:** (6 columns)
1. Month - Calendar month name
2. Puts - Number of put trades closed that month
3. Calls - Number of call trades closed that month
4. Puts Total - Total profit from puts that month
5. Calls Total - Total profit from calls that month
6. Total - Combined profit for the month

**Data Requirements:**
- Aggregated options data grouped by close month
- **🆕 V2:** Additional filter: WHERE account_id = [selected_account] (if account selected)
- COUNT and SUM calculations per month per option type

**Business Rules:**
- Group options by MONTH(closed_date)
- Puts Total = SUM(profit) WHERE type = 'Put' AND closed in month
- Calls Total = SUM(profit) WHERE type = 'Call' AND closed in month
- Total = Puts Total + Calls Total
- Only show months with activity

**Interactions:**
- View only - no user actions available

**Empty State:** Table empty if no closed options exist (no explicit empty state message)

**Validation:** N/A (read-only display)

---

## Current State Notes

**Working Features:**
- All sections display data correctly based on screenshot
- Entity Summary Panel shows comprehensive metrics
- Tables are functional with sorting capabilities
- Row actions (edit/delete) present and accessible
- Add buttons visible in appropriate section headers
- Empty states display helpful messages

**UI/UX Observations:**
- Dense information display - many columns in Options table
- Consistent dark theme throughout
- Color coding: Green for profits, Red for delete actions, Blue for primary actions
- Type badges use single letter ("P"/"C") for compactness
- Sortable columns indicated by ↕ icon in headers

**Data Integrity:**
- All entity relationships maintained (symbol → options, positions, dividends)
- Metric calculations appear in real-time in Entity Summary Panel
- Empty states prevent confusion when no data exists

**Business Logic Location:**
- Frontend: Table sorting, UI interactions, empty state display
- Backend: All calculations (P&L, AROI, aggregations), data fetching
- Mixed: Form validation (frontend + backend)

---

## Technical Notes

**Backend Requirements:**
- Symbol entity fetch
- **🆕 V2:** Accounts list fetch for dropdown
- Options list query with calculated fields **🆕 V2:** + account filter
- Long positions list query with calculated fields **🆕 V2:** + account filter
- Dividends list query **🆕 V2:** + account filter
- Monthly aggregation query **🆕 V2:** + account filter
- Summary metrics calculation (6 aggregate queries) **🆕 V2:** + account filter

**Frontend Capabilities:**
- Client-side table sorting
- Modal management (open/close)
- Form submission
- Delete confirmation flow
- Empty state rendering
- **🆕 V2:** Account selector state management
- **🆕 V2:** Page refresh on account selection change

**Performance Considerations:**
- Multiple database queries for single page load
- Calculation-heavy Entity Summary Panel
- Large options table could impact performance with many trades
- Consider pagination if table grows beyond 50-100 rows
- **🆕 V2:** Account filtering reduces data volume per query (performance benefit)
- **🆕 V2:** Account selector change triggers full page data refresh (consider caching strategy)

---

## V2 Change Notation

**Format Guide:**
- **🆕 V2 NEW COMPONENT** - Entirely new section or major component
- **🆕 V2:** Inline note - Enhancement or modification to existing feature
- All V2 changes summarized in table at top of document

**Change Types:**
- **NEW** - Brand new component/feature
- **ENHANCED** - Improvement to existing feature
- **MODIFIED** - Change to existing behavior
- **DEPRECATED** - Feature being removed or replaced