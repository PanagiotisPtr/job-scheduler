package config

import (
	"flag"
	"os"

	"github.com/spf13/viper"
)

type ServiceConfig struct {
	Port int `mapstructure:"port"`
}

type GitHubRepositoryArgs struct {
	Owner string `mapstructure:"owner"`
	Name  string `mapstructure:"name"`
	Path  string `mapstructure:"path"`
}

type GitHubConfig struct {
	AccessToken string                 `mapstructure:"accessToken"`
	Locations   []GitHubRepositoryArgs `mapstructure:"locations"`
}

type Config struct {
	Service      ServiceConfig `mapstructure:"service"`
	GitHubCinfig GitHubConfig  `mapstructure:"githubConfig"`
}

func loadConfig(filename string) (*Config, error) {
	viper.SetConfigFile(filename)
	viper.AddConfigPath(".")

	var config Config

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func ProvideConfig() (*Config, error) {
	configFilename := *flag.String("config", "config.dev.yml", "configuration file")
	flag.Parse()

	if os.Getenv("CONFIG") != "" {
		configFilename = os.Getenv("CONFIG")
	}

	return loadConfig(configFilename)
}

func ProvideTestConfig() (*Config, error) {
	configFilename := *flag.String("test-config", "config.test.yml", "configuration file")
	flag.Parse()

	if os.Getenv("TEST_CONFIG") != "" {
		configFilename = os.Getenv("TEST_CONFIG")
	}

	return loadConfig(configFilename)
}
