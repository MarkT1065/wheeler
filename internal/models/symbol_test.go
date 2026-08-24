package models

import (
	"path/filepath"
	"stonks/internal/database"
	"testing"
	"time"
)

func TestSymbolServiceQualifyLegacySymbolMovesDependencies(t *testing.T) {
	db, err := database.NewDB(filepath.Join(t.TempDir(), "symbols.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.DB.Exec(`INSERT INTO symbols (symbol, price) VALUES ('PETR4', 40.50)`); err != nil {
		t.Fatal(err)
	}
	optionService := NewOptionService(db.DB)
	positionService := NewLongPositionService(db.DB)
	dividendService := NewDividendService(db.DB)
	opened := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	if _, err := optionService.Create("PETR4", "Put", opened, 38, opened.AddDate(0, 1, 0), 1, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := positionService.Create("PETR4", opened, 100, 39); err != nil {
		t.Fatal(err)
	}
	if _, err := dividendService.Create("PETR4", opened, 0.25); err != nil {
		t.Fatal(err)
	}

	service := NewSymbolService(db.DB)
	symbol, err := service.QualifyLegacySymbol("PETR4", "BVMF")
	if err != nil {
		t.Fatal(err)
	}
	if symbol.Symbol != "BVMF:PETR4" {
		t.Fatalf("symbol = %q", symbol.Symbol)
	}
	if _, err := service.GetBySymbol("PETR4"); err == nil {
		t.Fatal("legacy symbol still exists")
	}
	if options, _ := optionService.GetBySymbol("BVMF:PETR4"); len(options) != 1 {
		t.Fatalf("options not moved: %d", len(options))
	}
	if positions, _ := positionService.GetBySymbol("BVMF:PETR4"); len(positions) != 1 {
		t.Fatalf("positions not moved: %d", len(positions))
	}
	if dividends, _ := dividendService.GetBySymbol("BVMF:PETR4"); len(dividends) != 1 {
		t.Fatalf("dividends not moved: %d", len(dividends))
	}
}

func TestSymbolServiceCreateRejectsLegacySymbol(t *testing.T) {
	db, err := database.NewDB(filepath.Join(t.TempDir(), "symbols.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := NewSymbolService(db.DB).Create("AAPL"); err == nil {
		t.Fatal("expected bare ticker to be rejected")
	}
}
