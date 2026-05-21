package output

import (
	"fmt"
	"io"
	"strings"

	"cr-cli/internal/ai"
)

type ReviewResult struct {
	Files        []FileResult
	TotalIssues  int
	ErrorCount   int
	WarningCount int
	InfoCount    int
}

type FileResult struct {
	Filename string
	Language string
	Issues   []ai.Issue
}

type TerminalRenderer struct {
	w     io.Writer
	color bool
}

func NewTerminalRenderer(w io.Writer, color bool) *TerminalRenderer {
	return &TerminalRenderer{w: w, color: color}
}

func (r *TerminalRenderer) Render(result *ReviewResult) {
	fmt.Fprintln(r.w, "╔══════════════════════════════════════════════════════════════╗")
	fmt.Fprintln(r.w, "║                    CR-CLI 代码审查报告                       ║")
	fmt.Fprintln(r.w, "╚══════════════════════════════════════════════════════════════╝")
	fmt.Fprintln(r.w)

	for _, file := range result.Files {
		fmt.Fprintf(r.w, "📁 文件: %s (%s)\n", file.Filename, file.Language)
		fmt.Fprintln(r.w, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Fprintln(r.w)

		for _, issue := range file.Issues {
			icon := r.severityIcon(issue.Severity)
			fmt.Fprintf(r.w, "  %s 第 %d 行 [%s] %s\n", icon, issue.Line, strings.ToUpper(issue.Severity), issue.Message)
			fmt.Fprintf(r.w, "    %s\n", issue.Suggestion)
			fmt.Fprintln(r.w)
		}
	}

	fmt.Fprintln(r.w, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Fprintf(r.w, "📊 统计: %d 个错误, %d 个警告, %d 个提示\n",
		result.ErrorCount, result.WarningCount, result.InfoCount)
}

func (r *TerminalRenderer) severityIcon(severity string) string {
	switch severity {
	case "error":
		return "✗"
	case "warning":
		return "⚠"
	case "info":
		return "ℹ"
	default:
		return "•"
	}
}
