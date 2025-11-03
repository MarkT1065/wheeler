# Dividends Page

**Page Type:** Management Page

**Navigation:** Main Menu → Dividends

**Purpose:** Track dividend-paying stock positions, analyze dividend income by time period and symbol, view payment calendar, and manage dividend payment history.

---

## V2 Changes Summary

| Component | Change Type | Description |
|-----------|-------------|-------------|
| Account Selector | NEW | Add account filtering between summary metrics and content |
| Summary Metrics Panel | MODIFIED | All metrics filter by selected account |
| Date Range Selector | NEW | Add Last Year / Year to Date / All Time selector for Dividends section |
| By Ticker Chart | MODIFIED | Rename to "By Symbol" for consistency |
| By Month Chart | ENHANCED | Filter by date range and account |
| By Symbol Chart | ENHANCED | Filter by date range and account |
| Dividend Calendar | ENHANCED | Filter by account, show payments for selected range |
| Position Details | ENHANCED | Filter by account |
| Annual Income Calculation | ENHANCED | Support weekly, monthly, quarterly dividend frequencies |
| Dividend Frequency Support | NEW | Add frequency field to dividends (weekly/monthly/quarterly/annual) |

---

## Page Structure

### Page Header
**Purpose:** Display page title and primary actions  
**Component Type:** Header  
**Position:** Top of page, full width

**Elements:**
- Page Title: "Dividends"
- Action Buttons: "+ Add Position" button (top right)

🆕 **V2:** Page header remains unchanged

---

### Summary Metrics Panel
**Purpose:** Display key portfolio-wide dividend metrics  
**Component Type:** Metrics Card  
**Position:** Below page header, full width, dark card with 4 metrics

**Metrics:**

1. **Total Positions**
   - **Display:** Count of dividend-paying positions
   - **Format:** Integer
   - **Color:** White text
   - **Calculation:** Count of distinct symbols with active dividend positions

