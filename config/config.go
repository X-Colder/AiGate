package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config 应用全局配置，从 config.yaml 文件加载。
type Config struct {
	Server    ServerConfig              `yaml:"server"`
	Database  DatabaseConfig            `yaml:"database"`
	Redis     RedisConfig               `yaml:"redis"`
	LogLevel  string                    `yaml:"log_level"`
	JWTSecret string                    `yaml:"jwt_secret"`
	Providers map[string]ProviderConfig `yaml:"providers"`
}

type DatabaseConfig struct {
	Driver          string `yaml:"driver"`
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime int    `yaml:"conn_max_idle_time"`
	Path            string `yaml:"path"`
}

// RedisConfig Redis 连接配置
type RedisConfig struct {
	Addr         string `yaml:"addr"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
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
			applyEnvOverrides(cfg)
			return cfg, nil
		}
		return nil, fmt.Errorf("read config file error: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file error: %w", err)
	}

	applyEnvOverrides(cfg)

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
			Driver:          "mysql",
			Host:            "127.0.0.1",
			Port:            3306,
			Username:        "aigate",
			Password:        "aigate_pass",
			Database:        "aigate",
			MaxOpenConns:    25,
			MaxIdleConns:    10,
			ConnMaxLifetime: 300,
			ConnMaxIdleTime: 60,
		},
		Redis: RedisConfig{
			Addr:         "127.0.0.1:6379",
			Password:     "",
			DB:           0,
			PoolSize:     50,
			MinIdleConns: 10,
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

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("AIGATE_SERVER_PORT"); v != "" {
		cfg.Server.Port = v
	}
	if v := os.Getenv("AIGATE_SERVER_MODE"); v != "" {
		cfg.Server.Mode = v
	}
	if v := os.Getenv("AIGATE_DB_DRIVER"); v != "" {
		cfg.Database.Driver = v
	}
	if v := os.Getenv("AIGATE_DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("AIGATE_DB_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Database.Port = p
		}
	}
	if v := os.Getenv("AIGATE_DB_USERNAME"); v != "" {
		cfg.Database.Username = v
	}
	if v := os.Getenv("AIGATE_DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("AIGATE_DB_DATABASE"); v != "" {
		cfg.Database.Database = v
	}
	if v := os.Getenv("AIGATE_REDIS_ADDR"); v != "" {
		cfg.Redis.Addr = v
	}
	if v := os.Getenv("AIGATE_REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	if v := os.Getenv("AIGATE_JWT_SECRET"); v != "" {
		cfg.JWTSecret = v
	}
	if v := os.Getenv("AIGATE_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
}
