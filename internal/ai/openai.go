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

type openAIProvider struct {
	baseURL     string
	apiKey      string
	model       string
	maxTokens   int
	temperature float64
	debug       bool
}

func NewOpenAIProvider(baseURL, apiKey, model string, maxTokens int, temperature float64, debug bool) Provider {
	return &openAIProvider{
		baseURL:     baseURL,
		apiKey:      apiKey,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
		debug:       debug,
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

	jsonBody, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	if p.debug {
		fmt.Fprintf(os.Stderr, "\n========== REQUEST [%s] ==========\n", p.Name())
		fmt.Fprintf(os.Stderr, "POST %s/chat/completions\n", p.baseURL)
		fmt.Fprintf(os.Stderr, "%s\n", string(jsonBody))
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
