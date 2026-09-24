package config

import (
	"os"
	"path"

	"github.com/spf13/viper"
)

type Config struct {
	Output	OutputConfig `mapstructure:"output" validate:"required"`
	Log		LogConfig `mapstructure:"log" validate:"required"`
}

type OutputConfig struct {
	Base string `mapstructure:"base" validate:"requred,oneof=ext, date"`
	Mapping map[string]string `mapstructure:"mapping"`
}

type LogConfig struct {
	Format string `mapstructure:"format" validate:"oneof=json"`
	Level string  `mapstructure:"level" validate:"oneof=info debug"`
}

func LoadConfig() (*Config, error) {
	dir, err := os.Getwd(); if err != nil {
		return nil, err
	}
	viper.AddConfigPath(path.Join(dir, "./configs/config.yaml"))
	viper.SetDefault("output.base", "ext")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.level", "info")
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}
	return &config, nil
}