2. **Annual Income**
   - **Display:** Annualized dividend income projection
   - **Format:** Currency (green text)
   - **Color:** Green (#00ff00 style)
   - **Calculation:** Sum of (shares × annual_dividend_per_share) for all positions
   - **Business Rule:** Needs clarification on source - likely calculated from:
     - Most recent dividend amount × frequency per year × shares held
     - OR: User-entered annual dividend rate × shares held

3. **Average Yield**
   - **Display:** Weighted average yield across all positions
   - **Format:** Percentage (1 decimal place)
   - **Color:** White text
   - **Calculation:** (Total Annual Income / Total Position Value) × 100

4. **Total Paid (All Time)**
   - **Display:** Cumulative dividend payments received
   - **Format:** Currency (green text)
   - **Color:** Green
   - **Calculation:** Sum of all dividend payment amounts in database

🆕 **V2 ENHANCED:**
- All metrics filter by selected account (from Account Selector)
- When "All Accounts" selected: Aggregate across all accounts
- When specific account selected: Show metrics for that account only
- Annual Income calculation enhanced to support multiple dividend frequencies

---

### 🆕 V2 NEW COMPONENT: Account Selector
**Purpose:** Filter all page data by account  
**Component Type:** Dropdown Selector Card  
**Position:** Between Summary Metrics Panel and Dividends Section, full width

**Behavior:**
- **Default:** "-- All Accounts --"
- **Options:** List all active user accounts
- **Impact:** Filters Summary Metrics, all charts, calendar, and Position Details
- **Persistence:** Selected account persists across page navigation

**Frontend:**
- Dropdown component in full-width card
- Triggers page data refresh on change

**Backend:**
- All queries accept optional `account_id` parameter
- `NULL` account_id = aggregate across all accounts
- Specific account_id filters: `WHERE account_id = ?`

---

## Content Sections

### Dividends Section
**Purpose:** Visualize dividend income by time period and symbol distribution  
**Component Type:** Card with two charts side-by-side  
**Position:** Below Account Selector, full width card

**Layout:**
- Section Title: "Dividends"
- 🆕 **V2 NEW:** Date Range Selector (right side of section header)
- Two charts side-by-side:
  - By Month (left, ~50% width)
  - By Symbol (right, ~50% width)

---

#### 🆕 V2 NEW COMPONENT: Date Range Selector
**Purpose:** Filter dividend charts by time period  
**Component Type:** Dropdown or Button Group  
**Position:** Section header, right side

**Options:**
- **Last Year:** Previous 12 calendar months from today
- **Year to Date:** January 1 of current year through today
- **All Time:** All dividend history in database

**Default:** Year to Date

**Behavior:**
- Filters both "By Month" and "By Symbol" charts
- Does NOT affect Summary Metrics (those remain "All Time")
- Does NOT affect Dividend Calendar (calendar shows future payments)
- Selection persists within session

**Frontend:**
- Button group or dropdown selector
- Updates charts on selection change

**Backend:**
- Queries accept `date_range` parameter: 'last_year', 'ytd', 'all_time'
- Server calculates date boundaries based on current date
- Returns filtered dividend data

---

#### By Month Chart
**Purpose:** Show monthly dividend income over time, broken down by symbol  
**Component Type:** Stacked Bar Chart  
**Position:** Left side of Dividends section (~50% width)

**Title:** "By Month"

**Chart Configuration:**
- **Type:** Stacked Bar Chart (vertical bars)
- **X-Axis:** Month abbreviations (Mar, May, Jul, Aug, Sep, Oct, etc.)
- **Y-Axis:** Dollar amount ($0 to $1200+ scale)
- **Data Series:** One series per symbol (stacked within each bar)
- **Colors:** Distinct color per symbol (consistent with By Symbol chart)
- **Value Labels:** Display total amount at top of each bar

**Data Requirements:**
- **Query:** Get dividend payments grouped by month and symbol
- **Filters:** 
  - Date range (from Date Range Selector)
  - Account (from Account Selector)
- **Aggregation:** Sum of payment amounts by month by symbol

**Business Rules:**
- Only show months with dividend payments (no zero bars)
- Stack bars by symbol from bottom to top
- Color coding consistent with legend and By Symbol chart
- X-axis shows only months with data (not continuous timeline)

**Interactions:**
- **Hover:** Show tooltip with:
  - Month and Year
  - Symbol
  - Payment amount
  - Total for month
- **Click:** No click interaction (read-only visualization)

**Empty State:**
- Message: "No dividend payments in selected period"
- Display when no data for selected date range/account

🆕 **V2 ENHANCED:**
- Filter by date range (Last Year / YTD / All Time)
- Filter by account
- X-axis dynamically shows months with data in selected range
- Handle weekly/monthly dividend frequencies (may show same month multiple times)

**Current State Notes:**
- Currently shows 6 months of data (Mar, May, Jul, Aug, Sep, Oct)
- Stacked bars showing 4 symbols: FTAI, TSLL, TSLY, ULTY
- Values visible: $60, $60, $257, $1,147, $1,150+

---

#### By Symbol Chart (formerly "By Ticker")
**Purpose:** Show distribution of dividend income across symbols  
**Component Type:** Pie Chart  
**Position:** Right side of Dividends section (~50% width)

**Title:** "By Symbol" 🆕 **V2 MODIFIED:** (was "By Ticker")

**Chart Configuration:**
- **Type:** Pie Chart
- **Data:** Percentage/amount of dividends by symbol
- **Colors:** Distinct color per symbol (matches By Month chart colors)
- **Legend:** Symbol list with color coding (right side)
- **Value Labels:** Dollar amount displayed on largest slices

**Data Requirements:**
- **Query:** Get total dividend payments by symbol
- **Filters:**
  - Date range (from Date Range Selector)
  - Account (from Account Selector)
- **Aggregation:** Sum of payment amounts by symbol

**Business Rules:**
- Calculate percentage: (Symbol Total / Grand Total) × 100
- Sort slices by size (largest to smallest)
- Color coding consistent with By Month chart
- Show dollar amount on slice if space permits

**Legend:**
- Position: Right side of chart
- Format: Color box + Symbol ticker
- Order: Match slice order (largest to smallest)

**Interactions:**
- **Hover:** Show tooltip with:
  - Symbol ticker
  - Dollar amount
  - Percentage of total
- **Click:** No click interaction (read-only visualization)

**Empty State:**
- Message: "No dividend payments in selected period"
- Display when no data for selected date range/account

🆕 **V2 ENHANCED:**
- Rename from "By Ticker" to "By Symbol"
- Filter by date range (Last Year / YTD / All Time)
- Filter by account
- Recalculate percentages based on filtered data

**Current State Notes:**
- Shows 4 symbols: TSLL (blue), FTAI (green), TSLY (orange), ULTY (red)
- ULTY appears to be largest slice ($1,080 visible)
- Legend displays on right side with color coding

---

### Dividend Calendar
**Purpose:** Visual calendar showing dividend payment dates  
**Component Type:** Calendar Grid  
**Position:** Right side of page, aligned with Dividends section

**Title:** "Dividend Calendar"

**Calendar Configuration:**
- **Display:** Two-month view (current month + next month)
- **Format:** Standard monthly calendar grid
- **Headers:** Sun, Mon, Tue, Wed, Thu, Fri, Sat
- **Date Highlighting:** Dates with dividend payments highlighted in green

**Data Requirements:**
- **Query:** Get upcoming dividend payment dates (ex-dividend dates or payment dates)
- **Filters:** Account (from Account Selector)
- **Time Range:** Current month + next month (rolling window)

**Business Rules:**
- Highlight dates with expected dividend payments
- Color: Green highlight (#00ff00 style) for dividend dates
- Multiple dividends on same date: Single highlight (tooltip shows all)
- Show current date indicator (today)

**Date Display:**
- Current month: Full month name + year (e.g., "November 2025")
- Next month: Full month name + year (e.g., "December 2025")
- Days of week: Abbreviated (Sun-Sat)
- Date numbers: Standard calendar layout

**Interactions:**
- **Hover on highlighted date:** Show tooltip with:
  - Symbol(s) with payments
  - Expected payment amount(s)
  - Payment type (ex-dividend date or payment date)
- **Click on date:** Optional - could navigate to detail view or filter Position Details

**Empty State:**
- Calendar still displays with no highlights
- Message: "No upcoming dividend payments" (if no dates in next 2 months)

🆕 **V2 ENHANCED:**
- Filter by selected account
- Support for weekly/monthly/quarterly dividend frequencies
- Calculate expected payment dates based on frequency and last payment
- Handle multiple payments per month (weekly/monthly dividends)

**Current State Notes:**
- Shows November 2025 and December 2025
- November 3rd highlighted in green
- December 23rd highlighted in green
- Clean calendar layout with day-of-week headers

---

### Position Details Section
**Purpose:** Display detailed information for each dividend-paying position  
**Component Type:** Expandable Accordion List  
**Position:** Below Dividends section and Calendar, full width

**Title:** "Position Details"

**Accordion Row (Collapsed):**
Each position displays in a collapsed row showing:

1. **Symbol** (left, bold, colored - e.g., blue for TSLL)
2. **Shares** (label + value)
3. **Annual Income** (label + currency value)
4. **Yield** (label + percentage)
5. **Ex-Div Date** (label + date MM/DD format)
6. **Expand/Collapse Icon** (right side, chevron down/up)

**Layout:** Horizontal layout with labels and values, evenly spaced

**Accordion Row (Expanded):**
Shows additional details (TBD - current screenshot only shows collapsed state):
- Payment history table
- Dividend frequency
- Last payment date and amount
- Next expected payment date
- Edit/Delete actions

**Data Requirements:**
- **Query:** Get all dividend positions with summary data
- **Fields:**
  - Symbol ticker
  - Shares held
  - Annual income (calculated or stored)
  - Yield percentage
  - Ex-dividend date (next expected)
- **Filters:** Account (from Account Selector)
- **Sort:** Default sort by symbol alphabetically

**Business Rules:**
- One row per symbol per account
- Annual Income: shares × annual dividend per share
- Yield: (annual dividend per share / current stock price) × 100
- Ex-Div Date: Next upcoming ex-dividend date

**Interactions:**
- **Click row:** Expand/collapse accordion
- **Click symbol:** Navigate to Symbol Detail page (optional)
- **Edit button:** Open Edit Position modal (in expanded view)
- **Delete button:** Delete position with confirmation (in expanded view)

**Empty State:**
- Message: "No dividend positions found"
- Action: "+ Add Position" button

🆕 **V2 ENHANCED:**
- Filter by selected account
- Expanded view includes dividend payment history table
- Support for dividend frequency display (weekly/monthly/quarterly/annual)
- Edit modal includes account selector

**Current State Notes:**
- Shows 3 positions: TSLL, TSLY, ULTY
- Example values:
  - TSLL: 500 shares, $180 annual income, 1.79% yield, 12/23 ex-div
  - TSLY: 300 shares, $180 annual income, 7.28% yield, 10/23 ex-div
  - ULTY: 3,000 shares, $1,080 annual income, 7.39% yield, 10/22 ex-div

---

## Modal: Add/Edit Position

**Modal Title:** 
- Add: "Add Dividend Position"
- Edit: "Edit Dividend Position"

**Form Fields:**

1. **🆕 V2 NEW: Account** (Dropdown, required)
   - Options: List of user accounts
   - Default: Currently selected account (from Account Selector)
   - Validation: Required

2. **Symbol** (Text input, required)
   - Format: Uppercase ticker symbol
   - Validation: Required, 1-5 characters, uppercase
   - Autocomplete: Optional - search existing symbols

3. **Shares** (Number input, required)
   - Format: Integer (whole shares) or decimal (fractional shares)
   - Validation: Required, positive number
   - Default: Empty

4. **🆕 V2 NEW: Dividend Frequency** (Dropdown, required)
   - Options: Weekly, Monthly, Quarterly, Annual
   - Default: Quarterly
   - Validation: Required

5. **Dividend Per Share** (Currency input, required)
   - Format: Decimal (2 places)
   - Label: "Dividend Amount (per period)" 🆕 V2: Dynamic label based on frequency
   - Validation: Required, positive number
   - Help text: "Enter the dividend amount per share per payment period"

6. **🆕 V2 ENHANCED: Annual Dividend** (Currency, calculated/editable)
   - Format: Decimal (2 places)
   - Calculation: Dividend Per Share × Frequency multiplier
     - Weekly: × 52
     - Monthly: × 12
     - Quarterly: × 4
     - Annual: × 1
   - Editable: Allow manual override if calculated value is incorrect
   - Validation: Positive number

7. **Current Price** (Currency input, optional)
   - Format: Decimal (2 places)
   - Validation: Positive number if provided
   - Use: Calculate yield (yield = annual dividend / current price × 100)

8. **Ex-Dividend Date** (Date picker, optional)
   - Format: MM/DD/YYYY
   - Default: Empty
   - Use: Populate calendar and calculate next expected payments

9. **Notes** (Text area, optional)
   - Format: Free text
   - Max length: 500 characters
   - Use: Additional context or reminders

**Modal Actions:**
- **Cancel:** Close modal without saving
- **Save:** Validate and save position

**Frontend Validation:**
- All required fields must be filled
- Symbol format validation (uppercase, length)
- Positive number validation for numeric fields
- Date format validation

**Backend Validation:**
- Duplicate symbol check per account (one position per symbol per account)
- Valid account_id
- Data type and range validation

**Success:**
- Close modal
- Refresh page data (metrics, charts, position list)
- Show success toast: "Position added/updated successfully"

**Error:**
- Display inline error messages
- Keep modal open
- Show specific validation errors

🆕 **V2 NEW/ENHANCED:**
- Account selector field added
- Dividend frequency field added
- Annual dividend calculation based on frequency
- Support for weekly/monthly dividend schedules
- Help text and labels adjust based on frequency selection

---

## Modal: Delete Position Confirmation

**Modal Title:** "Delete Position?"

**Content:**
- Message: "Are you sure you want to delete the [SYMBOL] dividend position?"
- Warning: "This will remove the position but retain all historical payment records."

**Modal Actions:**
- **Cancel:** Close modal without deleting
- **Delete:** Delete position, close modal, refresh page

**Backend:**
- Soft delete or hard delete (TBD based on data model)
- Retain payment history even if position deleted
- Remove from Position Details list

---

## Annual Income Calculation - Clarification Needed

**Current State Issue:** Source of "Annual Income" metric is unclear. Possible sources:

1. **Calculated from most recent payments:**
   - Take most recent dividend payment
   - Multiply by frequency (4 for quarterly, 12 for monthly, etc.)
   - Multiply by shares held
   - Formula: `last_payment × frequency × shares`

2. **User-entered annual rate:**
   - User enters expected annual dividend per share
   - System calculates: `annual_rate × shares`

3. **Stored per position:**
   - Annual dividend per share stored in position record
   - Updated manually or via API
   - Formula: `stored_annual_dividend × shares`

**🆕 V2 Recommendation:**
- Add `dividend_frequency` field to dividends table (WEEKLY, MONTHLY, QUARTERLY, ANNUAL)
- Add `dividend_per_period` field (amount per payment period)
- Calculate `annual_dividend = dividend_per_period × frequency_multiplier`
- Store calculated value but allow manual override
- Recalculate when new payments recorded

**Backend Enhancement:**
- When dividend payment recorded, update position's `dividend_per_period` if higher
- Calculate new annual projection
- Support for special dividends (one-time, not included in annual calculation)

---

## Dividend Frequency Support - V2 Enhancement

**Current State:**
- System appears to assume quarterly dividends (standard for most stocks)
- No explicit frequency field in data model

**🆕 V2 Enhancement Requirements:**

### 1. Data Model Changes
- Add `frequency` enum field to dividends table: WEEKLY, MONTHLY, QUARTERLY, ANNUAL
- Add `dividend_per_period` decimal field
- Add `last_payment_date` date field
- Add `next_expected_date` date field (calculated)

### 2. UI Changes
- Display frequency in Position Details expanded view
- Dividend Calendar shows all payment dates (not just quarterly)
- By Month chart handles multiple payments per month

### 3. Calculation Changes
- Annual income: `dividend_per_period × frequency_multiplier × shares`
  - Weekly: ×52
  - Monthly: ×12
  - Quarterly: ×4
  - Annual: ×1
- Next payment date: `last_payment_date + period_days`
  - Weekly: +7 days
  - Monthly: +30 days (or same day next month)
  - Quarterly: +90 days (or ~3 months)
  - Annual: +365 days (or same date next year)

### 4. Calendar Population
- For weekly/monthly dividends, populate all expected payment dates in calendar
- Use last payment date + frequency to project future payments
- Handle month-end edge cases (e.g., monthly on 31st → becomes 30th in some months)

### 5. Chart Adjustments
- By Month chart may show multiple bars per month for weekly payers
- Consider grouping options: by symbol, by payment date, by week
- Aggregate properly in "All Time" and "YTD" views

### 6. Business Rules
- Validate frequency selection matches historical payment pattern
- Warn user if frequency changes (potential data inconsistency)
- Support "irregular" frequency for special dividends or one-time payments

---

## Payment History (Expanded Accordion) - Specification TBD

**Purpose:** Show detailed payment history for a position  
**Component Type:** Data Table  
**Position:** Within expanded accordion row in Position Details

**Suggested Columns:**
1. Payment Date
2. Ex-Dividend Date
3. Amount per Share
4. Total Amount (shares × amount per share)
5. Status (Paid, Expected, Missed)
6. Actions (Edit, Delete)

**Data Requirements:**
- Query: Get all dividend payments for symbol and account
- Sort: Default descending by payment date (most recent first)

**Interactions:**
- Edit payment: Open edit modal
- Delete payment: Confirm and delete
- Add payment: "+ Add Payment" button in section header

**🆕 V2:**
- Filter by account
- Support for multiple payment frequencies
- Calculate expected payments based on frequency

**Note:** Current screenshot does not show expanded state, so this section is speculative and should be refined when expanded view is documented.

---

## Current State Notes

### Working Features:
- Summary metrics display correctly
- By Month stacked bar chart with multi-symbol breakdown
- By Ticker pie chart (should be "By Symbol")
- Dividend Calendar with highlighted payment dates
- Position Details accordion with summary data
- Clean, dark-themed UI consistent with rest of app

### Data Observations:
- 3 positions currently displayed
- Annual income: $1,440 total across all positions
- Average yield: 5.49%
- Total paid all time: $3,872
- Payment history spans at least 6 months (Mar-Oct visible in chart)

### UI/UX Observations:
- Good visual hierarchy with metrics at top
- Charts side-by-side for comparison
- Calendar provides forward-looking view
- Accordion allows compact display with details on demand
- Color coding consistent between charts and legend

### Business Logic Location:
- Annual income calculation: Likely backend (needs clarification)
- Yield calculation: Likely backend based on current prices
- Calendar date population: Backend projects future payments
- Chart aggregations: Backend queries with GROUP BY

---

## Technical Notes

### Backend Requirements:

**Database Schema:**
- `dividends` table fields:
  - `id` (primary key)
  - `account_id` (foreign key) 🆕 V2
  - `symbol` (ticker)
  - `shares` (decimal)
  - `dividend_per_period` (decimal) 🆕 V2
  - `frequency` (enum: WEEKLY, MONTHLY, QUARTERLY, ANNUAL) 🆕 V2
  - `annual_dividend` (decimal, calculated)
  - `current_price` (decimal, nullable)
  - `yield` (decimal, calculated)
  - `ex_dividend_date` (date)
  - `last_payment_date` (date) 🆕 V2
  - `notes` (text, nullable)
  - `created_at`, `updated_at`

- `dividend_payments` table (for payment history):
  - `id` (primary key)
  - `dividend_id` (foreign key)
  - `payment_date` (date)
  - `ex_dividend_date` (date)
  - `amount_per_share` (decimal)
  - `total_amount` (decimal)
  - `shares` (decimal, snapshot at payment time)
  - `status` (enum: PAID, EXPECTED, MISSED)
  - `notes` (text, nullable)
  - `created_at`, `updated_at`

**API Endpoints:**
- `GET /api/dividends/summary` - Summary metrics (filtered by account)
- `GET /api/dividends/by-month` - Chart data (filtered by account, date range)
- `GET /api/dividends/by-symbol` - Chart data (filtered by account, date range)
- `GET /api/dividends/calendar` - Upcoming payment dates (filtered by account)
- `GET /api/dividends/positions` - Position list (filtered by account)
- `POST /api/dividends/positions` - Create position
- `PUT /api/dividends/positions/:id` - Update position
- `DELETE /api/dividends/positions/:id` - Delete position
- `GET /api/dividends/payments/:dividend_id` - Payment history for position
- `POST /api/dividends/payments` - Record new payment
- `PUT /api/dividends/payments/:id` - Update payment
- `DELETE /api/dividends/payments/:id` - Delete payment

**Query Performance:**
- Index on `account_id` for filtering
- Index on `symbol` for lookups
- Index on `payment_date` for time-range queries
- Consider materialized view for summary metrics

**Calculation Logic:**
- Annual dividend: `dividend_per_period × frequency_multiplier`
- Frequency multipliers: {WEEKLY: 52, MONTHLY: 12, QUARTERLY: 4, ANNUAL: 1}
- Yield: `(annual_dividend / current_price) × 100`
- Next payment date: `last_payment_date + period_interval`

### Frontend Requirements:

**State Management:**
- Selected account (from Account Selector)
- Selected date range (Last Year / YTD / All Time)
- Position Details accordion expand/collapse states
- Modal open/close states

**Chart Libraries:**
- Bar chart: Recharts or Chart.js (stacked bars)
- Pie chart: Recharts or Chart.js
- Ensure consistent colors across both charts

**Calendar Component:**
- React Calendar or custom grid component
- Date highlighting logic
- Tooltip on hover with payment details

**Data Refresh:**
- Refresh on account change
- Refresh on date range change
- Refresh after add/edit/delete operations

**Validation:**
- Form validation for Add/Edit modal
- Real-time validation feedback
- Required field indicators

---

## V2 Implementation Priority

**Phase 1: Core Multi-Account Support**
1. Add Account Selector component
2. Update all queries to filter by account
3. Test "All Accounts" aggregation
4. Test specific account filtering

**Phase 2: Date Range Filtering**
1. Add Date Range Selector component
2. Update chart queries to filter by date range
3. Test Last Year, YTD, All Time views
4. Ensure calendar remains unaffected

**Phase 3: Dividend Frequency Enhancement**
1. Add `frequency` and `dividend_per_period` fields to database
2. Update Add/Edit modal with frequency selector
3. Implement annual dividend calculation based on frequency
4. Update calendar to show all payment dates (weekly/monthly)
5. Test with weekly, monthly, quarterly, annual dividends

**Phase 4: Nomenclature and Polish**
1. Rename "By Ticker" to "By Symbol"
2. Verify all color coding consistency
3. Test all empty states
4. Verify all validation messages

**Phase 5: Payment History**
1. Implement payment history table in expanded accordion
2. Add payment CRUD operations
3. Test payment tracking and history display

---

## V2 Change Notation Reference

### Summary Table Change Types:
- **NEW** - Brand new component/feature (Account Selector, Date Range Selector)
- **ENHANCED** - Improvement to existing feature (frequency support, calculations)
- **MODIFIED** - Change to existing behavior (rename "By Ticker")
- **DEPRECATED** - Feature being removed (none on this page)

### Inline Notation:
- 🆕 **V2 NEW COMPONENT:** - Major new feature/section
- 🆕 **V2 NEW:** - New field or element
- 🆕 **V2 ENHANCED:** - Enhancement to existing feature
- 🆕 **V2 MODIFIED:** - Behavior change to existing feature
- 🆕 **V2 DEPRECATED:** - Feature being removed

---

## Questions for Review

1. **Annual Income Source:** Confirm calculation method and data source for "Annual Income" metric
2. **Payment History:** Confirm fields and interactions for expanded accordion state (not visible in screenshot)
3. **Dividend Frequency Priority:** Confirm priority and scope of weekly/monthly dividend support
4. **Calendar Date Type:** Clarify if calendar shows ex-dividend dates or payment dates (or both)
5. **Account Columns:** Confirm no account columns needed in tables when account is selected (implicit from selector)
6. **Edit/Delete Location:** Confirm Edit/Delete actions are in expanded accordion view (not visible in current screenshot)

---

## Testing Checklist

### Current State Verification:
- [ ] Summary metrics display correctly
- [ ] By Month chart shows stacked bars with correct values
- [ ] By Symbol pie chart shows distribution correctly
- [ ] Calendar highlights payment dates
- [ ] Position Details accordion expands/collapses
- [ ] Add Position modal creates new positions
- [ ] Edit Position modal updates existing positions
- [ ] Delete Position removes position with confirmation

### V2 Account Filtering:
- [ ] Account Selector displays all accounts
- [ ] "All Accounts" shows aggregated data
- [ ] Specific account filters all metrics
- [ ] Specific account filters both charts
- [ ] Specific account filters calendar
- [ ] Specific account filters position list
- [ ] Selected account persists across navigation

### V2 Date Range Filtering:
- [ ] Date Range Selector displays three options
- [ ] "Last Year" shows correct 12-month period
- [ ] "Year to Date" shows Jan 1 through today
- [ ] "All Time" shows all historical data
- [ ] Date range filters both charts correctly
- [ ] Date range does NOT affect summary metrics
- [ ] Date range does NOT affect calendar

### V2 Frequency Support:
- [ ] Frequency field in Add/Edit modal
- [ ] Annual dividend calculates correctly for each frequency
- [ ] Weekly dividends show multiple payments per month
- [ ] Monthly dividends show monthly payments
- [ ] Calendar shows all expected payment dates
- [ ] By Month chart handles multiple payments per month

### V2 Nomenclature:
- [ ] "By Ticker" renamed to "By Symbol" everywhere
- [ ] All references use "Symbol" not "Ticker"
- [ ] Consistency across page and modals

---

**Document Status:** Complete - Ready for Review  
**Artifact ID:** `dividends_page_spec`  
**Last Updated:** 2025-11-03
