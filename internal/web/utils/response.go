package utils

import (
	"encoding/json"
	"net/http"
)

// RespondWithJSON writes JSON response with given status code
func RespondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// RespondWithError writes error response as JSON
func RespondWithError(w http.ResponseWriter, statusCode int, message string) {
	RespondWithJSON(w, statusCode, map[string]string{"error": message})
}

// RespondWithSuccess writes success response as JSON
func RespondWithSuccess(w http.ResponseWriter, message string) {
	RespondWithJSON(w, http.StatusOK, map[string]string{"success": message})
}

// DecodeJSONRequest decodes JSON request body into dest
func DecodeJSONRequest(r *http.Request, dest interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dest)
}

// DecodeJSONRequestOrError combines decode with error response
func DecodeJSONRequestOrError(w http.ResponseWriter, r *http.Request, dest interface{}) bool {
	if err := DecodeJSONRequest(r, dest); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return false
	}
	return true
}
