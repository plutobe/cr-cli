package ai

import "context"

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
