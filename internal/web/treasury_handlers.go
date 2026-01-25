package web

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"stonks/internal/models"
	"strconv"
	"strings"
	"time"
)


// treasuriesHandler serves the treasuries view
func (s *Server) treasuriesHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[TREASURIES PAGE] %s %s - Start processing treasuries page request", r.Method, r.URL.Path)

	symbols, err := s.symbolService.GetDistinctSymbols()
	if err != nil {
		log.Printf("[TREASURIES PAGE] WARNING: Failed to get symbols for navigation: %v", err)
		// Log error but continue with empty symbols list
		symbols = []string{}
	} else {
		log.Printf("[TREASURIES PAGE] Retrieved %d symbols for navigation", len(symbols))
	}

	// Get all treasuries
	log.Printf("[TREASURIES PAGE] Fetching all treasuries from service")
	treasuries, err := s.treasuryService.GetAll()
	if err != nil {
		log.Printf("[TREASURIES PAGE] ERROR: Failed to get treasuries: %v", err)
		// Log error but continue with empty treasuries list
		treasuries = []*models.Treasury{}
	} else {
		log.Printf("[TREASURIES PAGE] Retrieved %d treasuries from service", len(treasuries))
	}

	// Get all options for put exposure chart
	log.Printf("[TREASURIES PAGE] Fetching all options from service")
	options, err := s.optionService.GetAll()
	if err != nil {
		log.Printf("[TREASURIES PAGE] ERROR: Failed to get options: %v", err)
		// Log error but continue with empty options list
		options = []*models.Option{}
	} else {
		log.Printf("[TREASURIES PAGE] Retrieved %d options from service", len(options))
	}

	// Sort treasuries by days remaining: active positions by days ascending, then sold positions
	sort.Slice(treasuries, func(i, j int) bool {
		iHasExit := treasuries[i].ExitPrice != nil
		jHasExit := treasuries[j].ExitPrice != nil

		// If one has exit price and other doesn't, put the one without exit price first
		if iHasExit != jHasExit {
			return !iHasExit
		}

		// If both are active (no exit price), sort by days remaining ascending
		if !iHasExit && !jHasExit {
			return treasuries[i].CalculateDaysRemaining() < treasuries[j].CalculateDaysRemaining()
		}

		// If both are sold, maintain original order (or could sort by exit date)
		return false
	})

	// Calculate summary data
	log.Printf("[TREASURIES PAGE] Calculating summary data for %d treasuries", len(treasuries))
	summary := calculateTreasuriesSummary(treasuries)
	log.Printf("[TREASURIES PAGE] Summary calculated: TotalAmount=%.2f, ActivePositions=%d",
		summary.TotalAmount, summary.ActivePositions)

	data := TreasuriesData{
		Symbols:    symbols,
		AllSymbols: symbols, // For navigation compatibility
		Treasuries: treasuries,
		Options:    options,
		Summary:    summary,
		CurrentDB:  s.getCurrentDatabaseName(),
		ActivePage: "treasuries",
	}

	log.Printf("[TREASURIES PAGE] Rendering treasuries.html template with %d treasuries", len(treasuries))
	s.renderTemplate(w, "treasuries.html", data)
	log.Printf("[TREASURIES PAGE] Successfully completed treasuries page request")
}

// calculateTreasuriesSummary calculates summary statistics for treasuries
func calculateTreasuriesSummary(treasuries []*models.Treasury) TreasuriesSummary {
	var totalAmount, totalBuyPrice, totalProfitLoss, totalInterest float64
	var currentlyHeld float64 // Only sum open positions for "Currently Held"
	activePositions := 0
	var totalDuration int // Sum of durations for averaging

	for _, treasury := range treasuries {
		// Always include in totals for full portfolio view
		totalAmount += treasury.Amount
		totalBuyPrice += treasury.BuyPrice
		totalProfitLoss += treasury.CalculateProfitLoss()
		totalInterest += treasury.CalculateInterest()

		// Count as active if no exit price is set
		if treasury.ExitPrice == nil {
			activePositions++
			currentlyHeld += treasury.BuyPrice // Only include open positions in "Currently Held"
		}
		
		// Calculate duration from purchase to maturity
		duration := int(treasury.Maturity.Sub(treasury.Purchased).Hours() / 24)
		totalDuration += duration
	}

	averageReturn := 0.0
	if totalBuyPrice > 0 {
		// Average Return = (Total Interest Earned / Total Buy Prices) * 100
		averageReturn = (totalInterest / totalBuyPrice) * 100
	}
	
	averageDuration := 0
	if len(treasuries) > 0 {
		averageDuration = totalDuration / len(treasuries)
	}

	return TreasuriesSummary{
		TotalAmount:     totalAmount,
		TotalBuyPrice:   currentlyHeld, // Now represents "Currently Held" (open positions only)
		TotalProfitLoss: totalProfitLoss,
		TotalInterest:   totalInterest,
		AverageReturn:   averageReturn,
		ActivePositions: activePositions,
		AverageDuration: averageDuration,
	}
}

