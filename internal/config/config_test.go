package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigFromFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "cr-cli.yaml")
	content := `
provider:
  type: openai
  api_key: test-key
  base_url: https://api.test.com/v1
  model: gpt-4
  max_tokens: 2048
  temperature: 0.2
review:
  severity: warning
  batch_size: 5
  languages: [go, python]
  ignore_patterns: ["vendor/**"]
  rules: ["check SQL injection"]
output:
  format: terminal
  color: false
webhook:
  enabled: false
`
	os.WriteFile(cfgPath, []byte(content), 0644)

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider.Type != "openai" {
		t.Errorf("provider type = %q, want %q", cfg.Provider.Type, "openai")
	}
	if cfg.Provider.APIKey != "test-key" {
		t.Errorf("api_key = %q, want %q", cfg.Provider.APIKey, "test-key")
	}
	if cfg.Review.Severity != "warning" {
		t.Errorf("severity = %q, want %q", cfg.Review.Severity, "warning")
	}
	if cfg.Review.BatchSize != 5 {
		t.Errorf("batch_size = %d, want %d", cfg.Review.BatchSize, 5)
	}
}

func TestLoadConfigWithEnvVar(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "cr-cli.yaml")
	content := `provider:
  type: openai
  api_key: ${TEST_API_KEY}
`
	os.WriteFile(cfgPath, []byte(content), 0644)
	t.Setenv("TEST_API_KEY", "sk-secret")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider.APIKey != "sk-secret" {
		t.Errorf("api_key = %q, want %q", cfg.Provider.APIKey, "sk-secret")
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "cr-cli.yaml")
	content := `provider:
  type: openai
  api_key: test
`
	os.WriteFile(cfgPath, []byte(content), 0644)

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("default base_url = %q, want %q", cfg.Provider.BaseURL, "https://api.openai.com/v1")
	}
	if cfg.Review.Severity != "blocking" {
		t.Errorf("default severity = %q, want %q", cfg.Review.Severity, "blocking")
	}
	if cfg.Review.BatchSize != 10 {
		t.Errorf("default batch_size = %d, want %d", cfg.Review.BatchSize, 10)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadConfig_MalformedYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "bad.yaml")
	os.WriteFile(cfgPath, []byte("not: [valid: yaml"), 0644)
	_, err := Load(cfgPath)
	if err == nil {
		t.Error("expected error for malformed YAML")
	}
}

func TestLoadConfig_MissingRequiredFields(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "empty.yaml")
	os.WriteFile(cfgPath, []byte("provider:\n  type: openai\n"), 0644)
	_, err := Load(cfgPath)
	if err == nil {
		t.Error("expected error for missing api_key")
	}
}
