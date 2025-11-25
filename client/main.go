package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run client/main.go <command>")
		fmt.Println("Commands:")
		fmt.Println("  status     - Check server status")
		fmt.Println("  health     - Check server health")
		return
	}

	command := os.Args[1]

	switch command {
	case "status":
		getStatus()
	case "health":
		checkHealth()
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}

func getStatus() {
	resp, err := http.Get("http://localhost:8080")
	if err != nil {
		fmt.Printf("❌ Error: Cannot connect to server - %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("✅ Server Status: %s\n", string(body))
}

func checkHealth() {
	resp, err := http.Get("http://localhost:8080/health")
	if err != nil {
		fmt.Printf("❌ Error: Cannot connect to server - %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("✅ Server Health: %s\n", string(body))
}
