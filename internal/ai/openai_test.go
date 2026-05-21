package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIProvider_Review(t *testing.T) {
	issues := []Issue{
		{Severity: "error", Line: 42, Message: "SQL injection risk", Suggestion: "Use parameterized query"},
	}
	respBody, _ := json.Marshal(map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"content": string(mustMarshal(issues)),
				},
			},
		},
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("missing Authorization header")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBody)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(server.URL, "test-key", "gpt-4", 4096, 0.1)
	resp, err := provider.Review(context.Background(), &ReviewRequest{
		Language: "go",
		Filename: "main.go",
		Diff:     "--- a/main.go\n+++ b/main.go\n@@ -1 +1 @@\n-old\n+new",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(resp.Issues))
	}
	if resp.Issues[0].Severity != "error" {
		t.Errorf("severity = %q, want %q", resp.Issues[0].Severity, "error")
	}
	if resp.Issues[0].Line != 42 {
		t.Errorf("line = %d, want 42", resp.Issues[0].Line)
	}
}

func TestOpenAIProvider_EmptyResponse(t *testing.T) {
	respBody, _ := json.Marshal(map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"content": `{"issues": []}`,
				},
			},
		},
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBody)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(server.URL, "test-key", "gpt-4", 4096, 0.1)
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

func mustMarshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}
