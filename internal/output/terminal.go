package output

import (
	"fmt"
	"io"
	"strings"

	"cr-cli/internal/ai"
	"cr-cli/internal/i18n"
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
	fmt.Fprintf(r.w, "║          %s          ║\n", i18n.T("CR-CLI 代码审查报告", "CR-CLI Code Review Report"))
	fmt.Fprintln(r.w, "╚══════════════════════════════════════════════════════════════╝")
	fmt.Fprintln(r.w)

	for _, file := range result.Files {
		fmt.Fprintf(r.w, "📁 %s: %s (%s)\n", i18n.T("文件", "File"), file.Filename, file.Language)
		fmt.Fprintln(r.w, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Fprintln(r.w)

		for _, issue := range file.Issues {
			icon := r.severityIcon(issue.Severity)
			fmt.Fprintf(r.w, "  %s %s %d [%s] %s\n", icon, i18n.T("第", "Line"), issue.Line, strings.ToUpper(issue.Severity), issue.Message)
			fmt.Fprintf(r.w, "    %s %s\n", i18n.T("💡 建议:", "💡 Suggestion:"), issue.Suggestion)
			fmt.Fprintln(r.w)
		}
	}

	fmt.Fprintln(r.w, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Fprintf(r.w, "📊 %s: %d %s, %d %s, %d %s\n",
		i18n.T("统计", "Stats"),
		result.ErrorCount, i18n.T("个错误", "error(s)"),
		result.WarningCount, i18n.T("个警告", "warning(s)"),
		result.InfoCount, i18n.T("个提示", "info"))
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
