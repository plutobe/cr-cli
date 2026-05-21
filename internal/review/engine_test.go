package review

import (
	"context"
	"testing"

	"cr-cli/internal/ai"
	"cr-cli/internal/git"
)

type mockProvider struct {
	responses []*ai.ReviewResponse
	callCount int
}

func (m *mockProvider) Name() string { return "mock" }

func (m *mockProvider) Review(ctx context.Context, req *ai.ReviewRequest) (*ai.ReviewResponse, error) {
	if m.callCount < len(m.responses) {
		resp := m.responses[m.callCount]
		m.callCount++
		return resp, nil
	}
	m.callCount++
	return &ai.ReviewResponse{}, nil
}

func TestEngine_ReviewBatches(t *testing.T) {
	files := []git.FileChange{
		{Filename: "a.go", Language: "go", Diff: "+change1"},
		{Filename: "b.go", Language: "go", Diff: "+change2"},
		{Filename: "c.go", Language: "go", Diff: "+change3"},
	}

	provider := &mockProvider{
		responses: []*ai.ReviewResponse{
			{Issues: []ai.Issue{{Severity: "error", Line: 1, Message: "issue1"}}},
			{Issues: []ai.Issue{{Severity: "warning", Line: 2, Message: "issue2"}}},
			{Issues: nil},
		},
	}

	engine := NewEngine(provider, Options{BatchSize: 2})
	result, err := engine.Review(context.Background(), files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.FileResults) != 3 {
		t.Errorf("expected 3 file results, got %d", len(result.FileResults))
	}
	if len(result.FileResults[0].Issues) != 1 {
		t.Errorf("expected 1 issue in first file, got %d", len(result.FileResults[0].Issues))
	}
	if result.TotalFiles != 3 {
		t.Errorf("total files = %d, want 3", result.TotalFiles)
	}
	if result.Batches != 2 {
		t.Errorf("batches = %d, want 2", result.Batches)
	}
}

func TestEngine_DefaultBatchSize(t *testing.T) {
	provider := &mockProvider{}
	engine := NewEngine(provider, Options{})
	if engine.options.BatchSize != 10 {
		t.Errorf("default batch size = %d, want 10", engine.options.BatchSize)
	}
}

func TestEngine_EmptyFiles(t *testing.T) {
	provider := &mockProvider{}
	engine := NewEngine(provider, Options{BatchSize: 5})
	result, err := engine.Review(context.Background(), []git.FileChange{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalFiles != 0 {
		t.Errorf("total files = %d, want 0", result.TotalFiles)
	}
	if result.Batches != 0 {
		t.Errorf("batches = %d, want 0", result.Batches)
	}
}
