package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"stonks/internal/models"
	"strconv"
	"strings"
	"time"
)


// optionsHandler serves the options analysis view
func (s *Server) optionsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[OPTIONS PAGE] %s %s - Start processing options page request", r.Method, r.URL.Path)

	symbols, err := s.symbolService.GetDistinctSymbols()
	if err != nil {
		log.Printf("[OPTIONS PAGE] WARNING: Failed to get symbols for navigation: %v", err)
		symbols = []string{}
	} else {
		log.Printf("[OPTIONS PAGE] Retrieved %d symbols for navigation", len(symbols))
	}

	// Get options summary by symbol
	log.Printf("[OPTIONS PAGE] Fetching options summary data")
	optionsSummary, err := s.optionService.GetOptionsSummaryBySymbol()
	if err != nil {
		log.Printf("[OPTIONS PAGE] ERROR: Failed to get options summary: %v", err)
		optionsSummary = []*models.OptionSummary{}
	} else {
		log.Printf("[OPTIONS PAGE] Retrieved %d options summaries", len(optionsSummary))
	}

	// Get open positions with details
	log.Printf("[OPTIONS PAGE] Fetching open positions data")
	openPositions, err := s.optionService.GetOpenPositionsWithDetails()
	if err != nil {
		log.Printf("[OPTIONS PAGE] ERROR: Failed to get open positions: %v", err)
		openPositions = []*models.OpenPositionData{}
	} else {
		log.Printf("[OPTIONS PAGE] Retrieved %d open positions", len(openPositions))
	}

	// Get summary totals
	log.Printf("[OPTIONS PAGE] Calculating summary totals")
	summaryTotals, err := s.optionService.GetOptionsSummaryTotals()
	if err != nil {
		log.Printf("[OPTIONS PAGE] ERROR: Failed to get summary totals: %v", err)
		summaryTotals = &models.OptionSummary{}
	} else {
		log.Printf("[OPTIONS PAGE] Calculated summary totals: %d total positions", summaryTotals.TotalPositions)
	}

	data := OptionsData{
		Symbols:        symbols,
		AllSymbols:     symbols, // For navigation compatibility
		OptionsSummary: optionsSummary,
		OpenPositions:  openPositions,
		SummaryTotals:  summaryTotals,
		CurrentDB:      s.getCurrentDatabaseName(),
		ActivePage:     "options",
	}

	log.Printf("[OPTIONS PAGE] Rendering options.html template with %d summaries and %d open positions", len(optionsSummary), len(openPositions))
	s.renderTemplate(w, "options.html", data)
	log.Printf("[OPTIONS PAGE] Successfully completed options page request")
}

// allOptionsHandler serves the all options view with complete sortable table
func (s *Server) allOptionsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ALL OPTIONS PAGE] %s %s - Start processing all options page request", r.Method, r.URL.Path)

	symbols, err := s.symbolService.GetDistinctSymbols()
	if err != nil {
		log.Printf("[ALL OPTIONS PAGE] WARNING: Failed to get symbols for navigation: %v", err)
		symbols = []string{}
	} else {
		log.Printf("[ALL OPTIONS PAGE] Retrieved %d symbols for navigation", len(symbols))
	}

	// Create the options index using the OptionService Index method
	log.Printf("[ALL OPTIONS PAGE] Creating options index")
	optionsIndex, err := s.optionService.Index()
	if err != nil {
		log.Printf("[ALL OPTIONS PAGE] ERROR: Failed to create options index: %v", err)
		optionsIndex = make(map[string]interface{})
	} else {
		log.Printf("[ALL OPTIONS PAGE] Successfully created options index")
	}

	// Count total options from index for logging
	var totalOptions int
	if idIndex, ok := optionsIndex["id"].(map[string]*models.Option); ok {
		totalOptions = len(idIndex)
	}

	// Convert options index to JSON for template
	indexJSON, err := json.Marshal(optionsIndex)
	if err != nil {
		log.Printf("[ALL OPTIONS PAGE] ERROR: Failed to marshal options index to JSON: %v", err)
		indexJSON = []byte("{}")
	}

	data := AllOptionsDataWithJSON{
		Symbols:         symbols,
		AllSymbols:      symbols, // For navigation compatibility
		OptionsIndex:    optionsIndex,
		OptionsIndexJSON: template.JS(string(indexJSON)),
		CurrentDB:       s.getCurrentDatabaseName(),
		ActivePage:      "options",
	}

	log.Printf("[ALL OPTIONS PAGE] Rendering all-options.html template with index containing %d options", totalOptions)
	s.renderTemplate(w, "all-options.html", data)
	log.Printf("[ALL OPTIONS PAGE] Successfully completed all options page request")
}

