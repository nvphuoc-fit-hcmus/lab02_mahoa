package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const BaseURL = "http://localhost:8080/api"

type APIClient struct {
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
}

// Note represents a note from the server
type Note struct {
	ID               uint   `json:"id"`
	Title            string `json:"title"`
	EncryptedContent string `json:"encrypted_content"`
	IV               string `json:"iv"`
}

// Register creates a new user account
func (c *APIClient) Register(username, password string) error {
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
func (c *APIClient) Login(username, password string) (string, error) {
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

// CreateNote creates a new encrypted note
func (c *APIClient) CreateNote(title, encryptedContent, iv string) error {
	reqBody := CreateNoteRequest{
		Title:            title,
		EncryptedContent: encryptedContent,
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
func (c *APIClient) ListNotes() ([]Note, error) {
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

	var notes []Note
	if err := json.NewDecoder(resp.Body).Decode(&notes); err != nil {
		return nil, err
	}

	return notes, nil
}

// DeleteNote deletes a note by ID
func (c *APIClient) DeleteNote(id uint) error {
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
