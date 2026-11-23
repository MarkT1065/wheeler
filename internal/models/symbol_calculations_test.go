package models

import (
	"testing"
	"time"
)

func TestDollarDayCalculations(t *testing.T) {
	opened := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	closed := opened.AddDate(0, 0, 10)

	position := &LongPosition{
		Opened:   opened,
		Closed:   &closed,
		BuyPrice: 25,
		Shares:   40,
	}
	if got, want := position.CalculateDaysHeld(), 10; got != want {
		t.Fatalf("CalculateDaysHeld() = %d, want %d", got, want)
	}
	if got, want := position.CalculateDollarDays(), 10000.0; got != want {
		t.Fatalf("CalculateDollarDays() = %.2f, want %.2f", got, want)
	}

	put := &Option{
		Type:      "Put",
		Opened:    opened,
		Closed:    &closed,
		Strike:    50,
		Contracts: 2,
	}
	if got, want := put.CalculateDaysOpen(), 10; got != want {
		t.Fatalf("CalculateDaysOpen() = %d, want %d", got, want)
	}
	if got, want := put.CalculatePutExposureDollarDays(), 100000.0; got != want {
		t.Fatalf("CalculatePutExposureDollarDays() = %.2f, want %.2f", got, want)
	}

	call := &Option{Type: "Call", Opened: opened, Closed: &closed, Strike: 50, Contracts: 2}
	if got := call.CalculatePutExposureDollarDays(); got != 0 {
		t.Fatalf("call exposure dollar-days = %.2f, want 0", got)
	}
}

func TestDollarDayCalculationsUseAtLeastOneDay(t *testing.T) {
	opened := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	closed := opened

	position := &LongPosition{Opened: opened, Closed: &closed, BuyPrice: 10, Shares: 5}
	if got := position.CalculateDaysHeld(); got != 1 {
		t.Fatalf("same-day CalculateDaysHeld() = %d, want 1", got)
	}

	put := &Option{Type: "Put", Opened: opened, Closed: &closed, Strike: 10, Contracts: 1}
	if got := put.CalculateDaysOpen(); got != 1 {
		t.Fatalf("same-day CalculateDaysOpen() = %d, want 1", got)
	}
}
