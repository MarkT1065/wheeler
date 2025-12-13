# All Options Page

**Page Type:** Analytics Page (with filtering)

**Navigation:** Main Menu → Options → All Options 

**Purpose:** Provide comprehensive overview of all options trading activity with advanced filtering, visual analytics, and detailed trade listing.

---

## V2 Changes Summary

| Component | Change Type | Description |
|-----------|-------------|-------------|
| Filter Panel | **ENHANCED** | Add Account filter between Symbol and Type |
| Summary Metrics | **ENHANCED** | Metrics calculate based on selected account filter |
| Monthly Performance Chart | **ENHANCED ** | Display data consistent with filter or current year (12 month view) if no filter |
| Data Tab | **DEPRECATED** | Remove raw JSON data view (limited utility) |

---

## Page Structure

### Summary Metrics Panel
**Purpose:** Display key aggregate statistics for all options trading activity (filtered or unfiltered).

**Component Type:** Horizontal metrics bar

**Metrics Displayed:** (6 key metrics)
- Total Profit: Sum of all options P&L
- Contracts: Total number of option contracts traded
- Puts: Count of put option trades
- Calls: Count of call option trades
- Avg Strike: Average strike price across all trades
- Avg Premium: Average premium received/paid per contract

**Data Requirements:**
- Aggregated options data: COUNT, SUM, AVG calculations
- **🆕 V2:** Filter by account_id if account selected

**Business Rules:**
- Total Profit = SUM(net profit) for all options
- Contracts = SUM(contracts) for all options
- Puts = COUNT(*) WHERE type = 'Put'
- Calls = COUNT(*) WHERE type = 'Call'
- Avg Strike = AVG(strike) across all trades
- Avg Premium = AVG(premium) across all trades
- All metrics update dynamically based on active filters
- **🆕 V2:** All metrics filter by selected account when account filter active

**Styling:**
- Large, prominent display at top of page
- Total Profit emphasized (larger font, green color)
- Other metrics in smaller, secondary style

---

### Filter Panel
**Purpose:** Allow users to narrow down options data by multiple criteria for focused analysis.

**Component Type:** Multi-field filter form

**Filter Controls:** (9 filters + clear action)

1. **Symbol** - Dropdown
   - Options: "All Symbols" + list of tracked symbols
   - Default: "All Symbols"

2. **🆕 V2: Account** - Dropdown  
   - Options: "-- All Accounts --" + list of user accounts
   - Default: "-- All Accounts --"
   - Position: Between Symbol and Type

3. **Type** - Dropdown
   - Options: "All Types", "Put", "Call"
   - Default: "All Types"

4. **Status** - Dropdown
   - Options: "All", "Open", "Closed", "Assigned", "Expired"
   - Default: "All"

5. **Expiration From** - Date Picker
   - Format: mm/dd/yyyy
   - Default: Empty (no filter)

6. **Expiration To** - Date Picker
   - Format: mm/dd/yyyy
   - Default: Empty (no filter)

7. **Opened From** - Date Picker
   - Format: mm/dd/yyyy
   - Default: Empty (no filter)

8. **Opened To** - Date Picker
   - Format: mm/dd/yyyy
   - Default: Empty (no filter)

9. **Closed From** - Date Picker
   - Format: mm/dd/yyyy
   - Default: Empty (no filter)

10. **Closed To** - Date Picker
    - Format: mm/dd/yyyy
    - Default: Empty (no filter)

**Actions:**
- "Clear All" button - Resets all filters to defaults

**Data Requirements:**
- Symbol list for dropdown
- **🆕 V2:** Account list for dropdown
- Filter values persist in session/state

**Business Rules:**
- Multiple filters combine with AND logic
- Date range filters: FROM is inclusive start, TO is inclusive end
- Empty date fields mean no date constraint
- Status "All" shows both open and closed positions
- Filters apply to both chart and table simultaneously
- **🆕 V2:** Account filter applies across entire page (metrics, chart, table)

**Interactions:**
- Change any filter → Page data refreshes immediately
- Click "Clear All" → All filters reset to defaults, page refreshes

**Frontend Validation:**
- Date From cannot be after Date To (show error if invalid)
- All date inputs must be valid dates or empty

**Backend Requirements:**
- Options query accepts all filter parameters
- Query builds WHERE clause dynamically based on active filters
- **🆕 V2:** Account filter parameter added to query

---

### Content Sections

#### Section 1: Monthly Performance Chart
**Purpose:** Visualize options trading performance by month showing maximum potential profit, actual realized profit, and open position value.

**Component Type:** Stacked/Grouped Bar Chart

**Chart Configuration:**
- **🆕 V2:** X-axis: Display data consistent with filter or current year (12 month view) if no filter
- Y-axis: Dollar amounts ($0 to max value)
- Three data series (bars):
  - **Max Profit** (Yellow) - Maximum possible profit if all positions closed favorably
  - **Actual Profit** (Green) - Realized profit from closed positions
  - **Open Value** (Blue) - Current value of open positions
- Legend at top of chart

**Data Requirements:**
- Aggregated options data grouped by month
- Calculations per month:
  - Max Profit = SUM(premium × contracts × 100) for that month
  - Actual Profit = SUM(actual P&L) for closed positions that month
  - Open Value = SUM(current value) for positions still open
- **🆕 V2:** Filter by account_id if account selected
- Applies all active filter criteria from Filter Panel

