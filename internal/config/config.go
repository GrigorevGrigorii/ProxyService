package config

import (
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

type ProxyServerConfig struct {
	Port         int    `mapstructure:"port"`
	ServicesPath string `mapstructure:"services_path"`
}

type MockServerConfig struct {
	Port               int `mapstructure:"port"`
	ResponseStatusCode int `mapstructure:"response_status_code"`
}

type AdminServerConfig struct {
	PGConfig PGConfig `mapstructure:"pg"`
	Port     int      `mapstructure:"port"`
}

func newViper() *viper.Viper {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AutomaticEnv()
	return v
}

func LoadProxyServer() (*ProxyServerConfig, error) {
	v := newViper()
	v.BindEnv("proxy_server.port", "PORT")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg ProxyServerConfig
	if err := v.UnmarshalKey("proxy_server", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadMockServer() (*MockServerConfig, error) {
	v := newViper()
	v.BindEnv("mock_server.port", "PORT")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg MockServerConfig
	if err := v.UnmarshalKey("mock_server", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadAdminServer() (*AdminServerConfig, error) {
	v := newViper()
	v.BindEnv("admin_server.port", "PORT")
	v.BindEnv("admin_server.pg.user", "PG_USER")
	v.BindEnv("admin_server.pg.password", "PG_PASSWORD")
	v.BindEnv("admin_server.pg.host", "PG_HOST")
	v.BindEnv("admin_server.pg.database", "PG_DATABASE")
	v.BindEnv("admin_server.pg.sslmode", "PG_USE_SSL")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg AdminServerConfig
	if err := v.UnmarshalKey("admin_server", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
