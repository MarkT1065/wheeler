package providers

import (
	"bytes"
	"context"
	"log"
	"path/filepath"
	"stonks/internal/database"
	"stonks/internal/models"
	"strings"
	"testing"
	"time"
)

// setupProvidersTestDB creates a temporary test database for provider tests
func setupProvidersTestDB(t *testing.T) *database.DB {
	testDBPath := filepath.Join(t.TempDir(), "providers_test.db")

	db, err := database.NewDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// TestGetAPIKeyStatus_Polygon tests API key status with Polygon provider
func TestGetAPIKeyStatus_Polygon(t *testing.T) {
	tests := []struct {
		name           string
		providerType   string
		apiKey         string
		wantConfigured bool
		wantMasked     string
	}{
		{
			name:           "configured with long key",
			providerType:   "polygon",
			apiKey:         "abcdefghijklmnop",
			wantConfigured: true,
			wantMasked:     "abc...nop",
		},
		{
			name:           "configured with short key",
			providerType:   "polygon",
			apiKey:         "abc",
			wantConfigured: true,
			wantMasked:     "***",
		},
		{
			name:           "not configured",
			providerType:   "polygon",
			apiKey:         "",
			wantConfigured: false,
			wantMasked:     "",
		},
		{
			name:           "default to polygon when type not set",
			providerType:   "",
			apiKey:         "testkey123",
			wantConfigured: true,
			wantMasked:     "tes...123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupProvidersTestDB(t)

			settingService := models.NewSettingService(db.DB)
			if tt.providerType != "" {
				if err := settingService.SetValue("DATA_PROVIDER_TYPE", tt.providerType, ""); err != nil {
					t.Fatalf("failed to set DATA_PROVIDER_TYPE: %v", err)
				}
			}
			if tt.apiKey != "" {
				if err := settingService.SetValue("POLYGON_API_KEY", tt.apiKey, ""); err != nil {
					t.Fatalf("failed to set POLYGON_API_KEY: %v", err)
				}
			}

			symbolService := models.NewSymbolService(db.DB)
			service := NewService(symbolService, settingService)

			status := service.GetAPIKeyStatus()

			if status.Configured != tt.wantConfigured {
				t.Errorf("Expected Configured=%v, got %v", tt.wantConfigured, status.Configured)
			}

			if status.Masked != tt.wantMasked {
				t.Errorf("Expected Masked=%q, got %q", tt.wantMasked, status.Masked)
			}
		})
	}
}

// TestGetAPIKeyStatus_UnsupportedProvider tests API key status with unsupported provider
func TestGetAPIKeyStatus_UnsupportedProvider(t *testing.T) {
	db := setupProvidersTestDB(t)

	settingService := models.NewSettingService(db.DB)
	if err := settingService.SetValue("DATA_PROVIDER_TYPE", "unknown_provider", ""); err != nil {
		t.Fatalf("failed to set DATA_PROVIDER_TYPE: %v", err)
	}
	if err := settingService.SetValue("POLYGON_API_KEY", "some_key", ""); err != nil {
		t.Fatalf("failed to set POLYGON_API_KEY: %v", err)
	}

	symbolService := models.NewSymbolService(db.DB)
	service := NewService(symbolService, settingService)

	status := service.GetAPIKeyStatus()

	if status.Configured {
		t.Error("Expected Configured=false for unsupported provider")
	}

	if status.Masked != "" {
		t.Errorf("Expected empty Masked for unsupported provider, got %q", status.Masked)
	}
}

func TestGetAPIKeyStatus_KeylessProvidersAreReadyWithoutAKey(t *testing.T) {
	for _, providerType := range []string{providerTypeGoogle, providerTypeYFinance} {
		t.Run(providerType, func(t *testing.T) {
			db := setupProvidersTestDB(t)
			settings := models.NewSettingService(db.DB)
			if err := settings.SetValue("DATA_PROVIDER_TYPE", providerType, ""); err != nil {
				t.Fatal(err)
			}
			service := NewService(models.NewSymbolService(db.DB), settings)
			status := service.GetAPIKeyStatus()
			if status.Configured || !status.Ready || status.RequiresAPIKey || status.Masked != "Not required" {
				t.Fatalf("unexpected keyless status: %+v", status)
			}
		})
	}
}

