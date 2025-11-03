# Monthly Page

**Page Type:** Analytics Page

**Navigation:** Main Menu → Monthly

**Purpose:** Analyze monthly portfolio performance across all income categories (options, capital gains, dividends) with both cumulative and month-by-month views, and detailed breakdowns by symbol.

---

## V2 Changes Summary

| Component | Change Type | Description |
|-----------|-------------|-------------|
| Account Selector | NEW | Add account filtering between totals panel and content |
| Date Range Selector | NEW | Add Last Year / Year to Date / All Time selector |
| Aggregated Totals Panel | ENHANCED | Filter by account and date range |
| Gains Over Time Chart | ENHANCED | Filter by account and date range |
| All "By Ticker" References | MODIFIED | Rename to "By Symbol" for consistency |
| Puts Section Charts | ENHANCED | Filter by account and date range |
| Calls Section Charts | ENHANCED | Filter by account and date range |
| Capital Gains Section Charts | ENHANCED | Filter by account and date range |
| Dividends Section Charts | ENHANCED | Filter by account and date range |
| Monthly Premiums Chart | ENHANCED | Filter by account and date range |
| Total Profit by Symbol Table | ENHANCED | Filter by account and date range, no account column |
| Page Simplification | REVIEW | Consider consolidating/removing redundant charts |

---

## Page Structure

### Aggregated Totals Panel
**Purpose:** Display portfolio totals across all income categories for selected period  
**Component Type:** Metrics Card  
**Position:** Top of page, below page header, full width

**Metrics (5 totals displayed horizontally):**

1. **Total** - Grand total of all categories
2. **Puts** - Put option premiums
3. **Calls** - Call option premiums  
4. **Capital Gains** - Realized gains/losses
5. **Dividends** - Dividend income

