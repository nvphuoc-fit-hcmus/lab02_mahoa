package main

import (
	"fmt"
	"lab02_mahoa/server/database"
	"lab02_mahoa/server/handlers"
	"lab02_mahoa/server/models"
	"log"
	"net/http"
	"os"
)

func main() {
	// Create storage directory if it doesn't exist
	if err := os.MkdirAll("storage", 0755); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}

	// Initialize database with models
	if err := database.InitDB(&models.User{}, &models.Note{}, &models.SharedLink{}); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Setup routes
	setupRoutes()

	// Start server
	fmt.Println("🚀 Server is running on http://localhost:8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// CORS middleware to handle cross-origin requests
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func setupRoutes() {
	// Root endpoint - API info
	http.HandleFunc("/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","message":"Secure Note Sharing API","version":"1.0"}`)
	}))

	// Health check endpoint
	http.HandleFunc("/health", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy"}`)
	}))

	// Authentication endpoints (RESTful)
	http.HandleFunc("/api/auth/register", corsMiddleware(handlers.RegisterHandler))
	http.HandleFunc("/api/auth/login", corsMiddleware(handlers.LoginHandler))
	http.HandleFunc("/api/auth/logout", corsMiddleware(handlers.LogoutHandler))

	// Note management endpoints (RESTful)
	http.HandleFunc("/api/notes", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CreateNoteHandler(w, r)
		case http.MethodGet:
			handlers.ListNotesHandler(w, r)
		default:
			handlers.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}))

	http.HandleFunc("/api/notes/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetNoteHandler(w, r)
		case http.MethodDelete:
			handlers.DeleteNoteHandler(w, r)
		default:
			handlers.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}))
}
