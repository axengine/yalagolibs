package slack

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendSlackAlert(t *testing.T) {
	// Create a mock server to handle the Slack API request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request method is POST
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check if the Content-Type header is set to application/json
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		// Unmarshal the request body into an AlertMessage struct
		var receivedMsg AlertMessage
		if err := json.Unmarshal(body, &receivedMsg); err != nil {
			t.Fatalf("Failed to unmarshal request body: %v", err)
		}

		// Check the content of the received message
		expectedText := "⚠️ *[P1]* AI service interface request P95 data anomaly increased by 123%"
		if receivedMsg.Text != expectedText {
			t.Errorf("Expected text '%s', got '%s'", expectedText, receivedMsg.Text)
		}

		// Check the number of attachments
		if len(receivedMsg.Attachments) != 1 {
			t.Fatalf("Expected 1 attachment, got %d", len(receivedMsg.Attachments))
		}

		// Respond with a 200 OK status
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a sample alert message
	alert := AlertMessage{
		Text: "⚠️ *[P1]* AI service interface request P95 data anomaly increased by 123%",
		Attachments: []Attachment{
			{
				Color: "#ff0000",
				Fields: []AttachmentField{
					{Title: "📛 Alarm Service", Value: "yala-ai", Short: true},
					{Title: "⏰ Alarm Time", Value: "2025-07-25 14:32:10 (UTC+8)", Short: true},
					{Title: "💼 Alarm Setter", Value: "Lee", Short: true},
					{Title: "🛡️ Responsible Person", Value: "@Lee", Short: true},
					{Title: "📋 Processing Suggestions", Value: "TEST-Please restart the service and check the error log.- TEST", Short: false},
				},
				Footer: "Monitoring and Early Warning BOT",
				Ts:     1721889130,
			},
		},
	}

	// Send the alert using the function to be tested
	err := SendSlackAlert(context.Background(), server.URL, alert)
	if err != nil {
		t.Fatalf("SendSlackAlert failed: %v", err)
	}
}