// addOptionHandler handles form submission for adding new options
func (s *Server) addOptionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	symbol := r.FormValue("symbol")
	optionType := r.FormValue("type")
	strikeStr := r.FormValue("strike")
	expirationStr := r.FormValue("expiration")
	premiumStr := r.FormValue("premium")
	contractsStr := r.FormValue("contracts")

	strike, err := strconv.ParseFloat(strikeStr, 64)
	if err != nil {
		http.Error(w, "Invalid strike price", http.StatusBadRequest)
		return
	}

	expiration, err := time.Parse("2006-01-02", expirationStr)
	if err != nil {
		http.Error(w, "Invalid expiration date", http.StatusBadRequest)
		return
	}

	premium, err := strconv.ParseFloat(premiumStr, 64)
	if err != nil {
		http.Error(w, "Invalid premium", http.StatusBadRequest)
		return
	}

	contracts, err := strconv.Atoi(contractsStr)
	if err != nil {
		http.Error(w, "Invalid contracts", http.StatusBadRequest)
		return
	}

	// Get commission per contract from settings and calculate total
	commissionPerContract := s.optionService.GetCommissionPerContract()
	totalCommission := commissionPerContract * float64(contracts)

	_, err = s.optionService.CreateWithCommission(symbol, optionType, time.Now(), strike, expiration, premium, contracts, totalCommission)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// optionAPIHandler handles CRUD operations for options
