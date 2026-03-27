package config

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/spf13/viper"
)

type PGConfig struct {
	Username string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"sslmode"`
}

type RedisConfig struct {
	MasterName string   `mapstructure:"master_name"`
	Hosts      []string `mapstructure:"hosts"`
	Password   string   `mapstructure:"password"`
	Database   int      `mapstructure:"database"`
	PoolSize   int      `mapstructure:"pool_size"`
	EnableTLS  bool     `mapstructure:"enable_tls"`
	ReadOnly   bool     `mapstructure:"read_only"`
}

type ProxyServerConfig struct {
	Port        int         `mapstructure:"port"`
	PGConfig    PGConfig    `mapstructure:"pg"`
	RedisConfig RedisConfig `mapstructure:"redis"`
}

type MockServerConfig struct {
	Port               int `mapstructure:"port"`
	ResponseStatusCode int `mapstructure:"response_status_code"`
}

type AdminServerConfig struct {
	Port     int      `mapstructure:"port"`
	PGConfig PGConfig `mapstructure:"pg"`
}

type BackgroundWorkerConfig struct {
	RedisConfig RedisConfig `mapstructure:"redis"`
	Concurrency int         `mapstructure:"concurrency"`
}

type BackgroundSchedulerConfig struct {
	RedisConfig RedisConfig `mapstructure:"redis"`
	PGConfig    PGConfig    `mapstructure:"pg"`
}

// Configs of all services must end with "Config"
type LoadableConfig interface {
	ProxyServerConfig | MockServerConfig | AdminServerConfig | BackgroundWorkerConfig | BackgroundSchedulerConfig
}

func LoadConfig[T LoadableConfig]() (*T, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg T
	if err := v.UnmarshalKey(configName(cfg), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func configName(cfg any) string {
	typeStr := fmt.Sprintf("%T", cfg)                                                   // example: config.ProxyServerConfig
	trimmedType := strings.TrimSuffix(strings.TrimPrefix(typeStr, "config."), "Config") // example: ProxyServer
	return strcase.ToSnake(trimmedType)                                                 // example: proxy_server
}
