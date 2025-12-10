package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AttachmentField represents a field in a Slack attachment.
// It contains a title, a value, and a boolean indicating if it's short.
type AttachmentField struct {
	Title string `json:"title,omitempty"`
	Value string `json:"value,omitempty"`
	Short bool   `json:"short,omitempty"`
}

// Attachment represents a Slack message attachment.
// It can contain a color, fields, a footer, and a timestamp.
type Attachment struct {
	Color  string            `json:"color,omitempty"`
	Fields []AttachmentField `json:"fields,omitempty"`
	Footer string            `json:"footer,omitempty"`
	Ts     int64             `json:"ts,omitempty"`
}

// AlertMessage represents a Slack alert message.
// It contains the main text and a list of attachments.
type AlertMessage struct {
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// SendSlackAlert sends a structured alert message to a Slack webhook URL.
// It takes a context, the webhook URL, and an AlertMessage struct as input.
// It returns an error if the request fails or if the Slack API returns an error.
func SendSlackAlert(ctx context.Context, webhookURL string, message AlertMessage) error {
	// Marshal the message struct to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create a new HTTP request with the context
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set the Content-Type header to application/json
	req.Header.Set("Content-Type", "application/json")

	// Create an HTTP client with a timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}