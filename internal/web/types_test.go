package web

import (
	"math"
	"testing"
	"time"

	"stonks/internal/models"
)

func TestCalculateAnnualizedCashOnCash(t *testing.T) {
	if got := CalculateAnnualizedCashOnCash(100, 10000); math.Abs(got-365) > 0.0001 {
		t.Fatalf("CalculateAnnualizedCashOnCash() = %.4f, want 365", got)
	}
	if got := CalculateAnnualizedCashOnCash(100, 0); got != 0 {
		t.Fatalf("CalculateAnnualizedCashOnCash() with zero dollar-days = %.4f, want 0", got)
	}
	if got := CalculateAnnualizedCashOnCash(100, -1); got != 0 {
		t.Fatalf("CalculateAnnualizedCashOnCash() with negative dollar-days = %.4f, want 0", got)
	}
}

func TestAccumulateDollarDays(t *testing.T) {
	opened := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	closed := opened.AddDate(0, 0, 2)
	longPosition := &models.LongPosition{Opened: opened, Closed: &closed, BuyPrice: 10, Shares: 10}
	put := &models.Option{Type: "Put", Opened: opened, Closed: &closed, Strike: 20, Contracts: 1}
	call := &models.Option{Type: "Call", Opened: opened, Closed: &closed, Strike: 20, Contracts: 1}

	longDays := longPosition.CalculateDollarDays()
	putDays := put.CalculatePutExposureDollarDays()
	if got, want := AccumulateDollarDays([]*models.LongPosition{longPosition}, []*models.Option{put, call}), longDays+putDays; got != want {
		t.Fatalf("AccumulateDollarDays() = %.2f, want %.2f", got, want)
	}
}
