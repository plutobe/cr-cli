# CR-CLI

AI-powered code review CLI tool. Supports OpenAI and Anthropic providers. Integrates with GitLab CI.

[中文文档](README.zh-CN.md)

## Features

- **Dual AI Providers**: OpenAI and Anthropic API, switch freely
- **Self-hosted**: Custom API endpoint for enterprise intranet
- **Smart Detection**: Auto-detect unstaged/staged/latest commit changes
- **GitLab MR**: Review specific Merge Requests
- **Comprehensive Review**: Security, code quality, performance, best practices, readability
- **Batch Processing**: Auto-batch review for large changesets (default 10 files per batch)
- **Webhook Notifications**: DingTalk, WeCom, Feishu
- **CI Integration**: Non-zero exit code in blocking mode when issues found

## Installation

### Build from source

```bash
git clone <repo-url>
cd cr-cli
make build
```

### Install to system path

```bash
make install
```

### Cross-platform compilation

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o cr-cli-linux-amd64 .

# Windows
GOOS=windows GOARCH=amd64 go build -o cr-cli-windows-amd64.exe .

# macOS ARM
GOOS=darwin GOARCH=arm64 go build -o cr-cli-darwin-arm64 .
```

## Quick Start

### 1. Use directly

Run in any Git project directory:

```bash
cr-cli review
```

On first run, an interactive setup guides you through configuration. Only API Key is required, everything else has sensible defaults:

```
Please configure AI Provider:

  Provider type (openai/anthropic) [openai]:
  Base URL [https://api.openai.com/v1]:
  API Key: sk-xxx
  Model [gpt-4]:

Config saved to ~/.cr-cli/config.yaml
```

Config is saved to `~/.cr-cli/config.yaml` and shared across all projects.

### 2. Run review

```bash
# Auto-detect: review uncommitted changes, or latest commit if clean
cr-cli review

# Review a specific commit
cr-cli review --commit abc1234

# Review branch diff against base
cr-cli review --branch feature/login --base main

# Review GitLab MR
cr-cli review --mr 42

# Debug mode (print request/response payloads)
cr-cli --debug review
```

## Commands

### cr-cli review

Review code changes.

```bash
cr-cli review [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--commit` | string | - | Review specific commit (SHA or ref) |
| `--mr` | int | - | Review specific GitLab MR |
| `--branch` | string | - | Review specific branch |
| `--base` | string | `main` | Base branch for comparison |
| `--batch-size` | int | `10` | Files per review batch |
| `--severity` | string | `warning` | Review mode: blocking / warning |
| `--config` | string | - | Config file path |
| `--debug` | bool | `false` | Print debug info (request/response) |

**Auto-detect priority:**

1. Unstaged changes (`git diff`)
2. Staged changes (`git diff --cached`)
3. Most recent commit with code changes (skips binary-only commits)

### cr-cli config show

Show current loaded config.

```bash
cr-cli config show
```

## Configuration

### Config file lookup order

1. `--config` flag path
2. `./cr-cli.yaml`
3. `./cr-cli.yml`
4. `~/.cr-cli/config.yaml` (global, recommended)

On first run of `cr-cli review`, if no config exists, interactive setup is triggered and saved to `~/.cr-cli/config.yaml`.

### Environment variables

Config files support `${ENV_VAR}` syntax:

```yaml
provider:
  api_key: ${OPENAI_API_KEY}
gitlab:
  token: ${GITLAB_TOKEN}
```

### Review modes

| Mode | Description | Exit code |
|------|-------------|-----------|
| `warning` (default) | Report only, no exit code impact | Always 0 |
| `blocking` | Exit code 1 when error-level issues found | 0 or 1 |

### Webhook types

| Type | Description |
|------|-------------|
| `dingtalk` | DingTalk robot |
| `wecom` | WeCom robot |
| `feishu` | Feishu robot |
| `custom` | Generic JSON format |

## GitLab CI Integration

### .gitlab-ci.yml example

```yaml
stages:
  - review

code-review:
  stage: review
  image: golang:1.21
  before_script:
    - apt-get update && apt-get install -y git
  script:
    - make build
    - ./cr-cli review --severity blocking
  variables:
    OPENAI_API_KEY: $OPENAI_API_KEY
  only:
    - merge_requests
```

### As pipeline gate

```yaml
code-review-gate:
  stage: review
  script:
    - ./cr-cli review --severity blocking
  allow_failure: false
```

## Output Example

### Terminal output

```
╔══════════════════════════════════════════════════════════════╗
║                    CR-CLI Code Review Report                 ║
╚══════════════════════════════════════════════════════════════╝

📁 File: main.go (Go)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  ✗ Line 42 [ERROR] SQL Injection Risk
    User input directly concatenated into SQL statement
    💡 Suggestion: Use parameterized queries instead of string concatenation

  ⚠ Line 78 [WARNING] Incomplete Error Handling
    err return value ignored
    💡 Suggestion: Add if err != nil check

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 Stats: 1 error, 1 warning, 0 info
```

### Webhook Markdown notification

```markdown
## 🔍 Code Review Report

**Time**: 2024-01-15 14:30:00
**Repo**: /path/to/repo
**Commit**: HEAD

### Issues Found

| Severity | File | Line | Issue |
|----------|------|------|-------|
| 🔴 ERROR | main.go | 42 | SQL Injection Risk |
| 🟡 WARNING | main.go | 78 | Incomplete Error Handling |

**Stats**: 1 error, 1 warning, 0 info
```

## Project Structure

```
cr-cli/
├── cmd/                        # Cobra CLI commands
│   ├── root.go                # Root command
│   ├── review.go              # review subcommand
│   └── config.go              # config subcommand (with interactive setup)
├── internal/
│   ├── config/                # Config loading
│   │   ├── config.go          # YAML parsing + env vars + defaults
│   │   └── config_test.go
│   ├── git/                   # Git operations
│   │   ├── diff.go            # Diff retrieval (auto-detect/commit/branch)
│   │   ├── diff_test.go
│   │   ├── changeset.go       # Changeset parsing + language detection
│   │   └── changeset_test.go
│   ├── gitlab/                # GitLab API
│   │   ├── client.go          # MR diff retrieval
│   │   └── client_test.go
│   ├── ai/                    # AI Provider
│   │   ├── provider.go        # Interface definition
│   │   ├── prompt.go          # Prompt templates
│   │   ├── openai.go          # OpenAI implementation
│   │   ├── openai_test.go
│   │   ├── anthropic.go       # Anthropic implementation
│   │   └── anthropic_test.go
│   ├── review/                # Review engine
│   │   ├── engine.go          # Batch processing + result aggregation
│   │   └── engine_test.go
│   └── output/                # Output
│       ├── terminal.go        # Terminal colored output
│       ├── terminal_test.go
│       ├── webhook.go         # Webhook notifications
│       └── webhook_test.go
├── main.go                    # Entry point
├── go.mod
├── go.sum
└── Makefile
```

## Development

### Requirements

- Go 1.21+
- Git

### Workflow

```bash
git clone <repo-url>
cd cr-cli
go mod tidy
make test
make build
./cr-cli --help
```

### Adding a new AI Provider

1. Create a new file under `internal/ai/`, e.g. `deepseek.go`
2. Implement the `Provider` interface:

```go
type Provider interface {
    Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error)
    Name() string
}
```

3. Register in `cmd/review.go` `createProvider` function:

```go
case "deepseek":
    return ai.NewDeepSeekProvider(...), nil
```

### Adding a new Webhook type

Add a new case in `internal/output/webhook.go` `Send` method with the corresponding payload format.

### Modifying review rules

Edit the `buildSystemPrompt` function in `internal/ai/prompt.go`.

### Running tests

```bash
# All tests
make test

# Specific package
go test ./internal/ai/ -v

# Specific test
go test ./internal/config/ -v -run TestLoadConfigFromFile
```

### Code standards

- Follow Go standard project layout (`internal/` for private packages)
- Use `gofmt` for formatting
- Run `go vet ./...` before committing
- Every module has corresponding test files

## Review Rules

AI review focuses on:

1. **Security**: SQL injection, XSS, sensitive info leaks, insecure encryption
2. **Code Quality**: Code duplication, high complexity, poor naming
3. **Performance**: Memory leaks, goroutine leaks, unnecessary copies
4. **Best Practices**: Error handling, logging standards, concurrency safety
5. **Readability**: Missing comments, complex logic, magic numbers

Custom rules can be added via `review.rules` in config.

## Supported Languages

Go, Python, Java, JavaScript, TypeScript, C, C++, Rust, Ruby, PHP, Swift, Kotlin, Scala, Shell, SQL, HTML, CSS

## License

MIT
