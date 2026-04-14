package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server    ServerConfig              `yaml:"server"`
	LogLevel  string                    `yaml:"log_level"`
	Providers map[string]ProviderConfig `yaml:"providers"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port string `yaml:"port"`
	Mode string `yaml:"mode"` // debug, release, test
}

// ProviderConfig AI 服务提供者配置
type ProviderConfig struct {
	Enabled bool   `yaml:"enabled"`
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
	Timeout int    `yaml:"timeout"` // 超时秒数
}

// Load 加载配置文件
func Load() (*Config, error) {
	cfg := defaultConfig()

	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 配置文件不存在，使用默认配置
			return cfg, nil
		}
		return nil, fmt.Errorf("read config file error: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file error: %w", err)
	}

	return cfg, nil
}

// getConfigPath 获取配置文件路径
func getConfigPath() string {
	path := os.Getenv("AIGATE_CONFIG")
	if path != "" {
		return path
	}
	return "config.yaml"
}

// defaultConfig 返回默认配置
func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: "8080",
			Mode: "debug",
		},
		LogLevel: "info",
		Providers: map[string]ProviderConfig{
			"openai": {
				Enabled: false,
				BaseURL: "https://api.openai.com/v1",
				Model:   "gpt-3.5-turbo",
				Timeout: 30,
			},
			"anthropic": {
				Enabled: false,
				BaseURL: "https://api.anthropic.com/v1",
				Model:   "claude-3-sonnet-20240229",
				Timeout: 30,
			},
		},
	}
}
