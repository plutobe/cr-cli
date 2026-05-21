package output

import (
	"bytes"
	"strings"
	"testing"

	"cr-cli/internal/ai"
)

func TestTerminalRenderer_Render(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewTerminalRenderer(&buf, false) // no color for testing

	result := &ReviewResult{
		Files: []FileResult{
			{
				Filename: "main.go",
				Language: "go",
				Issues: []ai.Issue{
					{Severity: "error", Line: 42, Message: "SQL injection", Suggestion: "Use parameterized query"},
					{Severity: "warning", Line: 10, Message: "missing error check", Suggestion: "add err != nil"},
				},
			},
		},
		TotalIssues:  2,
		ErrorCount:   1,
		WarningCount: 1,
		InfoCount:    0,
	}

	renderer.Render(result)
	output := buf.String()

	if !strings.Contains(output, "main.go") {
		t.Error("output should contain filename")
	}
	if !strings.Contains(output, "SQL injection") {
		t.Error("output should contain issue message")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("output should contain severity")
	}
	if !strings.Contains(output, "1 个错误") {
		t.Error("output should contain error count")
	}
}

func TestTerminalRenderer_EmptyResult(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewTerminalRenderer(&buf, false)

	result := &ReviewResult{}
	renderer.Render(result)
	output := buf.String()

	if !strings.Contains(output, "0 个错误") {
		t.Error("output should show zero counts")
	}
}
