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

func TestIndex(t *testing.T) {
	//testDB, err := database.NewDB("/home/mturansk/projects/src/github.com/markturansky/stonks/data/wheeler.db")
	//if err != nil {
	//	t.Fatalf("Failed to setup test database: %v", err)
	//}
	//defer testDB.Close()
	//
	//opts := NewOptionService(testDB.DB)
	//all, err := opts.GetAll()
	//if err != nil {
	//	t.Fatalf("Failed to get all: %v", err)
	//}
	//index, err := Index(all)
	//if err != nil {
	//	t.Fatalf("Failed to create index: %v", err)
	//}
	//t.Logf("index: %+v", index)

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

	index, _ := Index([]*Option{opt1, opt2})

	m := map[string]interface{}{
		"id": map[string]*Option{
			"1": opt1,
			"2": opt2,
		},
		"symbol": map[string][]*Option{
			"VZ":  {opt1},
			"CVX": {opt2},
		},
		"type": map[string][]*Option{
			"Call": {opt1},
			"Put":  {opt2},
		},
		"opened": map[int]map[time.Month][]*Option{
			2025: {
				time.September: {opt1, opt2},
			},
		},
		"expiration": map[int]map[time.Month]map[int][]*Option{
			2025: {
				time.September: {
					19: {opt1},
				},
				time.October: {
					17: {opt2},
				},
			},
		},
		"open": []*Option{
			opt2,
		},
		"closed": map[int]map[time.Month][]*Option{
			2025: {
				time.September: {opt1},
			},
		},
	}

	if !Compare(index, m) {
		t.Fatalf("%#v", index)
	}
}
