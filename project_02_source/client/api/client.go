package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const BaseURL = "http://localhost:8080/api"

var (
	CurrentUserID   uint
	CurrentUsername string
	CurrentPassword string
	AuthToken       string // JWT Token để gọi API
)

type Client struct {
	Token string
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest represents login data
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

// CreateNoteRequest represents note creation data
type CreateNoteRequest struct {
	Title            string `json:"title"`
	EncryptedContent string `json:"encrypted_content"`
	EncryptedKey     string `json:"encrypted_key"`
	IV               string `json:"iv"`
}

// Note represents a note from the server
type Note struct {
	ID               uint      `json:"id"`
	Title            string    `json:"title"`
	EncryptedContent string    `json:"encrypted_content"`
	EncryptedKey     string    `json:"encrypted_key"`
	IV               string    `json:"iv"`
	CreatedAt        time.Time `json:"created_at"`
	IsShared         bool      `json:"is_shared"`
}

// ListNotesResponse represents the response from listing notes
type ListNotesResponse struct {
	Notes []Note `json:"notes"`
	Count int    `json:"count"`
}

// Register creates a new user account
func (c *Client) Register(username, password string) error {
	reqBody := RegisterRequest{
		Username: username,
		Password: password,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(BaseURL+"/auth/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed: %s", string(body))
	}

	return nil
}

// Login authenticates user and returns JWT token
func (c *Client) Login(username, password string) (string, error) {
	reqBody := LoginRequest{
		Username: username,
		Password: password,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(BaseURL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed: %s", string(body))
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", err
	}

	c.Token = loginResp.Token
	return loginResp.Token, nil
}

// CreateNote creates a new encrypted note with encrypted key
func (c *Client) CreateNote(title, encryptedContent, encryptedKey, iv string) error {
	reqBody := CreateNoteRequest{
		Title:            title,
		EncryptedContent: encryptedContent,
		EncryptedKey:     encryptedKey,
		IV:               iv,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", BaseURL+"/notes", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create note failed: %s", string(body))
	}

	return nil
}

// ListNotes retrieves all notes for the authenticated user
func (c *Client) ListNotes() ([]Note, error) {
	req, err := http.NewRequest("GET", BaseURL+"/notes", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list notes failed: %s", string(body))
	}

	// Parse response with nested notes array
	var response ListNotesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Notes, nil
}

// DeleteNote deletes a note by ID
func (c *Client) DeleteNote(id uint) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/notes/%d", BaseURL, id), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete note failed: %s", string(body))
	}

	return nil
}

// RevokeShare revokes all sharing links for a note
func (c *Client) RevokeShare(id uint) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/notes/%d/revoke", BaseURL, id), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("revoke share failed: %s", string(body))
	}

	return nil
}

// CreateShare creates a share link for a note
func (c *Client) CreateShare(id uint, durationHours int) (string, error) {
	if durationHours == 0 {
		durationHours = 24
	}

	reqBody := map[string]int{
		"duration_hours": durationHours,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/notes/%d/share", BaseURL, id), bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create share failed: %s", string(body))
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	if shareToken, ok := response["share_token"].(string); ok {
		return shareToken, nil
	}

	return "", fmt.Errorf("no share token in response")
}

// CreateShareWithMinutes creates a share link with duration in minutes (for testing)
func (c *Client) CreateShareWithMinutes(id uint, durationMinutes int) (string, error) {
	reqBody := map[string]int{
		"duration_minutes": durationMinutes,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/notes/%d/share", BaseURL, id), bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create share failed: %s", string(body))
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	if shareToken, ok := response["share_token"].(string); ok {
		return shareToken, nil
	}

	return "", fmt.Errorf("no share token in response")
}
