package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if cfg.Server.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Server.Port)
	}
	if cfg.Server.Mode != "debug" {
		t.Errorf("expected default mode debug, got %s", cfg.Server.Mode)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected default log level info, got %s", cfg.LogLevel)
	}
	if cfg.Database.Path != "aigate.db" {
		t.Errorf("expected default db path aigate.db, got %s", cfg.Database.Path)
	}
	if len(cfg.Providers) != 6 {
		t.Errorf("expected 6 default providers, got %d", len(cfg.Providers))
	}
	if cfg.Providers["openai"].Model != "gpt-3.5-turbo" {
		t.Errorf("expected openai model gpt-3.5-turbo, got %s", cfg.Providers["openai"].Model)
	}
	if cfg.Providers["anthropic"].Model != "claude-3-sonnet-20240229" {
		t.Errorf("expected anthropic model claude-3-sonnet-20240229, got %s", cfg.Providers["anthropic"].Model)
	}
}

func TestGetConfigPath_Default(t *testing.T) {
	os.Unsetenv("AIGATE_CONFIG")
	path := getConfigPath()
	if path != "config.yaml" {
		t.Errorf("expected config.yaml, got %s", path)
	}
}

func TestGetConfigPath_Env(t *testing.T) {
	os.Setenv("AIGATE_CONFIG", "/tmp/custom.yaml")
	defer os.Unsetenv("AIGATE_CONFIG")

	path := getConfigPath()
	if path != "/tmp/custom.yaml" {
		t.Errorf("expected /tmp/custom.yaml, got %s", path)
	}
}

func TestLoad_FileNotExist(t *testing.T) {
	os.Setenv("AIGATE_CONFIG", "/tmp/nonexistent_aigate_config.yaml")
	defer os.Unsetenv("AIGATE_CONFIG")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error for missing config, got %v", err)
	}
	if cfg.Server.Port != "8080" {
		t.Errorf("expected default port, got %s", cfg.Server.Port)
	}
}

func TestLoad_ValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
server:
  port: "9090"
  mode: "release"
log_level: "warn"
providers:
  openai:
    enabled: true
    api_key: "test-key"
    base_url: "https://api.openai.com/v1"
    model: "gpt-4"
    timeout: 60
`
	os.WriteFile(configPath, []byte(yamlContent), 0644)
	os.Setenv("AIGATE_CONFIG", configPath)
	defer os.Unsetenv("AIGATE_CONFIG")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Server.Port)
	}
	if cfg.Server.Mode != "release" {
		t.Errorf("expected mode release, got %s", cfg.Server.Mode)
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("expected log level warn, got %s", cfg.LogLevel)
	}
	if !cfg.Providers["openai"].Enabled {
		t.Error("expected openai provider to be enabled")
	}
	if cfg.Providers["openai"].APIKey != "test-key" {
		t.Errorf("expected api key test-key, got %s", cfg.Providers["openai"].APIKey)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "bad.yaml")

	os.WriteFile(configPath, []byte("{{invalid yaml"), 0644)
	os.Setenv("AIGATE_CONFIG", configPath)
	defer os.Unsetenv("AIGATE_CONFIG")

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}
