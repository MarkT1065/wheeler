package web

import (
	"context"
	"encoding/json"
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

func (s *Server) polygonStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := s.polygonService.GetAPIKeyStatus()

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

func (s *Server) polygonWorkerStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := s.polygonService.GetWorkerStatus()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Printf("[POLYGON API] Error encoding worker status response: %v", err)
	}
}
