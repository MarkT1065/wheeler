package web

import (
	"encoding/json"
	"fmt"
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
		OptionsSummary: optionsSummary,
		OpenPositions:  openPositions,
		SummaryTotals:  summaryTotals,
		CurrentDB:      s.getCurrentDatabaseName(),
	}

	log.Printf("[OPTIONS PAGE] Rendering options.html template with %d summaries and %d open positions", len(optionsSummary), len(openPositions))
	s.renderTemplate(w, "options.html", data)
	log.Printf("[OPTIONS PAGE] Successfully completed options page request")
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

	_, err = s.optionService.Create(symbol, optionType, time.Now(), strike, expiration, premium, contracts)
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

	// Create the option
	option, err := s.optionService.CreateWithCommission(req.Symbol, req.Type, opened, req.Strike, expiration, req.Premium, req.Contracts, req.Commission)
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

	// Update the option
	option, err := s.optionService.UpdateByID(*req.ID, req.Symbol, req.Type, opened, req.Strike, expiration, req.Premium, req.Contracts, req.Commission, closed, req.ExitPrice)
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