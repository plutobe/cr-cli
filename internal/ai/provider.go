package ai

import (
	"context"
	"encoding/json"
	"fmt"
)

type Provider interface {
	Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error)
	Name() string
}

type ReviewRequest struct {
	Language string
	Filename string
	Diff     string
	Context  string
	Rules    []string
}

type ReviewResponse struct {
	Issues []Issue
}

type Issue struct {
	Severity   string `json:"severity"`
	Line       int    `json:"line"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
}

func parseIssues(content string) (*ReviewResponse, error) {
	var resp struct {
		Issues []Issue `json:"issues"`
	}
	if err := json.Unmarshal([]byte(content), &resp); err != nil {
		var issues []Issue
		if err2 := json.Unmarshal([]byte(content), &issues); err2 != nil {
			return nil, fmt.Errorf("parse issues JSON: %w (content: %s)", err, content)
		}
		return &ReviewResponse{Issues: issues}, nil
	}
	return &ReviewResponse{Issues: resp.Issues}, nil
}
