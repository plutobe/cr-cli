package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	content = extractJSON(content)

	var resp struct {
		Issues []Issue `json:"issues"`
	}
	if err := json.Unmarshal([]byte(content), &resp); err != nil {
		repaired := repairJSON(content)
		if err2 := json.Unmarshal([]byte(repaired), &resp); err2 == nil {
			return &ReviewResponse{Issues: resp.Issues}, nil
		}
		var issues []Issue
		if err3 := json.Unmarshal([]byte(repaired), &issues); err3 == nil {
			return &ReviewResponse{Issues: issues}, nil
		}
		return nil, fmt.Errorf("parse issues JSON: %w (content: %s)", err, content)
	}
	return &ReviewResponse{Issues: resp.Issues}, nil
}

// extractJSON extracts the first JSON object or array from content.
func extractJSON(content string) string {
	content = strings.TrimSpace(content)
	// Try to find JSON block in markdown code fence
	if idx := strings.Index(content, "```"); idx >= 0 {
		rest := content[idx+3:]
		if nl := strings.Index(rest, "\n"); nl >= 0 {
			rest = rest[nl+1:]
		}
		if end := strings.Index(rest, "```"); end >= 0 {
			content = strings.TrimSpace(rest[:end])
		}
	}
	// Find first { or [
	for i, r := range content {
		if r == '{' || r == '[' {
			return content[i:]
		}
	}
	return content
}

// repairJSON attempts to fix common JSON issues from AI responses.
func repairJSON(s string) string {
	// Fix trailing commas before ] or }
	s = fixTrailingCommas(s)

	// Try parsing as-is first
	if json.Valid([]byte(s)) {
		return s
	}

	// Try to fix unescaped quotes inside JSON string values.
	// Strategy: walk through the JSON, track string boundaries, and escape
	// stray quotes inside string values.
	fixed, ok := fixUnescapedQuotes(s)
	if ok && json.Valid([]byte(fixed)) {
		return fixed
	}

	return s
}

func fixTrailingCommas(s string) string {
	var result []byte
	inString := false
	escaped := false

	for i := 0; i < len(s); i++ {
		c := s[i]

		if escaped {
			result = append(result, c)
			escaped = false
			continue
		}

		if c == '\\' && inString {
			result = append(result, c)
			escaped = true
			continue
		}

		if c == '"' {
			inString = !inString
			result = append(result, c)
			continue
		}

		if inString {
			result = append(result, c)
			continue
		}

		// Not in string - check for trailing comma
		if c == ',' {
			// Look ahead for ] or }
			j := i + 1
			for j < len(s) && (s[j] == ' ' || s[j] == '\t' || s[j] == '\n' || s[j] == '\r') {
				j++
			}
			if j < len(s) && (s[j] == ']' || s[j] == '}') {
				continue // skip trailing comma
			}
		}

		result = append(result, c)
	}

	return string(result)
}

// fixUnescapedQuotes tries to escape unescaped double quotes inside JSON string values.
func fixUnescapedQuotes(s string) (string, bool) {
	var result strings.Builder
	inString := false
	escaped := false
	inValue := false // true after colon, inside the value string

	for i := 0; i < len(s); i++ {
		c := s[i]

		if escaped {
			result.WriteByte(c)
			escaped = false
			continue
		}

		if c == '\\' && inString {
			result.WriteByte(c)
			escaped = true
			continue
		}

		if c == '"' {
			if !inString {
				// Opening quote
				inString = true
				inValue = false
				result.WriteByte(c)
				continue
			}

			// We're in a string and hit a quote. Is this the closing quote?
			// Look ahead: if next non-whitespace is : , ] } then it's likely a closing quote.
			j := i + 1
			for j < len(s) && (s[j] == ' ' || s[j] == '\t' || s[j] == '\n' || s[j] == '\r') {
				j++
			}
			if j < len(s) && (s[j] == ':' || s[j] == ',' || s[j] == ']' || s[j] == '}') {
				// This is the closing quote
				inString = false
				result.WriteByte(c)
				continue
			}

			// If we're right after a : and this is the opening value quote
			if !inValue {
				inValue = true
				result.WriteByte(c)
				continue
			}

			// Check if this could be an apostrophe-like usage (e.g., "don't")
			// or if the next char is a letter/number (likely embedded quote)
			// Heuristic: if the character after this quote is alphanumeric or $ or {,
			// and the previous character is also alphanumeric or special,
			// then this is an unescaped quote inside a value.
			if i+1 < len(s) && i > 0 {
				next := s[i+1]
				prev := s[i-1]
				if (isAlphaNumOrSpecial(prev) && isAlphaNumOrSpecial(next)) ||
					(prev == ' ' && next == ' ') {
					// Likely an unescaped quote - escape it
					result.WriteString("\\\"")
					continue
				}
			}

			// Default: treat as closing quote
			inString = false
			result.WriteByte(c)
			continue
		}

		result.WriteByte(c)
	}

	return result.String(), true
}

func isAlphaNumOrSpecial(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
		c == '_' || c == '-' || c == '.' || c == '/' || c == '@' || c == '$' || c == '{' || c == '}'
}
