# CR-CLI

AI 驱动的代码审查命令行工具，支持 OpenAI 和 Anthropic Provider，可集成 GitLab CI。

[English](README.md)

## 功能特性

- **双 AI Provider**：支持 OpenAI 和 Anthropic API，可自由切换
- **私有化部署**：支持自定义 API 地址，适配企业内网环境
- **智能检测**：自动检测未暂存/已暂存/最近一次提交的代码变更
- **GitLab MR**：支持审查指定 Merge Request
- **综合审查**：安全性、代码质量、性能、最佳实践、可读性
- **分批处理**：文件较多时自动分批审查（默认每批 10 个）
- **Webhook 通知**：支持钉钉、企业微信、飞书
- **CI 集成**：blocking 模式下发现问题返回非零退出码

## 安装

### 从源码编译

```bash
git clone <repo-url>
cd cr-cli
make build
```

### 安装到系统路径

```bash
make install
```

### 跨平台编译

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o cr-cli-linux-amd64 .

# Windows
GOOS=windows GOARCH=amd64 go build -o cr-cli-windows-amd64.exe .

# macOS ARM
GOOS=darwin GOARCH=arm64 go build -o cr-cli-darwin-arm64 .
```

## 快速开始

### 1. 直接使用

在任意 Git 项目目录下直接运行：

```bash
cr-cli review
```

首次使用时会自动进入交互式配置，只需输入 API Key 即可，其余参数均有默认值：

```
首次使用，请配置 AI Provider：

  Provider 类型 (openai/anthropic) [openai]:
  Base URL [https://api.openai.com/v1]:
  API Key: sk-xxx
  Model [gpt-4]:

配置已保存到 ~/.cr-cli/config.yaml
```

配置保存到 `~/.cr-cli/config.yaml`，后续所有项目共享，无需重复配置。

### 2. 运行审查

```bash
# 自动检测：有未提交改动则审查改动，否则审查最近一次提交
cr-cli review

# 审查指定提交
cr-cli review --commit abc1234

# 审查指定分支与主分支的差异
cr-cli review --branch feature/login --base main

# 审查 GitLab MR
cr-cli review --mr 42

# 调试模式（打印请求/响应报文）
cr-cli --debug review
```

## 命令详解

### cr-cli review

审查代码变更。

```bash
cr-cli review [flags]
```

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--commit` | string | - | 审查指定提交（SHA 或 ref） |
| `--mr` | int | - | 审查指定 GitLab MR |
| `--branch` | string | - | 审查指定分支 |
| `--base` | string | `main` | 比较的基础分支 |
| `--batch-size` | int | `10` | 每批审查文件数 |
| `--severity` | string | `warning` | 审查模式：blocking / warning |
| `--config` | string | - | 指定配置文件路径 |
| `--debug` | bool | `false` | 打印调试信息（请求/响应报文） |

**代码获取优先级（自动检测模式）：**

1. 未暂存的修改（`git diff`）
2. 已暂存的修改（`git diff --cached`）
3. 最近一次有代码变更的提交（跳过纯二进制提交）

### cr-cli config show

显示当前加载的配置。

```bash
cr-cli config show
```

## 配置说明

### 配置文件查找顺序

1. `--config` 参数指定的路径
2. 当前目录 `cr-cli.yaml`
3. 当前目录 `cr-cli.yml`
4. `~/.cr-cli/config.yaml`（全局配置，推荐）

首次运行 `cr-cli review` 时如果没有配置文件，会自动进入交互式配置并保存到 `~/.cr-cli/config.yaml`。

### 环境变量

配置文件中支持 `${ENV_VAR}` 语法引用环境变量：

```yaml
provider:
  api_key: ${OPENAI_API_KEY}
gitlab:
  token: ${GITLAB_TOKEN}
```

### 审查模式

| 模式 | 说明 | 退出码 |
|------|------|--------|
| `warning`（默认） | 仅输出报告，不影响退出码 | 始终 0 |
| `blocking` | 发现 error 级别问题时返回退出码 1 | 0 或 1 |

### Webhook 类型

| 类型 | 说明 |
|------|------|
| `dingtalk` | 钉钉机器人 |
| `wecom` | 企业微信机器人 |
| `feishu` | 飞书机器人 |
| `custom` | 通用 JSON 格式 |

## GitLab CI 集成

### .gitlab-ci.yml 示例

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

### 作为 Pipeline 闸门

```yaml
code-review-gate:
  stage: review
  script:
    - ./cr-cli review --severity blocking
  allow_failure: false  # 审查失败则 Pipeline 失败
```

## 输出示例

### 终端输出

```
╔══════════════════════════════════════════════════════════════╗
║                    CR-CLI 代码审查报告                       ║
╚══════════════════════════════════════════════════════════════╝

📁 文件: main.go (Go)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  ✗ 第 42 行 [ERROR] SQL 注入风险
    用户输入直接拼接 SQL 语句，存在注入漏洞
    💡 建议: 使用参数化查询替代字符串拼接

  ⚠ 第 78 行 [WARNING] 错误处理不完整
    忽略了 err 返回值
    💡 建议: 添加 if err != nil 检查

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 统计: 1 个错误, 1 个警告, 0 个提示
```

### Webhook Markdown 通知

```markdown
## 🔍 代码审查报告

**时间**: 2024-01-15 14:30:00
**仓库**: /path/to/repo
**提交**: HEAD

### 发现问题

| 严重程度 | 文件 | 行号 | 问题 |
|---------|------|------|------|
| 🔴 ERROR | main.go | 42 | SQL 注入风险 |
| 🟡 WARNING | main.go | 78 | 错误处理不完整 |

**统计**: 1 个错误, 1 个警告, 0 个提示
```

## 项目结构

```
cr-cli/
├── cmd/                        # Cobra CLI 命令
│   ├── root.go                # 根命令
│   ├── review.go              # review 子命令
│   └── config.go              # config 子命令（含首次交互式配置）
├── internal/
│   ├── config/                # 配置加载
│   │   ├── config.go          # YAML 解析 + 环境变量 + 默认值
│   │   └── config_test.go
│   ├── git/                   # Git 操作
│   │   ├── diff.go            # Diff 获取（自动检测/指定提交/分支）
│   │   ├── diff_test.go
│   │   ├── changeset.go       # 变更集解析 + 语言检测
│   │   └── changeset_test.go
│   ├── gitlab/                # GitLab API
│   │   ├── client.go          # MR Diff 获取
│   │   └── client_test.go
│   ├── ai/                    # AI Provider
│   │   ├── provider.go        # 接口定义
│   │   ├── prompt.go          # Prompt 模板
│   │   ├── openai.go          # OpenAI 实现
│   │   ├── openai_test.go
│   │   ├── anthropic.go       # Anthropic 实现
│   │   └── anthropic_test.go
│   ├── review/                # 审查引擎
│   │   ├── engine.go          # 分批处理 + 结果聚合
│   │   └── engine_test.go
│   └── output/                # 输出
│       ├── terminal.go        # 终端彩色输出
│       ├── terminal_test.go
│       ├── webhook.go         # Webhook 通知
│       └── webhook_test.go
├── main.go                    # 入口
├── go.mod
├── go.sum
└── Makefile
```

## 开发

### 环境要求

- Go 1.21+
- Git

### 开发流程

```bash
git clone <repo-url>
cd cr-cli
go mod tidy
make test
make build
./cr-cli --help
```

### 添加新的 AI Provider

1. 在 `internal/ai/` 下创建新文件，如 `deepseek.go`
2. 实现 `Provider` 接口：

```go
type Provider interface {
    Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error)
    Name() string
}
```

3. 在 `cmd/review.go` 的 `createProvider` 函数中注册：

```go
case "deepseek":
    return ai.NewDeepSeekProvider(...), nil
```

### 添加新的 Webhook 类型

在 `internal/output/webhook.go` 的 `Send` 方法中添加新的 case，实现对应的 payload 格式。

### 修改审查规则

编辑 `internal/ai/prompt.go` 中的 `buildSystemPrompt` 函数。

### 运行测试

```bash
# 运行所有测试
make test

# 运行特定包的测试
go test ./internal/ai/ -v

# 运行特定测试
go test ./internal/config/ -v -run TestLoadConfigFromFile
```

### 代码规范

- 遵循 Go 标准项目布局（`internal/` 私有包）
- 使用 `gofmt` 格式化代码
- 提交前运行 `go vet ./...` 检查
- 每个模块都有对应的测试文件

## 审查规则

AI 审查默认关注以下方面：

1. **安全性**：SQL 注入、XSS、敏感信息泄露、不安全的加密
2. **代码质量**：代码重复、复杂度过高、命名不规范
3. **性能**：内存泄漏、goroutine 泄漏、不必要的拷贝
4. **最佳实践**：错误处理、日志规范、并发安全
5. **可读性**：注释缺失、逻辑复杂、魔法数字

可通过配置文件 `review.rules` 添加自定义规则。

## 支持的语言

Go, Python, Java, JavaScript, TypeScript, C, C++, Rust, Ruby, PHP, Swift, Kotlin, Scala, Shell, SQL, HTML, CSS

## License

MIT
