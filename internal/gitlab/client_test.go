package gitlab

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMergeRequestDiff(t *testing.T) {
	changes := []map[string]interface{}{
		{
			"old_path": "main.go",
			"new_path": "main.go",
			"new_file": false,
			"diff":     "@@ -1 +1 @@\n-old\n+new",
		},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("PRIVATE-TOKEN") == "" {
			t.Error("missing PRIVATE-TOKEN header")
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"changes": changes,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	diff, err := client.GetMergeRequestDiff("group/project", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff == "" {
		t.Error("expected non-empty diff")
	}
}

func TestGetMergeRequestDiff_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"404 Not Found"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	_, err := client.GetMergeRequestDiff("group/project", 999)
	if err == nil {
		t.Error("expected error for 404 response")
	}
}
