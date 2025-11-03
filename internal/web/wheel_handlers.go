package web

import (
	"encoding/json"
	"net/http"
	"time"

	"stonks/internal/models"
)

// WheelDashboardHandler displays the wheel management dashboard
func (s *Server) WheelDashboardHandler(w http.ResponseWriter, r *http.Request) {
	symbols, _ := s.symbolService.GetDistinctSymbols()

	data := struct {
		ActivePage string
		CurrentDB  string
		AllSymbols []string
	}{
		ActivePage: "wheels",
		CurrentDB:  s.getCurrentDatabaseName(),
		AllSymbols: symbols,
	}

	s.renderTemplate(w, "wheels.html", data)
}

// GetWheelDataHandler returns all wheel prototype data
func (s *Server) GetWheelDataHandler(w http.ResponseWriter, r *http.Request) {
	data, err := models.LoadWheelPrototypeData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// GetWheelsHandler returns wheels filtered by account and status
func (s *Server) GetWheelsHandler(w http.ResponseWriter, r *http.Request) {
	accountID := r.URL.Query().Get("account")
	status := r.URL.Query().Get("status")

	data, err := models.LoadWheelPrototypeData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wheels := data.GetWheelsByStatus(accountID, status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wheels)
}

// CreateWheelRequest represents a request to create a new wheel
type CreateWheelRequest struct {
	AccountID            string `json:"accountId"`
	Symbol               string `json:"symbol"`
	ContractSize         int    `json:"contractSize"`
	TargetStrikeStrategy string `json:"targetStrikeStrategy"`
	Notes                string `json:"notes"`
}

// CreateWheelHandler creates a new wheel campaign
func (s *Server) CreateWheelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateWheelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Load data
	data, err := models.LoadWheelPrototypeData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create new wheel
	wheel := models.Wheel{
		ID:                   data.GetNextWheelID(),
		AccountID:            req.AccountID,
		Symbol:               req.Symbol,
		Status:               "active",
		CreatedAt:            time.Now().UTC().Format(time.RFC3339),
		ClosedAt:             nil,
		ContractSize:         req.ContractSize,
		TargetStrikeStrategy: req.TargetStrikeStrategy,
		Notes:                req.Notes,
		Summary: models.WheelSummary{
			TotalPremiumCollected: 0.00,
			TotalCapitalGains:     0.00,
			TotalDividends:        0.00,
			TotalProfit:           0.00,
			CurrentStockPosition:  0,
			CurrentCashSecured:    0.00,
			NumberOfCycles:        0,
			NumberOfPutTrades:     0,
			NumberOfCallTrades:    0,
			NumberOfAssignments:   0,
			NumberOfExpirations:   0,
			NumberOfRolls:         0,
		},
		Trades: []models.WheelTrade{},
	}

	// Add wheel to data
	data.Wheels = append(data.Wheels, wheel)

	// Save data
	if err := models.SaveWheelPrototypeData(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wheel)
}

// AddWheelTradeRequest represents a request to add a trade to a wheel
type AddWheelTradeRequest struct {
	WheelID        string   `json:"wheelId"`
	TradeType      string   `json:"tradeType"`
	OpenDate       string   `json:"openDate"`
	ExpirationDate string   `json:"expirationDate"`
	Strike         float64  `json:"strike"`
	Contracts      int      `json:"contracts"`
	Premium        float64  `json:"premium"`
	Commission     float64  `json:"commission"`
	StockPrice     float64  `json:"stockPrice"`
	Notes          string   `json:"notes"`
	// For assignments
	AssignmentPrice *float64 `json:"assignmentPrice,omitempty"`
	// For rolls
	ClosePremium    *float64 `json:"closePremium,omitempty"`
	CloseCommission *float64 `json:"closeCommission,omitempty"`
}

