package web

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// polygonTestHandler tests the Polygon API connection
func (s *Server) polygonTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("[POLYGON API] Testing API connection")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test the connection
	err := s.polygonService.TestConnection(ctx)
	
	response := map[string]interface{}{
		"success": err == nil,
	}

	if err != nil {
		response["error"] = err.Error()
		log.Printf("[POLYGON API] Connection test failed: %v", err)
	} else {
		response["message"] = "API key is valid and connection successful"
		log.Printf("[POLYGON API] Connection test successful")
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[POLYGON API] Error encoding test response: %v", err)
	}
}

// polygonUpdatePricesHandler triggers price updates for symbols
func (s *Server) polygonUpdatePricesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("[POLYGON API] Starting price update request")

	// Parse request body to get specific symbols (optional)
	var request struct {
		Symbols []string `json:"symbols,omitempty"`
		All     bool     `json:"all,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// If decoding fails, default to updating all symbols
		request.All = true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var updated, failed int
	var errors []string

	if request.All || len(request.Symbols) == 0 {
		// Update all symbols
		symbols, err := s.symbolService.GetDistinctSymbols()
		if err != nil {
			log.Printf("[POLYGON API] Error getting symbols: %v", err)
			http.Error(w, "Failed to get symbols", http.StatusInternalServerError)
			return
		}

		log.Printf("[POLYGON API] Updating prices for %d symbols", len(symbols))

		for _, symbol := range symbols {
			if err := s.polygonService.UpdateSymbolPrice(ctx, symbol); err != nil {
				log.Printf("[POLYGON API] Failed to update %s: %v", symbol, err)
				errors = append(errors, symbol+": "+err.Error())
				failed++
			} else {
				updated++
			}

			// Rate limiting for free tier (5 requests per minute)
			time.Sleep(12 * time.Second)
		}
	} else {
		// Update specific symbols
		log.Printf("[POLYGON API] Updating prices for specific symbols: %v", request.Symbols)

		for _, symbol := range request.Symbols {
			if err := s.polygonService.UpdateSymbolPrice(ctx, symbol); err != nil {
				log.Printf("[POLYGON API] Failed to update %s: %v", symbol, err)
				errors = append(errors, symbol+": "+err.Error())
				failed++
			} else {
				updated++
			}

			// Rate limiting
			if len(request.Symbols) > 1 {
				time.Sleep(12 * time.Second)
			}
		}
	}

	response := map[string]interface{}{
		"success": updated > 0,
		"updated": updated,
		"failed":  failed,
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	if updated > 0 {
		response["message"] = "Price update completed"
	} else {
		response["message"] = "No prices were updated"
	}

	log.Printf("[POLYGON API] Price update completed: %d updated, %d failed", updated, failed)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[POLYGON API] Error encoding update response: %v", err)
	}
}

// polygonSymbolInfoHandler gets detailed symbol information from Polygon
func (s *Server) polygonSymbolInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract symbol from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/polygon/symbol-info/")
	if path == "" {
		http.Error(w, "Symbol required", http.StatusBadRequest)
		return
	}

	symbol := strings.ToUpper(path)
	log.Printf("[POLYGON API] Getting symbol info for: %s", symbol)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get symbol info from Polygon
	info, err := s.polygonService.FetchSymbolDetails(ctx, symbol)
	if err != nil {
		log.Printf("[POLYGON API] Error getting symbol info for %s: %v", symbol, err)
		response := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get dividend history (optional)
	dividends, err := s.polygonService.FetchDividendHistory(ctx, symbol, 5)
	if err != nil {
		log.Printf("[POLYGON API] Warning: failed to get dividend history for %s: %v", symbol, err)
		// Continue without dividends
	}

	response := map[string]interface{}{
		"success":    true,
		"symbolInfo": info,
	}

	if len(dividends) > 0 {
		response["dividends"] = dividends
	}

	log.Printf("[POLYGON API] Successfully retrieved info for %s", symbol)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[POLYGON API] Error encoding symbol info response: %v", err)
	}
}

// polygonStatusHandler returns the status of the Polygon integration  
func (s *Server) polygonStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := s.polygonService.GetAPIKeyStatus()

	// Test connection if API key is configured
	if status.Configured {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.polygonService.TestConnection(ctx); err != nil {
			status.Valid = false
			status.Error = err.Error()
		} else {
			status.Valid = true
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Printf("[POLYGON API] Error encoding status response: %v", err)
	}
}

// polygonFetchDividendsHandler fetches dividend data for all symbols
func (s *Server) polygonFetchDividendsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("[POLYGON API] Starting bulk dividend fetch request")

	// Parse request body to get specific symbols (optional)
	var request struct {
		Symbols []string `json:"symbols,omitempty"`
		All     bool     `json:"all,omitempty"`
		Limit   int      `json:"limit,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// If decoding fails, default to fetching all symbols
		request.All = true
	}

	if request.Limit == 0 {
		request.Limit = 10 // Default to 10 recent dividends
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var processed int
	var results []map[string]interface{}
	var errors []string

	if request.All || len(request.Symbols) == 0 {
		// Fetch dividends for all symbols
		symbols, err := s.symbolService.GetDistinctSymbols()
		if err != nil {
			log.Printf("[POLYGON API] Error getting symbols: %v", err)
			http.Error(w, "Failed to get symbols", http.StatusInternalServerError)
			return
		}

		log.Printf("[POLYGON API] Fetching dividends for %d symbols", len(symbols))

		for _, symbol := range symbols {
			dividends, err := s.polygonService.FetchDividendHistory(ctx, symbol, request.Limit)
			processed++
			
			if err != nil {
				log.Printf("[POLYGON API] Failed to fetch dividends for %s: %v", symbol, err)
				errors = append(errors, symbol+": "+err.Error())
			} else {
				results = append(results, map[string]interface{}{
					"symbol":    symbol,
					"dividends": dividends,
					"count":     len(dividends),
				})
			}

			// Rate limiting for free tier (5 requests per minute)
			time.Sleep(12 * time.Second)
		}
	} else {
		// Fetch dividends for specific symbols
		log.Printf("[POLYGON API] Fetching dividends for specific symbols: %v", request.Symbols)

		for _, symbol := range request.Symbols {
			dividends, err := s.polygonService.FetchDividendHistory(ctx, symbol, request.Limit)
			processed++
			
			if err != nil {
				log.Printf("[POLYGON API] Failed to fetch dividends for %s: %v", symbol, err)
				errors = append(errors, symbol+": "+err.Error())
			} else {
				results = append(results, map[string]interface{}{
					"symbol":    symbol,
					"dividends": dividends,
					"count":     len(dividends),
				})
			}

			// Rate limiting
			if len(request.Symbols) > 1 {
				time.Sleep(12 * time.Second)
			}
		}
	}

	response := map[string]interface{}{
		"success":   true,
		"processed": processed,
		"results":   results,
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	totalDividends := 0
	for _, result := range results {
		if count, ok := result["count"].(int); ok {
			totalDividends += count
		}
	}
	
	response["message"] = fmt.Sprintf("Dividend fetch completed: %d symbols processed, %d total dividends found", processed, totalDividends)

	log.Printf("[POLYGON API] Dividend fetch completed: %d symbols processed, %d total dividends found", processed, totalDividends)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[POLYGON API] Error encoding dividend fetch response: %v", err)
	}
}