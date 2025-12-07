package main

import (
	"fmt"
	"lab02_mahoa/server/database"
	"lab02_mahoa/server/handlers"
	"lab02_mahoa/server/models"
	"log"
	"net/http"
	"os"
	"strings"
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

	fmt.Println("🚀 Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// CORS middleware
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

// setupRoutes configures all API routes
func setupRoutes() {
	// Auth routes
	http.HandleFunc("/api/auth/register", corsMiddleware(handlers.RegisterHandler))
	http.HandleFunc("/api/auth/login", corsMiddleware(handlers.LoginHandler))
	http.HandleFunc("/api/auth/logout", corsMiddleware(handlers.LogoutHandler))

	// Note routes (using custom router for method handling)
	http.HandleFunc("/api/notes", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		NotesRouter(w, r)
	}))

	// Note detail routes (for delete, revoke)
	http.HandleFunc("/api/notes/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		NotesDetailRouter(w, r)
	}))

	// Share routes
	http.HandleFunc("/api/shares/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		SharesRouter(w, r)
	}))
}

// NotesRouter handles /api/notes endpoint (list and create)
func NotesRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handlers.ListNotesHandler(w, r)
	case http.MethodPost:
		handlers.CreateNoteHandler(w, r)
	default:
		handlers.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// NotesDetailRouter handles /api/notes/:id endpoint (get, delete, revoke, share)
func NotesDetailRouter(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL: /api/notes/:id or /api/notes/:id/revoke or /api/notes/:id/share
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/notes/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		handlers.RespondWithError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	// Check if this is a revoke request
	if len(pathParts) >= 2 && pathParts[1] == "revoke" {
		if r.Method != http.MethodPost {
			handlers.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		handlers.RevokeShareHandler(w, r)
		return
	}

	// Check if this is a share creation request
	if len(pathParts) >= 2 && pathParts[1] == "share" {
		if r.Method != http.MethodPost {
			handlers.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		handlers.CreateShareHandler(w, r)
		return
	}

	// Otherwise handle GET or DELETE
	switch r.Method {
	case http.MethodGet:
		handlers.GetNoteHandler(w, r)
	case http.MethodDelete:
		handlers.DeleteNoteHandler(w, r)
	default:
		handlers.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// SharesRouter handles share-related endpoints
func SharesRouter(w http.ResponseWriter, r *http.Request) {
	// Placeholder for share routes
	handlers.RespondWithError(w, http.StatusNotImplemented, "Share endpoint not yet implemented")
}
