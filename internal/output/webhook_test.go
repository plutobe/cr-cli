package output

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cr-cli/internal/ai"
)

func TestWebhookSender_DingTalk(t *testing.T) {
	var receivedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewWebhookSender(server.URL, "dingtalk", "")
	result := &ReviewResult{
		Files: []FileResult{
			{
				Filename: "main.go",
				Issues: []ai.Issue{
					{Severity: "error", Line: 1, Message: "test issue", Suggestion: "fix it"},
				},
			},
		},
		ErrorCount: 1,
	}

	err := sender.Send(result, "test-repo", "abc123", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedBody["msgtype"] != "markdown" {
		t.Errorf("msgtype = %v, want markdown", receivedBody["msgtype"])
	}
}

func TestWebhookSender_Wecom(t *testing.T) {
	var receivedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewWebhookSender(server.URL, "wecom", "")
	result := &ReviewResult{ErrorCount: 0}

	err := sender.Send(result, "test-repo", "abc123", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedBody["msgtype"] != "markdown" {
		t.Errorf("msgtype = %v, want markdown", receivedBody["msgtype"])
	}
}

func TestWebhookSender_Custom(t *testing.T) {
	var receivedBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewWebhookSender(server.URL, "custom", "")
	result := &ReviewResult{ErrorCount: 1, WarningCount: 2}

	err := sender.Send(result, "test-repo", "abc123", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedBody["errors"].(float64) != 1 {
		t.Errorf("errors = %v, want 1", receivedBody["errors"])
	}
}
