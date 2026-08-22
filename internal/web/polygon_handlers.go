package web

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"stonks/internal/providers"
)

func (s *Server) providerLogPrefix() string {
	name := strings.ToUpper(strings.TrimSpace(s.providerService.Name()))
	if name == "" {
		name = "PROVIDER"
	}
	return fmt.Sprintf("[%s API]", name)
}

// providerTestHandler tests the data provider API connection
func (s *Server) providerTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	providerName := s.providerService.Name()
	logPrefix := s.providerLogPrefix()
	log.Printf("%s Testing API connection", logPrefix)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.providerService.TestConnection(ctx)

	response := map[string]interface{}{
		"success":  err == nil,
		"provider": providerName,
	}

	if err != nil {
		response["error"] = err.Error()
		log.Printf("%s Connection test failed: %v", logPrefix, err)
	} else {
		response["message"] = "API key is valid and connection successful"
		log.Printf("%s Connection test successful", logPrefix)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("%s Error encoding test response: %v", logPrefix, err)
	}
}

// providerUpdatePricesHandler triggers price updates for symbols
func (s *Server) providerUpdatePricesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logPrefix := s.providerLogPrefix()
	log.Printf("%s Starting price update request", logPrefix)

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

	var (
		result *providers.BulkPriceUpdateResult
		err    error
	)

	if request.All || len(request.Symbols) == 0 {
		result, err = s.providerService.UpdateAllSymbolPrices(ctx)
	} else {
		result, err = s.providerService.UpdateSymbolPrices(ctx, request.Symbols)
	}

	if result == nil {
		result = &providers.BulkPriceUpdateResult{}
	}

	success := err == nil && result.Updated > 0
	response := map[string]interface{}{
		"success": success,
		"updated": result.Updated,
		"failed":  result.Failed,
	}

	if len(result.Errors) > 0 {
		response["errors"] = result.Errors
	}

	if err != nil {
		response["message"] = err.Error()
		log.Printf("%s Price update failed: %v", logPrefix, err)
	} else if success {
		response["message"] = "Price update completed"
		log.Printf("%s Price update completed: %d updated, %d failed", logPrefix, result.Updated, result.Failed)
	} else {
		response["message"] = "No prices were updated"
		log.Printf("%s Price update completed with no updates", logPrefix)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("%s Error encoding update response: %v", logPrefix, err)
	}
}

// providerSymbolInfoHandler gets detailed symbol information from the configured provider
func (s *Server) providerSymbolInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract symbol from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/provider/symbol-info/")
	if path == "" {
		http.Error(w, "Symbol required", http.StatusBadRequest)
		return
	}

	symbol := strings.ToUpper(path)
	logPrefix := s.providerLogPrefix()
	log.Printf("%s Getting symbol info for: %s", logPrefix, symbol)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get symbol info from configured data provider
	info, err := s.providerService.FetchSymbolDetails(ctx, symbol)
	if err != nil {
		log.Printf("%s Error getting symbol info for %s: %v", logPrefix, symbol, err)
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
	dividends, err := s.providerService.FetchDividendHistory(ctx, symbol, 5)
	if err != nil {
		log.Printf("%s Warning: failed to get dividend history for %s: %v", logPrefix, symbol, err)
		// Continue without dividends
	}

	response := map[string]interface{}{
		"success":    true,
		"symbolInfo": info,
	}

	if len(dividends) > 0 {
		response["dividends"] = dividends
	}

	log.Printf("%s Successfully retrieved info for %s", logPrefix, symbol)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("%s Error encoding symbol info response: %v", logPrefix, err)
	}
}

// providerStatusHandler returns the status of the provider integration
func (s *Server) providerStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := s.providerService.GetAPIKeyStatus()
	logPrefix := s.providerLogPrefix()

	// Test connection if API key is configured
	if status.Configured {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.providerService.TestConnection(ctx); err != nil {
			status.Valid = false
			status.Error = err.Error()
		} else {
			status.Valid = true
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Printf("%s Error encoding status response: %v", logPrefix, err)
	}
}

// providerFetchDividendsHandler fetches dividend data for symbols
func (s *Server) providerFetchDividendsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logPrefix := s.providerLogPrefix()
	log.Printf("%s Starting bulk dividend fetch request", logPrefix)

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

	var symbols []string
	if request.All || len(request.Symbols) == 0 {
		var err error
		symbols, err = s.symbolService.GetPrioritizedSymbols()
		if err != nil {
			log.Printf("%s Error getting prioritized symbols: %v", logPrefix, err)
			http.Error(w, "Failed to get symbols", http.StatusInternalServerError)
			return
		}
	} else {
		symbols = request.Symbols
	}

	result, err := s.providerService.FetchDividendHistoryForSymbols(ctx, symbols, request.Limit)
	if result == nil {
		result = &providers.BulkDividendFetchResult{}
	}

	response := map[string]interface{}{
		"success":   err == nil,
		"processed": result.Processed,
		"results":   result.Results,
	}

	if len(result.Errors) > 0 {
		response["errors"] = result.Errors
	}

	totalDividends := 0
	for _, record := range result.Results {
		totalDividends += len(record.Dividends)
	}

	if err != nil {
		response["message"] = err.Error()
		log.Printf("%s Dividend fetch failed: %v", logPrefix, err)
	} else {
		response["message"] = fmt.Sprintf("Dividend fetch completed: %d symbols processed, %d total dividends found", result.Processed, totalDividends)
		log.Printf("%s Dividend fetch completed: %d symbols processed, %d total dividends found", logPrefix, result.Processed, totalDividends)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("%s Error encoding dividend fetch response: %v", logPrefix, err)
	}
}
