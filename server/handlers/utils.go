package handlers

import (
	"encoding/json"
	"lab02_mahoa/server/models"
	"log"
	"net/http"
)

// RespondWithJSON writes JSON response
func RespondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

// RespondWithError writes error JSON response
func RespondWithError(w http.ResponseWriter, statusCode int, message string) {
	RespondWithJSON(w, statusCode, models.ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}
