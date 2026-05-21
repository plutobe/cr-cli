package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnthropicProvider_Review(t *testing.T) {
	issues := []Issue{
		{Severity: "warning", Line: 10, Message: "error handling missing", Suggestion: "add err check"},
	}
	content, _ := json.Marshal(issues)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") == "" {
			t.Error("missing x-api-key header")
		}
		if r.Header.Get("anthropic-version") == "" {
			t.Error("missing anthropic-version header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": string(content)},
			},
		})
	}))
	defer server.Close()

	provider := NewAnthropicProvider(server.URL+"/v1", "test-key", "claude-sonnet-4-20250514", 4096, 0.1)
	resp, err := provider.Review(context.Background(), &ReviewRequest{
		Language: "go",
		Filename: "main.go",
		Diff:     "+new line",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(resp.Issues))
	}
	if resp.Issues[0].Severity != "warning" {
		t.Errorf("severity = %q, want %q", resp.Issues[0].Severity, "warning")
	}
	if resp.Issues[0].Line != 10 {
		t.Errorf("line = %d, want 10", resp.Issues[0].Line)
	}
}

func TestAnthropicProvider_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": `{"issues": []}`},
			},
		})
	}))
	defer server.Close()

	provider := NewAnthropicProvider(server.URL+"/v1", "test-key", "claude-sonnet-4-20250514", 4096, 0.1)
	resp, err := provider.Review(context.Background(), &ReviewRequest{
		Language: "go",
		Filename: "main.go",
		Diff:     "+no issues",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(resp.Issues))
	}
}
