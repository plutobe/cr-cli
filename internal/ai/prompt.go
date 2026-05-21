package ai

import (
	"fmt"
	"strings"
)

func buildSystemPrompt(rules []string) string {
	base := `你是一个专业的代码审查专家。请审查以下代码变更。

审查规则：
1. 安全性：SQL注入、XSS、敏感信息泄露、不安全的加密等
2. 代码质量：代码重复、复杂度过高、命名不规范等
3. 性能：内存泄漏、goroutine泄漏、不必要的拷贝等
4. 最佳实践：错误处理、日志规范、并发安全等
5. 可读性：注释缺失、逻辑复杂、魔法数字等`

	if len(rules) > 0 {
		base += "\n\n额外规则：\n"
		for i, r := range rules {
			base += fmt.Sprintf("%d. %s\n", i+1, r)
		}
	}

	base += `

输出格式（JSON对象，包含issues数组）：
{"issues":[{"severity":"error|warning|info","line":行号,"message":"问题描述","suggestion":"修复建议"}]}

如果没有发现问题，返回 {"issues": []}`
	return base
}

func buildUserPrompt(req *ReviewRequest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "文件：%s\n", req.Filename)
	fmt.Fprintf(&b, "语言：%s\n", req.Language)
	fmt.Fprintf(&b, "变更内容：\n%s\n", req.Diff)
	if req.Context != "" {
		fmt.Fprintf(&b, "上下文：\n%s\n", req.Context)
	}
	return b.String()
}
