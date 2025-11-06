package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// WheelPrototypeData represents the complete prototype data structure
type WheelPrototypeData struct {
	Accounts       []WheelAccount     `json:"accounts"`
	Wheels         []Wheel            `json:"wheels"`
	TradeSequence  int                `json:"tradeSequence"`
	WheelSequence  int                `json:"wheelSequence"`
	AccountSequence int               `json:"accountSequence"`
	Metadata       WheelMetadata      `json:"metadata"`
	TradeTypes     map[string]TradeTypeInfo `json:"tradeTypes"`
}

// WheelAccount represents a trading account for wheel campaigns
type WheelAccount struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	StartingCash float64   `json:"startingCash"`
	CurrentCash  float64   `json:"currentCash"`
	CreatedAt    string    `json:"createdAt"`
}

// Wheel represents a wheel campaign
type Wheel struct {
	ID                   string       `json:"id"`
	AccountID            string       `json:"accountId"`
	Symbol               string       `json:"symbol"`
	Status               string       `json:"status"` // active, inactive, closed
	CreatedAt            string       `json:"createdAt"`
	ClosedAt             *string      `json:"closedAt"`
	ContractSize         int          `json:"contractSize"`
	TargetStrikeStrategy string       `json:"targetStrikeStrategy"`
	Notes                string       `json:"notes"`
	Summary              WheelSummary `json:"summary"`
	Trades               []WheelTrade `json:"trades"`
}

// WheelSummary contains aggregated statistics for a wheel
type WheelSummary struct {
	TotalPremiumCollected float64 `json:"totalPremiumCollected"`
	TotalCapitalGains     float64 `json:"totalCapitalGains"`
	TotalDividends        float64 `json:"totalDividends"`
	TotalProfit           float64 `json:"totalProfit"`
	CurrentStockPosition  int     `json:"currentStockPosition"`
	CurrentCashSecured    float64 `json:"currentCashSecured"`
	NumberOfCycles        int     `json:"numberOfCycles"`
	NumberOfPutTrades     int     `json:"numberOfPutTrades"`
	NumberOfCallTrades    int     `json:"numberOfCallTrades"`
	NumberOfAssignments   int     `json:"numberOfAssignments"`
	NumberOfExpirations   int     `json:"numberOfExpirations"`
	NumberOfRolls         int     `json:"numberOfRolls"`
}

// WheelTrade represents a trade in a wheel campaign
type WheelTrade struct {
	ID              string       `json:"id"`
	WheelID         string       `json:"wheelId"`
	TradeType       string       `json:"tradeType"`
	SequenceNumber  int          `json:"sequenceNumber"`
	CycleNumber     int          `json:"cycleNumber"`
	OpenDate        string       `json:"openDate"`
	ExpirationDate  string       `json:"expirationDate"`
	CloseDate       *string      `json:"closeDate"`
	Strike          float64      `json:"strike"`
	Contracts       int          `json:"contracts"`
	Premium         float64      `json:"premium"`
	Commission      float64      `json:"commission"`
	StockPrice      float64      `json:"stockPrice"`
	AssignmentPrice *float64     `json:"assignmentPrice"`
	NetProceeds     float64      `json:"netProceeds"`
	Status          string       `json:"status"` // open, closed
	Outcome         *string      `json:"outcome"`
	Notes           string       `json:"notes"`
	RollDetails     *RollDetails `json:"rollDetails,omitempty"`
}

// RollDetails contains information about rolled positions
type RollDetails struct {
	ClosedTradeID   *string `json:"closedTradeId"`
	ClosePremium    float64 `json:"closePremium"`
	CloseCommission float64 `json:"closeCommission"`
	NetRollCredit   float64 `json:"netRollCredit"`
}

// TradeTypeInfo describes a trade type
type TradeTypeInfo struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	OptionType  string `json:"optionType"`
	Action      string `json:"action"`
}

// WheelMetadata contains file metadata
type WheelMetadata struct {
	Version      string `json:"version"`
	LastModified string `json:"lastModified"`
	Description  string `json:"description"`
}

const wheelPrototypeFile = "data/wheels_prototype.json"

// LoadWheelPrototypeData loads the prototype data from JSON file
func LoadWheelPrototypeData() (*WheelPrototypeData, error) {
	// Get absolute path
	absPath, err := filepath.Abs(wheelPrototypeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Read file
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wheel prototype file: %w", err)
	}

	// Parse JSON
	var prototypeData WheelPrototypeData
	if err := json.Unmarshal(data, &prototypeData); err != nil {
		return nil, fmt.Errorf("failed to parse wheel prototype JSON: %w", err)
	}

	return &prototypeData, nil
}

// SaveWheelPrototypeData saves the prototype data to JSON file
func SaveWheelPrototypeData(data *WheelPrototypeData) error {
	// Update metadata
	data.Metadata.LastModified = time.Now().UTC().Format(time.RFC3339)

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal wheel prototype data: %w", err)
	}

	// Get absolute path
	absPath, err := filepath.Abs(wheelPrototypeFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Write file
	if err := os.WriteFile(absPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write wheel prototype file: %w", err)
	}

	return nil
}

// GetNextTradeID generates the next trade ID
func (d *WheelPrototypeData) GetNextTradeID() string {
	d.TradeSequence++
	return fmt.Sprintf("trade_%03d", d.TradeSequence)
}

// GetNextWheelID generates the next wheel ID
func (d *WheelPrototypeData) GetNextWheelID() string {
	d.WheelSequence++
	return fmt.Sprintf("wheel_%03d", d.WheelSequence)
}

// GetNextAccountID generates the next account ID
func (d *WheelPrototypeData) GetNextAccountID() string {
	d.AccountSequence++
	return fmt.Sprintf("acc_%03d", d.AccountSequence)
}

// GetAccountByID retrieves an account by ID
func (d *WheelPrototypeData) GetAccountByID(accountID string) *WheelAccount {
	for i := range d.Accounts {
		if d.Accounts[i].ID == accountID {
			return &d.Accounts[i]
		}
	}
	return nil
}

// GetWheelByID retrieves a wheel by ID
func (d *WheelPrototypeData) GetWheelByID(wheelID string) *Wheel {
	for i := range d.Wheels {
		if d.Wheels[i].ID == wheelID {
			return &d.Wheels[i]
		}
	}
	return nil
}

// GetWheelsByAccount retrieves all wheels for an account
func (d *WheelPrototypeData) GetWheelsByAccount(accountID string) []Wheel {
	var wheels []Wheel
	for _, wheel := range d.Wheels {
		if accountID == "" || wheel.AccountID == accountID {
			wheels = append(wheels, wheel)
		}
	}
	return wheels
}

// GetWheelsByStatus retrieves wheels filtered by status
func (d *WheelPrototypeData) GetWheelsByStatus(accountID, status string) []Wheel {
	var wheels []Wheel
	for _, wheel := range d.Wheels {
		if (accountID == "" || wheel.AccountID == accountID) &&
		   (status == "" || wheel.Status == status) {
			wheels = append(wheels, wheel)
		}
	}
	return wheels
}
