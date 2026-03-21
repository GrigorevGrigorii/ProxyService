package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	ProxyServer ProxyServerConfig `mapstructure:"proxy_server"`
	MockServer  MockServerConfig  `mapstructure:"mock_server"`
}

type ProxyServerConfig struct {
	Port         int    `mapstructure:"port"`
	ServicesPath string `mapstructure:"services_path"`
}

type MockServerConfig struct {
	Port               int `mapstructure:"port"`
	ResponseStatusCode int `mapstructure:"response_status_code"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Environment variables override
	viper.AutomaticEnv()
	viper.BindEnv("server.port", "PORT")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
