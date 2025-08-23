package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)


// dividendsAPIHandler handles CRUD operations for dividends
func (s *Server) dividendsAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.createDividendHandler(w, r)
	case http.MethodDelete:
		s.deleteDividendHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// createDividendHandler handles POST requests to create new dividends
func (s *Server) createDividendHandler(w http.ResponseWriter, r *http.Request) {
	var req DividendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Parse the received date
	receivedDate, err := time.Parse("2006-01-02", req.Received)
	if err != nil {
		http.Error(w, "Invalid received date format", http.StatusBadRequest)
		return
	}

	// Validate amount
	if req.Amount <= 0 {
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return
	}

	// Create the dividend
	dividend, err := s.dividendService.Create(req.Symbol, receivedDate, req.Amount)
	if err != nil {
		http.Error(w, "Failed to create dividend", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dividend)
}

// deleteDividendHandler handles DELETE requests to remove dividends
func (s *Server) deleteDividendHandler(w http.ResponseWriter, r *http.Request) {
	var req DividendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Delete the dividend using ID if provided, otherwise use compound key
	if req.ID != nil && *req.ID > 0 {
		// Delete by ID (preferred method)
		err := s.dividendService.DeleteByID(*req.ID)
		if err != nil {
			log.Printf("Error deleting dividend by ID: %v", err)
			http.Error(w, fmt.Sprintf("Failed to delete dividend: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		// Fallback to compound key deletion
		receivedDate, err := time.Parse("2006-01-02", req.Received)
		if err != nil {
			http.Error(w, "Invalid received date format", http.StatusBadRequest)
			return
		}

		err = s.dividendService.Delete(req.Symbol, receivedDate, req.Amount)
		if err != nil {
			log.Printf("Error deleting dividend: %v", err)
			http.Error(w, fmt.Sprintf("Failed to delete dividend: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

// longPositionsAPIHandler handles CRUD operations for long positions
func (s *Server) longPositionsAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.createLongPositionHandler(w, r)
	case http.MethodPut:
		s.updateLongPositionHandler(w, r)
	case http.MethodDelete:
		s.deleteLongPositionHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// createLongPositionHandler creates a new long position
func (s *Server) createLongPositionHandler(w http.ResponseWriter, r *http.Request) {
	var req LongPositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Symbol == "" || req.Opened == "" || req.Shares == 0 || req.BuyPrice == 0 {
		http.Error(w, "Symbol, opened date, shares, and buy price are required", http.StatusBadRequest)
		return
	}

	// Parse opened date
	openedDate, err := time.Parse("2006-01-02", req.Opened)
	if err != nil {
		http.Error(w, "Invalid opened date format", http.StatusBadRequest)
		return
	}

	// Create the long position
	position, err := s.longPositionService.Create(req.Symbol, openedDate, req.Shares, req.BuyPrice)
	if err != nil {
		log.Printf("Error creating long position: %v", err)
		http.Error(w, "Failed to create long position", http.StatusInternalServerError)
		return
	}

	// If closed date and/or exit price are provided, update them
	if req.Closed != nil && *req.Closed != "" {
		closedDate, err := time.Parse("2006-01-02", *req.Closed)
		if err != nil {
			http.Error(w, "Invalid closed date format", http.StatusBadRequest)
			return
		}

		var exitPrice float64
		if req.ExitPrice != nil {
			exitPrice = *req.ExitPrice
		}

		err = s.longPositionService.CloseByID(position.ID, closedDate, exitPrice)
		if err != nil {
			log.Printf("Error closing long position: %v", err)
			// Position was created, but closing failed - return success but log error
			log.Printf("Long position created but failed to close: %v", err)
		}

		// Refresh position data to get updated values
		position, _ = s.longPositionService.GetByID(position.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(position)
}

// updateLongPositionHandler updates an existing long position
func (s *Server) updateLongPositionHandler(w http.ResponseWriter, r *http.Request) {
	var req LongPositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// ID is required for updates
	if req.ID == nil {
		http.Error(w, "ID is required for updates", http.StatusBadRequest)
		return
	}

	// Parse dates
	openedDate, err := time.Parse("2006-01-02", req.Opened)
	if err != nil {
		http.Error(w, "Invalid opened date format", http.StatusBadRequest)
		return
	}

	var closedDate *time.Time
	if req.Closed != nil && *req.Closed != "" {
		parsed, err := time.Parse("2006-01-02", *req.Closed)
		if err != nil {
			http.Error(w, "Invalid closed date format", http.StatusBadRequest)
			return
		}
		closedDate = &parsed
	}

	// Update the long position
	position, err := s.longPositionService.UpdateByID(*req.ID, req.Symbol, openedDate, req.Shares, req.BuyPrice, closedDate, req.ExitPrice)
	if err != nil {
		log.Printf("Error updating long position: %v", err)
		http.Error(w, "Failed to update long position", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(position)
}

// deleteLongPositionHandler deletes a long position
func (s *Server) deleteLongPositionHandler(w http.ResponseWriter, r *http.Request) {
	var req LongPositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Try ID-based deletion first
	if req.ID != nil {
		err := s.longPositionService.DeleteByID(*req.ID)
		if err != nil {
			log.Printf("Error deleting long position by ID: %v", err)
			http.Error(w, "Failed to delete long position", http.StatusInternalServerError)
			return
		}
	} else {
		// Fallback to compound key deletion
		openedDate, err := time.Parse("2006-01-02", req.Opened)
		if err != nil {
			http.Error(w, "Invalid opened date format", http.StatusBadRequest)
			return
		}

		err = s.longPositionService.Delete(req.Symbol, openedDate, req.Shares, req.BuyPrice)
		if err != nil {
			log.Printf("Error deleting long position: %v", err)
			http.Error(w, fmt.Sprintf("Failed to delete long position: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}