package web

import (
	"log"
	"net/http"
)

// tutorialHandler handles tutorial page requests and redirects to dashboard with tutorial modal
func (s *Server) tutorialHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HANDLER] Processing Tutorial page request")
	
	// For now, just redirect to dashboard where the tutorial modal will be triggered
	// This provides a simple endpoint that the tutorial nav link can point to
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// helpHandler renders the help page
func (s *Server) helpHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HANDLER] Processing Help page request")
	
	symbols, err := s.symbolService.GetDistinctSymbols()
	if err != nil {
		log.Printf("[HANDLER] ERROR: Failed to get symbols for Help page: %v", err)
		http.Error(w, "Failed to load symbols", http.StatusInternalServerError)
		return
	}

	data := HelpData{
		AllSymbols: symbols,
		CurrentDB:  s.getCurrentDatabaseName(),
		ActivePage: "help",
	}
	
	log.Printf("[HANDLER] Rendering Help page template with %d symbols", len(symbols))
	s.renderTemplate(w, "help.html", data)
}

// zenHandler renders the zen page with animated market flow
func (s *Server) zenHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HANDLER] Processing Zen page request")
	
	symbols, err := s.symbolService.GetDistinctSymbols()
	if err != nil {
		log.Printf("[HANDLER] ERROR: Failed to get symbols for Zen page: %v", err)
		http.Error(w, "Failed to load symbols", http.StatusInternalServerError)
		return
	}

	data := ZenData{
		AllSymbols: symbols,
		CurrentDB:  s.getCurrentDatabaseName(),
		ActivePage: "zen",
	}
	
	log.Printf("[HANDLER] Rendering Zen page template")
	s.renderTemplate(w, "zen.html", data)
}