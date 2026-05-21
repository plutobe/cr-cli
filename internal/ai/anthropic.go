package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type anthropicProvider struct {
	baseURL     string
	apiKey      string
	model       string
	maxTokens   int
	temperature float64
	debug       bool
}

func NewAnthropicProvider(baseURL, apiKey, model string, maxTokens int, temperature float64, debug bool) Provider {
	return &anthropicProvider{
		baseURL:     baseURL,
		apiKey:      apiKey,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
		debug:       debug,
	}
}

func (p *anthropicProvider) Name() string { return "anthropic" }

func (p *anthropicProvider) Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error) {
	systemPrompt := buildSystemPrompt(req.Rules)
	userPrompt := buildUserPrompt(req)

	body := map[string]interface{}{
		"model":       p.model,
		"max_tokens":  p.maxTokens,
		"temperature": p.temperature,
		"system":      systemPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
	}

	jsonBody, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	if p.debug {
		fmt.Fprintf(os.Stderr, "\n========== REQUEST [%s] ==========\n", p.Name())
		fmt.Fprintf(os.Stderr, "POST %s/messages\n", p.baseURL)
		fmt.Fprintf(os.Stderr, "%s\n", string(jsonBody))
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if p.debug {
		fmt.Fprintf(os.Stderr, "\n========== RESPONSE [%s] (status %d) ==========\n", p.Name(), resp.StatusCode)
		var pretty json.RawMessage
		if json.Unmarshal(respData, &pretty) == nil {
			out, _ := json.MarshalIndent(pretty, "", "  ")
			fmt.Fprintf(os.Stderr, "%s\n", string(out))
		} else {
			fmt.Fprintf(os.Stderr, "%s\n", string(respData))
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respData))
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respData, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(result.Content) == 0 {
		return &ReviewResponse{}, nil
	}

	return parseIssues(result.Content[0].Text)
}
