# Treasuries and Cash Page

**Page Type:** Management Page

**Navigation:** Main Menu → Treasuries & Cash

**Purpose:** Manage Treasury securities holdings and cash transactions with performance tracking and comprehensive transaction history.

---

## V2 Changes Summary

| Component | Change Type | Description |
|-----------|-------------|-------------|
| Page Title | **MODIFIED** | Rename from "Performance Summary" to "Treasuries and Cash" |
| Account Selector | **NEW** | Add account filtering dropdown between metrics and content sections |
| Cash Section | **NEW** | Add cash management card with transaction list and Add button |
| Cash Transaction Modal | **NEW** | Modal for adding/editing cash transactions (Deposit, Withdrawal, Interest) |
| New Treasury Button | **MODIFIED** | Move from page header to "Bonds Held" section header |
| Performance Metrics | **ENHANCED** | Metrics calculate based on selected account filter |

---

## Page Structure

### Page Header
**Page Title:** "Performance Summary" **🆕 V2:** Rename to "Treasuries and Cash"

**Page Actions:**
- "+ New Treasury" button (blue) **🆕 V2:** MOVED to Bonds Held section header

---

### Performance Summary Panel
**Purpose:** Display aggregate metrics for treasury and cash performance.

**Component Type:** Horizontal metrics bar (4 metrics)

**Metrics Displayed:**
- **Total Interest Earned** - Cumulative interest from treasuries **🆕 V2:** + interest income from cash
- **Average Return** - Average yield percentage across holdings
- **Currently Held** - Current value of all treasuries **🆕 V2:** + cash balance
- **Active Positions** - Count of active treasury holdings **🆕 V2:** (cash not included in count)

**Data Requirements:**
- Aggregate treasury data: SUM(interest), AVG(yield), SUM(current_value), COUNT(active)
- **🆕 V2:** Cash balance from cash transactions
- **🆕 V2:** Interest income from cash transactions
- **🆕 V2:** Filter by account_id if account selected

**Business Rules:**
- Total Interest Earned = SUM(interest earned) from treasuries **🆕 V2:** + SUM(interest income) from cash
- Average Return = AVG(yield) for held treasuries
- Currently Held = SUM(current_value) for active treasuries **🆕 V2:** + current cash balance
- Active Positions = COUNT(*) WHERE status = 'HELD'
- **🆕 V2:** All metrics filter by selected account when account filter active

---

### Account Selector
**🆕 V2 NEW COMPONENT**

**Purpose:** Allow users to filter all page data by specific account or view aggregated data across all accounts.

**Component Type:** Dropdown selector

**Position:** Between Performance Summary Panel and Content Sections (full-width card)

**Display Options:**
- Default: "-- All Accounts --"
- Account List: All user accounts (e.g., "Primary Account", "IRA Account")

**Optional Enhancement:** Account Summary Card
- Account Name
- Account Type (Cash, Margin, IRA)
- Total Treasury + Cash Value for this account
- Account Status (Active/Archived)

**Behavior:**
- **"-- All Accounts --" selected:** 
  - Shows aggregated data across all accounts
  - Performance metrics show combined values
  - Both Cash and Bonds sections show all data
  
- **Specific Account selected:**
  - Performance metrics recalculate for selected account only
  - Cash section filters: WHERE account_id = [selected_account]
  - Bonds Held section filters: WHERE account_id = [selected_account]

**Data Requirements:**
- List of all accounts for dropdown: `SELECT id, name, account_type FROM accounts WHERE status = 'ACTIVE' ORDER BY name`
- Selected account persists in session/state

**Business Rules:**
- Account filter applies to ALL sections simultaneously
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

---

### Content Sections

#### Section 1: Cash
**🆕 V2 NEW SECTION**

**Purpose:** Track cash balance and all cash transaction activity (deposits, withdrawals, interest income).

**Component Type:** Card with summary and transaction table

**Section Title:** "Cash"

**Section Header:**
- Section Title: "Cash"
- Total Cash Display: "$[X,XXX.XX]" (prominent display of current cash balance)
- Section Actions: "+ Add Cash Transaction" button

**Table Columns:** (6 columns)
1. Date - Transaction date
2. Type - Transaction type (Deposit, Withdrawal, Interest Income)
3. Amount - Dollar amount (signed: + for deposits/interest, - for withdrawals)
4. Account - Account name (if "All Accounts" selected; hidden if specific account)
5. Notes - Optional transaction description
6. Actions - Edit and Delete icons

