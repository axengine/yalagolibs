package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func SendTelegramMessage(ctx context.Context, botToken, chatID, message string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	// Create form data payload
	formData := url.Values{
		"chat_id": {chatID},
		"text":    {message},
	}

	// Create context-aware request
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		apiURL,
		strings.NewReader(formData.Encode()),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set content-type header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Configure HTTP client with safety timeout
	client := &http.Client{
		Timeout: 30 * time.Second, // Fallback timeout protection
	}

	// Execute HTTP request
	resp, err := client.Do(req)
	if err != nil {
		// Check context cancellation
		if ctx.Err() != nil {
			return fmt.Errorf("request canceled: %w", ctx.Err())
		}
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status: %d, response: %s",
			resp.StatusCode, string(body))
	}

	// Decode JSON response
	var apiResponse struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return fmt.Errorf("JSON decode failed: %w", err)
	}

	if !apiResponse.OK {
		return fmt.Errorf("telegram API error: %s", apiResponse.Description)
	}

	return nil
}
