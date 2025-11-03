# Options Overview Page

**Page Type:** Analytics/Management Page (hybrid)

**Navigation:** Main Menu → Options

**Purpose:** Visualize open options positions by expiration date with interactive charts and provide detailed position management through expandable accordions.

---

## V2 Changes Summary

| Component | Change Type | Description |
|-----------|-------------|-------------|
| Account Selector | **NEW** | Add account filtering dropdown between charts and Open Positions section |
| Charts | **ENHANCED** | Both charts filter by selected account |
| Open Positions | **ENHANCED** | Accordion groups filter by selected account |

---

## Page Structure

### Charts Section
**Purpose:** Provide visual analytics of options positions by expiration date and put exposure risk.

**Layout:** Two charts side-by-side (split view)

---

#### Chart 1: Options by Expiration (Left, ~75% width)
**Purpose:** Display individual option positions as bubbles plotted by expiration date and profit, with bubble size indicating risk/nominal value.

**Component Type:** Bubble Chart (Scatter Plot with sized points)

**Chart Title:** "Options by Expiration"

**Chart Configuration:**
- X-axis: Expiration Date (date scale)
- Y-axis: Individual Option Profit ($ amount, labeled "Individual Option Profit ($)")
- Bubble size: Based on nominal value (strike × contracts × 100)
- Color coding:
  - Blue bubbles: Put Options
  - Green bubbles: Call Options
- Legend: "Put Options" and "Call Options" at top

**Data Requirements:**
- All open options positions
- For each option:
  - Expiration date
  - Calculated profit (premium × contracts × 100 - commission)
  - Nominal value for bubble sizing
  - Option type (Put/Call) for color
- **🆕 V2:** Filter by account_id if account selected

**Business Rules:**
- Only show open positions (closed = NULL)
- Bubble size scaled based on risk exposure:
  - Min size: Small positions ($1K-5K nominal)
  - Max size: Large positions ($50K+ nominal)
  - Puts receive 1.5× size multiplier (higher risk)
- Y-axis represents current profit if position closed today
- Profit = premium received - current cost to close

**Interactions:**
- Hover over bubble → Tooltip shows:
  - Symbol, Type, Strike
  - Contracts
  - Nominal Value
  - Total Profit
- Click bubble → Opens corresponding accordion in Open Positions section below
- Click legend item → Toggle Put/Call visibility

**Empty State:** Chart empty if no open options exist

---

#### Chart 2: Put Exposure (Right, ~25% width)
**Purpose:** Show total capital at risk from put positions grouped by expiration date.

**Component Type:** Bar Chart

**Chart Title:** "Put Exposure"

**Chart Configuration:**
- X-axis: Expiration dates (showing as MM/DD format, e.g., "11/07", "11/14")
- Y-axis: Put Exposure amount ($ scale, e.g., $80K, $70K, $60K)
- Bar color: Blue
- Data labels: Total amount displayed on top of each bar (e.g., "$65K", "$76K")

**Data Requirements:**
- All open put positions grouped by expiration
- For each expiration date:
  - Total exposure = SUM(strike × contracts × 100) for all puts
- **🆕 V2:** Filter by account_id if account selected

**Business Rules:**
- Only include open put positions
- Group by expiration date
- Calculate total capital at risk per expiration
- Exposure = maximum amount required to purchase stock if assigned

**Interactions:**
- Hover over bar → Tooltip shows exact put exposure amount
- Click bar → Opens corresponding accordion in Open Positions section below
- Chart is clickable for drill-down to detail

**Empty State:** Chart empty if no open put positions exist

**Data Labels:**
- Display exposure amount in thousands format ("$65K", "$76K")
- Positioned at top of each bar

---

### Account Selector
**🆕 V2 NEW COMPONENT**

**Purpose:** Allow users to filter all page data by specific account or view aggregated data across all accounts.

**Component Type:** Dropdown selector

**Position:** Between Charts Section and Open Positions Section (full-width card)

**Display Options:**
- Default: "-- All Accounts --"
- Account List: All user accounts (e.g., "Primary Account", "IRA Account", "Margin Account")

**Optional Enhancement:** Account Summary Card
- Account Name
- Account Type (Cash, Margin, IRA)
- Total Open Options Value
- Account Status (Active/Archived)

**Behavior:**
- **"-- All Accounts --" selected:** 
  - Shows aggregated data across all accounts
  - Charts show all open positions
  - Open Positions section shows all positions grouped by expiration
  
- **Specific Account selected:**
  - Both charts recalculate for selected account only
  - Open Positions section filters: WHERE account_id = [selected_account]
  - No account column needed (implicit from selector)

**Data Requirements:**
- List of all accounts for dropdown: `SELECT id, name, account_type FROM accounts WHERE status = 'ACTIVE' ORDER BY name`
- Selected account persists in session/state

**Business Rules:**
- Account filter applies to ALL sections simultaneously (both charts + open positions)
- Account selection persists when navigating to other pages
- Default to "-- All Accounts --" on initial page load
- Cannot filter by archived accounts

