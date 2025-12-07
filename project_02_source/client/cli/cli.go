package cli

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"lab02_mahoa/client/api"
	"lab02_mahoa/client/crypto"
	"net/http"
	"os"
	"text/tabwriter"
)

// Run executes CLI commands
func Run(args []string) {
	if len(args) == 0 {
		printUsage()
		return
	}

	command := args[0]

	switch command {
	case "list":
		handleList()
	case "delete":
		handleDelete(args[1:])
	case "revoke":
		handleRevoke(args[1:])
	case "login":
		handleLogin(args[1:])
	case "register":
		handleRegister(args[1:])
	case "upload":
		handleUpload(args[1:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`
Secure Notes CLI - Usage:
  list                         List all notes
  delete -id <note_id>         Delete a note by ID
  revoke -id <note_id>         Revoke sharing for a note
  login -token <jwt_token>     Save JWT token for authentication
  register -u <user> -p <pass> Register new account
  upload -t <title> -c <file>  Upload and encrypt a note from file
`)
}

// handleList lists all notes
func handleList() {
	token := loadToken()
	if token == "" {
		fmt.Println("❌ Error: No token found. Please login first.")
		fmt.Println("   Use: secure-notes login -token <your_jwt_token>")
		return
	}

	// Create client with token
	client := &api.Client{Token: token}

	// Call API
	notes, err := client.ListNotes()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	if len(notes) == 0 {
		fmt.Println("📭 No notes found")
		return
	}

	// Display in table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTitle\tSize\tCreated At")
	fmt.Fprintln(w, "--\t-----\t----\t----------")

	for _, note := range notes {
		size := len(note.EncryptedContent)
		fmt.Fprintf(w, "%d\t%s\t%d bytes\t%s\n", note.ID, note.Title, size, note.CreatedAt.Format("2006-01-02 15:04"))
	}

	w.Flush()
	fmt.Printf("\n✅ Total: %d notes\n", len(notes))
}

// handleDelete deletes a note
func handleDelete(args []string) {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	noteID := fs.String("id", "", "Note ID to delete")
	fs.Parse(args)

	if *noteID == "" {
		fmt.Println("❌ Error: Please provide -id <note_id>")
		fmt.Println("   Usage: secure-notes delete -id 123")
		return
	}

	token := loadToken()
	if token == "" {
		fmt.Println("❌ Error: No token found. Please login first.")
		return
	}

	// Parse ID and create client
	var id uint
	fmt.Sscanf(*noteID, "%d", &id)
	client := &api.Client{Token: token}

	if err := client.DeleteNote(id); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Println("✅ Note deleted successfully")
}

// handleRevoke revokes sharing for a note
func handleRevoke(args []string) {
	fs := flag.NewFlagSet("revoke", flag.ContinueOnError)
	noteID := fs.String("id", "", "Note ID to revoke sharing")
	fs.Parse(args)

	if *noteID == "" {
		fmt.Println("❌ Error: Please provide -id <note_id>")
		fmt.Println("   Usage: secure-notes revoke -id 123")
		return
	}

	token := loadToken()
	if token == "" {
		fmt.Println("❌ Error: No token found. Please login first.")
		return
	}

	// Parse ID and create client
	var id uint
	fmt.Sscanf(*noteID, "%d", &id)
	client := &api.Client{Token: token}

	if err := client.RevokeShare(id); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Println("✅ Sharing revoked successfully")
}

// handleLogin saves JWT token to file
func handleLogin(args []string) {
	fs := flag.NewFlagSet("login", flag.ContinueOnError)
	token := fs.String("token", "", "JWT token")
	fs.Parse(args)

	if *token == "" {
		fmt.Println("❌ Error: Please provide -token <jwt_token>")
		fmt.Println("   Usage: secure-notes login -token eyJhbGc...")
		return
	}

	// Save token to .cli_token file
	tokenFile := ".cli_token"
	if err := os.WriteFile(tokenFile, []byte(*token), 0600); err != nil {
		fmt.Printf("❌ Error saving token: %v\n", err)
		return
	}

	fmt.Println("✅ Token saved to .cli_token")
}

// handleRegister registers a new user
func handleRegister(args []string) {
	fs := flag.NewFlagSet("register", flag.ContinueOnError)
	username := fs.String("u", "", "Username")
	password := fs.String("p", "", "Password")
	fs.Parse(args)

	if *username == "" || *password == "" {
		fmt.Println("❌ Error: Please provide -u <username> -p <password>")
		fmt.Println("   Usage: secure-notes register -u alice -p pass123")
		return
	}

	// Call API
	regReq := api.RegisterRequest{
		Username: *username,
		Password: *password,
	}

	jsonData, err := json.Marshal(regReq)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	resp, err := http.Post(api.BaseURL+"/auth/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		fmt.Println("✅ Registration successful!")
		fmt.Println("   Now use: secure-notes login -token <your_jwt_token>")
	} else {
		fmt.Printf("❌ Error: %s\n", string(body))
	}
}

// handleUpload uploads and encrypts a note
func handleUpload(args []string) {
	fs := flag.NewFlagSet("upload", flag.ContinueOnError)
	title := fs.String("t", "", "Note title")
	filePath := fs.String("c", "", "File path or content")
	fs.Parse(args)

	if *title == "" || *filePath == "" {
		fmt.Println("❌ Error: Please provide -t <title> -c <content>")
		fmt.Println("   Usage: secure-notes upload -t \"My Note\" -c \"/path/to/file.txt\"")
		return
	}

	token := loadToken()
	if token == "" {
		fmt.Println("❌ Error: No token found. Please login first.")
		return
	}

	// Read file content or treat as direct content
	var content string
	if _, err := os.Stat(*filePath); err == nil {
		// File exists, read it
		data, err := os.ReadFile(*filePath)
		if err != nil {
			fmt.Printf("❌ Error reading file: %v\n", err)
			return
		}
		content = string(data)
	} else {
		// Treat as direct content
		content = *filePath
	}

	// Generate encryption key
	key, err := crypto.GenerateKey()
	if err != nil {
		fmt.Printf("❌ Error generating key: %v\n", err)
		return
	}
	keyStr := fmt.Sprintf("key_%s", *title)

	// Encrypt content
	encryptedContent, iv, err := crypto.EncryptAES(content, key)
	if err != nil {
		fmt.Printf("❌ Error encrypting: %v\n", err)
		return
	}

	// Encrypt key
	encryptedKey, _, err := crypto.EncryptAES(keyStr, key)
	if err != nil {
		fmt.Printf("❌ Error encrypting key: %v\n", err)
		return
	}

	// Create client and upload
	client := &api.Client{Token: token}
	if err := client.CreateNote(*title, encryptedContent, encryptedKey, iv); err != nil {
		fmt.Printf("❌ Error uploading: %v\n", err)
		return
	}

	fmt.Println("✅ Note uploaded and encrypted successfully!")
}

// loadToken loads JWT token from file or environment variable
func loadToken() string {
	// Try to read from .cli_token file first
	if data, err := os.ReadFile(".cli_token"); err == nil {
		return string(bytes.TrimSpace(data))
	}

	// Fall back to environment variable
	return os.Getenv("CLI_TOKEN")
}
