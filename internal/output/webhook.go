package output

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type WebhookSender struct {
	url    string
	whType string
	secret string
}

func NewWebhookSender(url, whType, secret string) *WebhookSender {
	return &WebhookSender{url: url, whType: whType, secret: secret}
}

func (s *WebhookSender) Send(result *ReviewResult, repo, commit string, ts time.Time) error {
	markdown := s.buildMarkdown(result, repo, commit, ts)

	var payload interface{}
	switch s.whType {
	case "dingtalk":
		payload = map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"title": "代码审查报告",
				"text":  markdown,
			},
		}
	case "wecom":
		payload = map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"content": markdown,
			},
		}
	case "feishu":
		payload = map[string]interface{}{
			"msg_type": "interactive",
			"card": map[string]interface{}{
				"header": map[string]interface{}{
					"title": map[string]string{"content": "代码审查报告", "tag": "plain_text"},
				},
				"elements": []interface{}{
					map[string]interface{}{"tag": "markdown", "content": markdown},
				},
			},
		}
	default:
		payload = map[string]interface{}{
			"text":      markdown,
			"repo":      repo,
			"commit":    commit,
			"timestamp": ts.Unix(),
			"errors":    result.ErrorCount,
			"warnings":  result.WarningCount,
		}
	}

	return s.sendPayload(payload)
}

func (s *WebhookSender) sendPayload(payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	webhookURL := s.url
	if s.secret != "" {
		webhookURL = s.signURL(webhookURL, s.secret)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook error (status %d): %s", resp.StatusCode, string(body))
	}
	return nil
}

func (s *WebhookSender) signURL(webhookURL, secret string) string {
	timestamp := time.Now().UnixMilli()
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))

	u, err := url.Parse(webhookURL)
	if err != nil {
		return webhookURL
	}
	q := u.Query()
	q.Set("timestamp", fmt.Sprintf("%d", timestamp))
	q.Set("sign", sign)
	u.RawQuery = q.Encode()
	return u.String()
}

func (s *WebhookSender) buildMarkdown(result *ReviewResult, repo, commit string, ts time.Time) string {
	var b strings.Builder
	b.WriteString("## 代码审查报告\n\n")
	fmt.Fprintf(&b, "**时间**: %s\n", ts.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&b, "**仓库**: %s\n", repo)
	fmt.Fprintf(&b, "**提交**: %s\n\n", commit)

	if result.ErrorCount > 0 || result.WarningCount > 0 {
		b.WriteString("### 发现问题\n\n")
		b.WriteString("| 严重程度 | 文件 | 行号 | 问题 |\n")
		b.WriteString("|---------|------|------|------|\n")
		for _, f := range result.Files {
			for _, issue := range f.Issues {
				icon := "🔵"
				if issue.Severity == "error" {
					icon = "🔴"
				} else if issue.Severity == "warning" {
					icon = "🟡"
				}
				fmt.Fprintf(&b, "| %s %s | %s | %d | %s |\n",
					icon, strings.ToUpper(issue.Severity), f.Filename, issue.Line, issue.Message)
			}
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(&b, "**统计**: %d 个错误, %d 个警告, %d 个提示\n",
		result.ErrorCount, result.WarningCount, result.InfoCount)
	return b.String()
}
