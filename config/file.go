package config

import (
	"charm.land/log/v2"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type IO interface {
	write(config *Config) error
	read() (*Config, error)
}

type defaultConfigIO struct {
	configPath string
}

func NewDefaultIO() IO {
	return &defaultConfigIO{configPath()}
}

func (c *defaultConfigIO) read() (*Config, error) {
	// Check if the file exists
	if _, err := os.Stat(c.configPath); os.IsNotExist(err) {
		// Ensure the directory structure exists
		if err := os.MkdirAll(filepath.Dir(c.configPath), 0755); err != nil {
			return nil, err
		}

		// Create the file
		if _, err := os.Create(c.configPath); err != nil {
			return nil, err
		}
	}

	data, err := os.ReadFile(c.configPath)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	if err = yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}

func (c *defaultConfigIO) write(config *Config) error {
	out, err := yaml.Marshal(config)
	if err != nil {
		log.Fatalf("Error marshalling config: %v", err)
		return err
	}

	err = os.WriteFile(c.configPath, out, 0644)
	if err != nil {
		log.Fatalf("Error writing config file: %v", err)
		return err
	}

	return nil
}

func configPath() string {
	return filepath.Join(configDir(), "config.yaml")
}

// ThemesDir is where user themes live, one <name>.yaml per theme.
func ThemesDir() string {
	return filepath.Join(configDir(), "themes")
}

func configDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		os.Exit(-1)
	}
	return filepath.Join(homeDir, ".config", "ktea")
}
