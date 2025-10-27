# WHEELER - ENHANCE COMMISSION HANDLING
## Detailed Migration Plan - Simplified Approach

**Document Version:** 1.0
**Date:** October 26, 2024
**Status:** Ready for Implementation

---

## TABLE OF CONTENTS

1. [Executive Summary](#executive-summary)
2. [Phase 1: Database & Settings Infrastructure](#phase-1-database--settings-infrastructure)
3. [Phase 2: Detailed Requirements](#phase-2-detailed-requirements)
4. [Phase 3: Implementation Plan](#phase-3-implementation-plan)
5. [Phase 4: Comprehensive Testing Plan](#phase-4-comprehensive-testing-plan)
6. [Phase 5: Migration & Rollout Plan](#phase-5-migration--rollout-plan)
7. [Phase 6: Documentation Updates](#phase-6-documentation-updates)
8. [Phase 7: Known Limitations & Future Enhancements](#phase-7-known-limitations--future-enhancements)
9. [Summary & Recommendations](#summary--recommendations)

---

## EXECUTIVE SUMMARY

This plan addresses four critical improvements to the Wheeler options tracking system:

1. **Make commission configurable** - Move from hardcoded $0.65 to database-driven settings
2. **Fix commission calculation bug** - Only charge closing commission when position is bought to close (not expired/assigned)
3. **Fix maxProfit calculation** - Correctly calculate opening commission for percentage calculations
4. **Protect data integrity** - Prevent editing contracts when position has exit date

**Key Principle:** Keep it simple - trust the user to enter correct values, protect closed positions from commission changes.

### Key Design Decisions

- **Storage:** Database settings table (not JSON file) for consistency with POLYGON_API_KEY
- **User Control:** Commission editable per-trade with sensible defaults
- **Data Protection:** Closed positions cannot have commission edited
- **Integrity:** Contracts field locked when exit date present (prevents partial closes)

---

## PHASE 1: DATABASE & SETTINGS INFRASTRUCTURE

### 1.1 Analysis of Current System

**Existing Settings Infrastructure:**
- ✅ Settings table already exists in schema
- ✅ SettingService already implements CRUD operations (`internal/models/setting.go`)
- ✅ UI already exists for settings (`internal/web/templates/settings.html`)
- ✅ Helper methods: `GetValue()`, `GetValueWithDefault()`, `SetValue()`, `Upsert()`

**Current Commission Implementation:**
- ❌ Hardcoded constant: `OptionCommissionPerContract = 0.65` (`internal/models/option.go:11`)
- ❌ Used in 3 places: `Create()`, `Close()`, `CloseByID()`
- ❌ No UI to modify value

### 1.2 Design Decision: Database Settings

**Store commission in `settings` table as `OPTION_COMMISSION_PER_CONTRACT`**

**Why Database over JSON:**

| Factor | Database | JSON File |
|--------|----------|-----------|
| Consistency | ✅ Matches existing POLYGON_API_KEY pattern | ❌ New pattern to maintain |
| Multi-database | ✅ Per-database settings | ❌ Global across all databases |
| UI Integration | ✅ Already implemented | ❌ Need new file I/O handlers |
| Backup | ✅ Included in DB backups | ❌ Separate file to track |
| Migration | ✅ Automatic via schema | ❌ Need file creation logic |

---

## PHASE 2: DETAILED REQUIREMENTS

### 2.1 Functional Requirements

#### FR-1: Configurable Commission Setting

- Setting name: `OPTION_COMMISSION_PER_CONTRACT`
- Default value: `0.65` (maintain backward compatibility)
- Data type: Decimal (stored as string, parsed as float64)
- Description: "Commission charged per options contract (opening and closing)"

#### FR-2: Commission Calculation Rules

```
Opening Position:
  User enters commission per contract (default from settings)
  Store: commission_per_contract × number_of_contracts

Closing Position with exitPrice = 0 (expired):
  Commission field unchanged (opening commission only)

Closing Position with exitPrice > 0 (bought to close):
  Commission = commission × 2 (double it)
```

#### FR-3: MaxProfit Calculation (Corrected)

```
MaxProfit is CONSTANT = (Premium × Contracts × 100) - Opening Commission

To get opening commission from stored commission field:
  - If position is open: commission field = opening commission
  - If position expired (exitPrice = 0): commission field = opening commission
  - If position bought back (exitPrice > 0): opening commission = commission / 2
```

**Key Insight:** MaxProfit never changes - it represents the best possible outcome (option expires worthless).

#### FR-4: Commission Field Editability

```
NEW position:          Commission editable ✓
OPEN position:         Commission editable ✓
CLOSING position:      Commission editable ✓ (as part of close action)
CLOSED position:       Commission READ-ONLY ✗
```

**Rationale:** Protects historical data integrity while allowing corrections before finalization.

#### FR-5: Contracts Field Protection

```
NEW position:              Contracts editable ✓
OPEN position (no exit):   Contracts editable ✓
Position with exit date:   Contracts READ-ONLY ✗
```

**Reason:** Application doesn't support partial closing of positions.

#### FR-6: Settings UI Enhancement

- Add "Trading Settings" card to `internal/web/templates/settings.html`
- Display current commission value
- Allow decimal input (e.g., 0.65, 1.00, 0.50)
- Validate: must be >= 0, <= 50.00 (sanity check)
- Show preview: "Round-trip cost for 10 contracts: $13.00"

### 2.2 Non-Functional Requirements

**NFR-1: Backward Compatibility**
- Existing databases without setting → default to $0.65
- Existing commission values in options table → remain unchanged

**NFR-2: Performance**
- Commission lookup cached in OptionService (not queried per operation)
- Settings loaded once at service initialization

**NFR-3: Data Integrity**
- Historical commissions remain unchanged (cannot edit closed positions)
- Contracts cannot be changed once exit date entered (prevents partial closes)

---

## PHASE 3: IMPLEMENTATION PLAN

### 3.1 File Changes Overview

| File | Change Type | Description |
|------|-------------|-------------|
| `internal/models/option.go` | Major | Remove constant, add commission parameter to methods |
| `internal/models/symbol.go` | Minor | Fix CalculatePercentOfProfit() logic |
| `internal/models/setting.go` | Minor | Add helper method `GetFloatValue()` |
| `internal/web/settings_handlers.go` | Minor | Add commission setting to settings page |
| `internal/web/templates/settings.html` | Major | Add Trading Settings card |
| `internal/web/options_handlers.go` | Minor | Update create/update to handle commission |
| Option edit UI (symbol page or modal) | Major | Add commission field, protect contracts field |
| `test/option_test.go` | New | Unit tests for commission logic |
| `test/setting_test.go` | New | Unit tests for settings retrieval |

### 3.2 Step-by-Step Implementation

---

#### STEP 1: Enhance SettingService with Type Helpers

**File:** `internal/models/setting.go`

**Add new methods at the end of the file:**

```go
// GetFloatValue returns the setting value as float64, or default if not found/invalid
func (s *SettingService) GetFloatValue(name string, defaultValue float64) float64 {
	value := s.GetValue(name)
	if value == "" {
		return defaultValue
	}

	// Parse as float
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}

	return floatValue
}

// GetFloatValueWithValidation returns float64 with min/max validation
func (s *SettingService) GetFloatValueWithValidation(name string, defaultValue, min, max float64) float64 {
	value := s.GetFloatValue(name, defaultValue)

	// Validate range
	if value < min {
		return min
	}
	if value > max {
		return max
	}

	return value
}
```

**Import needed:** Add `"strconv"` to imports at top of file

**Testing:**
- Test with valid float string: "0.65" → 0.65
- Test with invalid string: "abc" → default value
- Test with empty string: "" → default value
- Test with out-of-range: 100 → clamped to max

---

#### STEP 2: Refactor OptionService to Use Dynamic Commission

**File:** `internal/models/option.go`

**A. Remove hardcoded constant (line 11):**

```go
// REMOVE THIS LINE:
// const OptionCommissionPerContract = 0.65

// REPLACE WITH:
const DefaultCommissionPerContract = 0.65 // Fallback if setting not configured
```

**B. Add settingService to OptionService struct (around line 13):**

```go
type OptionService struct {
	db             *sql.DB
	settingService *SettingService
}
```

**C. Update NewOptionService constructor (around line 17):**

```go
func NewOptionService(db *sql.DB, settingService *SettingService) *OptionService {
	return &OptionService{
		db:             db,
		settingService: settingService,
	}
}
```

**D. Add helper method to get commission rate:**

```go
// GetCommissionPerContract retrieves the current commission setting
func (s *OptionService) GetCommissionPerContract() float64 {
	if s.settingService == nil {
		return DefaultCommissionPerContract
	}

	// Get from settings with validation (min: 0, max: 50)
	return s.settingService.GetFloatValueWithValidation(
		"OPTION_COMMISSION_PER_CONTRACT",
		DefaultCommissionPerContract,
		0.0,   // min
		50.0,  // max
	)
}
```

**E. Update Create() method (line 21-25):**

The Create() method will now receive commission as a parameter from the UI:

```go
func (s *OptionService) Create(symbol, optionType string, opened time.Time, strike float64, expiration time.Time, premium float64, contracts int, commissionPerContract float64) (*Option, error) {
	// Calculate total opening commission from per-contract rate
	openingCommission := commissionPerContract * float64(contracts)
	return s.CreateWithCommission(symbol, optionType, opened, strike, expiration, premium, contracts, openingCommission)
}
```

**NOTE:** This changes the signature - all callers must be updated to pass commissionPerContract.

**F. Update Close() method (line 133-156) - FIX BUG:**

```go
func (s *OptionService) Close(symbol, optionType string, opened time.Time, strike float64, expiration time.Time, premium float64, contracts int, closed time.Time, exitPrice float64) error {
	// Calculate closing commission: only if position was bought to close (exitPrice > 0)
	// For expired positions (exitPrice = 0), no closing commission is added
	closingCommission := 0.0
	if exitPrice > 0 {
		// Get current commission setting
		commissionPerContract := s.GetCommissionPerContract()
		closingCommission = commissionPerContract * float64(contracts)
	}

	query := `UPDATE options
			  SET closed = ?, exit_price = ?, commission = commission + ?, updated_at = CURRENT_TIMESTAMP
			  WHERE symbol = ? AND type = ? AND opened = ? AND strike = ? AND expiration = ? AND premium = ? AND contracts = ?`

	result, err := s.db.Exec(query, closed, exitPrice, closingCommission, symbol, optionType, opened, strike, expiration, premium, contracts)
	if err != nil {
		return fmt.Errorf("failed to close option: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("option not found")
	}

	return nil
}
```

**G. Update CloseByID() method (line 246-275) - FIX BUG:**

```go
func (s *OptionService) CloseByID(id int, closed time.Time, exitPrice float64) error {
	// First get the option to find out the number of contracts
	option, err := s.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get option for commission calculation: %w", err)
	}

	// Calculate closing commission: only if position was bought to close (exitPrice > 0)
	// For expired positions (exitPrice = 0), no closing commission is added
	closingCommission := 0.0
	if exitPrice > 0 {
		// Get current commission setting
		commissionPerContract := s.GetCommissionPerContract()
		closingCommission = commissionPerContract * float64(option.Contracts)
	}

	query := `UPDATE options
			  SET closed = ?, exit_price = ?, commission = commission + ?, updated_at = CURRENT_TIMESTAMP
			  WHERE id = ?`

	result, err := s.db.Exec(query, closed, exitPrice, closingCommission, id)
	if err != nil {
		return fmt.Errorf("failed to close option: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("option not found")
	}

	return nil
}
```

---

#### STEP 3: Fix MaxProfit Calculation

**File:** `internal/models/symbol.go`

**Update CalculatePercentOfProfit() method (around line 146-154):**

```go
func (o *Option) CalculatePercentOfProfit() float64 {
	if o.Premium == 0 {
		return 0
	}

	// MaxProfit is CONSTANT = (Premium × Contracts × 100) - Opening Commission
	// Extract opening commission from the commission field:
	// - Open positions: commission = opening commission
	// - Expired positions (exitPrice = 0): commission = opening commission
	// - Bought-back positions (exitPrice > 0): commission = opening + closing, so divide by 2

	openingCommission := o.Commission
	if o.Closed != nil && o.ExitPrice != nil && *o.ExitPrice > 0 {
		// Position was bought to close - commission includes both opening and closing
		// Assume equal rates, so opening = total / 2
		openingCommission = o.Commission / 2.0
	}

	maxProfit := (o.Premium * float64(o.Contracts) * 100) - openingCommission
	actualProfit := o.CalculateTotalProfit()

	if maxProfit <= 0 {
		return 0
	}

	return (actualProfit / maxProfit) * 100
}
```

**Add comment explaining the assumption:**

```go
// Note: For positions bought to close, this assumes opening and closing
// commissions are equal. If commission rates changed between opening and
// closing, the calculation uses an approximation.
```

---

#### STEP 4: Update Server Initialization

**File:** `internal/web/server.go`

**Find where OptionService is initialized and update to pass settingService:**

Look for something like:
```go
optionService := models.NewOptionService(db.DB)
```

**Replace with:**
```go
optionService := models.NewOptionService(db.DB, settingService)
```

**This is a breaking change** - ensure settingService is initialized before optionService.

---

#### STEP 5: Update Option Handlers to Handle Commission

**File:** `internal/web/options_handlers.go`

**A. Update createOption handler (around line 230-294):**

```go
func (s *Server) createOption(w http.ResponseWriter, r *http.Request) {
	log.Printf("[CREATE OPTION] Starting POST request")

	var req OptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[CREATE OPTION] ERROR: Invalid JSON payload: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Symbol == "" || req.Type == "" || req.Opened == "" || req.Expiration == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	if req.Type != "Put" && req.Type != "Call" {
		http.Error(w, "Type must be 'Put' or 'Call'", http.StatusBadRequest)
		return
	}

	// Parse dates
	opened, err := time.Parse("2006-01-02", req.Opened)
	if err != nil {
		http.Error(w, "Invalid opened date format", http.StatusBadRequest)
		return
	}

	expiration, err := time.Parse("2006-01-02", req.Expiration)
	if err != nil {
		http.Error(w, "Invalid expiration date format", http.StatusBadRequest)
		return
	}

	// Get commission per contract
	// If provided in request, use it; otherwise get from settings
	commissionPerContract := req.CommissionPerContract
	if commissionPerContract <= 0 {
		commissionPerContract = s.optionService.GetCommissionPerContract()
	}

	// Create the option with commission per contract
	option, err := s.optionService.Create(req.Symbol, req.Type, opened, req.Strike, expiration, req.Premium, req.Contracts, commissionPerContract)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create option: %v", err), http.StatusInternalServerError)
		return
	}

	// If closed date and exit price are provided, close the option immediately
	if req.Closed != nil && *req.Closed != "" {
		closed, err := time.Parse("2006-01-02", *req.Closed)
		if err != nil {
			http.Error(w, "Invalid closed date format", http.StatusBadRequest)
			return
		}

		exitPrice := 0.0
		if req.ExitPrice != nil {
			exitPrice = *req.ExitPrice
		}

		err = s.optionService.CloseByID(option.ID, closed, exitPrice)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to close option: %v", err), http.StatusInternalServerError)
			return
		}

		// Retrieve updated option with commission
		option, err = s.optionService.GetByID(option.ID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to retrieve closed option: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(option)
}
```

**B. Update updateOption handler (around line 297-354):**

```go
func (s *Server) updateOption(w http.ResponseWriter, r *http.Request) {
	var req OptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate ID is provided for update
	if req.ID == nil {
		http.Error(w, "Option ID is required for update", http.StatusBadRequest)
		return
	}

	// Get existing option to check if it's closed
	existingOption, err := s.optionService.GetByID(*req.ID)
	if err != nil {
		http.Error(w, "Option not found", http.StatusNotFound)
		return
	}

	// Validate required fields
	if req.Symbol == "" || req.Type == "" || req.Opened == "" || req.Expiration == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	if req.Type != "Put" && req.Type != "Call" {
		http.Error(w, "Type must be 'Put' or 'Call'", http.StatusBadRequest)
		return
	}

	// Parse dates
	opened, err := time.Parse("2006-01-02", req.Opened)
	if err != nil {
		http.Error(w, "Invalid opened date format", http.StatusBadRequest)
		return
	}

	expiration, err := time.Parse("2006-01-02", req.Expiration)
	if err != nil {
		http.Error(w, "Invalid expiration date format", http.StatusBadRequest)
		return
	}

	// Parse closed date if provided
	var closed *time.Time
	if req.Closed != nil && *req.Closed != "" {
		closedDate, err := time.Parse("2006-01-02", *req.Closed)
		if err != nil {
			http.Error(w, "Invalid closed date format", http.StatusBadRequest)
			return
		}
		closed = &closedDate
	}

	// Calculate commission based on whether position is already closed
	var totalCommission float64

	if existingOption.Closed != nil {
		// Position is already closed - commission cannot be edited
		// Use existing commission value
		totalCommission = existingOption.Commission
	} else {
		// Position is open or being closed - commission can be edited
		commissionPerContract := req.CommissionPerContract
		if commissionPerContract <= 0 {
			// If not provided, calculate from existing commission
			commissionPerContract = existingOption.Commission / float64(existingOption.Contracts)
		}

		// Calculate total commission
		totalCommission = commissionPerContract * float64(req.Contracts)

		// If closing the position with exitPrice > 0, double the commission
		if closed != nil && req.ExitPrice != nil && *req.ExitPrice > 0 && existingOption.Closed == nil {
			totalCommission = totalCommission * 2.0
		}
	}

	// Update the option
	option, err := s.optionService.UpdateByID(*req.ID, req.Symbol, req.Type, opened, req.Strike, expiration, req.Premium, req.Contracts, totalCommission, closed, req.ExitPrice)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update option: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(option)
}
```

**C. Update OptionRequest struct in `types.go`:**

```go
type OptionRequest struct {
	ID                    *int     `json:"id,omitempty"`
	Symbol                string   `json:"symbol"`
	Type                  string   `json:"type"`
	Strike                float64  `json:"strike"`
	Expiration            string   `json:"expiration"`
	Premium               float64  `json:"premium"`
	Contracts             int      `json:"contracts"`
	Opened                string   `json:"opened"`
	Closed                *string  `json:"closed,omitempty"`
	ExitPrice             *float64 `json:"exit_price,omitempty"`
	Commission            float64  `json:"commission"`
	CommissionPerContract float64  `json:"commission_per_contract,omitempty"` // NEW FIELD
}
```

---

#### STEP 6: Update Settings UI

**File:** `internal/web/templates/settings.html`

**Add new card after Polygon.io card (around line 98, before the closing content-section div):**

```html
<!-- Trading Settings Card -->
<div class="settings-form-container">
    <div class="settings-card">
        <div class="settings-card-header">
            <i class="fas fa-coins"></i>
            <h3>Trading Configuration</h3>
        </div>
        <div class="settings-card-body">
            <form id="commissionForm">
                <div class="form-group">
                    <label for="commissionInput" class="form-label">Commission Per Contract ($)</label>
                    <input type="number" id="commissionInput" class="form-input"
                           placeholder="0.65" step="0.01" min="0" max="50"
                           value="{{if .CommissionPerContract}}{{printf "%.2f" .CommissionPerContract}}{{else}}0.65{{end}}">
                    <div class="form-help">
                        <i class="fas fa-info-circle"></i>
                        Commission charged per options contract for opening and closing trades
                    </div>
                </div>

                <!-- Preview Calculation -->
                <div class="commission-preview" id="commissionPreview">
                    <div class="preview-item">
                        <span class="preview-label">Opening (10 contracts):</span>
                        <span class="preview-value" id="openingCost">$6.50</span>
                    </div>
                    <div class="preview-item">
                        <span class="preview-label">Closing (10 contracts):</span>
                        <span class="preview-value" id="closingCost">$6.50</span>
                    </div>
                    <div class="preview-item total">
                        <span class="preview-label">Round-trip Total:</span>
                        <span class="preview-value" id="roundTripCost">$13.00</span>
                    </div>
                </div>

                <div class="form-group">
                    <button type="submit" class="btn btn-primary" id="saveCommissionBtn">
                        <i class="fas fa-save"></i>
                        Save Commission Rate
                    </button>
                </div>
            </form>
        </div>
    </div>
</div>
```

**Add CSS styles in the existing `<style>` section (around line 267):**

```css
.commission-preview {
    background: #1e1e1e;
    border: 1px solid #404040;
    border-radius: 6px;
    padding: 15px;
    margin: 15px 0;
}

.preview-item {
    display: flex;
    justify-content: space-between;
    padding: 8px 0;
    color: #b0b0b0;
}

.preview-item.total {
    border-top: 1px solid #404040;
    margin-top: 10px;
    padding-top: 10px;
    font-weight: 600;
    color: #e0e0e0;
}

.preview-label {
    font-size: 13px;
}

.preview-value {
    font-size: 13px;
    font-weight: 500;
    color: #27ae60;
}
```

**Add JavaScript in the existing `<script>` section (before the closing script tag around line 464):**

```javascript
// Update commission preview
function updateCommissionPreview() {
    const commission = parseFloat(document.getElementById('commissionInput').value) || 0;
    const contracts = 10;

    const openingCost = commission * contracts;
    const closingCost = commission * contracts;
    const roundTripCost = openingCost + closingCost;

    document.getElementById('openingCost').textContent = `$${openingCost.toFixed(2)}`;
    document.getElementById('closingCost').textContent = `$${closingCost.toFixed(2)}`;
    document.getElementById('roundTripCost').textContent = `$${roundTripCost.toFixed(2)}`;
}

// Live update preview as user types
document.getElementById('commissionInput').addEventListener('input', updateCommissionPreview);

// Form submission
document.getElementById('commissionForm').addEventListener('submit', function(e) {
    e.preventDefault();

    const commission = document.getElementById('commissionInput').value.trim();
    const saveBtn = document.getElementById('saveCommissionBtn');

    // Validate
    const commissionFloat = parseFloat(commission);
    if (isNaN(commissionFloat) || commissionFloat < 0 || commissionFloat > 50) {
        showNotification('Commission must be between $0.00 and $50.00', 'error');
        return;
    }

    // Disable button and show loading state
    saveBtn.disabled = true;
    const originalText = saveBtn.innerHTML;
    saveBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Saving...';

    // Update commission via API
    fetch('/api/settings/OPTION_COMMISSION_PER_CONTRACT', {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            value: commission,
            description: 'Commission charged per options contract (opening and closing)'
        })
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Failed to save commission setting');
        }
        return response.json();
    })
    .then(data => {
        showNotification('Commission rate saved successfully!', 'success');
    })
    .catch(error => {
        console.error('Error saving commission:', error);
        showNotification('Error saving commission: ' + error.message, 'error');
    })
    .finally(() => {
        // Restore button state
        saveBtn.disabled = false;
        saveBtn.innerHTML = originalText;
    });
});

// Initialize preview on page load when commission form exists
if (document.getElementById('commissionForm')) {
    updateCommissionPreview();
}
```

---

#### STEP 7: Update Settings Handler

**File:** `internal/web/settings_handlers.go`

**Update the settings page handler to include commission value:**

Find the `settingsHandler` function and update the data struct:

```go
func (s *Server) settingsHandler(w http.ResponseWriter, r *http.Request) {
	symbols, err := s.symbolService.GetDistinctSymbols()
	if err != nil {
		log.Printf("[SETTINGS] WARNING: Failed to get symbols for navigation: %v", err)
		symbols = []string{}
	}

	// Get Polygon API key
	apiKey := s.settingService.GetValue("POLYGON_API_KEY")

	// Get commission per contract (use GetFloatValue for proper type handling)
	commissionPerContract := s.settingService.GetFloatValue("OPTION_COMMISSION_PER_CONTRACT", 0.65)

	data := struct {
		Symbols               []string
		AllSymbols            []string
		CurrentDB             string
		ActivePage            string
		ApiKey                string
		CommissionPerContract float64
	}{
		Symbols:               symbols,
		AllSymbols:            symbols,
		CurrentDB:             s.getCurrentDatabaseName(),
		ActivePage:            "settings",
		ApiKey:                apiKey,
		CommissionPerContract: commissionPerContract,
	}

	s.renderTemplate(w, "settings.html", data)
}
```

---

#### STEP 8: Update Option Edit UI

**This is the most complex part - needs to be done wherever options are edited (symbol page modal, options page, etc.)**

**Key Requirements:**

1. **Add Commission Per Contract field**
   - Label: "Commission per Contract"
   - Type: number, step="0.01", min="0", max="50"
   - Default on create: Load from settings
   - Default on edit: Calculate from stored commission

2. **Calculate displayed commission rate for editing:**

```javascript
// When loading option for edit
function loadOptionForEdit(option) {
    // ... existing code ...

    // Calculate commission per contract for display
    let commissionPerContract = 0.65; // default

    if (option.id) {
        // Editing existing option
        if (option.closed && option.exit_price && option.exit_price > 0) {
            // Bought to close - commission has both opening and closing
            commissionPerContract = (option.commission / 2.0) / option.contracts;
        } else {
            // Open or expired - commission is opening only
            commissionPerContract = option.commission / option.contracts;
        }
    }

    document.getElementById('commissionPerContract').value = commissionPerContract.toFixed(2);

    // Set field editability based on position status
    const isClosed = (option.closed != null);
    document.getElementById('commissionPerContract').disabled = isClosed;

    // Contracts field protection
    const hasExitDate = (option.closed != null || document.getElementById('exitPrice').value !== '');
    document.getElementById('contracts').disabled = hasExitDate;
}
```

3. **Handle contracts field protection:**

```javascript
// Add event listener to exit price field
document.getElementById('exitPrice').addEventListener('input', function() {
    const hasExitPrice = this.value !== '';
    document.getElementById('contracts').disabled = hasExitPrice;

    if (hasExitPrice) {
        // Optionally show tooltip
        document.getElementById('contracts').title = 'Cannot change contracts when exit price is set';
    }
});

// Add event listener to closed date field
document.getElementById('closedDate').addEventListener('input', function() {
    const hasClosedDate = this.value !== '';
    document.getElementById('contracts').disabled = hasClosedDate;

    if (hasClosedDate) {
        document.getElementById('contracts').title = 'Cannot change contracts when closing position';
    }
});
```

4. **Update form submission:**

```javascript
function saveOption() {
    // ... existing validation ...

    const optionData = {
        id: editingOptionId, // null for new, number for edit
        symbol: document.getElementById('symbol').value,
        type: document.getElementById('type').value,
        strike: parseFloat(document.getElementById('strike').value),
        expiration: document.getElementById('expiration').value,
        premium: parseFloat(document.getElementById('premium').value),
        contracts: parseInt(document.getElementById('contracts').value),
        opened: document.getElementById('opened').value,
        closed: document.getElementById('closedDate').value || null,
        exit_price: document.getElementById('exitPrice').value ? parseFloat(document.getElementById('exitPrice').value) : null,
        commission_per_contract: parseFloat(document.getElementById('commissionPerContract').value)
    };

    // ... rest of save logic ...
}
```

**Note:** The exact file location for option edit UI depends on your implementation. Common locations:
- `internal/web/templates/symbol.html` (symbol page modal)
- `internal/web/templates/options.html` (options page modal)
- Shared modal in `_symbol_modal.html` or similar

---

#### STEP 9: Database Schema Migration (Optional)

**File:** `internal/database/schema.sql`

**Add default commission setting at the end of the file:**

```sql
-- Insert default commission setting if not exists
INSERT OR IGNORE INTO settings (name, value, description, created_at, updated_at)
VALUES ('OPTION_COMMISSION_PER_CONTRACT', '0.65', 'Commission charged per options contract (opening and closing)', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
```

This ensures new databases start with the default value. Existing databases will use the fallback logic in code.

---

## PHASE 4: COMPREHENSIVE TESTING PLAN

### 4.1 Unit Tests

#### Test File: `test/setting_test.go` (NEW FILE)

Create a new file with comprehensive tests for the float value helpers.

**Key Test Cases:**
- Valid float string parsing
- Invalid string handling (returns default)
- Not found setting (returns default)
- Min/max validation and clamping
- Commission setting default behavior
- Custom commission setting storage and retrieval

#### Test File: `test/option_commission_test.go` (NEW FILE)

Create comprehensive tests for commission calculations.

**Key Test Cases:**
- Opening commission with default rate
- Opening commission with custom rate
- Closing commission for bought-to-close positions
- No closing commission for expired positions
- Total profit calculation for expired options
- Total profit calculation for bought-back options
- Percent of profit for expired (should be 100%)
- Percent of profit for bought-back (should be partial)

### 4.2 Integration Tests

**Test Scenarios:**

1. **End-to-End: Create → Close (bought back) → Verify Profit**
   - Create option with commission rate from settings
   - Close by buying back at lower price
   - Verify full round-trip commission charged
   - Verify profit calculation correct

2. **End-to-End: Create → Expire → Verify Profit**
   - Create option with commission rate from settings
   - Close with exitPrice = 0 (expired)
   - Verify only opening commission charged
   - Verify profit calculation correct (100% of max)

3. **Settings Change Mid-Trade**
   - Create option with commission rate 0.65
   - Change commission setting to 1.00
   - Close option
   - Verify closing uses rate 1.00 (current rate at close time)

4. **Edit Open Position Commission**
   - Create option with 0.65 rate
   - Edit and change commission to 1.00
   - Verify stored commission updated correctly

5. **Cannot Edit Closed Position Commission**
   - Create and close option
   - Attempt to edit commission
   - Verify commission field is read-only

6. **Cannot Edit Contracts With Exit Date**
   - Create open option
   - Enter exit date
   - Verify contracts field becomes read-only

### 4.3 Manual Testing Checklist

**Settings Page:**
- [ ] Settings page loads without errors
- [ ] Commission input accepts decimal values (0.65, 1.00, etc.)
- [ ] Commission preview updates in real-time as you type
- [ ] Save button works and displays success notification
- [ ] Validation rejects negative values
- [ ] Validation rejects values > $50
- [ ] Settings persist across page refreshes
- [ ] Settings persist across server restarts

**Creating Options:**
- [ ] Commission per contract field displays default from settings
- [ ] User can edit commission per contract value
- [ ] Option is created with correct total commission (rate × contracts)
- [ ] Custom commission rate is used (not default) if edited

**Editing Open Options:**
- [ ] Commission per contract field displays calculated value (commission / contracts)
- [ ] Commission field is editable
- [ ] Contracts field is editable (when no exit date)
- [ ] Changing commission updates stored value correctly
- [ ] Changing contracts updates stored commission correctly

**Closing Options - Expired:**
- [ ] Set exit price to 0
- [ ] Contracts field becomes read-only
- [ ] Commission field is still editable (can adjust before closing)
- [ ] On save, commission is NOT doubled
- [ ] Profit calculation shows 100% (or near 100%)

**Closing Options - Bought to Close:**
- [ ] Set exit price > 0
- [ ] Contracts field becomes read-only
- [ ] Commission field is still editable
- [ ] On save, closing commission is added (using current rate from settings)
- [ ] Profit calculation shows partial percentage

**Editing Closed Options:**
- [ ] Commission field is read-only
- [ ] Contracts field is read-only
- [ ] Exit price can still be edited
- [ ] Closed date can still be edited
- [ ] Display shows correct commission rate (handles bought-back vs expired)

**Profit Calculations:**
- [ ] Expired option shows correct max profit (premium - opening commission)
- [ ] Expired option shows 100% of profit achieved
- [ ] Bought-back option shows correct max profit (premium - opening commission)
- [ ] Bought-back option shows partial % of profit
- [ ] All profit displays are correct on dashboard, monthly, symbol pages

---

## PHASE 5: MIGRATION & ROLLOUT PLAN

### 5.1 Pre-Deployment Checklist

- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Code review completed
- [ ] Documentation updated (CLAUDE.md)
- [ ] Backup current database before deployment
- [ ] Test on development environment first

### 5.2 Deployment Steps

1. **Backup Production Database**
   ```bash
   cp data/wheeler.db data/backups/wheeler_pre_commission_update_$(date +%Y%m%d_%H%M%S).db
   ```

2. **Deploy Code Changes**
   ```bash
   git pull origin main
   go build .
   ```

3. **Verify Settings Table**
   ```sql
   SELECT * FROM settings WHERE name = 'OPTION_COMMISSION_PER_CONTRACT';
   ```
   If not exists, will default to 0.65 (no action needed)

4. **Start Application**
   ```bash
   ./wheeler
   ```

5. **Verify Functionality**
   - Visit `/settings`
   - Verify commission card appears with default value
   - Set commission value and save
   - Create test option
   - Verify commission calculated correctly
   - Close test option (expired)
   - Verify no closing commission added
   - Create another test option
   - Close with exit price
   - Verify closing commission added

### 5.3 Rollback Plan

If critical issues arise:

1. **Stop Application**
   ```bash
   # Kill the process
   pkill wheeler
   ```

2. **Restore Backup**
   ```bash
   cp data/backups/wheeler_pre_commission_update_YYYYMMDD_HHMMSS.db data/wheeler.db
   ```

3. **Revert Code**
   ```bash
   git log --oneline  # Find previous commit
   git checkout <previous-commit-hash>
   go build .
   ./wheeler
   ```

### 5.4 Post-Deployment Validation

- [ ] Verify existing options display correct historical commissions
- [ ] Create new option → verify commission calculated from settings
- [ ] Edit open option → verify commission editable
- [ ] Close option (expired) → verify no closing commission
- [ ] Close option (bought) → verify closing commission added
- [ ] Edit closed option → verify commission is read-only
- [ ] Change commission setting → verify new rate used for new trades
- [ ] Check all dashboards still calculate P&L correctly
- [ ] Verify monthly analysis includes correct commissions
- [ ] Verify symbol pages show correct profit percentages

---

## PHASE 6: DOCUMENTATION UPDATES

### 6.1 Update CLAUDE.md

Add section after "Commission Model" (around line 177):

```markdown
### Commission Configuration

Wheeler uses configurable commission rates for options trading:

**Settings:**
- **Default Rate**: $0.65 per contract (industry standard)
- **Location**: Admin → Settings → Trading Configuration
- **Range**: $0.00 - $50.00 per contract
- **Storage**: Database setting `OPTION_COMMISSION_PER_CONTRACT`

**Behavior:**
- **Opening Position**: Always charged (rate × contracts)
- **Closing Position**:
  - Bought to close (exitPrice > 0): Charged (rate × contracts)
  - Expired/Assigned (exitPrice = 0): No closing commission
- **User Override**: Commission can be customized per-trade when creating/editing

**Commission Field Editability:**
- ✓ Editable: New positions, open positions, closing positions
- ✗ Read-only: Already closed positions (protects historical data)

**Contracts Field Protection:**
- ✓ Editable: New positions, open positions without exit date
- ✗ Read-only: Positions with exit date (prevents partial closes)

**Database Storage:**
- Commission stored as total amount (not per-contract rate)
- For open/expired positions: commission = opening commission only
- For bought-back positions: commission = opening + closing commission

**MaxProfit Calculation:**
- MaxProfit = (Premium × Contracts × 100) - Opening Commission
- MaxProfit is constant regardless of how position closes
- Achieved when option expires worthless (100% of profit)
- Percent of Profit = (Actual Profit / Max Profit) × 100
```

### 6.2 Update README.md

Add to features list:
- Configurable commission rates per contract
- User-editable commission per trade (with defaults)
- Accurate commission tracking for expired vs. bought-to-close positions
- Protected fields to prevent data integrity issues

---

## PHASE 7: KNOWN LIMITATIONS & FUTURE ENHANCEMENTS

### Known Limitations

1. **Commission Rate Changes Mid-Trade**
   - If commission settings change between opening and closing, the PercentOfProfit calculation assumes equal rates
   - Impact: Minor - affects percentage display, not actual P&L
   - Workaround: User can manually override commission when closing

2. **No Commission History**
   - System doesn't track what commission rate was "current" at time of trade
   - Can't retroactively show "what rate was used"
   - Mitigation: Actual commission paid is stored accurately

3. **Partial Position Closes Not Supported**
   - Cannot close half of a 10-contract position
   - Contracts field is protected once exit date entered
   - This is by design to keep system simple

### Future Enhancements

1. **Commission History Tracking**
   - Store opening and closing commissions separately
   - Track commission rate changes over time
   - Show "effective rate" on historical trades

2. **Per-Broker Commission Profiles**
   - Support multiple brokers with different rates
   - Tag trades with broker ID
   - Calculate blended commission rates

3. **Commission Reporting**
   - Total commissions paid (monthly, yearly, all-time)
   - Commission as % of premium collected
   - Compare commission costs across strategies
   - Identify high-commission periods

4. **Advanced Fee Types**
   - Regulatory fees (ORF, TAF, SEC fees)
   - Exchange fees
   - Assignment fees (when put is assigned to stock)
   - Exercise fees (when call is exercised)

5. **Commission Validation & Warnings**
   - Warn if commission seems unusually high
   - Suggest reviewing broker fees
   - Track commission rate changes with notifications

6. **Bulk Edit Commission**
   - Update commission for multiple open positions
   - Useful if broker changed rates
   - With audit trail

---

## SUMMARY & RECOMMENDATIONS

### Key Benefits of This Migration

✅ **User Flexibility** - Commission can be changed per-trade or globally
✅ **Accuracy** - No longer charges commission on expired options
✅ **Data Integrity** - Protects closed positions and prevents partial closes
✅ **Simplicity** - User controls commission, no complex auto-calculations
✅ **Backward Compatible** - Defaults to $0.65 if not configured
✅ **Well Tested** - Comprehensive unit and integration tests

### Implementation Timeline

| Phase | Duration | Owner |
|-------|----------|-------|
| Code Changes | 6-8 hours | Developer |
| Unit Testing | 3-4 hours | Developer |
| Integration Testing | 2-3 hours | Developer |
| UI Testing | 2 hours | Developer |
| Code Review | 1-2 hours | Team |
| Deployment | 1 hour | Ops |
| **Total** | **15-20 hours** | |

### Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Breaking existing trades | Low | High | Comprehensive tests, backward compatibility |
| Database corruption | Very Low | High | Mandatory backup before deployment |
| Commission calculation wrong | Low | Medium | Unit tests cover all edge cases |
| UI confusion | Medium | Low | Clear labels, tooltips, read-only fields |
| Edit protection not working | Low | Medium | Integration tests verify field protection |

### Go/No-Go Criteria

✅ **GO if:**
- All unit tests pass
- Integration tests pass
- Manual testing completed
- Code reviewed
- Backup created
- Tested in dev environment

❌ **NO-GO if:**
- Any test failures
- No backup available
- Code review not complete
- Critical bugs found in testing

---

## APPENDIX A: TEST CODE SAMPLES

### A.1 Setting Service Tests

See `test/setting_test.go` for complete implementation. Key tests include:

- `TestSettingService_GetFloatValue` - Tests float parsing and defaults
- `TestSettingService_GetFloatValueWithValidation` - Tests min/max clamping
- `TestSettingService_CommissionSetting` - Tests commission-specific behavior

### A.2 Option Commission Tests

See `test/option_commission_test.go` for complete implementation. Key tests include:

- `TestOptionService_OpeningCommission` - Tests opening commission calculation
- `TestOptionService_ClosingCommission_BoughtToClose` - Tests closing commission added
- `TestOptionService_ClosingCommission_Expired` - Tests no closing commission
- `TestOption_TotalProfit` - Tests profit calculations
- `TestOption_PercentOfProfit` - Tests maxProfit and percentage calculations

---

## APPENDIX B: MIGRATION CHECKLIST

Use this checklist during implementation:

### Code Changes
- [ ] Step 1: SettingService enhanced with GetFloatValue()
- [ ] Step 2: OptionService refactored (constructor, Create, Close, CloseByID)
- [ ] Step 3: CalculatePercentOfProfit() fixed in symbol.go
- [ ] Step 4: Server.go updated (OptionService initialization)
- [ ] Step 5: Option handlers updated (create, update)
- [ ] Step 6: Settings UI enhanced (HTML/CSS/JS)
- [ ] Step 7: Settings handler updated
- [ ] Step 8: Option edit UI updated (commission field, contracts protection)
- [ ] Step 9: Schema migration added (optional)

### Testing
- [ ] Unit tests written for SettingService
- [ ] Unit tests written for Option commission logic
- [ ] Integration tests written
- [ ] All tests passing
- [ ] Manual testing completed

### Documentation
- [ ] CLAUDE.md updated
- [ ] README.md updated
- [ ] Migration plan reviewed

### Deployment
- [ ] Pre-deployment checklist complete
- [ ] Backup created
- [ ] Code deployed
- [ ] Post-deployment validation complete

---

## DOCUMENT HISTORY

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2024-10-26 | Claude | Initial comprehensive migration plan |

---

**END OF DOCUMENT**
