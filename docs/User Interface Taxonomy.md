# Wheeler UI Taxonomy & Nomenclature

The purpose of this document is to capture a taxonomy that is used in design documents and Claude conversations to develop the Wheeler v2 user interface.

## 1. Navigation Structure

### 1.1 Navigation Menu (Left Sidebar)
The persistent left sidebar containing all application navigation.

**Components:**
- **Menu Item** - Top-level navigation entry (e.g., "Dashboard", "Monthly")
- **Menu Group** - Expandable collection of related items (e.g., "Symbols", "Admin")
- **Submenu Item** - Item within an expanded Menu Group (e.g., "ANET" under Symbols)
- **Action Item** - Special submenu item that triggers creation (e.g., "+ New Symbol")

**Menu Item Types:**
- **Primary Menu Item** - Direct link to a Page (Dashboard, Monthly, Options, Treasuries, Dividends, Metrics, Help)
- **Expandable Menu Group** - Container for Submenu Items (Symbols, Admin, Accounts)

### 1.2 Menu Structure Terminology

```
Navigation Menu
├─ Primary Menu Item → Page
├─ Expandable Menu Group
│  ├─ Action Item → Modal
│  └─ Submenu Item → Detail Page
└─ ...
```

**Example:**
```
Symbols (Expandable Menu Group)
├─ + New Symbol (Action Item) → Symbol Creation Modal
├─ ANET (Submenu Item) → Symbol Detail Page
└─ BULL (Submenu Item) → Symbol Detail Page
```

---

## 2. Page Structure

### 2.1 Page Types

**Page** - The primary content area that loads when a navigation item is selected.

**Three Page Categories:**

1. **Analytics Page** - Read-only views focused on data visualization and reporting
   - Purpose: Display portfolio performance, trends, and metrics
   - Components: Charts, summary cards, read-only tables
   - Examples: Dashboard, Monthly, Metrics

2. **Management Page** - Interactive views for CRUD operations on entities
   - Purpose: Create, read, update, delete data
   - Components: Data tables, forms, action buttons, modals
   - Examples: Options, Treasuries, Dividends

3. **Utility Page** - Tools and configuration interfaces
   - Purpose: Application configuration, data import/export, system admin
   - Components: Forms, file uploads, action buttons, settings panels
   - Examples: Import, Database, Polygon, Settings, Help

4. **Detail Page** - Focused view of a single entity instance
   - Purpose: Show comprehensive information about one entity
   - Components: Entity header, related data sections, action buttons
   - Examples: Symbol Detail Page, Account Detail Page

---

### 2.2 Page Components

Every **Page** is composed of:

#### A. Page Header
- **Page Title** - Main heading (e.g., "Options", "Monthly View")
- **Page Subtitle** (optional) - Descriptive text below title
- **Page Actions** (optional) - Primary action buttons in header (e.g., "+ Add Option")

#### B. Entity Summary Panel (Detail Pages only)
- Prominent header showing key entity metrics
- Entity name/identifier
- Grid of metric cards
- Entity-level actions (Edit, Delete)

#### C. Content Sections
- **Section** - Logical grouping of related content within a page
- **Section Title** - Heading for the section
- **Section Actions** (optional) - Action buttons in section header (e.g., "+ Add")
- **Section Content** - The actual content (charts, tables, forms, etc.)

```
Page
├─ Page Header
│  ├─ Page Title
│  ├─ Page Subtitle (optional)
│  └─ Page Actions (optional)
├─ Entity Summary Panel (Detail Pages only)
│  ├─ Entity Title
│  ├─ Metric Cards
│  └─ Entity Actions
└─ Content Sections
   ├─ Section 1
   │  ├─ Section Title
   │  ├─ Section Actions (optional)
   │  └─ Section Content
   └─ Section 2
      ├─ Section Title
      ├─ Section Actions (optional)
      └─ Section Content
```

---

## 3. Content Components

### 3.1 Data Display Components

