package config

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/spf13/viper"
)

type PGConfig struct {
	Username    string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Database    string `mapstructure:"database"`
	SSLMode     string `mapstructure:"sslmode"`
	SSLRootCert string `mapstructure:"sslrootcert"`
}

type RedisConfig struct {
	MasterName string   `mapstructure:"master_name"`
	Addrs      []string `mapstructure:"addrs"`
	Password   string   `mapstructure:"password"`
	Database   int      `mapstructure:"database"`
	PoolSize   int      `mapstructure:"pool_size"`
	EnableTLS  bool     `mapstructure:"enable_tls"`
	ReadOnly   bool     `mapstructure:"read_only"`
}

type ProxyServerConfig struct {
	Port         int         `mapstructure:"port"`
	AllowOrigins []string    `mapstructure:"allow_origins"`
	AllowMethods []string    `mapstructure:"allow_methods"`
	SwaggerHost  string      `mapstructure:"swagger_host"`
	PGConfig     PGConfig    `mapstructure:"pg"`
	RedisConfig  RedisConfig `mapstructure:"redis"`
}

type AdminServerConfig struct {
	Port         int      `mapstructure:"port"`
	AllowOrigins []string `mapstructure:"allow_origins"`
	AllowMethods []string `mapstructure:"allow_methods"`
	SwaggerHost  string   `mapstructure:"swagger_host"`
	PGConfig     PGConfig `mapstructure:"pg"`
}

type BackgroundWorkerConfig struct {
	RedisConfig RedisConfig `mapstructure:"redis"`
	Concurrency int         `mapstructure:"concurrency"`
}

type BackgroundSchedulerConfig struct {
	RedisConfig RedisConfig `mapstructure:"redis"`
	PGConfig    PGConfig    `mapstructure:"pg"`
}

type MigrationConfig struct {
	Source                  string   `mapstructure:"source"`
	PGConfig                PGConfig `mapstructure:"pg"`
	PGPasswordAWSSecretName string   `mapstructure:"pg_password_aws_secret_name"`
	AWSRegion               string   `mapstructure:"aws_region"`
}

type MockServerConfig struct {
	Port               int `mapstructure:"port"`
	ResponseStatusCode int `mapstructure:"response_status_code"`
}

// Configs of all services must end with "Config"
type LoadableConfig interface {
	ProxyServerConfig | MockServerConfig | AdminServerConfig | BackgroundWorkerConfig | BackgroundSchedulerConfig | MigrationConfig
}

func LoadConfig[T LoadableConfig]() (T, error) {
	var cfg T

	v := viper.New()
	v.SetConfigName(configName(cfg))
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.SetTypeByDefaultValue(true)

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return cfg, err
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func configName(cfg any) string {
	typeStr := fmt.Sprintf("%T", cfg)                                                   // example: config.ProxyServerConfig
	trimmedType := strings.TrimSuffix(strings.TrimPrefix(typeStr, "config."), "Config") // example: ProxyServer
	return strcase.ToSnake(trimmedType)                                                 // example: proxy_server
}
