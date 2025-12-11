package main

import (
	"fmt"
	"lab02_mahoa/server/database"
	"lab02_mahoa/server/handlers"
	"lab02_mahoa/server/jobs"
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
	if err := database.InitDB(&models.User{}, &models.Note{}, &models.SharedLink{}, &models.E2EEShare{}); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Start background cleanup job for expired shares and links
	db := database.GetDB()
	jobs.StartCleanupJob(db)

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

	// Note detail routes (for delete, revoke, e2ee)
	http.HandleFunc("/api/notes/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is an E2EE share creation request
		if strings.HasSuffix(r.URL.Path, "/e2ee") {
			if r.Method == http.MethodPost {
				handlers.CreateE2EEShareHandler(w, r)
				return
			}
		}
		NotesDetailRouter(w, r)
	}))

	// Share routes
	http.HandleFunc("/api/shares/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		SharesRouter(w, r)
	}))

	// E2EE routes
	http.HandleFunc("/api/e2ee", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		E2EEListRouter(w, r)
	}))

	http.HandleFunc("/api/e2ee/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		E2EEDetailRouter(w, r)
	}))

	// User public key routes
	http.HandleFunc("/api/user/publickey", corsMiddleware(handlers.UpdatePublicKeyHandler))
	http.HandleFunc("/api/users/", corsMiddleware(handlers.GetPublicKeyHandler))
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
	// Extract token from path: /api/shares/:token
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/shares/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		handlers.RespondWithError(w, http.StatusBadRequest, "Share token is required")
		return
	}

	// Handle GET request to access shared note
	if r.Method == http.MethodGet {
		handlers.GetSharedNoteHandler(w, r)
		return
	}

	handlers.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

// E2EEListRouter handles /api/e2ee endpoint (list E2EE shares)
func E2EEListRouter(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		handlers.ListE2EESharesHandler(w, r)
		return
	}
	handlers.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

// E2EEDetailRouter handles /api/e2ee/:id endpoint (get, delete)
func E2EEDetailRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handlers.GetE2EEShareHandler(w, r)
	case http.MethodDelete:
		handlers.DeleteE2EEShareHandler(w, r)
	default:
		handlers.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}


