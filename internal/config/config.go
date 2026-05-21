package config

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Provider ProviderConfig `yaml:"provider"`
	Review   ReviewConfig   `yaml:"review"`
	Output   OutputConfig   `yaml:"output"`
	Webhook  WebhookConfig  `yaml:"webhook"`
	GitLab   GitLabConfig   `yaml:"gitlab"`
}

type ProviderConfig struct {
	Type        string  `yaml:"type"`
	APIKey      string  `yaml:"api_key"`
	BaseURL     string  `yaml:"base_url"`
	Model       string  `yaml:"model"`
	MaxTokens   int     `yaml:"max_tokens"`
	Temperature float64 `yaml:"temperature"`
}

type ReviewConfig struct {
	Severity       string   `yaml:"severity"`
	BatchSize      int      `yaml:"batch_size"`
	IgnorePatterns []string `yaml:"ignore_patterns"`
	Rules          []string `yaml:"rules"`
}

type OutputConfig struct {
	Format string `yaml:"format"`
	Color  bool   `yaml:"color"`
}

type WebhookConfig struct {
	Enabled   bool   `yaml:"enabled"`
	URL       string `yaml:"url"`
	Type      string `yaml:"type"`
	Secret    string `yaml:"secret"`
	OnSuccess bool   `yaml:"on_success"`
	OnFailure bool   `yaml:"on_failure"`
}

type GitLabConfig struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	expanded := expandEnvVars(string(data))

	cfg := &Config{}
	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	applyDefaults(cfg)
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Provider.Type == "" {
		return fmt.Errorf("provider.type is required")
	}
	if c.Provider.APIKey == "" {
		return fmt.Errorf("provider.api_key is required")
	}
	return nil
}

var envVarRe = regexp.MustCompile(`\$\{([^}]+)\}`)

func expandEnvVars(s string) string {
	return envVarRe.ReplaceAllStringFunc(s, func(match string) string {
		key := match[2 : len(match)-1]
		if val, ok := os.LookupEnv(key); ok {
			return val
		}
		return match
	})
}

// applyDefaults fills unset fields with sensible defaults.
// Note: because defaults are applied via zero-value checks (== "" / == 0),
// explicitly setting a field to its zero value in the config file (e.g.
// temperature: 0) will be overridden by the default. There is no way to
// distinguish "not set" from "explicitly set to zero" with the current
// approach.
func applyDefaults(cfg *Config) {
	if cfg.Provider.BaseURL == "" {
		switch cfg.Provider.Type {
		case "anthropic":
			cfg.Provider.BaseURL = "https://api.anthropic.com"
		default:
			cfg.Provider.BaseURL = "https://api.openai.com/v1"
		}
	}
	if cfg.Provider.MaxTokens == 0 {
		cfg.Provider.MaxTokens = 4096
	}
	if cfg.Provider.Temperature == 0 {
		cfg.Provider.Temperature = 0.1
	}
	if cfg.Review.Severity == "" {
		cfg.Review.Severity = "warning"
	}
	if cfg.Review.BatchSize == 0 {
		cfg.Review.BatchSize = 10
	}
	if cfg.Output.Format == "" {
		cfg.Output.Format = "terminal"
	}
}