**Business Rules:**
- Group by month of trade open date
- Max Profit includes both open and closed positions
- Actual Profit only includes closed positions
- Open Value only shows for positions not yet closed
- Chart updates dynamically when filters change
- Months with no activity show $0 bars

**Interactions:**
- Hover over bar → Tooltip shows exact values
- Click legend item → Toggle series visibility
- Chart is read-only (no drill-down functionality shown)

**Empty State:** Chart shows empty grid if no data matches filters

---

#### Section 2: Options Data Table
**Purpose:** Display detailed line-item data for all options trades matching current filter criteria.

**Component Type:** Tabs with Interactive Sortable Table

**Tab Navigation:**
- **"Options" tab** (active by default) - Shows options trade table
- **"Data" tab** - Shows raw JSON data **🆕 V2 DEPRECATED:** Remove this tab

**Table Header:**
- Shows count: "Showing all [N] options"

**Table Columns:** (10 columns)
1. Symbol - Stock ticker (clickable link to Symbol Detail Page)
2. Type - Badge indicator (C for Call, P for Put)
3. Strike - Strike price
4. Expiration - Expiration date
5. Opened - Trade open date
6. Closed - Trade close date (empty if open)
7. Contracts - Number of contracts
8. Premium - Premium per contract
9. Max Profit - Maximum potential profit
10. Actual - Actual realized profit (for closed trades)

**Row Styling:**
- Symbol as clickable link (blue)
- Type badge: Green for Call, Blue for Put
- Actual values color-coded: Green for profit, Red for loss
- Sortable columns indicated by ↕ icon

**Data Requirements:**
- All options matching filter criteria
- Calculated fields: Max Profit, Actual P&L
- **🆕 V2:** Filter by account_id if account selected
- Sorted by most recent first (default)

**Business Rules:**
- Max Profit = premium × contracts × 100 - commission
- Actual = (premium - exit_price) × contracts × 100 - commission (for closed)
- Actual shows "-" for open positions
- Apply all active filters from Filter Panel
- **🆕 V2:** When account filter active, only show trades for that account

**Interactions:**
- Click column header → Sort table by that column (ascending/descending toggle)
- Click Symbol link → Navigate to Symbol Detail Page for that symbol
- Table rows are read-only (no inline edit/delete actions)
- Scroll vertically if many rows

**Empty State:** "No options match the selected filters" (if filters yield no results)

**Pagination:** Not shown in current implementation (loads all matching records)

---

#### Section 2b: Data Tab (Raw JSON)
**🆕 V2 DEPRECATED**

**Current State:**
Shows raw JSON structure of options index data with label: "Options Index Data - Raw JSON data structure used to build the options table"

**Deprecation Rationale:**
- Limited utility for end users
- Debugging/technical view not needed in production UI
- Data already visible in formatted Options table
- Increases page complexity without user benefit

**Migration Path:**
- Remove "Data" tab from tab navigation
- Options tab becomes default and only view
- No user-facing impact (rarely used feature)

---

## Current State Notes

**Working Features:**
- Comprehensive filtering with 9 filter criteria
- Real-time chart updates based on filters
- Summary metrics dynamically calculated
- Sortable table with color-coded values
- Clear All button for quick filter reset
- Clickable symbol links for drill-down

**UI/UX Observations:**
- Clean, focused analytics interface
- Filter panel prominently placed for easy access
- Chart provides visual monthly trend
- Table shows detailed line items
- Consistent dark theme
- Type badges use single letter for compactness
- Performance metrics at a glance in summary panel

**Data Integrity:**
- All filters work independently and in combination
- Metrics accurately reflect filtered data
- Chart and table stay synchronized
- Date range validation prevents invalid ranges

**Business Logic Location:**
- Frontend: Filter state management, table sorting, chart rendering
- Backend: All calculations (P&L, aggregations), filtered data queries
- Mixed: Date validation (frontend + backend)

---

## Technical Notes

**Backend Requirements:**
- Options list query with all filter parameters
- Monthly aggregation query with filter support
- Summary metrics calculation with filter support
- **🆕 V2:** Account list fetch for new dropdown
- **🆕 V2:** All queries accept account_id filter parameter

**Frontend Capabilities:**
- Multi-field filter form management
- Client-side table sorting
- Chart.js rendering and updates
- Tab switching (Options/Data)
- Filter state persistence during session
- **🆕 V2:** Account filter state management
- **🆕 V2 DEPRECATED:** Remove Data tab rendering

**Performance Considerations:**
- Large result sets (hundreds of options) may impact load time
- Chart rendering with 12 months of data is performant
- Consider implementing pagination for table (currently loads all)
- Filter queries should use indexed columns (symbol, type, expiration, opened, closed)
- **🆕 V2:** Account filtering reduces query result size (performance benefit)
- Clear All provides quick reset without multiple server roundtrips

**Query Optimization:**
- Single query for table data
- Single query for chart data
- Single query for summary metrics
- All three queries use same base filters
- Consider query result caching if filters unchanged

---

## V2 Change Notation

**Format Guide:**
- **🆕 V2 NEW COMPONENT** - Entirely new section or major component
- **🆕 V2:** Inline note - Enhancement or modification to existing feature
- **🆕 V2 DEPRECATED** - Feature being removed
- All V2 changes summarized in table at top of document

**Change Types:**
- **NEW** - Brand new component/feature
- **ENHANCED** - Improvement to existing feature
- **MODIFIED** - Change to existing behavior
- **DEPRECATED** - Feature being removed or replaced