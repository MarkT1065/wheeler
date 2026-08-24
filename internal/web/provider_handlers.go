package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"stonks/internal/markets"
	"strings"
	"time"

	"stonks/internal/providers"
)

func validateSymbols(symbols []string) ([]string, error) {
	validated := make([]string, 0, len(symbols))
	for _, raw := range symbols {
		instrument, err := markets.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid symbol %q: %w", raw, err)
		}
		validated = append(validated, instrument.Key())
	}
	return validated, nil
}

func providerLogPrefix(providerService *providers.Service) string {
	name := strings.ToUpper(strings.TrimSpace(providerService.Name()))
	if name == "" {
		name = "PROVIDER"
	}
	return fmt.Sprintf("[%s API]", name)
}

func (s *Server) providerServicesForRequest(r *http.Request) providerServicesSnapshot {
	if snapshot, ok := providerServicesFromContext(r.Context()); ok {
		return snapshot
	}
	return providerServicesSnapshot{providerService: s.providerService, symbolService: s.symbolService}
}

// providerTestHandler tests the data provider API connection
func (s *Server) providerTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	providerName := s.providerService.Name()
	logPrefix := providerLogPrefix(s.providerService)
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
		response["message"] = "Provider connection successful"
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

	services := s.providerServicesForRequest(r)
	providerService := services.providerService
	logPrefix := providerLogPrefix(providerService)
	log.Printf("%s Starting price update request", logPrefix)

	// Parse request body to get specific symbols (optional)
	var request struct {
		Symbols []string `json:"symbols,omitempty"`
		All     bool     `json:"all,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err == io.EOF {
		request.All = true
	} else if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var (
		result *providers.BulkPriceUpdateResult
		err    error
	)

	if request.All || len(request.Symbols) == 0 {
		result, err = providerService.UpdateAllSymbolPrices(ctx)
	} else {
		validated, validateErr := validateSymbols(request.Symbols)
		if validateErr != nil {
			http.Error(w, validateErr.Error(), http.StatusBadRequest)
			return
		}
		result, err = providerService.UpdateSymbolPrices(ctx, validated)
	}

	if result == nil {
		result = &providers.BulkPriceUpdateResult{}
	}

	success := err == nil && result.Updated > 0
	response := map[string]interface{}{
		"success": success,
		"updated": result.Updated,
		"failed":  result.Failed,
		"skipped": result.Skipped,
		"limit":   providers.BulkPriceUpdateLimit(),
	}

	if len(result.Errors) > 0 {
		response["errors"] = result.Errors
	}

	if err != nil {
		response["message"] = err.Error()
		log.Printf("%s Price update failed: %v", logPrefix, err)
	} else if success {
		response["message"] = "Price update completed"
		if result.Skipped > 0 {
			response["message"] = fmt.Sprintf("Price update completed; %d symbols remain for a later request", result.Skipped)
		}
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

	instrument, err := markets.Parse(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	symbol := instrument.Key()
	logPrefix := providerLogPrefix(s.providerService)
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

	var dividends []*providers.DividendInfo
	if s.providerService.ActiveProfile().SupportsDividends {
		dividends, err = s.providerService.FetchDividendHistory(ctx, symbol, 5)
		if err != nil {
			log.Printf("%s Warning: failed to get dividend history for %s: %v", logPrefix, symbol, err)
		}
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
	logPrefix := providerLogPrefix(s.providerService)

	// Keyless providers are ready to test without an API key.
	if status.Ready {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		status.Valid, status.Error = s.cachedProviderValidation(status.Provider, func() error {
			return s.providerService.TestConnection(ctx)
		})
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

	services := s.providerServicesForRequest(r)
	providerService := services.providerService
	symbolService := services.symbolService
	logPrefix := providerLogPrefix(providerService)
	log.Printf("%s Starting bulk dividend fetch request", logPrefix)
	if !providerService.ActiveProfile().SupportsDividends {
		http.Error(w, "The configured provider does not support dividend history", http.StatusConflict)
		return
	}

	// Parse request body to get specific symbols (optional)
	var request struct {
		Symbols []string `json:"symbols,omitempty"`
		All     bool     `json:"all,omitempty"`
		Limit   int      `json:"limit,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err == io.EOF {
		request.All = true
	} else if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if request.Limit == 0 {
		request.Limit = 10 // Default to 10 recent dividends
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var symbols []string
	if request.All || len(request.Symbols) == 0 {
		var err error
		symbols, err = symbolService.GetPrioritizedSymbols()
		if err != nil {
			log.Printf("%s Error getting prioritized symbols: %v", logPrefix, err)
			http.Error(w, "Failed to get symbols", http.StatusInternalServerError)
			return
		}
	} else {
		validated, validateErr := validateSymbols(request.Symbols)
		if validateErr != nil {
			http.Error(w, validateErr.Error(), http.StatusBadRequest)
			return
		}
		symbols = validated
	}

	result, err := providerService.FetchDividendHistoryForSymbols(ctx, symbols, request.Limit)
	if result == nil {
		result = &providers.BulkDividendFetchResult{Results: []providers.DividendFetchRecord{}}
	}

	response := map[string]interface{}{
		"success":   err == nil,
		"processed": result.Processed,
		"skipped":   result.Skipped,
		"limit":     providers.BulkPriceUpdateLimit(),
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
		if result.Skipped > 0 {
			response["message"] = fmt.Sprintf("Dividend fetch completed: %d symbols processed, %d total dividends found; %d symbols remain for a later request", result.Processed, totalDividends, result.Skipped)
		}
		log.Printf("%s Dividend fetch completed: %d symbols processed, %d total dividends found", logPrefix, result.Processed, totalDividends)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("%s Error encoding dividend fetch response: %v", logPrefix, err)
	}
}