// addTreasuryHandler handles form submission for adding new treasuries
func (s *Server) addTreasuryHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ADD TREASURY] Starting POST request to add new treasury")

	if r.Method != http.MethodPost {
		log.Printf("[ADD TREASURY] ERROR: Invalid method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cuspid := r.FormValue("cuspid")
	purchasedStr := r.FormValue("purchased")
	maturityStr := r.FormValue("maturity")
	amountStr := r.FormValue("amount")
	yieldStr := r.FormValue("yield")
	buyPriceStr := r.FormValue("buyPrice")
	currentValueStr := r.FormValue("currentValue")
	exitPriceStr := r.FormValue("exitPrice")

	log.Printf("[ADD TREASURY] Form values: CUSPID=%s, Purchased=%s, Maturity=%s, Amount=%s, Yield=%s, BuyPrice=%s, CurrentValue=%s, ExitPrice=%s",
		cuspid, purchasedStr, maturityStr, amountStr, yieldStr, buyPriceStr, currentValueStr, exitPriceStr)

	purchased, err := time.Parse("2006-01-02", purchasedStr)
	if err != nil {
		log.Printf("[ADD TREASURY] ERROR: Invalid purchased date '%s': %v", purchasedStr, err)
		http.Error(w, "Invalid purchased date", http.StatusBadRequest)
		return
	}

	maturity, err := time.Parse("2006-01-02", maturityStr)
	if err != nil {
		log.Printf("[ADD TREASURY] ERROR: Invalid maturity date '%s': %v", maturityStr, err)
		http.Error(w, "Invalid maturity date", http.StatusBadRequest)
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		log.Printf("[ADD TREASURY] ERROR: Invalid amount '%s': %v", amountStr, err)
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	yield, err := strconv.ParseFloat(yieldStr, 64)
	if err != nil {
		log.Printf("[ADD TREASURY] ERROR: Invalid yield '%s': %v", yieldStr, err)
		http.Error(w, "Invalid yield", http.StatusBadRequest)
		return
	}

	buyPrice, err := strconv.ParseFloat(buyPriceStr, 64)
	if err != nil {
		log.Printf("[ADD TREASURY] ERROR: Invalid buy price '%s': %v", buyPriceStr, err)
		http.Error(w, "Invalid buy price", http.StatusBadRequest)
		return
	}

	// Parse optional fields
	var currentValue, exitPrice *float64
	if currentValueStr != "" {
		if cv, err := strconv.ParseFloat(currentValueStr, 64); err == nil {
			currentValue = &cv
		}
	}
	if exitPriceStr != "" {
		if ep, err := strconv.ParseFloat(exitPriceStr, 64); err == nil {
			exitPrice = &ep
		}
	}

	log.Printf("[ADD TREASURY] Parsed values: CUSPID=%s, Purchased=%v, Maturity=%v, Amount=%.2f, Yield=%.3f, BuyPrice=%.2f, CurrentValue=%v, ExitPrice=%v",
		cuspid, purchased, maturity, amount, yield, buyPrice, currentValue, exitPrice)
	log.Printf("[ADD TREASURY] Calling CreateFull service for CUSPID: %s", cuspid)

	_, err = s.treasuryService.CreateFull(cuspid, purchased, maturity, amount, yield, buyPrice, currentValue, exitPrice)
	if err != nil {
		log.Printf("[ADD TREASURY] ERROR: Service layer failed to create CUSPID %s: %v", cuspid, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[ADD TREASURY] Successfully created treasury for CUSPID: %s", cuspid)
	log.Printf("[ADD TREASURY] Redirecting to /treasuries")

	http.Redirect(w, r, "/treasuries", http.StatusSeeOther)
}

// treasuryAPIHandler handles all treasury API requests
func (s *Server) treasuryAPIHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[TREASURY API] %s %s - Start processing", r.Method, r.URL.Path)

	// Set JSON content type
	w.Header().Set("Content-Type", "application/json")

	// Extract CUSPID from URL path
	cuspid := strings.TrimPrefix(r.URL.Path, "/api/treasuries/")
	if cuspid == "" {
		log.Printf("[TREASURY API] ERROR: No CUSPID provided in URL path: %s", r.URL.Path)
		http.Error(w, "CUSPID is required", http.StatusBadRequest)
		return
	}

	log.Printf("[TREASURY API] Extracted CUSPID: '%s' from path: %s", cuspid, r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		log.Printf("[TREASURY API] Routing to GET handler for CUSPID: %s", cuspid)
		s.getTreasuryHandler(w, r, cuspid)
	case http.MethodPut:
		log.Printf("[TREASURY API] Routing to PUT handler for CUSPID: %s", cuspid)
		s.updateTreasuryHandler(w, r, cuspid)
	case http.MethodDelete:
		log.Printf("[TREASURY API] Routing to DELETE handler for CUSPID: %s", cuspid)
		s.deleteTreasuryHandler(w, r, cuspid)
	default:
		log.Printf("[TREASURY API] ERROR: Unsupported method: %s for CUSPID: %s", r.Method, cuspid)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getTreasuryHandler handles GET requests for a specific treasury
func (s *Server) getTreasuryHandler(w http.ResponseWriter, r *http.Request, cuspid string) {
	log.Printf("[GET TREASURY] Starting GET request for CUSPID: %s", cuspid)

	treasury, err := s.treasuryService.GetByCUSPID(cuspid)
	if err != nil {
		log.Printf("[GET TREASURY] ERROR: Failed to get treasury for CUSPID %s: %v", cuspid, err)
		http.Error(w, "Treasury not found", http.StatusNotFound)
		return
	}

	log.Printf("[GET TREASURY] Successfully retrieved treasury for CUSPID: %s", cuspid)
	log.Printf("[GET TREASURY] Treasury data: Amount=%.2f, Yield=%.3f, BuyPrice=%.2f",
		treasury.Amount, treasury.Yield, treasury.BuyPrice)

	if err := json.NewEncoder(w).Encode(treasury); err != nil {
		log.Printf("[GET TREASURY] ERROR: Failed to encode JSON response for CUSPID %s: %v", cuspid, err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("[GET TREASURY] Successfully sent response for CUSPID: %s", cuspid)
}

// updateTreasuryHandler handles PUT requests to update a treasury
func (s *Server) updateTreasuryHandler(w http.ResponseWriter, r *http.Request, cuspid string) {
	log.Printf("[UPDATE TREASURY] Starting PUT request for CUSPID: %s", cuspid)

	var updateReq TreasuryUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		log.Printf("[UPDATE TREASURY] ERROR: Failed to decode JSON request for CUSPID %s: %v", cuspid, err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("[UPDATE TREASURY] Decoded request for CUSPID %s: Purchased=%s, Maturity=%s, Amount=%.2f, Yield=%.3f, BuyPrice=%.2f",
		cuspid, updateReq.Purchased, updateReq.Maturity, updateReq.Amount, updateReq.Yield, updateReq.BuyPrice)

	// Parse dates
	purchased, err := time.Parse("2006-01-02", updateReq.Purchased)
	if err != nil {
		log.Printf("[UPDATE TREASURY] ERROR: Invalid purchased date format for CUSPID %s: '%s' - %v",
			cuspid, updateReq.Purchased, err)
		http.Error(w, "Invalid purchased date format", http.StatusBadRequest)
		return
	}

	maturity, err := time.Parse("2006-01-02", updateReq.Maturity)
	if err != nil {
		log.Printf("[UPDATE TREASURY] ERROR: Invalid maturity date format for CUSPID %s: '%s' - %v",
			cuspid, updateReq.Maturity, err)
		http.Error(w, "Invalid maturity date format", http.StatusBadRequest)
		return
	}

	log.Printf("[UPDATE TREASURY] Parsed dates for CUSPID %s: Purchased=%v, Maturity=%v", cuspid, purchased, maturity)
	log.Printf("[UPDATE TREASURY] Calling UpdateFull service for CUSPID: %s", cuspid)

	// Update treasury using UpdateFull method
	updatedTreasury, err := s.treasuryService.UpdateFull(
		cuspid,
		purchased,
		maturity,
		updateReq.Amount,
		updateReq.Yield,
		updateReq.BuyPrice,
		updateReq.CurrentValue,
		updateReq.ExitPrice,
	)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[UPDATE TREASURY] ERROR: Treasury not found for CUSPID %s: %v", cuspid, err)
			http.Error(w, "Treasury not found", http.StatusNotFound)
		} else {
			log.Printf("[UPDATE TREASURY] ERROR: Service layer failed to update CUSPID %s: %v", cuspid, err)
			http.Error(w, "Failed to update treasury", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("[UPDATE TREASURY] Successfully updated treasury for CUSPID: %s", cuspid)
	log.Printf("[UPDATE TREASURY] Updated treasury data: Amount=%.2f, Yield=%.3f, BuyPrice=%.2f",
		updatedTreasury.Amount, updatedTreasury.Yield, updatedTreasury.BuyPrice)

	if err := json.NewEncoder(w).Encode(updatedTreasury); err != nil {
		log.Printf("[UPDATE TREASURY] ERROR: Failed to encode JSON response for CUSPID %s: %v", cuspid, err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("[UPDATE TREASURY] Successfully sent response for CUSPID: %s", cuspid)
}

// deleteTreasuryHandler handles DELETE requests for a treasury
func (s *Server) deleteTreasuryHandler(w http.ResponseWriter, r *http.Request, cuspid string) {
	log.Printf("[DELETE TREASURY] Starting DELETE request for CUSPID: %s", cuspid)

	err := s.treasuryService.Delete(cuspid)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[DELETE TREASURY] ERROR: Treasury not found for CUSPID %s: %v", cuspid, err)
			http.Error(w, "Treasury not found", http.StatusNotFound)
		} else {
			log.Printf("[DELETE TREASURY] ERROR: Service layer failed to delete CUSPID %s: %v", cuspid, err)
			http.Error(w, "Failed to delete treasury", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("[DELETE TREASURY] Successfully deleted treasury for CUSPID: %s", cuspid)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))

	log.Printf("[DELETE TREASURY] Successfully sent response for CUSPID: %s", cuspid)
}