func (s *Server) optionAPIHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[OPTION API] %s %s - Processing option API request", r.Method, r.URL.Path)

	switch r.Method {
	case http.MethodPost:
		s.createOption(w, r)
	case http.MethodPut:
		s.updateOption(w, r)
	case http.MethodDelete:
		s.deleteOption(w, r)
	default:
		log.Printf("[OPTION API] ERROR: Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// individualOptionAPIHandler handles GET requests for individual options by ID
func (s *Server) individualOptionAPIHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INDIVIDUAL OPTION API] %s %s - Processing individual option API request", r.Method, r.URL.Path)

	if r.Method != http.MethodGet {
		log.Printf("[INDIVIDUAL OPTION API] ERROR: Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract option ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/options/")
	if path == "" {
		log.Printf("[INDIVIDUAL OPTION API] ERROR: No option ID provided")
		http.Error(w, "Option ID is required", http.StatusBadRequest)
		return
	}

	optionID, err := strconv.Atoi(path)
	if err != nil {
		log.Printf("[INDIVIDUAL OPTION API] ERROR: Invalid option ID: %s", path)
		http.Error(w, "Invalid option ID", http.StatusBadRequest)
		return
	}

	// Fetch option by ID
	option, err := s.optionService.GetByID(optionID)
	if err != nil {
		log.Printf("[INDIVIDUAL OPTION API] ERROR: Failed to get option by ID %d: %v", optionID, err)
		http.Error(w, "Option not found", http.StatusNotFound)
		return
	}

	log.Printf("[INDIVIDUAL OPTION API] Successfully retrieved option: %d", optionID)

	// Return option data as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(option); err != nil {
		log.Printf("[INDIVIDUAL OPTION API] ERROR: Failed to encode option to JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// createOption handles POST requests to create new options
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

	// Calculate total commission for storage
	// For new options, just multiply by contracts
	totalCommission := commissionPerContract * float64(req.Contracts)

	// Create the option - store total commission in database
	option, err := s.optionService.CreateWithCommission(req.Symbol, req.Type, opened, req.Strike, expiration, req.Premium, req.Contracts, totalCommission)
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

		err = s.optionService.Close(req.Symbol, req.Type, opened, req.Strike, expiration, req.Premium, req.Contracts, closed, exitPrice)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to close option: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(option)
}

// updateOption handles PUT requests to update existing options
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

	// Get commission per contract
	// If provided in request, use it; otherwise get from settings
	commissionPerContract := req.CommissionPerContract
	if commissionPerContract <= 0 {
		commissionPerContract = s.optionService.GetCommissionPerContract()
	}

	// Calculate total commission for storage
	totalCommission := commissionPerContract * float64(req.Contracts)

	// If option is closed with exit price > 0, double the commission (opening + closing)
	if closed != nil && req.ExitPrice != nil && *req.ExitPrice > 0 {
		totalCommission = totalCommission * 2.0
	}

	// Update the option - store total commission in database
	option, err := s.optionService.UpdateByID(*req.ID, req.Symbol, req.Type, opened, req.Strike, expiration, req.Premium, req.Contracts, totalCommission, closed, req.ExitPrice)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update option: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(option)
}

// deleteOption handles DELETE requests to remove options
func (s *Server) deleteOption(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DELETE OPTION] Starting DELETE request")

	var req OptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[DELETE OPTION] ERROR: Invalid JSON payload: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("[DELETE OPTION] Request decoded: ID=%v, Symbol=%s, Type=%s", req.ID, req.Symbol, req.Type)

	// Prefer ID-based deletion if ID is provided
	if req.ID != nil {
		log.Printf("[DELETE OPTION] Using ID-based deletion for ID: %d", *req.ID)
		err := s.optionService.DeleteByID(*req.ID)
		if err != nil {
			log.Printf("[DELETE OPTION] ERROR: ID-based deletion failed for ID %d: %v", *req.ID, err)
			http.Error(w, fmt.Sprintf("Failed to delete option: %v", err), http.StatusInternalServerError)
			return
		}
		log.Printf("[DELETE OPTION] Successfully deleted option with ID: %d", *req.ID)
	} else {
		log.Printf("[DELETE OPTION] Using compound key deletion")
		// Fallback to compound key deletion
		// Validate required fields for deletion
		if req.Symbol == "" || req.Type == "" || req.Opened == "" || req.Expiration == "" {
			log.Printf("[DELETE OPTION] ERROR: Missing required fields for compound key deletion")
			http.Error(w, "Missing required fields for deletion", http.StatusBadRequest)
			return
		}

		// Parse dates
		opened, err := time.Parse("2006-01-02", req.Opened)
		if err != nil {
			log.Printf("[DELETE OPTION] ERROR: Invalid opened date format '%s': %v", req.Opened, err)
			http.Error(w, "Invalid opened date format", http.StatusBadRequest)
			return
		}

		expiration, err := time.Parse("2006-01-02", req.Expiration)
		if err != nil {
			log.Printf("[DELETE OPTION] ERROR: Invalid expiration date format '%s': %v", req.Expiration, err)
			http.Error(w, "Invalid expiration date format", http.StatusBadRequest)
			return
		}

		log.Printf("[DELETE OPTION] Attempting compound key deletion: Symbol=%s, Type=%s, Opened=%s, Strike=%f, Expiration=%s",
			req.Symbol, req.Type, req.Opened, req.Strike, req.Expiration)

		// Delete the option using compound key
		err = s.optionService.Delete(req.Symbol, req.Type, opened, req.Strike, expiration, req.Premium, req.Contracts)
		if err != nil {
			log.Printf("[DELETE OPTION] ERROR: Compound key deletion failed: %v", err)
			http.Error(w, fmt.Sprintf("Failed to delete option: %v", err), http.StatusInternalServerError)
			return
		}
		log.Printf("[DELETE OPTION] Successfully deleted option using compound key")
	}

	log.Printf("[DELETE OPTION] Sending success response")
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "success"}); err != nil {
		log.Printf("[DELETE OPTION] ERROR: Failed to encode success response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	log.Printf("[DELETE OPTION] Request completed successfully")
}

// optionsFilterHandler handles filtered queries on the options index
func (s *Server) optionsFilterHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[OPTIONS FILTER API] %s %s - Processing options filter request", r.Method, r.URL.Path)

	if r.Method != http.MethodPost {
		log.Printf("[OPTIONS FILTER API] ERROR: Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var filters models.FilterOptions
	if err := json.NewDecoder(r.Body).Decode(&filters); err != nil {
		log.Printf("[OPTIONS FILTER API] ERROR: Invalid JSON payload: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("[OPTIONS FILTER API] Filter request: %+v", filters)

	// Get the options index
	optionsIndex, err := s.optionService.Index()
	if err != nil {
		log.Printf("[OPTIONS FILTER API] ERROR: Failed to create options index: %v", err)
		http.Error(w, "Failed to create index", http.StatusInternalServerError)
		return
	}

	// Apply filters
	filteredOptions := models.GetByFilters(optionsIndex, filters)
	log.Printf("[OPTIONS FILTER API] Filtered results: %d options", len(filteredOptions))

	// Return filtered results
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(filteredOptions); err != nil {
		log.Printf("[OPTIONS FILTER API] ERROR: Failed to encode options to JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("[OPTIONS FILTER API] Successfully returned %d filtered options", len(filteredOptions))
}