#### Charts
**Chart** - Visual data representation using Chart.js

**Chart Types:**
- **Pie Chart** - Circular chart showing proportional data
- **Bar Chart** - Vertical bars comparing values
- **Stacked Bar Chart** - Bars with multiple stacked segments
- **Line Chart** - Trend line over time
- **Scatter Plot** - Individual data points on X/Y axis
- **Bubble Chart** - Scatter plot with size dimension

**Chart Structure:**
- **Chart Title** - Label above chart
- **Chart Legend** - Key to chart elements
- **Chart Axes** - X and Y axis labels and scales
- **Data Labels** - Values displayed on chart elements

#### Tables
**Table** - Tabular data display

**Table Types:**
- **Data Table** - Read-only information display
- **Interactive Table** - Includes row actions (edit, delete)
- **Sortable Table** - Column headers clickable for sorting (indicated by ↕ icon)
- **Expandable Table** - Rows expand to show detail

**Table Structure:**
- **Table Header** - Column headings (may include sort indicators)
- **Table Row** - Single record
- **Table Cell** - Individual data value
- **Row Actions** - Buttons/links for row operations (edit, delete icons)
- **Table Footer** (optional) - Summary row with totals
- **Empty State** - Message displayed when table has no data (e.g., "No stock positions recorded for ANET")

#### Cards
**Card** - Self-contained content container with border/background

**Card Types:**
- **Summary Card** - Key metric display (number + label)
- **Info Card** - Informational content block
- **Action Card** - Card with clickable action

**Card Structure:**
- **Card Header** - Title and optional icon
- **Card Body** - Main content
- **Card Footer** (optional) - Secondary info or actions

#### Lists
**List** - Vertical arrangement of items

**List Types:**
- **Feature List** - Bulleted list of features/capabilities
- **Data List** - List of data items
- **Action List** - List items with clickable actions

#### Badges and Indicators
**Badge** - Small visual indicator showing status or type

**Badge Types:**
- **Type Badge** - Indicates entity type (e.g., "P" for Put, "C" for Call)
- **Status Badge** - Shows entity state (Open, Closed, etc.)
- **Count Badge** - Displays numerical count

**Badge Styles:**
- Colored background
- Compact size
- Positioned inline or in table cells

---

### 3.2 Interactive Components

#### Modals
**Modal** - Overlay dialog that appears on top of page content

**Modal Types:**
- **Form Modal** - Contains data entry form
- **Confirmation Modal** - Yes/No confirmation dialog
- **Info Modal** - Display-only information

**Modal Structure:**
- **Modal Header** - Title and close button
- **Modal Body** - Main content (form, message, etc.)
- **Modal Footer** - Action buttons (Submit, Cancel)

#### Forms
**Form** - Data entry interface

**Form Components:**
- **Form Field** - Single input element with label
- **Field Label** - Text describing the field
- **Input Control** - Text box, dropdown, date picker, etc.
- **Field Validation** - Rules and error messages
- **Field Help Text** (optional) - Guidance below input
- **Form Actions** - Submit and Cancel buttons

**Input Control Types:**
- **Text Input** - Single-line text entry
- **Number Input** - Numeric value entry
- **Date Picker** - Calendar date selection
- **Dropdown** - Select from predefined options
- **Autocomplete** - Text input with suggestions
- **Radio Button Group** - Select one from options
- **Checkbox** - Boolean on/off toggle
- **Text Area** - Multi-line text entry

#### Buttons
**Button** - Clickable action trigger

**Button Types:**
- **Primary Button** - Main action (e.g., "Save", "Submit")
- **Secondary Button** - Alternative action (e.g., "Cancel")
- **Icon Button** - Icon-only button
- **Action Button** - Row or card action (e.g., "Edit", "Delete")

#### Accordions
**Accordion** - Expandable/collapsible content section

**Accordion Structure:**
- **Accordion Header** - Clickable title bar
- **Accordion Content** - Hidden/shown content area
- **Accordion Toggle** - Expand/collapse icon