**Display Format:**
- Layout: Horizontal row with spacing
- Font: ~20px, green text (#27ae60)
- Labels: Gray (#a0a0a0)
- Background: Dark card (#2d2d2d)
- Negative values: "-$1,234" format

**Current Behavior:**
- Shows year-to-date totals (Jan 1 - current date)
- All values green regardless of positive/negative

🆕 **V2 ENHANCED:**
- Filter by selected account (from Account Selector)
- Filter by selected date range (from Date Range Selector)
- **Default**: "All Accounts" + "Year to Date" = current behavior
- Update totals when either selector changes

**Data Requirements:**
- Query: Filter transactions by account_id and date_range
- Group by: Transaction type (PUT, CALL, STOCK_SALE, DIVIDEND)
- Sum amounts for each category

---

### 🆕 V2 NEW COMPONENT: Account Selector
**Purpose:** Filter all page data by account  
**Component Type:** Dropdown Selector Card  
**Position:** Between Aggregated Totals Panel and Gains Over Time Section, full width

**Behavior:**
- **Default:** "-- All Accounts --"
- **Options:** List all active user accounts
- **Impact:** Filters totals, charts, and table
- **Persistence:** Selected account persists across navigation

**Backend:**
- All queries accept optional `account_id` parameter
- `NULL` account_id = aggregate across all accounts
- Specific account_id: `WHERE account_id = ?`

---

### 🆕 V2 NEW COMPONENT: Date Range Selector
**Purpose:** Filter page data by time period  
**Component Type:** Dropdown or Button Group  
**Position:** Gains Over Time section header (center, between title and toggle)

**Options:**
- **Last Year:** Previous 12 calendar months from today
- **Year to Date:** January 1 of current year through today
- **All Time:** All historical data in database

**Default:** Year to Date

**Behavior:**
- Filters Totals Panel, all charts, and table
- Updates X-axis months dynamically for selected range
- Selection persists within session

**Backend:**
- Queries accept `date_range` parameter: 'last_year', 'ytd', 'all_time'
- Server calculates date boundaries based on current date
- Returns filtered data for charts and totals

**Frontend:**
- Dropdown or button group component
- Updates all data on selection change
- Visual indication of active selection

---

## Content Sections

### Gains Over Time Section
**Purpose:** Visualize total portfolio gains with cumulative/monthly toggle  
**Component Type:** Card with controls and stacked bar chart  
**Position:** Below Account Selector, full width

**Section Header Controls (3 components):**
- **Title:** "Gains Over Time" (left)
- **🆕 V2 NEW: Date Range Selector** (center)
- **Cumulative / Monthly Toggle** (right - EXISTING)

---

#### Cumulative/Monthly Toggle (EXISTING FEATURE)
**Purpose:** Switch between cumulative and monthly view  
**Component Type:** Button Group  
**Position:** Section header, right side

**Buttons:**
- **Cumulative** (default active): Green bg, white text when active
- **Monthly**: Gray bg, light gray text when inactive

**Behavior:**
- Toggle switches chart data between cumulative and monthly
- Visual state change on button
- Does NOT affect date range or account filter
- Persists within page session

**Styling:**
- Grouped buttons (no gap)
- Rounded corners on outer edges only
- Active: #27ae60 bg, white text
- Inactive: #34495e bg, #bdc3c7 text

---

#### Gains Over Time Chart
**Purpose:** Stacked bar chart showing portfolio gains breakdown  
**Component Type:** Stacked Bar Chart  
**Position:** Below section header, full width, large height

**Chart Configuration:**
- **Type:** Stacked bars (4 series)
- **X-Axis:** Months (dynamic based on date range)
- **Y-Axis:** Dollar amount
- **Legend:** Top, horizontal

**4 Data Series:**
1. **Puts** - Blue (rgba(31, 119, 180, 0.8))
2. **Calls** - Orange (rgba(255, 127, 14, 0.8))
3. **Capital Gains** - Green/Red dynamic (positive/negative)
4. **Dividends** - Gold (rgba(255, 215, 0, 0.8))

**View Modes:**
- **Cumulative:** Y-Axis = "Cumulative Gains ($)", shows running totals
- **Monthly:** Y-Axis = "Monthly Gains ($)", shows individual month values

🆕 **V2 ENHANCED:**
- Filter by selected account
- Filter by selected date range
- X-axis adjusts to show months in selected range
- Maintain cumulative/monthly toggle with filtered data

**Interactions:**
- Hover: Tooltip with category breakdown + total
- Toggle: Switches between cumulative/monthly views

---

### Options Analysis Sections (Side-by-Side)
**Purpose:** Detailed analysis of puts and calls separately  
**Component Type:** Two cards side-by-side  
**Position:** Below Gains Over Time section

**Layout:** 50/50 split, two sections:
- **Left:** Puts section
- **Right:** Calls section

Each section contains:
- Section title ("Puts" or "Calls")
- Two charts side-by-side:
  - **By Month** (left): Bar + cumulative line chart
  - **By Symbol** (right): Pie chart 🆕 **V2 MODIFIED:** (was "By Ticker")

---

#### Puts Section

**Section Title:** "Puts"

**Charts (side-by-side):**

1. **By Month Chart**
   - Type: Bar chart with cumulative line overlay
   - Bars: Monthly put premiums (blue)
   - Line: Cumulative total (gold)
   - Dual Y-axes: Left = monthly, Right = cumulative

2. **By Symbol Chart** 🆕 **V2 MODIFIED:** (was "By Ticker")
   - Type: Pie chart
   - Data: Put premiums by symbol
   - Legend: Right side
   - Colors: Consistent color per symbol
   - Data labels: Dollar amounts on slices

🆕 **V2 ENHANCED:**
- Both charts filter by account
- Both charts filter by date range
- Rename "By Ticker" to "By Symbol"

---

#### Calls Section

**Section Title:** "Calls"

**Charts (side-by-side):**

1. **By Month Chart**
   - Type: Bar chart with cumulative line overlay
   - Bars: Monthly call premiums (orange/blue)
   - Line: Cumulative total (gold)
   - Dual Y-axes: Left = monthly, Right = cumulative

2. **By Symbol Chart** 🆕 **V2 MODIFIED:** (was "By Ticker")
   - Type: Pie chart
   - Data: Call premiums by symbol
   - Legend: Right side
   - Colors: Consistent color per symbol
   - Data labels: Dollar amounts on slices

🆕 **V2 ENHANCED:**
- Both charts filter by account
- Both charts filter by date range
- Rename "By Ticker" to "By Symbol"

---

### Capital Gains & Dividends Sections (Side-by-Side)
**Purpose:** Detailed analysis of capital gains and dividends  
**Component Type:** Two cards side-by-side  
**Position:** Below Options sections

**Layout:** 50/50 split, two sections:
- **Left:** Capital Gains section
- **Right:** Dividends section

Each section identical structure to Puts/Calls sections above.

---

#### Capital Gains Section

**Section Title:** "Capital Gains"

**Charts (side-by-side):**

1. **By Month Chart**
   - Type: Bar + cumulative line
   - Bars: Monthly capital gains (green/red for profit/loss)
   - Line: Cumulative (gold)

2. **By Symbol Chart** 🆕 **V2 MODIFIED:** (was "By Ticker")
   - Type: Pie chart
   - Data: Capital gains by symbol

🆕 **V2 ENHANCED:**
- Filter by account and date range
- Rename "By Ticker" to "By Symbol"

---

#### Dividends Section

**Section Title:** "Dividends"

**Charts (side-by-side):**

1. **By Month Chart**
   - Type: Bar + cumulative line
   - Bars: Monthly dividends (blue)
   - Line: Cumulative (gold)

2. **By Symbol Chart** 🆕 **V2 MODIFIED:** (was "By Ticker")
   - Type: Pie chart
   - Data: Dividends by symbol

🆕 **V2 ENHANCED:**
- Filter by account and date range
- Rename "By Ticker" to "By Symbol"

---

### Monthly Premiums by Symbol (Open/Closed) Section
**Purpose:** Show option premiums grouped by open vs closed positions  
**Component Type:** Card with stacked grouped bar chart  
**Position:** Below Capital Gains & Dividends sections, full width

**Section Title:** "Monthly Premiums by Symbol (Open/Closed)"

**Chart Configuration:**
- **Type:** Stacked grouped bar chart
- **X-Axis:** Months
- **Bars:** Two stacks per month (Closed positions, Open positions)
- **Colors:** Consistent per symbol, with transparency for open positions
- **Legend:** Right side, compact

**Stacking Logic:**
- **Stack 1 (Closed):** All symbols' closed position premiums stacked
- **Stack 2 (Open):** All symbols' open position premiums stacked (with transparency)
- Each symbol has consistent color across both stacks

**Data Requirements:**
- Options data grouped by month, symbol, and open/closed status
- Calculate: (premium - exit_price - commission) per option
- Month determined by opened date

🆕 **V2 ENHANCED:**
- Filter by account
- Filter by date range
- Dynamic X-axis for selected date range

**Interactions:**
- Hover: Detailed tooltip with symbol breakdown and totals

**Note:** This chart may be complex and could be considered for simplification in V2 review.

---

### Total Profit by Symbol Table
**Purpose:** Tabular view of profit by symbol across all months  
**Component Type:** Data Table (scrollable)  
**Position:** Below Monthly Premiums chart, full width

**Section Title:** "Total Profit by Symbol"

**Table Structure:**

**Columns:**
- **Ticker** (left, bold, clickable link to Symbol Detail page)
- **Total** (sum across all months)
- **January** through **December** (12 individual month columns)

**Rows:**
- One row per symbol with any profit data
- Click symbol: Navigate to Symbol Detail page
- Values: Currency format ($1,234.56)

**Footer Row:**
- **Label:** "Total"
- **Values:** Sum of all symbols per column (Total + each month)

**Data Requirements:**
- Query: All transactions grouped by symbol and month
- Aggregation: Sum profit by symbol by month
- Sort: Alphabetically by symbol

🆕 **V2 ENHANCED:**
- Filter by account (no account column needed when filtering)
- Filter by date range (show only months in range)
- Dynamically adjust month columns based on date range
- When "All Accounts" selected: Aggregate across accounts, no account column

**Interactions:**
- Click symbol: Navigate to Symbol Detail page
- Scrollable: Horizontal scroll for 12+ columns

**Empty State:**
- Message: "No profit data available"

---

## Current State Notes

### Working Features:
- Aggregated totals panel displays correctly
- Gains Over Time chart with cumulative/monthly toggle (already implemented)
- 4 sections with dual charts (Puts, Calls, Cap Gains, Dividends)
- Each section has: By Month (bar + line) and By Ticker (pie)
- Monthly Premiums grouped chart by open/closed status
- Total Profit by Symbol table with monthly breakdown
- Symbol links to detail pages
- Consistent color coding across charts

### UI/UX Observations:
- Page has many charts (potentially too many)
- Clear visual hierarchy with sections
- Dark theme consistent with rest of app
- Charts use Chart.js library
- Toggle button implementation is clean
- Pie charts labeled with dollar amounts

### Data Observations:
- All charts pull from same transaction data
- Consistent date handling (using opened date for options)
- Color coding: Blue/orange for options, green/red for gains, gold for cumulative
- Symbol color consistency across related charts

### Business Logic Location:
- Chart rendering: Frontend (Chart.js)
- Data aggregation: Backend (Go templates)
- Cumulative calculations: Frontend JavaScript
- Toggle state management: Frontend JavaScript

---

## Technical Notes

### Backend Requirements:

**API Endpoints (NEW/UPDATED for V2):**
- `GET /api/monthly/summary` - Totals panel data (filtered by account_id, date_range)
- `GET /api/monthly/gains-over-time` - Chart data (filtered by account_id, date_range)
- `GET /api/monthly/puts` - Puts chart data (filtered by account_id, date_range)
- `GET /api/monthly/calls` - Calls chart data (filtered by account_id, date_range)
- `GET /api/monthly/cap-gains` - Capital gains chart data (filtered by account_id, date_range)
- `GET /api/monthly/dividends` - Dividends chart data (filtered by account_id, date_range)
- `GET /api/monthly/premiums-grouped` - Monthly premiums chart (filtered by account_id, date_range)
- `GET /api/monthly/profit-table` - Table data (filtered by account_id, date_range)

**Query Parameters:**
- `account_id` (optional): Filter by account, NULL = all accounts
- `date_range` (required): 'last_year', 'ytd', 'all_time'

**Date Range Calculations:**
- **Last Year:** Current date - 12 months
- **YTD:** January 1 of current year through current date
- **All Time:** No date filter

**Performance Considerations:**
- Aggregate queries may be expensive with large datasets
- Consider caching for "All Time" view
- Index on account_id, transaction_date

### Frontend Requirements:

**State Management:**
- Selected account (from Account Selector)
- Selected date range (Last Year / YTD / All Time)
- Cumulative/Monthly toggle state
- Chart instances (for updates/destruction)

**Chart Libraries:**
- Chart.js for all charts
- ChartDataLabels plugin for value labels
- Consistent colors across symbol references

**Data Refresh:**
- Refresh all charts on account change
- Refresh all charts on date range change
- Maintain toggle state during filter changes
- Update totals panel on any filter change

**Dynamic X-Axis:**
- "Last Year": Show 12 months rolling
- "YTD": Show Jan through current month only
- "All Time": Show all months with data (or Jan-Dec for current year)

---

## V2 Page Simplification - Review Needed

**Potential Issues:**
1. **Too Many Charts:** Page has 11 charts total (1 main + 8 dual + 1 grouped + 1 table)
2. **Redundancy:** By Month charts all show similar cumulative patterns
3. **Cognitive Load:** Difficult to scan all charts at once

**Simplification Options to Consider:**

**Option 1: Consolidate By Symbol Charts**
- Remove individual By Symbol pie charts from each section
- Keep only Gains Over Time stacked bar chart
- Add consolidated pie chart showing all symbols across all categories

**Option 2: Remove Cumulative Lines**
- Remove cumulative line overlays from By Month charts
- Keep only monthly bars
- Gains Over Time chart already shows cumulative view

**Option 3: Tabs/Accordion**
- Keep Gains Over Time chart prominent
- Put detailed breakdowns (Puts, Calls, etc.) in tabs or accordions
- User selects which category to deep-dive

**Option 4: Remove Monthly Premiums Chart**
- Chart shows open/closed distinction
- Information may be redundant with Options Overview page
- Consider removing or moving to different page

**Recommendation:** Discuss with product owner/users which charts provide most value.

---

## V2 Implementation Priority

**Phase 1: Core Filtering**
1. Add Account Selector component
2. Add Date Range Selector component
3. Update backend queries to accept account_id and date_range parameters
4. Test "All Accounts" + "YTD" = current behavior
5. Test specific account filtering

**Phase 2: Chart Updates**
1. Update Gains Over Time chart to filter by account and date range
2. Update all By Month charts to filter
3. Update all By Symbol charts to filter
4. Update Monthly Premiums chart to filter
5. Update table to filter
6. Test dynamic X-axis for date ranges

**Phase 3: Nomenclature**
1. Find/replace "By Ticker" with "By Symbol" in all templates
2. Update chart titles
3. Update legends
4. Verify consistency

**Phase 4: Simplification Review**
1. Gather user feedback on chart usage
2. Identify least-used charts
3. Prototype simplified layouts
4. A/B test if possible
5. Implement approved simplifications

---

## Testing Checklist

### Current State Verification:
- [ ] Totals panel displays correctly
- [ ] Gains Over Time chart renders
- [ ] Cumulative/Monthly toggle works
- [ ] All 4 sections render (Puts, Calls, Cap Gains, Dividends)
- [ ] Monthly Premiums chart renders
- [ ] Profit table renders with all months
- [ ] Symbol links navigate correctly

### V2 Account Filtering:
- [ ] Account Selector displays all accounts
- [ ] "All Accounts" shows aggregated data
- [ ] Specific account filters totals panel
- [ ] Specific account filters all charts
- [ ] Specific account filters table
- [ ] Selected account persists across navigation

### V2 Date Range Filtering:
- [ ] Date Range Selector displays three options
- [ ] "Last Year" shows correct 12-month period
- [ ] "YTD" shows Jan 1 through today
- [ ] "All Time" shows all data
- [ ] Totals panel updates for selected range
- [ ] All charts update for selected range
- [ ] Table updates for selected range
- [ ] X-axis adjusts dynamically for range

### V2 Combined Filtering:
- [ ] Account + Date Range work together correctly
- [ ] "All Accounts" + "YTD" = original behavior
- [ ] Specific account + "Last Year" filters correctly
- [ ] Toggle works with filtered data

### V2 Nomenclature:
- [ ] All "By Ticker" renamed to "By Symbol"
- [ ] Chart titles updated
- [ ] Legends updated
- [ ] Consistency across page

---

## Questions for Review

1. **Date Range Scope:** Should Date Range Selector affect Totals Panel, or should panel always show YTD?
2. **X-Axis Display:** For "All Time", show all historical months or just current year Jan-Dec?
3. **Simplification Priority:** Which charts are most valuable? Which could be removed/consolidated?
4. **Monthly Premiums Chart:** Keep, simplify, or remove this complex chart?
5. **Table Columns:** For date ranges less than full year, hide empty month columns or show all 12?
6. **Default Range:** Keep "Year to Date" as default, or use "All Time"?

---

**Document Status:** Complete - Ready for Review  
**Artifact ID:** `monthly_page_spec`  
**Last Updated:** 2025-11-03

**Notes:**
- Page currently has 11 visualizations (many charts)
- Cumulative/Monthly toggle already working (preserve in V2)
- Strong candidate for simplification/consolidation
- Consider user research on chart usage before simplifying