package config

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/spf13/viper"
)

const emptyVal = "!empty!"

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

func (cfg RedisConfig) GetPassword() string {
	if cfg.Password == emptyVal {
		return ""
	}
	return cfg.Password
}

func (cfg RedisConfig) GetMasterName() string {
	if cfg.MasterName == emptyVal {
		return ""
	}
	return cfg.MasterName
}

type ProxyServerConfig struct {
	Port            int         `mapstructure:"port"`
	SwaggerHost     string      `mapstructure:"swagger_host"`
	AWSCognitoGroup string      `mapstructure:"aws_cognito_group"`
	PGConfig        PGConfig    `mapstructure:"pg"`
	RedisConfig     RedisConfig `mapstructure:"redis"`
}

type AdminServerConfig struct {
	Port            int      `mapstructure:"port"`
	SwaggerHost     string   `mapstructure:"swagger_host"`
	AWSCognitoGroup string   `mapstructure:"aws_cognito_group"`
	PGConfig        PGConfig `mapstructure:"pg"`
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

func LoadConfig[T LoadableConfig]() (*T, error) {
	var cfg T

	v := viper.New()
	v.SetConfigName(configName(cfg))
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func configName(cfg any) string {
	typeStr := fmt.Sprintf("%T", cfg)                                                   // example: config.ProxyServerConfig
	trimmedType := strings.TrimSuffix(strings.TrimPrefix(typeStr, "config."), "Config") // example: ProxyServer
	return strcase.ToSnake(trimmedType)                                                 // example: proxy_server
}
