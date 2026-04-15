package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 应用全局配置，从 config.yaml 文件加载。
type Config struct {
	Server    ServerConfig              `yaml:"server"`
	Database  DatabaseConfig            `yaml:"database"`
	LogLevel  string                    `yaml:"log_level"`
	JWTSecret string                    `yaml:"jwt_secret"`
	Providers map[string]ProviderConfig `yaml:"providers"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Path string `yaml:"path"` // SQLite 数据库文件路径
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Port string `yaml:"port"` // 监听端口
	Mode string `yaml:"mode"` // Gin 运行模式: debug, release, test
}

// ProviderConfig AI 服务提供者的连接配置。
// 每个 Provider 独立配置 API Key、接口地址、默认模型和超时时间。
type ProviderConfig struct {
	Enabled bool   `yaml:"enabled"`  // 是否启用该提供者
	APIKey  string `yaml:"api_key"`  // API 认证密钥
	BaseURL string `yaml:"base_url"` // API 基础地址（不含路径）
	Model   string `yaml:"model"`    // 默认模型名称
	Timeout int    `yaml:"timeout"`  // HTTP 请求超时（秒）
}

// Load 从配置文件加载并解析配置。
// 优先使用环境变量 AIGATE_CONFIG 指定的路径，否则使用当前目录下的 config.yaml。
// 当配置文件不存在时，返回内置默认配置而不报错。
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

// getConfigPath 获取配置文件路径，支持通过环境变量 AIGATE_CONFIG 覆盖
func getConfigPath() string {
	path := os.Getenv("AIGATE_CONFIG")
	if path != "" {
		return path
	}
	return "config.yaml"
}

// defaultConfig 返回内置默认配置，所有 Provider 默认禁用。
// 包含所有支持的 Provider 的默认接口地址和推荐模型。
func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: "8080",
			Mode: "debug",
		},
		Database: DatabaseConfig{
			Path: "aigate.db",
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
			"deepseek": {
				Enabled: false,
				BaseURL: "https://api.deepseek.com/v1",
				Model:   "deepseek-chat",
				Timeout: 60,
			},
			"doubao": {
				Enabled: false,
				BaseURL: "https://ark.cn-beijing.volces.com/api/v3",
				Model:   "doubao-pro-32k",
				Timeout: 60,
			},
			"qwen": {
				Enabled: false,
				BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
				Model:   "qwen-turbo",
				Timeout: 60,
			},
			"kimi": {
				Enabled: false,
				BaseURL: "https://api.moonshot.cn/v1",
				Model:   "moonshot-v1-8k",
				Timeout: 60,
			},
		},
	}
}
