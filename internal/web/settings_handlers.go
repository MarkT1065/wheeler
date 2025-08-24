package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"stonks/internal/models"
)

// SettingsData holds data for the settings template
type SettingsData struct {
	Settings   []*models.Setting `json:"settings"`
	AllSymbols []string          `json:"allSymbols"`
	CurrentDB  string            `json:"currentDB"`
	ApiKey     string            `json:"apiKey"`
}

// settingsHandler serves the settings management page
func (s *Server) settingsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[SETTINGS] Handling settings page request")

	// Get all symbols for navigation
	symbols, err := s.symbolService.GetDistinctSymbols()
	if err != nil {
		log.Printf("[SETTINGS] Error getting symbols: %v", err)
		symbols = []string{}
	}

	// Get all settings
	settings, err := s.settingService.GetAll()
	if err != nil {
		log.Printf("[SETTINGS] Error getting settings: %v", err)
		settings = []*models.Setting{}
	}

	// Get API key value specifically
	apiKey := s.settingService.GetValue("POLYGON_API_KEY")

	data := SettingsData{
		Settings:   settings,
		AllSymbols: symbols,
		CurrentDB:  s.getCurrentDatabaseName(),
		ApiKey:     apiKey,
	}

	s.renderTemplate(w, "settings.html", data)
}

// settingsAPIHandler handles CRUD operations for settings collection
func (s *Server) settingsAPIHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[SETTINGS API] %s request to /api/settings", r.Method)

	switch r.Method {
	case http.MethodGet:
		s.getAllSettingsAPI(w, r)
	case http.MethodPost:
		s.createSettingAPI(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// individualSettingAPIHandler handles operations on individual settings
func (s *Server) individualSettingAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Extract setting name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/settings/")
	if path == "" {
		http.Error(w, "Setting name required", http.StatusBadRequest)
		return
	}

	settingName := strings.ToUpper(path)
	log.Printf("[SETTINGS API] %s request for setting: %s", r.Method, settingName)

	switch r.Method {
	case http.MethodGet:
		s.getSettingAPI(w, r, settingName)
	case http.MethodPut:
		s.updateSettingAPI(w, r, settingName)
	case http.MethodDelete:
		s.deleteSettingAPI(w, r, settingName)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getAllSettingsAPI returns all settings as JSON
func (s *Server) getAllSettingsAPI(w http.ResponseWriter, r *http.Request) {
	settings, err := s.settingService.GetAll()
	if err != nil {
		log.Printf("[SETTINGS API] Error getting all settings: %v", err)
		http.Error(w, "Failed to get settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(settings); err != nil {
		log.Printf("[SETTINGS API] Error encoding settings response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// getSettingAPI returns a specific setting as JSON
func (s *Server) getSettingAPI(w http.ResponseWriter, r *http.Request, name string) {
	setting, err := s.settingService.GetByName(name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Setting not found", http.StatusNotFound)
			return
		}
		log.Printf("[SETTINGS API] Error getting setting %s: %v", name, err)
		http.Error(w, "Failed to get setting", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(setting); err != nil {
		log.Printf("[SETTINGS API] Error encoding setting response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// SettingRequest represents the request body for creating/updating settings
type SettingRequest struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

// createSettingAPI creates a new setting
func (s *Server) createSettingAPI(w http.ResponseWriter, r *http.Request) {
	var req SettingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[SETTINGS API] Error decoding create request: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "Setting name is required", http.StatusBadRequest)
		return
	}

	// Create the setting
	setting, err := s.settingService.Create(req.Name, req.Value, req.Description)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, fmt.Sprintf("Setting '%s' already exists", req.Name), http.StatusConflict)
			return
		}
		log.Printf("[SETTINGS API] Error creating setting: %v", err)
		http.Error(w, "Failed to create setting", http.StatusInternalServerError)
		return
	}

	log.Printf("[SETTINGS API] Successfully created setting: %s", setting.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(setting); err != nil {
		log.Printf("[SETTINGS API] Error encoding create response: %v", err)
	}
}

// updateSettingAPI updates an existing setting
func (s *Server) updateSettingAPI(w http.ResponseWriter, r *http.Request, name string) {
	var req SettingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[SETTINGS API] Error decoding update request: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Update the setting
	setting, err := s.settingService.Update(name, req.Value, req.Description)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Setting not found", http.StatusNotFound)
			return
		}
		log.Printf("[SETTINGS API] Error updating setting %s: %v", name, err)
		http.Error(w, "Failed to update setting", http.StatusInternalServerError)
		return
	}

	log.Printf("[SETTINGS API] Successfully updated setting: %s", setting.Name)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(setting); err != nil {
		log.Printf("[SETTINGS API] Error encoding update response: %v", err)
	}
}

// deleteSettingAPI deletes a setting
func (s *Server) deleteSettingAPI(w http.ResponseWriter, r *http.Request, name string) {
	err := s.settingService.Delete(name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Setting not found", http.StatusNotFound)
			return
		}
		log.Printf("[SETTINGS API] Error deleting setting %s: %v", name, err)
		http.Error(w, "Failed to delete setting", http.StatusInternalServerError)
		return
	}

	log.Printf("[SETTINGS API] Successfully deleted setting: %s", name)

	w.WriteHeader(http.StatusNoContent)
}