// AddWheelTradeHandler adds a trade to a wheel campaign
func (s *Server) AddWheelTradeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AddWheelTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Load data
	data, err := models.LoadWheelPrototypeData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Find wheel
	wheel := data.GetWheelByID(req.WheelID)
	if wheel == nil {
		http.Error(w, "Wheel not found", http.StatusNotFound)
		return
	}

	// Determine cycle and sequence numbers
	sequenceNumber := len(wheel.Trades) + 1
	cycleNumber := 1
	if len(wheel.Trades) > 0 {
		lastTrade := wheel.Trades[len(wheel.Trades)-1]
		cycleNumber = lastTrade.CycleNumber
		// If last trade was a call assignment, start new cycle
		if lastTrade.TradeType == "CALL_ASSIGN" {
			cycleNumber++
		}
	}

	// Calculate net proceeds
	netProceeds := (req.Premium * float64(req.Contracts) * 100) - req.Commission

	// Create trade
	trade := models.WheelTrade{
		ID:              data.GetNextTradeID(),
		WheelID:         req.WheelID,
		TradeType:       req.TradeType,
		SequenceNumber:  sequenceNumber,
		CycleNumber:     cycleNumber,
		OpenDate:        req.OpenDate,
		ExpirationDate:  req.ExpirationDate,
		CloseDate:       nil,
		Strike:          req.Strike,
		Contracts:       req.Contracts,
		Premium:         req.Premium,
		Commission:      req.Commission,
		StockPrice:      req.StockPrice,
		AssignmentPrice: req.AssignmentPrice,
		NetProceeds:     netProceeds,
		Status:          "open",
		Outcome:         nil,
		Notes:           req.Notes,
	}

	// Handle roll details
	if req.TradeType == "PUT_ROLL" || req.TradeType == "CALL_ROLL" {
		if req.ClosePremium != nil && req.CloseCommission != nil {
			closeProceeds := (*req.ClosePremium * float64(req.Contracts) * 100) - *req.CloseCommission
			netRollCredit := netProceeds - closeProceeds

			trade.RollDetails = &models.RollDetails{
				ClosedTradeID:   nil, // Could link to previous trade
				ClosePremium:    *req.ClosePremium,
				CloseCommission: *req.CloseCommission,
				NetRollCredit:   netRollCredit,
			}
			trade.NetProceeds = netRollCredit
		}
	}

	// Add trade to wheel
	wheel.Trades = append(wheel.Trades, trade)

	// Update wheel summary
	updateWheelSummary(wheel, &trade)

	// Update wheel in data
	for i := range data.Wheels {
		if data.Wheels[i].ID == req.WheelID {
			data.Wheels[i] = *wheel
			break
		}
	}

	// Save data
	if err := models.SaveWheelPrototypeData(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trade)
}

// updateWheelSummary updates the wheel summary based on a new trade
func updateWheelSummary(wheel *models.Wheel, trade *models.WheelTrade) {
	// Update premium collected
	if trade.TradeType == "CSP" || trade.TradeType == "CC" ||
	   trade.TradeType == "PUT_ROLL" || trade.TradeType == "CALL_ROLL" {
		wheel.Summary.TotalPremiumCollected += trade.NetProceeds
	}

	// Update trade counts
	switch trade.TradeType {
	case "CSP", "PUT_ROLL", "PUT_EXPIRE":
		wheel.Summary.NumberOfPutTrades++
	case "CC", "CALL_ROLL", "CALL_EXPIRE":
		wheel.Summary.NumberOfCallTrades++
	}

	// Update assignment/expiration/roll counts
	switch trade.TradeType {
	case "PUT_ASSIGN", "CALL_ASSIGN":
		wheel.Summary.NumberOfAssignments++
	case "PUT_EXPIRE", "CALL_EXPIRE":
		wheel.Summary.NumberOfExpirations++
	case "PUT_ROLL", "CALL_ROLL":
		wheel.Summary.NumberOfRolls++
	}

	// Update stock position
	if trade.TradeType == "PUT_ASSIGN" {
		wheel.Summary.CurrentStockPosition += trade.Contracts * 100
	} else if trade.TradeType == "CALL_ASSIGN" {
		wheel.Summary.CurrentStockPosition -= trade.Contracts * 100
		if wheel.Summary.CurrentStockPosition == 0 {
			wheel.Summary.NumberOfCycles++
		}
	}

	// Update cash secured amount for open puts
	if trade.TradeType == "CSP" {
		wheel.Summary.CurrentCashSecured += trade.Strike * float64(trade.Contracts) * 100
	} else if trade.TradeType == "PUT_ASSIGN" || trade.TradeType == "PUT_EXPIRE" {
		wheel.Summary.CurrentCashSecured -= trade.Strike * float64(trade.Contracts) * 100
	}

	// Update total profit
	wheel.Summary.TotalProfit = wheel.Summary.TotalPremiumCollected +
		wheel.Summary.TotalCapitalGains + wheel.Summary.TotalDividends
}

// UpdateWheelStatusRequest represents a request to update wheel status
type UpdateWheelStatusRequest struct {
	Status string `json:"status"`
}

// UpdateWheelStatusHandler updates a wheel's status
func (s *Server) UpdateWheelStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	wheelID := r.URL.Query().Get("id")
	if wheelID == "" {
		http.Error(w, "Wheel ID required", http.StatusBadRequest)
		return
	}

	var req UpdateWheelStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Load data
	data, err := models.LoadWheelPrototypeData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Find and update wheel
	found := false
	for i := range data.Wheels {
		if data.Wheels[i].ID == wheelID {
			data.Wheels[i].Status = req.Status
			if req.Status == "closed" {
				now := time.Now().UTC().Format(time.RFC3339)
				data.Wheels[i].ClosedAt = &now
			}
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Wheel not found", http.StatusNotFound)
		return
	}

	// Save data
	if err := models.SaveWheelPrototypeData(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