---

## 4. Interaction Patterns

### 4.1 Data Entry Flows

#### Create Flow
```
1. User clicks Action Item or Page Action button
2. Modal opens with Form
3. User fills Form Fields
4. User clicks Submit button
5. Validation executes
6. If valid: Entity created, Modal closes, Page refreshes
7. If invalid: Error messages displayed
```

#### Edit Flow
```
1. User clicks Edit action on Table Row or Card
2. Modal opens with Form pre-populated
3. User modifies Form Fields
4. User clicks Save button
5. Validation executes
6. If valid: Entity updated, Modal closes, Page refreshes
7. If invalid: Error messages displayed
```

#### Delete Flow
```
1. User clicks Delete action on Table Row or Card
2. Confirmation Modal opens
3. User confirms or cancels
4. If confirmed: Entity deleted, Page refreshes
```

### 4.2 Navigation Flows

#### Page Navigation
```
User clicks Menu Item → Page loads in main content area
```

#### Detail Navigation
```
User clicks Submenu Item → Detail Page loads
User clicks link in Table → Detail Page loads
```

#### Modal Navigation
```
User clicks Action Item → Modal opens
User completes or cancels → Modal closes
```

---

## 5. Data Entity Terminology

### 5.1 Core Entities (Nouns)

- **Account** - Trading account container
- **Symbol** - Stock ticker being tracked
- **Option** - Option contract (put/call)
- **Stock Position** - Equity holdings (also "Long Position")
- **Trade** - Paired open/close transactions
- **Transaction** - Atomic financial event
- **Wheel** - Campaign of related trades
- **Treasury** - Treasury security
- **Dividend** - Dividend payment
- **Commission** - Trading fee
- **Premium** - Option price

### 5.2 Entity States

- **Open** - Active, not yet closed
- **Closed** - Completed, no longer active
- **Partially Closed** - Some but not all contracts/shares closed
- **Assigned** - Option assigned (became stock)
- **Expired** - Option expired worthless
- **Active** - In use (account, wheel)
- **Archived** - Retained but inactive

### 5.3 Financial Terms

- **Strike Price** - Option exercise price
- **Expiration Date** - Option expiration
- **DTE** (Days To Expiration) - Days until option expires
- **Cost Basis** - Total purchase cost
- **P&L** (Profit & Loss) - Net profit or loss
- **AROI** (Annualized Return on Investment) - Yearly return percentage
- **Cash Flow** - Money in/out
- **Exposure** - Capital at risk

---

## 6. Detail Page Components

### 6.1 Entity Summary Panel
A prominent header panel showing key metrics and quick actions for an entity.

**Structure:**
- **Entity Title** - Name/identifier (e.g., "ANET - ANET Inc")
- **Metric Cards** - Grid of key values (Price, Yield, P/E, Total Profits, etc.)
- **Entity Actions** - Primary buttons (Edit, Delete)

**Example:** Symbol Detail Page top panel showing symbol name, price, dividend, and multiple financial metrics

### 6.2 Section Actions
Action buttons positioned in section headers (not page headers).

**Location:** Right side of Section Title
**Common Actions:** "+ Add [Entity Type]" button

**Example:** "+ Add" button in "Options" section header

---

## 7. Consistent Naming Conventions

### 7.1 File Naming

- **HTML Templates** - lowercase with underscores: `symbol_detail.html`, `account_detail.html`
- **Go Handlers** - camelCase: `symbolHandler`, `accountDetailHandler`
- **CSS Classes** - kebab-case: `.section-title`, `.form-modal`, `.data-table`
- **JavaScript Functions** - camelCase: `openModal()`, `saveOption()`, `calculateAROI()`

### 7.2 Documentation Structure

When documenting a Page, use this template:

```markdown
# [Page Name]

**Page Type:** Analytics / Management / Utility / Detail

**Navigation:** [Path to reach page]

**Purpose:** [One sentence description]

## Page Structure

### Page Header
- Page Title: [Title text]
- Page Subtitle: [Subtitle if present]
- Page Actions: [List of buttons]

### Entity Summary Panel (Detail Pages only)
- Entity Title: [Dynamic entity name]
- Metric Cards: [List of metrics displayed]
- Entity Actions: [Edit, Delete, other buttons]

### Content Sections

#### Section 1: [Section Title]
**Purpose:** [What this section shows/does]

**Section Actions:** [List action buttons in section header]

**Components:**
- [List components: charts, tables, forms, etc.]

**Data Requirements:**
- [What data needs to be fetched from backend]

**Business Rules:**
- [Calculations, validations, constraints]

**Interactions:**
- [What user can do]

**Empty State:** [Message shown when no data]

---

[Repeat for each section]
```

---

## 8. Visual Hierarchy

### 8.1 Text Styles

- **Page Title** - Largest, primary heading
- **Section Title** - Secondary heading
- **Subsection Title** - Tertiary heading
- **Body Text** - Standard paragraph text
- **Label Text** - Form field labels
- **Help Text** - Small, muted guidance text
- **Error Text** - Red validation message

### 8.2 Spacing

- **Page** - Full main content area
- **Section** - Separated by margin/padding
- **Component** - Contained within section
- **Element** - Smallest unit (button, field, cell)

---

## 9. Validation Terminology

### 9.1 Validation Types

- **Required Field Validation** - Field cannot be empty
- **Format Validation** - Input must match pattern (email, date, number)
- **Range Validation** - Number within min/max bounds
- **Business Rule Validation** - Domain-specific constraint
- **Cross-Field Validation** - Multiple fields must satisfy relationship

### 9.2 Error Display

- **Inline Error** - Message below/beside field
- **Summary Error** - List of all errors at top of form
- **Toast Notification** - Temporary popup message

---

## 10. Complete Example: Symbol Detail Page

**Page Type:** Detail Page

**Navigation Path:** Main Menu → Symbols → [Symbol Name]

### Page Structure

**Entity Summary Panel**
- Entity Title: "ANET - ANET Inc"
- Metric Cards: Price, Div, Yield, P/E, Options Gains, Cap Gains, Dividends, Total Profits, Cash on Cash, Long Value, Put Exposed
- Entity Actions: "Edit" button, "Delete" button

**Content Sections**

**Section 1: Options**
- Component Type: Interactive Sortable Table
- Section Actions: "+ Add" button
- Columns: Call/Put, Date Sold, Closed Date, Strike, OTM, Expiration, Remaining, DTE, DTC, Contracts, Premium, Exit Price, Commission, Total, % of Profit, % of Time, Multiplier
- Row Actions: Edit (pencil icon), Delete (trash icon)
- Type Badge: "P" for Put, "C" for Call
- Empty State: Shows data when present, no empty state in example

**Section 2: Stock Positions**  
- Component Type: Interactive Sortable Table
- Section Actions: "+ Add" button
- Columns: Purchased, Closed Date, Yield, Shares, Buy Price, Exit Price, Profit/Loss, ROI, Amount, Total Invested
- Empty State: "No stock positions recorded for ANET"

**Section 3: Dividends**
- Component Type: Interactive Sortable Table  
- Section Actions: "+ Add" button
- Columns: Received, Amount
- Row Actions: Delete (trash icon)
- Empty State: "No dividends recorded for ANET"

**Section 4: Monthly Results**
- Component Type: Data Table (Read-only)
- Columns: Month, Puts, Calls, Puts Total, Calls Total, Total
- No Section Actions (summary/analytics only)

**Interactive Elements**
- Click column headers → Sort table
- Click Edit icon → Opens Edit Modal (pre-populated)
- Click Delete icon → Opens Confirmation Modal
- Click "+ Add" → Opens Create Modal for that entity type

---

This taxonomy establishes consistent vocabulary for all Wheeler documentation going forward.