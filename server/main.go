package main

import (
	"fmt"
	"net/http"
)

func main() {
	// Simple test endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","message":"Server is running"}`)
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy"}`)
	})

	fmt.Println("🚀 Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
