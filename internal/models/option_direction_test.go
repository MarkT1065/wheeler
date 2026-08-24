package models

import (
	"stonks/internal/database"
	"testing"
	"time"
)

func TestOptionDirectionPersistsAndChangesProfitSign(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec("INSERT INTO symbols (symbol) VALUES ('NASDAQ:AAPL')"); err != nil {
		t.Fatal(err)
	}

	service := NewOptionService(db.DB)
	opened := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	expires := opened.AddDate(0, 1, 0)
	short, err := service.CreateWithCommissionAndDirection("NASDAQ:AAPL", "Put", "Short", opened, 100, expires, 2, 1, 0.65)
	if err != nil {
		t.Fatal(err)
	}
	long, err := service.CreateWithCommissionAndDirection("NASDAQ:AAPL", "Put", "Long", opened, 100, expires, 2, 1, 0.65)
	if err != nil {
		t.Fatal(err)
	}
	if short.Direction != "Short" || long.Direction != "Long" {
		t.Fatalf("directions = %q, %q", short.Direction, long.Direction)
	}
	if short.CalculateTotalProfit() != 199.35 || long.CalculateTotalProfit() != -200.65 {
		t.Fatalf("opening P&L = %.2f, %.2f", short.CalculateTotalProfit(), long.CalculateTotalProfit())
	}
	exit := 1.25
	closed := opened.AddDate(0, 0, 7)
	if _, err := service.UpdateByIDWithDirection(long.ID, long.Symbol, long.Type, long.Direction, long.Opened, long.Strike, long.Expiration, long.Premium, long.Contracts, long.Commission, &closed, &exit); err != nil {
		t.Fatal(err)
	}
	updated, err := service.GetByID(long.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.CalculateTotalProfit() != -75.65 {
		t.Fatalf("long closing P&L = %.2f", updated.CalculateTotalProfit())
	}
	if updated.HasPerformanceMetrics() {
		t.Fatal("long options must not expose short-option performance metrics")
	}
}

func TestOptionDirectionRejectsInvalidValue(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	service := NewOptionService(db.DB)
	_, err = service.CreateWithCommissionAndDirection("NASDAQ:AAPL", "Put", "Diagonal", time.Now(), 100, time.Now().AddDate(0, 1, 0), 1, 1, 0)
	if err == nil {
		t.Fatal("expected invalid direction error")
	}

	if err := service.CloseWithDirection("NASDAQ:AAPL", "Put", "Diagonal", time.Now(), 100, time.Now().AddDate(0, 1, 0), 1, 1, time.Now(), 0); err == nil {
		t.Fatal("expected CloseWithDirection to reject an invalid direction")
	}
}

func TestOptionDirectionTreatsUnknownValuesAsNonShort(t *testing.T) {
	if (&Option{Direction: ""}).IsShort() != true {
		t.Fatal("legacy empty direction must remain short")
	}
	if (&Option{Direction: "Diagonal"}).IsShort() {
		t.Fatal("unknown direction must not be treated as short")
	}
}

func TestDeleteWithDirectionKeepsOppositeSpreadLeg(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec("INSERT INTO symbols (symbol) VALUES ('NASDAQ:SPREAD')"); err != nil {
		t.Fatal(err)
	}
	service := NewOptionService(db.DB)
	opened := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	expires := opened.AddDate(0, 1, 0)
	for _, direction := range []string{"Short", "Long"} {
		if _, err := service.CreateWithCommissionAndDirection("NASDAQ:SPREAD", "Put", direction, opened, 100, expires, 2, 1, 0); err != nil {
			t.Fatal(err)
		}
	}
	if err := service.DeleteWithDirection("NASDAQ:SPREAD", "Put", "Short", opened, 100, expires, 2, 1); err != nil {
		t.Fatal(err)
	}
	options, err := service.GetBySymbol("NASDAQ:SPREAD")
	if err != nil {
		t.Fatal(err)
	}
	if len(options) != 1 || options[0].Direction != "Long" {
		t.Fatalf("remaining option = %#v", options)
	}
}