**Row Actions:**
- Edit icon (pencil) - Opens Edit Cash Transaction Modal (pre-populated)
- Delete icon (trash) - Opens Delete Confirmation Modal

**Data Requirements:**
- All cash transactions for account(s)
- Current cash balance = SUM(all transaction amounts)
- **🆕 V2:** Filter by account_id if account selected
- Sorted by date DESC (most recent first)

**Business Rules:**
- Cash Balance = SUM(amount) for all transactions
- Deposits: Positive amount
- Withdrawals: Negative amount
- Interest Income: Positive amount
- Account column only shown when "All Accounts" selected
- Transactions cannot be deleted if referenced by other entities (business rule TBD)

**Interactions:**
- Click "+ Add Cash Transaction" → Opens Add Cash Transaction Modal
- Click Edit icon → Opens Edit Cash Transaction Modal with form pre-populated
- Click Delete icon → Opens Confirmation Modal, then deletes record
- Click column header → Sort table by that column

**Empty State:** "No cash transactions. Add your first transaction" (with link to open modal)

**Validation:**
- Date required
- Amount required, must be non-zero
- Type required (Deposit, Withdrawal, Interest Income)

---

#### Cash Transaction Modal (Add/Edit)
**🆕 V2 NEW COMPONENT**

**Purpose:** Create or edit cash transaction records.

**Component Type:** Form Modal

**Modal Title:** 
- "Add Cash Transaction" (when creating)
- "Edit Cash Transaction" (when editing)

**Form Fields:**

1. **Account** - Dropdown (required)
   - List of user's accounts
   - Default: Currently selected account from page (if specific account)
   - Not editable when editing existing transaction (locked to original account)

2. **Transaction Type** - Dropdown (required)
   - Options: "Deposit", "Withdrawal", "Interest Income"
   - Default: None

3. **Date** - Date Picker (required)
   - Format: mm/dd/yyyy
   - Default: Today

4. **Amount** - Number Input (required)
   - Format: Currency ($X,XXX.XX)
   - Positive numbers only (sign determined by transaction type)
   - Validation: Must be > 0

5. **Notes** - Text Area (optional)
   - Multi-line text for transaction description
   - Max 500 characters

**Modal Actions:**
- "Save" button (primary) - Submit form
- "Cancel" button (secondary) - Close modal without saving

**Data Requirements:**
- Account list for dropdown
- Existing transaction data (for edit mode)

**Business Rules:**
- Amount stored as signed value:
  - Deposit: Store as positive
  - Withdrawal: Store as negative
  - Interest Income: Store as positive
- Date cannot be in future
- Account required and immutable after creation
- Transaction type cannot be changed after creation (delete and recreate instead)

**Frontend Validation:**
- All required fields must be filled
- Amount must be positive number
- Date must be valid date, not future
- Notes max length 500 characters

**Backend Validation:**
- Account must exist and be active
- Amount must be non-zero
- Transaction type must be valid enum value
- Duplicate detection (same date, amount, type) - warn user?

**Interactions:**
- Open modal → Form fields clear (add) or pre-populate (edit)
- Change transaction type → No other fields affected
- Submit valid form → Save transaction, close modal, refresh page data
- Submit invalid form → Show inline errors, keep modal open
- Click Cancel → Close modal without saving

**Success Message:** "Cash transaction saved successfully" (toast notification)

**Error Handling:** Display field-level errors inline below each field

---

#### Section 2: Bonds Held
**Purpose:** Display all treasury securities holdings with performance metrics and management actions.

**Component Type:** Card with data table

**Section Title:** "Bonds Held"

**Section Header:**
- Section Title: "Bonds Held"
- **🆕 V2:** Section Actions: "+ New Treasury" button (moved from page header)

**Table Columns:** (11 columns)
1. CUSPID - Treasury security identifier (sortable)
2. PURCHASED - Purchase date (sortable)
3. MATURITY DATE - Maturity date (sortable)
4. REMAINING - Days until maturity (sortable)
5. AMOUNT - Face value amount (sortable)
6. YIELD - Annual yield percentage (sortable)
7. BUY PRICE - Purchase price (sortable)
8. CURRENT VALUE - Current market value (sortable)
9. EXIT PRICE - Sale price if sold (sortable)
10. PROFIT/LOSS - Realized or unrealized gain/loss (sortable)
11. ACTIONS - Edit and Delete icons

