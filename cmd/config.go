package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cr-cli/internal/config"
	"cr-cli/internal/i18n"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: i18n.T("管理配置", "Manage configuration"),
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: i18n.T("显示当前配置", "Show current configuration"),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if cfg == nil {
			return fmt.Errorf("%s", i18n.T(
				"未找到配置文件，请先运行 cr-cli review 进行配置",
				"No config file found, run 'cr-cli review' to set up",
			))
		}
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("marshal config: %w", err)
		}
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configCmd)
}

func loadConfig() (*config.Config, error) {
	if cfgFile != "" {
		return config.Load(cfgFile)
	}
	for _, path := range []string{"cr-cli.yaml", "cr-cli.yml"} {
		if _, err := os.Stat(path); err == nil {
			return config.Load(path)
		}
	}
	home, err := os.UserHomeDir()
	if err == nil {
		path := filepath.Join(home, ".cr-cli", "config.yaml")
		if _, err := os.Stat(path); err == nil {
			return config.Load(path)
		}
	}
	return nil, nil
}

func loadConfigOrSetup() (*config.Config, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	if cfg != nil {
		return cfg, nil
	}
	return interactiveSetup()
}

func interactiveSetup() (*config.Config, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(i18n.T("首次使用，请配置 AI Provider：", "First run, please configure AI Provider:"))
	fmt.Println()

	providerType := prompt(reader, "Provider (openai/anthropic)", "openai")
	baseURL := prompt(reader, "Base URL", defaultBaseURL(providerType))
	apiKey := promptRequired(reader, "API Key")
	model := prompt(reader, "Model", "gpt-4")

	cfg := &config.Config{
		Provider: config.ProviderConfig{
			Type:        providerType,
			APIKey:      apiKey,
			BaseURL:     baseURL,
			Model:       model,
			MaxTokens:   4096,
			Temperature: 0.1,
		},
		Review: config.ReviewConfig{
			Severity:  "warning",
			BatchSize: 10,
		},
		Output: config.OutputConfig{
			Format: "terminal",
			Color:  true,
		},
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return cfg, nil
	}
	dir := filepath.Join(home, ".cr-cli")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return cfg, nil
	}
	path := filepath.Join(dir, "config.yaml")
	data := buildConfigFile(cfg)
	os.WriteFile(path, data, 0644)
	fmt.Printf("\n%s %s\n", i18n.T("配置已保存到", "Config saved to"), path)

	return cfg, nil
}

func prompt(reader *bufio.Reader, label, defaultValue string) string {
	fmt.Printf("  %s [%s]: ", label, defaultValue)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

func promptRequired(reader *bufio.Reader, label string) string {
	for {
		fmt.Printf("  %s: ", label)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			return input
		}
		fmt.Println(i18n.T("  此项为必填项，请重新输入。", "  Required, please try again."))
	}
}

func defaultBaseURL(providerType string) string {
	switch providerType {
	case "anthropic":
		return "https://api.anthropic.com"
	default:
		return "https://api.openai.com/v1"
	}
}

func buildConfigFile(cfg *config.Config) []byte {
	return []byte(fmt.Sprintf(`# CR-CLI Config

provider:
  # Provider type: openai | anthropic
  type: %s
  # API endpoint
  base_url: %s
  # API Key (required)
  api_key: %s
  # Model name
  model: %s
  # Max output tokens
  max_tokens: 4096
  # Temperature (0.0-1.0)
  temperature: 0.1

review:
  # Review mode: warning | blocking
  severity: warning
  # Files per batch
  batch_size: 10
  # Ignore patterns (glob)
  ignore_patterns:
    - "vendor/**"
    - "node_modules/**"
    - "*.pb.go"
    - "*.min.js"
    - "*.min.css"
    - "*.lock"
    - "*.sum"
  # Custom review rules
  rules: []

output:
  format: terminal
  color: true

webhook:
  enabled: false
  url: https://oapi.dingtalk.com/robot/send?access_token=xxx
  type: dingtalk
  secret: ""

gitlab:
  url: https://gitlab.example.com
  token: ""
`, cfg.Provider.Type, cfg.Provider.BaseURL, cfg.Provider.APIKey, cfg.Provider.Model))
}
