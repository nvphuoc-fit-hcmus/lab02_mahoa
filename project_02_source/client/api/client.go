package api

import (
	"bytes"
	"crypto/ecdh"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const BaseURL = "http://localhost:8080/api"

var (
	CurrentUserID        uint
	CurrentUsername      string
	CurrentPassword      string
	AuthToken            string           // JWT Token để gọi API
	CurrentDHPrivateKey  *ecdh.PrivateKey // User's DH private key for E2EE
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
	IV               string `json:"iv"`
	EncryptedKey     string `json:"encrypted_key"`
	EncryptedKeyIV   string `json:"encrypted_key_iv"`
}

// Note represents a note from the server
type Note struct {
	ID               uint      `json:"id"`
	Title            string    `json:"title"`
	EncryptedContent string    `json:"encrypted_content"`
	IV               string    `json:"iv"`
	EncryptedKey     string    `json:"encrypted_key"`
	EncryptedKeyIV   string    `json:"encrypted_key_iv"`
	CreatedAt        time.Time `json:"created_at"`
	IsShared         bool      `json:"is_shared"`
}

// ListNotesResponse represents the response from listing notes
type ListNotesResponse struct {
	Notes []Note `json:"notes"`
	Count int    `json:"count"`
}

// SharedNote represents a note accessed via share link
type SharedNote struct {
	ID               uint      `json:"id"`
	Title            string    `json:"title"`
	EncryptedContent string    `json:"encrypted_content"`
	IV               string    `json:"iv"`
	CreatedAt        time.Time `json:"created_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	OwnerUsername    string    `json:"owner_username"`
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
func (c *Client) CreateNote(title, encryptedContent, iv, encryptedKey, encryptedKeyIV string) error {
	reqBody := CreateNoteRequest{
		Title:            title,
		EncryptedContent: encryptedContent,
		IV:               iv,
		EncryptedKey:     encryptedKey,
		EncryptedKeyIV:   encryptedKeyIV,
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

// CreateShareWithOptions creates a share link with password and/or max_access_count
func (c *Client) CreateShareWithOptions(id uint, durationHours int, password string, maxAccessCount int) (string, error) {
	if durationHours == 0 {
		durationHours = 24
	}

	reqBody := map[string]interface{}{
		"duration_hours": durationHours,
	}

	// Add optional password
	if password != "" {
		reqBody["password"] = password
	}

	// Add optional max_access_count
	if maxAccessCount > 0 {
		reqBody["max_access_count"] = maxAccessCount
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

// GetSharedNote retrieves a note via share token (with optional password)
func (c *Client) GetSharedNote(shareToken string, password string) (SharedNote, error) {
	var reqBody []byte
	var err error

	// If password provided, send it in request body
	if password != "" {
		reqData := map[string]string{"password": password}
		reqBody, err = json.Marshal(reqData)
		if err != nil {
			return SharedNote{}, err
		}
	}

	var req *http.Request
	if password != "" {
		req, err = http.NewRequest("GET", fmt.Sprintf("%s/shares/%s", BaseURL, shareToken), bytes.NewBuffer(reqBody))
	} else {
		req, err = http.NewRequest("GET", fmt.Sprintf("%s/shares/%s", BaseURL, shareToken), nil)
	}
	
	if err != nil {
		return SharedNote{}, err
	}

	if password != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return SharedNote{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		body, _ := io.ReadAll(resp.Body)
		return SharedNote{}, fmt.Errorf("unauthorized: %s", string(body))
	}

	if resp.StatusCode == http.StatusGone {
		return SharedNote{}, fmt.Errorf("share link has expired or reached maximum access count")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return SharedNote{}, fmt.Errorf("get shared note failed: %s", string(body))
	}

	var sharedNote SharedNote
	if err := json.NewDecoder(resp.Body).Decode(&sharedNote); err != nil {
		return SharedNote{}, err
	}

	return sharedNote, nil
}

// GetNote retrieves a specific note by ID
func (c *Client) GetNote(id uint) (Note, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/notes/%d", BaseURL, id), nil)
	if err != nil {
		return Note{}, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Note{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return Note{}, fmt.Errorf("get note failed: %s", string(body))
	}

	var note Note
	if err := json.NewDecoder(resp.Body).Decode(&note); err != nil {
		return Note{}, err
	}

	return note, nil
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

// E2EEShare represents an E2EE share
type E2EEShare struct {
	ID               uint      `json:"id"`
	NoteTitle        string    `json:"note_title"`
	SenderUsername   string    `json:"sender_username"`
	SenderPublicKey  string    `json:"sender_public_key"`
	EncryptedContent string    `json:"encrypted_content"`
	ContentIV        string    `json:"content_iv"`
	ExpiresAt        time.Time `json:"expires_at"`
	CreatedAt        time.Time `json:"created_at"`
}

// ListE2EESharesResponse represents the response from listing E2EE shares
type ListE2EESharesResponse struct {
	Shares []E2EEShare `json:"shares"`
	Count  int         `json:"count"`
}

// CreateE2EEShareRequest represents E2EE share creation data
type CreateE2EEShareRequest struct {
	RecipientUsername string `json:"recipient_username"`
	SenderPublicKey   string `json:"sender_public_key"`
	EncryptedContent  string `json:"encrypted_content"`
	ContentIV         string `json:"content_iv"`
	DurationHours     int    `json:"duration_hours,omitempty"`
}

// CreateE2EEShare creates an E2EE share with a specific user
func (c *Client) CreateE2EEShare(noteID uint, recipientUsername, senderPublicKey, encryptedContent, contentIV string, durationHours int) (uint, error) {
	reqBody := CreateE2EEShareRequest{
		RecipientUsername: recipientUsername,
		SenderPublicKey:   senderPublicKey,
		EncryptedContent:  encryptedContent,
		ContentIV:         contentIV,
		DurationHours:     durationHours,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/notes/%d/e2ee", BaseURL, noteID), bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("create E2EE share failed: %s", string(body))
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, err
	}

	if shareID, ok := response["share_id"].(float64); ok {
		return uint(shareID), nil
	}

	return 0, fmt.Errorf("no share ID in response")
}

// ListE2EEShares retrieves all E2EE shares received by the user
func (c *Client) ListE2EEShares() ([]E2EEShare, error) {
	req, err := http.NewRequest("GET", BaseURL+"/e2ee", nil)
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
		return nil, fmt.Errorf("list E2EE shares failed: %s", string(body))
	}

	var response ListE2EESharesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Shares, nil
}

// GetE2EEShare retrieves a specific E2EE share by ID
func (c *Client) GetE2EEShare(shareID uint) (E2EEShare, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/e2ee/%d", BaseURL, shareID), nil)
	if err != nil {
		return E2EEShare{}, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return E2EEShare{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return E2EEShare{}, fmt.Errorf("get E2EE share failed: %s", string(body))
	}

	var share E2EEShare
	if err := json.NewDecoder(resp.Body).Decode(&share); err != nil {
		return E2EEShare{}, err
	}

	return share, nil
}

// DeleteE2EEShare deletes an E2EE share (revokes sharing)
func (c *Client) DeleteE2EEShare(shareID uint) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/e2ee/%d", BaseURL, shareID), nil)
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
		return fmt.Errorf("delete E2EE share failed: %s", string(body))
	}

	return nil
}

// UpdatePublicKey updates user's DH public key on server
func (c *Client) UpdatePublicKey(publicKey string) error {
	reqBody := map[string]string{
		"dh_public_key": publicKey,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", BaseURL+"/user/publickey", bytes.NewBuffer(jsonData))
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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update public key failed: %s", string(body))
	}

	return nil
}

// GetUserPublicKey retrieves a user's DH public key
func (c *Client) GetUserPublicKey(username string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/users/%s/publickey", BaseURL, username), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("get public key failed: %s", string(body))
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	if publicKey, ok := response["dh_public_key"].(string); ok {
		return publicKey, nil
	}

	return "", fmt.Errorf("no public key found")
}