// TestUpdateSymbolPrice_MissingAPIKey tests error when API key is missing
func TestUpdateSymbolPrice_MissingAPIKey(t *testing.T) {
	ctx := context.Background()
	db := setupProvidersTestDB(t)

	settingService := models.NewSettingService(db.DB)
	if err := settingService.SetValue("DATA_PROVIDER_TYPE", "polygon", ""); err != nil {
		t.Fatalf("failed to set DATA_PROVIDER_TYPE: %v", err)
	}
	// Intentionally don't set POLYGON_API_KEY to test missing key scenario

	symbolService := models.NewSymbolService(db.DB)
	service := NewService(symbolService, settingService)

	err := service.UpdateSymbolPrice(ctx, "NASDAQ:AAPL")
	if err == nil {
		t.Fatal("Expected error when API key is missing, got nil")
	}

	if !strings.Contains(err.Error(), "failed to get provider") {
		t.Errorf("Expected error about provider, got: %v", err)
	}
}

// TestUpdateSymbolPrice_UnsupportedProvider tests error with unsupported provider
func TestUpdateSymbolPrice_UnsupportedProvider(t *testing.T) {
	ctx := context.Background()
	db := setupProvidersTestDB(t)

	settingService := models.NewSettingService(db.DB)
	if err := settingService.SetValue("DATA_PROVIDER_TYPE", "unsupported", ""); err != nil {
		t.Fatalf("failed to set DATA_PROVIDER_TYPE: %v", err)
	}
	if err := settingService.SetValue("POLYGON_API_KEY", "test_key", ""); err != nil {
		t.Fatalf("failed to set POLYGON_API_KEY: %v", err)
	}

	symbolService := models.NewSymbolService(db.DB)
	service := NewService(symbolService, settingService)

	err := service.UpdateSymbolPrice(ctx, "NASDAQ:AAPL")
	if err == nil {
		t.Fatal("Expected error when provider is unsupported, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported data provider type") {
		t.Errorf("Expected error about provider, got: %v", err)
	}
}

// TestUpdateAllSymbolPrices_MissingAPIKey tests bulk update without API key
func TestUpdateAllSymbolPrices_MissingAPIKey(t *testing.T) {
	ctx := context.Background()
	db := setupProvidersTestDB(t)

	settingService := models.NewSettingService(db.DB)
	if err := settingService.SetValue("DATA_PROVIDER_TYPE", "polygon", ""); err != nil {
		t.Fatalf("failed to set DATA_PROVIDER_TYPE: %v", err)
	}

	// No API key
	symbolService := models.NewSymbolService(db.DB)

	// Create test symbols
	_, err := symbolService.Create("NASDAQ:AAPL")
	if err != nil {
		t.Fatalf("failed to create symbol AAPL: %v", err)
	}
	_, err = symbolService.Create("NASDAQ:GOOGL")
	if err != nil {
		t.Fatalf("failed to create symbol GOOGL: %v", err)
	}

	service := NewService(symbolService, settingService)

	_, err = service.UpdateAllSymbolPrices(ctx)
	if err == nil {
		t.Fatal("Expected error without API key, got nil")
	}
}

// TestFetchSymbolDetails_MissingAPIKey tests fetch details without API key
func TestFetchSymbolDetails_MissingAPIKey(t *testing.T) {
	ctx := context.Background()
	db := setupProvidersTestDB(t)

	settingService := models.NewSettingService(db.DB)
	if err := settingService.SetValue("DATA_PROVIDER_TYPE", "polygon", ""); err != nil {
		t.Fatalf("failed to set DATA_PROVIDER_TYPE: %v", err)
	}
	// No API key - should fail

	symbolService := models.NewSymbolService(db.DB)
	service := NewService(symbolService, settingService)

	_, err := service.FetchSymbolDetails(ctx, "NASDAQ:AAPL")
	if err == nil {
		t.Fatal("Expected error without API key, got nil")
	}

	if !strings.Contains(err.Error(), "failed to get provider") {
		t.Errorf("Expected provider error, got: %v", err)
	}
}

// TestFetchDividendHistory_MissingAPIKey tests dividend history without API key
func TestFetchDividendHistory_MissingAPIKey(t *testing.T) {
	ctx := context.Background()
	db := setupProvidersTestDB(t)

	settingService := models.NewSettingService(db.DB)
	if err := settingService.SetValue("DATA_PROVIDER_TYPE", "polygon", ""); err != nil {
		t.Fatalf("failed to set DATA_PROVIDER_TYPE: %v", err)
	}
	// No API key

	symbolService := models.NewSymbolService(db.DB)
	service := NewService(symbolService, settingService)

	_, err := service.FetchDividendHistory(ctx, "NASDAQ:AAPL", 10)
	if err == nil {
		t.Fatal("Expected error without API key, got nil")
	}

	if !strings.Contains(err.Error(), "failed to get provider") {
		t.Errorf("Expected provider error, got: %v", err)
	}
}

// TestTestConnection_MissingAPIKey tests connection test without API key
func TestTestConnection_MissingAPIKey(t *testing.T) {
	ctx := context.Background()
	db := setupProvidersTestDB(t)

	settingService := models.NewSettingService(db.DB)
	if err := settingService.SetValue("DATA_PROVIDER_TYPE", "polygon", ""); err != nil {
		t.Fatalf("failed to set DATA_PROVIDER_TYPE: %v", err)
	}
	// No API key

	symbolService := models.NewSymbolService(db.DB)
	service := NewService(symbolService, settingService)

	err := service.TestConnection(ctx)
	if err == nil {
		t.Fatal("Expected error without API key, got nil")
	}
}

// TestGetRateLimitDelay tests rate limit delay configuration
func TestGetRateLimitDelay(t *testing.T) {
	tests := []struct {
		name         string
		providerType string
		wantDelay    time.Duration
	}{
		{
			name:         "polygon provider",
			providerType: "polygon",
			wantDelay:    12 * time.Second,
		},
		{
			name:         "default provider",
			providerType: "",
			wantDelay:    12 * time.Second, // Defaults to polygon
		},
		{
			name:         "unknown provider",
			providerType: "unknown",
			wantDelay:    10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupProvidersTestDB(t)

			settingService := models.NewSettingService(db.DB)
			if tt.providerType != "" {
				if err := settingService.SetValue("DATA_PROVIDER_TYPE", tt.providerType, ""); err != nil {
					t.Fatalf("failed to set DATA_PROVIDER_TYPE: %v", err)
				}
			}

			symbolService := models.NewSymbolService(db.DB)
			service := NewService(symbolService, settingService)

			delay := service.GetRateLimitDelay()
			if delay != tt.wantDelay {
				t.Errorf("Expected delay %v, got %v", tt.wantDelay, delay)
			}
		})
	}
}

// TestDebugLogging exercises the provider-selection branch and verifies that
// debug output masks, rather than exposes, the configured API key.
func TestDebugLogging(t *testing.T) {
	db := setupProvidersTestDB(t)

	settingService := models.NewSettingService(db.DB)
	if err := settingService.SetValue("DATA_PROVIDER_TYPE", "polygon", ""); err != nil {
		t.Fatalf("failed to set DATA_PROVIDER_TYPE: %v", err)
	}
	if err := settingService.SetValue("POLYGON_API_KEY", "test_key_12345", ""); err != nil {
		t.Fatalf("failed to set POLYGON_API_KEY: %v", err)
	}

	symbolService := models.NewSymbolService(db.DB)
	service := NewService(symbolService, settingService)
	t.Setenv("DEBUG_PROVIDERS", "1")

	var logs bytes.Buffer
	previousOutput := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(previousOutput) })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := service.TestConnection(ctx); err == nil {
		t.Fatal("Expected cancelled connection test to fail")
	}

	if !strings.Contains(logs.String(), "Using POLYGON provider with API key: tes...345") {
		t.Fatalf("Expected masked provider debug log, got %q", logs.String())
	}
}
