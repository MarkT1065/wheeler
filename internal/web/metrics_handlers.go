package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"stonks/internal/models"
	"strconv"
	"strings"
)

// metricsHandler serves the metrics view
func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[METRICS PAGE] %s %s - Start processing metrics page request", r.Method, r.URL.Path)

	symbols, err := s.symbolService.GetDistinctSymbols()
	if err != nil {
		log.Printf("[METRICS PAGE] WARNING: Failed to get symbols for navigation: %v", err)
		symbols = []string{}
	} else {
		log.Printf("[METRICS PAGE] Retrieved %d symbols for navigation", len(symbols))
	}

	// Get all metrics
	log.Printf("[METRICS PAGE] Fetching all metrics")
	metrics, err := s.metricService.GetAll()
	if err != nil {
		log.Printf("[METRICS PAGE] ERROR: Failed to get metrics: %v", err)
		metrics = []*models.Metric{}
	} else {
		log.Printf("[METRICS PAGE] Retrieved %d metrics", len(metrics))
	}

	data := MetricsData{
		PageTitle:  "Metrics",
		Symbols:    symbols,
		AllSymbols: symbols, // For navigation compatibility
		Metrics:    metrics,
		CurrentDB:  s.getCurrentDatabaseName(),
		ActivePage: "metrics",
	}

	log.Printf("[METRICS PAGE] Rendering template with %d metrics", len(metrics))
	if err := s.templates.ExecuteTemplate(w, "metrics.html", data); err != nil {
		log.Printf("[METRICS PAGE] ERROR: Failed to execute template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("[METRICS PAGE] %s %s - Successfully served metrics page", r.Method, r.URL.Path)
}

// API Handlers

// createMetricHandler handles POST /api/metrics
func (s *Server) createMetricHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[API] POST /api/metrics - Start creating metric")

	if r.Method != http.MethodPost {
		log.Printf("[API] POST /api/metrics - Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Type  string  `json:"type"`
		Value float64 `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[API] POST /api/metrics - Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[API] POST /api/metrics - Creating metric: type=%s, value=%.2f", req.Type, req.Value)

	metric, err := s.metricService.Create(models.MetricType(req.Type), req.Value)
	if err != nil {
		log.Printf("[API] POST /api/metrics - Failed to create metric: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create metric: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metric); err != nil {
		log.Printf("[API] POST /api/metrics - Failed to encode response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("[API] POST /api/metrics - Successfully created metric with ID %d", metric.ID)
}

// getMetricsHandler handles GET /api/metrics
func (s *Server) getMetricsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[API] GET /api/metrics - Start fetching metrics")

	if r.Method != http.MethodGet {
		log.Printf("[API] GET /api/metrics - Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics, err := s.metricService.GetAll()
	if err != nil {
		log.Printf("[API] GET /api/metrics - Failed to get metrics: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get metrics: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[API] GET /api/metrics - Retrieved %d metrics", len(metrics))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		log.Printf("[API] GET /api/metrics - Failed to encode response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("[API] GET /api/metrics - Successfully returned %d metrics", len(metrics))
}

// updateMetricHandler handles PUT /api/metrics/{id}
func (s *Server) updateMetricHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[API] PUT /api/metrics/{id} - Start updating metric")

	if r.Method != http.MethodPut {
		log.Printf("[API] PUT /api/metrics/{id} - Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[3] == "" {
		log.Printf("[API] PUT /api/metrics/{id} - Invalid URL path: %s", path)
		http.Error(w, "Invalid metric ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[3])
	if err != nil {
		log.Printf("[API] PUT /api/metrics/{id} - Invalid metric ID: %s", parts[3])
		http.Error(w, "Invalid metric ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Value float64 `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[API] PUT /api/metrics/{id} - Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[API] PUT /api/metrics/{id} - Updating metric ID %d with value %.2f", id, req.Value)

	metric, err := s.metricService.Update(id, req.Value)
	if err != nil {
		log.Printf("[API] PUT /api/metrics/{id} - Failed to update metric: %v", err)
		http.Error(w, fmt.Sprintf("Failed to update metric: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metric); err != nil {
		log.Printf("[API] PUT /api/metrics/{id} - Failed to encode response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("[API] PUT /api/metrics/{id} - Successfully updated metric ID %d", id)
}

// deleteMetricHandler handles DELETE /api/metrics/{id}
func (s *Server) deleteMetricHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[API] DELETE /api/metrics/{id} - Start deleting metric")

	if r.Method != http.MethodDelete {
		log.Printf("[API] DELETE /api/metrics/{id} - Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[3] == "" {
		log.Printf("[API] DELETE /api/metrics/{id} - Invalid URL path: %s", path)
		http.Error(w, "Invalid metric ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[3])
	if err != nil {
		log.Printf("[API] DELETE /api/metrics/{id} - Invalid metric ID: %s", parts[3])
		http.Error(w, "Invalid metric ID", http.StatusBadRequest)
		return
	}

	log.Printf("[API] DELETE /api/metrics/{id} - Deleting metric ID %d", id)

	if err := s.metricService.Delete(id); err != nil {
		log.Printf("[API] DELETE /api/metrics/{id} - Failed to delete metric: %v", err)
		http.Error(w, fmt.Sprintf("Failed to delete metric: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Printf("[API] DELETE /api/metrics/{id} - Successfully deleted metric ID %d", id)
}

// createMetricsSnapshotHandler handles POST /api/metrics/snapshot
func (s *Server) createMetricsSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[API] POST /api/metrics/snapshot - Start creating comprehensive metrics snapshot")

	if r.Method != http.MethodPost {
		log.Printf("[API] POST /api/metrics/snapshot - Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request to get days parameter (optional, defaults to 1 for current day)
	var req struct {
		Days int `json:"days,omitempty"`
	}
	
	// Try to decode request body, but don't fail if it's empty
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[API] POST /api/metrics/snapshot - Failed to decode request body (using defaults): %v", err)
		}
	}

	// Default to 1 day (today only) if not specified
	days := req.Days
	if days <= 0 {
		days = 1
	}

	log.Printf("[API] POST /api/metrics/snapshot - Creating comprehensive snapshot for %d days", days)

	// Use the new ComprehensiveSnapshot function
	err := s.metricService.ComprehensiveSnapshot(days)
	if err != nil {
		log.Printf("[API] POST /api/metrics/snapshot - Failed to create comprehensive snapshot: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create metrics snapshot: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Comprehensive metrics snapshot created successfully for %d days", days),
		"days":    days,
	}); err != nil {
		log.Printf("[API] POST /api/metrics/snapshot - Failed to encode response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("[API] POST /api/metrics/snapshot - Successfully created comprehensive snapshot for %d days", days)
}

// getMetricsChartDataHandler handles GET /api/metrics/chart-data
func (s *Server) getMetricsChartDataHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[API] GET /api/metrics/chart-data - Start fetching chart data")

	if r.Method != http.MethodGet {
		log.Printf("[API] GET /api/metrics/chart-data - Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all metrics ordered by date and type
	query := `
		SELECT DATE(created) as date, type, value 
		FROM metrics 
		ORDER BY created ASC, type ASC
	`
	
	rows, err := s.db.Query(query)
	if err != nil {
		log.Printf("[API] GET /api/metrics/chart-data - Failed to query metrics: %v", err)
		http.Error(w, fmt.Sprintf("Failed to query metrics: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Build chart data structure
	chartData := make(map[string][]ChartPoint)
	
	for rows.Next() {
		var date, metricType string
		var value float64
		
		if err := rows.Scan(&date, &metricType, &value); err != nil {
			log.Printf("[API] GET /api/metrics/chart-data - Failed to scan row: %v", err)
			continue
		}
		
		chartData[metricType] = append(chartData[metricType], ChartPoint{
			Date:  date,
			Value: value,
		})
	}

	log.Printf("[API] GET /api/metrics/chart-data - Retrieved chart data for %d metric types", len(chartData))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(chartData); err != nil {
		log.Printf("[API] GET /api/metrics/chart-data - Failed to encode response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("[API] GET /api/metrics/chart-data - Successfully returned chart data")
}

