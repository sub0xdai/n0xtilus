package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	APIKey         string
	APISecret      string
  APIBaseURL     string
	RiskPercentage float64
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