**Interactions:**
- User selects account from dropdown → Page reloads/refreshes with filtered data
- Account selection shows in URL or state

**Frontend Validation:**
- Dropdown always has valid selection

**Backend Requirements:**
- Account list API endpoint
- All data queries accept optional account_id filter parameter
- Charts and accordion data filtered by account

---

### Content Sections

#### Section: Open Positions
**Purpose:** Display detailed information for all open options positions, grouped by expiration date in expandable accordions.

**Component Type:** Accordion groups (expandable/collapsible)

**Section Title:** "Open Positions"

**Accordion Groups:** One per expiration date, sorted by expiration (nearest first)

**Accordion Header:** (for each expiration group)
- Expiration icon (calendar)
- "Expiration: [Date]" (e.g., "Expiration: 11/07/2025")
- DTE indicator: "DTE: [N] days" (color-coded by urgency)
  - Critical (red): ≤3 days
  - Warning (yellow): 4-7 days
  - Caution (orange): 8-15 days
  - Safe (green): 16+ days
- Summary metrics:
  - "[N] Positions - $[X] premium"
  - "[N] Call[s] - $[X]" (if calls exist)
  - "[N] Put[s] - $[X]" (if puts exist)
- Expand/collapse toggle (chevron icon)

**Accordion Content:** Interactive table for that expiration date

**Table Columns:** (7 columns minimum visible in screenshot)
1. Symbol - Stock ticker (clickable link)
2. Type - Badge (P for Put, C for Call)
3. Strike - Strike price
4. Quantity - Number of contracts
5. Nominal - Total exposure (strike × contracts × 100)
6. Total Profit - Current P&L
7. Entry Date - Date position opened
8. Actions (implied) - Edit/Delete icons (not visible in screenshot but standard pattern)

**Data Requirements:**
- All open options (WHERE closed IS NULL)
- Grouped by expiration date
- Sorted by expiration ASC (nearest first)
- For each position:
  - Calculate nominal value
  - Calculate current profit
  - Calculate DTE
- **🆕 V2:** Filter by account_id if account selected

**Business Rules:**
- DTE = Days between current date and expiration date
- Nominal for Puts = strike × contracts × 100
- Nominal for Calls = current_price × contracts × 100
- Total Profit = premium × contracts × 100 - commission - current_cost_to_close
- Group positions by exact expiration date
- Show summary metrics in header for quick overview

**Interactions:**
- Click accordion header → Expand/collapse that group
- Only one accordion open at a time (or multiple? - screenshot shows collapsed state)
- Click Symbol link → Navigate to Symbol Detail Page
- Click chart bubble/bar → Auto-expand corresponding accordion and scroll to it
- Edit/Delete actions (if present) → Open modals

**Empty State:** "No open options positions" (if no open positions exist)

**Color Coding:**
- DTE indicator uses color scale (red → yellow → green)
- Type badges: P (blue), C (green)
- Profit values: Green (positive), Red (negative)

---

## Current State Notes

**Working Features:**
- Dual chart layout provides visual and risk analysis
- Bubble chart shows individual positions with risk-sizing
- Bar chart highlights put exposure by date
- Clickable charts drill down to accordion detail
- Accordion headers show comprehensive summaries
- DTE color coding provides urgency indicators
- All data synchronized across charts and accordions

**UI/UX Observations:**
- Clean split-screen chart layout maximizes space
- Bubble sizing conveys risk at a glance
- Put exposure chart focuses on assignment risk
- Accordion groups organize by expiration for easy planning
- Collapsed accordions keep page compact
- Summary metrics in headers reduce need to expand
- Consistent dark theme

**Data Integrity:**
- Charts and accordions show same underlying data
- Click interactions connect visual to detail
- DTE calculations accurate and color-coded
- Exposure calculations based on strike × contracts

**Business Logic Location:**
- Frontend: Chart rendering, accordion expand/collapse, click interactions
- Backend: All calculations (profit, nominal, DTE, exposure), data grouping
- Mixed: Chart click → accordion mapping

---

## Technical Notes

**Backend Requirements:**
- Open options list query grouped by expiration
- Individual position calculations (profit, nominal, DTE)
- Put exposure aggregation by expiration
- Summary metrics per expiration group
- **🆕 V2:** Account list fetch for dropdown
- **🆕 V2:** All queries accept account_id filter parameter

**Frontend Capabilities:**
- Dual chart rendering (Chart.js)
- Bubble chart with dynamic sizing
- Bar chart with data labels
- Accordion expand/collapse management
- Chart click → accordion interaction
- Tooltip rendering
- **🆕 V2:** Account selector state management
- **🆕 V2:** Page refresh on account selection

**Performance Considerations:**
- Bubble chart handles ~50-100 positions well
- Bar chart simple and fast
- Accordion lazy-loads table data on expand (recommended)
- Chart re-render on data change is performant
- **🆕 V2:** Account filtering reduces data volume (performance benefit)

**Chart Click Interaction:**
- Map chart data point to expiration date
- Find matching accordion by expiration
- Scroll accordion into view
- Expand accordion automatically
- Highlight or focus on specific row if symbol clicked

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