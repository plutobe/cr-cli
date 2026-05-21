package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type openAIProvider struct {
	baseURL     string
	apiKey      string
	model       string
	maxTokens   int
	temperature float64
}

func NewOpenAIProvider(baseURL, apiKey, model string, maxTokens int, temperature float64) Provider {
	return &openAIProvider{
		baseURL:     baseURL,
		apiKey:      apiKey,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
	}
}

func (p *openAIProvider) Name() string { return "openai" }

func (p *openAIProvider) Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error) {
	systemPrompt := buildSystemPrompt(req.Rules)
	userPrompt := buildUserPrompt(req)

	body := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"max_tokens":      p.maxTokens,
		"temperature":     p.temperature,
		"response_format": map[string]string{"type": "json_object"},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respData))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respData, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(result.Choices) == 0 {
		return &ReviewResponse{}, nil
	}

	return parseIssues(result.Choices[0].Message.Content)
}

func parseIssues(content string) (*ReviewResponse, error) {
	var resp struct {
		Issues []Issue `json:"issues"`
	}
	if err := json.Unmarshal([]byte(content), &resp); err != nil {
		// Try parsing as bare array for compatibility
		var issues []Issue
		if err2 := json.Unmarshal([]byte(content), &issues); err2 != nil {
			return nil, fmt.Errorf("parse issues JSON: %w (content: %s)", err, content)
		}
		return &ReviewResponse{Issues: issues}, nil
	}
	return &ReviewResponse{Issues: resp.Issues}, nil
}
