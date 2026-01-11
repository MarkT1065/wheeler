package models

import (
	"log"
	"testing"
	"time"
)

func eztime(dt string) time.Time {
	ti, err := time.Parse("2006-01-02 15:04:05", "2025-09-01 10:30:05")
	if err != nil {
		log.Fatal(err)
	}
	return ti
}

func eztimeptr(dt string) *time.Time {
	d := eztime(dt)
	return &d
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TestIndex(t *testing.T) {
	opt1exit := .12
	opt1opened := eztime("2025-09-01 10:30:05")
	opt1 := &Option{
		ID:         1,
		Symbol:     "VZ",
		Type:       "Call",
		Opened:     opt1opened,
		Strike:     44.0,
		Expiration: eztime("2025-09-19"),
		Premium:    .44,
		Contracts:  1,
		Commission: 1.30,
		Closed:     eztimeptr("2025-09-12"),
		ExitPrice:  &opt1exit,
	}

	opt2 := &Option{
		ID:         2,
		Symbol:     "CVX",
		Type:       "Put",
		Opened:     eztime("2025-09-01 10:35:00"),
		Strike:     145.0,
		Expiration: eztime("2025-10-17"),
		Premium:    1.56,
		Contracts:  1,
		Commission: .65,
	}

	index, err := Index([]*Option{opt1, opt2})
	if err != nil {
		t.Fatalf("Index() returned error: %v", err)
	}

	// Test that all expected keys exist
	expectedKeys := []string{"id", "symbol", "type", "opened", "expiration", "open", "closed"}
	for _, key := range expectedKeys {
		if _, exists := index[key]; !exists {
			t.Errorf("Index missing expected key: %s", key)
		}
	}

	// Test ID index
	idIndex := index["id"].(map[string]*Option)
	if len(idIndex) != 2 {
		t.Errorf("Expected 2 entries in ID index, got %d", len(idIndex))
	}
	if idIndex["1"] != opt1 {
		t.Error("ID index[\"1\"] doesn't match opt1")
	}
	if idIndex["2"] != opt2 {
		t.Error("ID index[\"2\"] doesn't match opt2")
	}

	// Test symbol index
	symbolIndex := index["symbol"].(map[string][]*Option)
	if len(symbolIndex["VZ"]) != 1 || symbolIndex["VZ"][0] != opt1 {
		t.Error("Symbol index for VZ incorrect")
	}
	if len(symbolIndex["CVX"]) != 1 || symbolIndex["CVX"][0] != opt2 {
		t.Error("Symbol index for CVX incorrect")
	}

	// Test type index
	typeIndex := index["type"].(map[string][]*Option)
	if len(typeIndex["Call"]) != 1 || typeIndex["Call"][0] != opt1 {
		t.Error("Type index for Call incorrect")
	}
	if len(typeIndex["Put"]) != 1 || typeIndex["Put"][0] != opt2 {
		t.Error("Type index for Put incorrect")
	}

	// Test open index (only opt2 is open)
	openIndex := index["open"].([]*Option)
	if len(openIndex) != 1 {
		t.Errorf("Expected 1 open option, got %d", len(openIndex))
	}
	if len(openIndex) > 0 && openIndex[0] != opt2 {
		t.Error("Open index doesn't contain opt2")
	}

	// Test closed index (only opt1 is closed)
	closedIndex := index["closed"].(map[int]map[time.Month][]*Option)
	sept2025Closed := closedIndex[2025][time.September]
	if len(sept2025Closed) != 1 || sept2025Closed[0] != opt1 {
		t.Error("Closed index for September 2025 incorrect")
	}
}
