package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"stonks/internal/markets"
	"stonks/internal/models"
	"stonks/internal/providers"
	"strings"
	"unicode"
)

// SettingsData holds data for the settings template
type SettingsData struct {
	Settings         []*models.Setting           `json:"settings"`
	AllSymbols       []string                    `json:"allSymbols"`
	LegacySymbols    []string                    `json:"legacySymbols"`
	CurrentDB        string                      `json:"currentDB"`
	ApiKey           string                      `json:"apiKey"`
	ActiveProvider   string                      `json:"activeProvider"`
	ProviderName     string                      `json:"providerName"`
	ProviderProfiles []providers.ProviderProfile `json:"providerProfiles"`
	ActivePage       string                      `json:"activePage"`
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
	legacySymbols := make([]string, 0)
	for _, symbol := range symbols {
		if markets.IsLegacyKey(symbol) {
			legacySymbols = append(legacySymbols, symbol)
		}
	}

	// Get all settings
	settings, err := s.settingService.GetAll()
	if err != nil {
		log.Printf("[SETTINGS] Error getting settings: %v", err)
		settings = []*models.Setting{}
	}

	profile := s.providerService.ActiveProfile()
	providerName := profile.DisplayName
	if providerName == "" {
		providerName = formatProviderDisplayName(s.providerService.Name())
	}

	apiKey := s.settingService.GetValue("POLYGON_API_KEY")
	data := SettingsData{
		Settings:         settings,
		AllSymbols:       symbols,
		LegacySymbols:    legacySymbols,
		CurrentDB:        s.getCurrentDatabaseName(),
		ApiKey:           apiKey,
		ActiveProvider:   s.providerService.Name(),
		ProviderName:     providerName,
		ProviderProfiles: providers.ProviderProfiles(),
		ActivePage:       "settings",
	}

	s.renderTemplate(w, "settings.html", data)
}

func formatProviderDisplayName(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "Market Data Provider"
	}
	lower := strings.ToLower(trimmed)
	switch lower {
	case "polygon":
		return "Polygon.io"
	case "google_finance":
		return "Google Finance"
	case "yfinance":
		return "Yahoo Finance (yfinance)"
	default:
		runes := []rune(trimmed)
		runes[0] = unicode.ToUpper(runes[0])
		return string(runes)
	}
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

// ProviderConfigurationRequest groups settings that must change together when
// selecting a market-data provider. Keeping this separate from the generic
// settings CRUD API makes a provider switch atomic.
type ProviderConfigurationRequest struct {
	ProviderType  string `json:"provider_type"`
	PolygonAPIKey string `json:"polygon_api_key"`
}

func (s *Server) providerConfigurationAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ProviderConfigurationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	providerType, ok := providers.CanonicalProviderType(req.ProviderType)
	if !ok {
		http.Error(w, fmt.Sprintf("unsupported data provider type: %s", strings.TrimSpace(req.ProviderType)), http.StatusBadRequest)
		return
	}
	profile := providers.ProviderProfile{}
	foundProfile := false
	for _, candidate := range providers.ProviderProfiles() {
		if candidate.ID == providerType {
			profile = candidate
			foundProfile = true
			break
		}
	}
	if !foundProfile {
		http.Error(w, fmt.Sprintf("unsupported data provider type: %s", providerType), http.StatusBadRequest)
		return
	}

	apiKey := strings.TrimSpace(req.PolygonAPIKey)
	if profile.RequiresAPIKey && apiKey == "" {
		http.Error(w, "Polygon API key is required when Polygon.io is selected", http.StatusBadRequest)
		return
	}
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		log.Printf("[SETTINGS API] Error starting provider configuration transaction: %v", err)
		http.Error(w, "Failed to save provider configuration", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	upsert := func(name, value, description string) error {
		_, err := tx.Exec(`INSERT INTO settings (name, value, description) VALUES (?, ?, ?)
			ON CONFLICT(name) DO UPDATE SET value = excluded.value, description = excluded.description, updated_at = CURRENT_TIMESTAMP`, name, value, description)
		return err
	}
	if profile.RequiresAPIKey {
		if err := upsert("POLYGON_API_KEY", apiKey, "API key for Polygon.io stock market data integration"); err != nil {
			log.Printf("[SETTINGS API] Error saving Polygon API key: %v", err)
			http.Error(w, "Failed to save provider configuration", http.StatusInternalServerError)
			return
		}
	}
	// Write the active provider last. The transaction makes every change
	// visible together only after all dependent configuration has succeeded.
	if err := upsert("DATA_PROVIDER_TYPE", providerType, "Active market data provider"); err != nil {
		log.Printf("[SETTINGS API] Error saving provider type: %v", err)
		http.Error(w, "Failed to save provider configuration", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(); err != nil {
		log.Printf("[SETTINGS API] Error committing provider configuration: %v", err)
		http.Error(w, "Failed to save provider configuration", http.StatusInternalServerError)
		return
	}

	s.invalidateProviderValidation()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"provider": providerType})
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
	if err := normalizeProviderSetting(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	s.invalidateProviderValidation()

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
	req.Name = name
	if err := normalizeProviderSetting(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	s.invalidateProviderValidation()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(setting); err != nil {
		log.Printf("[SETTINGS API] Error encoding update response: %v", err)
	}
}

func normalizeProviderSetting(req *SettingRequest) error {
	name := strings.ToUpper(strings.TrimSpace(req.Name))
	switch name {
	case "DATA_PROVIDER_TYPE":
		providerType, ok := providers.CanonicalProviderType(req.Value)
		if !ok {
			return fmt.Errorf("unsupported data provider type: %s", strings.TrimSpace(req.Value))
		}
		req.Value = providerType
	case "GOOGLE_FINANCE_EXCHANGE":
		req.Value = strings.ToUpper(strings.TrimSpace(req.Value))
		if req.Value == "" {
			req.Value = "NASDAQ"
		}
	}
	return nil
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
	s.invalidateProviderValidation()

	w.WriteHeader(http.StatusNoContent)
}
