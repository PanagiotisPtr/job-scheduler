package config

import (
	"flag"
	"os"

	"github.com/spf13/viper"
)

type ServiceConfig struct {
	Port int
}

type GitHubRepositoryArgs struct {
	Owner string
	Name  string
	Path  string
}

type GitHubConfig struct {
	AccessToken string
	Locations   []GitHubRepositoryArgs
}

type Config struct {
	Service      ServiceConfig
	GitHubCinfig GitHubConfig
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