**Row Actions:**
- Edit icon (pencil) - Opens Edit Treasury Modal (pre-populated)
- Delete icon (trash) - Opens Delete Confirmation Modal

**Data Requirements:**
- All treasury records for account(s)
- **🆕 V2:** Filter by account_id if account selected
- Calculated fields: REMAINING (days to maturity), PROFIT/LOSS

**Business Rules:**
- REMAINING = Days between current date and maturity date
  - Color code: Red if < 30 days, Yellow if < 90 days, Green otherwise
- PROFIT/LOSS = (Current Value or Exit Price) - Buy Price
- Current Value used if still held
- Exit Price used if sold
- Sort by maturity date ascending by default
- **🆕 V2:** When account filter active, only show treasuries for that account

**Interactions:**
- Click column header → Sort table by that column (ascending/descending toggle)
- Click Edit icon → Opens Edit Treasury Modal with form pre-populated
- Click Delete icon → Opens Confirmation Modal, then deletes record
- Click "+ New Treasury" → Opens Add Treasury Modal
- **🆕 V2:** All actions respect current account filter

**Empty State:** "No treasuries found. Add your first treasury" (with clickable link to open modal)

**Color Coding:**
- REMAINING: Red (< 30 days), Yellow (< 90 days), Green (90+ days)
- PROFIT/LOSS: Green (positive), Red (negative), Gray (zero)

**Validation:** N/A (display only, validation occurs in modals)

---

## Current State Notes

**Working Features:**
- Performance summary metrics at top
- Sortable treasury table with all key fields
- Edit/Delete actions on treasury rows
- Empty state with helpful call-to-action
- Color-coded remaining days indicator
- New Treasury button prominent in header

**UI/UX Observations:**
- Clean, focused interface for treasury management
- All key treasury data visible in single table
- Performance metrics provide quick portfolio overview
- Consistent dark theme
- Empty state encourages first action

**Data Integrity:**
- Performance metrics calculated from treasury data
- Remaining days calculated dynamically
- Profit/loss based on current vs purchase price
- All sortable columns for flexible analysis

**Business Logic Location:**
- Frontend: Table sorting, modal management, empty state display
- Backend: All calculations (interest, yield, remaining days, P&L), data fetching
- Mixed: Form validation (frontend + backend)

---

## Technical Notes

**Backend Requirements:**
- Treasury list query with calculated fields (remaining, P&L)
- Performance metrics aggregation query
- **🆕 V2:** Account list fetch for dropdown
- **🆕 V2:** Cash transactions CRUD API
- **🆕 V2:** Cash balance calculation endpoint
- **🆕 V2:** All queries accept account_id filter parameter

**Frontend Capabilities:**
- Client-side table sorting
- Modal management (open/close)
- Form submission and validation
- Delete confirmation flow
- Empty state rendering
- **🆕 V2:** Account selector state management
- **🆕 V2:** Cash transaction form with type-specific logic

**Performance Considerations:**
- Treasury table loads all records (reasonable for most portfolios)
- Consider pagination if > 100 treasuries
- Performance metrics query separate from table data
- **🆕 V2:** Cash transaction query separate from treasury query
- **🆕 V2:** Account filtering reduces query size (performance benefit)

**Data Model Notes:**
- **🆕 V2:** Cash transactions stored in transactions table
- **🆕 V2:** Transaction types: DEPOSIT, WITHDRAWAL, INTEREST_INCOME
- **🆕 V2:** Cash balance derived, not stored (SUM of transactions)
- Treasuries remain in treasuries table (existing model)

---

## V2 Change Notation

**Format Guide:**
- **🆕 V2 NEW COMPONENT** - Entirely new section or major component
- **🆕 V2 NEW SECTION** - Entirely new content section
- **🆕 V2:** Inline note - Enhancement or modification to existing feature
- **🆕 V2:** MOVED - Component relocated
- All V2 changes summarized in table at top of document

**Change Types:**
- **NEW** - Brand new component/feature
- **ENHANCED** - Improvement to existing feature
- **MODIFIED** - Change to existing behavior
- **DEPRECATED** - Feature being removed or replaced