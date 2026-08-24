package web

import (
	"encoding/json"
	"net/http"
	"stonks/internal/markets"
)

func (s *Server) marketsAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(markets.Definitions())
}
