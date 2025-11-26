package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Create storage directory if it doesn't exist
	if err := os.MkdirAll("storage", 0755); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}

	// Initialize database
	if err := InitDB(); err != nil {
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
	http.HandleFunc("/api/auth/register", corsMiddleware(RegisterHandler))
	http.HandleFunc("/api/auth/login", corsMiddleware(LoginHandler))
	http.HandleFunc("/api/auth/logout", corsMiddleware(LogoutHandler))

	// Note management endpoints (RESTful)
	http.HandleFunc("/api/notes", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			CreateNoteHandler(w, r)
		case http.MethodGet:
			ListNotesHandler(w, r)
		default:
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}))

	http.HandleFunc("/api/notes/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetNoteHandler(w, r)
		case http.MethodDelete:
			DeleteNoteHandler(w, r)
		default:
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}))
}
