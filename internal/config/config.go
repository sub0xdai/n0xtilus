package config

import (
	"fmt"
	"log"
	"github.com/spf13/viper"
)

type Config struct {
	APIKey         string  `mapstructure:"api_key"`
	APISecret      string  `mapstructure:"api_secret"`
	APIBaseURL     string  `mapstructure:"api_base_url"`
	RiskPercentage float64 `mapstructure:"risk_percentage"`
	TestMode       bool    `mapstructure:"test_mode"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	log.Printf("Config file used: %s", viper.ConfigFileUsed())

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	log.Printf("Config loaded: %+v", config)
	return &config, nil
}
