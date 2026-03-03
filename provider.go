package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// chatRequest is the request body for the OpenAI-compatible chat completions API.
type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// chatResponse is the response body from the chat completions API.
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// errorResponse represents an API error response body.
type errorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// callLLM sends the messages to the configured LLM provider and returns the response text.
func callLLM(cfg *Config, messages []Message) (string, error) {
	reqBody := chatRequest{
		Model:    cfg.Provider.Model,
		Messages: messages,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	url := strings.TrimRight(cfg.Provider.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add auth header for providers that require it
	if cfg.Provider.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.Provider.APIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not reach %s. Check your connection", cfg.Provider.BaseURL)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// Handle HTTP errors
	if resp.StatusCode != http.StatusOK {
		return "", handleHTTPError(resp.StatusCode, respBody, cfg.Provider.BaseURL)
	}

	// Parse successful response
	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("empty response from API (no choices returned)")
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}

// handleHTTPError returns a user-friendly error for known HTTP status codes.
func handleHTTPError(status int, body []byte, baseURL string) error {
	switch status {
	case http.StatusUnauthorized:
		return fmt.Errorf("authentication failed. Check your API key")
	case http.StatusTooManyRequests:
		return fmt.Errorf("rate limited. Try again later")
	default:
		// Try to extract error message from response body
		var errResp errorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			return fmt.Errorf("API error (HTTP %d): %s", status, errResp.Error.Message)
		}
		return fmt.Errorf("API error (HTTP %d)", status)
	}
}
