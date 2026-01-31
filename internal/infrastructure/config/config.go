package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nezdemkovski/folio212/internal/shared/constants"
	"github.com/nezdemkovski/folio212/internal/shared/validation"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Environment      string `mapstructure:"environment" yaml:"environment"`
	Workspace        string `mapstructure:"workspace" yaml:"workspace,omitempty"`
	Trading212Env    string `mapstructure:"trading212_env" yaml:"trading212_env,omitempty"` // "demo" or "live"
	Trading212APIKey string `mapstructure:"trading212_api_key" yaml:"trading212_api_key,omitempty"`
}

var cfg *Config

func Default() *Config {
	return &Config{
		Environment:   "local",
		Trading212Env: "demo",
	}
}

func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, constants.ConfigDirName), nil
}

func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, constants.ConfigFileName), nil
}

func Load() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	cfg = Default()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

func Save(c *Config) error {
	if c == nil {
		return fmt.Errorf("config is required")
	}
	if err := validation.ValidateNonEmpty("environment", c.Environment); err != nil {
		return err
	}

	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	cfg = c
	return nil
}
